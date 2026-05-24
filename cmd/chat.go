package cmd

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/sirrryasir/archon/ai"
	"github.com/sirrryasir/archon/config"
	"github.com/sirrryasir/archon/modes"
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
	lipgloss.SetHasDarkBackground(true)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// Detect Greenfield vs Brownfield mode
	res, err := modes.DetectMode(cwd)
	if err != nil {
		res = modes.DetectResult{Mode: modes.Greenfield, Reason: err.Error()}
	}

	// Generate a stable, project-specific session ID (short hash of project path)
	hash := md5.Sum([]byte(cwd))
	projectName := filepath.Base(cwd)
	sessionID := fmt.Sprintf("%s-%x", strings.ToLower(projectName), hash[:4])

	// Initialize session manager
	sm, err := session.NewSessionManager(sessionID, cwd)
	if err != nil {
		fmt.Println("Error initializing session:", err)
		return
	}
	defer sm.Close()

	// Load existing history and trim to stay within context limits
	history, _ := sm.LoadHistory()
	if len(history) > 0 {
		modelName := config.GetModel()
		limit := ai.GetModelContextLimit(modelName)
		safeLimit := int(float64(limit) * 0.70) // use 70% of limit to leave room for new msgs

		// Keep trimming from the front (oldest messages) until we're under the limit
		for len(history) > 2 {
			tokens := ai.EstimateMessageTokens(history)
			if tokens <= safeLimit {
				break
			}
			// Remove oldest non-system message
			history = append(history[:1], history[2:]...)
		}

		// If still too large, start fresh
		if ai.EstimateMessageTokens(history) > safeLimit {
			history = nil
		}
	}

	// Initialize AI Engine
	engine := ai.NewEngine()

	// Initialize BubbleTea Model with asynchronous scanning context
	model := tui.InitialModelWithHistory(context.Background(), engine, sm, cwd, res.Mode, history)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI session: %v\n", err)
		os.Exit(1)
	}
}
