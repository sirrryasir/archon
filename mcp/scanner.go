package mcp

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const maxGlobalBytes = 2 * 1024 * 1024 // 2MB global limit

var ignoredDirs = map[string]bool{
	"node_modules": true, ".git": true, "dist": true, "build": true, ".next": true,
	".cache": true, "coverage": true, ".turbo": true, ".bun": true,
	".opencode": true, ".gemini": true, ".factory": true, ".codebuddy": true,
	".commandcode": true, ".pi": true, ".th-client": true, ".zencoder": true,
}
var ignoredFiles = map[string]bool{
	"bun.lock": true, "package-lock.json": true, "yarn.lock": true, "pnpm-lock.yaml": true,
	".DS_Store": true,
}

var sensitiveFiles = map[string]bool{
	".env": true, ".env.local": true, ".env.development": true, ".env.production": true,
	"id_rsa": true, "id_ed25519": true, "credentials.json": true, "service-account.json": true,
}

var codeExtensions = map[string]bool{
	".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".json": true, ".md": true,
	".yaml": true, ".yml": true, ".toml": true, ".sh": true, ".go": true, ".mod": true,
}

var secretRegex = regexp.MustCompile(`(?i)(?:sk-|AIza|ghp_|gho_|ghu_|ghs_|ghr_|SECRET|PASSWORD|TOKEN|KEY|PASS)[\w-]{10,}`)

// FileEntry represents a file or directory in the workspace.
type FileEntry struct {
	Path      string
	Type      string
	Extension string
	SizeBytes int64
	Content   string
}

// ProjectContext holds the aggregated metadata for a project.
type ProjectContext struct {
	RootPath      string
	Files         []FileEntry
	PackageJSON   map[string]interface{}
	HasTypeScript bool
	HasBun        bool
	HasGo         bool
	Directories   []string
	EntryPoints   []string
}

var archFiles = map[string]bool{
	"GEMINI.md": true, "ARCHON.md": true, "README.md": true, "go.mod": true, "package.json": true,
}

