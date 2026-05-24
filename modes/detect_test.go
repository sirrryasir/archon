package modes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectMode_Greenfield(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "archon-detect-green-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Greenfield: empty directory
	res, err := DetectMode(tmpDir)
	if err != nil {
		t.Fatalf("failed to detect mode: %v", err)
	}

	if res.Mode != Greenfield {
		t.Errorf("expected Greenfield mode, got: %s (reason: %s)", res.Mode, res.Reason)
	}
}

func TestDetectMode_BrownfieldMarker(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "archon-detect-brown-marker-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json marker
	err = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	res, err := DetectMode(tmpDir)
	if err != nil {
		t.Fatalf("failed to detect mode: %v", err)
	}

	if res.Mode != Brownfield {
		t.Errorf("expected Brownfield mode (due to marker), got: %s", res.Mode)
	}
}

func TestDetectMode_BrownfieldFiles(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "archon-detect-brown-files-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create three source code files (which is our Brownfield threshold)
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "helper.py"), []byte("def run(): pass"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "utils.ts"), []byte("export const a = 1;"), 0644)

	res, err := DetectMode(tmpDir)
	if err != nil {
		t.Fatalf("failed to detect mode: %v", err)
	}

	if res.Mode != Brownfield {
		t.Errorf("expected Brownfield mode (due to file count), got: %s", res.Mode)
	}
}
