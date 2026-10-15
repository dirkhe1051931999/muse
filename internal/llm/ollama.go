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

type Ollama struct {
	BaseURL     string
	Model       string
	Temperature float64
	MaxTokens   int
	TimeoutSec  int
	Proxy       string
}

func (p Ollama) Name() string { return "ollama" }

func (p Ollama) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	system, user := messages(req)
	payload := map[string]any{
		"model":  p.Model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"options": map[string]any{"temperature": p.Temperature, "num_predict": p.MaxTokens},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return GenerateResponse{}, err
	}

	url := strings.TrimRight(p.BaseURL, "/") + "/api/chat"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return GenerateResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client, err := newHTTPClient(p.TimeoutSec, p.Proxy)
	if err != nil {
		return GenerateResponse{}, err
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return GenerateResponse{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if err := rateLimitError(resp.StatusCode, resp.Status, respBody); err != nil {
			return GenerateResponse{}, err
		}
		return GenerateResponse{}, fmt.Errorf("Ollama request failed: %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return GenerateResponse{}, err
	}
	if strings.TrimSpace(parsed.Message.Content) == "" {
		return GenerateResponse{}, fmt.Errorf("Ollama returned empty response")
	}
	return decodeCandidates([]byte(stripCodeFence(parsed.Message.Content)))
}
