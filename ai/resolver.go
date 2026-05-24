package ai

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var pathRegex = regexp.MustCompile(`'(/[a-zA-Z0-9_\-\./]+)'|"(/[a-zA-Z0-9_\-\./]+)"|(/[a-zA-Z0-9_\-\./]+)`)
var secretRegex = regexp.MustCompile(`(?i)(?:sk-|AIza|ghp_|gho_|ghu_|ghs_|ghr_|SECRET|PASSWORD|TOKEN|KEY|PASS)[\w-]{10,}`)

// ResolvePromptFiles parses a user prompt, detects absolute paths of existing files,
// reads their contents, and appends them to the prompt as structured context.
func ResolvePromptFiles(prompt string) string {
	matches := pathRegex.FindAllStringSubmatch(prompt, -1)
	if len(matches) == 0 {
		return prompt
	}

	var sb strings.Builder
	sb.WriteString(prompt)

	resolvedAny := false
	seen := map[string]bool{}

	for _, match := range matches {
		var path string
		if match[1] != "" {
			path = match[1]
		} else if match[2] != "" {
			path = match[2]
		} else {
			path = match[3]
		}

		path = strings.TrimSpace(path)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true

		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			// Limit to 500KB to avoid excessive token size
			if info.Size() < 500*1024 {
				raw, err := os.ReadFile(path)
				if err == nil {
					if !resolvedAny {
						sb.WriteString("\n\n=== INJECTED FILE CONTEXT ===")
						resolvedAny = true
					}
					// Redact any secrets before injecting
					content := secretRegex.ReplaceAllString(string(raw), "[REDACTED]")
					sb.WriteString(fmt.Sprintf("\n\n### File: %s\n```\n%s\n```", path, content))
				}
			}
		}
	}

	if resolvedAny {
		sb.WriteString("\n=== END INJECTED FILE CONTEXT ===")
		return sb.String()
	}

	return prompt
}
