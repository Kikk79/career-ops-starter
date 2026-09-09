package app

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/santifer/career-ops/dashboard/internal/model"
)

// makeOutputDir legt ein career-ops-artiges Verzeichnis mit den angegebenen
// HTML-Fassungen unter output/ an.
func makeOutputDir(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	out := filepath.Join(root, "output")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatalf("output/ nicht anlegbar: %v", err)
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(out, f), []byte("<html></html>"), 0o644); err != nil {
			t.Fatalf("%s nicht schreibbar: %v", f, err)
		}
	}
	return root
}

func TestCompanyWords(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"Rechtsform entfällt", "Globex GmbH", []string{"globex"}},
		{"mehrteilig", "Initech Klinikum Musterstadt", []string{"initech", "klinikum", "musterstadt"}},
		{"Umlaut umschrieben", "Landkreis Bad Tölz-Starkhausen", []string{"landkreis", "toelz", "starkhausen"}},
		{"kurze Teile entfallen", "Bad AG", nil},
		{"unbekannte Firma", "?", nil},
		{"leer", "", nil},
		{"Bindestrich trennt", "Stark Industries", []string{"stark", "industries"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := companyWords(tc.in)
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("companyWords(%q) = %#v, erwartet %#v", tc.in, got, tc.want)
			}
		})
	}
}

// Der Rückfallweg muss genau die Fälle lösen, an denen die geteilte
// Datenschicht scheitert: Dateiname trägt eine Kurzform des Firmennamens.
func TestResolveCVHTMLFindsShortenedFilenames(t *testing.T) {
	root := makeOutputDir(t,
		"cv-max-mustermann-globex.html",
		"cv-max-mustermann-initech.html",
		"cv-max-mustermann-lra-toelz.html",
		"cv-max-mustermann-stark-industries.html",
	)
	tests := []struct {
		company  string
		wantFile string
	}{
		{"Globex GmbH", "output/cv-max-mustermann-globex.html"},
		{"Initech Klinikum Musterstadt", "output/cv-max-mustermann-initech.html"},
		{"Landkreis Bad Tölz-Starkhausen", "output/cv-max-mustermann-lra-toelz.html"},
		{"Stark Industries", "output/cv-max-mustermann-stark-industries.html"},
	}
	for _, tc := range tests {
		t.Run(tc.company, func(t *testing.T) {
			html, pdf, found := resolveCVHTML(root, model.CareerApplication{Company: tc.company})
			if !found {
				t.Fatalf("keine Fassung für %q gefunden", tc.company)
			}
			if html != tc.wantFile {
				t.Errorf("HTML = %q, erwartet %q", html, tc.wantFile)
			}
			if want := tc.wantFile[:len(tc.wantFile)-len(".html")] + ".pdf"; pdf != want {
				t.Errorf("PDF = %q, erwartet %q", pdf, want)
			}
		})
	}
}

// Eine Bewerbung über eine Agentur führt die Firma als "?" — dafür gibt es
// keinen automatischen Treffer, und das muss die Oberfläche erfahren, damit
// sie die Auswahl anbietet.
func TestResolveCVHTMLUnknownCompanyReportsNotFound(t *testing.T) {
	root := makeOutputDir(t, "cv-max-mustermann-umbrella.html")
	if _, _, found := resolveCVHTML(root, model.CareerApplication{Company: "?"}); found {
		t.Error("unbekannte Firma dürfte keinen Treffer liefern")
	}
}

// Eine Rechtsform allein darf nicht auf eine fremde Firma passen.
func TestResolveCVHTMLDoesNotMatchOnLegalFormAlone(t *testing.T) {
	root := makeOutputDir(t, "cv-max-mustermann-gmbh-holding.html")
	if _, _, found := resolveCVHTML(root, model.CareerApplication{Company: "Muster GmbH"}); found {
		t.Error("Treffer allein über die Rechtsform")
	}
}

func TestResolveCVHTMLNoFilesAtAll(t *testing.T) {
	root := makeOutputDir(t)
	if _, _, found := resolveCVHTML(root, model.CareerApplication{Company: "Globex GmbH"}); found {
		t.Error("Treffer ohne jede Datei")
	}
}

func TestListCVHTMLReturnsRepoRelativeSlashPaths(t *testing.T) {
	root := makeOutputDir(t, "cv-a.html", "cv-b.html")
	got := listCVHTML(root)
	if len(got) != 2 {
		t.Fatalf("%d Dateien, erwartet 2", len(got))
	}
	for _, p := range got {
		if filepath.IsAbs(p) {
			t.Errorf("%q ist absolut, erwartet repo-relativ", p)
		}
		if p[:7] != "output/" {
			t.Errorf("%q beginnt nicht mit output/", p)
		}
	}
}
