package app

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/santifer/career-ops/dashboard/internal/data"
	"github.com/santifer/career-ops/dashboard/internal/model"

	"github.com/santifer/career-ops/dashboard/guiapp/internal/ui"
)

// barWidth ist die Zeichenbreite der Textbalken in den Auswertungstabellen.
const barWidth = 24

// analyticsView zeigt Trichter, Bewertungsverteilung, Quoten, Wochenverlauf und
// die Aufschlüsselungen aus der Statistikschicht des Terminal-Dashboards.
type analyticsView struct {
	app     *App
	content fyne.CanvasObject

	kpis     *fyne.Container
	funnel   *widget.TextGrid
	scores   *widget.TextGrid
	weeks    *widget.TextGrid
	types    *widget.TextGrid
	places   *widget.TextGrid
	pay      *widget.TextGrid
	insights *widget.Label
}

func (a *App) newAnalyticsView() *analyticsView {
	v := &analyticsView{app: a}

	v.kpis = container.NewGridWithColumns(4)
	v.funnel = widget.NewTextGrid()
	v.scores = widget.NewTextGrid()
	v.weeks = widget.NewTextGrid()
	v.types = widget.NewTextGrid()
	v.places = widget.NewTextGrid()
	v.pay = widget.NewTextGrid()
	v.insights = widget.NewLabel("")
	v.insights.Wrapping = fyne.TextWrapWord

	body := container.NewVBox(
		v.kpis,
		widget.NewSeparator(),
		section("Bewerbungstrichter", v.funnel),
		section("Bewertungsverteilung", v.scores),
		section("Aktivität je Kalenderwoche", v.weeks),
		section("Rollentypen", v.types),
		section("Standorte und Arbeitsmodell", v.places),
		section("Gehälter", v.pay),
		section("Beobachtungen", v.insights),
	)

	v.content = container.NewVScroll(body)
	return v
}

// section rahmt einen Auswertungsblock mit Überschrift.
func section(title string, body fyne.CanvasObject) fyne.CanvasObject {
	head := widget.NewLabel(title)
	head.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewVBox(head, body, widget.NewSeparator())
}

// refresh zeichnet alle Auswertungen aus dem aktuellen Datenstand neu.
func (v *analyticsView) refresh() {
	m, p, s := v.app.metrics, v.app.progress, v.app.statsData

	v.kpis.RemoveAll()
	v.kpis.Add(kpi("Bewerbungen gesamt", itoa(m.Total)))
	v.kpis.Add(kpi("Aktiv im Verfahren", itoa(p.ActiveApps)))
	v.kpis.Add(kpi("Durchschnittswert", ui.FormatScore(m.AvgScore)))
	v.kpis.Add(kpi("Bestwert", ui.FormatScore(m.TopScore)))
	v.kpis.Add(kpi("Antwortquote", ui.FormatPct(p.ResponseRate)))
	v.kpis.Add(kpi("Gesprächsquote", ui.FormatPct(p.InterviewRate)))
	v.kpis.Add(kpi("Angebotsquote", ui.FormatPct(p.OfferRate)))
	v.kpis.Add(kpi("Mit PDF", itoa(m.WithPDF)))
	v.kpis.Refresh()

	v.funnel.SetText(renderFunnel(p.FunnelStages))
	v.scores.SetText(renderBuckets(p.ScoreBuckets))
	v.weeks.SetText(renderWeeks(p.WeeklyActivity))
	v.types.SetText(renderArchetypes(s.Archetypes))
	v.places.SetText(renderPlaces(s))
	v.pay.SetText(renderPay(s))

	if len(v.app.apps) == 0 {
		v.insights.SetText("Noch keine Daten. Sobald Bewerbungen in data/applications.md stehen, erscheinen hier Auswertungen.")
		return
	}
	lines := data.GenerateInsights(s)
	if len(lines) == 0 {
		v.insights.SetText("Keine auffälligen Muster im aktuellen Bestand.")
		return
	}
	v.insights.SetText("• " + strings.Join(lines, "\n• "))
}

// kpi baut eine Kennzahlkachel.
func kpi(title, value string) fyne.CanvasObject {
	t := widget.NewLabel(title)
	val := widget.NewLabel(value)
	val.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewVBox(t, val)
}

func renderFunnel(stages []model.FunnelStage) string {
	if len(stages) == 0 {
		return "  keine Daten"
	}
	max := 0
	for _, s := range stages {
		if s.Count > max {
			max = s.Count
		}
	}
	var sb strings.Builder
	for _, s := range stages {
		fmt.Fprintf(&sb, "  %-22s %5d  %s  %5.1f %%\n",
			ui.Truncate(funnelLabel(s.Label), 22), s.Count, ui.Bar(s.Count, max, barWidth), s.Pct)
	}
	return sb.String()
}

