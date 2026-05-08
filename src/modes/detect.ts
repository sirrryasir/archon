import { readdir, stat, access } from 'node:fs/promises';
import { join } from 'node:path';

/**
 * Archon Mode Detection
 * Determines whether Archon should run in Greenfield or Brownfield mode
 * based on the presence of project markers in the working directory.
 */

export type ProjectMode = 'greenfield' | 'brownfield';

/** Files whose mere existence signals "this is a real project" */
const PROJECT_MARKERS = [
  'package.json',
  'Cargo.toml',
  'pyproject.toml',
  'go.mod',
  'pom.xml',
  'build.gradle',
  'Makefile',
  'CMakeLists.txt',
  'composer.json',
  'Gemfile',
  'requirements.txt',
  'setup.py',
  'tsconfig.json',
  'deno.json',
  '.sln',
];

const SOURCE_EXTENSIONS = new Set([
  '.ts', '.tsx', '.js', '.jsx', '.py', '.rs', '.go',
  '.java', '.rb', '.php', '.c', '.cpp', '.cs', '.swift',
]);

/**
 * Detect whether the given directory is greenfield (empty/new) or brownfield (existing project).
 */
export async function detectMode(rootPath: string): Promise<{
  mode: ProjectMode;
  reason: string;
  fileCount: number;
}> {
  // Check for project marker files
  for (const marker of PROJECT_MARKERS) {
    try {
      await access(join(rootPath, marker));
      return {
        mode: 'brownfield',
        reason: `Found ${marker}`,
        fileCount: -1, // will be counted later by scanner
      };
    } catch {
      // File doesn't exist, keep checking
    }
  }

  // Count source files in root and immediate subdirectories
  let sourceFileCount = 0;
  try {
    const entries = await readdir(rootPath, { withFileTypes: true });

    for (const entry of entries) {
      if (entry.isFile()) {
        const ext = entry.name.slice(entry.name.lastIndexOf('.'));
        if (SOURCE_EXTENSIONS.has(ext)) {
          sourceFileCount++;
        }
      } else if (entry.isDirectory() && !entry.name.startsWith('.')) {
        // Check one level deep
        try {
          const subEntries = await readdir(join(rootPath, entry.name), { withFileTypes: true });
          for (const sub of subEntries) {
            if (sub.isFile()) {
              const ext = sub.name.slice(sub.name.lastIndexOf('.'));
              if (SOURCE_EXTENSIONS.has(ext)) {
                sourceFileCount++;
              }
            }
          }
        } catch {
          // Can't read subdirectory
        }
      }
    }
  } catch {
    // Can't read directory at all — treat as greenfield
    return { mode: 'greenfield', reason: 'Cannot read directory', fileCount: 0 };
  }

  if (sourceFileCount >= 3) {
    return {
      mode: 'brownfield',
      reason: `Found ${sourceFileCount} source files`,
      fileCount: sourceFileCount,
    };
  }

  return {
    mode: 'greenfield',
    reason: sourceFileCount === 0
      ? 'Empty directory — no project files detected'
      : `Only ${sourceFileCount} source file(s) — treating as new project`,
    fileCount: sourceFileCount,
  };
}
