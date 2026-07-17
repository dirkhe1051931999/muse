package commitmsg

import "strings"

type Candidate struct {
	Type    string
	Scope   string
	Subject string
	Body    []string
}

func Render(c Candidate) string {
	header := c.Type
	if c.Scope != "" {
		header += "(" + c.Scope + ")"
	}
	header += ": " + c.Subject

	if len(c.Body) == 0 {
		return header + "\n"
	}
	return header + "\n\n" + strings.Join(c.Body, "\n") + "\n"
}