// funnelLabel übersetzt die englischen Stufennamen der Datenschicht.
func funnelLabel(raw string) string {
	if l, ok := ui.StatusLabels[strings.TrimSpace(raw)]; ok {
		return l
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "total", "all", "tracked":
		return "Erfasst"
	case "evaluated":
		return "Bewertet"
	case "applied":
		return "Beworben"
	case "responded":
		return "Antwort erhalten"
	case "interview":
		return "Im Gespräch"
	case "offer":
		return "Angebot"
	case "hired":
		return "Eingestellt"
	default:
		return raw
	}
}

func renderBuckets(buckets []model.ScoreBucket) string {
	if len(buckets) == 0 {
		return "  keine Daten"
	}
	max := 0
	for _, b := range buckets {
		if b.Count > max {
			max = b.Count
		}
	}
	var sb strings.Builder
	for _, b := range buckets {
		fmt.Fprintf(&sb, "  %-10s %5d  %s\n", b.Label, b.Count, ui.Bar(b.Count, max, barWidth))
	}
	return sb.String()
}

func renderWeeks(weeks []model.WeekActivity) string {
	if len(weeks) == 0 {
		return "  keine Daten"
	}
	max := 0
	for _, w := range weeks {
		if w.Count > max {
			max = w.Count
		}
	}
	var sb strings.Builder
	for _, w := range weeks {
		fmt.Fprintf(&sb, "  %-10s %5d  %s\n", w.Week, w.Count, ui.Bar(w.Count, max, barWidth))
	}
	return sb.String()
}

func renderArchetypes(rows []model.ArchetypeStat) string {
	if len(rows) == 0 {
		return "  keine Daten — Rollentypen stammen aus den Berichten unter reports/"
	}
	max := 0
	for _, r := range rows {
		if r.Count > max {
			max = r.Count
		}
	}
	var sb strings.Builder
	sb.WriteString("  Typ                        Anzahl  Anteil   Ø Wert\n")
	for _, r := range rows {
		fmt.Fprintf(&sb, "  %-26s %5d  %5.1f %%   %s  %s\n",
			ui.Truncate(r.Label, 26), r.Count, r.Pct, ui.FormatScore(r.AvgScore), ui.Bar(r.Count, max, 12))
	}
	return sb.String()
}

func renderPlaces(s model.StatsMetrics) string {
	var sb strings.Builder
	sb.WriteString("  Arbeitsmodell\n")
	if len(s.WorkModes) == 0 {
		sb.WriteString("    keine Daten\n")
	}
	for _, r := range s.WorkModes {
		fmt.Fprintf(&sb, "    %-24s %5d  %5.1f %%\n", ui.Truncate(workModeLabel(r.Label), 24), r.Count, r.Pct)
	}
	sb.WriteString("\n  Standorte\n")
	if len(s.Locations) == 0 {
		sb.WriteString("    keine Daten\n")
	}
	for _, r := range s.Locations {
		fmt.Fprintf(&sb, "    %-24s %5d  %5.1f %%\n", ui.Truncate(r.Label, 24), r.Count, r.Pct)
	}
	return sb.String()
}

// workModeLabel übersetzt die Arbeitsmodell-Werte der Datenschicht.
func workModeLabel(raw string) string {
	switch strings.TrimSpace(raw) {
	case "Remote":
		return "Vollständig remote"
	case "RemoteFlex":
		return "Remote mit Präsenzanteil"
	case "Hybrid":
		return "Hybrid"
	case "Full":
		return "Vor Ort"
	case "":
		return "Unbekannt"
	default:
		return raw
	}
}

func renderPay(s model.StatsMetrics) string {
	p := s.Pay
	if p.Count == 0 {
		return "  Keine Gehaltsangaben in den Notizen gefunden."
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "  Angaben insgesamt   %d  (davon %d aus der Anzeige, %d geschätzt)\n",
		p.Count, p.PostedCount, p.EstCount)
	fmt.Fprintf(&sb, "  Durchschnitt        %s\n", ui.FormatMoney(p.AvgPayMax))
	fmt.Fprintf(&sb, "  Median              %s\n", ui.FormatMoney(p.MedianPayMax))
	fmt.Fprintf(&sb, "  Höchstwert          %s\n", ui.FormatMoney(p.MaxPayMax))
	if len(s.PayHistogram) > 0 {
		max := 0
		for _, r := range s.PayHistogram {
			if r.Count > max {
				max = r.Count
			}
		}
		sb.WriteString("\n  Verteilung\n")
		for _, r := range s.PayHistogram {
			fmt.Fprintf(&sb, "    %-16s %4d  %s\n", r.Label, r.Count, ui.Bar(r.Count, max, 18))
		}
	}
	return sb.String()
}
