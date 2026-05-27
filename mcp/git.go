package mcp

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// IsGitRepo checks if a given directory is inside a git repository.
func IsGitRepo(rootPath string) bool {
	_, err := os.Stat(filepath.Join(rootPath, ".git"))
	return err == nil
}

// runGit executes a git command in the specified directory and returns its output.
func runGit(rootPath string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = rootPath

	// Set a reasonable timeout to prevent hanging
	timer := time.AfterFunc(5*time.Second, func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	})
	defer timer.Stop()

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}

// GetGitDiffSummary returns a summary of staged and unstaged changes.
func GetGitDiffSummary(rootPath string) string {
	unstaged := runGit(rootPath, "diff", "--stat")
	staged := runGit(rootPath, "diff", "--cached", "--stat")

	var parts []string
	if staged != "" {
		parts = append(parts, "### Staged Changes:\n"+staged)
	}
	if unstaged != "" {
		parts = append(parts, "### Unstaged Changes:\n"+unstaged)
	}

	if len(parts) == 0 {
		return "No uncommitted changes detected."
	}
	return strings.Join(parts, "\n\n")
}

// GetGitDiff returns the full diff of uncommitted changes.
func GetGitDiff(rootPath string, maxBytes int) string {
	diff := runGit(rootPath, "diff")
	stagedDiff := runGit(rootPath, "diff", "--cached")

	var combined string
	if stagedDiff != "" {
		combined += "--- STAGED ---\n" + stagedDiff + "\n"
	}
	if diff != "" {
		combined += "--- WORKING TREE ---\n" + diff + "\n"
	}

	if len(combined) > maxBytes {
		return combined[:maxBytes] + fmt.Sprintf("\n\n[... diff truncated at %dKB ...]", maxBytes/1024)
	}
	if combined == "" {
		return "No diff output."
	}
	return combined
}

// GetRecentCommits returns recent commit history (short format).
func GetRecentCommits(rootPath string, count int) string {
	return runGit(rootPath, "log", "--oneline", fmt.Sprintf("-%d", count), "--no-merges")
}

// GetCurrentBranch returns the current branch name.
func GetCurrentBranch(rootPath string) string {
	return runGit(rootPath, "rev-parse", "--abbrev-ref", "HEAD")
}

// BuildGitContext builds a complete git context block for the system prompt.
func BuildGitContext(rootPath string) string {
	if !IsGitRepo(rootPath) {
		return ""
	}

	branch := GetCurrentBranch(rootPath)
	if branch == "" {
		branch = "detached HEAD"
	}

	diffSummary := GetGitDiffSummary(rootPath)
	recentCommits := GetRecentCommits(rootPath, 5)

	var lines []string
	lines = append(lines, "## Git Context", fmt.Sprintf("Branch: %s", branch), "", diffSummary)

	if recentCommits != "" {
		lines = append(lines, "", "### Recent Commits:", recentCommits)
	}

	return strings.Join(lines, "\n")
}
