package app

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/santifer/career-ops/dashboard/guiapp/internal/runner"
)

// maxOutputLines begrenzt den Ausgabepuffer. Manche Scans liefern zehntausende
// Zeilen; ohne Grenze würde die Oberfläche zäh.
const maxOutputLines = 5000

// toolsView listet die Repo-Werkzeuge und führt sie aus.
type toolsView struct {
	app     *App
	content fyne.CanvasObject

	tree     *widget.Tree
	groups   []runner.Group
	selected *runner.Tool

	name    *widget.Label
	desc    *widget.Label
	cmdLine *widget.Label
	warn    *widget.Label

	argsEntry *widget.Entry
	runBtn    *widget.Button
	stopBtn   *widget.Button
	clearBtn  *widget.Button

	output     *widget.TextGrid
	outputWrap *container.Scroll
	lines      []string

	// split wird von der App nach dem ersten Layout auf sein Verhältnis
	// gesetzt; davor überschreibt Fyne den Wert.
	split *container.Split
}

func (a *App) newToolsView() *toolsView {
	t := &toolsView{app: a, groups: runner.Catalog()}

	t.tree = t.buildTree()

	t.name = widget.NewLabel("Kein Werkzeug ausgewählt")
	t.name.TextStyle = fyne.TextStyle{Bold: true}
	t.desc = widget.NewLabel("Wählen Sie links ein Werkzeug aus.")
	t.desc.Wrapping = fyne.TextWrapWord
	t.cmdLine = widget.NewLabel("")
	t.cmdLine.TextStyle = fyne.TextStyle{Monospace: true}
	t.warn = widget.NewLabel("")
	t.warn.Wrapping = fyne.TextWrapWord

	t.argsEntry = widget.NewEntry()
	t.argsEntry.SetPlaceHolder("Zusätzliche Argumente")
	t.argsEntry.OnChanged = func(string) { t.updateCommandLine() }

	t.runBtn = widget.NewButtonWithIcon("Starten", theme.MediaPlayIcon(), t.run)
	t.runBtn.Importance = widget.HighImportance
	t.runBtn.Disable()

	t.stopBtn = widget.NewButtonWithIcon("Abbrechen", theme.MediaStopIcon(), func() {
		t.app.runner.Cancel()
		t.app.setStatus("Abbruch angefordert …")
	})
	t.stopBtn.Disable()

	t.clearBtn = widget.NewButtonWithIcon("Ausgabe leeren", theme.ContentClearIcon(), func() {
		t.lines = nil
		t.output.SetText("")
	})

	t.output = widget.NewTextGrid()
	t.outputWrap = container.NewVScroll(t.output)

	right := container.NewBorder(
		container.NewVBox(
			t.name, t.desc, t.warn,
			widget.NewSeparator(),
			t.argsEntry,
			t.cmdLine,
			container.NewHBox(t.runBtn, t.stopBtn, t.clearBtn),
			widget.NewSeparator(),
		),
		nil, nil, nil,
		t.outputWrap,
	)

	t.split = container.NewHSplit(t.tree, right)
	t.content = t.split
	return t
}

// Knoten-IDs des Werkzeugbaums: "g:<gruppe>" für eine Gruppe,
// "t:<gruppe>:<werkzeug>" für ein Werkzeug.
const (
	groupPrefix = "g:"
	toolPrefix  = "t:"
)

func groupID(gi int) widget.TreeNodeID { return groupPrefix + itoa(gi) }

func toolID(gi, ti int) widget.TreeNodeID { return toolPrefix + itoa(gi) + ":" + itoa(ti) }

// childIDs liefert die Kindknoten. Die leere ID ist die Wurzel des Baums.
func (t *toolsView) childIDs(id widget.TreeNodeID) []widget.TreeNodeID {
	if id == "" {
		ids := make([]widget.TreeNodeID, 0, len(t.groups))
		for gi := range t.groups {
			ids = append(ids, groupID(gi))
		}
		return ids
	}
	gi, ok := parseGroupID(id)
	if !ok || gi >= len(t.groups) {
		return nil
	}
	ids := make([]widget.TreeNodeID, 0, len(t.groups[gi].Tools))
	for ti := range t.groups[gi].Tools {
		ids = append(ids, toolID(gi, ti))
	}
	return ids
}

// isBranch meldet, ob ein Knoten aufklappbar ist. Die Wurzel zählt dazu,
// sonst zeigt der Baum die Gruppen gar nicht erst an.
func (t *toolsView) isBranch(id widget.TreeNodeID) bool {
	return id == "" || strings.HasPrefix(id, groupPrefix)
}

// labelFor liefert den Anzeigetext eines Knotens.
func (t *toolsView) labelFor(id widget.TreeNodeID) string {
	if gi, ok := parseGroupID(id); ok && gi < len(t.groups) {
		return t.groups[gi].Name
	}
	tool := t.toolByID(id)
	if tool == nil {
		return ""
	}
	if !t.app.runner.ScriptExists(*tool) {
		return tool.Name + "  (nicht vorhanden)"
	}
	return tool.Name
}

