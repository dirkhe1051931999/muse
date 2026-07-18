package prompt

import (
	"strings"
	"testing"

	"git-muse/internal/config"
	"git-muse/internal/diffctx"
)

func TestSystemEnglish(t *testing.T) {
	got := System(config.Config{Language: "en"})
	if !strings.Contains(got, "Write subject and body in English.") {
		t.Fatalf("missing English instruction: %s", got)
	}
	if !strings.Contains(got, "Use imperative mood for English.") {
		t.Fatalf("missing imperative rule: %s", got)
	}
	if strings.Contains(got, "Simplified Chinese") {
		t.Fatalf("unexpected Chinese instruction: %s", got)
	}
}

func TestSystemChinese(t *testing.T) {
	got := System(config.Config{Language: "zh"})
	if !strings.Contains(got, "Write subject and body in Simplified Chinese.") {
		t.Fatalf("missing Chinese instruction: %s", got)
	}
	if strings.Contains(got, "Use imperative mood for English.") {
		t.Fatalf("English imperative should not appear for zh: %s", got)
	}
}

func TestUserLanguageDirective(t *testing.T) {
	data := diffctx.Context{Branch: "main", Diff: "+x"}
	en := User(config.Config{Language: "en", Format: "conventional", Count: 1}, data)
	if !strings.Contains(en, "Write subject and body in English.") {
		t.Fatalf("missing English user directive: %s", en)
	}
	zh := User(config.Config{Language: "zh", Format: "conventional", Count: 1}, data)
	if !strings.Contains(zh, "Write subject and body in Simplified Chinese.") {
		t.Fatalf("missing Chinese user directive: %s", zh)
	}
}
