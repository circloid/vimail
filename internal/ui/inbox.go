package ui

import (
	"context"
	"fmt"
	"veloci_mail/internal/email"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InboxModelImpl represents the inbox view state
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

// NewInboxModelImpl creates a new inbox model
func NewInboxModelImpl(ctx context.Context, emailClient *email.Client) *InboxModelImpl {
	return &InboxModelImpl{
		emailClient: emailClient,
		ctx:         ctx,
		messages:    []*email.Message{},
		selected:    0,
		loading:     false,
	}
}

// LoadMessagesMsg represents a message loading command
type LoadMessagesMsg struct {
	Messages []*email.Message
	Error    error
}

// Init initializes the inbox model
func (m *InboxModelImpl) Init() tea.Cmd {
	return m.LoadMessages()
}

// Update handles inbox-specific updates
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
		case "home":
			m.selected = 0
		case "end":
			if len(m.messages) > 0 {
				m.selected = len(m.messages) - 1
			}
		case "ctrl+r":
			return m, m.LoadMessages()
		}

	case LoadMessagesMsg:
		m.loading = false
		if msg.Error != nil {
			m.err = msg.Error
		} else {
			m.messages = msg.Messages
			m.err = nil
			// Keep selection within bounds
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

// View renders the inbox view
func (m *InboxModelImpl) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	if len(m.messages) == 0 {
		return m.renderEmpty()
	}

	return m.renderMessageList()
}

// renderLoading renders the loading state
func (m *InboxModelImpl) renderLoading() string {
	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render("📨 Loading messages...")

	return content
}

// renderError renders error state
func (m *InboxModelImpl) renderError() string {
	errorMsg := fmt.Sprintf("Failed to load messages: %v", m.err)
	content := ErrorStyle.Render(errorMsg)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// renderEmpty renders empty inbox state
func (m *InboxModelImpl) renderEmpty() string {
	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render("📭 No messages in inbox")

	return content
}

// renderMessageList renders the list of messages
func (m *InboxModelImpl) renderMessageList() string {
	var lines []string

	// Calculate visible range
	visibleHeight := m.height - 2 // Account for padding
	startIdx := 0
	endIdx := len(m.messages)

	// Ensure selected item is visible
	if m.selected >= startIdx+visibleHeight {
		startIdx = m.selected - visibleHeight + 1
	}
	if m.selected < startIdx {
		startIdx = m.selected
	}
	if startIdx+visibleHeight < endIdx {
		endIdx = startIdx + visibleHeight
	}

	// Render visible messages
	for i := startIdx; i < endIdx && i < len(m.messages); i++ {
		msg := m.messages[i]
		selected := i == m.selected

		line := FormatEmailListItem(
			msg.GetDisplayFrom(),
			msg.Subject,
			msg.FormatDate(),
			msg.IsUnread(),
			selected,
			m.width-4, // Account for padding
		)

		lines = append(lines, line)
	}

	// Add scroll indicators if needed
	if startIdx > 0 {
		lines = append([]string{
			lipgloss.NewStyle().
				Foreground(MutedColor).
				Align(lipgloss.Center).
				Width(m.width-4).
				Render("⬆ More messages above ⬆"),
		}, lines...)
	}

	if endIdx < len(m.messages) {
		lines = append(lines,
			lipgloss.NewStyle().
				Foreground(MutedColor).
				Align(lipgloss.Center).
				Width(m.width-4).
				Render("⬇ More messages below ⬇"),
		)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	return BorderStyle.
		Width(m.width-2).
		Height(m.height-1).
		Render(content)
}

// LoadMessages creates a command to load messages
func (m *InboxModelImpl) LoadMessages() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		messages, err := m.emailClient.GetInboxMessages(m.ctx, 50)
		return LoadMessagesMsg{
			Messages: messages,
			Error:    err,
		}
	}
}

// Refresh refreshes the message list
func (m *InboxModelImpl) Refresh() tea.Cmd {
	return m.LoadMessages()
}

// SetSize updates the size of the inbox view
func (m *InboxModelImpl) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetSelectedMessage returns the currently selected message
func (m *InboxModelImpl) GetSelectedMessage() *email.Message {
	if m.selected >= 0 && m.selected < len(m.messages) {
		return m.messages[m.selected]
	}
	return nil
}

// MessageCount returns the number of messages
func (m *InboxModelImpl) MessageCount() int {
	return len(m.messages)
}

// GetMessages returns all messages
func (m *InboxModelImpl) GetMessages() []*email.Message {
	return m.messages
}

// IsLoading returns whether the inbox is currently loading
func (m *InboxModelImpl) IsLoading() bool {
	return m.loading
}

// HasError returns whether there's an error
func (m *InboxModelImpl) HasError() bool {
	return m.err != nil
}

// GetError returns the current error
func (m *InboxModelImpl) GetError() error {
	return m.err
}
