package llm

import (
	"strings"
	"testing"
)

func TestExtractCompatibleContent(t *testing.T) {
	body := []byte(`{"choices":[{"finish_reason":"stop","message":{"content":"{\"candidates\":[]}"}}]}`)
	content, reason, err := extractCompatibleContent(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reason != "stop" || content == "" {
		t.Fatalf("content=%q reason=%q", content, reason)
	}
}

func TestExtractCompatibleContentNullUsesReasoning(t *testing.T) {
	body := []byte(`{"choices":[{"finish_reason":"length","message":{"content":null,"reasoning_content":"{\"candidates\":[{\"type\":\"feat\",\"subject\":\"x\"}]}"}}]}`)
	content, reason, err := extractCompatibleContent(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reason != "length" {
		t.Fatalf("reason=%q", reason)
	}
	if content == "" {
		t.Fatalf("expected reasoning fallback")
	}
}

func TestEmptyCompatibleErrorMentionsLength(t *testing.T) {
	err := emptyCompatibleError("length", []byte(`{"choices":[]}`))
	if err == nil || !strings.Contains(err.Error(), "empty response") || !strings.Contains(err.Error(), "max_tokens") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildCompatiblePayloadDisablesThinking(t *testing.T) {
	p := Compatible{Model: "deepseek-v4-flash", MaxTokens: 2048}
	payload := buildCompatiblePayload(p, nil, true)
	thinking, ok := payload["thinking"].(map[string]string)
	if !ok || thinking["type"] != "disabled" {
		t.Fatalf("thinking = %#v", payload["thinking"])
	}
	payload = buildCompatiblePayload(p, nil, false)
	if _, ok := payload["thinking"]; ok {
		t.Fatalf("thinking should be omitted")
	}
}
