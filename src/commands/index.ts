import chalk from 'chalk';
import type { SlashCommand, ChatContext } from './types.ts';
import { gatherProjectContext } from '../mcp/index.ts';
import { chatStream } from '../ai/index.ts';
import {
  extractConversationContext,
  buildDesignPrompt,
  exportDesignDocs,
  type DesignDocs,
} from '../design/generator.ts';

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Built-in Slash Commands
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const helpCommand: SlashCommand = {
  name: 'help',
  aliases: ['h', '?'],
  description: 'Show all available commands',
  execute: async (_args, _ctx) => {
    const commands = getCommands();
    const lines = [
      '',
      chalk.cyan.bold('  ARCHON COMMANDS'),
      chalk.gray('  ─────────────────────────────────────'),
    ];

    for (const cmd of commands) {
      const aliases = cmd.aliases?.length
        ? chalk.gray(` (${cmd.aliases.map(a => `/${a}`).join(', ')})`)
        : '';
      lines.push(
        `  ${chalk.cyan(`/${cmd.name}`)}${aliases}  ${chalk.white(cmd.description)}`
      );
    }

    lines.push(chalk.gray('  ─────────────────────────────────────'));
    lines.push('');
    lines.push(chalk.cyan.bold('  SHORTCUTS'));
    lines.push(chalk.gray('  ─────────────────────────────────────'));
    lines.push(`  ${chalk.cyan('Ctrl+L')}      ${chalk.white('Clear and redraw screen')}`);
    lines.push(`  ${chalk.cyan('Ctrl+C')}      ${chalk.white('Exit session')}`);
    lines.push(`  ${chalk.cyan('Alt+M')}       ${chalk.white('Toggle Greenfield/Brownfield mode')}`);
    lines.push(`  ${chalk.cyan('Esc')}         ${chalk.white('Clear input / Cancel thinking')}`);
    lines.push(chalk.gray('  ─────────────────────────────────────'));
    lines.push('');
    return lines.join('\n');
  },
};

const clearCommand: SlashCommand = {
  name: 'clear',
  aliases: ['c'],
  description: 'Clear conversation history',
  execute: async (_args, ctx) => {
    // Keep only system messages
    ctx.setMessages(ctx.messages.filter(m => m.role === 'system'));
    return chalk.green('  ✓ Conversation cleared. System context preserved.');
  },
};

const compactCommand: SlashCommand = {
  name: 'compact',
  description: 'Summarize conversation to reduce context size',
  execute: async (_args, ctx) => {
    const systemMsgs = ctx.messages.filter(m => m.role === 'system');
    const chatMsgs = ctx.messages.filter(m => m.role !== 'system');

    if (chatMsgs.length < 4) {
      return chalk.yellow('  ⚠ Not enough conversation to compact.');
    }

    // Ask AI to summarize the conversation
    const summaryMessages = [
      ...systemMsgs,
      ...chatMsgs,
      {
        role: 'user' as const,
        content: 'Summarize the key architectural points, decisions, and insights from our conversation so far in a concise format. This summary will replace the full conversation history to save context space.',
      },
    ];

    let summary = '';
    for await (const chunk of chatStream(summaryMessages)) {
      summary += chunk;
    }

    // Replace conversation with compact summary
    const compactedMessages = [
      ...systemMsgs,
      {
        role: 'assistant' as const,
        content: `[Conversation compacted — ${chatMsgs.length} messages summarized]\n\n${summary}`,
      },
    ];

    ctx.setMessages(compactedMessages);
    return chalk.green(`  ✓ Compacted ${chatMsgs.length} messages into a summary.`);
  },
};

const reviewCommand: SlashCommand = {
  name: 'review',
  aliases: ['r'],
  description: 'Review current project architecture',
  execute: async (_args, ctx) => {
    const context = await gatherProjectContext(ctx.projectPath);

    // Inject as a user message so AI responds with analysis
    const reviewPrompt = `Review this project architecture. Identify strengths, weaknesses, missing patterns, and suggest improvements.\n\n${context}`;

    ctx.setMessages([
      ...ctx.messages,
      { role: 'user' as const, content: reviewPrompt },
    ]);

    // Signal that the AI should respond to this
    return undefined; // undefined = let AI respond
  },
};

const filesCommand: SlashCommand = {
  name: 'files',
  aliases: ['f', 'ls'],
  description: 'List scanned project files',
  execute: async (_args, ctx) => {
    const { scanWorkspace } = await import('../mcp/scanner.ts');
    const projectCtx = await scanWorkspace(ctx.projectPath);

    const lines = [
      '',
      chalk.cyan.bold(`  PROJECT: ${ctx.projectPath}`),
      chalk.gray(`  ─────────────────────────────────────`),
      chalk.white(`  Directories: ${projectCtx.directories.length}`),
      chalk.white(`  Files: ${projectCtx.files.length}`),
      chalk.white(`  TypeScript: ${projectCtx.hasTypeScript ? 'Yes' : 'No'}`),
      chalk.white(`  Runtime: ${projectCtx.hasBun ? 'Bun' : 'Node.js'}`),
    ];

    if (projectCtx.entryPoints.length > 0) {
      lines.push('');
      lines.push(chalk.cyan('  Entry Points:'));
      for (const ep of projectCtx.entryPoints) {
        lines.push(chalk.gray(`    → ${ep}`));
      }
    }

    lines.push('');
    for (const dir of projectCtx.directories.slice(0, 20)) {
      lines.push(chalk.gray(`    ${dir}/`));
    }
    if (projectCtx.directories.length > 20) {
      lines.push(chalk.gray(`    ... and ${projectCtx.directories.length - 20} more`));
    }

    lines.push(chalk.gray('  ─────────────────────────────────────'));
    lines.push('');
    return lines.join('\n');
  },
};

