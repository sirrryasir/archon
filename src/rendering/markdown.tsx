import React from 'react';
import { Box, Text } from 'ink';
import chalk from 'chalk';
import { highlight } from 'cli-highlight';

/**
 * Archon Terminal Markdown Renderer
 *
 * Inspired by Gemini CLI's MarkdownDisplay.tsx approach:
 * Line-by-line regex parsing → Ink Box/Text components.
 *
 * Why not marked + marked-terminal?
 * - Deep nesting breaks in Ink layouts
 * - Bullet points get excessive indentation
 * - Code blocks lose formatting
 * - Tables are unreadable
 *
 * This renders markdown as clean, readable terminal output using
 * Ink's native Box/Text components for proper layout.
 */

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Colors
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const COLORS = {
  h1: '#22d3ee',        // cyan-400
  h2: '#38bdf8',        // sky-400
  h3: '#818cf8',        // indigo-400
  h4: '#a78bfa',        // violet-400
  text: '#f1f5f9',      // slate-100
  dim: '#64748b',       // slate-500
  code: '#fbbf24',      // amber-400
  codeBg: '#1e293b',    // slate-800
  bullet: '#22d3ee',    // cyan-400
  number: '#38bdf8',    // sky-400
  bold: '#ffffff',      // white
  link: '#38bdf8',      // sky-400
  hr: '#334155',        // slate-700
  blockquote: '#94a3b8', // slate-400
  tableHeader: '#22d3ee',
  tableBorder: '#334155',
  diagram: '#10b981',   // emerald-500
};

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Inline Formatting
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

/**
 * Apply inline markdown formatting: **bold**, *italic*, `code`, [links](url)
 */
