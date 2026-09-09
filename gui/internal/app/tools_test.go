package app

import (
	"testing"

	"fyne.io/fyne/v2/widget"

	"github.com/santifer/career-ops/dashboard/guiapp/internal/runner"
)

// newTestToolsView baut eine Werkzeugansicht ohne Fyne-Oberfläche, nur mit der
// Baumlogik — die Widgets brauchen für diese Tests nicht zu existieren.
func newTestToolsView() *toolsView {
	return &toolsView{groups: runner.Catalog()}
}

// Ein leerer Baum war der Fehler, den der erste Lauftest zeigte: ohne Kinder
// unter der Wurzel bleibt der ganze Werkzeugbereich leer.
func TestTreeRootHasAllGroups(t *testing.T) {
	tv := newTestToolsView()
	roots := tv.childIDs("")
	if len(roots) != len(runner.Catalog()) {
		t.Fatalf("Wurzel hat %d Kinder, erwartet %d", len(roots), len(runner.Catalog()))
	}
	if len(roots) == 0 {
		t.Fatal("Wurzel hat keine Kinder — der Baum bliebe leer")
	}
	for _, id := range roots {
		if !tv.isBranch(id) {
			t.Errorf("%q müsste ein aufklappbarer Knoten sein", id)
		}
		if tv.labelFor(id) == "" {
			t.Errorf("%q hat keinen Anzeigetext", id)
		}
	}
}

// Die Wurzel selbst muss als Zweig gelten, sonst fragt der Baum ihre Kinder
// nicht ab.
func TestTreeRootIsBranch(t *testing.T) {
	if !newTestToolsView().isBranch("") {
		t.Error("die Wurzel muss ein Zweig sein")
	}
}

func TestTreeGroupsHaveTools(t *testing.T) {
	tv := newTestToolsView()
	total := 0
	for gi, g := range tv.groups {
		children := tv.childIDs(groupID(gi))
		if len(children) != len(g.Tools) {
			t.Errorf("Gruppe %q hat %d Kinder, erwartet %d", g.Name, len(children), len(g.Tools))
		}
		for _, id := range children {
			if tv.isBranch(id) {
				t.Errorf("%q darf kein Zweig sein", id)
			}
			if tv.toolByID(id) == nil {
				t.Errorf("%q lässt sich nicht auf ein Werkzeug auflösen", id)
			}
			total++
		}
	}
	if total == 0 {
		t.Fatal("der Baum enthält kein einziges Werkzeug")
	}
}

// Jede erzeugte Werkzeug-ID muss genau das Werkzeug zurückliefern, aus dem sie
// gebaut wurde — sonst startet ein Klick das falsche Skript.
func TestToolByIDRoundTrip(t *testing.T) {
	tv := newTestToolsView()
	for gi, g := range tv.groups {
		for ti, want := range g.Tools {
			got := tv.toolByID(toolID(gi, ti))
			if got == nil {
				t.Fatalf("toolByID(%q) = nil", toolID(gi, ti))
			}
			if got.Script != want.Script || got.Name != want.Name {
				t.Errorf("toolByID(%q) = %q/%q, erwartet %q/%q",
					toolID(gi, ti), got.Name, got.Script, want.Name, want.Script)
			}
		}
	}
}

func TestToolByIDRejectsInvalid(t *testing.T) {
	tv := newTestToolsView()
	for _, id := range []widget.TreeNodeID{"", "g:0", "t:", "t:0", "t:x:1", "t:0:x", "t:999:0", "t:0:999", "t:-1:0", "quatsch"} {
		if got := tv.toolByID(id); got != nil {
			t.Errorf("toolByID(%q) = %v, erwartet nil", id, got.Name)
		}
	}
}

func TestParseGroupID(t *testing.T) {
	tests := []struct {
		in     widget.TreeNodeID
		want   int
		wantOK bool
	}{
		{"g:0", 0, true},
		{"g:12", 12, true},
		{"g:-1", 0, false},
		{"g:", 0, false},
		{"g:x", 0, false},
		{"t:0:1", 0, false},
		{"", 0, false},
	}
	for _, tc := range tests {
		got, ok := parseGroupID(tc.in)
		if ok != tc.wantOK || (ok && got != tc.want) {
			t.Errorf("parseGroupID(%q) = (%d, %v), erwartet (%d, %v)", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestPlaceholderFor(t *testing.T) {
	if got := placeholderFor(runner.Tool{}); got != "Keine weiteren Argumente nötig" {
		t.Errorf("ohne Beispiel = %q", got)
	}
	if got := placeholderFor(runner.Tool{ArgHint: "--summary"}); got != "Zusätzliche Argumente, z. B.: --summary" {
		t.Errorf("mit Beispiel = %q", got)
	}
}
