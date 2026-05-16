package tui

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirrryasir/archon/ai"
	"github.com/sirrryasir/archon/config"
	"github.com/sirrryasir/archon/session"
)

type errMsg error

type streamChunkMsg string
type streamDoneMsg string

// ChatModel implements tea.Model for the Archon conversational interface.
type ChatModel struct {
	viewport    viewport.Model
	textarea    textarea.Model
	engine      ai.Engine
	session     *session.SessionManager
	messages    []ai.Message
	err         error
	
	isStreaming bool
	streamBuffer string
	
	FileCount    int
	mdRenderer   *glamour.TermRenderer
}

func InitialModel(ctx context.Context, engine ai.Engine, sm *session.SessionManager, initialContext string, fileCount int) ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Ask Archon (type /help for commands)..."
	ta.Focus()
	ta.Prompt = InputPromptStyle.Render("❯ ")
	ta.CharLimit = 10000
	ta.SetWidth(138)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(140, 20)
	vp.SetContent("\nLoading workspace context...")

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(136),
	)

	m := ChatModel{
		textarea:   ta,
		viewport:   vp,
		engine:     engine,
		session:    sm,
		mdRenderer: renderer,
		FileCount:  fileCount,
	}

	welcomeMsg := `**Brownfield Mode** — Analyzing existing project

I've scanned your codebase and loaded the project context. 
Ask me anything about your architecture, and I'll reference specific files and patterns I observe.

Use **/review** for a full architecture review, or **/design** to generate design documents from our conversation.`

	// Initialize System and Welcome Message
	sysMsg := ai.Message{Role: "system", Content: initialContext}
	assistantWelcome := ai.Message{Role: "assistant", Content: welcomeMsg}
	
	m.messages = append(m.messages, sysMsg, assistantWelcome)
	_ = sm.AppendMessage(sysMsg)
	_ = sm.AppendMessage(assistantWelcome)

	return m
}

