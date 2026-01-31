package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors - Professional dark theme
var (
	// Primary colors
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	accentColor    = lipgloss.Color("#F59E0B") // Amber
	dangerColor    = lipgloss.Color("#EF4444") // Red

	// Neutral colors
	bgColor       = lipgloss.Color("#0F172A") // Dark blue
	surfaceColor  = lipgloss.Color("#1E293B") // Lighter dark
	borderColor   = lipgloss.Color("#334155") // Border
	textColor     = lipgloss.Color("#F8FAFC") // White
	textMutedColor = lipgloss.Color("#94A3B8") // Muted

	// Money colors
	incomeColor  = lipgloss.Color("#22C55E") // Green
	expenseColor = lipgloss.Color("#EF4444") // Red
)

// Base styles
var (
	// Container styles
	baseStyle = lipgloss.NewStyle().
			Background(bgColor).
			Foreground(textColor)

	// Header
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 2).
			Width(60)

	// Tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(primaryColor).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(textMutedColor).
				Padding(0, 2)

	// Card styles
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Width(56)

	cardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Money styles
	moneyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor)

	incomeStyle = lipgloss.NewStyle().
			Foreground(incomeColor)

	expenseStyle = lipgloss.NewStyle().
			Foreground(expenseColor)

	// Help bar
	helpStyle = lipgloss.NewStyle().
			Foreground(textMutedColor).
			Padding(0, 1)

	// Progress bar colors
	progressFullStyle  = lipgloss.NewStyle().Foreground(secondaryColor)
	progressEmptyStyle = lipgloss.NewStyle().Foreground(borderColor)
)

// renderProgressBar membuat visual progress bar.
func renderProgressBar(percent float64, width int) string {
	filled := int(percent / 100.0 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += progressFullStyle.Render("█")
		} else {
			bar += progressEmptyStyle.Render("░")
		}
	}

	return bar
}
