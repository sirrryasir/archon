package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirrryasir/archon/ai"
	"github.com/sirrryasir/archon/config"
	"github.com/spf13/viper"
)

func (m *ChatModel) executeSlashCommand(cmd string) {
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
	case "/help", "/?":
		help := "**Available Commands:**\n\n• **/review** - Perform a deep architectural audit.\n• **/design** - Generate design documentation.\n• **/status** - Show usage stats.\n• **/clear** - Reset history.\n• **/exit** - Close Archon."
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: help})

	case "/clear":
		m.messages = m.messages[:1] // Keep system prompt
		_ = m.session.OverwriteHistory(m.messages)
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "_Session cleared. Context retained._"})

	case "/status":
		tokens := ai.EstimateMessageTokens(m.messages)
		limit := ai.GetModelContextLimit(config.GetModel())
		ratio := float64(tokens) / float64(limit) * 100
		status := fmt.Sprintf("**System Status**\n- **Model**: %s\n- **Provider**: %s\n- **Tokens**: %d / %d (%.2f%%)",
			config.GetModel(), config.GetProvider(), tokens, limit, ratio)
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: status})

	case "/compact":
		compacted, err := ai.CompactHistory(context.Background(), m.engine, m.messages)
		if err == nil {
			m.messages = compacted
			_ = m.session.OverwriteHistory(m.messages)
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "_Context compaction successful._"})
		} else {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: fmt.Sprintf("**Error during compaction**: %v", err)})
		}

	case "/model":
		if args == "" {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "Usage: `/model <alias or name>`"})
			return
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

	case "/review":
		prompt := "Please conduct a comprehensive architectural review of the codebase based on the provided context."
		m.startStreamingPrompt(prompt)

	case "/design":
		prompt := "Generate a complete `ARCHITECTURE.md` file for this project."
		m.startStreamingPrompt(prompt)

	default:
		m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "Unknown command. Try **/help**."})
	}
}

func (m *ChatModel) startStreamingPrompt(prompt string) {
	userMsg := ai.Message{Role: "user", Content: prompt}
	m.messages = append(m.messages, userMsg)
	_ = m.session.AppendMessage(userMsg)

	m.isStreaming = true
	m.streamBuffer = ""
}