func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	// Filter out background query artifacts (OSC sequences)
	val := m.textarea.Value()
	if strings.Contains(val, "\x1b]") || strings.Contains(val, "]11;") || strings.Contains(val, "rgb:") {
		// Remove anything between \x1b] and \x07 or \x1b\\
		oscRegex := regexp.MustCompile(`\x1b\].*?(\x07|\x1b\\)`)
		cleanVal := oscRegex.ReplaceAllString(val, "")
		// Also catch fragments or raw responses
		cleanVal = strings.ReplaceAll(cleanVal, "]11;rgb:1818/1a1a/1d1d", "")
		cleanVal = strings.ReplaceAll(cleanVal, "]11;rgb:1818/1a1a/1d1d\\", "")
		if cleanVal != val {
			m.textarea.SetValue(cleanVal)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if msg.Alt {
				return m, tiCmd
			}
			
			if m.isStreaming {
				return m, nil
			}

			v := strings.TrimSpace(m.textarea.Value())
			if v == "" {
				return m, nil
			}

			if strings.HasPrefix(v, "/") {
				if v == "/exit" || v == "/quit" {
					return m, tea.Quit
				}
				m.textarea.Reset()
				m.executeSlashCommand(v)
				
				if m.isStreaming {
					m.textarea.Prompt = InputPromptProcessingStyle.Render("❯ ")
					return m, m.startStream(v)
				}
				
				// For non-streaming commands (like /help, /status)
				// The output is in m.streamBuffer (temporary) or m.messages
				m.refreshView()
				return m, nil
			}

			m.textarea.Reset()
			m.textarea.Prompt = InputPromptProcessingStyle.Render("❯ ")

			userMsg := ai.Message{Role: "user", Content: v}
			m.messages = append(m.messages, userMsg)
			_ = m.session.AppendMessage(userMsg)

			m.isStreaming = true
			m.streamBuffer = ""
			
			m.refreshView()
			
			return m, m.startStream(v)
		}

	case streamChunkMsg:
		m.streamBuffer += string(msg)
		m.refreshView()
		return m, nil

	case streamDoneMsg:
		m.isStreaming = false
		m.textarea.Prompt = InputPromptStyle.Render("❯ ")
		
		fullContent := string(msg)
		if fullContent != "" {
			archonMsg := ai.Message{Role: "assistant", Content: fullContent}
			m.messages = append(m.messages, archonMsg)
			_ = m.session.AppendMessage(archonMsg)
		}
		
		m.streamBuffer = ""
		
		if ai.NeedsCompaction(m.messages, config.GetModel()) {
			compacted, _ := ai.CompactHistory(context.Background(), m.engine, m.messages)
			m.messages = compacted
			_ = m.session.OverwriteHistory(m.messages)
		}
		
		m.refreshView()
		return m, nil

	case errMsg:
		m.err = msg
		m.textarea.Prompt = InputPromptStyle.Render("❯ ")
		return m, nil

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width - 32 // 30 for sidebar + 2 padding
		m.textarea.SetWidth(msg.Width - 4)
		
		// Update glamorous renderer for new width
		m.mdRenderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.viewport.Width-6),
		)

		// Calculate height
		bannerH := lipgloss.Height(m.renderBanner())
		statusH := 2 
		inputH := lipgloss.Height(InputBoxStyle.Width(msg.Width).Render(m.textarea.View()))
		
		m.viewport.Height = msg.Height - bannerH - statusH - inputH
		if m.viewport.Height < 0 { m.viewport.Height = 1 }
		
		m.refreshView()
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *ChatModel) refreshView() {
	var b strings.Builder

	for _, msg := range m.messages {
		if msg.Role == "system" {
			continue
		}
		if msg.Role == "user" {
			b.WriteString(UserLabel + "\n")
			b.WriteString(UserMsgStyle.Render(msg.Content) + "\n")
		} else if msg.Role == "assistant" {
			trimmed := strings.TrimSpace(msg.Content)
			// Robust JSON/Architecture detection
			isJSON := (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) || 
			          (strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))
			
			if isJSON && strings.Contains(trimmed, ":") {
				b.WriteString(ArchonLabel + "\n")
				b.WriteString(JSONBoxStyle.Render(trimmed) + "\n")
			} else if strings.Contains(trimmed, "## SYSTEM QUALITIES EVALUATION") || strings.Contains(trimmed, "## CONCLUSION") {
				b.WriteString(ArchonLabel + "\n")
				rendered, _ := m.mdRenderer.Render(msg.Content)
				b.WriteString(ReviewBoxStyle.Render(rendered) + "\n")
			} else if strings.HasSuffix(trimmed, "?") {
				b.WriteString(ArchonLabel + "\n")
				rendered, _ := m.mdRenderer.Render(msg.Content)
				b.WriteString(SocraticQuestStyle.Render(rendered) + "\n")
			} else if strings.Contains(trimmed, "Advice:") || strings.Contains(trimmed, "Strategy:") {
				b.WriteString(ArchonLabel + "\n")
				rendered, _ := m.mdRenderer.Render(msg.Content)
				b.WriteString(GoldenAdviceStyle.Render(rendered) + "\n")
			} else if strings.HasPrefix(msg.Content, "**System Status**") {
				b.WriteString("\n" + CommandResultStyle.Render(msg.Content) + "\n")
			} else {
				b.WriteString(ArchonLabel + "\n")
				rendered, err := m.mdRenderer.Render(msg.Content)
				if err != nil {
					b.WriteString(ArchonMsgStyle.Render(msg.Content) + "\n")
				} else {
					b.WriteString(ArchonMsgStyle.Render(rendered))
				}
			}
		}
	}

	if m.isStreaming {
		b.WriteString(ArchonLabel + "\n")
		if m.streamBuffer == "" {
			b.WriteString(ArchonMsgStyle.Render(lipgloss.NewStyle().Foreground(archonGray).Italic(true).Render("Archon is thinking...")) + "\n")
		} else {
			rendered, err := m.mdRenderer.Render(m.streamBuffer)
			if err != nil {
				b.WriteString(ArchonMsgStyle.Render(m.streamBuffer))
			} else {
				b.WriteString(ArchonMsgStyle.Render(rendered))
			}
		}
	}

	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

func (m ChatModel) renderSidebar() string {
	sidebarWidth := 30
	style := SidebarStyle.Width(sidebarWidth).Height(m.viewport.Height)
	
	titleStyle := lipgloss.NewStyle().Foreground(archonCyan).Bold(true).Underline(true)
	itemStyle := lipgloss.NewStyle().Foreground(archonGray)
	activeItemStyle := lipgloss.NewStyle().Foreground(archonGreen).Bold(true)

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("SYSTEM STATE"),
		itemStyle.Render(fmt.Sprintf("Mode:   %s", "BROWNFIELD")),
		itemStyle.Render(fmt.Sprintf("Model:  %s", "QWEN3-480B")),
		itemStyle.Render(fmt.Sprintf("Files:  %d", m.FileCount)),
		"",
		titleStyle.Render("COMMANDS"),
		itemStyle.Render("/review - Full Audit"),
		itemStyle.Render("/design - Architect"),
		itemStyle.Render("/help   - List All"),
		"",
		titleStyle.Render("PROJECT INFO"),
		activeItemStyle.Render("• archon (current)"),
		itemStyle.Render("  ├─ ai/"),
		itemStyle.Render("  ├─ mcp/"),
		itemStyle.Render("  └─ tui/"),
	)
	
	return style.Render(content)
}

