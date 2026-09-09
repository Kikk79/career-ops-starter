package app

import (
	"testing"

	"github.com/santifer/career-ops/dashboard/internal/model"
)

func TestCanonicalStatus(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"bereits kanonisch", "Applied", "Applied"},
		{"kleingeschrieben", "applied", "Applied"},
		{"spanischer Alias", "aplicado", "Applied"},
		{"türkischer Alias", "mülakat", "Interview"},
		{"SKIP", "no_aplicar", "SKIP"},
		{"unbekannt bleibt erhalten", "Wartet auf Rückruf", "Wartet auf Rückruf"},
		{"leer", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := canonicalStatus(tc.in); got != tc.want {
				t.Errorf("canonicalStatus(%q) = %q, erwartet %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	app := model.CareerApplication{
		Company:   "Acme AI",
		Role:      "Head of Applied AI",
		Notes:     "über Portal beworben",
		Archetype: "Plattform",
		Location:  "Berlin, DE",
	}
	tests := []struct {
		name   string
		needle string
		want   bool
	}{
		{"Firma", "acme", true},
		{"Rolle", "head", true},
		{"Notiz", "portal", true},
		{"Archetyp", "plattform", true},
		{"Ort", "berlin", true},
		{"kein Treffer", "globex", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := matches(app, tc.needle); got != tc.want {
				t.Errorf("matches(%q) = %v, erwartet %v", tc.needle, got, tc.want)
			}
		})
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{{0, "0"}, {7, "7"}, {42, "42"}, {1234, "1234"}, {-9, "-9"}}
	for _, tc := range tests {
		if got := itoa(tc.in); got != tc.want {
			t.Errorf("itoa(%d) = %q, erwartet %q", tc.in, got, tc.want)
		}
	}
}

func TestNonEmpty(t *testing.T) {
	got := nonEmpty([]string{"a", "", "b", "", ""})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("nonEmpty = %#v, erwartet [a b]", got)
	}
}

func TestLabelled(t *testing.T) {
	if got := labelled("Ort", "Berlin"); got != "Ort: Berlin" {
		t.Errorf("labelled = %q", got)
	}
	for _, empty := range []string{"", "   ", "—"} {
		if got := labelled("Ort", empty); got != "" {
			t.Errorf("labelled(%q) = %q, erwartet leer", empty, got)
		}
	}
}
