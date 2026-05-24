package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanWorkspace(t *testing.T) {
	// Create a temp workspace directory
	tmpDir, err := os.MkdirTemp("", "archon-scanner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create structure:
	// - README.md (arch file)
	// - go.mod (arch file)
	// - src/main.go (clean source)
	// - src/config.json (contains secret)
	// - .git/config (should be ignored)
	// - node_modules/index.js (should be ignored)

	err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Archon Project"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	srcDir := filepath.Join(tmpDir, "src")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(srcDir, "config.json"), []byte(`{"api_key": "sk-1234567890abcdef123456"}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(gitDir, "config"), []byte("some git config"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	nodeDir := filepath.Join(tmpDir, "node_modules")
	if err := os.Mkdir(nodeDir, 0755); err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(nodeDir, "index.js"), []byte("console.log('hello');"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Run concurrent scan
	ctx, err := ScanWorkspace(tmpDir, 4)
	if err != nil {
		t.Fatalf("ScanWorkspace failed: %v", err)
	}

	// Verify tech stack
	if !ctx.HasGo {
		t.Error("expected HasGo to be true")
	}

	// Verify ignored directories are not present
	for _, dir := range ctx.Directories {
		if strings.HasPrefix(dir, ".git") || strings.HasPrefix(dir, "node_modules") {
			t.Errorf("found ignored directory in results: %s", dir)
		}
	}

	// Verify files count
	if len(ctx.Files) == 0 {
		t.Fatal("expected files to be scanned, got 0")
	}

	// Verify secret redaction
	foundSecretConfig := false
	for _, file := range ctx.Files {
		if file.Path == filepath.Join("src", "config.json") || file.Path == "src/config.json" {
			foundSecretConfig = true
			if !strings.Contains(file.Content, "[REDACTED]") {
				t.Errorf("expected secret config.json to be redacted, got: %s", file.Content)
			}
			if strings.Contains(file.Content, "sk-12345") {
				t.Error("secret was not redacted completely")
			}
		}
	}

	if !foundSecretConfig {
		t.Error("expected to find src/config.json in scanned files")
	}
}
