package app

import (
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/santifer/career-ops/dashboard/internal/data"
	"github.com/santifer/career-ops/dashboard/internal/model"

	"github.com/santifer/career-ops/dashboard/guiapp/internal/ui"
)

// Sortierarten der Bewerbungsliste.
const (
	sortByNumber = "Nummer"
	sortByScore  = "Bewertung"
	sortByDate   = "Datum"
	sortByStatus = "Dringlichkeit"
	sortByComp   = "Firma"
)

var sortOptions = []string{sortByNumber, sortByScore, sortByDate, sortByStatus, sortByComp}

// filterAll ist der Eintrag „kein Statusfilter“ im Auswahlfeld.
const filterAll = "Alle Status"

// Spaltenbreiten der Tabelle in Pixeln.
var columnWidths = []float32{52, 96, 168, 260, 78, 132}

var columnTitles = []string{"Nr.", "Datum", "Firma", "Rolle", "Wert", "Status"}

// pipelineView ist die Bewerbungsliste mit Suche, Statusfilter und Sortierung.
type pipelineView struct {
	app     *App
	content fyne.CanvasObject

	table     *widget.Table
	search    *widget.Entry
	filter    *widget.Select
	sortBy    *widget.Select
	sortDesc  bool
	sortBtn   *widget.Button
	countInfo *widget.Label

	// visible sind die nach Suche und Filter übrig gebliebenen Zeilen, in der
	// gerade gewählten Sortierung. Nur diese Liste bestimmt die Tabelle.
	visible []model.CareerApplication
}

func (a *App) newPipelineView() *pipelineView {
	p := &pipelineView{app: a, sortDesc: true}

	p.search = widget.NewEntry()
	p.search.SetPlaceHolder("Suchen in Firma, Rolle, Notizen …")
	p.search.OnChanged = func(string) { p.refresh() }

	p.filter = widget.NewSelect(p.filterOptions(), func(string) { p.refresh() })
	p.filter.SetSelected(filterAll)

	p.sortBy = widget.NewSelect(sortOptions, func(string) { p.refresh() })
	p.sortBy.SetSelected(sortByNumber)

	p.sortBtn = widget.NewButtonWithIcon("Absteigend", theme.MoveDownIcon(), func() {
		p.sortDesc = !p.sortDesc
		if p.sortDesc {
			p.sortBtn.SetText("Absteigend")
			p.sortBtn.SetIcon(theme.MoveDownIcon())
		} else {
			p.sortBtn.SetText("Aufsteigend")
			p.sortBtn.SetIcon(theme.MoveUpIcon())
		}
		p.refresh()
	})

	p.countInfo = widget.NewLabel("")

	clear := widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		p.search.SetText("")
		p.filter.SetSelected(filterAll)
	})

	controls := container.NewVBox(
		container.NewBorder(nil, nil, widget.NewLabel("Suche"), clear, p.search),
		container.NewHBox(
			widget.NewLabel("Status"), p.filter,
			widget.NewSeparator(),
			widget.NewLabel("Sortieren"), p.sortBy, p.sortBtn,
		),
		p.countInfo,
	)

	p.table = p.buildTable()
	p.content = container.NewBorder(controls, nil, nil, nil, p.table)
	return p
}

// filterOptions liefert die Einträge des Statusfilters: „Alle“ plus die
// kanonischen Status in Verlaufsreihenfolge, auf Deutsch beschriftet.
func (p *pipelineView) filterOptions() []string {
	opts := []string{filterAll}
	for _, s := range ui.CanonicalStatuses {
		opts = append(opts, ui.StatusLabel(s))
	}
	return opts
}

// buildTable erzeugt die Tabelle. Jede Zelle ist ein canvas.Text, damit
// Bewertung und Status farbig dargestellt werden können.
func (p *pipelineView) buildTable() *widget.Table {
	t := widget.NewTable(
		func() (int, int) { return len(p.visible), len(columnTitles) },
		func() fyne.CanvasObject {
			txt := canvas.NewText("", theme.Color(theme.ColorNameForeground))
			txt.TextSize = theme.TextSize()
			return container.NewPadded(txt)
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			txt := obj.(*fyne.Container).Objects[0].(*canvas.Text)
			if id.Row < 0 || id.Row >= len(p.visible) {
				txt.Text = ""
				txt.Refresh()
				return
			}
			app := p.visible[id.Row]
			txt.Color = theme.Color(theme.ColorNameForeground)
			txt.TextStyle = fyne.TextStyle{}

			switch id.Col {
			case 0:
				txt.Text = itoa(app.Number)
			case 1:
				txt.Text = app.Date
			case 2:
				txt.Text = ui.Truncate(app.Company, 22)
				txt.TextStyle = fyne.TextStyle{Bold: true}
			case 3:
				txt.Text = ui.Truncate(app.Role, 36)
			case 4:
				txt.Text = ui.FormatScore(app.Score)
				txt.Color = ui.ScoreColor(app.Score)
				txt.TextStyle = fyne.TextStyle{Bold: true}
			case 5:
				txt.Text = ui.StatusLabel(canonicalStatus(app.Status))
				txt.Color = ui.StatusColor(app.Status)
			}
			txt.Refresh()
		},
	)

	t.ShowHeaderRow = true
	t.CreateHeader = func() fyne.CanvasObject {
		l := widget.NewLabel("")
		l.TextStyle = fyne.TextStyle{Bold: true}
		return l
	}
	t.UpdateHeader = func(id widget.TableCellID, obj fyne.CanvasObject) {
		l := obj.(*widget.Label)
		if id.Col >= 0 && id.Col < len(columnTitles) {
			l.SetText(columnTitles[id.Col])
		}
	}
	for i, w := range columnWidths {
		t.SetColumnWidth(i, w)
	}

	t.OnSelected = func(id widget.TableCellID) {
		if id.Row >= 0 && id.Row < len(p.visible) {
			p.app.detail.show(p.visible[id.Row])
		}
	}
	return t
}

