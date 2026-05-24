# AGENTS.md — Archon Agent Interface

This file describes how other AI agents and automated systems can use **Archon** as a tool.

---

## Overview

Archon is a **Socratic AI Software Architect** CLI. It helps design and critique system architecture through guided questioning — it deliberately does NOT write code. Instead, it helps you think through:

- System decomposition and service boundaries  
- Data flow and consistency tradeoffs
- Failure modes and resilience strategies
- API contracts and integration patterns

---

## Non-Interactive Mode (Agent-to-Agent)

Use `archon prompt` to interact without a TUI:

```bash
archon prompt "How should we structure the auth service boundary?"
```

### Stdin / Stdout Contract

| Stream | Content |
|--------|---------|
| `stdout` | The AI response (Markdown) |
| `stderr` | Status messages: `[ Archon ] Thinking...` |

This separation makes it safe to pipe stdout to other tools:

```bash
# Capture only the response
archon prompt "Review my API design" 2>/dev/null > review.md

# Pipe to another tool
archon prompt "Analyze this service diagram" | pandoc -o review.pdf

# Use in a multi-agent pipeline
ANALYSIS=$(archon prompt "What are the weak points of this design?")
echo "$ANALYSIS" | another-agent summarize
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--timeout` | `3m` | Response timeout (e.g. `30s`, `5m`, `10m`) |
| `--provider` | from config | Override AI provider |
| `--model` | from config | Override model |

### Example Pipeline

```bash
# Get architecture review with a 2-minute timeout
archon prompt --timeout=2m "Review the service boundaries in this monorepo" 2>/dev/null

# Use with a small fast local model
archon prompt --provider=ollama --model=qwen2.5:3b "Is this design over-engineered?"
```

---

## Configuration

Archon resolves configuration in this order (later overrides earlier):

1. Defaults (`ollama` / `qwen3-coder:480b-cloud`)
2. Global file: `~/.archon/config.json`
3. Local file: `.archon/config.json` (project-specific)
4. Environment variables: `ARCHON_PROVIDER`, `ARCHON_MODEL`
5. CLI flags: `--provider`, `--model`

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ARCHON_PROVIDER` | Provider: `ollama`, `anthropic`, `openai`, `google` |
| `ARCHON_MODEL` | Model name |
| `ARCHON_MOCK_RESPONSE` | **Testing only** — returns this string as the response |
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `GOOGLE_GENERATIVE_AI_API_KEY` | Google AI API key |
| `OLLAMA_URL` | Ollama base URL (default: `http://localhost:11434`) |

---

## Testing / Offline Use

Use `ARCHON_MOCK_RESPONSE` to test agent integrations without a real LLM:

```bash
ARCHON_MOCK_RESPONSE="Consider breaking this into 3 services..." archon prompt "Review design"
# Returns the mock response immediately, no network needed
```

---

## Health Check

Before using Archon in a pipeline, verify connectivity:

```bash
archon doctor
```

Exit codes:
- `0` — Ready
- `1` — Configuration or connectivity error

---

## Commands

| Command | Description |
|---------|-------------|
| `archon` | Start interactive TUI chat session |
| `archon chat` | Explicit TUI mode |
| `archon prompt [msg]` | Non-interactive single prompt |
| `archon init` | Interactive setup wizard |
| `archon doctor` | Verify connectivity and config |
| `archon --version` | Print version |

---

## The Guardian

Archon enforces a **no-implementation-code** rule. Responses containing code blocks with implementation languages (Go, TypeScript, Python, etc.) are intercepted and replaced with architectural guidance. This is by design — Archon is an architect, not a code generator.

```
> ⚠️ [GUARDIAN INTERCEPTED]: I don't write implementation code.
  Let's focus on the architecture. What problem are you actually trying to solve?
```

---

## Context Awareness

When running in any directory, Archon automatically scans the workspace and injects:
- File/directory structure (up to depth 4)
- Git branch and recent commit messages
- Project mode detection (Greenfield vs Brownfield)

This means a prompt like `"Review my service design"` will have full project context without manual file attachment.
