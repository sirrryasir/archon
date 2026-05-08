import { writeFile, mkdir } from 'node:fs/promises';
import { join } from 'node:path';
import type { Message } from '../ai/index.ts';

/**
 * Design Document Generator
 *
 * Takes conversation history and generates structured design documents
 * that humans can read and AI agents can use as implementation blueprints.
 */

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Types
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export interface DesignDocs {
  architecture: string;
  design: string;
  system: string;
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Prompts for each document type
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export const ARCHITECTURE_DOC_PROMPT = `
Based on our conversation, generate a comprehensive ARCHITECTURE.md document.

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

Write ONLY the markdown document content. No preamble or explanation.
`;

export const DESIGN_DOC_PROMPT = `
Based on our conversation, generate a comprehensive DESIGN.md document.

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

Write ONLY the markdown document content. No preamble or explanation.
`;

export const SYSTEM_DOC_PROMPT = `
Based on our conversation, generate a comprehensive SYSTEM.md document.

FORMAT REQUIREMENTS:
- Start with "# System Specification" as the title
- Include a Stack section listing: runtime, framework, database, auth, etc.
- Include a Directory Structure section with a proposed file tree using ASCII
- Include a Module Contracts section: for each module list inputs, outputs, side effects
- Include a Database Schema section (if applicable) using markdown tables
- Include an Environment Variables section listing each var, its purpose, and example value
- Include a Build & Deploy section with commands
- Include an Implementation Priority section ordering what to build first

This document should be detailed enough that an AI coding agent (Claude Code, Codex, etc.)
could read it and implement the entire system without asking clarifying questions.

Write ONLY the markdown document content. No preamble or explanation.
`;

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Conversation Extraction
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

/**
 * Extract the architectural conversation as a condensed context string.
 * Strips system messages and focuses on user + assistant exchanges.
 */
export function extractConversationContext(messages: Message[]): string {
  const relevant = messages
    .filter(m => m.role === 'user' || m.role === 'assistant')
    .map(m => `[${m.role.toUpperCase()}]: ${m.content}`)
    .join('\n\n');

  return relevant;
}

/**
 * Build the prompt for generating a specific design document.
 */
export function buildDesignPrompt(
  docType: 'architecture' | 'design' | 'system',
  conversationContext: string,
): string {
  const prompts = {
    architecture: ARCHITECTURE_DOC_PROMPT,
    design: DESIGN_DOC_PROMPT,
    system: SYSTEM_DOC_PROMPT,
  };

  return `
You are Archon, an expert software architect generating design documentation.

Here is the full architectural conversation that has taken place:

═══════════════════════════════════════════════
CONVERSATION:
═══════════════════════════════════════════════
${conversationContext}

═══════════════════════════════════════════════
TASK:
═══════════════════════════════════════════════
${prompts[docType]}
`;
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// File Export
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

/**
 * Write design documents to the specified directory.
 */
export async function exportDesignDocs(
  docs: DesignDocs,
  outputDir: string,
): Promise<string[]> {
  await mkdir(outputDir, { recursive: true });

  const files = [
    { name: 'ARCHITECTURE.md', content: docs.architecture },
    { name: 'DESIGN.md', content: docs.design },
    { name: 'SYSTEM.md', content: docs.system },
  ];

  const written: string[] = [];

  for (const file of files) {
    if (file.content && file.content.trim().length > 0) {
      const fullPath = join(outputDir, file.name);
      await writeFile(fullPath, file.content, 'utf-8');
      written.push(fullPath);
    }
  }

  return written;
}
