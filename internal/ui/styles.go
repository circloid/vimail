// internal/ui/styles.go - Minimal Zen Design
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Minimal color palette
var (
	// Simple colors
	White    = lipgloss.Color("#ffffff")
	Gray     = lipgloss.Color("#888888")
	DarkGray = lipgloss.Color("#444444")
	Blue     = lipgloss.Color("#5555ff")
	Black    = lipgloss.Color("#000000")
)

// Minimal styles
var (
	// Clean email list item
	EmailItemStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Padding(0, 2)

	// Selected email - simple highlight
	SelectedEmailStyle = lipgloss.NewStyle().
				Foreground(White).
				Background(Blue).
				Padding(0, 2)

	// Email text - clean white text
	EmailTextStyle = lipgloss.NewStyle().
			Foreground(White).
			Padding(1, 2)

	// Simple border
	SimpleBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(DarkGray).
				Padding(1)

	// Clean header
	HeaderStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(DarkGray).
			Padding(0, 2)
)

// Simple utility functions
func FormatEmailLine(from, subject string, selected bool) string {
	text := from + " - " + subject
	if selected {
		return SelectedEmailStyle.Render(text)
	}
	return EmailItemStyle.Render(text)
}

func RenderEmailText(text string) string {
	return EmailTextStyle.Render(text)
}