function formatInline(text: string): string {
  return text
    // Bold + Italic ***text***
    .replace(/\*\*\*(.+?)\*\*\*/g, (_m, t) => chalk.bold.italic(t))
    // Bold **text**
    .replace(/\*\*(.+?)\*\*/g, (_m, t) => chalk.bold.hex(COLORS.bold)(t))
    // Italic *text* (not at word boundaries to avoid matching * in paths)
    .replace(/(?<!\w)\*([^*\n]+?)\*(?!\w)/g, (_m, t) => chalk.italic(t))
    // Inline code `text`
    .replace(/`([^`\n]+?)`/g, (_m, t) => chalk.hex(COLORS.code)(t))
    // Links [text](url)
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_m, t, url) =>
      `${chalk.hex(COLORS.link)(t)} ${chalk.dim(`(${url})`)}`
    )
    // Strikethrough ~~text~~
    .replace(/~~(.+?)~~/g, (_m, t) => chalk.strikethrough(t));
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Block-Level Parsing
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

interface Block {
  type: 'heading' | 'paragraph' | 'code' | 'list-item' | 'hr' | 'blockquote' | 'table' | 'empty';
  content: string;
  level?: number;       // heading level or list indent
  lang?: string;        // code block language
  ordered?: boolean;    // list type
  marker?: string;      // list marker (-, *, 1., 2., etc.)
  rows?: string[][];    // table rows
  headers?: string[];   // table headers
}

/**
 * Parse raw markdown text into structured blocks.
 * Gemini CLI style: line-by-line with state tracking for code blocks and tables.
 */
export function parseMarkdown(text: string): Block[] {
  const lines = text.split(/\r?\n/);
  const blocks: Block[] = [];

  let inCodeBlock = false;
  let codeLines: string[] = [];
  let codeLang = '';
  let codeFence = '';

  let inTable = false;
  let tableHeaders: string[] = [];
  let tableRows: string[][] = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];

    // ── Code Block State ──
    if (inCodeBlock) {
      const fenceMatch = line.match(/^(\s*)(```|~~~)\s*$/);
      if (fenceMatch && line.trim().startsWith(codeFence)) {
        blocks.push({
          type: 'code',
          content: codeLines.join('\n'),
          lang: codeLang || undefined,
        });
        inCodeBlock = false;
        codeLines = [];
        codeLang = '';
        codeFence = '';
      } else {
        codeLines.push(line);
      }
      continue;
    }

    // ── Code Fence Start ──
    const codeFenceMatch = line.match(/^\s*(```|~~~)(\w*)\s*$/);
    if (codeFenceMatch) {
      // Flush any pending table
      if (inTable) {
        blocks.push({ type: 'table', content: '', headers: tableHeaders, rows: tableRows });
        inTable = false;
        tableHeaders = [];
        tableRows = [];
      }
      inCodeBlock = true;
      codeFence = codeFenceMatch[1];
      codeLang = codeFenceMatch[2] || '';
      codeLines = [];
      continue;
    }

    // ── Table Detection ──
    const tableRowMatch = line.match(/^\s*\|(.+)\|\s*$/);
    const tableSepMatch = line.match(/^\s*\|?\s*(:?-+:?)\s*(\|\s*(:?-+:?)\s*)+\|?\s*$/);

    if (inTable) {
      if (tableSepMatch) continue; // skip separator rows
      if (tableRowMatch) {
        tableRows.push(tableRowMatch[1].split('|').map(c => c.trim()));
        continue;
      }
      // End of table
      blocks.push({ type: 'table', content: '', headers: tableHeaders, rows: tableRows });
      inTable = false;
      tableHeaders = [];
      tableRows = [];
      // fall through to process current line
    }

    if (!inTable && tableRowMatch) {
      // Check if next line is separator
      if (i + 1 < lines.length && lines[i + 1].match(/^\s*\|?\s*(:?-+:?)\s*(\|\s*(:?-+:?)\s*)+\|?\s*$/)) {
        inTable = true;
        tableHeaders = tableRowMatch[1].split('|').map(c => c.trim());
        continue;
      }
    }

    // ── Heading ──
    const headingMatch = line.match(/^(#{1,4})\s+(.+)/);
    if (headingMatch) {
      blocks.push({
        type: 'heading',
        content: headingMatch[2],
        level: headingMatch[1].length,
      });
      continue;
    }

    // ── Horizontal Rule ──
    if (line.match(/^\s*([-*_]\s*){3,}\s*$/)) {
      blocks.push({ type: 'hr', content: '' });
      continue;
    }

    // ── Blockquote ──
    if (line.match(/^\s*>\s?/)) {
      blocks.push({
        type: 'blockquote',
        content: line.replace(/^\s*>\s?/, ''),
      });
      continue;
    }

    // ── Unordered List ──
    const ulMatch = line.match(/^(\s*)([-*+])\s+(.*)/);
    if (ulMatch) {
      const indent = Math.floor(ulMatch[1].length / 2);
      blocks.push({
        type: 'list-item',
        content: ulMatch[3],
        level: indent,
        ordered: false,
        marker: ulMatch[2],
      });
      continue;
    }

    // ── Ordered List ──
    const olMatch = line.match(/^(\s*)(\d+)\.\s+(.*)/);
    if (olMatch) {
      const indent = Math.floor(olMatch[1].length / 2);
      blocks.push({
        type: 'list-item',
        content: olMatch[3],
        level: indent,
        ordered: true,
        marker: olMatch[2],
      });
      continue;
    }

    // ── Empty Line ──
    if (line.trim() === '') {
      // Only add if last block wasn't empty
      if (blocks.length === 0 || blocks[blocks.length - 1].type !== 'empty') {
        blocks.push({ type: 'empty', content: '' });
      }
      continue;
    }

    // ── Paragraph Text ──
    blocks.push({ type: 'paragraph', content: line });
  }

  // Flush pending code block (streaming)
  if (inCodeBlock && codeLines.length > 0) {
    blocks.push({ type: 'code', content: codeLines.join('\n'), lang: codeLang || undefined });
  }

  // Flush pending table
  if (inTable && tableHeaders.length > 0) {
    blocks.push({ type: 'table', content: '', headers: tableHeaders, rows: tableRows });
  }

  return blocks;
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Ink Components
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Heading({ text, level }: { text: string; level: number }) {
  const colors: Record<number, string> = {
    1: COLORS.h1,
    2: COLORS.h2,
    3: COLORS.h3,
    4: COLORS.h4,
  };
  
  const prefixes: Record<number, string> = {
    1: '',
    2: '',
    3: '',
    4: '',
  };

  const c = colors[level] || COLORS.text;
  const prefix = prefixes[level] || '';
  const formatted = formatInline(text);

  return (
    <Box marginBottom={level <= 2 ? 1 : 0} marginTop={level === 1 ? 1 : 0}>
      <Text bold color={c}>{prefix}{formatted}</Text>
    </Box>
  );
}

function CodeBlock({ code, lang }: { code: string; lang?: string }) {
  const isDiagram = lang === 'ascii' || lang === 'diagram' || lang === 'architecture';
  const borderColor = isDiagram ? COLORS.diagram : COLORS.dim;
  const label = isDiagram ? ' ARCHITECTURE DIAGRAM ' : (lang ? ` ${lang.toUpperCase()} ` : ' CODE ');
  
  let highlighted: string;
  if (isDiagram) {
    highlighted = code; // Don't highlight diagrams, keep raw ASCII
  } else {
    try {
      highlighted = highlight(code, {
        language: lang || 'plaintext',
        ignoreIllegals: true,
      });
    } catch {
      highlighted = code;
    }
  }

  const width = 80;

  return (
    <Box flexDirection="column" marginBottom={1} marginTop={1}>
      <Box>
        <Text color={borderColor}>{'┏' + '━'.repeat(2)}</Text>
        <Text color={borderColor} bold italic>{label}</Text>
        <Text color={borderColor}>{'━'.repeat(Math.max(2, width - label.length - 3))}</Text>
      </Box>
      <Box paddingLeft={0} flexDirection="column">
        {highlighted.split('\n').map((line, i) => (
          <Box key={i} flexDirection="row">
            <Text color={borderColor}>{'┃ '}</Text>
            <Text>{line}</Text>
          </Box>
        ))}
      </Box>
      <Box>
        <Text color={borderColor}>{'┗' + '━'.repeat(width - 1)}</Text>
      </Box>
    </Box>
  );
}

function ListItem({ text, indent, ordered, marker }: {
  text: string; indent: number; ordered: boolean; marker: string;
}) {
  const padding = indent * 2;
  const bullet = ordered
    ? chalk.hex(COLORS.number)(`${marker}.`)
    : chalk.hex(COLORS.bullet)('•');
  const formatted = formatInline(text);

  return (
    <Box paddingLeft={padding} flexDirection="row">
      <Box width={ordered ? 3 : 2} flexShrink={0}>
        <Text>{bullet}</Text>
      </Box>
      <Box flexGrow={1}>
        <Text wrap="wrap">{formatted}</Text>
      </Box>
    </Box>
  );
}

function Blockquote({ text }: { text: string }) {
  const formatted = formatInline(text);
  return (
    <Box paddingLeft={1}>
      <Text>{chalk.hex(COLORS.blockquote)('▌')} {chalk.italic.hex(COLORS.blockquote)(formatted)}</Text>
    </Box>
  );
}

function HorizontalRule() {
  return (
    <Box marginY={0}>
      <Text color={COLORS.hr}>{'─'.repeat(60)}</Text>
    </Box>
  );
}

function Table({ headers, rows }: { headers: string[]; rows: string[][] }) {
  // Calculate column widths
  const colWidths = headers.map((h, i) => {
    let max = h.length;
    for (const row of rows) {
      const cell = row[i] || '';
      max = Math.max(max, cell.length);
    }
    return Math.min(max + 2, 40); // cap at 40 chars
  });

  const hLine = colWidths.map(w => '─'.repeat(w)).join('┬');
  const sLine = colWidths.map(w => '─'.repeat(w)).join('┼');
  const bLine = colWidths.map(w => '─'.repeat(w)).join('┴');

  const padCell = (text: string, width: number) => {
    const pad = Math.max(0, width - text.length);
    return ' ' + text + ' '.repeat(pad - 1);
  };

  const headerRow = headers.map((h, i) => padCell(h, colWidths[i])).join('│');

  return (
    <Box flexDirection="column" marginBottom={1}>
      <Text color={COLORS.tableBorder}>{'┌' + hLine + '┐'}</Text>
      <Text>
        {chalk.hex(COLORS.tableBorder)('│')}
        {chalk.bold.hex(COLORS.tableHeader)(headerRow)}
        {chalk.hex(COLORS.tableBorder)('│')}
      </Text>
      <Text color={COLORS.tableBorder}>{'├' + sLine + '┤'}</Text>
      {rows.map((row, ri) => {
        const cells = headers.map((_, ci) => padCell(row[ci] || '', colWidths[ci])).join('│');
        return (
          <Text key={ri}>
            {chalk.hex(COLORS.tableBorder)('│')}
            {cells}
            {chalk.hex(COLORS.tableBorder)('│')}
          </Text>
        );
      })}
      <Text color={COLORS.tableBorder}>{'└' + bLine + '┘'}</Text>
    </Box>
  );
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Main Renderer
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export function MarkdownRenderer({ text }: { text: string }) {
  const blocks = parseMarkdown(text);

  return (
    <Box flexDirection="column">
      {blocks.map((block, idx) => {
        switch (block.type) {
          case 'heading':
            return <Heading key={idx} text={block.content} level={block.level || 1} />;
          case 'code':
            return <CodeBlock key={idx} code={block.content} lang={block.lang} />;
          case 'list-item':
            return (
              <ListItem
                key={idx}
                text={block.content}
                indent={block.level || 0}
                ordered={block.ordered || false}
                marker={block.marker || '-'}
              />
            );
          case 'blockquote':
            return <Blockquote key={idx} text={block.content} />;
          case 'hr':
            return <HorizontalRule key={idx} />;
          case 'table':
            return (
              <Table
                key={idx}
                headers={block.headers || []}
                rows={block.rows || []}
              />
            );
          case 'empty':
            return <Box key={idx} height={1} />;
          case 'paragraph':
          default:
            return (
              <Box key={idx}>
                <Text wrap="wrap">{formatInline(block.content)}</Text>
              </Box>
            );
        }
      })}
    </Box>
  );
}
