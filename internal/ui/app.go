package ui

import (
	"context"
	"fmt"
	"veloci_mail/internal/config"
	"veloci_mail/internal/email"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Forward declarations for models defined in other files
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
	SetEmailClient(client *email.Client)
	IsSent() bool
	IsCancelled() bool
}

// ViewMode represents different application views
type ViewMode int

const (
	InboxView ViewMode = iota
	ReaderView
	ComposerView
)

// Model represents the main application state
type Model struct {
	// Core components
	emailClient *email.Client
	config      *config.Config
	ctx         context.Context

	// UI state
	viewMode     ViewMode
	width        int
	height       int
	ready        bool
	err          error

	// Sub-models (using concrete types from other files)
	inbox    *InboxModelImpl
	reader   *ReaderModelImpl
	composer *ComposerModelImpl

	// Navigation
	previousView ViewMode
}

// NewModel creates a new application model
func NewModel(ctx context.Context, emailClient *email.Client, cfg *config.Config) Model {
	return Model{
		emailClient: emailClient,
		config:      cfg,
		ctx:         ctx,
		viewMode:    InboxView,
		inbox:       NewInboxModelImpl(ctx, emailClient),
		reader:      NewReaderModelImpl(),
		composer:    NewComposerModelImpl(emailClient.GetUserEmail()),
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.inbox.Init(),
		tea.EnterAltScreen,
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update sub-models with new dimensions
		m.inbox.SetSize(msg.Width, msg.Height-4) // Account for header and help
		m.reader.SetSize(msg.Width, msg.Height-4)
		m.composer.SetSize(msg.Width, msg.Height-4)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.viewMode == InboxView {
				return m, tea.Quit
			} else {
				// Return to previous view or inbox
				m.viewMode = m.previousView
				if m.viewMode == 0 {
					m.viewMode = InboxView
				}
				return m, nil
			}

		case "c":
			if m.viewMode == InboxView {
				m.previousView = InboxView
				m.viewMode = ComposerView
				m.composer = NewComposerModelImpl(m.emailClient.GetUserEmail())
				return m, m.composer.Init()
			}

		case "r":
			if m.viewMode == InboxView {
				// Get selected message for reply
				if selectedMsg := m.inbox.GetSelectedMessage(); selectedMsg != nil {
					m.previousView = InboxView
					m.viewMode = ComposerView
					m.composer = NewReplyComposerModelImpl(m.emailClient.GetUserEmail(), selectedMsg)
					return m, m.composer.Init()
				}
			}

		case "enter":
			if m.viewMode == InboxView {
				// Open selected message in reader
				if selectedMsg := m.inbox.GetSelectedMessage(); selectedMsg != nil {
					m.previousView = InboxView
					m.viewMode = ReaderView
					m.reader.SetMessage(selectedMsg)
					return m, nil
				}
			}
		}

	case ErrorMsg:
		m.err = msg.Error
		return m, nil

	case StatusMsg:
		// Handle status updates
		return m, nil
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

		// Handle reader-specific navigation
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc":
				m.viewMode = m.previousView
			}
		}

	case ComposerView:
		updatedComposer, composerCmd := m.composer.Update(msg)
		if updatedModel, ok := updatedComposer.(*ComposerModelImpl); ok {
			m.composer = updatedModel
		}
		cmds = append(cmds, composerCmd)

		// Handle composer completion
		if m.composer.IsSent() || m.composer.IsCancelled() {
			m.viewMode = m.previousView
			// Refresh inbox if we sent a message
			if m.composer.IsSent() {
				cmds = append(cmds, m.inbox.Refresh())
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the current view
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.err != nil {
		return m.renderError()
	}

	header := m.renderHeader()
	content := m.renderContent()
	help := m.renderHelp()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		content,
		help,
	)
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	var title string
	var status string

	switch m.viewMode {
	case InboxView:
		title = "📧 Terminal Email Client - Inbox"
		status = fmt.Sprintf("%s | %d messages", m.emailClient.GetUserEmail(), m.inbox.MessageCount())
	case ReaderView:
		title = "📖 Reading Message"
		status = fmt.Sprintf("%s", m.emailClient.GetUserEmail())
	case ComposerView:
		title = "✏️  Compose Message"
		status = fmt.Sprintf("From: %s", m.emailClient.GetUserEmail())
	}

	return HeaderStyle.Width(m.width).Render(
		RenderHeader(title, status, m.width),
	)
}

// renderContent renders the main content area
func (m Model) renderContent() string {
	contentHeight := m.height - 4 // Header + help + padding

	switch m.viewMode {
	case InboxView:
		return m.inbox.View()
	case ReaderView:
		return m.reader.View()
	case ComposerView:
		return m.composer.View()
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Render("Unknown view")
}

// renderHelp renders context-sensitive help
func (m Model) renderHelp() string {
	var helps []string

	switch m.viewMode {
	case InboxView:
		helps = []string{
			FormatKeyHelp("↑/↓", "navigate"),
			FormatKeyHelp("enter", "read"),
			FormatKeyHelp("c", "compose"),
			FormatKeyHelp("r", "reply"),
			FormatKeyHelp("q", "quit"),
		}
	case ReaderView:
		helps = []string{
			FormatKeyHelp("esc", "back"),
			FormatKeyHelp("r", "reply"),
			FormatKeyHelp("q", "back to inbox"),
		}
	case ComposerView:
		helps = []string{
			FormatKeyHelp("ctrl+s", "send"),
			FormatKeyHelp("esc", "cancel"),
			FormatKeyHelp("tab", "next field"),
		}
	}

	return RenderHelp(helps, m.width)
}

// renderError renders error messages
func (m Model) renderError() string {
	errorContent := ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(errorContent)

	help := RenderHelp([]string{FormatKeyHelp("q", "quit")}, m.width)

	return lipgloss.JoinVertical(lipgloss.Top, content, help)
}

// Message types for communication between components
type ErrorMsg struct {
	Error error
}

type StatusMsg struct {
	Message string
}

// Helper functions for creating messages
func NewErrorMsg(err error) ErrorMsg {
	return ErrorMsg{Error: err}
}

func NewStatusMsg(message string) StatusMsg {
	return StatusMsg{Message: message}
}
