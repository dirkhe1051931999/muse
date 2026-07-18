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
	data := `{"llm":{"provider":"compatible","baseUrl":"https://token.sensenova.cn/v1/chat/completions","model":"deepseek-v4-flash","apiKey":"sk-test","maxTokens":900,"timeoutSeconds":240},"output":{"format":"conventional","language":"zh","count":2},"diff":{"maxBytes":200000,"ignore":["*.lock"],"allow":[".env.example"]}}`
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
	if len(cfg.Diff.Ignore) != 1 || cfg.Diff.Ignore[0] != "*.lock" {
		t.Fatalf("unexpected ignore: %#v", cfg.Diff.Ignore)
	}
	if len(cfg.Diff.Allow) != 1 || cfg.Diff.Allow[0] != ".env.example" {
		t.Fatalf("unexpected allow: %#v", cfg.Diff.Allow)
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

func TestNormalizeLanguage(t *testing.T) {
	got, err := normalizeLanguage("ZH")
	if err != nil || got != "zh" {
		t.Fatalf("normalizeLanguage(ZH) = %q, %v", got, err)
	}
	got, err = normalizeLanguage("en")
	if err != nil || got != "en" {
		t.Fatalf("normalizeLanguage(en) = %q, %v", got, err)
	}
	if _, err := normalizeLanguage("fr"); err == nil {
		t.Fatalf("expected unsupported language error")
	}
}

func TestApplyFileConfigDiffGlobs(t *testing.T) {
	var fileCfg fileConfig
	fileCfg.Diff.MaxBytes = 123
	fileCfg.Diff.Ignore = []string{"*.lock"}
	fileCfg.Diff.Allow = []string{".env.example"}
	cfg := Config{Language: "en"}
	applyFileConfig(&cfg, fileCfg)
	if cfg.MaxDiffBytes != 123 {
		t.Fatalf("MaxDiffBytes = %d", cfg.MaxDiffBytes)
	}
	if len(cfg.IgnoreGlobs) != 1 || cfg.IgnoreGlobs[0] != "*.lock" {
		t.Fatalf("IgnoreGlobs = %#v", cfg.IgnoreGlobs)
	}
	if len(cfg.AllowGlobs) != 1 || cfg.AllowGlobs[0] != ".env.example" {
		t.Fatalf("AllowGlobs = %#v", cfg.AllowGlobs)
	}
}

func TestSplitCSV(t *testing.T) {
	got := splitCSV(" *.lock , coverage/* , ")
	if len(got) != 2 || got[0] != "*.lock" || got[1] != "coverage/*" {
		t.Fatalf("splitCSV = %#v", got)
	}
}

func TestNormalizeProxy(t *testing.T) {
	got, err := NormalizeProxy("127.0.0.1:7890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "http://127.0.0.1:7890" {
		t.Fatalf("got %q", got)
	}
	got, err = NormalizeProxy("http://127.0.0.1:7890")
	if err != nil || got != "http://127.0.0.1:7890" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if _, err := NormalizeProxy("ftp://127.0.0.1:7890"); err == nil {
		t.Fatalf("expected scheme error")
	}
	got, err = NormalizeProxy("")
	if err != nil || got != "" {
		t.Fatalf("empty proxy: %q %v", got, err)
	}
}
