package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"git-muse/internal/git"
)

type Config struct {
	Provider     string
	Model        string
	BaseURL      string
	Format       string
	Language     string
	Count        int
	MaxDiffBytes int
	Temperature  float64
	APIKey       string
	APIKeyName   string
}

func Load(ctx context.Context, client git.Client) Config {
	cfg := Config{
		Format:       "conventional",
		Language:     "en",
		Count:        1,
		MaxDiffBytes: 100000,
		Temperature:  0.2,
	}
	cfg.Provider = value(ctx, client, "muse.provider", "GIT_MUSE_PROVIDER", cfg.Provider)
	cfg.Model = value(ctx, client, "muse.model", "GIT_MUSE_MODEL", cfg.Model)
	cfg.BaseURL = value(ctx, client, "muse.baseUrl", "GIT_MUSE_BASE_URL", cfg.BaseURL)
	cfg.Format = value(ctx, client, "muse.format", "GIT_MUSE_FORMAT", cfg.Format)
	cfg.Language = value(ctx, client, "muse.language", "GIT_MUSE_LANGUAGE", cfg.Language)
	cfg.Count = intValue(ctx, client, "muse.count", "GIT_MUSE_COUNT", cfg.Count)
	cfg.MaxDiffBytes = intValue(ctx, client, "muse.maxDiffBytes", "GIT_MUSE_MAX_DIFF_BYTES", cfg.MaxDiffBytes)
	cfg.Temperature = floatValue(ctx, client, "muse.temperature", "GIT_MUSE_TEMPERATURE", cfg.Temperature)

	switch cfg.Provider {
	case "compatible", "openai", "openrouter":
		cfg.APIKeyName = apiKeyName(cfg.Provider)
		cfg.APIKey = os.Getenv(cfg.APIKeyName)
		if cfg.BaseURL == "" && cfg.Provider == "openrouter" {
			cfg.BaseURL = "https://openrouter.ai/api/v1"
		}
		if cfg.BaseURL == "" {
			cfg.BaseURL = "https://api.openai.com/v1"
		}
	case "ollama":
		if cfg.BaseURL == "" {
			cfg.BaseURL = "http://localhost:11434"
		}
	}
	return cfg
}

func (c Config) ValidateProvider() error {
	if c.Provider == "" {
		return fmt.Errorf("no LLM provider configured\n\nRun:\n\n  git config --global muse.provider ollama\n  git config --global muse.model qwen2.5-coder")
	}
	if c.Model == "" {
		return fmt.Errorf("no LLM model configured\n\nRun:\n\n  git config --global muse.model qwen2.5-coder")
	}
	if c.Provider != "ollama" && c.APIKeyName != "" && c.APIKey == "" {
		return fmt.Errorf("%s not set", c.APIKeyName)
	}
	return nil
}

func value(ctx context.Context, client git.Client, key, env, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(env)); v != "" {
		return v
	}
	if v, err := client.Config(ctx, key); err == nil && v != "" {
		return v
	}
	return fallback
}

func intValue(ctx context.Context, client git.Client, key, env string, fallback int) int {
	v := value(ctx, client, key, env, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func floatValue(ctx context.Context, client git.Client, key, env string, fallback float64) float64 {
	v := value(ctx, client, key, env, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return n
}

func apiKeyName(provider string) string {
	switch provider {
	case "openrouter":
		return "OPENROUTER_API_KEY"
	default:
		return "OPENAI_API_KEY"
	}
}
