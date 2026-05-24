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
	"github.com/sirrryasir/archon/design"
	"github.com/sirrryasir/archon/mcp"
	"github.com/sirrryasir/archon/modes"
	"github.com/sirrryasir/archon/session"
)

var availableCommands = []string{
	"/review",
	"/design",
	"/copy",
	"/clear",
	"/status",
	"/compact",
	"/model",
	"/export",
	"/diff",
	"/mode",
	"/help",
	"/quit",
}

// Braille spinner frames — same as Gemini CLI / Claude Code
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type errMsg error
type tickMsg time.Time

// nextChunkMsg represents a streamed chunk or error from the AI channel.
type nextChunkMsg struct {
	ch   <-chan string
	val  string
	err  error
	done bool
}

type scanCompleteMsg struct {
	projectCtx *mcp.ProjectContext
	err        error
}

// ChatModel implements tea.Model for the Archon conversational interface.
type ChatModel struct {
	viewport    viewport.Model
	textarea    textarea.Model
	engine      ai.Engine
	session     *session.SessionManager
	messages    []ai.Message
	err         error

	isStreaming  bool
	streamBuffer string
	spinnerIdx   int
	elapsedSecs  int

	// Prompt History
	history    []string
	historyIdx int
	tempInput  string

	// Slash Commands Autocomplete Menu
	showSuggest bool
	suggestIdx  int
	suggestions []string

	width      int
	height     int
	FileCount  int
	mdRenderer *glamour.TermRenderer

	// Missing Greenfield/Brownfield mode integration
	cwd        string
	mode       modes.ProjectMode
	isScanning bool
	designDocs *design.DesignDocs
}

func InitialModel(_ context.Context, engine ai.Engine, sm *session.SessionManager, cwd string, mode modes.ProjectMode) ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Ask anything... or type /help for commands"
	ta.Focus()
	ta.Prompt = InputPromptStyle.Render("❯ ")
	ta.CharLimit = 10000
	ta.SetWidth(100)
	ta.SetHeight(1)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(100, 20)

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(96),
	)

	m := ChatModel{
		textarea:   ta,
		viewport:   vp,
		engine:     engine,
		session:    sm,
		mdRenderer: renderer,
		FileCount:  0,
		width:      100,
		height:     40,
		history:    []string{},
		historyIdx: 0,
		cwd:        cwd,
		mode:       mode,
	}

	var welcomeMsg string
	var sysPrompt string

	if mode == modes.Greenfield {
		m.isScanning = false
		welcomeMsg = `**Archon v1.0.0 • Software Architect**
Using model **` + strings.ToUpper(config.GetModel()) + `**

` + modes.GreenfieldWelcome
		sysPrompt = modes.GreenfieldSystemPrompt
	} else {
		m.isScanning = true
		welcomeMsg = "\nLoading workspace context..."
		sysPrompt = ai.GuardianMessage // temporary baseline, will be updated asynchronously
	}

	vp.SetContent(welcomeMsg)

	sysMsg := ai.Message{Role: "system", Content: sysPrompt}
	assistantWelcome := ai.Message{Role: "assistant", Content: welcomeMsg}

	m.messages = append(m.messages, sysMsg, assistantWelcome)
	_ = sm.AppendMessage(sysMsg)
	_ = sm.AppendMessage(assistantWelcome)

	// Call refreshView immediately to render the welcome metadata and purge the "Loading" placeholder
	m.refreshView()

	return m
}

// InitialModelWithHistory creates a model and restores trimmed history from a previous session.
func InitialModelWithHistory(_ context.Context, engine ai.Engine, sm *session.SessionManager, cwd string, mode modes.ProjectMode, history []ai.Message) ChatModel {
	m := InitialModel(context.Background(), engine, sm, cwd, mode)

	if len(history) > 0 {
		// Replace default messages with trimmed history (skip system messages in history,
		// we already injected the current system prompt)
		var restored []ai.Message
		restored = append(restored, m.messages[0]) // keep new system prompt
		for _, msg := range history {
			if msg.Role != "system" {
				restored = append(restored, msg)
			}
		}
		m.messages = restored
	}

	// Call refreshView immediately to render the history and purge the "Loading" placeholder
	m.refreshView()

	return m
}