const modelCommand: SlashCommand = {
  name: 'model',
  aliases: ['m'],
  description: 'Show or switch AI model',
  usage: '/model [model-name]',
  execute: async (args, ctx) => {
    if (!args) {
      return chalk.cyan(`  Current model: ${chalk.bold(ctx.model)}`);
    }
    ctx.setModel(args);
    return chalk.green(`  ✓ Model switched to: ${chalk.bold(args)}`);
  },
};

const modeCommand: SlashCommand = {
  name: 'mode',
  description: 'Show or switch mode (greenfield/brownfield)',
  usage: '/mode [greenfield|brownfield]',
  execute: async (args, ctx) => {
    if (!args) {
      return chalk.cyan(`  Current mode: ${chalk.bold(ctx.mode)}`);
    }
    const newMode = args.toLowerCase().trim();
    if (newMode !== 'greenfield' && newMode !== 'brownfield') {
      return chalk.yellow(`  ⚠ Invalid mode. Use: /mode greenfield  or  /mode brownfield`);
    }
    ctx.setMode(newMode);
    return chalk.green(`  ✓ Switched to ${chalk.bold(newMode)} mode.`);
  },
};

const designCommand: SlashCommand = {
  name: 'design',
  aliases: ['d'],
  description: 'Generate design documents from conversation',
  execute: async (_args, ctx) => {
    const chatMsgs = ctx.messages.filter(m => m.role !== 'system');

    if (chatMsgs.length < 2) {
      return chalk.yellow('  ⚠ Have a conversation first! I need context to generate design docs.');
    }

    const conversationContext = extractConversationContext(ctx.messages);
    const docs: DesignDocs = { architecture: '', design: '', system: '' };
    const docTypes = ['architecture', 'design', 'system'] as const;

    const lines: string[] = [
      '',
      chalk.cyan.bold('  Generating Design Documents...'),
      chalk.gray('  ─────────────────────────────────────'),
    ];

    for (const docType of docTypes) {
      const label = docType.toUpperCase();
      lines.push(chalk.gray(`  → Generating ${label}.md...`));

      const prompt = buildDesignPrompt(docType, conversationContext);
      const genMessages = [
        { role: 'user' as const, content: prompt },
      ];

      let content = '';
      for await (const chunk of chatStream(genMessages)) {
        content += chunk;
      }

      docs[docType] = content;
      lines.push(chalk.green(`  ✓ ${label}.md generated (${content.length} chars)`));
    }

    ctx.setDesignDocs(docs);
    lines.push('');
    lines.push(chalk.cyan('  Design docs ready! Use /export to write them to disk.'));
    lines.push(chalk.gray('  ─────────────────────────────────────'));
    lines.push('');
    return lines.join('\n');
  },
};

const exportCommand: SlashCommand = {
  name: 'export',
  aliases: ['e'],
  description: 'Export design documents to disk',
  usage: '/export [output-dir]',
  execute: async (args, ctx) => {
    if (!ctx.designDocs) {
      return chalk.yellow('  ⚠ No design documents generated yet. Run /design first.');
    }

    const outputDir = args || ctx.projectPath;
    try {
      const writtenFiles = await exportDesignDocs(ctx.designDocs, outputDir);

      const lines = [
        '',
        chalk.green.bold('  ✓ Design documents exported!'),
        chalk.gray('  ─────────────────────────────────────'),
      ];

      for (const filePath of writtenFiles) {
        lines.push(chalk.white(`  → ${filePath}`));
      }

      lines.push(chalk.gray('  ─────────────────────────────────────'));
      lines.push('');
      return lines.join('\n');
    } catch (err: any) {
      return chalk.red(`  ✗ Export failed: ${err.message}`);
    }
  },
};

const exitCommand: SlashCommand = {
  name: 'exit',
  aliases: ['quit', 'q'],
  description: 'Exit the chat session',
  execute: async (_args, ctx) => {
    ctx.exit();
  },
};

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Command Registry
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const COMMANDS: SlashCommand[] = [
  helpCommand,
  clearCommand,
  compactCommand,
  reviewCommand,
  filesCommand,
  modelCommand,
  modeCommand,
  designCommand,
  exportCommand,
  exitCommand,
];

export function getCommands(): SlashCommand[] {
  return COMMANDS;
}

export function findCommand(name: string): SlashCommand | undefined {
  return COMMANDS.find(
    cmd => cmd.name === name || cmd.aliases?.includes(name)
  );
}

export { parseSlashCommand } from './types.ts';

