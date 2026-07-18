package llm

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestDecodeCandidatesObject(t *testing.T) {
	data := []byte(`{"candidates":[{"type":"feat","subject":"add config reader"}]}`)
	got, err := decodeCandidates(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Candidates[0].Type != "feat" {
		t.Fatalf("unexpected candidate: %#v", got.Candidates[0])
	}
}

func TestDecodeCandidatesRejectsBadType(t *testing.T) {
	data := []byte(`{"candidates":[{"type":"bad","subject":"add config reader"}]}`)
	if _, err := decodeCandidates(data); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDecodeCandidatesClampsLongSubject(t *testing.T) {
	long := strings.Repeat("中", 90)
	data := []byte(`{"candidates":[{"type":"feat","subject":"` + long + `"}]}`)
	got, err := decodeCandidates(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := utf8.RuneCountInString(got.Candidates[0].Subject); n > 72 {
		t.Fatalf("subject not clamped: %d runes", n)
	}
}
