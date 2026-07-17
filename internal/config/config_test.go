package config

import (
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
