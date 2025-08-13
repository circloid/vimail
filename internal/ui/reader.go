// internal/ui/reader.go - Minimal Zen Email Reader
package ui

import (
	"strings"
	"vimail/internal/email"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ReaderModelImpl struct {
	message   *email.Message
	width     int
	height    int
	scrollY   int
	bodyLines []string
}

func NewReaderModelImpl() *ReaderModelImpl {
	return &ReaderModelImpl{
		scrollY: 0,
	}
}

func (m *ReaderModelImpl) Init() tea.Cmd {
	return nil
}

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
			maxScroll := len(m.bodyLines) - (m.height - 6)
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollY < maxScroll {
				m.scrollY++
			}
		}
	}

	return m, nil
}

func (m *ReaderModelImpl) View() string {
	if m.message == nil {
		return lipgloss.NewStyle().
			Foreground(Gray).
			Padding(5, 2).
			Render("No message")
	}

	// Simple header - just subject in white
	header := lipgloss.NewStyle().
		Foreground(White).
		Bold(true).
		Padding(1, 2).
		Render(m.message.Subject)

	// Clean email body - white text, no decorations
	bodyHeight := m.height - 6
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	startLine := m.scrollY
	endLine := startLine + bodyHeight
	if endLine > len(m.bodyLines) {
		endLine = len(m.bodyLines)
	}

	var visibleLines []string
	for i := startLine; i < endLine; i++ {
		if i < len(m.bodyLines) {
			visibleLines = append(visibleLines, m.bodyLines[i])
		}
	}

	bodyText := strings.Join(visibleLines, "\n")
	body := EmailTextStyle.Render(bodyText)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		"",
		body,
	)
}

func (m *ReaderModelImpl) SetMessage(message *email.Message) {
	m.message = message
	m.scrollY = 0

	if message != nil {
		m.bodyLines = strings.Split(message.Body, "\n")
		// Clean up empty lines at the end
		for len(m.bodyLines) > 0 && strings.TrimSpace(m.bodyLines[len(m.bodyLines)-1]) == "" {
			m.bodyLines = m.bodyLines[:len(m.bodyLines)-1]
		}
		if len(m.bodyLines) == 0 {
			m.bodyLines = []string{"(Empty message)"}
		}
	} else {
		m.bodyLines = []string{}
	}
}

func (m *ReaderModelImpl) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ReaderModelImpl) GetMessage() *email.Message {
	return m.message
}
