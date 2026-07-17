package commitmsg

import (
	"fmt"
	"strings"
)

type Candidate struct {
	Type       string   `json:"type"`
	Scope      string   `json:"scope,omitempty"`
	Subject    string   `json:"subject"`
	Body       []string `json:"body,omitempty"`
	Confidence float64  `json:"confidence,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
}

var allowedTypes = map[string]bool{
	"feat": true, "fix": true, "docs": true, "style": true, "refactor": true,
	"perf": true, "test": true, "build": true, "ci": true, "chore": true, "revert": true,
}

func Validate(c Candidate) error {
	c.Type = strings.TrimSpace(c.Type)
	c.Subject = strings.TrimSpace(c.Subject)
	if !allowedTypes[c.Type] {
		return fmt.Errorf("invalid commit type %q", c.Type)
	}
	if c.Subject == "" {
		return fmt.Errorf("commit subject is empty")
	}
	if len([]rune(c.Subject)) > 72 {
		return fmt.Errorf("commit subject exceeds 72 characters")
	}
	return nil
}

func Render(c Candidate) string {
	typ := strings.TrimSpace(c.Type)
	scope := strings.TrimSpace(c.Scope)
	subject := strings.TrimSuffix(strings.TrimSpace(c.Subject), ".")

	header := typ
	if scope != "" {
		header += "(" + scope + ")"
	}
	header += ": " + subject

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
