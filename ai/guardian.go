package ai

import (
	"regexp"
	"strings"
)

// forbiddenLanguages is a list of language identifiers that Archon is not allowed to generate code for.
var forbiddenLanguages = []string{
	"javascript", "js", "typescript", "ts", "jsx", "tsx",
	"python", "py", "java", "go", "golang", "c", "cpp", "c++",
	"csharp", "cs", "ruby", "rb", "php", "rust", "rs", "swift",
	"kotlin", "kt", "scala", "dart", "html", "css", "scss", "sql",
}

var codeBlockRegex = regexp.MustCompile("(?i)^```([a-z0-9+#]+)")

// ScanChunkForViolations analyzes a chunk of text (or an accumulated line) to see
// if it attempts to start a forbidden code block.
func ScanChunkForViolations(text string) bool {
	// Look for standard markdown code block starts
	matches := codeBlockRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		lang := strings.ToLower(strings.TrimSpace(matches[1]))
		
		// If it's a generic block (no language), or shell, bash, json, yaml, etc., we allow it
		// We only block specific programming languages
		for _, forbidden := range forbiddenLanguages {
			if lang == forbidden {
				return true
			}
		}
	}
	return false
}

// GuardianMessage is the strict reminder injected into the prompt
const GuardianMessage = "\n" +
	"[SOCRATIC ARCHITECT DIRECTIVE]\n" +
	"You are Archon, the Socratic AI Software Architect. Your mission is to engineer minds, not repositories.\n" +
	"1. NEVER provide direct implementation code. If asked for code, explain WHY architectural thinking precedes syntax.\n" +
	"2. ALWAYS use the Socratic method. Ask probing questions to guide the user toward their own solutions.\n" +
	"3. FOCUS EXCLUSIVELY on Systems Architecture, Scalability, Design Patterns, and Security.\n" +
	"4. You may use ```mermaid for diagrams, ```json/```yaml for config, and ```bash for commands.\n" +
	"5. DO NOT use ```typescript, ```go, ```python, or any other programming language block.\n" +
	"6. If you violate this directive, the system will intercept and terminate your response.\n"

