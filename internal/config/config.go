package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git-muse/internal/git"
)

type Config struct {
	Provider         string
	Model            string
	BaseURL          string
	Proxy            string
	Format           string
	Language         string
	Count            int
	MaxDiffBytes     int
	IgnoreGlobs      []string
	AllowGlobs       []string
	MaxTokens        int
	TimeoutSec       int
	Temperature      float64
	APIKey           string
	APIKeyName       string
	Fallbacks        []Config
	RateLimitRetries int
	FileError        error
}

func (c Config) LLMLabel() string {
	provider := strings.TrimSpace(c.Provider)
	if provider == "" {
		provider = "unknown"
	}
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "unspecified"
	}
	return provider + "/" + model
}

func (c Config) Chain() []Config {
	primary := c
	primary.Fallbacks = nil
	out := make([]Config, 0, 1+len(c.Fallbacks))
	out = append(out, primary)
	out = append(out, c.Fallbacks...)
	return out
}

func Load(ctx context.Context, client git.Client) (Config, error) {
	cfg := Config{
		Format:           "conventional",
		Language:         "en",
		Count:            1,
		MaxDiffBytes:     100000,
		MaxTokens:        2048,
		TimeoutSec:       180,
		Temperature:      0.2,
		RateLimitRetries: 3,
	}
	cfg.Provider = gitValue(ctx, client, "muse.provider", cfg.Provider)
	cfg.Model = gitValue(ctx, client, "muse.model", cfg.Model)
	cfg.BaseURL = gitValue(ctx, client, "muse.baseUrl", cfg.BaseURL)
	cfg.Proxy = gitValue(ctx, client, "muse.proxy", cfg.Proxy)
	cfg.Format = gitValue(ctx, client, "muse.format", cfg.Format)
	cfg.Language = gitValue(ctx, client, "muse.language", cfg.Language)
	cfg.Count = intGitValue(ctx, client, "muse.count", cfg.Count)
	cfg.MaxDiffBytes = intGitValue(ctx, client, "muse.maxDiffBytes", cfg.MaxDiffBytes)
	cfg.IgnoreGlobs = csvGitValue(ctx, client, "muse.diffIgnore", cfg.IgnoreGlobs)
	cfg.AllowGlobs = csvGitValue(ctx, client, "muse.diffAllow", cfg.AllowGlobs)
	cfg.MaxTokens = intGitValue(ctx, client, "muse.maxTokens", cfg.MaxTokens)
	cfg.TimeoutSec = intGitValue(ctx, client, "muse.timeoutSeconds", cfg.TimeoutSec)
	cfg.Temperature = floatGitValue(ctx, client, "muse.temperature", cfg.Temperature)
	if fileCfg, err := LoadFile(); err == nil {
		applyFileConfig(&cfg, fileCfg)
	} else if !os.IsNotExist(err) {
		cfg.FileError = err
	}
	cfg.Provider = envValue("GIT_MUSE_PROVIDER", cfg.Provider)
	cfg.Model = envValue("GIT_MUSE_MODEL", cfg.Model)
	cfg.BaseURL = envValue("GIT_MUSE_BASE_URL", cfg.BaseURL)
	cfg.Proxy = envValue("GIT_MUSE_PROXY", cfg.Proxy)
	if env := strings.TrimSpace(os.Getenv("GIT_MUSE_PROXY")); env != "" {
		for i := range cfg.Fallbacks {
			cfg.Fallbacks[i].Proxy = env
		}
	}
	cfg.Format = envValue("GIT_MUSE_FORMAT", cfg.Format)
	cfg.Language = envValue("GIT_MUSE_LANGUAGE", cfg.Language)
	cfg.Count = intEnvValue("GIT_MUSE_COUNT", cfg.Count)
	cfg.MaxDiffBytes = intEnvValue("GIT_MUSE_MAX_DIFF_BYTES", cfg.MaxDiffBytes)
	cfg.IgnoreGlobs = csvEnvValue("GIT_MUSE_DIFF_IGNORE", cfg.IgnoreGlobs)
	cfg.AllowGlobs = csvEnvValue("GIT_MUSE_DIFF_ALLOW", cfg.AllowGlobs)
	cfg.MaxTokens = intEnvValue("GIT_MUSE_MAX_TOKENS", cfg.MaxTokens)
	cfg.TimeoutSec = intEnvValue("GIT_MUSE_TIMEOUT_SECONDS", cfg.TimeoutSec)
	cfg.Temperature = floatEnvValue("GIT_MUSE_TEMPERATURE", cfg.Temperature)
	cfg.RateLimitRetries = NormalizeRateLimitRetries(cfg.RateLimitRetries)

	lang, err := normalizeLanguage(cfg.Language)
	if err != nil {
		return cfg, err
	}
	cfg.Language = lang

	proxy, err := NormalizeProxy(cfg.Proxy)
	if err != nil {
		return cfg, err
	}
	cfg.Proxy = proxy

	if err := finalizeEndpoint(&cfg, true); err != nil {
		return cfg, err
	}
	for i := range cfg.Fallbacks {
		if err := finalizeEndpoint(&cfg.Fallbacks[i], false); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

func finalizeEndpoint(cfg *Config, envOverridesKey bool) error {
	proxy, err := NormalizeProxy(cfg.Proxy)
	if err != nil {
		return err
	}
	cfg.Proxy = proxy
	switch cfg.Provider {
	case "compatible", "openai", "openrouter":
		cfg.APIKeyName = apiKeyName(cfg.Provider)
		if key := strings.TrimSpace(os.Getenv(cfg.APIKeyName)); key != "" {
			if envOverridesKey || cfg.APIKey == "" {
				cfg.APIKey = key
			}
		}
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
	return nil
}

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find user home: %w", err)
	}
	return filepath.Join(home, ".git-muse.json"), nil
}

func normalizeLanguage(lang string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "", "en":
		return "en", nil
	case "zh":
		return "zh", nil
	default:
		return "", fmt.Errorf("unsupported language %q\n\nSupported: en, zh\n\nSet output.language in ~/.git-muse.json or:\n\n  git config muse.language en", lang)
	}
}

