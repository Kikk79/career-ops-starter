package ui

import "testing"

// Jeder kanonische Status aus templates/states.yml braucht eine deutsche
// Beschriftung, die sich verlustfrei zurückübersetzen lässt — sonst schreibt
// die Oberfläche einen Wert in den Tracker, den das Repo nicht kennt.
func TestStatusLabelRoundTrip(t *testing.T) {
	for _, canonical := range CanonicalStatuses {
		label := StatusLabel(canonical)
		if label == canonical {
			t.Errorf("%s hat keine deutsche Beschriftung", canonical)
		}
		if got := StatusFromLabel(label); got != canonical {
			t.Errorf("StatusFromLabel(%q) = %q, erwartet %q", label, got, canonical)
		}
	}
}

// Ein unbekannter Trackerwert darf nicht verschluckt werden.
func TestStatusLabelUnknownPassesThrough(t *testing.T) {
	if got := StatusLabel("Sonstiges"); got != "Sonstiges" {
		t.Errorf("StatusLabel = %q, erwartet Durchreichen", got)
	}
	if got := StatusLabel(""); got != "—" {
		t.Errorf("leerer Status = %q, erwartet Gedankenstrich", got)
	}
}

func TestFormatScore(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "—"},
		{-1, "—"},
		{4.6, "4.6"},
		{4.26, "4.3"},
		{3, "3.0"},
	}
	for _, tc := range tests {
		if got := FormatScore(tc.in); got != tc.want {
			t.Errorf("FormatScore(%v) = %q, erwartet %q", tc.in, got, tc.want)
		}
	}
}

func TestBar(t *testing.T) {
	tests := []struct {
		name              string
		value, max, width int
		wantLen           int
	}{
		{"voll", 10, 10, 10, 10},
		{"halb", 5, 10, 10, 10},
		{"null", 0, 10, 10, 10},
		{"max null", 5, 0, 10, 0},
		{"breite null", 5, 10, 0, 0},
		{"über max", 20, 10, 10, 10},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := len([]rune(Bar(tc.value, tc.max, tc.width))); got != tc.wantLen {
				t.Errorf("Bar-Länge = %d, erwartet %d", got, tc.wantLen)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		in   string
		max  int
		want string
	}{
		{"kurz", 10, "kurz"},
		{"genau zehn", 10, "genau zehn"},
		{"viel zu langer Text", 8, "viel zu…"},
		{"Müßiggang", 5, "Müßi…"}, // zählt Zeichen, nicht Bytes
	}
	for _, tc := range tests {
		if got := Truncate(tc.in, tc.max); got != tc.want {
			t.Errorf("Truncate(%q, %d) = %q, erwartet %q", tc.in, tc.max, got, tc.want)
		}
	}
}

func TestFormatMoney(t *testing.T) {
	if got := FormatMoney(0); got != "—" {
		t.Errorf("FormatMoney(0) = %q", got)
	}
	if got := FormatMoney(210000); got != "$210K" {
		t.Errorf("FormatMoney(210000) = %q, erwartet $210K", got)
	}
}
