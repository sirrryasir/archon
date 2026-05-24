# Contributing to Archon CLI

Thank you for your interest in contributing to Archon! This document outlines the process for setting up your Go development environment, running tests, and submitting pull requests.

---

## Development Environment

Archon is built using Go. To get started, make sure you have [Go](https://go.dev/) (version 1.20 or later) installed on your system.

### 1. Clone the Repository
```bash
git clone https://github.com/sirrryasir/archon.git
cd archon
```

### 2. Download Dependencies
```bash
go mod download
```

### 3. Run in Development Mode
You can compile and run Archon directly using:
```bash
go run main.go chat
# or with prompt arguments
go run main.go prompt "How should I structure my database layer?"
```

### 4. Build the Executable
```bash
make build
# This generates the native binary at ./bin/archon
```

---

## Project Structure

- `main.go`: The main entry point of the CLI application.
- `cmd/`: Cobra CLI commands (e.g. `rootCmd`, `chatCmd`, `promptCmd`, `initCmd`, `doctorCmd`).
- `ai/`: Providers (`anthropic`, `google`, `ollama`, `openai`), model estimation, aliases, and the streaming **Output Guardian** middleware.
- `tui/`: Bubble Tea TUI chat loop, input, styles, and command integrations.
- `mcp/`: Workspace scanner, git metadata builder, and directory mapping functions.
- `modes/`: Greenfield vs. Brownfield mode detector and prompt templates.
- `design/`: Markdown blueprints generator (`ARCHITECTURE.md`, `DESIGN.md`, `SYSTEM.md`) and downstream instructions (`AGENTS.md`).
- `config/`: Configuration cascade loading using Viper.
- `session/`: Conversational history database mapping and management.
- `scripts/`: Integration test suites and build automation.

---

## Code Quality & Testing

Before submitting a pull request, ensure your code builds and passes all tests.

### Run Unit Tests
```bash
make test
```

### Run Integration & E2E Tests
Make sure Ollama is running and you have internet connectivity to verify the end-to-end piping flow:
```bash
make e2e
```

---

## Architectural Guidelines

We adhere strictly to the **Socratic Architect** design:
1. **No-Implementation-Code**: Do not bypass or weaken the Output Guardian in `ai/guardian.go`. Archon is designed to be an architect, not an execution coding tool. Code blocks containing syntax fences (`go`, `ts`, `py`, etc.) are actively blocked.
2. **Socratic Dialog**: Answers must prioritize questions that guide the user's thinking rather than doing the thinking for them.
3. **No Unwanted Logs**: Keep the console output of the `prompt` command clean of diagnostic logs on `stdout`. Always pipe warnings/thinking/diagnostic text to `stderr`.

---

## Pull Request Process

1. Fork the repository and create your branch from `main`.
2. Commit your changes with clear, descriptive commit messages.
3. Run `make test` and `make e2e` to verify your changes.
4. Open a pull request describing the motivation, implementation, and architectural impact of your change.

---

## Code of Conduct

Please be respectful and inclusive in your interactions with other contributors. We aim to foster an open and welcoming community. For detailed guidelines, please read our [Code of Conduct](CODE_OF_CONDUCT.md).

