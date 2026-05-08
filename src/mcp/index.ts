import { scanWorkspace, type ProjectContext } from './scanner.ts';

/**
 * Xogta mashruuca oo dhan soo ururiyo
 * kadibna u rogo qoraal (string) aan AI-ga siino.
 * Tani waa "Context-ka" ee AI-gu u baahan yahay
 * si uu Architecture taladajile sax ah ka bixiyo.
 */
export async function gatherProjectContext(rootPath: string): Promise<string> {
  const ctx = await scanWorkspace(rootPath);
  return formatContext(ctx);
}

/**
 * ProjectContext-ka u beddelo string-ka la siiyo AI-ga
 */
function formatContext(ctx: ProjectContext): string {
  const lines: string[] = [];

  lines.push('=== PROJECT CONTEXT (scanned by Archon MCP) ===');
  lines.push('');

  // Tech stack
  lines.push('## Tech Stack Detection:');
  lines.push(`- Runtime: ${ctx.hasBun ? 'Bun' : 'Node.js'}`);
  lines.push(`- TypeScript: ${ctx.hasTypeScript ? 'Yes' : 'No'}`);
  lines.push('');

  // Package.json dependencies
  if (ctx.packageJson) {
    const deps = ctx.packageJson.dependencies ?? {};
    const depNames = Object.keys(deps);

    if (depNames.length > 0) {
      lines.push('## Dependencies:');
      for (const name of depNames) {
        lines.push(`- ${name}: ${deps[name]}`);
      }
      lines.push('');
    }

    // Scripts
    const scripts = ctx.packageJson.scripts ?? {};
    const scriptEntries = Object.entries(scripts);
    if (scriptEntries.length > 0) {
      lines.push('## NPM/Bun Scripts:');
      for (const [key, val] of scriptEntries) {
        lines.push(`- ${key}: ${val}`);
      }
      lines.push('');
    }
  }

  // Directory structure (qaabdhismeedka galka)
  lines.push('## Directory Structure:');
  for (const dir of ctx.directories) {
    const depth = dir.split('/').length - 1;
    const indent = '  '.repeat(depth);
    const dirName = dir.split('/').pop();
    lines.push(`${indent}${dirName}/`);
  }
  lines.push('');

  // Faylasha muhiimka ah (code files only)
  const codeFiles = ctx.files.filter((f) =>
    ['.ts', '.tsx', '.js', '.jsx', '.json', '.md'].includes(f.extension ?? '')
  );
  
  // Source Code (Nuxurka faylasha) - Xadidan 100 line halkii file
  const filesWithContent = codeFiles.filter(f => f.content);
  if (filesWithContent.length > 0) {
    lines.push('## Source Code (Key Files Content - First 100 lines):');
    for (const file of filesWithContent) {
      lines.push(`### File: ${file.path}`);
      lines.push('```' + (file.extension?.slice(1) || 'text'));
      const truncatedContent = file.content!.split('\n').slice(0, 100).join('\n');
      lines.push(truncatedContent);
      if (file.content!.split('\n').length > 100) {
        lines.push('// ... [Content truncated after 100 lines]');
      }
      lines.push('```');
      lines.push('');
    }
  }

  lines.push('=== END PROJECT CONTEXT ===');
  return lines.join('\n');
}

export type { ProjectContext };
