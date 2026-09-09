package runner

import "os"

// fileExists meldet, ob unter path eine reguläre Datei liegt.
func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
