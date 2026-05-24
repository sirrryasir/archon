package modes

import (
	"os"
	"path/filepath"
	"strings"
)

// ProjectMode represents the state of the project: empty/new (greenfield) or existing (brownfield).
type ProjectMode string

const (
	Greenfield ProjectMode = "greenfield"
	Brownfield ProjectMode = "brownfield"
)

// ProjectMarkers are files whose existence immediately flags the project as brownfield.
var ProjectMarkers = []string{
	"package.json",
	"Cargo.toml",
	"pyproject.toml",
	"go.mod",
	"pom.xml",
	"build.gradle",
	"Makefile",
	"CMakeLists.txt",
	"composer.json",
	"Gemfile",
	"requirements.txt",
	"setup.py",
	"tsconfig.json",
	"deno.json",
	".sln",
}

// SourceExtensions are extensions that represent real code in a project.
var SourceExtensions = map[string]bool{
	".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".py": true, ".rs": true, ".go": true,
	".java": true, ".rb": true, ".php": true, ".c": true, ".cpp": true, ".cs": true, ".swift": true,
}

// DetectResult contains the detected mode and description of the detection result.
type DetectResult struct {
	Mode      ProjectMode
	Reason    string
	FileCount int
}

// DetectMode determines whether a directory is Greenfield or Brownfield.
func DetectMode(rootPath string) (DetectResult, error) {
	// 1. Check for standard project markers
	for _, marker := range ProjectMarkers {
		path := filepath.Join(rootPath, marker)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return DetectResult{
				Mode:      Brownfield,
				Reason:    "Found " + marker,
				FileCount: -1, // will be counted by the workspace scanner
			}, nil
		}
	}

	// 2. Scan the root directory and 1 level deep for source files
	sourceCount := 0
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return DetectResult{
			Mode:      Greenfield,
			Reason:    "Cannot read directory: " + err.Error(),
			FileCount: 0,
		}, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if SourceExtensions[ext] {
				sourceCount++
			}
		} else {
			// Skip hidden dirs (like .git, .archon, node_modules)
			if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" {
				continue
			}

			// Read 1 level deep
			subEntries, err := os.ReadDir(filepath.Join(rootPath, entry.Name()))
			if err == nil {
				for _, sub := range subEntries {
					if !sub.IsDir() {
						ext := strings.ToLower(filepath.Ext(sub.Name()))
						if SourceExtensions[ext] {
							sourceCount++
						}
					}
				}
			}
		}
	}

	if sourceCount >= 3 {
		return DetectResult{
			Mode:      Brownfield,
			Reason:    "Found existing source files",
			FileCount: sourceCount,
		}, nil
	}

	reason := "Empty directory — no project files detected"
	if sourceCount > 0 {
		reason = "Only minor source files detected — treating as Greenfield"
	}

	return DetectResult{
		Mode:      Greenfield,
		Reason:    reason,
		FileCount: sourceCount,
	}, nil
}
