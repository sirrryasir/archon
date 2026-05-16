package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base Colors
	archonCyan   = lipgloss.Color("#00f0ff")
	archonGray   = lipgloss.Color("#6b7280")
	archonDarkG  = lipgloss.Color("#4b5563")
	archonGreen  = lipgloss.Color("#10b981")
	archonYellow = lipgloss.Color("#f59e0b")
	archonRed    = lipgloss.Color("#ef4444")
	archonWhite  = lipgloss.Color("#ffffff")
	archonBlue   = lipgloss.Color("#3b82f6")
	emerald      = lipgloss.Color("#10b981")

	// Component Styles
	BannerStyle = lipgloss.NewStyle().
			Foreground(archonCyan).
			Bold(true)

	BannerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#d97706")).
			Padding(0, 2).
			Align(lipgloss.Center)

	BadgeStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 1)

	BrownfieldBadge = BadgeStyle.Copy().
			BorderForeground(archonYellow).
			Foreground(archonYellow).
			Bold(true).
			Render("BROWNFIELD")

	GreenfieldBadge = BadgeStyle.Copy().
			BorderForeground(archonGreen).
			Foreground(archonGreen).
			Bold(true).
			Render("GREENFIELD")

	ModelBadge = BadgeStyle.Copy().
			BorderForeground(archonGray).
			Foreground(archonGray)

	FilesBadge = BadgeStyle.Copy().
			BorderForeground(archonGray).
			Foreground(archonGray)

	// Status Bar Styles
	StatusBarContainer = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), true, false, false, false).
				BorderForeground(archonGray).
				Padding(0, 1).
				MarginTop(1).
				Width(140)

	StatusLabel = lipgloss.NewStyle().Foreground(archonGray).Render
	StatusValue = lipgloss.NewStyle().Foreground(archonWhite).Bold(true).Render
	StatusSep   = lipgloss.NewStyle().Foreground(archonDarkG).Render(" | ")

	// Chat Messages
	UserLabel = lipgloss.NewStyle().
			Foreground(archonGreen).
			Bold(true).
			PaddingLeft(2).
			MarginTop(1).
			Render("YOU")

	UserMsgStyle = lipgloss.NewStyle().
			Foreground(archonWhite).
			PaddingLeft(4).
			MarginBottom(1)

	ArchonLabel = lipgloss.NewStyle().
			Foreground(archonCyan).
			Bold(true).
			PaddingLeft(2).
			MarginTop(1).
			Render("ARCHON")

	ArchonMsgStyle = lipgloss.NewStyle().
			PaddingLeft(4).
			MarginBottom(1)

	JSONBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(emerald).
			Padding(1, 2).
			MarginLeft(4).
			MarginBottom(1)

	ReviewBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(archonBlue).
			Padding(1, 2).
			MarginLeft(4).
			MarginBottom(1)

	CommandResultStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(archonDarkG).
				Foreground(lipgloss.Color("#cbd5e1")).
				PaddingLeft(4).
				MarginBottom(1)

	// Layout Styles
	SidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(archonDarkG).
			Padding(0, 1).
			MarginLeft(2)

	HeaderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(archonDarkG).
			MarginBottom(1).
			PaddingBottom(1)

	// Input Box
	InputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(archonCyan).
			Padding(0, 1)

	InputBoxProcessingStyle = InputBoxStyle.Copy().
				BorderForeground(archonGray)

	InputPromptStyle = lipgloss.NewStyle().
				Foreground(archonCyan).
				Bold(true)

	InputPromptProcessingStyle = lipgloss.NewStyle().
	                                Foreground(archonGray).
	                                Bold(true)

	// Socratic Specialized Styles
	SocraticQuestStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(archonCyan).
	Padding(1, 2).
	MarginLeft(4).
	MarginBottom(1).
	Italic(true).
	Foreground(archonWhite)

	GoldenAdviceStyle = lipgloss.NewStyle().
	Border(lipgloss.DoubleBorder()).
	BorderForeground(archonYellow).
	Padding(1, 2).
	MarginLeft(4).
	MarginBottom(1).
	Bold(true)
	)

