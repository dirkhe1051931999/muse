package diffctx

import (
	"path/filepath"
	"regexp"
	"strings"
)

var secretLinePattern = regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password|authorization|bearer)\s*[:=]\s*[^\s]+`)

func Filter(diff string, maxBytes int) (string, []ExcludedFile, bool) {
	if maxBytes <= 0 {
		maxBytes = 100000
	}

	var out strings.Builder
	var excluded []ExcludedFile
	var currentPath string
	skip := false

	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			currentPath = parseDiffPath(line)
			reason := excludeReason(currentPath)
			skip = reason != ""
			if skip {
				excluded = append(excluded, ExcludedFile{Path: currentPath, Reason: reason})
				continue
			}
		}
		if skip {
			continue
		}
		out.WriteString(redactLine(line))
		out.WriteByte('\n')
		if out.Len() >= maxBytes {
			return out.String()[:maxBytes], excluded, true
		}
	}
	return out.String(), excluded, false
}

func parseDiffPath(line string) string {
	parts := strings.Fields(line)
	if len(parts) < 4 {
		return ""
	}
	path := strings.TrimPrefix(parts[3], "b/")
	return strings.Trim(path, "\"")
}

func excludeReason(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	base := filepath.Base(path)
	patterns := []string{".env", ".env.*", "*.pem", "*.key", "*.crt", "*.p12", "*.min.js"}
	for _, pattern := range patterns {
		if ok, _ := filepath.Match(pattern, base); ok {
			return "secret or generated pattern"
		}
	}
	for _, prefix := range []string{"dist/", "build/", "vendor/", "node_modules/"} {
		if strings.HasPrefix(path, prefix) {
			return "generated or vendor path"
		}
	}
	return ""
}

func redactLine(line string) string {
	return secretLinePattern.ReplaceAllString(line, "$1=<redacted>")
}
