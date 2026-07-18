package docs

import (
	"strings"
	"testing"
)

func TestHTMLContainsUsage(t *testing.T) {
	got := HTML()
	if !strings.Contains(got, "git-muse") {
		t.Fatalf("missing title content")
	}
	if !strings.Contains(got, "git muse help") {
		t.Fatalf("missing help guidance")
	}
	if !strings.Contains(got, "<!DOCTYPE html>") {
		t.Fatalf("missing html doctype")
	}
}

func TestUsageMentionsGitHelpInterception(t *testing.T) {
	if !strings.Contains(Usage, "git-muse.html") {
		t.Fatalf("usage should mention git-muse.html")
	}
}
