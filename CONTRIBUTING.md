# Contributing to Archon CLI

Thank you for considering contributing to Archon CLI. This document outlines the process for setting up your development environment and submitting changes.

## Development Environment

Archon is built using Bun and TypeScript. To get started:

1. Install Bun:
```bash
curl -fsSL https://bun.sh/install | bash
```

2. Clone the repository:
```bash
git clone https://github.com/yourusername/archon.git
cd archon
```

3. Install dependencies:
```bash
bun install
```

4. Run in development mode:
```bash
bun run dev
```

## Project Structure

- bin/: The main executable entry point.
- src/ai/: AI middleware, Output Guardian, and Ollama integration.
- src/cli/: Ink-based React UI components for the terminal.
- src/commands/: Slash commands.
- src/design/: Architecture document generation and rules engine.
- src/mcp/: Workspace context scanners.
- src/modes/: Greenfield vs Brownfield detection logic.
- src/rendering/: Markdown parsing and terminal rendering.
- src/session/: Local JSONL session persistence.

## Architecture Guidelines

We adhere to the Socratic methodology. When contributing to the AI layer (src/ai/), do not bypass the Output Guardian. Archon is an Architect, not an execution coding agent. Generating raw, un-diagrammatic code blocks is restricted by design.

## Pull Request Process

1. Fork the repository and create your branch from main.
2. Ensure your code passes the TypeScript compiler check:
```bash
bun run typecheck
```
3. Describe your changes clearly in the pull request, including the motivation and architectural impact.
4. Wait for a maintainer to review and merge your changes.

## Compiling Binaries

To test binary compilation locally before submitting:
```bash
bun run build:all
```
This will generate executables for Linux, macOS, and Windows in the dist/ directory.
