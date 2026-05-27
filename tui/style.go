package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base Colors
	archonCyan   = lipgloss.Color("#00f0ff")
	archonGray   = lipgloss.Color("#4b5563") // muted gray
	archonDarkG  = lipgloss.Color("#1f2937") // very dark gray
	archonGreen  = lipgloss.Color("#10b981")
	archonYellow = lipgloss.Color("#f59e0b")
	archonRed    = lipgloss.Color("#ef4444")
	archonWhite  = lipgloss.Color("#f3f4f6")
	archonBlue   = lipgloss.Color("#3b82f6")
	emerald      = lipgloss.Color("#10b981")

	// --- Badges ---
	BadgeStyle = lipgloss.NewStyle().
			Padding(0, 1)

	BrownfieldBadge = BadgeStyle.Copy().
			Foreground(archonYellow).
			Bold(true).
			Render("BROWNFIELD")

	GreenfieldBadge = BadgeStyle.Copy().
			Foreground(archonGreen).
			Bold(true).
			Render("GREENFIELD")

	// --- Status Bar (Subtle, single top border line) ---
	StatusBarOuter = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(archonGray).
			Padding(0, 1)

	StatusLabel = lipgloss.NewStyle().Foreground(archonGray).Render
	StatusValue = lipgloss.NewStyle().Foreground(archonWhite).Bold(true).Render
	StatusSep   = lipgloss.NewStyle().Foreground(archonGray).Render("  •  ")

	// --- Chat Messages ---
	MsgBarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(archonGray).
			PaddingLeft(2)

	UserLabel = lipgloss.NewStyle().
			Foreground(archonGreen).
			Bold(true).
			PaddingLeft(1).
			MarginTop(1).
			Render("you >")

	UserMsgStyle = lipgloss.NewStyle().
			Foreground(archonWhite).
			PaddingLeft(3).
			MarginBottom(1)

	ArchonLabel = lipgloss.NewStyle().
			Foreground(archonCyan).
			Bold(true).
			PaddingLeft(1).
			MarginTop(1).
			Render("archon >")

	ArchonMsgStyle = lipgloss.NewStyle().
			PaddingLeft(3).
			MarginBottom(1)

	// --- Specialized Content Boxes (Minimalist left border only, no heavy boxes) ---
	JSONBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(emerald).
			Padding(0, 2).
			MarginLeft(3).
			MarginBottom(1)

	ReviewBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(archonBlue).
			Padding(0, 2).
			MarginLeft(3).
			MarginBottom(1)

	CommandResultStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(archonGray).
				Foreground(lipgloss.Color("#cbd5e1")).
				PaddingLeft(2).
				MarginBottom(1)

	// --- Input Box (Fully Borderless, minimal top padding) ---
	InputBoxStyle = lipgloss.NewStyle().
			Padding(0, 0).
			MarginTop(0).
			MarginBottom(0)

	InputBoxProcessingStyle = InputBoxStyle.Copy()

	InputPromptStyle = lipgloss.NewStyle().
				Foreground(archonCyan).
				Bold(true)

	InputPromptProcessingStyle = lipgloss.NewStyle().
					Foreground(archonGray).
					Bold(true)

	// Dim Help Text
	HelpTextHint = lipgloss.NewStyle().
			Foreground(archonGray).
			Italic(true).
			PaddingLeft(1)
)
