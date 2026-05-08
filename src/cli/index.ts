import chalk from 'chalk';
import ora from 'ora';
import type { Command } from 'commander';

import { askArchon } from '../ai/index.ts';
import { gatherProjectContext } from '../mcp/index.ts';
import { startChatSession } from './chat.tsx';

export function setupCLI(program: Command): void {
  const cwd = process.cwd();

  // 1. Amarka `init` - Si mashruuc cusub qaabdhismeedkiisa loo dejiyo
  program
    .command('init')
    .description('Initialize Archon in the current directory')
    .action(() => {
      console.log(chalk.cyan.bold('\n🏛️  Archon: Welcome, Architect.'));
      const spinner = ora('Initializing system state...').start();

      setTimeout(() => {
        spinner.succeed(chalk.green('Archon initialized successfully.'));
        console.log(chalk.gray('Ready to engineer minds and systems.\n'));
      }, 1000);
    });

  // 2. Amarka `prompt` - Mashruucaaga akhriyo kadibna AI-ga weydiiso
  program
    .command('prompt <query...>')
    .description('Ask Archon for architectural guidance')
    .option('-p, --path <path>', 'Project path to analyze', cwd)
    .action(async (queryArray: string[], opts: { path: string }) => {
      const query = queryArray.join(' ');
      console.log(chalk.cyan.bold('\n🏛️  Archon is analyzing your query:'));
      console.log(chalk.gray(`"${query}"\n`));

      const spinner = ora('Scanning workspace & consulting Socratic Middleware...').start();
      spinner.stop();

      await askArchon(query, { projectPath: opts.path });
    });

  // 3. Amarka `review` - MCP Context Manager-ka isticmaalayo si uu u akhriyo mashruuca
  program
    .command('review')
    .description('Review the current project architecture')
    .option('-p, --path <path>', 'Project path to review', cwd)
    .action(async (opts: { path: string }) => {
      const spinner = ora('Scanning local workspace using MCP Context Manager...').start();

      try {
        const context = await gatherProjectContext(opts.path);
        spinner.succeed(chalk.green('Workspace scanned successfully.\n'));

        console.log(chalk.blue.bold('🔍 Architecture Review:\n'));
        console.log(chalk.white(context));
        console.log('');

        // AI-ga sii context-ka oo weydiiso talo
        const reviewSpinner = ora('Archon is reviewing your architecture...').start();
        reviewSpinner.stop();

        await askArchon(
          'Review this project architecture. Identify strengths, weaknesses, missing patterns, and suggest improvements.',
          { projectPath: opts.path }
        );
      } catch (error: any) {
        spinner.fail(chalk.red('Failed to scan workspace.'));
        console.error(chalk.red(error.message));
      }
    });

  // 4. Amarka `chat` - Interactive REPL session (Claude Code UX + Socratic Brain)
  program
    .command('chat')
    .description('Start an interactive Socratic chat session with Archon')
    .action(async () => {
      await startChatSession();
    });
}