// ScanWorkspace traverses the given root path up to maxDepth to extract architectural context.
func ScanWorkspace(rootPath string, maxDepth int) (*ProjectContext, error) {
	ctx := &ProjectContext{
		RootPath:    rootPath,
		Files:       []FileEntry{},
		Directories: []string{},
	}

	var totalBytesRead int64 = 0

	// First pass: Find and read architectural files
	for archFile := range archFiles {
		path := filepath.Join(rootPath, archFile)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			raw, err := os.ReadFile(path)
			if err == nil {
				totalBytesRead += info.Size()
				ctx.Files = append(ctx.Files, FileEntry{
					Path:      archFile,
					Type:      "file",
					Extension: filepath.Ext(archFile),
					SizeBytes: info.Size(),
					Content:   secretRegex.ReplaceAllString(string(raw), "[REDACTED]"),
				})
			}
		}
	}

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return nil // ignore inaccessible files
		}

		relPath, _ := filepath.Rel(rootPath, path)
		if relPath == "." {
			return nil
		}

		depth := strings.Count(relPath, string(os.PathSeparator))
		if depth > maxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if ignoredDirs[d.Name()] {
				return fs.SkipDir
			}
			ctx.Directories = append(ctx.Directories, relPath)
		} else {
			if ignoredFiles[d.Name()] || sensitiveFiles[d.Name()] || archFiles[d.Name()] {
				return nil
			}
			ext := filepath.Ext(d.Name())
			info, err := d.Info()
			if err != nil {
				return nil
			}

			content := ""
			if codeExtensions[ext] {
				if totalBytesRead+info.Size() > maxGlobalBytes {
					content = "[FILE CONTENT OMITTED: GLOBAL CONTEXT LIMIT REACHED]"
				} else if info.Size() < 1024*1024 { // less than 1MB
					raw, err := os.ReadFile(path)
					if err == nil {
						totalBytesRead += info.Size()
						content = secretRegex.ReplaceAllString(string(raw), "[REDACTED]")
					}
				}
			}

			ctx.Files = append(ctx.Files, FileEntry{
				Path:      relPath,
				Type:      "file",
				Extension: ext,
				SizeBytes: info.Size(),
				Content:   content,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Try reading package.json
	pkgRaw, err := os.ReadFile(filepath.Join(rootPath, "package.json"))
	if err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(pkgRaw, &pkg) == nil {
			ctx.PackageJSON = pkg
		}
	}

	// Determine tech stack
	for _, f := range ctx.Files {
		if f.Path == "tsconfig.json" || f.Extension == ".ts" || f.Extension == ".tsx" {
			ctx.HasTypeScript = true
		}
		if f.Path == "bun.lock" {
			ctx.HasBun = true
		}
		if f.Path == "go.mod" || f.Extension == ".go" {
			ctx.HasGo = true
		}
		
		name := filepath.Base(f.Path)
		if name == "index.ts" || name == "index.js" || name == "main.ts" || name == "main.js" || name == "app.ts" || name == "main.go" {
			ctx.EntryPoints = append(ctx.EntryPoints, f.Path)
		}
	}

	if ctx.PackageJSON != nil {
		if devDeps, ok := ctx.PackageJSON["devDependencies"].(map[string]interface{}); ok {
			if _, hasTypesBun := devDeps["@types/bun"]; hasTypesBun {
				ctx.HasBun = true
			}
		}
	}

	return ctx, nil
}

// FormatContext transforms a ProjectContext into a formatted string for AI prompting.
func FormatContext(ctx *ProjectContext) string {
	var lines []string

	lines = append(lines, "=== PROJECT CONTEXT (scanned by Archon MCP) ===")
	lines = append(lines, "")
	lines = append(lines, "## Tech Stack Detection:")
	
	if ctx.HasGo {
		lines = append(lines, "- Runtime: Go")
	} else if ctx.HasBun {
		lines = append(lines, "- Runtime: Bun")
	} else {
		lines = append(lines, "- Runtime: Node.js")
	}

	if ctx.HasTypeScript {
		lines = append(lines, "- TypeScript: Yes")
	}
	lines = append(lines, "")

	if ctx.PackageJSON != nil {
		if deps, ok := ctx.PackageJSON["dependencies"].(map[string]interface{}); ok && len(deps) > 0 {
			lines = append(lines, "## Dependencies:")
			for k, v := range deps {
				lines = append(lines, fmt.Sprintf("- %s: %v", k, v))
			}
			lines = append(lines, "")
		}
		if scripts, ok := ctx.PackageJSON["scripts"].(map[string]interface{}); ok && len(scripts) > 0 {
			lines = append(lines, "## NPM/Bun Scripts:")
			for k, v := range scripts {
				lines = append(lines, fmt.Sprintf("- %s: %v", k, v))
			}
			lines = append(lines, "")
		}
	}

	lines = append(lines, "## Directory Structure:")
	for _, dir := range ctx.Directories {
		depth := strings.Count(dir, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)
		dirName := filepath.Base(dir)
		lines = append(lines, fmt.Sprintf("%s%s/", indent, dirName))
	}
	lines = append(lines, "")

	var filesWithContent []FileEntry
	for _, f := range ctx.Files {
		if f.Content != "" {
			filesWithContent = append(filesWithContent, f)
		}
	}

	if len(filesWithContent) > 0 {
		lines = append(lines, "## Source Code (Key Files Content - First 100 lines):")
		for _, f := range filesWithContent {
			lines = append(lines, fmt.Sprintf("### File: %s", f.Path))
			ext := f.Extension
			if len(ext) > 0 {
				ext = ext[1:]
			} else {
				ext = "text"
			}
			lines = append(lines, "```"+ext)
			
			contentLines := strings.Split(f.Content, "\n")
			limit := 100
			if len(contentLines) < limit {
				limit = len(contentLines)
			}
			
			lines = append(lines, strings.Join(contentLines[:limit], "\n"))
			if len(contentLines) > 100 {
				lines = append(lines, "// ... [Content truncated after 100 lines]")
			}
			lines = append(lines, "```", "")
		}
	}

	lines = append(lines, "=== END PROJECT CONTEXT ===")
	return strings.Join(lines, "\n")
}
