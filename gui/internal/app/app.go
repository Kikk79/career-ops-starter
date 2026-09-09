// Package app baut die Oberfläche des career-ops-GUI zusammen.
//
// Die Daten kommen unverändert aus der Schicht, die auch das Terminal-Dashboard
// verwendet (dashboard/internal/data). Alles, was das Repo an Werkzeugen
// mitbringt, wird über den Runner als Node-Prozess aufgerufen — die Logik der
// .mjs-Skripte wird nicht nachgebaut.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/santifer/career-ops/dashboard/internal/data"
	"github.com/santifer/career-ops/dashboard/internal/model"

	"github.com/santifer/career-ops/dashboard/guiapp/internal/runner"
	"github.com/santifer/career-ops/dashboard/guiapp/internal/ui"
)

// App hält den gesamten Zustand der Oberfläche.
type App struct {
	fyneApp fyne.App
	win     fyne.Window

	// careerOpsPath ist das Wurzelverzeichnis des career-ops-Repos.
	careerOpsPath string
	dark          bool

	// needsSetup merkt sich, dass beim letzten Laden keine Bewerbungsliste
	// vorlag — der Normalfall vor der Ersteinrichtung.
	needsSetup bool

	// Geladene Daten. Werden von reload() gemeinsam ersetzt, nie einzeln
	// verändert, damit die Ansichten immer einen zusammenpassenden Stand sehen.
	apps        []model.CareerApplication
	metrics     model.PipelineMetrics
	progress    model.ProgressMetrics
	statsData   model.StatsMetrics
	pdfManifest data.PDFManifest

	runner *runner.Runner

	pipeline  *pipelineView
	detail    *detailView
	analytics *analyticsView
	tools     *toolsView

	statusBar *widget.Label
}

// New erzeugt die Anwendung für ein career-ops-Verzeichnis.
func New(careerOpsPath string) *App {
	abs, err := filepath.Abs(careerOpsPath)
	if err != nil {
		abs = careerOpsPath
	}
	a := &App{
		fyneApp:       fyneapp.NewWithID("io.careerops.gui"),
		careerOpsPath: abs,
		dark:          true,
		runner:        runner.New(abs),
	}
	a.fyneApp.Settings().SetTheme(ui.NewAppTheme(a.dark))
	a.win = a.fyneApp.NewWindow("Career-Ops")
	// Bewusst moderat: bei 150-%-Anzeigeskalierung entspricht das rund
	// 1770x1110 echten Bildpunkten und passt damit noch auf einen Bildschirm
	// mit 1512 Punkten Höhe samt Taskleiste.
	a.win.Resize(fyne.NewSize(1180, 740))
	return a
}

// Run baut die Oberfläche auf und startet die Ereignisschleife.
func (a *App) Run() {
	a.statusBar = widget.NewLabel("")

	a.pipeline = a.newPipelineView()
	a.detail = a.newDetailView()
	a.analytics = a.newAnalyticsView()
	a.tools = a.newToolsView()

	// Liste links, Detail rechts — dasselbe Aufteilungsprinzip wie im TUI,
	// nur nebeneinander statt hintereinander.
	pipelineSplit := container.NewHSplit(a.pipeline.content, a.detail.content)

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Bewerbungen", theme.ListIcon(), pipelineSplit),
		container.NewTabItemWithIcon("Auswertung", theme.InfoIcon(), a.analytics.content),
		container.NewTabItemWithIcon("Werkzeuge", theme.SettingsIcon(), a.tools.content),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	a.win.SetContent(container.NewBorder(a.buildToolbar(), a.buildStatusBar(), nil, nil, tabs))
	a.win.CenterOnScreen()

	a.reload()

	// Beides braucht ein fertig aufgebautes Fenster: Ein Dialog vor
	// ShowAndRun hätte kein Elternfenster, und ein Teiler-Verhältnis, das vor
	// dem ersten Layout gesetzt wird, überschreibt Fyne wieder mit der
	// Mindestbreite der Kinder.
	go func() {
		time.Sleep(400 * time.Millisecond)
		fyne.Do(func() {
			pipelineSplit.SetOffset(0.66)
			a.tools.split.SetOffset(0.28)
			if a.needsSetup {
				a.showMissingTracker()
			}
		})
	}()
	a.win.ShowAndRun()
}

