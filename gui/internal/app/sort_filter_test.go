package app

import (
	"testing"

	"github.com/santifer/career-ops/dashboard/internal/model"
)

// sampleApps bildet einen kleinen, aber realistischen Tracker nach.
func sampleApps() []model.CareerApplication {
	return []model.CareerApplication{
		{Number: 1, Date: "2026-07-27", Company: "ZeroClamp", Role: "IT-Systemadministrator", Score: 4.2, Status: "Interview"},
		{Number: 2, Date: "2026-08-03", Company: "rotis", Role: "System Engineer", Score: 3.6, Status: "Evaluated"},
		{Number: 3, Date: "2026-07-30", Company: "Kathrein", Role: "Senior IT-Systemadministrator", Score: 4.0, Status: "Applied"},
		{Number: 4, Date: "2026-08-01", Company: "TUNAP", Role: "IT Support Specialist", Score: 3.8, Status: "Rejected"},
	}
}

func numbers(apps []model.CareerApplication) []int {
	out := make([]int, len(apps))
	for i, a := range apps {
		out[i] = a.Number
	}
	return out
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSortApps(t *testing.T) {
	tests := []struct {
		name string
		key  string
		desc bool
		want []int
	}{
		{"Nummer aufsteigend", sortByNumber, false, []int{1, 2, 3, 4}},
		{"Nummer absteigend", sortByNumber, true, []int{4, 3, 2, 1}},
		{"Bewertung absteigend", sortByScore, true, []int{1, 3, 4, 2}},
		{"Bewertung aufsteigend", sortByScore, false, []int{2, 4, 3, 1}},
		{"Datum aufsteigend", sortByDate, false, []int{1, 3, 4, 2}},
		{"Datum absteigend", sortByDate, true, []int{2, 4, 3, 1}},
		// Sortierung nach Firma ignoriert Groß-/Kleinschreibung:
		// Kathrein, rotis, TUNAP, ZeroClamp
		{"Firma aufsteigend", sortByComp, false, []int{3, 2, 4, 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			apps := sampleApps()
			sortApps(apps, tc.key, tc.desc)
			if got := numbers(apps); !equalInts(got, tc.want) {
				t.Errorf("Reihenfolge = %v, erwartet %v", got, tc.want)
			}
		})
	}
}

// Nach Status wird nach Handlungsdringlichkeit sortiert, nicht alphabetisch
// und nicht chronologisch: data.StatusPriority stellt laufende Verfahren nach
// vorne (Interview → Angebot → Antwort → Beworben → Bewertet) und abgeschlossene
// nach hinten (Nicht passend → Abgelehnt → Verworfen).
func TestSortAppsByStatusRanksByUrgency(t *testing.T) {
	apps := sampleApps()
	sortApps(apps, sortByStatus, false)
	got := numbers(apps)
	// #1 Interview, #3 Applied, #2 Evaluated, #4 Rejected
	if want := []int{1, 3, 2, 4}; !equalInts(got, want) {
		t.Errorf("Reihenfolge = %v, erwartet %v", got, want)
	}
}

// Absteigend kehrt die Dringlichkeit um: erledigte Fälle zuerst.
func TestSortAppsByStatusDescending(t *testing.T) {
	apps := sampleApps()
	sortApps(apps, sortByStatus, true)
	if got, want := numbers(apps), []int{4, 2, 3, 1}; !equalInts(got, want) {
		t.Errorf("Reihenfolge = %v, erwartet %v", got, want)
	}
}

// Eine unbekannte Sortierart darf keine Zeile verlieren.
func TestSortAppsUnknownKeyFallsBackToNumber(t *testing.T) {
	apps := sampleApps()
	sortApps(apps, "gibt es nicht", false)
	if got := numbers(apps); !equalInts(got, []int{1, 2, 3, 4}) {
		t.Errorf("Reihenfolge = %v, erwartet Nummernfolge", got)
	}
}

func TestFilterApps(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus string
		needle     string
		want       []int
	}{
		{"ohne Filter", "", "", []int{1, 2, 3, 4}},
		{"nur Interview", "interview", "", []int{1}},
		{"nur Abgelehnt", "rejected", "", []int{4}},
		{"Suche nach Firma", "", "zeroclamp", []int{1}},
		{"Suche nach Rolle", "", "support", []int{4}},
		{"Suche ohne Treffer", "", "quantenphysik", []int{}},
		{"Status und Suche kombiniert", "evaluated", "system", []int{2}},
		{"Status und Suche schließen sich aus", "interview", "tunap", []int{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := numbers(filterApps(sampleApps(), tc.wantStatus, tc.needle))
			if !equalInts(got, tc.want) {
				t.Errorf("Ergebnis = %v, erwartet %v", got, tc.want)
			}
		})
	}
}

// Der Filter darf die Reihenfolge der Eingabe nicht durcheinanderbringen.
func TestFilterAppsPreservesOrder(t *testing.T) {
	got := numbers(filterApps(sampleApps(), "", "i"))
	last := 0
	for _, n := range got {
		if n < last {
			t.Fatalf("Reihenfolge zerstört: %v", got)
		}
		last = n
	}
}
