package version

import (
	"fmt"
	"os"
	"path/filepath"
)

func Identity() string {
	return fmt.Sprintf("git-muse %s\nExecutable: %s\n", Version, Executable())
}

func Executable() string {
	exe, err := os.Executable()
	if err != nil {
		return "unknown"
	}
	exe = filepath.Clean(exe)
	if resolved, err := filepath.EvalSymlinks(exe); err == nil && resolved != "" {
		return resolved
	}
	return exe
}
