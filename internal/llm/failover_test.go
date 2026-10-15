package llm

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"git-muse/internal/commitmsg"
	"git-muse/internal/config"
)

type stubProvider struct {
	name string
	err  error
	resp GenerateResponse
}

func (s stubProvider) Name() string { return s.name }

func (s stubProvider) Generate(context.Context, GenerateRequest) (GenerateResponse, error) {
	if s.err != nil {
		return GenerateResponse{}, s.err
	}
	return s.resp, nil
}

func TestLooksLikeRateLimit(t *testing.T) {
	if !looksLikeRateLimit(429, "") {
		t.Fatalf("429 should be rate limited")
	}
	if !looksLikeRateLimit(403, `{"error":{"code":"rate_limit_exceeded"}}`) {
		t.Fatalf("rate_limit_exceeded should match")
	}
	if !looksLikeRateLimit(200, "RPM exceeded for this key") {
		t.Fatalf("RPM token should match")
	}
	if looksLikeRateLimit(400, "invalid rpmbuild argument") {
		t.Fatalf("rpmbuild should not match")
	}
	if !looksLikeRateLimit(503, "触发限流，请稍后重试") {
		t.Fatalf("chinese rate limit should match")
	}
}

func TestFailoverSwitchesOnRateLimit(t *testing.T) {
	ok := GenerateResponse{Candidates: []commitmsg.Candidate{{Type: "feat", Subject: "ok"}}}
	var log bytes.Buffer
	p := Failover{
		Providers: []Provider{
			stubProvider{name: "a", err: &RateLimitError{Status: "429 Too Many Requests", Body: "rpm"}},
			stubProvider{name: "b", resp: ok},
		},
		Labels:           []string{"compatible/model-a", "compatible/model-b"},
		RateLimitRetries: 1,
		Log:              &log,
	}
	got, err := p.Generate(context.Background(), GenerateRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Candidates[0].Subject != "ok" {
		t.Fatalf("unexpected response: %+v", got)
	}
	out := log.String()
	if !strings.Contains(out, "rate limited") || !strings.Contains(out, "compatible/model-a") || !strings.Contains(out, "compatible/model-b") {
		t.Fatalf("missing switch log: %q", out)
	}
}

type countStub struct {
	name  string
	err   error
	resp  GenerateResponse
	fail  int
	calls int
}

func (s *countStub) Name() string { return s.name }

func (s *countStub) Generate(context.Context, GenerateRequest) (GenerateResponse, error) {
	s.calls++
	if s.calls <= s.fail {
		return GenerateResponse{}, s.err
	}
	return s.resp, nil
}

func TestFailoverRetriesBeforeSwitch(t *testing.T) {
	ok := GenerateResponse{Candidates: []commitmsg.Candidate{{Type: "feat", Subject: "ok"}}}
	first := &countStub{name: "a", fail: 2, err: &RateLimitError{Status: "429 Too Many Requests"}, resp: ok}
	second := &countStub{name: "b", fail: 0, resp: ok}
	var log bytes.Buffer
	p := Failover{
		Providers:        []Provider{first, second},
		Labels:           []string{"compatible/model-a", "compatible/model-b"},
		RateLimitRetries: 3,
		Log:              &log,
	}
	if _, err := p.Generate(context.Background(), GenerateRequest{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first.calls != 3 {
		t.Fatalf("first.calls = %d, want 3", first.calls)
	}
	if second.calls != 0 {
		t.Fatalf("second.calls = %d, want 0", second.calls)
	}
	out := log.String()
	if !strings.Contains(out, "retry 1/3") || !strings.Contains(out, "retry 2/3") {
		t.Fatalf("missing retry log: %q", out)
	}
	if strings.Contains(out, "switching") {
		t.Fatalf("should not switch: %q", out)
	}
}

func TestFailoverSwitchesAfterRetryBudget(t *testing.T) {
	ok := GenerateResponse{Candidates: []commitmsg.Candidate{{Type: "feat", Subject: "ok"}}}
	first := &countStub{name: "a", fail: 10, err: &RateLimitError{Status: "429 Too Many Requests"}}
	second := &countStub{name: "b", fail: 0, resp: ok}
	var log bytes.Buffer
	p := Failover{
		Providers:        []Provider{first, second},
		Labels:           []string{"compatible/model-a", "compatible/model-b"},
		RateLimitRetries: 3,
		Log:              &log,
	}
	if _, err := p.Generate(context.Background(), GenerateRequest{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first.calls != 3 {
		t.Fatalf("first.calls = %d, want 3", first.calls)
	}
	if second.calls != 1 {
		t.Fatalf("second.calls = %d, want 1", second.calls)
	}
	out := log.String()
	if !strings.Contains(out, "retry 1/3") || !strings.Contains(out, "retry 2/3") || !strings.Contains(out, "switching") {
		t.Fatalf("missing retry/switch log: %q", out)
	}
}

func TestFailoverDoesNotSwitchOnOtherErrors(t *testing.T) {
	p := Failover{
		Providers: []Provider{
			stubProvider{name: "a", err: errors.New("LLM request failed: 401 Unauthorized")},
			stubProvider{name: "b", resp: GenerateResponse{Candidates: []commitmsg.Candidate{{Type: "feat", Subject: "ok"}}}},
		},
		Log: ioDiscard(),
	}
	if _, err := p.Generate(context.Background(), GenerateRequest{}); err == nil {
		t.Fatalf("expected auth error")
	}
}

func TestNewChainUsesFallbacks(t *testing.T) {
	cfg := config.Config{
		Provider: "ollama",
		Model:    "a",
		BaseURL:  "http://localhost:11434",
		Fallbacks: []config.Config{
			{Provider: "ollama", Model: "b", BaseURL: "http://localhost:11434"},
		},
	}
	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	fo, ok := p.(Failover)
	if !ok || len(fo.Providers) != 2 {
		t.Fatalf("expected failover chain, got %#v", p)
	}
}

func ioDiscard() *bytes.Buffer {
	return &bytes.Buffer{}
}
