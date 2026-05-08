import type { Message } from '../ai/index.ts';
import type { ProjectMode } from '../modes/detect.ts';
import type { DesignDocs } from '../design/generator.ts';

/**
 * Archon Slash Command System
 * Inspired by Claude Code's command architecture
 */

export interface ChatContext {
  /** Current working directory */
  cwd: string;
  /** Current conversation messages */
  messages: Message[];
  /** Set messages (for /clear, /compact) */
  setMessages: (msgs: Message[]) => void;
  /** Current model name */
  model: string;
  /** Set model */
  setModel: (model: string) => void;
  /** Exit the chat session */
  exit: () => void;
  /** Project path being analyzed */
  projectPath: string;
  /** Current mode: greenfield or brownfield */
  mode: ProjectMode;
  /** Set mode */
  setMode: (mode: ProjectMode) => void;
  /** Generated design documents (in-memory) */
  designDocs: DesignDocs | null;
  /** Set design documents */
  setDesignDocs: (docs: DesignDocs) => void;
}

export interface SlashCommand {
  /** Command name (without /) */
  name: string;
  /** Alternative names */
  aliases?: string[];
  /** Short description for /help */
  description: string;
  /** Detailed usage hint */
  usage?: string;
  /** Execute the command */
  execute: (args: string, context: ChatContext) => Promise<string | void>;
}

/**
 * Parse user input for slash commands
 * Returns the command name and arguments, or null if not a command
 */
export function parseSlashCommand(input: string): { name: string; args: string } | null {
  const trimmed = input.trim();
  if (!trimmed.startsWith('/')) return null;

  const spaceIndex = trimmed.indexOf(' ');
  if (spaceIndex === -1) {
    return { name: trimmed.slice(1).toLowerCase(), args: '' };
  }

  return {
    name: trimmed.slice(1, spaceIndex).toLowerCase(),
    args: trimmed.slice(spaceIndex + 1).trim(),
  };
}
