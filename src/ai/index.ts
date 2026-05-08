import chalk from 'chalk';
import { gatherProjectContext } from '../mcp/index.ts';

/**
 * Archon AI Module
 * Handles communication with Ollama (and future providers)
 */

const OLLAMA_BASE = process.env.OLLAMA_URL ?? 'http://localhost:11434';
const OLLAMA_CHAT_URL = `${OLLAMA_BASE}/api/chat`;
const OLLAMA_GENERATE_URL = `${OLLAMA_BASE}/api/generate`;

function getModel(): string {
  return process.env.ARCHON_MODEL ?? 'qwen3:14b';
}

const SOCRATIC_PROMPT = `
You are Archon, an elite Socratic AI Software Architect.
CRITICAL RULE: DO NOT write complete code solutions or boilerplate for the user. 
Your goal is to ENGINEER MINDS, not just systems.

When the user asks a question:
1. First analyze the PROJECT CONTEXT provided to understand what they are building.
2. Identify the core architectural challenge or system implication based on their actual project.
3. Ask 1-2 probing questions to make the user think about scalability, latency, memory, or security.
4. Reference specific files, dependencies, or patterns you see in their project.
5. Use high-level engineering principles (Microservices, DB indexing, Event-driven architecture).
6. If useful, provide a brief ASCII diagram of how components should interact.
7. Keep responses concise, direct, and authoritative (like a Tech Lead).
`;

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Types
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export interface Message {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

interface AskOptions {
  projectPath?: string;
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Single-shot prompt (for `archon prompt` command)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export async function askArchon(query: string, options: AskOptions = {}): Promise<void> {
  let contextBlock = '';
  if (options.projectPath) {
    try {
      contextBlock = await gatherProjectContext(options.projectPath);
    } catch {
      // Proceed without context
    }
  }

  const userPrompt = contextBlock
    ? `${contextBlock}\n\n---\nUser Query: ${query}\n\nAnalyze the project context above, then provide your Socratic architectural guidance.`
    : `User Query: ${query}\n\nProvide your Socratic architectural guidance.`;

  const payload = {
    model: getModel(),
    prompt: userPrompt,
    system: SOCRATIC_PROMPT,
    stream: true,
  };

  try {
    const response = await fetch(OLLAMA_GENERATE_URL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });

    if (!response.body) throw new Error('No response body from Ollama');

    const reader = response.body.getReader();
    const decoder = new TextDecoder('utf-8');
    let done = false;
    let buffer = '';

    process.stdout.write(chalk.cyan('🏛️  Archon: '));
    while (!done) {
      const { value, done: readerDone } = await reader.read();
      done = readerDone;
      if (value) {
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';

        for (const line of lines) {
          if (!line.trim()) continue;
          try {
            const parsed = JSON.parse(line);
            if (parsed.response) {
              process.stdout.write(chalk.white(parsed.response));
            }
            if (parsed.done) {
              done = true;
            }
          } catch {
            // Incomplete JSON chunk
          }
        }
      }
    }
    process.stdout.write('\n\n');
  } catch (error: any) {
    console.error(chalk.red('\nFailed to connect to local Ollama instance.'));
    console.error(chalk.gray(`Please ensure Ollama is running and model '${getModel()}' is pulled.`));
    console.error(chalk.red(`Error: ${error.message}\n`));
  }
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Streaming chat (async generator for Ink UI)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export async function* chatStream(
  messages: Message[],
  modelOverride?: string,
): AsyncGenerator<string, void, unknown> {
  const payload = {
    model: modelOverride ?? getModel(),
    messages: messages,
    stream: true,
  };

  try {
    const response = await fetch(OLLAMA_CHAT_URL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });

    if (!response.body) throw new Error('No response body from Ollama');

    const reader = response.body.getReader();
    const decoder = new TextDecoder('utf-8');
    let done = false;
    let buffer = '';

    while (!done) {
      const { value, done: readerDone } = await reader.read();
      done = readerDone;
      if (value) {
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';

        for (const line of lines) {
          if (!line.trim()) continue;
          try {
            const parsed = JSON.parse(line);
            if (parsed.message?.content) {
              yield parsed.message.content;
            }
            if (parsed.done) {
              done = true;
            }
          } catch {
            // Incomplete JSON chunk
          }
        }
      }
    }
  } catch (error: any) {
    yield `\n[Error communicating with Archon AI: ${error.message}]\n`;
  }
}
