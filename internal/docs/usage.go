package docs

import (
	"html"
	"strings"
)

// Usage is the CLI help text shown by git-muse and embedded into git-muse.html.
const Usage = `git-muse - AI commit message helper

Usage:
  git muse [command]
  git-muse [command]
  git muse help
  git-muse --help
  git muse --version
  git muse --proxy http://127.0.0.1:7890
  git muse --proxy 127.0.0.1:7890

Commands:
  init       Write example config to ~/.git-muse.json
  generate   Generate commit message from staged changes
  inspect    Show context that would be sent to LLM
  doctor     Check repository and configuration
  help       Show help

Global flags:
  -h, --help              Show help (use git-muse --help or git muse help)
  -V, --version           Show version
      --proxy <url>       HTTP proxy for LLM requests (http/https targets)

Notes:
  ~/.git-muse.json llm may be one object or an array. Rate-limited (RPM/429)
  requests automatically switch to the next entry.
  git muse --help is handled by Git itself and opens git-muse.html.
  Run git muse init once to install that HTML help page into Git's doc path.
`

// HTML returns a minimal page Git can open for `git muse --help`.
func HTML() string {
	escaped := html.EscapeString(strings.TrimRight(Usage, "\n"))
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>git-muse(1)</title>
<style>
body { font-family: Consolas, "Courier New", monospace; margin: 2rem; white-space: pre-wrap; line-height: 1.4; }
</style>
</head>
<body>
` + escaped + `
</body>
</html>
`
}
