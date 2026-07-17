package prompt

import (
	"fmt"
	"strings"

	"git-muse/internal/config"
	"git-muse/internal/diffctx"
)

func System() string {
	return `You generate Git commit messages from staged changes.
Return strict JSON only. Do not use Markdown.
JSON shape: {"candidates":[{"type":"feat","scope":"auth","subject":"add refresh token rotation","body":["optional body line"],"confidence":0.91,"warnings":[]}]}
Use Conventional Commits.
Allowed types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.
Subject must be non-empty and <= 72 characters.
Use imperative mood for English.
Do not invent changes not present in diff.`
}

func User(cfg config.Config, data diffctx.Context) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Format: %s\n", cfg.Format)
	fmt.Fprintf(&b, "Language: %s\n", cfg.Language)
	fmt.Fprintf(&b, "Candidate count: %d\n\n", cfg.Count)
	fmt.Fprintf(&b, "Branch:\n%s\n\n", data.Branch)
	fmt.Fprintf(&b, "Recent commits:\n%s\n\n", strings.Join(data.RecentSubjects, "\n"))
	fmt.Fprintf(&b, "Status:\n%s\n\n", data.Status)
	fmt.Fprintf(&b, "Stat:\n%s\n\n", data.Stat)
	fmt.Fprintf(&b, "Staged diff:\n%s\n", data.Diff)
	return b.String()
}
