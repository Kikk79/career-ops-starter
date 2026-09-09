package app

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/santifer/career-ops/dashboard/internal/data"
	"github.com/santifer/career-ops/dashboard/internal/model"
)

// legalFormWords sind Rechtsformen und Allerweltsbegriffe, die für sich
// genommen keine Firma identifizieren. Ein Treffer allein auf so einem Wort
// würde die falsche Datei liefern.
var legalFormWords = map[string]bool{
	"gmbh": true, "mbh": true, "aktiengesellschaft": true, "holding": true,
	"group": true, "gruppe": true, "company": true, "corporation": true,
	"deutschland": true, "germany": true, "international": true,
	"systems": true, "solutions": true, "services": true, "technologies": true,
}

// minCompanyWordLen verhindert, dass Kürzel wie "bad" oder "ag" auf eine
// fremde Datei passen.
const minCompanyWordLen = 4

// resolveCVHTML findet die zugeschnittene HTML-Fassung des Lebenslaufs für
// eine Bewerbung und den daraus abgeleiteten PDF-Pfad.
//
// Zuerst wird die geteilte Datenschicht befragt. Die matcht auf den
// vollständigen Firmen-Slug und geht deshalb leer aus, sobald die Datei nach
// einer Kurzform benannt ist — "Tyczka GmbH" gegen `cv-…-tyczka.html`,
// "RoMed Klinikum Rosenheim" gegen `cv-…-romed.html`. Genau das ist in der
// Praxis der Normalfall, weil die Dateinamen von Hand vergeben werden.
//
// Der Rückfallweg sucht deshalb nach einem einzelnen, aussagekräftigen Wort
// des Firmennamens. Bleibt auch das ergebnislos, meldet exact=false und die
// Oberfläche lässt die Datei von Hand auswählen, statt kommentarlos nichts zu tun.
func resolveCVHTML(careerOpsPath string, app model.CareerApplication) (htmlPath, pdfPath string, found bool) {
	if h, p := data.ResolveHTML(careerOpsPath, app); h != "" {
		return h, p, true
	}
	candidates := listCVHTML(careerOpsPath)
	if len(candidates) == 0 {
		return "", "", false
	}
	for _, word := range companyWords(app.Company) {
		for _, rel := range candidates {
			if strings.Contains(strings.ToLower(filepath.Base(rel)), word) {
				return rel, strings.TrimSuffix(rel, ".html") + ".pdf", true
			}
		}
	}
	return "", "", false
}

// companyWords zerlegt einen Firmennamen in die Bestandteile, die ihn
// tatsächlich identifizieren — ohne Rechtsformen und ohne zu kurze Teile.
// Umlaute werden wie in den Dateinamen umschrieben.
func companyWords(company string) []string {
	normalized := strings.ToLower(company)
	for from, to := range map[string]string{"ä": "ae", "ö": "oe", "ü": "ue", "ß": "ss"} {
		normalized = strings.ReplaceAll(normalized, from, to)
	}
	fields := strings.FieldsFunc(normalized, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if len(f) >= minCompanyWordLen && !legalFormWords[f] {
			out = append(out, f)
		}
	}
	return out
}

// listCVHTML liefert alle zugeschnittenen Lebenslauf-Fassungen unter output/,
// neueste zuerst, als repo-relative Pfade mit Schrägstrichen.
func listCVHTML(careerOpsPath string) []string {
	globbed, err := filepath.Glob(filepath.Join(careerOpsPath, "output", "cv-*.html"))
	if err != nil {
		return nil
	}
	sort.Slice(globbed, func(i, j int) bool {
		return modTime(globbed[i]).After(modTime(globbed[j]))
	})
	out := make([]string, 0, len(globbed))
	for _, p := range globbed {
		rel, err := filepath.Rel(careerOpsPath, p)
		if err != nil {
			continue
		}
		out = append(out, filepath.ToSlash(rel))
	}
	return out
}

// modTime liefert den Änderungszeitpunkt, bei Fehlern die Nullzeit — dann
// landet die Datei am Ende der Sortierung.
func modTime(path string) time.Time {
	st, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return st.ModTime()
}
