package runner

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ErrBusy meldet den Versuch, ein zweites Werkzeug zu starten, während noch
// eines läuft. Der Runner führt bewusst nur einen Prozess gleichzeitig aus,
// damit zwei Schreibvorgänge nicht um die Bewerbungsliste konkurrieren.
var ErrBusy = errors.New("es läuft bereits ein Werkzeug")

// ErrNodeMissing meldet, dass node nicht im Suchpfad liegt. Ohne node ist
// kein einziges Repo-Werkzeug aufrufbar, deshalb wird das gesondert gemeldet.
var ErrNodeMissing = errors.New("node wurde nicht gefunden — bitte Node.js installieren und in den PATH aufnehmen")

// LineFunc erhält jede Ausgabezeile des laufenden Prozesses.
type LineFunc func(line string)

// DoneFunc meldet das Ende eines Laufs: Exit-Code und, falls der Start selbst
// scheiterte, der Fehler.
type DoneFunc func(exitCode int, err error)

// Runner führt Node-Werkzeuge im career-ops-Verzeichnis aus. Genau ein Lauf
// gleichzeitig; ein laufender Lauf ist abbrechbar.
type Runner struct {
	// careerOpsPath ist das Arbeitsverzeichnis für jeden Aufruf.
	careerOpsPath string

	mu     sync.Mutex
	cancel context.CancelFunc
	active bool
}

// New erzeugt einen Runner für das angegebene career-ops-Verzeichnis.
func New(careerOpsPath string) *Runner {
	return &Runner{careerOpsPath: careerOpsPath}
}

// Busy meldet, ob gerade ein Werkzeug läuft.
func (r *Runner) Busy() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active
}

// Cancel bricht den laufenden Prozess ab. Ohne laufenden Prozess passiert nichts.
func (r *Runner) Cancel() {
	r.mu.Lock()
	cancel := r.cancel
	r.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// CommandLine baut die Anzeige der Befehlszeile, die Start ausführen würde.
// Die Oberfläche zeigt sie an, damit nachvollziehbar bleibt, was läuft.
func CommandLine(t Tool, extra []string) string {
	parts := append([]string{"node", t.Script}, t.Args...)
	parts = append(parts, extra...)
	return strings.Join(parts, " ")
}

// SplitArgs zerlegt eine eingegebene Argumentzeile. Anführungszeichen halten
// zusammengehörige Werte zusammen, damit --note "per Portal" ein Argument bleibt.
func SplitArgs(s string) []string {
	var (
		args    []string
		current strings.Builder
		quote   rune
		open    bool
	)
	flush := func() {
		if current.Len() > 0 || open {
			args = append(args, current.String())
			current.Reset()
		}
	}
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '"' || r == '\'':
			quote = r
			open = true
		case r == ' ' || r == '\t':
			flush()
			open = false
		default:
			current.WriteRune(r)
			open = true
		}
	}
	flush()
	return args
}

// Start führt ein Werkzeug aus. onLine wird für jede Ausgabezeile aufgerufen,
// onDone einmal am Ende — beide aus einer Hintergrund-Goroutine, die
// Oberfläche muss die Aufrufe selbst in den UI-Thread überführen.
func (r *Runner) Start(t Tool, extra []string, onLine LineFunc, onDone DoneFunc) error {
	r.mu.Lock()
	if r.active {
		r.mu.Unlock()
		return ErrBusy
	}
	if _, err := exec.LookPath("node"); err != nil {
		r.mu.Unlock()
		return ErrNodeMissing
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.active = true
	r.mu.Unlock()

	args := append([]string{t.Script}, t.Args...)
	args = append(args, extra...)

	cmd := exec.CommandContext(ctx, "node", args...)
	cmd.Dir = r.careerOpsPath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.finish(cancel)
		return fmt.Errorf("Ausgabe konnte nicht gelesen werden: %w", err)
	}
	// stderr auf dieselbe Pipe legen: Fortschritts- und Fehlermeldungen der
	// Skripte gehören in der Oberfläche in eine gemeinsame, zeitlich richtige
	// Ausgabe. Ein zweiter StderrPipe-Aufruf wäre hier ein Fehler, weil
	// cmd.Stderr dann bereits belegt ist.
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		r.finish(cancel)
		return fmt.Errorf("%s konnte nicht gestartet werden: %w", t.Script, err)
	}

	go func() {
		pump(stdout, onLine)
		waitErr := cmd.Wait()

		// Den Abbruchzustand VOR finish festhalten: finish ruft cancel auf,
		// wodurch ctx.Err() danach immer gesetzt wäre und jeder normal
		// beendete Lauf fälschlich als abgebrochen gälte.
		cancelled := ctx.Err() != nil
		r.finish(cancel)

		if cancelled {
			onDone(-1, errors.New("abgebrochen"))
			return
		}
		code := 0
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			code = exitErr.ExitCode()
		} else if waitErr != nil {
			onDone(-1, waitErr)
			return
		}
		onDone(code, nil)
	}()
	return nil
}

// finish gibt den Runner wieder frei.
func (r *Runner) finish(cancel context.CancelFunc) {
	cancel()
	r.mu.Lock()
	r.active = false
	r.cancel = nil
	r.mu.Unlock()
}

// pump liest zeilenweise und reicht jede Zeile weiter. Der Puffer ist großzügig,
// weil einige Skripte lange JSON-Zeilen ausgeben.
func pump(rc io.ReadCloser, onLine LineFunc) {
	defer rc.Close()
	sc := bufio.NewScanner(rc)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		onLine(sc.Text())
	}
}

// ScriptExists prüft, ob ein Werkzeug im career-ops-Verzeichnis vorhanden ist.
// Die Oberfläche graut fehlende Werkzeuge aus, statt sie scheitern zu lassen.
func (r *Runner) ScriptExists(t Tool) bool {
	return fileExists(filepath.Join(r.careerOpsPath, t.Script))
}
