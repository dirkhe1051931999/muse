package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"git-muse/internal/commitmsg"
	"git-muse/internal/config"
	"git-muse/internal/diffctx"
	"git-muse/internal/prompt"
)

type Provider interface {
	Generate(context.Context, GenerateRequest) (GenerateResponse, error)
	Name() string
}

type GenerateRequest struct {
	Config config.Config
	Diff   diffctx.Context
}

type GenerateResponse struct {
	Candidates []commitmsg.Candidate `json:"candidates"`
}

func New(cfg config.Config) (Provider, error) {
	return newChain(cfg.Chain(), cfg.RateLimitRetries)
}

func newOne(cfg config.Config) (Provider, error) {
	switch cfg.Provider {
	case "ollama":
		return Ollama{BaseURL: cfg.BaseURL, Model: cfg.Model, Temperature: cfg.Temperature, MaxTokens: cfg.MaxTokens, TimeoutSec: cfg.TimeoutSec, Proxy: cfg.Proxy}, nil
	case "compatible", "openai", "openrouter":
		return Compatible{BaseURL: cfg.BaseURL, Model: cfg.Model, APIKey: cfg.APIKey, Temperature: cfg.Temperature, MaxTokens: cfg.MaxTokens, TimeoutSec: cfg.TimeoutSec, Proxy: cfg.Proxy}, nil
	default:
		return nil, fmt.Errorf("unsupported provider %q", cfg.Provider)
	}
}

func decodeCandidates(data []byte) (GenerateResponse, error) {
	var response GenerateResponse
	if err := json.Unmarshal(data, &response); err == nil && len(response.Candidates) > 0 {
		return response, validateResponse(response)
	}

	var candidates []commitmsg.Candidate
	if err := json.Unmarshal(data, &candidates); err == nil && len(candidates) > 0 {
		response.Candidates = candidates
		return response, validateResponse(response)
	}
	return GenerateResponse{}, fmt.Errorf("model returned invalid JSON")
}

func validateResponse(response GenerateResponse) error {
	if len(response.Candidates) == 0 {
		return fmt.Errorf("model returned no candidates")
	}
	for i, candidate := range response.Candidates {
		candidate = commitmsg.Normalize(candidate)
		if err := commitmsg.Validate(candidate); err != nil {
			return err
		}
		response.Candidates[i] = candidate
	}
	return nil
}

func messages(req GenerateRequest) (string, string) {
	return prompt.System(req.Config), prompt.User(req.Config, req.Diff)
}
