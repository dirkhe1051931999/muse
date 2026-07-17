package app

import "testing"

func TestPrependPathMovesCurrentExeDirFirst(t *testing.T) {
	got := prependPath("old;tools;new", "new")
	want := "new;old;tools"
	if got != want {
		t.Fatalf("prependPath() = %q, want %q", got, want)
	}
}
