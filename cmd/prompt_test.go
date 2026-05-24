package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput intercepts stdout and stderr during the execution of f.
func captureOutput(f func()) (stdout string, stderr string, err error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	outChan := make(chan string)
	errChan := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outChan <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		errChan <- buf.String()
	}()

	// Execute the function
	f()

	// Restore original streams
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout = <-outChan
	stderr = <-errChan
	return stdout, stderr, nil
}

func TestRunPrompt_SocraticResponse(t *testing.T) {
	// Set mock response in environment
	mockResp := "How does your user service manage session caching?"
	os.Setenv("ARCHON_MOCK_RESPONSE", mockResp)
	defer os.Unsetenv("ARCHON_MOCK_RESPONSE")

	stdout, stderr, err := captureOutput(func() {
		runPrompt("Explain user service design pattern")
	})
	if err != nil {
		t.Fatalf("failed to capture output: %v", err)
	}

	// Verify that thinking logs appear in stderr and final response in stdout
	if !strings.Contains(stderr, "[ Archon ] Thinking...") {
		t.Errorf("expected stderr to contain thinking state, got: %s", stderr)
	}
	if !strings.Contains(stdout, mockResp) {
		t.Errorf("expected stdout to contain mock response, got: %s", stdout)
	}
}

func TestRunPrompt_GuardianIntercepted(t *testing.T) {
	// Set mock response containing a forbidden code fence block
	mockResp := "Here is the code:\n```go\npackage main\nfunc main() {}\n```"
	os.Setenv("ARCHON_MOCK_RESPONSE", mockResp)
	defer os.Unsetenv("ARCHON_MOCK_RESPONSE")

	stdout, _, err := captureOutput(func() {
		runPrompt("Write a Go function")
	})
	if err != nil {
		t.Fatalf("failed to capture output: %v", err)
	}

	// Verify that the Output Guardian intercepted the response
	if !strings.Contains(stdout, "[GUARDIAN INTERCEPTED]") {
		t.Errorf("expected output to be intercepted, got stdout: %q", stdout)
	}
	// Verify that the actual code block content was blocked from stdout
	if strings.Contains(stdout, "package main") {
		t.Error("expected code block to be completely redacted and not printed to stdout")
	}
}
