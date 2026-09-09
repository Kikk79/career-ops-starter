package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot liefert das career-ops-Wurzelverzeichnis relativ zu diesem Test.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("Wurzelverzeichnis nicht auflösbar: %v", err)
	}
	return root
}

// Ein Katalogeintrag, dessen Skript nicht existiert, ist ein toter Knopf in der
// Oberfläche. Dieser Test fängt das ab, sobald das Repo ein Skript umbenennt.
func TestCatalogScriptsExist(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Skipf("kein career-ops-Repo unter %s", root)
	}

	for _, g := range Catalog() {
		for _, tool := range g.Tools {
			if _, err := os.Stat(filepath.Join(root, tool.Script)); err != nil {
				t.Errorf("%s / %s: Skript %s fehlt im Repo", g.Name, tool.Name, tool.Script)
			}
		}
	}
}

// Jedes Werkzeug braucht Name, Skript und eine Beschreibung — ohne
// Beschreibung ist die Oberfläche für Einsteiger wertlos.
func TestCatalogEntriesComplete(t *testing.T) {
	for _, g := range Catalog() {
		if strings.TrimSpace(g.Name) == "" {
			t.Error("Gruppe ohne Namen")
		}
		if len(g.Tools) == 0 {
			t.Errorf("Gruppe %q ist leer", g.Name)
		}
		for _, tool := range g.Tools {
			switch {
			case strings.TrimSpace(tool.Name) == "":
				t.Errorf("%s: Werkzeug ohne Namen", g.Name)
			case !strings.HasSuffix(tool.Script, ".mjs"):
				t.Errorf("%s / %s: %q ist kein .mjs-Skript", g.Name, tool.Name, tool.Script)
			case len(strings.TrimSpace(tool.Desc)) < 20:
				t.Errorf("%s / %s: Beschreibung fehlt oder ist zu knapp", g.Name, tool.Name)
			}
		}
	}
}

// Namen müssen katalogweit eindeutig sein, sonst sind zwei Baumeinträge nicht
// unterscheidbar.
func TestCatalogNamesUnique(t *testing.T) {
	seen := map[string]string{}
	for _, g := range Catalog() {
		for _, tool := range g.Tools {
			if prev, dup := seen[tool.Name]; dup {
				t.Errorf("Name %q doppelt vergeben (%s und %s)", tool.Name, prev, g.Name)
			}
			seen[tool.Name] = g.Name
		}
	}
}

func TestCommandLine(t *testing.T) {
	tool := Tool{Script: "update-system.mjs", Args: []string{"check"}}
	if got, want := CommandLine(tool, []string{"--json"}), "node update-system.mjs check --json"; got != want {
		t.Errorf("CommandLine = %q, erwartet %q", got, want)
	}
}
