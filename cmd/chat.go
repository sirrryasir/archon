package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/sirrryasir/archon/ai"
	"github.com/sirrryasir/archon/mcp"
	"github.com/sirrryasir/archon/session"
	"github.com/sirrryasir/archon/tui"
)

func init() {
	rootCmd.AddCommand(chatCmd)
}

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive Socratic chat session",
	Run: func(cmd *cobra.Command, args []string) {
		StartChatSession()
	},
}

func StartChatSession() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	fmt.Println("Scanning workspace...")
	
	// Scan workspace (depth 4)
	projectCtx, err := mcp.ScanWorkspace(cwd, 4)
	if err != nil {
		fmt.Println("Error scanning workspace:", err)
		return
	}

	// Initialize session manager
	sm, err := session.NewSessionManager("default-session", cwd)
	if err != nil {
		fmt.Println("Error initializing session:", err)
		return
	}
	defer sm.Close()

	// Build system prompt
	contextBlock := mcp.FormatContext(projectCtx)
	gitCtx := mcp.BuildGitContext(cwd)
	if gitCtx != "" {
		contextBlock += "\n\n" + gitCtx
	}
	
	systemPrompt := ai.GuardianMessage + "\n\n" + contextBlock

	// Initialize AI Engine
	engine := ai.NewEngine()

	// Initialize BubbleTea Model
	p := tea.NewProgram(
		tui.InitialModel(context.Background(), engine, sm, systemPrompt, len(projectCtx.Files)),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
