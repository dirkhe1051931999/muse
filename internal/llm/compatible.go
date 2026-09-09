package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Compatible struct {
	BaseURL     string
	Model       string
	APIKey      string
	Temperature float64
	MaxTokens   int
	TimeoutSec  int
	Proxy       string
}

func (p Compatible) Name() string { return "compatible" }

func (p Compatible) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	system, user := messages(req)
	messages := []map[string]string{
		{"role": "system", "content": system},
		{"role": "user", "content": user},
	}
	// Prefer disabling thinking: deepseek-v4* defaults to CoT and can exhaust max_tokens.
	respBody, err := p.chat(ctx, buildCompatiblePayload(p, messages, true))
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "thinking") {
		respBody, err = p.chat(ctx, buildCompatiblePayload(p, messages, false))
	}
	if err != nil {
		return GenerateResponse{}, err
	}

	content, finishReason, err := extractCompatibleContent(respBody)
	if err != nil {
		return GenerateResponse{}, err
	}
	if strings.TrimSpace(content) == "" {
		return GenerateResponse{}, emptyCompatibleError(finishReason, respBody)
	}
	return decodeCandidates([]byte(stripCodeFence(content)))
}

func buildCompatiblePayload(p Compatible, messages []map[string]string, disableThinking bool) map[string]any {
	payload := map[string]any{
		"model":       p.Model,
		"temperature": p.Temperature,
		"max_tokens":  p.MaxTokens,
		"messages":    messages,
	}
	if disableThinking {
		payload["thinking"] = map[string]string{"type": "disabled"}
	}
	return payload
}

func (p Compatible) chat(ctx context.Context, payload map[string]any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(p.BaseURL, "/")
	if !strings.HasSuffix(url, "/chat/completions") {
		url += "/chat/completions"
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	client, err := newHTTPClient(p.TimeoutSec, p.Proxy)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if err := rateLimitError(resp.StatusCode, resp.Status, respBody); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("LLM request failed: %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	return respBody, nil
}

func extractCompatibleContent(respBody []byte) (content, finishReason string, err error) {
	var parsed struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content          json.RawMessage `json:"content"`
				ReasoningContent string          `json:"reasoning_content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", "", err
	}
	if len(parsed.Choices) == 0 {
		return "", "", nil
	}
	choice := parsed.Choices[0]
	content = rawJSONString(choice.Message.Content)
	if strings.TrimSpace(content) == "" {
		// Some gateways put the final answer only in reasoning when truncated oddly.
		content = strings.TrimSpace(choice.Message.ReasoningContent)
	}
	return content, choice.FinishReason, nil
}

func rawJSONString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return strings.TrimSpace(string(raw))
}

func emptyCompatibleError(finishReason string, respBody []byte) error {
	snippet := strings.TrimSpace(string(respBody))
	if len(snippet) > 400 {
		snippet = snippet[:400] + "..."
	}
	hint := "Increase llm.maxTokens, or disable thinking for this model."
	if finishReason == "length" {
		hint = "finish_reason=length: thinking/output hit max_tokens. Raise llm.maxTokens (e.g. 4096) or keep thinking disabled."
	}
	if snippet == "" {
		return fmt.Errorf("LLM returned empty response\n\n%s", hint)
	}
	return fmt.Errorf("LLM returned empty response\n\n%s\n\nResponse snippet:\n%s", hint, snippet)
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) >= 3 {
			return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
		}
	}
	return s
}
