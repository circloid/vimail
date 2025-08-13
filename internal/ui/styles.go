package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	PrimaryColor   = lipgloss.Color("#4A90E2")
	SecondaryColor = lipgloss.Color("#7ED321")
	AccentColor    = lipgloss.Color("#F5A623")
	ErrorColor     = lipgloss.Color("#D0021B")
	WarningColor   = lipgloss.Color("#F8E71C")
	SuccessColor   = lipgloss.Color("#7ED321")
	TextColor      = lipgloss.Color("#333333")
	MutedColor     = lipgloss.Color("#888888")
	BackgroundColor = lipgloss.Color("#FFFFFF")
	BorderColor    = lipgloss.Color("#CCCCCC")
)

// Base styles
var (
	BaseStyle = lipgloss.NewStyle().
		Foreground(TextColor).
		Background(BackgroundColor)

	TitleStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(MutedColor).
		Italic(true)

	SelectedStyle = lipgloss.NewStyle().
		Foreground(BackgroundColor).
		Background(PrimaryColor).
		Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
		Foreground(TextColor)

	BorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)

	WarningStyle = lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true)
)

// List styles
var (
	ListItemStyle = lipgloss.NewStyle().
		Padding(0, 2)

	SelectedListItemStyle = lipgloss.NewStyle().
		Foreground(BackgroundColor).
		Background(PrimaryColor).
		Padding(0, 2)

	UnreadItemStyle = lipgloss.NewStyle().
		Foreground(TextColor).
		Bold(true).
		Padding(0, 2)

	ReadItemStyle = lipgloss.NewStyle().
		Foreground(MutedColor).
		Padding(0, 2)
)

// Header styles
var (
	HeaderStyle = lipgloss.NewStyle().
		Foreground(BackgroundColor).
		Background(PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Width(100)

	StatusStyle = lipgloss.NewStyle().
		Foreground(MutedColor).
		Align(lipgloss.Right).
		Padding(0, 2)
)

// Input styles
var (
	InputStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(BorderColor).
		Padding(0, 1).
		Width(50)

	FocusedInputStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(PrimaryColor).
		Padding(0, 1).
		Width(50)

	InputLabelStyle = lipgloss.NewStyle().
		Foreground(TextColor).
		Bold(true).
		Width(12).
		Align(lipgloss.Right)

	InputPromptStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)
)

// Content styles
var (
	ContentStyle = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor)

	EmailHeaderStyle = lipgloss.NewStyle().
		Foreground(TextColor).
		Bold(true).
		Padding(0, 0, 1, 0)

	EmailMetaStyle = lipgloss.NewStyle().
		Foreground(MutedColor).
		Italic(true)

	EmailBodyStyle = lipgloss.NewStyle().
		Foreground(TextColor).
		Padding(1, 0)
)

// Help styles
var (
	HelpStyle = lipgloss.NewStyle().
		Foreground(MutedColor).
		Align(lipgloss.Center).
		Padding(1, 0)

	KeyStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	DescStyle = lipgloss.NewStyle().
		Foreground(MutedColor)
)

// Layout helpers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FormatKeyHelp formats key binding help text
func FormatKeyHelp(key, desc string) string {
	return KeyStyle.Render(key) + " " + DescStyle.Render(desc)
}

// TruncateString truncates a string to the specified length
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length < 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}

// PadRight pads a string to the specified width
func PadRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + lipgloss.NewStyle().Width(width-len(s)).Render("")
}

// CenterText centers text within the specified width
func CenterText(text string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(text)
}

// FormatEmailListItem formats an email for display in the inbox list
func FormatEmailListItem(from, subject, date string, unread bool, selected bool, width int) string {
	// Calculate available space for each field
	dateWidth := 8
	fromWidth := Max(15, (width-dateWidth-10)/3)
	subjectWidth := width - fromWidth - dateWidth - 6

	// Truncate fields to fit
	fromDisplay := TruncateString(from, fromWidth)
	subjectDisplay := TruncateString(subject, subjectWidth)
	dateDisplay := TruncateString(date, dateWidth)

	// Format the line
	line := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(fromWidth).Render(fromDisplay),
		" ",
		lipgloss.NewStyle().Width(subjectWidth).Render(subjectDisplay),
		" ",
		lipgloss.NewStyle().Width(dateWidth).Align(lipgloss.Right).Render(dateDisplay),
	)

	// Apply styling based on state
	var style lipgloss.Style
	if selected {
		style = SelectedListItemStyle
	} else if unread {
		style = UnreadItemStyle
	} else {
		style = ReadItemStyle
	}

	return style.Width(width).Render(line)
}

// RenderHeader renders the application header
func RenderHeader(title, status string, width int) string {
	titlePart := TitleStyle.Render(title)
	statusPart := StatusStyle.Render(status)
	
	// Calculate spacing
	usedWidth := len(title) + len(status) + 4 // padding
	spacer := Max(0, width-usedWidth)
	
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		titlePart,
		lipgloss.NewStyle().Width(spacer).Render(""),
		statusPart,
	)
}

// RenderHelp renders help text at the bottom
func RenderHelp(helps []string, width int) string {
	helpText := lipgloss.JoinHorizontal(lipgloss.Top, helps...)
	return HelpStyle.Width(width).Render(helpText)
}
