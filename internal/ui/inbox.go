// internal/ui/inbox.go - Minimal Zen Inbox
package ui

import (
	"context"
	"vimail/internal/email"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InboxModelImpl struct {
	emailClient *email.Client
	ctx         context.Context
	messages    []*email.Message
	selected    int
	width       int
	height      int
	loading     bool
	err         error
}

func NewInboxModelImpl(ctx context.Context, emailClient *email.Client) *InboxModelImpl {
	return &InboxModelImpl{
		emailClient: emailClient,
		ctx:         ctx,
		messages:    []*email.Message{},
		selected:    0,
		loading:     false,
	}
}

type LoadMessagesMsg struct {
	Messages []*email.Message
	Error    error
}

func (m *InboxModelImpl) Init() tea.Cmd {
	return m.LoadMessages()
}

func (m *InboxModelImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.messages)-1 {
				m.selected++
			}
		}

	case LoadMessagesMsg:
		m.loading = false
		if msg.Error != nil {
			m.err = msg.Error
		} else {
			m.messages = msg.Messages
			m.err = nil
			if m.selected >= len(m.messages) {
				m.selected = len(m.messages) - 1
			}
			if m.selected < 0 {
				m.selected = 0
			}
		}
	}

	return m, nil
}

func (m *InboxModelImpl) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Foreground(Gray).
			Padding(5, 2).
			Render("Loading...")
	}

	if len(m.messages) == 0 {
		return lipgloss.NewStyle().
			Foreground(Gray).
			Padding(5, 2).
			Render("No messages")
	}

	var lines []string
	for i, msg := range m.messages {
		selected := i == m.selected
		line := FormatEmailLine(msg.GetDisplayFrom(), msg.Subject, selected)
		lines = append(lines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return SimpleBorderStyle.Render(content)
}

func (m *InboxModelImpl) LoadMessages() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		messages, err := m.emailClient.GetInboxMessages(m.ctx, 20)
		return LoadMessagesMsg{
			Messages: messages,
			Error:    err,
		}
	}
}

func (m *InboxModelImpl) Refresh() tea.Cmd {
	return m.LoadMessages()
}

func (m *InboxModelImpl) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *InboxModelImpl) GetSelectedMessage() *email.Message {
	if m.selected >= 0 && m.selected < len(m.messages) {
		return m.messages[m.selected]
	}
	return nil
}

func (m *InboxModelImpl) MessageCount() int {
	return len(m.messages)
}

func (m *InboxModelImpl) GetMessages() []*email.Message {
	return m.messages
}

func (m *InboxModelImpl) IsLoading() bool {
	return m.loading
}

func (m *InboxModelImpl) HasError() bool {
	return m.err != nil
}

func (m *InboxModelImpl) GetError() error {
	return m.err
}
