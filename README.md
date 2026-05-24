<div align="center">
<pre>
 █████╗ ██████╗  ██████╗██╗  ██╗ ██████╗ ███╗   ██╗
██╔══██╗██╔══██╗██╔════╝██║  ██║██╔═══██╗████╗  ██║
███████║██████╔╝██║     ███████║██║   ██║██╔██╗ ██║
██╔══██║██╔══██╗██║     ██╔══██║██║   ██║██║╚██╗██║
██║  ██║██║  ██║╚██████╗██║  ██║╚██████╔╝██║ ╚████║
╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
</pre>
</div>

<p align="center">A Socratic AI Software Architect CLI for evaluated, constraint-driven systems engineering.</p>

---

## Elevator Pitch

Minor rant incoming: You ask an AI coding assistant to help design your new authentication service. Before you can blink, it spits out 400 lines of buggy boilerplate code using a database library you don't even use. You spend the next two hours debugging syntax errors, only to realize the fundamental architecture is flawed—it's a tight, synchronous monolith coupling that will fail the moment you introduce scale.

What a headache!

We have enough tools that write code. We don't have enough tools that **think**. 

Archon is designed to solve this. It is a strict **Socratic Software Architect** that runs directly in your terminal. **It deliberately refuses to write implementation code.** Instead, it scans your codebase, maps your dependencies, and engages you in Socratic dialogue—asking tough, probing questions about service boundaries, data consistency, scalability, and security. 

When you agree on a design, Archon generates architectural markdown blueprints and outputs a unified `AGENTS.md` spec file. You can then feed this spec directly to downstream code-generating agents (like Cursor, Claude Code, or Aider) to build it right the first time.

---

## Table of Contents

