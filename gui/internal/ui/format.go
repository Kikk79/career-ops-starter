package ui

import (
	"fmt"
	"image/color"
	"strings"
)

// StatusLabels bildet die kanonischen Status aus templates/states.yml auf
// deutsche Anzeigetexte ab. Geschrieben wird immer der kanonische Wert.
var StatusLabels = map[string]string{
	"Evaluated": "Bewertet",
	"Applied":   "Beworben",
	"Responded": "Antwort erhalten",
	"Interview": "Im Gespräch",
	"Offer":     "Angebot",
	"Hired":     "Eingestellt",
	"Rejected":  "Abgelehnt",
	"Discarded": "Verworfen",
	"SKIP":      "Nicht passend",
}

// CanonicalStatuses ist die Schreibreihenfolge des Status-Auswahlfeldes,
// entlang des Bewerbungsverlaufs.
var CanonicalStatuses = []string{
	"Evaluated", "Applied", "Responded", "Interview",
	"Offer", "Hired", "Rejected", "Discarded", "SKIP",
}

// StatusLabel liefert den deutschen Anzeigetext, oder den Rohwert, wenn der
// Tracker einen unbekannten Status enthält.
func StatusLabel(canonical string) string {
	if l, ok := StatusLabels[canonical]; ok {
		return l
	}
	if canonical == "" {
		return "—"
	}
	return canonical
}

// StatusFromLabel dreht StatusLabel um: vom Anzeigetext zurück zum Wert, der
// in data/applications.md geschrieben wird.
func StatusFromLabel(label string) string {
	for canonical, l := range StatusLabels {
		if l == label {
			return canonical
		}
	}
	return label
}

// StatusColor färbt einen Status nach seiner Stellung im Verlauf ein:
// Erfolg grün, aktiv blau/gelb, Absage rot, verworfen grau.
func StatusColor(status string) color.Color {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "hired":
		return ColGreen
	case "offer":
		return ColTeal
	case "interview":
		return ColBlue
	case "responded":
		return ColMauve
	case "applied":
		return ColYellow
	case "rejected":
		return ColRed
	case "discarded", "skip":
		return ColGray
	default:
		return ColGray
	}
}

// ScoreColor stuft die 0-5-Bewertung farblich ab; ab 4.0 gilt eine Stelle im
// Repo als gut passend (siehe modes/_shared.md).
func ScoreColor(score float64) color.Color {
	switch {
	case score >= 4.5:
		return ColGreen
	case score >= 4.0:
		return ColTeal
	case score >= 3.5:
		return ColYellow
	case score > 0:
		return ColRed
	default:
		return ColGray
	}
}

// FormatScore stellt die Bewertung einheitlich als "4.6" dar; 0 bedeutet
// "nicht bewertet".
func FormatScore(score float64) string {
	if score <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f", score)
}

// FormatPct rundet einen Prozentwert auf eine Nachkommastelle.
func FormatPct(pct float64) string {
	return fmt.Sprintf("%.1f %%", pct)
}

// FormatMoney stellt Dollarbeträge kompakt als "$210K" dar.
func FormatMoney(v float64) string {
	if v <= 0 {
		return "—"
	}
	return fmt.Sprintf("$%.0fK", v/1000)
}

// Bar zeichnet einen Textbalken fester Breite — genutzt in den
// Auswertungstabellen, damit Verhältnisse ohne Diagrammbibliothek lesbar sind.
func Bar(value, max, width int) string {
	if max <= 0 || width <= 0 {
		return ""
	}
	filled := value * width / max
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("·", width-filled)
}

// Truncate kürzt lange Zellen auf eine feste Breite mit Auslassungszeichen.
func Truncate(s string, max int) string {
	r := []rune(s)
	if max <= 1 || len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
