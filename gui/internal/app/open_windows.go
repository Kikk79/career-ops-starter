//go:build windows

package app

import (
	"os/exec"
	"strings"
	"syscall"
)

// openWithDefaultApp öffnet eine Datei oder URL mit dem Standardprogramm.
//
// Unter Windows übernimmt das rundll32 statt "cmd /c start": cmd würde den
// Pfad erneut durch seine eigene Zeichenersetzung schicken, wodurch Pfade mit
// & oder ^ zerbrechen. Fällt der Aufruf durch, greift explorer.exe.
func openWithDefaultApp(target string) error {
	t := strings.TrimSpace(target)
	if t == "" {
		return nil
	}
	cmd := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", t)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		return exec.Command("explorer.exe", t).Start()
	}
	return nil
}
