package design

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirrryasir/archon/ai"
)

// DesignDocs holds the generated markdown contents for each blueprint document.
type DesignDocs struct {
	Architecture string
	Design       string
	System       string
}

const ArchitectureDocPrompt = `Based on our conversation, generate a comprehensive ARCHITECTURE.md document.

FORMAT REQUIREMENTS:
- Start with "# Architecture" as the title
- Include a high-level system overview paragraph
- Include a System Diagram using Mermaid (graph TD or flowchart)
- List all Components with: name, responsibility, technology, dependencies
- Include a Data Flow section describing how data moves through the system
- Include an API Surface section listing key endpoints/interfaces
- Include a Non-Functional Requirements section (scalability, performance, security)

This document should be readable by:
1. A new engineer joining the team
2. An AI agent that needs to understand the system to implement it

Write ONLY the markdown document content. No preamble or explanation.`

const DesignDocPrompt = `Based on our conversation, generate a comprehensive DESIGN.md document.

FORMAT REQUIREMENTS:
- Start with "# Technical Design" as the title
- Include a Decision Log table with columns: Decision | Options Considered | Chosen | Reasoning
- Include a Patterns & Principles section listing each pattern used and why
- Include a Security Model section
- Include a Scaling Strategy section
- Include an Error Handling Strategy section
- Include a Testing Strategy section
- Include a Technical Debt & Risks section

This document captures the WHY behind architectural decisions.

Write ONLY the markdown document content. No preamble or explanation.`

const SystemDocPrompt = `Based on our conversation, generate a comprehensive SYSTEM.md document.

FORMAT REQUIREMENTS:
- Start with "# System Specification" as the title
- Include a Stack section listing: runtime, framework, database, auth, etc.
- Include a Directory Structure section with a proposed file tree using ASCII
- Include a Module Contracts section: for each module list inputs, outputs, side effects
- Include a Database Schema section (if applicable) using markdown tables
- Include an Environment Variables section listing each var, its purpose, and example value
- Include a Build & Deploy section with commands
- Include an Implementation Priority section ordering what to build first

This document should be detailed enough that an AI coding agent could read it and implement the entire system without asking clarifying questions.

Write ONLY the markdown document content. No preamble or explanation.`

// ExtractConversationContext condenses user and assistant messages for documentation generation.
func ExtractConversationContext(messages []ai.Message) string {
	var parts []string
	for _, m := range messages {
		if m.Role == "user" || m.Role == "assistant" {
			parts = append(parts, fmt.Sprintf("[%s]: %s", strings.ToUpper(m.Role), m.Content))
		}
	}
	return strings.Join(parts, "\n\n")
}

// BuildDesignPrompt constructs the AI prompt to generate a specific design document.
func BuildDesignPrompt(docType string, conversationContext string) string {
	var prompt string
	switch docType {
	case "architecture":
		prompt = ArchitectureDocPrompt
	case "design":
		prompt = DesignDocPrompt
	case "system":
		prompt = SystemDocPrompt
	}

	return fmt.Sprintf(`You are Archon, an expert software architect generating design documentation.

Here is the full architectural conversation that has taken place:

═══════════════════════════════════════════════
CONVERSATION:
═══════════════════════════════════════════════
%s

═══════════════════════════════════════════════
TASK:
═══════════════════════════════════════════════
%s`, conversationContext, prompt)
}

// ExportDesignDocs writes design documents and agent instructions to target directory.
func ExportDesignDocs(docs DesignDocs, outputDir string) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	cursorRulesContent := `# Archon Blueprint Directives
You are an Execution Agent / Junior Developer. Your task is to implement the system defined in this directory.

## CRITICAL RULES:
1. **Strict Adherence:** You MUST strictly adhere to the technical guidelines, component boundaries, and technology stack laid out in ` + "`ARCHITECTURE.md`, `DESIGN.md`, and `SYSTEM.md`." + `
2. **No Re-architecting:** Do not make high-level architectural decisions, introduce new design patterns, or fundamentally alter the data flow without explicit permission.
3. **Module Contracts:** Respect the module contracts and constraints defined in ` + "`SYSTEM.md`." + `
4. **Escalation:** If the implementation details are ambiguous or contradict best practices, STOP and ask the user to consult Archon for clarification instead of hallucinating a structure.

Archon has spoken. Execute flawlessly.
`

	files := []struct {
		name    string
		content string
	}{
		{"ARCHITECTURE.md", docs.Architecture},
		{"DESIGN.md", docs.Design},
		{"SYSTEM.md", docs.System},
		{"AGENTS.md", cursorRulesContent},
	}

	var written []string
	for _, file := range files {
		if strings.TrimSpace(file.content) != "" {
			fullPath := filepath.Join(outputDir, file.name)
			if err := os.WriteFile(fullPath, []byte(file.content), 0644); err != nil {
				return nil, err
			}
			written = append(written, fullPath)
		}
	}

	return written, nil
}
