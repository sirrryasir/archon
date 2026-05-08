import { readdir, readFile, writeFile, mkdir, appendFile } from 'node:fs/promises';
import { join } from 'node:path';
import { homedir } from 'node:os';
import { randomUUID } from 'node:crypto';
import type { Message } from '../ai/index.ts';

/**
 * Archon Session Manager
 * Persists conversation history as JSONL files
 * Inspired by Claude Code's history.jsonl approach
 */

const ARCHON_DIR = join(homedir(), '.archon');
const SESSIONS_DIR = join(ARCHON_DIR, 'sessions');

export interface SessionMeta {
  id: string;
  projectPath: string;
  createdAt: number;
  updatedAt: number;
  title: string;
  messageCount: number;
  model: string;
}

interface SessionLogEntry {
  role: Message['role'];
  content: string;
  timestamp: number;
}

/**
 * Ensure the sessions directory exists
 */
async function ensureSessionsDir(): Promise<void> {
  await mkdir(SESSIONS_DIR, { recursive: true });
}

/**
 * Generate a unique session ID
 */
export function createSessionId(): string {
  return randomUUID().slice(0, 8);
}

/**
 * Get the file path for a session
 */
function getSessionPath(sessionId: string): string {
  return join(SESSIONS_DIR, `${sessionId}.jsonl`);
}

function getMetaPath(sessionId: string): string {
  return join(SESSIONS_DIR, `${sessionId}.meta.json`);
}

/**
 * Save session metadata
 */
export async function saveSessionMeta(meta: SessionMeta): Promise<void> {
  await ensureSessionsDir();
  await writeFile(getMetaPath(meta.id), JSON.stringify(meta, null, 2), 'utf-8');
}

/**
 * Append a message to the session log
 */
export async function appendMessage(sessionId: string, message: Message): Promise<void> {
  await ensureSessionsDir();
  const entry: SessionLogEntry = {
    role: message.role,
    content: message.content,
    timestamp: Date.now(),
  };
  await appendFile(getSessionPath(sessionId), JSON.stringify(entry) + '\n', 'utf-8');
}

/**
 * Load all messages from a session
 */
export async function loadSession(sessionId: string): Promise<Message[]> {
  try {
    const content = await readFile(getSessionPath(sessionId), 'utf-8');
    const lines = content.trim().split('\n').filter(Boolean);
    return lines.map(line => {
      const entry: SessionLogEntry = JSON.parse(line);
      return { role: entry.role, content: entry.content };
    });
  } catch {
    return [];
  }
}

/**
 * List all available sessions, sorted by most recent
 */
export async function listSessions(): Promise<SessionMeta[]> {
  await ensureSessionsDir();
  const files = await readdir(SESSIONS_DIR);
  const metaFiles = files.filter(f => f.endsWith('.meta.json'));

  const sessions: SessionMeta[] = [];
  for (const file of metaFiles) {
    try {
      const content = await readFile(join(SESSIONS_DIR, file), 'utf-8');
      sessions.push(JSON.parse(content));
    } catch {
      // Skip corrupted meta files
    }
  }

  return sessions.sort((a, b) => b.updatedAt - a.updatedAt);
}

/**
 * Generate a title from the first user message
 */
export function generateTitle(messages: Message[]): string {
  const firstUser = messages.find(m => m.role === 'user');
  if (!firstUser) return 'New Session';
  // Truncate to first 60 chars
  const title = firstUser.content.slice(0, 60);
  return title.length < firstUser.content.length ? `${title}...` : title;
}
