package commitmsg

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Candidate struct {
	Type       string   `json:"type"`
	Scope      string   `json:"scope,omitempty"`
	Subject    string   `json:"subject"`
	Body       []string `json:"body,omitempty"`
	Confidence float64  `json:"confidence,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
}

const maxSubjectRunes = 72

var allowedTypes = map[string]bool{
	"feat": true, "fix": true, "docs": true, "style": true, "refactor": true,
	"perf": true, "test": true, "build": true, "ci": true, "chore": true, "revert": true,
}

// Normalize trims fields and clamps subject to maxSubjectRunes.
func Normalize(c Candidate) Candidate {
	c.Type = strings.TrimSpace(c.Type)
	c.Scope = strings.TrimSpace(c.Scope)
	c.Subject = strings.TrimSuffix(strings.TrimSpace(c.Subject), ".")
	c.Subject = clampSubject(c.Subject, maxSubjectRunes)
	return c
}

func Validate(c Candidate) error {
	c = Normalize(c)
	if !allowedTypes[c.Type] {
		return fmt.Errorf("invalid commit type %q", c.Type)
	}
	if c.Subject == "" {
		return fmt.Errorf("commit subject is empty")
	}
	return nil
}

func Render(c Candidate) string {
	c = Normalize(c)

	header := c.Type
	if c.Scope != "" {
		header += "(" + c.Scope + ")"
	}
	header += ": " + c.Subject

	if len(c.Body) == 0 {
		return header + "\n"
	}

	lines := make([]string, 0, len(c.Body))
	for _, line := range c.Body {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return header + "\n"
	}
	return header + "\n\n" + strings.Join(lines, "\n") + "\n"
}

func clampSubject(subject string, max int) string {
	if max <= 0 || utf8.RuneCountInString(subject) <= max {
		return subject
	}
	runes := []rune(subject)[:max]
	trimmed := strings.TrimSpace(string(runes))
	// Prefer cutting at the last space for English subjects when enough text remains.
	if i := strings.LastIndex(trimmed, " "); i >= max/2 {
		return strings.TrimSpace(trimmed[:i])
	}
	return trimmed
}
