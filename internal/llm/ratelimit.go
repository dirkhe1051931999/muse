package llm

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"
)

type RateLimitError struct {
	Status string
	Body   string
}

func (e *RateLimitError) Error() string {
	if e == nil {
		return "LLM rate limited"
	}
	body := strings.TrimSpace(e.Body)
	if body == "" {
		return fmt.Sprintf("LLM rate limited: %s", e.Status)
	}
	return fmt.Sprintf("LLM rate limited: %s: %s", e.Status, body)
}

func IsRateLimited(err error) bool {
	var rl *RateLimitError
	if errors.As(err, &rl) {
		return true
	}
	if err == nil {
		return false
	}
	return looksLikeRateLimit(0, err.Error())
}

func rateLimitError(statusCode int, status string, body []byte) error {
	snippet := strings.TrimSpace(string(body))
	if len(snippet) > 400 {
		snippet = snippet[:400] + "..."
	}
	if !looksLikeRateLimit(statusCode, snippet) {
		return nil
	}
	return &RateLimitError{Status: status, Body: snippet}
}

func looksLikeRateLimit(statusCode int, text string) bool {
	if statusCode == http.StatusTooManyRequests || statusCode == 529 {
		return true
	}
	lower := strings.ToLower(text)
	needles := []string{
		"rate_limit",
		"ratelimit",
		"rate limit",
		"too many requests",
		"requests per minute",
		"rpm limit",
		"rpm exceeded",
		"exceeded rpm",
		"限流",
		"请求过于频繁",
		"频率过快",
	}
	for _, needle := range needles {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return containsRPMToken(lower)
}

func containsRPMToken(lower string) bool {
	for idx := 0; ; {
		i := strings.Index(lower[idx:], "rpm")
		if i < 0 {
			return false
		}
		i += idx
		beforeOK := i == 0 || !isWordChar(rune(lower[i-1]))
		after := i + 3
		afterOK := after >= len(lower) || !isWordChar(rune(lower[after]))
		if beforeOK && afterOK {
			return true
		}
		idx = i + 1
	}
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