func (m ChatModel) renderBanner() string {
	ascii := `    █████╗ ██████╗  ██████╗██╗  ██╗ ██████╗ ███╗   ██╗    
   ██╔══██╗██╔══██╗██╔════╝██║  ██║██╔═══██╗████╗  ██║    
   ███████║██████╔╝██║     ███████║██║   ██║██╔██╗ ██║    
   ██╔══██║██╔══██╗██║     ██╔══██║██║   ██║██║╚██╗██║    
   ██║  ██║██║  ██║╚██████╗██║  ██║╚██████╔╝██║ ╚████║    
   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝    `
	
	bannerText := BannerStyle.Render(ascii)
	bannerBox := BannerBoxStyle.Render(bannerText)

	badges := lipgloss.JoinHorizontal(lipgloss.Top,
		BrownfieldBadge,
		"  ",
		ModelBadge.Render("MODEL: "+strings.ToUpper(config.GetModel())),
		"  ",
		FilesBadge.Render(fmt.Sprintf("FILES: %d", m.FileCount)),
	)

	header := lipgloss.JoinHorizontal(lipgloss.Center, 
		bannerBox, 
		lipgloss.NewStyle().MarginLeft(4).Render(badges),
	)

	return HeaderStyle.Width(m.viewport.Width + 32).Render(header)
}

func (m ChatModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n", m.err)
	}

	// 1. Calculate Status Bar
	tokens := ai.EstimateMessageTokens(m.messages)
	limit := ai.GetModelContextLimit(config.GetModel())
	
	tokenColor := archonGreen
	ratio := float64(tokens) / float64(limit)
	if ratio > 0.9 {
		tokenColor = archonRed
	} else if ratio > 0.8 {
		tokenColor = archonYellow
	}

	msgCount := len(m.messages) - 2 // Exclude system prompt and welcome message
	if msgCount < 0 { msgCount = 0 }

	leftStatus := lipgloss.JoinHorizontal(lipgloss.Top,
		StatusLabel("EXIT "), StatusValue("^C"), StatusSep,
		StatusLabel("HELP "), StatusValue("/?"),
	)

	rightStatus := lipgloss.JoinHorizontal(lipgloss.Top,
		StatusLabel("TOKENS: "), lipgloss.NewStyle().Foreground(tokenColor).Render(ai.FormatTokenCount(tokens)), 
		StatusLabel(" / "+ai.FormatTokenCount(limit)), StatusSep,
		StatusLabel("MSG: "), StatusValue(fmt.Sprintf("%d", msgCount)), StatusSep,
		StatusLabel("MODEL: "), lipgloss.NewStyle().Foreground(archonBlue).Render(strings.ToUpper(config.GetModel())), StatusSep,
		StatusLabel("SESSION: "), StatusValue("Go-Port"),
	)

	spaces := m.viewport.Width + 32 - lipgloss.Width(leftStatus) - lipgloss.Width(rightStatus) - 2
	if spaces < 1 { spaces = 1 }
	
	statusLine := lipgloss.JoinHorizontal(lipgloss.Top, leftStatus, strings.Repeat(" ", spaces), rightStatus)
	renderedStatus := StatusBarContainer.Width(m.viewport.Width + 32).Render(statusLine)

	// 2. Render Input Box
	var inputBox string
	if m.isStreaming {
		loadingText := lipgloss.NewStyle().Foreground(archonYellow).Bold(true).Italic(true).Render("Archon is thinking...") + 
					   lipgloss.NewStyle().Foreground(archonGray).Render(" (Press Esc to cancel)")
		inputBox = InputBoxProcessingStyle.Width(m.viewport.Width + 32).Render(m.textarea.View() + "\n" + loadingText)
	} else {
		inputBox = InputBoxStyle.Width(m.viewport.Width + 32).Render(m.textarea.View())
	}

	// 3. Assemble Layout with Split-View
	mainView := lipgloss.JoinHorizontal(lipgloss.Top,
		m.viewport.View(),
		m.renderSidebar(),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderBanner(),
		mainView,
		inputBox,
		renderedStatus,
	)
}

func (m ChatModel) startStream(prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		full, err := m.engine.ChatStream(ctx, m.messages, func(chunk string) {
			// Real-time streaming would go here if using channels
		})
		
		if err != nil {
			return errMsg(err)
		}
		
		return streamDoneMsg(full)
	}
}
