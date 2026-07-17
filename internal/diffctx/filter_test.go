package diffctx

import (
	"strings"
	"testing"
)

func TestFilterExcludesEnvFile(t *testing.T) {
	diff := "diff --git a/.env b/.env\n+OPENAI_API_KEY=secret\n" +
		"diff --git a/main.go b/main.go\n+package main\n"

	got, excluded, _ := Filter(diff, 100000)
	if strings.Contains(got, "OPENAI_API_KEY") {
		t.Fatalf("secret diff was not excluded: %s", got)
	}
	if !strings.Contains(got, "package main") {
		t.Fatalf("normal diff missing: %s", got)
	}
	if len(excluded) != 1 || excluded[0].Path != ".env" {
		t.Fatalf("unexpected excluded files: %#v", excluded)
	}
}

func TestFilterRedactsSecretLine(t *testing.T) {
	diff := "diff --git a/config.txt b/config.txt\n+token=secret\n"
	got, _, _ := Filter(diff, 100000)
	if strings.Contains(got, "secret") {
		t.Fatalf("secret line was not redacted: %s", got)
	}
	if !strings.Contains(got, "token=<redacted>") {
		t.Fatalf("redacted marker missing: %s", got)
	}
}

func TestFilterTruncates(t *testing.T) {
	diff := "diff --git a/a.txt b/a.txt\n" + strings.Repeat("+x\n", 100)
	got, _, truncated := Filter(diff, 20)
	if !truncated {
		t.Fatalf("expected truncated")
	}
	if len(got) != 20 {
		t.Fatalf("got length %d", len(got))
	}
}
