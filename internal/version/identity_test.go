package version

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIdentityIncludesVersionAndPath(t *testing.T) {
	got := Identity()
	if !strings.Contains(got, "git-muse "+Version) {
		t.Fatalf("missing version: %q", got)
	}
	if !strings.Contains(got, "Executable:") {
		t.Fatalf("missing executable: %q", got)
	}
	exe := Executable()
	if exe == "" || exe == "unknown" {
		t.Fatalf("executable = %q", exe)
	}
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("stat executable %s: %v", exe, err)
	}
	if filepath.Base(exe) == "" {
		t.Fatalf("empty base: %q", exe)
	}
}
