package llm

import "testing"

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
