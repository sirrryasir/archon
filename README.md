# Archon CLI

Archon is a terminal-based software architecture evaluation tool. It analyzes project context, enforces system design principles, and utilizes a Socratic methodology to guide architectural decision-making. 

Unlike traditional AI coding assistants that generate direct code implementations, Archon strictly operates as an architectural reviewer. Its middleware intercepts and blocks raw code generation, forcing the user to focus on system boundaries, data flow, and scalability.

## Architecture & Core Mechanics

- **Socratic Constraints**: Archon refuses to write code solutions. It evaluates your prompt against the workspace context and asks probing questions regarding latency, security, and scalability.
- **Output Guardian**: Stream interception middleware detects code blocks matching programming languages (e.g., `ts`, `py`, `rs`) and aborts the AI generation stream in real-time, permitting only `diagram` blocks.
- **Context Scanner**: Uses native I/O with global byte-limit constraints (2MB limit) to prevent out-of-memory errors when scanning large codebases.
- **Rules Export**: The `/export` command compiles current architectural decisions into strict `.cursorrules` and `.clauderules` constraints to govern downstream code-generating AI agents.
- **Provider Agnostic**: Integrated with the Vercel AI SDK to support Ollama, OpenAI, Anthropic, and Google Gemini endpoints natively.

## Installation

Archon is built on the Bun runtime and can be compiled into a standalone binary.

**Prerequisites**: [Bun](https://bun.sh/) must be installed.

```bash
git clone https://github.com/sirrryasir/archon.git
cd archon
bun install
```

### Build

To compile a native executable for your system:

```bash
bun run build:all
./dist/archon
```

## Configuration

Archon uses a `.env` file for LLM provider configuration. A template is provided in the repository.

```bash
cp .env.example .env
```

Edit the `.env` file to select your provider (`ARCHON_PROVIDER`) and input your API keys.

### Recommended Models

Archon relies heavily on advanced reasoning capabilities to analyze architecture rather than generating code. 

**High-Performance/Cloud Targets:**
- `qwen3-coder:480b-cloud` (Default)
- `deepseek-v4-pro:cloud`
- `gpt-oss:120b-cloud`

**Local Mid-Range Targets:**
- `qwen2.5:32b`
- `qwen3:14b`

## Usage

Start the interactive terminal interface:

```bash
bun run dev
# or
archon
```

Run a single-shot architectural evaluation:

```bash
archon prompt "How should I structure a React micro-frontend?"
```

Trigger an automated workspace review:

```bash
archon review
```

### Interactive Commands

When running the interactive CLI, the following slash commands are supported:

| Command | Description |
|---------|-------------|
| `/export` | Generates architecture documentation and downstream AI rules (`.cursorrules`). |
| `/review` | Triggers a full architectural review of the current workspace state. |
| `/clear` | Clears the terminal buffer. |
| `/exit` | Terminates the current session. |

## Contributing

Review `CONTRIBUTING.md` and `CODE_OF_CONDUCT.md` for guidelines on submitting pull requests and reporting issues.

## License

This project is licensed under the MIT License.
