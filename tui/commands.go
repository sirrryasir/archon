package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/sirrryasir/archon/ai"
	"github.com/sirrryasir/archon/config"
	"github.com/sirrryasir/archon/design"
	"github.com/sirrryasir/archon/mcp"
	"github.com/sirrryasir/archon/modes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

type designDocsReadyMsg struct {
	docs design.DesignDocs
	err  error
}

func (m *ChatModel) executeSlashCommand(cmd string) tea.Cmd {
	parts := strings.SplitN(cmd, " ", 2)
	baseCmd := parts[0]
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	m.isStreaming = false

	switch baseCmd {
	case "/exit", "/quit":
		// Handled directly in update loop
		return nil
	case "/help", "/?":
		help := "**Available Commands:**\n\n" +
			"• **/review** - Perform a deep Socratic review of the codebase.\n" +
			"• **/design** - Generate system architecture, design decisions, and specifications blueprints asynchronously.\n" +
			"• **/export [output-dir]** - Export generated blueprints and AGENTS.md instructions to disk.\n" +
			"• **/diff** - Review uncommitted git changes from an architectural perspective.\n" +
			"• **/mode [greenfield|brownfield]** - Show or switch project mode.\n" +
			"• **/copy** - Copy the last AI response to the system clipboard.\n" +
			"• **/status** - Show current context usage and settings.\n" +
			"• **/clear** - Reset the conversation history.\n" +
			"• **/exit** - Close Archon."
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: help})
		return nil

	case "/clear":
		m.messages = m.messages[:1] // Keep system prompt
		_ = m.session.OverwriteHistory(m.messages)
		welcomeMsg := `**Archon v1.0.0 • Software Architect**
Using model **` + strings.ToUpper(config.GetModel()) + `** • ` + fmt.Sprintf("**%d**", m.FileCount) + ` files loaded

` + modes.BrownfieldWelcome
		if m.mode == modes.Greenfield {
			welcomeMsg = `**Archon v1.0.0 • Software Architect**
Using model **` + strings.ToUpper(config.GetModel()) + `**

` + modes.GreenfieldWelcome
		}
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: welcomeMsg})
		return nil

	case "/copy":
		var lastAssistant string
		for i := len(m.messages) - 1; i >= 0; i-- {
			if m.messages[i].Role == "assistant" {
				lastAssistant = m.messages[i].Content
				break
			}
		}
		if lastAssistant != "" {
			err := clipboard.WriteAll(lastAssistant)
			if err == nil {
				m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "_Last assistant response copied to clipboard!_"})
			} else {
				m.messages = append(m.messages, ai.Message{Role: "assistant", Content: fmt.Sprintf("⚠️ **Failed to copy**: %v", err)})
			}
		} else {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "⚠️ No assistant message found to copy."})
		}
		return nil

	case "/status":
		tokens := ai.EstimateMessageTokens(m.messages)
		limit := ai.GetModelContextLimit(config.GetModel())
		ratio := float64(tokens) / float64(limit) * 100
		status := fmt.Sprintf("**System Status**\n- **Model**: %s\n- **Provider**: %s\n- **Mode**: %s\n- **Project Path**: %s\n- **Tokens**: %d / %d (%.2f%%)",
			config.GetModel(), config.GetProvider(), m.mode, m.cwd, tokens, limit, ratio)
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: status})
		return nil

	case "/compact":
		compacted, err := ai.CompactHistory(context.Background(), m.engine, m.messages)
		if err == nil {
			m.messages = compacted
			_ = m.session.OverwriteHistory(m.messages)
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "_Context compaction successful._"})
		} else {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: fmt.Sprintf("**Error during compaction**: %v", err)})
		}
		return nil

	case "/model":
		if args == "" {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "Usage: `/model <alias or name>`"})
			return nil
		}
		
		alias := ai.ResolveModelAlias(args)
		if alias != nil {
			viper.Set("model", alias.Model)
			viper.Set("provider", alias.Provider)
		} else {
			viper.Set("model", args)
		}
		
		m.engine = ai.NewEngine()
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: fmt.Sprintf("_Model switched to **%s** (%s)_", config.GetModel(), config.GetProvider())})
		return nil

	case "/review":
		prompt := "Please conduct a comprehensive architectural review of the codebase based on the provided context."
		m.startStreamingPrompt(prompt)

	case "/design":
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "_Generating design documents in the background (ARCHITECTURE.md, DESIGN.md, SYSTEM.md)..._"})
		m.isStreaming = true
		return m.generateDesignDocsCmd()

	case "/export":
		if m.designDocs == nil {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "⚠️ No design documents found. Please run `/design` first to generate blueprints."})
			return nil
		}
		outDir := m.cwd
		if args != "" {
			outDir = args
		}
		written, err := design.ExportDesignDocs(*m.designDocs, outDir)
		if err != nil {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: fmt.Sprintf("⚠️ **Export failed**: %v", err)})
			return nil
		}
		var builder strings.Builder
		builder.WriteString("**✓ Design documents successfully exported to disk!**\n")
		for _, path := range written {
			builder.WriteString(fmt.Sprintf("- %s\n", filepath.Base(path)))
		}
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: builder.String()})
		return nil

	case "/diff":
		if !mcp.IsGitRepo(m.cwd) {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "⚠️ Not a git repository."})
			return nil
		}
		diff := mcp.GetGitDiff(m.cwd, 50000)
		if diff == "No diff output." {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "✓ No uncommitted changes to review."})
			return nil
		}
		prompt := fmt.Sprintf("Please review the following uncommitted git changes from an architectural perspective. Identify any potential concerns or improvements:\n\n%s", diff)
		m.startStreamingPrompt(prompt)

	case "/mode":
		if args == "" {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: fmt.Sprintf("Current mode: **%s**", m.mode)})
			return nil
		}
		newMode := modes.ProjectMode(strings.ToLower(strings.TrimSpace(args)))
		if newMode != modes.Greenfield && newMode != modes.Brownfield {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "Usage: `/mode <greenfield|brownfield>`"})
			return nil
		}
		m.mode = newMode
		if m.mode == modes.Greenfield {
			m.messages[0] = ai.Message{Role: "system", Content: modes.GreenfieldSystemPrompt}
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: fmt.Sprintf("_Switched to **%s** mode._\n\n%s", m.mode, modes.GreenfieldWelcome)})
			_ = m.session.OverwriteHistory(m.messages)
		} else {
			m.isScanning = true
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: fmt.Sprintf("_Switched to **%s** mode. Scanning workspace..._", m.mode)})
			_ = m.session.OverwriteHistory(m.messages)
			return m.scanWorkspaceCmd()
		}
		return nil

	default:
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "Unknown command. Try **/help**."})
	}
	return nil
}