// refresh wendet Suche, Filter und Sortierung neu an und zeichnet die Tabelle.
func (p *pipelineView) refresh() {
	needle := strings.ToLower(strings.TrimSpace(p.search.Text))
	wantStatus := ""
	if p.filter != nil && p.filter.Selected != filterAll && p.filter.Selected != "" {
		wantStatus = strings.ToLower(ui.StatusFromLabel(p.filter.Selected))
	}

	p.visible = filterApps(p.app.apps, wantStatus, needle)
	p.sortVisible()

	if p.countInfo != nil {
		if len(p.visible) == len(p.app.apps) {
			p.countInfo.SetText(itoa(len(p.visible)) + " Bewerbungen")
		} else {
			p.countInfo.SetText(itoa(len(p.visible)) + " von " + itoa(len(p.app.apps)) + " Bewerbungen")
		}
	}
	if p.table != nil {
		p.table.UnselectAll()
		p.table.Refresh()
	}
}

// sortVisible ordnet die sichtbaren Zeilen nach dem gewählten Kriterium.
func (p *pipelineView) sortVisible() {
	key := sortByNumber
	if p.sortBy != nil && p.sortBy.Selected != "" {
		key = p.sortBy.Selected
	}
	sortApps(p.visible, key, p.sortDesc)
}

// sortApps sortiert an Ort und Stelle nach dem angegebenen Kriterium. Stabil,
// damit gleichrangige Zeilen ihre Reihenfolge aus dem Tracker behalten.
func sortApps(apps []model.CareerApplication, key string, desc bool) {
	less := func(i, j int) bool {
		a, b := apps[i], apps[j]
		switch key {
		case sortByScore:
			return a.Score < b.Score
		case sortByDate:
			return a.Date < b.Date
		case sortByStatus:
			// Nach Handlungsdringlichkeit, nicht alphabetisch: laufende
			// Verfahren nach vorne, abgeschlossene nach hinten.
			return data.StatusPriority(a.Status) < data.StatusPriority(b.Status)
		case sortByComp:
			return strings.ToLower(a.Company) < strings.ToLower(b.Company)
		default:
			return a.Number < b.Number
		}
	}
	sort.SliceStable(apps, func(i, j int) bool {
		if desc {
			return less(j, i)
		}
		return less(i, j)
	})
}

// filterApps wendet Statusfilter und Suchbegriff an. wantStatus ist der
// kanonische Status in Kleinschreibung oder "" für "alle"; needle ist der
// bereits kleingeschriebene Suchbegriff oder "".
func filterApps(apps []model.CareerApplication, wantStatus, needle string) []model.CareerApplication {
	out := make([]model.CareerApplication, 0, len(apps))
	for _, app := range apps {
		if wantStatus != "" && data.NormalizeStatus(app.Status) != wantStatus {
			continue
		}
		if needle != "" && !matches(app, needle) {
			continue
		}
		out = append(out, app)
	}
	return out
}

// matches prüft eine Zeile gegen den kleingeschriebenen Suchbegriff.
func matches(app model.CareerApplication, needle string) bool {
	for _, field := range []string{app.Company, app.Role, app.Notes, app.Archetype, app.Location} {
		if strings.Contains(strings.ToLower(field), needle) {
			return true
		}
	}
	return false
}

// canonicalStatus bringt einen Trackerwert auf die kanonische Schreibweise aus
// templates/states.yml zurück, damit die deutsche Beschriftung greift.
func canonicalStatus(raw string) string {
	normalized := data.NormalizeStatus(raw)
	for _, s := range ui.CanonicalStatuses {
		if strings.ToLower(s) == normalized {
			return s
		}
	}
	return strings.TrimSpace(raw)
}

// itoa vermeidet den strconv-Import in jeder Datei.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
