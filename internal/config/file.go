package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type LLMFile struct {
	Provider    string  `json:"provider"`
	BaseURL     string  `json:"baseUrl"`
	Model       string  `json:"model"`
	APIKey      string  `json:"apiKey"`
	Proxy       string  `json:"proxy"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
	TimeoutSec  int     `json:"timeoutSeconds"`
}

type LLMFileList []LLMFile

func (l *LLMFileList) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*l = nil
		return nil
	}
	if data[0] == '[' {
		var items []LLMFile
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		*l = items
		return nil
	}
	var item LLMFile
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}
	*l = LLMFileList{item}
	return nil
}

type fileConfig struct {
	LLM    LLMFileList `json:"llm"`
	Output struct {
		Format   string `json:"format"`
		Language string `json:"language"`
		Count    int    `json:"count"`
	} `json:"output"`
	Diff struct {
		MaxBytes int      `json:"maxBytes"`
		Ignore   []string `json:"ignore"`
		Allow    []string `json:"allow"`
	} `json:"diff"`
}

const Example = `{
  "llm": [
    {
      "provider": "compatible",
      "baseUrl": "https://token.sensenova.cn/v1/chat/completions",
      "model": "deepseek-v4-flash",
      "apiKey": "sk-keykeykey...keykeykey",
      "proxy": "",
      "temperature": 0.2,
      "maxTokens": 2048,
      "timeoutSeconds": 180
    },
    {
      "provider": "compatible",
      "baseUrl": "https://api.openai.com/v1",
      "model": "gpt-4.1-mini",
      "apiKey": "sk-another-key"
    }
  ],
  "output": {
    "format": "conventional",
    "language": "en",
    "count": 1
  },
  "diff": {
    "maxBytes": 100000,
    "ignore": [],
    "allow": []
  }
}
`

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
	if len(fileCfg.LLM) > 0 {
		applyLLMFile(cfg, fileCfg.LLM[0], true)
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
	if fileCfg.Diff.Ignore != nil {
		cfg.IgnoreGlobs = append([]string(nil), fileCfg.Diff.Ignore...)
	}
	if fileCfg.Diff.Allow != nil {
		cfg.AllowGlobs = append([]string(nil), fileCfg.Diff.Allow...)
	}
	if len(fileCfg.LLM) <= 1 {
		cfg.Fallbacks = nil
		return
	}
	cfg.Fallbacks = make([]Config, 0, len(fileCfg.LLM)-1)
	for _, item := range fileCfg.LLM[1:] {
		fb := fallbackBase(*cfg)
		applyLLMFile(&fb, item, false)
		cfg.Fallbacks = append(cfg.Fallbacks, fb)
	}
}

func fallbackBase(primary Config) Config {
	return Config{
		Provider:    primary.Provider,
		Model:       primary.Model,
		BaseURL:     primary.BaseURL,
		Proxy:       primary.Proxy,
		MaxTokens:   primary.MaxTokens,
		TimeoutSec:  primary.TimeoutSec,
		Temperature: primary.Temperature,
	}
}

func applyLLMFile(cfg *Config, item LLMFile, copyAPIKey bool) {
	if item.Provider != "" {
		cfg.Provider = item.Provider
	}
	if item.Model != "" {
		cfg.Model = item.Model
	}
	if item.BaseURL != "" {
		cfg.BaseURL = item.BaseURL
	}
	if item.Proxy != "" {
		cfg.Proxy = item.Proxy
	}
	if item.MaxTokens > 0 {
		cfg.MaxTokens = item.MaxTokens
	}
	if item.TimeoutSec > 0 {
		cfg.TimeoutSec = item.TimeoutSec
	}
	if item.Temperature != 0 {
		cfg.Temperature = item.Temperature
	}
	if copyAPIKey {
		if item.APIKey != "" {
			cfg.APIKey = item.APIKey
		}
		return
	}
	cfg.APIKey = item.APIKey
}
