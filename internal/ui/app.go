// internal/ui/app.go - Minimal Zen App
package ui

import (
	"context"
	"vimail/internal/config"
	"vimail/internal/email"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InboxModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	SetSize(width, height int)
	GetSelectedMessage() *email.Message
	MessageCount() int
	Refresh() tea.Cmd
}

type ReaderModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	SetSize(width, height int)
	SetMessage(message *email.Message)
	GetMessage() *email.Message
}

type ComposerModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	SetSize(width, height int)
	IsSent() bool
	IsCancelled() bool
}

type ViewMode int

const (
	InboxView ViewMode = iota
	ReaderView
	ComposerView
)

type Model struct {
	emailClient  *email.Client
	config       *config.Config
	ctx          context.Context
	viewMode     ViewMode
	width        int
	height       int
	ready        bool
	inbox        *InboxModelImpl
	reader       *ReaderModelImpl
	composer     *ComposerModelImpl
	previousView ViewMode
}

func NewModel(ctx context.Context, emailClient *email.Client, cfg *config.Config) Model {
	return Model{
		emailClient: emailClient,
		config:      cfg,
		ctx:         ctx,
		viewMode:    InboxView,
		inbox:       NewInboxModelImpl(ctx, emailClient),
		reader:      NewReaderModelImpl(),
		composer:    NewComposerModelImpl(emailClient.GetUserEmail(), emailClient),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.inbox.Init(),
		tea.EnterAltScreen,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.inbox.SetSize(msg.Width, msg.Height-3)
		m.reader.SetSize(msg.Width, msg.Height-3)
		m.composer.SetSize(msg.Width, msg.Height-3)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			if m.viewMode == ReaderView {
				m.viewMode = InboxView
			} else if m.viewMode == ComposerView {
				m.viewMode = m.previousView
			}

		case "c":
			if m.viewMode == InboxView {
				m.previousView = InboxView
				m.viewMode = ComposerView
				m.composer = NewComposerModelImpl(m.emailClient.GetUserEmail(), m.emailClient)
				return m, m.composer.Init()
			}

		case "enter":
			if m.viewMode == InboxView {
				if selectedMsg := m.inbox.GetSelectedMessage(); selectedMsg != nil {
					m.previousView = InboxView
					m.viewMode = ReaderView
					m.reader.SetMessage(selectedMsg)
					return m, nil
				}
			}
		}
	}

	// Update current view
	switch m.viewMode {
	case InboxView:
		updatedInbox, inboxCmd := m.inbox.Update(msg)
		if updatedModel, ok := updatedInbox.(*InboxModelImpl); ok {
			m.inbox = updatedModel
		}
		cmds = append(cmds, inboxCmd)

	case ReaderView:
		updatedReader, readerCmd := m.reader.Update(msg)
		if updatedModel, ok := updatedReader.(*ReaderModelImpl); ok {
			m.reader = updatedModel
		}
		cmds = append(cmds, readerCmd)

	case ComposerView:
		updatedComposer, composerCmd := m.composer.Update(msg)
		if updatedModel, ok := updatedComposer.(*ComposerModelImpl); ok {
			m.composer = updatedModel
		}
		cmds = append(cmds, composerCmd)

		// Handle composer completion
		if m.composer.IsSent() || m.composer.IsCancelled() {
			m.viewMode = m.previousView
			if m.composer.IsSent() {
				cmds = append(cmds, m.inbox.Refresh())
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Minimal header
	var header string
	switch m.viewMode {
	case InboxView:
		header = "Inbox"
	case ReaderView:
		header = "Reading"
	case ComposerView:
		header = "Compose"
	}

	headerBar := HeaderStyle.Width(m.width).Render(header)

	// Content
	var content string
	switch m.viewMode {
	case InboxView:
		content = m.inbox.View()
	case ReaderView:
		content = m.reader.View()
	case ComposerView:
		content = m.composer.View()
	}

	// Simple help
	help := lipgloss.NewStyle().
		Foreground(Gray).
		Align(lipgloss.Center).
		Width(m.width).
		Render("q: quit | ↑↓: navigate | enter: read | c: compose | esc: back")

	return lipgloss.JoinVertical(
		lipgloss.Top,
		headerBar,
		content,
		help,
	)
}

type ErrorMsg struct {
	Error error
}

type StatusMsg struct {
	Message string
}

func NewErrorMsg(err error) ErrorMsg {
	return ErrorMsg{Error: err}
}

func NewStatusMsg(message string) StatusMsg {
	return StatusMsg{Message: message}
}
