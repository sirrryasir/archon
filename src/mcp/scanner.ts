import { readdir, stat, readFile } from 'node:fs/promises';
import { join, extname, basename } from 'node:path';

// Faylasha aynu iska dhaafdeyno markaad akhriyso workspace-ka
const IGNORED_DIRS = new Set([
  'node_modules', '.git', 'dist', 'build', '.next',
  '.cache', 'coverage', '.turbo', '.bun',
  'opensource', '.opencode', '.gemini',
]);

const IGNORED_FILES = new Set([
  'bun.lock', 'package-lock.json', 'yarn.lock', 'pnpm-lock.yaml',
  'node_modules', '.DS_Store',
]);

const SENSITIVE_FILES = new Set([
  '.env', '.env.local', '.env.development', '.env.production',
  'id_rsa', 'id_ed25519', 'credentials.json', 'service-account.json',
]);

// Noocyada faylasha aynu ka xogta soo ururino
const CODE_EXTENSIONS = new Set([
  '.ts', '.tsx', '.js', '.jsx', '.json', '.md',
  '.yaml', '.yml', '.toml', '.sh',
]);

export interface FileEntry {
  path: string;
  type: 'file' | 'directory';
  extension?: string;
  sizeBytes?: number;
  content?: string;
}

export interface ProjectContext {
  rootPath: string;
  files: FileEntry[];
  packageJson?: Record<string, any>;
  hasTypeScript: boolean;
  hasBun: boolean;
  directories: string[];
  entryPoints: string[];
}

/**
 * Mashruuca ku jira ee directory-ga oo dhan uu akhriyo
 * si uu u fahmo qaabdhismeedka (Architecture)
 */
export async function scanWorkspace(rootPath: string, maxDepth = 4): Promise<ProjectContext> {
  const files: FileEntry[] = [];
  const directories: string[] = [];

  async function walk(dir: string, depth: number): Promise<void> {
    if (depth > maxDepth) return;

    const entries = await readdir(dir, { withFileTypes: true });

    for (const entry of entries) {
      const fullPath = join(dir, entry.name);
      const relativePath = fullPath.replace(rootPath + '/', '');

      if (entry.isDirectory()) {
        if (IGNORED_DIRS.has(entry.name)) continue;
        directories.push(relativePath);
        await walk(fullPath, depth + 1);
      } else if (entry.isFile()) {
        if (IGNORED_FILES.has(entry.name) || SENSITIVE_FILES.has(entry.name)) continue;
        const ext = extname(entry.name);
        const fileInfo = await stat(fullPath);
        
        let content: string | undefined;
        // 1MB xadka file-ka si aan token limit loo dhaafin
        if (CODE_EXTENSIONS.has(ext) && fileInfo.size < 1024 * 1024) {
          try {
            const rawContent = await readFile(fullPath, 'utf-8');
            // Redaction logic: Iska ilaali secrets-ka (basic regex)
            const secretRegex = /(?:sk-|AIza|ghp_|gho_|ghu_|ghs_|ghr_|SECRET|PASSWORD|TOKEN|KEY|PASS)[\w-]{10,}/gi;
            content = rawContent.replace(secretRegex, '[REDACTED]');
          } catch (e) {
            // iska indho tir error-ka akhriska
          }
        }

        files.push({
          path: relativePath,
          type: 'file',
          extension: ext,
          sizeBytes: fileInfo.size,
          content,
        });
      }
    }
  }

  await walk(rootPath, 0);

  // Package.json akhri
  let packageJson: Record<string, any> | undefined;
  try {
    const raw = await readFile(join(rootPath, 'package.json'), 'utf-8');
    packageJson = JSON.parse(raw);
  } catch {
    // Ma jiro package.json
  }

  // Ogaaw haddii TypeScript iyo Bun la isticmaalayo
  const hasTypeScript = files.some(
    (f) => f.path === 'tsconfig.json' || f.extension === '.ts' || f.extension === '.tsx'
  );
  const hasBun = files.some((f) => f.path === 'bun.lock') ||
    !!packageJson?.devDependencies?.['@types/bun'];

  // Entry points raadi (faylasha ugu muhiimsan)
  const entryPoints = files
    .filter((f) =>
      ['index.ts', 'index.js', 'main.ts', 'main.js', 'app.ts', 'app.js'].includes(basename(f.path))
    )
    .map((f) => f.path);

  return {
    rootPath,
    files,
    packageJson,
    hasTypeScript,
    hasBun,
    directories,
    entryPoints,
  };
}

/**
 * Fayl gaar ah akhriso, natiijona soo celi qoraalka oo dhan
 */
export async function readFileContent(filePath: string): Promise<string> {
  return readFile(filePath, 'utf-8');
}
