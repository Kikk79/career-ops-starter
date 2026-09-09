// Command career-ops-gui ist die grafische Oberfläche für career-ops.
//
// Sie zeigt dieselben Daten wie das Terminal-Dashboard unter dashboard/ — die
// Datenschicht wird unverändert weiterverwendet — und stellt zusätzlich die
// Kommandozeilen-Werkzeuge des Repos als bedienbare Liste bereit.
//
// Aufruf:
//
//	career-ops-gui                 # aktuelles Verzeichnis
//	career-ops-gui --path ..       # anderes career-ops-Verzeichnis
package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/santifer/career-ops/dashboard/guiapp/internal/app"
)

func main() {
	pathFlag := flag.String("path", "", "Pfad zum career-ops-Verzeichnis (Vorgabe: automatisch ermittelt)")
	flag.Parse()

	root := *pathFlag
	if root == "" {
		root = detectRoot()
	}
	st, err := os.Stat(root)
	if err != nil || !st.IsDir() {
		app.Fatal("Verzeichnis nicht gefunden: %s", root)
	}

	app.New(root).Run()
}

// detectRoot ermittelt das career-ops-Verzeichnis, wenn --path fehlt. Geprüft
// werden das Arbeitsverzeichnis und die Verzeichnisse oberhalb der Binary —
// so funktioniert sowohl "go run ." aus gui/ als auch ein Doppelklick auf die
// gebaute Datei in gui/ oder im Repo-Wurzelverzeichnis.
func detectRoot() string {
	if wd, err := os.Getwd(); err == nil && isCareerOpsRoot(wd) {
		return wd
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for i := 0; i < 3; i++ {
			if isCareerOpsRoot(dir) {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if wd, err := os.Getwd(); err == nil {
		// Aus gui/ heraus gestartet: eine Ebene höher liegt das Repo.
		if parent := filepath.Dir(wd); isCareerOpsRoot(parent) {
			return parent
		}
		return wd
	}
	return "."
}

// isCareerOpsRoot erkennt ein career-ops-Verzeichnis an Dateien, die dort
// immer liegen — auch bevor der Nutzer die Einrichtung abgeschlossen hat.
func isCareerOpsRoot(dir string) bool {
	for _, marker := range []string{"AGENTS.md", "doctor.mjs", "tracker.mjs"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}