// NormalizeProxy accepts host:port or a full URL. Empty input is allowed.
// Bare host:port becomes http://host:port (covers HTTP and HTTPS targets via CONNECT).
func NormalizeProxy(proxy string) (string, error) {
	proxy = strings.TrimSpace(proxy)
	if proxy == "" {
		return "", nil
	}
	if !strings.Contains(proxy, "://") {
		proxy = "http://" + proxy
	}
	u, err := url.Parse(proxy)
	if err != nil {
		return "", fmt.Errorf("invalid proxy %q: %w\n\nExample:\n\n  git muse --proxy http://127.0.0.1:7890", proxy, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported proxy scheme %q\n\nSupported: http, https\n\nExample:\n\n  git muse --proxy http://127.0.0.1:7890", u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid proxy %q: missing host\n\nExample:\n\n  git muse --proxy http://127.0.0.1:7890", proxy)
	}
	return u.String(), nil
}

func NormalizeRateLimitRetries(n int) int {
	if n <= 0 {
		return 3
	}
	return n
}

func (c Config) ValidateProvider() error {
	if c.FileError != nil {
		return c.FileError
	}
	if c.Provider == "" {
		path, err := Path()
		if err != nil {
			return fmt.Errorf("no LLM provider configured\n\nRun:\n\n  git muse init")
		}
		return fmt.Errorf("no LLM provider configured\n\nRun:\n\n  git muse init\n\nThen edit:\n\n  %s", path)
	}
	if c.Model == "" {
		path, err := Path()
		if err != nil {
			return fmt.Errorf("no LLM model configured\n\nRun:\n\n  git muse init")
		}
		return fmt.Errorf("no LLM model configured\n\nEdit:\n\n  %s", path)
	}
	if c.Provider != "ollama" && c.APIKeyName != "" && c.APIKey == "" {
		return fmt.Errorf("%s not set", c.APIKeyName)
	}
	return nil
}

func gitValue(ctx context.Context, client git.Client, key, fallback string) string {
	if v, err := client.Config(ctx, key); err == nil && v != "" {
		return v
	}
	return fallback
}

func envValue(env, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(env)); v != "" {
		return v
	}
	return fallback
}

func csvGitValue(ctx context.Context, client git.Client, key string, fallback []string) []string {
	v := gitValue(ctx, client, key, "")
	if v == "" {
		return fallback
	}
	return splitCSV(v)
}

func csvEnvValue(env string, fallback []string) []string {
	v := envValue(env, "")
	if v == "" {
		return fallback
	}
	return splitCSV(v)
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func intGitValue(ctx context.Context, client git.Client, key string, fallback int) int {
	v := gitValue(ctx, client, key, "")
	return parsePositiveInt(v, fallback)
}

func intEnvValue(env string, fallback int) int {
	v := envValue(env, "")
	return parsePositiveInt(v, fallback)
}

func parsePositiveInt(v string, fallback int) int {
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func floatGitValue(ctx context.Context, client git.Client, key string, fallback float64) float64 {
	v := gitValue(ctx, client, key, "")
	return parseFloat(v, fallback)
}

func floatEnvValue(env string, fallback float64) float64 {
	v := envValue(env, "")
	return parseFloat(v, fallback)
}

func parseFloat(v string, fallback float64) float64 {
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