// parseGroupID zerlegt "g:<n>".
func parseGroupID(id widget.TreeNodeID) (int, bool) {
	rest, found := strings.CutPrefix(string(id), groupPrefix)
	if !found {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// toolByID löst eine Werkzeug-Knoten-ID auf; für Gruppen- und Wurzelknoten nil.
func (t *toolsView) toolByID(id widget.TreeNodeID) *runner.Tool {
	rest, found := strings.CutPrefix(string(id), toolPrefix)
	if !found {
		return nil
	}
	gs, ts, split := strings.Cut(rest, ":")
	if !split {
		return nil
	}
	gi, err1 := strconv.Atoi(gs)
	ti, err2 := strconv.Atoi(ts)
	if err1 != nil || err2 != nil || gi < 0 || ti < 0 {
		return nil
	}
	if gi >= len(t.groups) || ti >= len(t.groups[gi].Tools) {
		return nil
	}
	return &t.groups[gi].Tools[ti]
}

// buildTree baut den nach Themen gruppierten Werkzeugbaum.
func (t *toolsView) buildTree() *widget.Tree {
	tree := widget.NewTree(
		t.childIDs,
		t.isBranch,
		func(branch bool) fyne.CanvasObject {
			l := widget.NewLabel("")
			if branch {
				l.TextStyle = fyne.TextStyle{Bold: true}
			}
			return l
		},
		func(id widget.TreeNodeID, _ bool, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(t.labelFor(id))
		},
	)

	tree.OnSelected = func(id widget.TreeNodeID) {
		if tool := t.toolByID(id); tool != nil {
			t.selectTool(tool)
		}
	}
	for gi := range t.groups {
		tree.OpenBranch(groupID(gi))
	}
	return tree
}

// selectTool zeigt Beschreibung und Befehlszeile eines Werkzeugs an.
func (t *toolsView) selectTool(tool *runner.Tool) {
	t.selected = tool
	t.name.SetText(tool.Name)
	t.desc.SetText(tool.Desc)
	t.argsEntry.SetPlaceHolder(placeholderFor(*tool))
	t.argsEntry.SetText("")

	switch {
	case !t.app.runner.ScriptExists(*tool):
		t.warn.SetText("Achtung: " + tool.Script + " ist in diesem Verzeichnis nicht vorhanden.")
		t.runBtn.Disable()
	case tool.Writes:
		t.warn.SetText("Dieses Werkzeug verändert Dateien im Projekt.")
		t.runBtn.Enable()
	default:
		t.warn.SetText("Dieses Werkzeug liest nur.")
		t.runBtn.Enable()
	}
	if t.app.runner.Busy() {
		t.runBtn.Disable()
	}
	t.updateCommandLine()
}

// placeholderFor beschriftet das Argumentfeld mit dem Beispiel des Werkzeugs.
func placeholderFor(tool runner.Tool) string {
	if tool.ArgHint == "" {
		return "Keine weiteren Argumente nötig"
	}
	return "Zusätzliche Argumente, z. B.: " + tool.ArgHint
}

// updateCommandLine zeigt an, welcher Befehl beim Start ausgeführt würde.
func (t *toolsView) updateCommandLine() {
	if t.selected == nil {
		t.cmdLine.SetText("")
		return
	}
	extra := runner.SplitArgs(t.argsEntry.Text)
	t.cmdLine.SetText("$ " + runner.CommandLine(*t.selected, extra))
}

// run startet das gewählte Werkzeug und streamt seine Ausgabe.
func (t *toolsView) run() {
	if t.selected == nil {
		return
	}
	tool := *t.selected
	extra := runner.SplitArgs(t.argsEntry.Text)

	t.appendLine("$ " + runner.CommandLine(tool, extra))
	t.runBtn.Disable()
	t.stopBtn.Enable()
	t.app.setStatus(tool.Name + " läuft …")

	err := t.app.runner.Start(tool, extra,
		func(line string) {
			fyne.Do(func() { t.appendLine(line) })
		},
		func(code int, runErr error) {
			fyne.Do(func() { t.finished(tool, code, runErr) })
		},
	)
	if err != nil {
		t.runBtn.Enable()
		t.stopBtn.Disable()
		t.appendLine("FEHLER: " + err.Error())
		t.app.setStatus(err.Error())
		dialog.ShowError(err, t.app.win)
	}
}

// finished schließt einen Lauf ab: Ergebnis melden und, wenn das Werkzeug
// schreibt, die Bewerbungsliste neu einlesen.
func (t *toolsView) finished(tool runner.Tool, code int, runErr error) {
	t.stopBtn.Disable()
	if t.selected != nil && t.app.runner.ScriptExists(*t.selected) {
		t.runBtn.Enable()
	}

	switch {
	case runErr != nil:
		t.appendLine("— " + tool.Name + ": " + runErr.Error())
		t.app.setStatus(tool.Name + ": " + runErr.Error())
	case code == 0:
		t.appendLine("— " + tool.Name + " erfolgreich beendet (Exit-Code 0)")
		t.app.setStatus(tool.Name + " erfolgreich beendet")
	default:
		t.appendLine(fmt.Sprintf("— %s mit Exit-Code %d beendet", tool.Name, code))
		t.app.setStatus(fmt.Sprintf("%s: Exit-Code %d", tool.Name, code))
	}

	if tool.Writes && runErr == nil {
		t.app.reload()
	}
}

// appendLine hängt eine Zeile an die Ausgabe und scrollt ans Ende.
func (t *toolsView) appendLine(line string) {
	t.lines = append(t.lines, line)
	if len(t.lines) > maxOutputLines {
		drop := len(t.lines) - maxOutputLines
		t.lines = append([]string{fmt.Sprintf("… %d ältere Zeilen ausgeblendet", drop)},
			t.lines[drop:]...)
	}
	t.output.SetText(strings.Join(t.lines, "\n"))
	t.outputWrap.ScrollToBottom()
}
