package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/santifer/career-ops/dashboard/internal/data"
	"github.com/santifer/career-ops/dashboard/internal/model"

	"github.com/santifer/career-ops/dashboard/guiapp/internal/ui"
)

// detailView zeigt eine ausgewählte Bewerbung: Kopfdaten, Statuspflege,
// Aktionen und den vollständigen Bewertungsbericht.
type detailView struct {
	app     *App
	content fyne.CanvasObject

	// current ist die gerade angezeigte Bewerbung; Number 0 heißt „keine“.
	current model.CareerApplication

	title    *widget.Label
	subtitle *widget.Label
	facts    *widget.Label

	statusSelect *widget.Select
	noteEntry    *widget.Entry
	saveBtn      *widget.Button

	openURLBtn *widget.Button
	openPDFBtn *widget.Button
	makePDFBtn *widget.Button
	reportBtn  *widget.Button

	report     *widget.RichText
	reportWrap *container.Scroll

	// suppressStatusEvent verhindert, dass das programmgesteuerte Setzen des
	// Auswahlfeldes beim Anzeigen einer Zeile als Benutzeraktion gilt.
	suppressStatusEvent bool
}

func (a *App) newDetailView() *detailView {
	d := &detailView{app: a}

	d.title = widget.NewLabel("")
	d.title.TextStyle = fyne.TextStyle{Bold: true}
	d.subtitle = widget.NewLabel("")
	d.facts = widget.NewLabel("")
	d.facts.Wrapping = fyne.TextWrapWord

	d.statusSelect = widget.NewSelect(statusLabels(), func(sel string) {
		if d.suppressStatusEvent || d.current.Number == 0 {
			return
		}
		d.saveBtn.Enable()
		_ = sel
	})

	d.noteEntry = widget.NewEntry()
	d.noteEntry.SetPlaceHolder("Notiz zum Statuswechsel (optional)")

	d.saveBtn = widget.NewButtonWithIcon("Status speichern", theme.DocumentSaveIcon(), d.saveStatus)
	d.saveBtn.Importance = widget.HighImportance
	d.saveBtn.Disable()

	d.openURLBtn = widget.NewButtonWithIcon("Anzeige öffnen", theme.SearchIcon(), func() {
		d.open(d.current.JobURL, "Zu dieser Bewerbung ist keine Stellen-URL hinterlegt.")
	})
	d.openPDFBtn = widget.NewButtonWithIcon("PDF öffnen", theme.DocumentIcon(), d.openPDF)
	d.makePDFBtn = widget.NewButtonWithIcon("PDF erzeugen", theme.MediaPlayIcon(), d.generatePDF)
	d.reportBtn = widget.NewButtonWithIcon("Bericht im Editor", theme.FileTextIcon(), func() {
		if d.current.ReportPath == "" {
			d.info("Zu dieser Bewerbung gibt es keinen Bericht.")
			return
		}
		d.open(filepath.Join(d.app.careerOpsPath, filepath.FromSlash(d.current.ReportPath)), "")
	})

	d.report = widget.NewRichTextFromMarkdown("")
	d.report.Wrapping = fyne.TextWrapWord
	d.reportWrap = container.NewVScroll(d.report)

	header := container.NewVBox(
		d.title,
		d.subtitle,
		d.facts,
		widget.NewSeparator(),
		container.NewHBox(widget.NewLabel("Status"), d.statusSelect, d.saveBtn),
		d.noteEntry,
		container.NewGridWithColumns(2,
			d.openURLBtn, d.openPDFBtn,
			d.makePDFBtn, d.reportBtn,
		),
		widget.NewSeparator(),
	)

	d.content = container.NewBorder(header, nil, nil, nil, d.reportWrap)
	d.clear()
	return d
}

// statusLabels sind die deutschen Beschriftungen der kanonischen Status.
func statusLabels() []string {
	out := make([]string, 0, len(ui.CanonicalStatuses))
	for _, s := range ui.CanonicalStatuses {
		out = append(out, ui.StatusLabel(s))
	}
	return out
}

// clear setzt die Detailansicht auf „nichts ausgewählt“ zurück.
func (d *detailView) clear() {
	d.current = model.CareerApplication{}
	d.title.SetText("Keine Bewerbung ausgewählt")
	d.subtitle.SetText("Wählen Sie links eine Zeile aus.")
	d.facts.SetText("")
	d.report.ParseMarkdown("")
	d.setActionsEnabled(false)
}

