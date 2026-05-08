#!/usr/bin/env bun

import { Command } from 'commander';
import { setupCLI } from '../src/cli/index.ts';
import { startChatSession } from '../src/cli/chat.tsx';

const program = new Command();

program
  .name('archon')
  .description('Archon CLI: The Socratic AI Software Architect')
  .version('1.0.0');

// Ku xir dhammaan commands-ka
setupCLI(program);

// Haddii amar la siinin, si toos ah u bilow chat session (like Claude Code)
if (process.argv.slice(2).length === 0) {
  startChatSession();
} else {
  program.parse(process.argv);
}