- [Elevator Pitch](#elevator-pitch)
- [Quickstart (For Absolute Beginners)](#quickstart-for-absolute-beginners)
- [Recommended Models Explained](#recommended-models-explained)
- [Installation](#installation)
- [Configuration Cascade](#configuration-cascade)
- [Command Reference](#command-reference)
- [Interactive Slash Commands](#interactive-slash-commands)
- [Testing](#testing)
- [Contributing](#contributing)
- [Alternatives](#alternatives)
- [License](#license)

---

## Quickstart (For Absolute Beginners)

If you have never touched Go, Ollama, or API keys, follow this step-by-step guide to get Archon running in 5 minutes.

### Step 1: Install Go
Go to [go.dev/dl](https://go.dev/dl/) and download the installer for your operating system (macOS, Windows, or Linux). Open the installer and click through the setup.

### Step 2: Download & Build Archon
Open your terminal application (Terminal on macOS, PowerShell/Command Prompt on Windows) and run these commands one by one:

```bash
# 1. Clone this codebase
git clone https://github.com/sirrryasir/archon.git
cd archon

# 2. Build the executable binary
go build -o bin/archon .
```

### Step 3: Configure Your Environment
You can either run the interactive configuration wizard:
```bash
./bin/archon init
```
Or configure it manually by copying the template file:
```bash
cp .env.example .env
```
Open `.env` in your editor and enter your API keys.

### Step 4: Install Globally (Optional)
To run `archon` from any directory on your computer:
```bash
sudo cp bin/archon /usr/local/bin/
```

### Step 5: Start Chatting!
Verify your configuration and start the interactive terminal session:
```bash
# Run doctor check to verify connectivity
archon doctor

# Start the Socratic Architect TUI
archon
```

---

## Recommended Models Explained

Archon does not write code, but it needs an AI model with deep reasoning capabilities to analyze systems architecture. Here is a beginner-friendly breakdown of how to configure models.

### Option A: Using Ollama Cloud (Recommended)
If you want to use state-of-the-art open-source models hosted in the cloud for free or low-cost:
1. When running `archon init`, choose **Ollama** as your provider.
2. Select `qwen3-coder:480b-cloud` (Default) or `deepseek-v4-pro:cloud`.
3. These models are incredibly smart and run instantly in the cloud, giving you top-tier architectural insights.

### Option B: Using Commercial Cloud APIs (Paid)
If you have an API key from Anthropic, OpenAI, or Google:
1. Run `archon init`.
2. Choose your provider (e.g. **Anthropic** or **OpenAI**).
3. Paste in your API key when prompted.
4. Recommended models:
   - **Anthropic**: `claude-3-5-sonnet` (Outstanding architectural analysis).
   - **OpenAI**: `gpt-4o` or `o1`.
   - **Google**: `gemini-1.5-pro`.

### Option C: Running 100% Locally (Private & Offline)
If you want to run models locally on your own computer without any API keys or network requests:
1. Download and install [Ollama](https://ollama.com/).
2. Run this command in your terminal to download a local model:
   ```bash
   ollama run qwen2.5:3b
   ```
3. Run `archon init`, select **Ollama** as your provider, and type `qwen2.5:3b` as the model.
4. *Note: Running locally on CPU can be slow depending on your computer's specs.*

---

## Installation

For power users, developers, and packaging systems:

### Homebrew (macOS & Linux)
```bash
brew install sirrryasir/tap/archon
```

### Go Package Manager
Ensure `~/go/bin` is in your `$PATH`:
```bash
go install github.com/sirrryasir/archon@latest
```

### Manual Build
```bash
make build
make install
```

---

## Configuration Cascade

Archon searches for settings in this order (later items override earlier ones):

1. **Defaults**: Providers default to `ollama` with `qwen3-coder:480b-cloud`.
2. **Global Config**: `~/.archon/config.json`.
3. **Local Config**: `.archon/config.json` inside the directory where you launch the command. This lets you override models/keys per-project.
4. **Environment Variables**:
   - `ARCHON_PROVIDER` (e.g. `anthropic`, `openai`, `google`, `ollama`)
   - `ARCHON_MODEL` (model string override)
   - `ANTHROPIC_API_KEY`
   - `OPENAI_API_KEY`
   - `GOOGLE_GENERATIVE_AI_API_KEY`
   - `OLLAMA_URL` (defaults to `http://localhost:11434`)
5. **CLI Flags**:
   - `--model` / `-m` (e.g. `sonnet`, `deepseek`)
   - `--provider` / `-p` (e.g. `anthropic`, `ollama`)

---

## Command Reference

| Command | Args | Description |
|---------|------|-------------|
| `archon` | None | Launches the interactive Bubble Tea TUI chat loop (default). |
| `archon chat` | None | Explicitly starts the TUI chat. |
| `archon prompt` | `[message]` | Sends a single prompt and streams the response directly to `stdout`. |
| `archon init` | None | Launches the interactive configuration setup wizard. |
| `archon doctor` | None | Runs diagnostic and connectivity checks on your configurations. |
| `archon --version` | None | Prints the current release version. |

### Prompt Flags
When using the non-interactive `prompt` mode:
* `--timeout` (Default: `3m`): Sets the maximum duration before canceling the request. Supports shorthand (e.g. `--timeout=30s`, `--timeout=5m`).

---

## Interactive Slash Commands

Type `/` within the interactive TUI to open the command autocompletion menu.

* **/review**: Triggers a comprehensive Socratic architecture review of your codebase.
* **/design**: Asynchronously generates system specification blueprints (`ARCHITECTURE.md`, `DESIGN.md`, `SYSTEM.md`).
* **/export `[dir]`**: Exports generated blueprints and the unified AI orchestration file (`AGENTS.md`) to your workspace.
* **/diff**: Feeds staged and unstaged Git diffs to the architect model for structural change audits.
* **/mode `[greenfield\|brownfield]`**: Inspects or toggles your current project mode.
* **/copy**: Copies the last AI assistant response to your system clipboard.
* **/status**: Displays active provider, model, tokens used, and project mode.
* **/compact**: Summarizes active conversation history to reduce context footprint.
* **/model `[name]`**: Switches the active AI model.
* **/clear**: Resets the chat history while maintaining base system instructions.
* **/exit** / **/quit**: Closes the interactive session.

---

## Testing

Archon includes unit tests and end-to-end integration tests.

Run unit tests:
```bash
make test
```

Run automated E2E piping, doctor, and live model connectivity checks:
```bash
make e2e
```

---

## Contributing

Please review [`CONTRIBUTING.md`](CONTRIBUTING.md) and [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) for guidelines on submitting pull requests and reporting issues.

For development logs and debugging:
* Run unit tests with `make test`.
* Keep the codebase lean and ensure any diagnostic logs in commands output to `stderr` to preserve piping compatibility.

---

## Alternatives

* **Plandex**: Excellent command-line tool, but focused on code execution. Archon focuses purely on Socratic inquiry and high-level architectural blueprints.
* **Aider**: Great CLI pair programmer. Archon works upstream of Aider, allowing you to design the structure before writing the code.
* **Claude Code**: Powerful terminal-based coding assistant. Archon acts as a lighter, provider-agnostic, architect-only tool.

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
