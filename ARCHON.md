# Archon CLI: The Socratic AI Software Architect

## 1. Project Overview
Archon CLI is a next-generation terminal-based AI assistant designed not to write code, but to engineer minds. In an era where AI can generate boilerplate code instantly, the true value of a developer lies in Systems Architecture, High-level Problem Solving, and Agent Orchestration. Archon serves as a Socratic Mentor, guiding developers to think like "System Bosses" rather than "Manual Workers."

## 2. The Core Problem & Vision
**The Problem:** Traditional AI coding tools (Copilot, Cursor) bypass the learning process by providing direct code completions. This degrades the user's architectural thinking and system design capabilities. 
**The Vision:** To build an AI that evaluates the system architecture, intercepts user queries, and responds with Socratic questioning, engineering principles, and structural analysis. It shifts the focus from syntax to scalability.

## 3. Key Features
* **Socratic Interception:** Instead of answering "write a function," Archon asks "how does this function impact the overall system's memory or latency?"
* **Local-First & Private:** Integrates directly with local LLMs (Ollama) to ensure code privacy and offline capability.
* **Context-Aware Architecture:** Utilizes the Model Context Protocol (MCP) to read and understand the entire project directory, not just isolated files.
* **Engineering Focus:** Prioritizes mentoring on macro-level concepts such as microservices, distributed systems (Kubernetes, Kafka, Redis), and system performance.
* **Terminal Native:** Designed entirely for the CLI to maintain workflow speed and efficiency for power users.

## 4. Technical Stack
* **Core Engine:** Node.js (with `commander` for CLI routing, `chalk` for styling, `ora` for loading states).
* **Context Layer:** Model Context Protocol (MCP) SDK for secure file system access and context gathering.
* **Inference Engine:** Ollama API (Local AI models for inference).
* **Format:** Pure Command Line Interface (CLI).

## 5. System Modules (Architecture Diagram)
The system is divided into three core micro-modules:

1.  **CLI Controller (`/cli`):** * Handles user input commands (e.g., `archon prompt`, `archon review`).
    * Manages terminal output, formatting, and UI/UX (spinners, colors, ascii diagrams).
2.  **MCP Context Manager (`/mcp`):** * Acts as the bridge to the user's local filesystem.
    * Scans project structures, reads configuration files, and understands the relationships between different microservices or components.
3.  **Socratic Middleware (`/ai`):** * The "Brain" of the operation.
    * Wraps user queries in strict Socratic System Prompts before sending them to Ollama.
    * Ensures the output is educational, architectural, and devoid of direct copy-paste code solutions.

## 6. Execution Flow
1.  **User Input:** The developer types a query into the terminal.
2.  **Context Gathering:** The CLI Controller triggers the MCP Manager to read relevant project files and architecture state.
3.  **Prompt Engineering:** The user's query + project context are combined with the core Socratic prompt.
4.  **Inference:** The payload is sent to the local Ollama instance.
5.  **Output:** Archon streams the response back to the terminal, providing architectural insights, flowcharts (text-based), and guiding questions.

## 7. Core Directives for Code Generation (For AI Agents assisting in building this)
When writing code for Archon CLI, AI agents must adhere to the following rules:
* **Modularity:** Keep the CLI logic, Context (MCP) logic, and AI (Ollama) logic strictly separated in different directories.
* **Performance:** Terminal responses must stream smoothly.
* **No GUI:** Do not suggest or implement graphical interfaces; this is strictly a terminal utility.