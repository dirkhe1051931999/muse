package docs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InstallGitHTML writes git-muse.html into `git --html-path` so that
// `git muse --help` can open documentation instead of failing.
func InstallGitHTML(ctx context.Context) (string, error) {
	dir, err := gitHTMLPath(ctx)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "git-muse.html")
	if err := os.WriteFile(path, []byte(HTML()), 0644); err != nil {
		return "", fmt.Errorf("write %s: %w\n\nRun as administrator, or use:\n\n  git muse help\n  git-muse --help", path, err)
	}
	return path, nil
}

func gitHTMLPath(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "--html-path")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git --html-path: %w", err)
	}
	dir := strings.TrimSpace(string(out))
	if dir == "" {
		return "", fmt.Errorf("git --html-path returned empty path")
	}
	return filepath.Clean(dir), nil
}