// setActionsEnabled schaltet alle Bedienelemente gemeinsam.
func (d *detailView) setActionsEnabled(on bool) {
	widgets := []interface{ Enable() }{d.statusSelect, d.openURLBtn, d.openPDFBtn, d.makePDFBtn, d.reportBtn}
	if on {
		for _, w := range widgets {
			w.Enable()
		}
		return
	}
	for _, w := range []interface{ Disable() }{d.statusSelect, d.openURLBtn, d.openPDFBtn, d.makePDFBtn, d.reportBtn, d.saveBtn} {
		w.Disable()
	}
}

// show stellt eine Bewerbung dar und lädt ihren Bericht nach.
func (d *detailView) show(app model.CareerApplication) {
	d.current = app
	d.setActionsEnabled(true)
	d.saveBtn.Disable()
	d.noteEntry.SetText("")

	d.title.SetText(fmt.Sprintf("#%d · %s", app.Number, app.Company))
	d.subtitle.SetText(app.Role)

	d.suppressStatusEvent = true
	d.statusSelect.SetSelected(ui.StatusLabel(canonicalStatus(app.Status)))
	d.suppressStatusEvent = false

	d.facts.SetText(strings.Join(nonEmpty([]string{
		labelled("Bewertung", ui.FormatScore(app.Score)),
		labelled("Beworben am", app.Date),
		labelled("Archetyp", app.Archetype),
		labelled("Ort", app.Location),
		labelled("Arbeitsmodell", app.WorkMode),
		labelled("Gehaltsrahmen", app.PayRange),
		labelled("Letzter Kontakt", app.LastContact),
	}), "   ·   "))

	d.report.ParseMarkdown(d.loadReport(app))
	d.reportWrap.ScrollToTop()
}

// loadReport liest den Bewertungsbericht. Fehlt er, wird stattdessen erklärt,
// was vorliegt — die Notizen des Trackers sind dann die einzige Information.
func (d *detailView) loadReport(app model.CareerApplication) string {
	if app.ReportPath != "" {
		full := filepath.Join(d.app.careerOpsPath, filepath.FromSlash(app.ReportPath))
		if b, err := os.ReadFile(full); err == nil {
			return string(b)
		}
	}
	var sb strings.Builder
	sb.WriteString("## Kein Bewertungsbericht vorhanden\n\n")
	if app.ReportPath != "" {
		sb.WriteString("Der Tracker verweist auf `" + app.ReportPath + "`, die Datei fehlt aber.\n\n")
	}
	if strings.TrimSpace(app.Notes) != "" {
		sb.WriteString("### Notizen aus dem Tracker\n\n" + app.Notes + "\n")
	}
	return sb.String()
}

// saveStatus schreibt den gewählten Status in data/applications.md. Der
// Schreibvorgang läuft über die Datenschicht des Terminal-Dashboards, die
// dieselbe Sperrdatei benutzt wie set-status.mjs.
func (d *detailView) saveStatus() {
	if d.current.Number == 0 {
		return
	}
	newStatus := ui.StatusFromLabel(d.statusSelect.Selected)
	note := strings.TrimSpace(d.noteEntry.Text)

	var err error
	if note != "" {
		err = data.UpdateApplicationStatusAndNotes(d.app.careerOpsPath, d.current, newStatus, note)
	} else {
		err = data.UpdateApplicationStatus(d.app.careerOpsPath, d.current, newStatus)
	}
	if err != nil {
		dialog.ShowError(fmt.Errorf("Status konnte nicht gespeichert werden: %w", err), d.app.win)
		d.app.setStatus("Statusänderung fehlgeschlagen")
		return
	}

	d.app.setStatus(fmt.Sprintf("#%d %s → %s gespeichert",
		d.current.Number, d.current.Company, ui.StatusLabel(newStatus)))
	number := d.current.Number
	d.app.reload()
	d.reselect(number)
}

// reselect stellt nach einem Neuladen wieder dieselbe Bewerbung dar.
func (d *detailView) reselect(number int) {
	for _, app := range d.app.apps {
		if app.Number == number {
			d.show(app)
			return
		}
	}
}

// openPDF öffnet die erzeugte Lebenslauf-PDF der Bewerbung. Gibt es mehrere
// Fassungen, wird die neueste genommen.
func (d *detailView) openPDF() {
	pdfs := data.ResolvePDFs(d.app.careerOpsPath, d.current, d.app.pdfManifest)
	if len(pdfs) == 0 {
		d.info("Für diese Bewerbung wurde noch keine PDF erzeugt. Über „PDF erzeugen“ können Sie sie anlegen.")
		return
	}
	d.open(pdfs[0], "")
}

