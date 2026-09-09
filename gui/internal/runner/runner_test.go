package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"leer", "", nil},
		{"nur Leerzeichen", "   ", nil},
		{"einfach", "--summary", []string{"--summary"}},
		{"mehrere", "42 Applied", []string{"42", "Applied"}},
		{"doppelte Anführungszeichen", `--note "per Portal beworben"`, []string{"--note", "per Portal beworben"}},
		{"einfache Anführungszeichen", `--note 'zwei Wörter'`, []string{"--note", "zwei Wörter"}},
		{"leerer Wert in Anführungszeichen", `--note ""`, []string{"--note", ""}},
		{"mehrfache Leerzeichen", "a    b", []string{"a", "b"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SplitArgs(tc.in)
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("SplitArgs(%q) = %#v, erwartet %#v", tc.in, got, tc.want)
			}
		})
	}
}

// writeScript legt ein kleines Node-Skript im temporären Verzeichnis ab.
func writeScript(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("Skript konnte nicht geschrieben werden: %v", err)
	}
}

// requireNode überspringt einen Test, wenn node fehlt.
func requireNode(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node ist nicht installiert")
	}
}

func TestStartStreamsOutputAndExitCode(t *testing.T) {
	requireNode(t)
	dir := t.TempDir()
	writeScript(t, dir, "ok.mjs", "console.log('erste');\nconsole.log('zweite');\n")

	r := New(dir)
	var (
		mu    sync.Mutex
		lines []string
		done  = make(chan struct {
			code int
			err  error
		}, 1)
	)

	err := r.Start(Tool{Script: "ok.mjs"}, nil,
		func(line string) { mu.Lock(); lines = append(lines, line); mu.Unlock() },
		func(code int, err error) {
			done <- struct {
				code int
				err  error
			}{code, err}
		},
	)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case res := <-done:
		if res.err != nil {
			t.Fatalf("unerwarteter Fehler: %v", res.err)
		}
		if res.code != 0 {
			t.Errorf("Exit-Code = %d, erwartet 0", res.code)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("Zeitüberschreitung")
	}

	mu.Lock()
	defer mu.Unlock()
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"erste", "zweite"} {
		if !strings.Contains(joined, want) {
			t.Errorf("Ausgabe enthält %q nicht: %q", want, joined)
		}
	}
}

func TestStartReportsNonZeroExitCode(t *testing.T) {
	requireNode(t)
	dir := t.TempDir()
	writeScript(t, dir, "fail.mjs", "console.error('kaputt');\nprocess.exit(3);\n")

	r := New(dir)
	done := make(chan int, 1)
	if err := r.Start(Tool{Script: "fail.mjs"}, nil, func(string) {}, func(code int, _ error) { done <- code }); err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case code := <-done:
		if code != 3 {
			t.Errorf("Exit-Code = %d, erwartet 3", code)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("Zeitüberschreitung")
	}
}

func TestStartRejectsSecondRun(t *testing.T) {
	requireNode(t)
	dir := t.TempDir()
	writeScript(t, dir, "slow.mjs", "setTimeout(() => {}, 5000);\n")

	r := New(dir)
	done := make(chan struct{}, 1)
	if err := r.Start(Tool{Script: "slow.mjs"}, nil, func(string) {}, func(int, error) { done <- struct{}{} }); err != nil {
		t.Fatalf("erster Start: %v", err)
	}
	defer func() {
		r.Cancel()
		<-done
	}()

	// Kurz warten, damit der erste Lauf sicher als aktiv gilt.
	for i := 0; i < 50 && !r.Busy(); i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if err := r.Start(Tool{Script: "slow.mjs"}, nil, func(string) {}, func(int, error) {}); err != ErrBusy {
		t.Errorf("zweiter Start = %v, erwartet ErrBusy", err)
	}
}

func TestCancelStopsRun(t *testing.T) {
	requireNode(t)
	dir := t.TempDir()
	writeScript(t, dir, "slow.mjs", "setTimeout(() => {}, 60000);\n")

	r := New(dir)
	done := make(chan error, 1)
	if err := r.Start(Tool{Script: "slow.mjs"}, nil, func(string) {}, func(_ int, err error) { done <- err }); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for i := 0; i < 100 && !r.Busy(); i++ {
		time.Sleep(10 * time.Millisecond)
	}
	r.Cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Error("nach Abbruch wurde kein Fehler gemeldet")
		}
	case <-time.After(30 * time.Second):
		t.Fatal("Abbruch hat den Prozess nicht beendet")
	}
	if r.Busy() {
		t.Error("Runner ist nach dem Abbruch noch belegt")
	}
}

func TestScriptExists(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "da.mjs", "")
	r := New(dir)

	if !r.ScriptExists(Tool{Script: "da.mjs"}) {
		t.Error("vorhandenes Skript wurde nicht erkannt")
	}
	if r.ScriptExists(Tool{Script: "fehlt.mjs"}) {
		t.Error("fehlendes Skript wurde als vorhanden gemeldet")
	}
}
