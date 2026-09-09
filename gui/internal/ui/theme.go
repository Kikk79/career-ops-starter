// Package ui bündelt Darstellung: Farbschema und Formatierungshelfer.
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Farben orientieren sich an Catppuccin, dem Farbschema des TUI unter
// dashboard/internal/theme, damit GUI und TUI gleich aussehen.
var (
	ColGreen  = color.NRGBA{R: 0x40, G: 0xa0, B: 0x2b, A: 0xff}
	ColBlue   = color.NRGBA{R: 0x1e, G: 0x66, B: 0xf5, A: 0xff}
	ColYellow = color.NRGBA{R: 0xdf, G: 0x8e, B: 0x1d, A: 0xff}
	ColRed    = color.NRGBA{R: 0xd2, G: 0x0f, B: 0x39, A: 0xff}
	ColMauve  = color.NRGBA{R: 0x88, G: 0x39, B: 0xef, A: 0xff}
	ColTeal   = color.NRGBA{R: 0x17, G: 0x92, B: 0x99, A: 0xff}
	ColGray   = color.NRGBA{R: 0x8c, G: 0x8f, B: 0xa1, A: 0xff}
)

// AppTheme ist das Standard-Fyne-Theme mit fest gewählter Hell-/Dunkelvariante,
// damit die Umschaltung im GUI unabhängig von der Systemeinstellung wirkt.
type AppTheme struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

// NewAppTheme liefert ein Theme in der gewünschten Variante.
func NewAppTheme(dark bool) *AppTheme {
	v := theme.VariantLight
	if dark {
		v = theme.VariantDark
	}
	return &AppTheme{Theme: theme.DefaultTheme(), variant: v}
}

// Color erzwingt die gewählte Variante statt der Systemvariante.
func (t *AppTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(n, t.variant)
}
