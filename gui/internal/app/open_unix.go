//go:build !windows

package app

import (
	"os/exec"
	"runtime"
	"strings"
)

// openWithDefaultApp öffnet eine Datei oder URL mit dem Standardprogramm:
// "open" unter macOS, "xdg-open" auf den übrigen Systemen.
func openWithDefaultApp(target string) error {
	t := strings.TrimSpace(target)
	if t == "" {
		return nil
	}
	opener := "xdg-open"
	if runtime.GOOS == "darwin" {
		opener = "open"
	}
	return exec.Command(opener, t).Start()
}
