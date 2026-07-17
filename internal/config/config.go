package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	MaxTokens    int
	TimeoutSec   int
	Temperature  float64
	APIKey       string
	APIKeyName   string
	FileError    error
}

type fileConfig struct {
	LLM struct {
		Provider    string  `json:"provider"`
		BaseURL     string  `json:"baseUrl"`
		Model       string  `json:"model"`
		APIKey      string  `json:"apiKey"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"maxTokens"`
		TimeoutSec  int     `json:"timeoutSeconds"`
	} `json:"llm"`
	Output struct {
		Format   string `json:"format"`
		Language string `json:"language"`
		Count    int    `json:"count"`
	} `json:"output"`
	Diff struct {
		MaxBytes int `json:"maxBytes"`
	} `json:"diff"`
}

const Example = `{
  "llm": {
    "provider": "compatible",
    "baseUrl": "https://token.sensenova.cn/v1/chat/completions",
    "model": "deepseek-v4-flash",
    "apiKey": "sk-keykeykey...keykeykey",
    "temperature": 0.2,
    "maxTokens": 700,
    "timeoutSeconds": 180
  },
  "output": {
    "format": "conventional",
    "language": "en",
    "count": 1
  },
  "diff": {
    "maxBytes": 100000
  }
}
`

func Load(ctx context.Context, client git.Client) Config {
	cfg := Config{
		Format:       "conventional",
		Language:     "en",
		Count:        1,
		MaxDiffBytes: 100000,
		MaxTokens:    700,
		TimeoutSec:   180,
		Temperature:  0.2,
	}
	cfg.Provider = gitValue(ctx, client, "muse.provider", cfg.Provider)
	cfg.Model = gitValue(ctx, client, "muse.model", cfg.Model)
	cfg.BaseURL = gitValue(ctx, client, "muse.baseUrl", cfg.BaseURL)
	cfg.Format = gitValue(ctx, client, "muse.format", cfg.Format)
	cfg.Language = gitValue(ctx, client, "muse.language", cfg.Language)
	cfg.Count = intGitValue(ctx, client, "muse.count", cfg.Count)
	cfg.MaxDiffBytes = intGitValue(ctx, client, "muse.maxDiffBytes", cfg.MaxDiffBytes)
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
	cfg.Format = envValue("GIT_MUSE_FORMAT", cfg.Format)
	cfg.Language = envValue("GIT_MUSE_LANGUAGE", cfg.Language)
	cfg.Count = intEnvValue("GIT_MUSE_COUNT", cfg.Count)
	cfg.MaxDiffBytes = intEnvValue("GIT_MUSE_MAX_DIFF_BYTES", cfg.MaxDiffBytes)
	cfg.MaxTokens = intEnvValue("GIT_MUSE_MAX_TOKENS", cfg.MaxTokens)
	cfg.TimeoutSec = intEnvValue("GIT_MUSE_TIMEOUT_SECONDS", cfg.TimeoutSec)
	cfg.Temperature = floatEnvValue("GIT_MUSE_TEMPERATURE", cfg.Temperature)

	switch cfg.Provider {
	case "compatible", "openai", "openrouter":
		cfg.APIKeyName = apiKeyName(cfg.Provider)
		if key := strings.TrimSpace(os.Getenv(cfg.APIKeyName)); key != "" {
			cfg.APIKey = key
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
	return cfg
}

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find user home: %w", err)
	}
	return filepath.Join(home, ".git-muse.json"), nil
}

func LoadFile() (fileConfig, error) {
	path, err := Path()
	if err != nil {
		return fileConfig{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}, err
	}
	var cfg fileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fileConfig{}, fmt.Errorf("read %s: %w", path, err)
	}
	return cfg, nil
}

func InitFile(overwrite bool) (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}
	flag := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flag |= os.O_TRUNC
	} else {
		flag |= os.O_EXCL
	}
	file, err := os.OpenFile(path, flag, 0600)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.WriteString(Example); err != nil {
		return "", err
	}
	return path, nil
}

func applyFileConfig(cfg *Config, fileCfg fileConfig) {
	if fileCfg.LLM.Provider != "" {
		cfg.Provider = fileCfg.LLM.Provider
	}
	if fileCfg.LLM.Model != "" {
		cfg.Model = fileCfg.LLM.Model
	}
	if fileCfg.LLM.BaseURL != "" {
		cfg.BaseURL = fileCfg.LLM.BaseURL
	}
	if fileCfg.Output.Format != "" {
		cfg.Format = fileCfg.Output.Format
	}
	if fileCfg.Output.Language != "" {
		cfg.Language = fileCfg.Output.Language
	}
	if fileCfg.Output.Count > 0 {
		cfg.Count = fileCfg.Output.Count
	}
	if fileCfg.Diff.MaxBytes > 0 {
		cfg.MaxDiffBytes = fileCfg.Diff.MaxBytes
	}
	if fileCfg.LLM.MaxTokens > 0 {
		cfg.MaxTokens = fileCfg.LLM.MaxTokens
	}
	if fileCfg.LLM.TimeoutSec > 0 {
		cfg.TimeoutSec = fileCfg.LLM.TimeoutSec
	}
	if fileCfg.LLM.Temperature != 0 {
		cfg.Temperature = fileCfg.LLM.Temperature
	}
	if fileCfg.LLM.APIKey != "" {
		cfg.APIKey = fileCfg.LLM.APIKey
	}
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
