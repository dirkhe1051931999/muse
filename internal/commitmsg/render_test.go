package commitmsg

import "testing"

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