func (m ChatModel) Init() tea.Cmd {
	if m.mode == modes.Brownfield {
		return tea.Batch(textarea.Blink, m.scanWorkspaceCmd())
	}
	return textarea.Blink
}

func (m ChatModel) scanWorkspaceCmd() tea.Cmd {
	return func() tea.Msg {
		projectCtx, err := mcp.ScanWorkspace(m.cwd, 4)
		return scanCompleteMsg{projectCtx: projectCtx, err: err}
	}
}

// tickCmd fires every 100ms to animate the spinner
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// readNextChunk reads a single chunk asynchronously from the streaming channel.
func readNextChunk(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		val, ok := <-ch
		if !ok {
			return nextChunkMsg{done: true}
		}
		return nextChunkMsg{ch: ch, val: val}
	}
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	// --- Aggressive Terminal Escape Sequence Filter (Garbage Text Leak) ---
	val := m.textarea.Value()
	if strings.Contains(val, "rgb:") || strings.Contains(val, "]11;") || strings.Contains(val, "\x1b") || val == "\\" || strings.Contains(val, "\\") {
		// Purge any and all variations of OSC background/color query codes
		oscRegex := regexp.MustCompile(`(?i)\x1b\]\d*;?[^\\\x07]*(\x07|\x1b\\)?|\]11;rgb:[0-9a-f/]+|rgb:[0-9a-f/]+|\\x1b.*|\\`)
		cleanVal := oscRegex.ReplaceAllString(val, "")
		cleanVal = strings.ReplaceAll(cleanVal, "\\", "")
		cleanVal = strings.TrimSpace(cleanVal)
		m.textarea.SetValue(cleanVal)
	}

	// --- slash autocomplete detection ---
	currentVal := m.textarea.Value()
	if strings.HasPrefix(currentVal, "/") && !strings.Contains(currentVal, " ") {
		m.showSuggest = true
		m.suggestions = nil
		for _, cmd := range availableCommands {
			if strings.HasPrefix(cmd, currentVal) {
				m.suggestions = append(m.suggestions, cmd)
			}
		}
		if len(m.suggestions) == 0 {
			m.showSuggest = false
		} else if m.suggestIdx >= len(m.suggestions) {
			m.suggestIdx = 0
		}
	} else {
		m.showSuggest = false
		m.suggestIdx = 0
	}

	switch msg := msg.(type) {

	case scanCompleteMsg:
		m.isScanning = false
		if msg.err != nil {
			m.err = msg.err
			m.viewport.SetContent(fmt.Sprintf("\n⚠️ **Error scanning workspace**: %v\n\nPress Esc to exit.", msg.err))
			return m, nil
		}

		m.FileCount = len(msg.projectCtx.Files)

		// Build brownfield Socratic system prompt
		contextBlock := mcp.FormatContext(msg.projectCtx)
		gitCtx := mcp.BuildGitContext(m.cwd)
		if gitCtx != "" {
			contextBlock += "\n\n" + gitCtx
		}
		systemPrompt := modes.BuildBrownfieldPrompt(contextBlock)

		// Set the system message
		m.messages[0] = ai.Message{Role: "system", Content: systemPrompt}
		_ = m.session.OverwriteHistory(m.messages)

		// Update welcome message text
		welcomeMsg := `**Archon v1.0.0 • Software Architect**
Using model **` + strings.ToUpper(config.GetModel()) + `** • ` + fmt.Sprintf("**%d**", m.FileCount) + ` files loaded

` + modes.BrownfieldWelcome

		m.messages[1] = ai.Message{Role: "assistant", Content: welcomeMsg}

		m.refreshView()
		return m, nil

	case tickMsg:
		if m.isStreaming {
			m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
			m.refreshView()
			return m, tickCmd()
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			if m.isStreaming {
				m.isStreaming = false
				m.messages = append(m.messages, ai.Message{Role: "assistant", Content: "\n\n_Thinking cancelled by user._"})
				m.textarea.Prompt = InputPromptStyle.Render("❯ ")
				m.refreshView()
				return m, nil
			}
			m.textarea.SetValue("")
			return m, nil
		case tea.KeyCtrlL:
			cmd := m.executeSlashCommand("/clear")
			m.refreshView()
			return m, cmd

		case tea.KeyTab:
			// Auto-complete if suggestions menu is active
			if m.showSuggest && len(m.suggestions) > 0 {
				m.textarea.SetValue(m.suggestions[m.suggestIdx] + " ")
				m.textarea.CursorEnd()
				m.showSuggest = false
				m.suggestIdx = 0
				return m, nil
			}

		case tea.KeyUp:
			// Navigate autocomplete list if active
			if m.showSuggest && len(m.suggestions) > 0 {
				m.suggestIdx = (m.suggestIdx - 1 + len(m.suggestions)) % len(m.suggestions)
				return m, nil
			}

			// Fallback to command history
			if len(m.history) == 0 {
				return m, nil
			}
			// Save current input if we are at the end of history
			if m.historyIdx == len(m.history) {
				m.tempInput = m.textarea.Value()
			}
			if m.historyIdx > 0 {
				m.historyIdx--
				m.textarea.SetValue(m.history[m.historyIdx])
				m.textarea.CursorEnd()
			}
			return m, nil

		case tea.KeyDown:
			// Navigate autocomplete list if active
			if m.showSuggest && len(m.suggestions) > 0 {
				m.suggestIdx = (m.suggestIdx + 1) % len(m.suggestions)
				return m, nil
			}

			// Fallback to command history
			if len(m.history) == 0 {
				return m, nil
			}
			if m.historyIdx < len(m.history)-1 {
				m.historyIdx++
				m.textarea.SetValue(m.history[m.historyIdx])
				m.textarea.CursorEnd()
			} else if m.historyIdx == len(m.history)-1 {
				m.historyIdx++
				m.textarea.SetValue(m.tempInput)
				m.textarea.CursorEnd()
			}
			return m, nil

		case tea.KeyEnter:
			if msg.Alt {
				return m, tiCmd
			}

			if m.isStreaming {
				return m, nil
			}

			v := m.textarea.Value()
			trimmed := strings.TrimSpace(v)
			if trimmed == "" {
				return m, nil
			}

			// Add to history
			m.history = append(m.history, v)
			m.historyIdx = len(m.history)
			m.tempInput = ""

			m.textarea.Reset()
			m.showSuggest = false
			m.suggestIdx = 0

			// Show slash commands as visible user messages
			if strings.HasPrefix(trimmed, "/") {
				if trimmed == "/exit" || trimmed == "/quit" {
					return m, tea.Quit
				}

				// Add as visible user message so user sees what they typed
				visibleMsg := ai.Message{Role: "user", Content: trimmed}
				m.messages = append(m.messages, visibleMsg)

				cmd := m.executeSlashCommand(trimmed)

				if m.isStreaming {
					m.textarea.Prompt = InputPromptProcessingStyle.Render("❯ ")
					m.elapsedSecs = 0
					m.spinnerIdx = 0
					m.refreshView()
					return m, tea.Batch(m.startStream(), tickCmd(), cmd)
				}

				m.refreshView()
				return m, cmd
			}

			// Regular message
			m.textarea.Prompt = InputPromptProcessingStyle.Render("❯ ")

			resolvedContent := ai.ResolvePromptFiles(trimmed)
			userMsg := ai.Message{Role: "user", Content: resolvedContent}
			m.messages = append(m.messages, userMsg)
			_ = m.session.AppendMessage(userMsg)

			m.isStreaming = true
			m.streamBuffer = ""
			m.elapsedSecs = 0
			m.spinnerIdx = 0

			m.refreshView()

			return m, tea.Batch(m.startStream(), tickCmd())
		}

	case designDocsReadyMsg:
		m.isStreaming = false
		m.textarea.Prompt = InputPromptStyle.Render("❯ ")
		if msg.err != nil {
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: fmt.Sprintf("⚠️ **Error generating design documents**: %v", msg.err)})
		} else {
			m.designDocs = &msg.docs
			m.messages = append(m.messages, ai.Message{
				Role:    "assistant",
				Content: "_Design blueprints (ARCHITECTURE.md, DESIGN.md, SYSTEM.md) and AGENTS.md generated successfully! Use `/export` to write them to disk._",
			})
		}
		m.refreshView()
		return m, nil

	case nextChunkMsg:
		if msg.err != nil {
			m.isStreaming = false
			m.textarea.Prompt = InputPromptStyle.Render("❯ ")
			errText := fmt.Sprintf("⚠️ **Error**: %v\n\n_Try again or check your API key / network connection._", msg.err)
			m.messages = append(m.messages, ai.Message{Role: "assistant", Content: errText})
			m.refreshView()
			return m, nil
		}

		if msg.done {
			m.isStreaming = false
			m.textarea.Prompt = InputPromptStyle.Render("❯ ")

			fullContent := m.streamBuffer
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
		}

		m.streamBuffer += msg.val
		m.refreshView()
		return m, readNextChunk(msg.ch)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		contentW := msg.Width - 4
		if contentW < 40 {
			contentW = 40
		}

		m.textarea.SetWidth(contentW)

		m.mdRenderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(contentW-6),
		)

		// Set precise dynamic viewport sizing
		inputH := 1
		helpH := 1
		statusH := 2
		dividerH := 1
		padding := 1
		suggestH := 0
		if m.showSuggest && len(m.suggestions) > 0 {
			suggestH = len(m.suggestions) + 2
		}

		m.viewport.Height = msg.Height - inputH - helpH - statusH - dividerH - padding - suggestH
		if m.viewport.Height < 5 {
			m.viewport.Height = 5
		}
		m.viewport.Width = contentW

		m.refreshView()
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *ChatModel) refreshView() {
	var b strings.Builder

	// Prepend the styled banner at the top of the viewport content
	b.WriteString(BannerStyle.Render(BannerText) + "\n\n")

	for _, msg := range m.messages {
		if msg.Role == "system" {
			continue
		}
		if msg.Role == "user" {
			b.WriteString(UserLabel + "\n")
			b.WriteString(UserMsgStyle.Render(msg.Content) + "\n")
		} else if msg.Role == "assistant" {
			trimmed := strings.TrimSpace(msg.Content)

			isJSON := (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
				(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))

			if isJSON && strings.Contains(trimmed, ":") {
				b.WriteString(ArchonLabel + "\n")
				b.WriteString(JSONBoxStyle.Render(trimmed) + "\n")
			} else if strings.Contains(trimmed, "## SYSTEM QUALITIES") || strings.Contains(trimmed, "## CONCLUSION") {
				b.WriteString(ArchonLabel + "\n")
				rendered, _ := m.mdRenderer.Render(msg.Content)
				b.WriteString(ReviewBoxStyle.Render(rendered) + "\n")
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
		spinner := spinnerFrames[m.spinnerIdx]
		spinLine := lipgloss.NewStyle().
			Foreground(archonCyan).
			Bold(true).
			PaddingLeft(4).
			Render(spinner + " Archon is thinking...")
		b.WriteString(ArchonLabel + "\n")
		if m.streamBuffer == "" {
			b.WriteString(spinLine + "\n")
		} else {
			rendered, err := m.mdRenderer.Render(m.streamBuffer)
			if err != nil {
				b.WriteString(ArchonMsgStyle.Render(m.streamBuffer))
			} else {
				b.WriteString(ArchonMsgStyle.Render(rendered))
			}
			// Show inline spinner at the end while still streaming
			b.WriteString(spinLine + "\n")
		}
	}

	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

// renderSuggestions renders a floating list box of available commands matching current input.
func (m ChatModel) renderSuggestions() string {
	if !m.showSuggest || len(m.suggestions) == 0 {
		return ""
	}

	var items []string
	titleStyle := lipgloss.NewStyle().Foreground(archonCyan).Bold(true)
	items = append(items, titleStyle.Render("── COMMANDS ──"))

	for i, sug := range m.suggestions {
		if i == m.suggestIdx {
			items = append(items, lipgloss.NewStyle().Foreground(archonCyan).Background(archonDarkG).Bold(true).Render(" ❯ "+sug))
		} else {
			items = append(items, lipgloss.NewStyle().Foreground(archonWhite).Render("   "+sug))
		}
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(archonGray).
		Padding(0, 1).
		MarginBottom(0)

	return boxStyle.Render(strings.Join(items, "\n"))
}

// View assembles the full TUI matching the minimal UNIX aesthetic.
func (m ChatModel) View() string {
	// --- Status Bar ---
	tokens := ai.EstimateMessageTokens(m.messages)
	limit := ai.GetModelContextLimit(config.GetModel())

	tokenColor := archonGreen
	ratio := float64(tokens) / float64(limit)
	if ratio > 0.9 {
		tokenColor = archonRed
	} else if ratio > 0.8 {
		tokenColor = archonYellow
	}

	msgCount := len(m.messages) - 2
	if msgCount < 0 {
		msgCount = 0
	}

	modeStr := GreenfieldBadge
	if m.mode == modes.Brownfield {
		modeStr = BrownfieldBadge
	}

	leftStatus := lipgloss.JoinHorizontal(lipgloss.Top,
		StatusLabel("EXIT ^C"), StatusSep,
		StatusLabel("HELP "), StatusValue("/?"), StatusSep,
		modeStr,
	)

	rightStatus := lipgloss.JoinHorizontal(lipgloss.Top,
		StatusLabel("TOKENS: "),
		lipgloss.NewStyle().Foreground(tokenColor).Render(ai.FormatTokenCount(tokens)),
		StatusLabel("  /  "+ai.FormatTokenCount(limit)), StatusSep,
		StatusLabel("MSG: "), StatusValue(fmt.Sprintf("%d", msgCount)), StatusSep,
		StatusLabel("MODEL: "),
		lipgloss.NewStyle().Foreground(archonBlue).Render(strings.ToUpper(config.GetModel())), StatusSep,
		StatusLabel("SESSION: "), StatusValue(m.session.SessionID),
	)

	innerW := m.width - 4
	spaces := innerW - lipgloss.Width(leftStatus) - lipgloss.Width(rightStatus)
	if spaces < 1 {
		spaces = 1
	}
	statusLine := leftStatus + strings.Repeat(" ", spaces) + rightStatus
	renderedStatus := StatusBarOuter.Width(m.width - 2).Render(statusLine)

	// --- Input Box ---
	var inputBox string
	if m.isStreaming {
		inputBox = InputBoxProcessingStyle.Width(m.width - 4).Render(m.textarea.View())
	} else {
		inputBox = InputBoxStyle.Width(m.width - 4).Render(m.textarea.View())
	}

	divider := lipgloss.NewStyle().Foreground(archonGray).Render(strings.Repeat("─", m.width))
	helpLine := HelpTextHint.Render("↑/↓ history/commands  •  tab select  •  ctrl+l clear  •  esc exit")

	// If suggestions are showing, render them directly above the input box and below the divider
	var suggestMenu string
	suggestH := 0
	if m.showSuggest && len(m.suggestions) > 0 {
		suggestMenu = m.renderSuggestions() + "\n"
		suggestH = len(m.suggestions) + 2
	}

	// --- Dynamic Viewport Height Calculation ---
	inputH := lipgloss.Height(inputBox)
	helpH := lipgloss.Height(helpLine)
	statusH := lipgloss.Height(renderedStatus)
	dividerH := 1
	padding := 1

	m.viewport.Height = m.height - inputH - helpH - statusH - dividerH - padding - suggestH
	if m.viewport.Height < 5 {
		m.viewport.Height = 5
	}
	m.viewport.Width = m.width - 4

	// --- Assemble Full Layout (completely minimalist, no arcade headers) ---
	return lipgloss.JoinVertical(lipgloss.Left,
		m.viewport.View(),
		divider,
		suggestMenu,
		inputBox,
		helpLine,
		renderedStatus,
	)
}

func (m ChatModel) startStream() tea.Cmd {
	return func() tea.Msg {
		ch := make(chan string, 256)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		go func() {
			defer cancel()
			defer close(ch)

			// Execute ChatStream asynchronously and forward each chunk to the channel
			_, err := m.engine.ChatStream(ctx, m.messages, func(chunk string) {
				ch <- chunk
			})

			if err != nil {
				// Propagate error via a pseudo-chunk or standard nextChunkMsg structure
			}
		}()

		// Start reading from the channel immediately
		return nextChunkMsg{ch: ch}
	}
}
