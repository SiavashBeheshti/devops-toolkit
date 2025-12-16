package output

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

// Theme colors for consistent styling
var (
	// Primary colors
	PrimaryColor   = lipgloss.Color("#7C3AED") // Purple
	SecondaryColor = lipgloss.Color("#06B6D4") // Cyan
	AccentColor    = lipgloss.Color("#F59E0B") // Amber

	// Status colors
	SuccessColor = lipgloss.Color("#10B981") // Green
	WarningColor = lipgloss.Color("#F59E0B") // Amber
	ErrorColor   = lipgloss.Color("#EF4444") // Red
	InfoColor    = lipgloss.Color("#3B82F6") // Blue

	// Neutral colors
	MutedColor      = lipgloss.Color("#6B7280") // Gray
	BackgroundColor = lipgloss.Color("#1F2937") // Dark gray
	BorderColor     = lipgloss.Color("#374151") // Medium gray
)

// Styled text helpers
var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Italic(true)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(InfoColor)

	MutedStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2)

	HeaderBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 2).
			Bold(true)

	// Badge styles
	SuccessBadge = lipgloss.NewStyle().
			Background(SuccessColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	WarningBadge = lipgloss.NewStyle().
			Background(WarningColor).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1).
			Bold(true)

	ErrorBadge = lipgloss.NewStyle().
			Background(ErrorColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	InfoBadge = lipgloss.NewStyle().
			Background(InfoColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)
)

// Color print helpers using fatih/color for table compatibility
var (
	ColorSuccess = color.New(color.FgGreen, color.Bold)
	ColorWarning = color.New(color.FgYellow, color.Bold)
	ColorError   = color.New(color.FgRed, color.Bold)
	ColorInfo    = color.New(color.FgBlue)
	ColorMuted   = color.New(color.FgHiBlack)
	ColorPrimary = color.New(color.FgMagenta, color.Bold)
	ColorCyan    = color.New(color.FgCyan)
)

// Status icons
const (
	IconSuccess  = "✓"
	IconWarning  = "⚠"
	IconError    = "✗"
	IconInfo     = "ℹ"
	IconRunning  = "●"
	IconPending  = "○"
	IconArrow    = "→"
	IconBullet   = "•"
	IconCheck    = "✔"
	IconCross    = "✘"
	IconStar     = "★"
	IconDot      = "·"
	IconPipe     = "│"
	IconCorner   = "└"
	IconTee      = "├"
	IconDash     = "─"
	IconDoubleDash = "═"
)