func (m *ChatModel) generateDesignDocsCmd() tea.Cmd {
	return func() tea.Msg {
		convCtx := design.ExtractConversationContext(m.messages)

		// Generate Architecture.md
		archPrompt := design.BuildDesignPrompt("architecture", convCtx)
		archRes, err := m.engine.ChatStream(context.Background(), []ai.Message{{Role: "user", Content: archPrompt}}, func(chunk string) {})
		if err != nil {
			return designDocsReadyMsg{err: err}
		}

		// Generate Design.md
		designPrompt := design.BuildDesignPrompt("design", convCtx)
		designRes, err := m.engine.ChatStream(context.Background(), []ai.Message{{Role: "user", Content: designPrompt}}, func(chunk string) {})
		if err != nil {
			return designDocsReadyMsg{err: err}
		}

		// Generate System.md
		systemPrompt := design.BuildDesignPrompt("system", convCtx)
		systemRes, err := m.engine.ChatStream(context.Background(), []ai.Message{{Role: "user", Content: systemPrompt}}, func(chunk string) {})
		if err != nil {
			return designDocsReadyMsg{err: err}
		}

		return designDocsReadyMsg{
			docs: design.DesignDocs{
				Architecture: archRes,
				Design:       designRes,
				System:       systemRes,
			},
		}
	}
}

func (m *ChatModel) startStreamingPrompt(prompt string) {
	userMsg := ai.Message{Role: "user", Content: prompt}
	m.messages = append(m.messages, userMsg)
	_ = m.session.AppendMessage(userMsg)

	m.isStreaming = true
	m.streamBuffer = ""
}