// generatePDF ruft generate-pdf.mjs für die HTML-Fassung dieser Bewerbung auf
// und öffnet das Ergebnis. Findet sich keine Fassung automatisch, wird sie zur
// Auswahl gestellt, statt die Aktion kommentarlos verpuffen zu lassen.
func (d *detailView) generatePDF() {
	htmlPath, pdfPath, found := resolveCVHTML(d.app.careerOpsPath, d.current)
	if !found {
		d.chooseCVHTML()
		return
	}
	d.runGeneratePDF(d.current, htmlPath, pdfPath)
}

// chooseCVHTML lässt den Nutzer die zugeschnittene Lebenslauf-Fassung selbst
// wählen. Nötig, wenn der Dateiname nicht zum Firmennamen im Tracker passt —
// etwa bei einer Bewerbung über eine Agentur, wo die Firma als „?“ geführt wird.
func (d *detailView) chooseCVHTML() {
	candidates := listCVHTML(d.app.careerOpsPath)
	if len(candidates) == 0 {
		d.info("Unter output/ liegt noch keine zugeschnittene HTML-Fassung des " +
			"Lebenslaufs.\n\nErzeugen Sie zuerst eine — im Reiter „Werkzeuge“ über " +
			"„Lebenslauf als HTML bauen“.")
		return
	}

	app := d.current
	sel := widget.NewSelect(candidates, nil)
	sel.SetSelected(candidates[0])

	body := container.NewVBox(
		widget.NewLabel(fmt.Sprintf(
			"Zu „%s“ ließ sich keine Lebenslauf-Fassung eindeutig zuordnen.\n"+
				"Welche soll für die PDF verwendet werden?", app.Company)),
		sel,
	)
	dialog.ShowCustomConfirm("Lebenslauf-Fassung wählen", "PDF erzeugen", "Abbrechen", body,
		func(ok bool) {
			if !ok || sel.Selected == "" {
				return
			}
			html := sel.Selected
			d.runGeneratePDF(app, html, strings.TrimSuffix(html, ".html")+".pdf")
		}, d.app.win)
}

// runGeneratePDF startet generate-pdf.mjs in einer Goroutine, damit die
// Oberfläche während des Renderns bedienbar bleibt.
func (d *detailView) runGeneratePDF(app model.CareerApplication, htmlPath, pdfPath string) {
	d.app.setStatus("PDF wird erzeugt …")
	d.makePDFBtn.Disable()

	go func() {
		args := []string{"generate-pdf.mjs", htmlPath, pdfPath}
		if app.ReportNumber != "" {
			args = append(args, "--report="+app.ReportNumber)
		}
		cmd := exec.Command("node", args...)
		cmd.Dir = d.app.careerOpsPath
		out, err := cmd.CombinedOutput()

		fyne.Do(func() {
			d.makePDFBtn.Enable()
			if err != nil {
				d.app.setStatus("PDF-Erzeugung fehlgeschlagen")
				dialog.ShowError(fmt.Errorf("generate-pdf.mjs ist fehlgeschlagen:\n\n%s", lastLine(out, err)), d.app.win)
				return
			}
			abs := filepath.Join(d.app.careerOpsPath, filepath.FromSlash(pdfPath))
			d.app.setStatus("PDF erzeugt: " + pdfPath)
			if openErr := openWithDefaultApp(abs); openErr != nil {
				d.app.setStatus("PDF erzeugt, konnte aber nicht geöffnet werden: " + openErr.Error())
			}
		})
	}()
}

// open öffnet Pfad oder URL mit dem Standardprogramm des Betriebssystems.
// Ist target leer, wird emptyMsg angezeigt.
func (d *detailView) open(target, emptyMsg string) {
	if strings.TrimSpace(target) == "" {
		if emptyMsg != "" {
			d.info(emptyMsg)
		}
		return
	}
	if err := openWithDefaultApp(target); err != nil {
		dialog.ShowError(fmt.Errorf("konnte nicht geöffnet werden: %w", err), d.app.win)
	}
}

func (d *detailView) info(msg string) {
	dialog.ShowInformation("Hinweis", msg, d.app.win)
}

// labelled formatiert ein Feld als "Titel: Wert" und liefert "" für Leerwerte,
// damit nonEmpty sie herausfiltern kann.
func labelled(title, value string) string {
	v := strings.TrimSpace(value)
	if v == "" || v == "—" {
		return ""
	}
	return title + ": " + v
}

func nonEmpty(in []string) []string {
	out := in[:0]
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// lastLine zieht die letzte nicht leere Ausgabezeile heraus — dort schreiben
// die Node-Skripte ihre eigentliche Fehlermeldung hin.
func lastLine(out []byte, fallback error) string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if l := strings.TrimSpace(lines[i]); l != "" {
			return l
		}
	}
	return fallback.Error()
}
