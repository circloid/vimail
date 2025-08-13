package ui

import (
	"strings"
	"veloci_mail/internal/email"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ReaderModelImpl represents the email reading view state
type ReaderModelImpl struct {
	message    *email.Message
	width      int
	height     int
	scrollY    int
	bodyLines  []string
}

// NewReaderModelImpl creates a new reader model
func NewReaderModelImpl() *ReaderModelImpl {
	return &ReaderModelImpl{
		scrollY: 0,
	}
}

// Init initializes the reader model
func (m *ReaderModelImpl) Init() tea.Cmd {
	return nil
}

// Update handles reader-specific updates
func (m *ReaderModelImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.message == nil {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.scrollY > 0 {
				m.scrollY--
			}
		case "down", "j":
			maxScroll := len(m.bodyLines) - (m.height - 8) // Account for header space
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollY < maxScroll {
				m.scrollY++
			}
		case "home":
			m.scrollY = 0
		case "end":
			maxScroll := len(m.bodyLines) - (m.height - 8)
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.scrollY = maxScroll
		case "pgup":
			pageSize := m.height - 8
			m.scrollY -= pageSize
			if m.scrollY < 0 {
				m.scrollY = 0
			}
		case "pgdn":
			pageSize := m.height - 8
			maxScroll := len(m.bodyLines) - pageSize
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.scrollY += pageSize
			if m.scrollY > maxScroll {
				m.scrollY = maxScroll
			}
		}
	}

	return m, nil
}

// View renders the email reading view
func (m *ReaderModelImpl) View() string {
	if m.message == nil {
		return m.renderEmpty()
	}

	header := m.renderMessageHeader()
	body := m.renderMessageBody()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		body,
	)
}

// renderEmpty renders empty state
func (m *ReaderModelImpl) renderEmpty() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render("No message selected")
}

// renderMessageHeader renders the email header information
func (m *ReaderModelImpl) renderMessageHeader() string {
	var headerLines []string

	// Subject
	subject := m.message.Subject
	if subject == "" {
		subject = "(No Subject)"
	}
	headerLines = append(headerLines, EmailHeaderStyle.Render("Subject: "+subject))

	// From
	from := m.message.GetDisplayFrom()
	headerLines = append(headerLines, EmailMetaStyle.Render("From: "+from))

	// To
	if m.message.To != "" {
		headerLines = append(headerLines, EmailMetaStyle.Render("To: "+m.message.To))
	}

	// Date
	date := m.message.Date.Format("Mon, Jan 2, 2006 at 3:04 PM")
	headerLines = append(headerLines, EmailMetaStyle.Render("Date: "+date))

	// Add separator
	headerLines = append(headerLines, "")
	headerLines = append(headerLines, strings.Repeat("─", m.width-4))
	headerLines = append(headerLines, "")

	header := lipgloss.JoinVertical(lipgloss.Left, headerLines...)

	return BorderStyle.
		Width(m.width-2).
		Render(header)
}

// renderMessageBody renders the scrollable message body
func (m *ReaderModelImpl) renderMessageBody() string {
	if len(m.bodyLines) == 0 {
		return BorderStyle.
			Width(m.width-2).
			Height(m.height-8). // Account for header space
			Render(EmailBodyStyle.Render("(Empty message)"))
	}

	// Calculate visible area
	bodyHeight := m.height - 10 // Account for header, borders, and padding
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	startLine := m.scrollY
	endLine := startLine + bodyHeight
	if endLine > len(m.bodyLines) {
		endLine = len(m.bodyLines)
	}

	// Get visible lines
	var visibleLines []string
	for i := startLine; i < endLine; i++ {
		if i < len(m.bodyLines) {
			line := m.bodyLines[i]
			// Wrap long lines to fit width
			wrappedLine := email.WrapText(line, m.width-6) // Account for borders and padding
			visibleLines = append(visibleLines, wrappedLine)
		}
	}

	// Add scroll indicators
	var indicators []string
	if m.scrollY > 0 {
		indicators = append(indicators,
			lipgloss.NewStyle().
				Foreground(MutedColor).
				Align(lipgloss.Center).
				Width(m.width-6).
				Render("⬆ Scroll up for more ⬆"))
	}

	indicators = append(indicators, visibleLines...)

	if endLine < len(m.bodyLines) {
		indicators = append(indicators,
			lipgloss.NewStyle().
				Foreground(MutedColor).
				Align(lipgloss.Center).
				Width(m.width-6).
				Render("⬇ Scroll down for more ⬇"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, indicators...)

	return BorderStyle.
		Width(m.width-2).
		Height(bodyHeight+2). // Add border space
		Render(EmailBodyStyle.Render(content))
}

// SetMessage sets the message to display and prepares it for rendering
func (m *ReaderModelImpl) SetMessage(message *email.Message) {
	m.message = message
	m.scrollY = 0

	if message != nil {
		// Prepare body lines for scrolling
		m.bodyLines = strings.Split(message.Body, "\n")

		// Remove empty lines at the end
		for len(m.bodyLines) > 0 && strings.TrimSpace(m.bodyLines[len(m.bodyLines)-1]) == "" {
			m.bodyLines = m.bodyLines[:len(m.bodyLines)-1]
		}

		// Ensure we have at least one line
		if len(m.bodyLines) == 0 {
			m.bodyLines = []string{"(Empty message)"}
		}
	} else {
		m.bodyLines = []string{}
	}
}

// SetSize updates the size of the reader view
func (m *ReaderModelImpl) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Adjust scroll position if necessary
	if m.message != nil {
		maxScroll := len(m.bodyLines) - (height - 8)
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.scrollY > maxScroll {
			m.scrollY = maxScroll
		}
	}
}

// GetMessage returns the current message
func (m *ReaderModelImpl) GetMessage() *email.Message {
	return m.message
}

// CanScrollUp returns whether scrolling up is possible
func (m *ReaderModelImpl) CanScrollUp() bool {
	return m.scrollY > 0
}

// CanScrollDown returns whether scrolling down is possible
func (m *ReaderModelImpl) CanScrollDown() bool {
	maxScroll := len(m.bodyLines) - (m.height - 8)
	if maxScroll < 0 {
		maxScroll = 0
	}
	return m.scrollY < maxScroll
}
