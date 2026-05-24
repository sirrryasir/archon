package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sirrryasir/archon/ai"
	"github.com/sirrryasir/archon/mcp"
)

var promptTimeout time.Duration

func init() {
	rootCmd.AddCommand(promptCmd)
	promptCmd.Flags().DurationVar(&promptTimeout, "timeout", 3*time.Minute, "Timeout for the prompt response (e.g. 30s, 5m)")
}

var promptCmd = &cobra.Command{
	Use:   "prompt [message]",
	Short: "Send a single prompt to Archon and print the response (non-interactive, for agent-to-agent use)",
	Long: `Send a single prompt to Archon and stream the response to stdout.
Diagnostic messages (thinking status) are written to stderr.
This makes it safe to pipe stdout to other tools:

  archon prompt "How should I structure my API?" | jq -R '.'
  archon prompt "Review my auth design" 2>/dev/null > review.md`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userInput := strings.Join(args, " ")
		runPrompt(userInput)
	},
}

func runPrompt(userInput string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting cwd: %v\n", err)
		os.Exit(1)
	}

	// Scan workspace for context
	projectCtx, err := mcp.ScanWorkspace(cwd, 4)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning workspace: %v\n", err)
		os.Exit(1)
	}

	contextBlock := mcp.FormatContext(projectCtx)
	gitCtx := mcp.BuildGitContext(cwd)
	if gitCtx != "" {
		contextBlock += "\n\n" + gitCtx
	}

	systemPrompt := ai.GuardianMessage + "\n\n" + contextBlock

	resolvedInput := ai.ResolvePromptFiles(userInput)
	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: resolvedInput},
	}

	engine := ai.NewEngine()

	ctx, cancel := context.WithTimeout(context.Background(), promptTimeout)
	defer cancel()

	fmt.Fprintf(os.Stderr, "[ Archon ] Thinking...\n")

	response, err := engine.ChatStream(ctx, messages, func(chunk string) {
		fmt.Print(chunk)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n[ Archon ] Error: %v\n", err)
		os.Exit(1)
	}

	// Ensure response ends with newline
	if response != "" && !strings.HasSuffix(response, "\n") {
		fmt.Println()
	}
}
