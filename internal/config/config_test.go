package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateProviderRequiresProvider(t *testing.T) {
	if err := (Config{}).ValidateProvider(); err == nil {
		t.Fatalf("expected missing provider error")
	}
}

func TestValidateProviderAllowsOllamaWithoutKey(t *testing.T) {
	cfg := Config{Provider: "ollama", Model: "qwen2.5-coder"}
	if err := cfg.ValidateProvider(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateProviderRequiresKeyForCompatible(t *testing.T) {
	cfg := Config{Provider: "compatible", Model: "gpt", APIKeyName: "OPENAI_API_KEY"}
	if err := cfg.ValidateProvider(); err == nil {
		t.Fatalf("expected missing API key error")
	}
}

func TestLoadFileReadsJSONConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))
	path, err := Path()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	data := `{"llm":{"provider":"compatible","baseUrl":"https://token.sensenova.cn/v1/chat/completions","model":"deepseek-v4-flash","apiKey":"sk-test","maxTokens":900,"timeoutSeconds":240},"output":{"format":"conventional","language":"zh","count":2},"diff":{"maxBytes":200000}}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := LoadFile()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.LLM.Provider != "compatible" || cfg.LLM.Model != "deepseek-v4-flash" || cfg.LLM.APIKey != "sk-test" {
		t.Fatalf("unexpected llm config: %+v", cfg.LLM)
	}
	if cfg.Output.Language != "zh" || cfg.Output.Count != 2 || cfg.Diff.MaxBytes != 200000 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestInitFileCreatesExample(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	path, err := InitFile(false)
	if err != nil {
		t.Fatalf("init config: %v", err)
	}
	if path != filepath.Join(home, ".git-muse.json") {
		t.Fatalf("unexpected path: %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if _, err := InitFile(true); err != nil {
		t.Fatalf("overwrite config: %v", err)
	}
}
