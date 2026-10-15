package diffctx

import (
	"strings"
	"testing"
)

func TestFilterExcludesEnvFile(t *testing.T) {
	diff := "diff --git a/.env b/.env\n+OPENAI_API_KEY=secret\n" +
		"diff --git a/main.go b/main.go\n+package main\n"

	got, excluded, _ := Filter(diff, FilterOptions{MaxBytes: 100000})
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
	got, _, _ := Filter(diff, FilterOptions{MaxBytes: 100000})
	if strings.Contains(got, "secret") {
		t.Fatalf("secret line was not redacted: %s", got)
	}
	if !strings.Contains(got, "token=<redacted>") {
		t.Fatalf("redacted marker missing: %s", got)
	}
}

func TestFilterTruncates(t *testing.T) {
	diff := "diff --git a/a.txt b/a.txt\n" + strings.Repeat("+x\n", 100)
	got, _, truncated := Filter(diff, FilterOptions{MaxBytes: 20})
	if !truncated {
		t.Fatalf("expected truncated")
	}
	if len(got) != 20 {
		t.Fatalf("got length %d", len(got))
	}
}

func TestFilterUserIgnore(t *testing.T) {
	diff := "diff --git a/go.sum b/go.sum\n+x\n" +
		"diff --git a/main.go b/main.go\n+package main\n"
	got, excluded, _ := Filter(diff, FilterOptions{
		MaxBytes: 100000,
		Ignore:   []string{"go.sum"},
	})
	if strings.Contains(got, "go.sum") && strings.Contains(got, "+x") {
		t.Fatalf("ignored file still present: %s", got)
	}
	if !strings.Contains(got, "package main") {
		t.Fatalf("normal diff missing: %s", got)
	}
	if len(excluded) != 1 || excluded[0].Reason != "user ignore" {
		t.Fatalf("unexpected excluded: %#v", excluded)
	}
}

func TestFilterAllowOverridesBuiltin(t *testing.T) {
	diff := "diff --git a/.env.example b/.env.example\n+FOO=bar\n"
	got, excluded, _ := Filter(diff, FilterOptions{
		MaxBytes: 100000,
		Allow:    []string{".env.example"},
	})
	if len(excluded) != 0 {
		t.Fatalf("expected allow to include file, excluded=%#v", excluded)
	}
	if !strings.Contains(got, "FOO=bar") {
		t.Fatalf("allowed file missing: %s", got)
	}
}

func TestFilterAllowBeatsIgnore(t *testing.T) {
	diff := "diff --git a/vendor/keep.go b/vendor/keep.go\n+package keep\n"
	got, excluded, _ := Filter(diff, FilterOptions{
		MaxBytes: 100000,
		Ignore:   []string{"vendor/*"},
		Allow:    []string{"vendor/keep.go"},
	})
	if len(excluded) != 0 {
		t.Fatalf("allow should win over ignore: %#v", excluded)
	}
	if !strings.Contains(got, "package keep") {
		t.Fatalf("allowed path missing: %s", got)
	}
}