// buildToolbar baut die obere Leiste: Neu laden, Verzeichnis öffnen, Theme.
func (a *App) buildToolbar() fyne.CanvasObject {
	reload := widget.NewButtonWithIcon("Neu laden", theme.ViewRefreshIcon(), a.reload)

	openDir := widget.NewButtonWithIcon("Ordner öffnen", theme.FolderOpenIcon(), func() {
		if err := openWithDefaultApp(a.careerOpsPath); err != nil {
			a.setStatus("Ordner konnte nicht geöffnet werden: " + err.Error())
		}
	})

	themeBtn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), nil)
	themeBtn.OnTapped = func() {
		a.dark = !a.dark
		a.fyneApp.Settings().SetTheme(ui.NewAppTheme(a.dark))
		a.setStatus(map[bool]string{true: "Dunkles Design", false: "Helles Design"}[a.dark])
	}

	pathLabel := widget.NewLabel(a.careerOpsPath)
	pathLabel.TextStyle = fyne.TextStyle{Monospace: true}

	return container.NewBorder(nil, widget.NewSeparator(), nil,
		container.NewHBox(reload, openDir, themeBtn),
		container.NewHScroll(pathLabel))
}

// buildStatusBar baut die untere Meldungszeile.
func (a *App) buildStatusBar() fyne.CanvasObject {
	return container.NewVBox(widget.NewSeparator(), a.statusBar)
}

// setStatus schreibt eine Meldung in die untere Zeile. Darf nur aus dem
// UI-Thread aufgerufen werden; aus Goroutinen über fyne.Do.
func (a *App) setStatus(msg string) {
	if a.statusBar != nil {
		a.statusBar.SetText(msg)
	}
}

// reload liest den Tracker neu ein und aktualisiert sämtliche Ansichten.
// Bewusst synchron: das Einlesen einer Markdown-Tabelle dauert Millisekunden,
// und ein Nebenläufigkeitsmodell dafür wäre nur zusätzliche Fehlerquelle.
func (a *App) reload() {
	apps := data.ParseApplications(a.careerOpsPath)
	if apps == nil {
		a.needsSetup = true
		a.apps = nil
		a.metrics = model.PipelineMetrics{}
		a.progress = model.ProgressMetrics{}
		a.statsData = model.StatsMetrics{}
		a.pipeline.refresh()
		a.analytics.refresh()
		a.detail.clear()
		a.setStatus("Keine Bewerbungsliste gefunden — siehe Reiter „Werkzeuge“, Eintrag „Einrichtung prüfen“")
		return
	}
	a.needsSetup = false

	// Berichtsdaten je Zeile nachladen: Archetyp und Kurzfassung stehen nicht
	// im Tracker, sondern im zugehörigen Bericht unter reports/.
	for i := range apps {
		if apps[i].ReportPath == "" {
			continue
		}
		archetype, _, _, _ := data.LoadReportSummary(a.careerOpsPath, apps[i].ReportPath)
		if archetype != "" {
			apps[i].Archetype = archetype
		}
	}

	a.apps = apps
	a.metrics = data.ComputeMetrics(apps)
	a.progress = data.ComputeProgressMetrics(apps)
	a.statsData = data.ComputeStatsMetrics(apps)
	a.pdfManifest = data.LoadPDFManifest(a.careerOpsPath)

	a.pipeline.refresh()
	a.analytics.refresh()
	a.detail.clear()

	a.setStatus(fmt.Sprintf("%d Bewerbungen geladen · %d bewertet · Ø %.1f · Bestwert %.1f",
		a.metrics.Total, a.countScored(), a.metrics.AvgScore, a.metrics.TopScore))
}

// showMissingTracker erklärt in einem Dialog, was zu tun ist, wenn noch keine
// Bewerbungsliste existiert — das ist der Normalfall bei einer frischen
// Installation, kein Fehler.
func (a *App) showMissingTracker() {
	msg := fmt.Sprintf(
		"In %s wurde weder applications.md noch data/applications.md gefunden.\n\n"+
			"Das ist bei einer frischen Installation normal. Öffnen Sie den Reiter "+
			"„Werkzeuge“ und starten Sie dort „Einrichtung prüfen“ "+
			"(doctor.mjs) — das Skript sagt Ihnen, welche Dateien noch fehlen.",
		a.careerOpsPath)
	dialog.ShowInformation("Noch nichts einzurichten", msg, a.win)
}

// countScored zählt die Bewerbungen mit einer Bewertung. Der Tracker enthält
// auch nachgetragene Zeilen ohne Bewertung, die den Durchschnitt verfälschen
// würden.
func (a *App) countScored() int {
	n := 0
	for _, app := range a.apps {
		if app.Score > 0 {
			n++
		}
	}
	return n
}

// Fehlerausgabe für den Programmstart.
func Fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
