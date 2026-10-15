package commitmsg

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRenderConventionalCommit(t *testing.T) {
	got := Render(Candidate{
		Type:    "feat",
		Scope:   "auth",
		Subject: "add refresh token rotation",
	})
	want := "feat(auth): add refresh token rotation\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenderBody(t *testing.T) {
	got := Render(Candidate{
		Type:    "fix",
		Subject: "handle empty config",
		Body:    []string{"Return clear error.", "Keep existing config untouched."},
	})
	want := "fix: handle empty config\n\nReturn clear error.\nKeep existing config untouched.\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNormalizeClampsLongSubject(t *testing.T) {
	long := strings.Repeat("中", 80)
	got := Normalize(Candidate{Type: "feat", Subject: long})
	if utf8.RuneCountInString(got.Subject) > maxSubjectRunes {
		t.Fatalf("subject still too long: %d", utf8.RuneCountInString(got.Subject))
	}
	if err := Validate(got); err != nil {
		t.Fatalf("Validate after clamp: %v", err)
	}
}

func TestNormalizeCutsEnglishAtWordBoundary(t *testing.T) {
	subject := strings.Repeat("word ", 20) // longer than 72
	got := Normalize(Candidate{Type: "fix", Subject: subject})
	if utf8.RuneCountInString(got.Subject) > maxSubjectRunes {
		t.Fatalf("subject still too long: %d", utf8.RuneCountInString(got.Subject))
	}
	if strings.HasSuffix(got.Subject, " ") {
		t.Fatalf("trailing space left: %q", got.Subject)
	}
}

func TestDecodeLongSubjectNoLongerFails(t *testing.T) {
	c := Candidate{Type: "chore", Subject: strings.Repeat("改", 100)}
	if err := Validate(c); err != nil {
		t.Fatalf("expected clamp path to pass Validate: %v", err)
	}
}
