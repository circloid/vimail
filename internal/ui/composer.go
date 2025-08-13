package ui

import (
	"context"
	"strings"
	"veloci_mail/internal/email"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ComposerField represents the current input field
type ComposerField int

const (
	ToField ComposerField = iota
	SubjectField
	BodyField
)

// ComposerModelImpl represents the email composition view state
type ComposerModelImpl struct {
	fromEmail    string
	to           string
	subject      string
	body         []string
	currentField ComposerField
	cursorPos    int
	bodyLine     int
	width        int
	height       int
	sending      bool
	Sent         bool
	Cancelled    bool
	err          error
	emailClient  *email.Client
}

// NewComposerModelImpl creates a new composer model
func NewComposerModelImpl(fromEmail string) *ComposerModelImpl {
	return &ComposerModelImpl{
		fromEmail:    fromEmail,
		currentField: ToField,
		body:         []string{""},
		bodyLine:     0,
		cursorPos:    0,
	}
}

// NewReplyComposerModelImpl creates a composer model for replying to a message
func NewReplyComposerModelImpl(fromEmail string, originalMsg *email.Message) *ComposerModelImpl {
	composer := NewComposerModelImpl(fromEmail)
	composer.to = originalMsg.From
	composer.subject = email.PrepareReplySubject(originalMsg.Subject)

	// Add quoted original message
	quotedBody := []string{
		"",
		"",
		"On " + originalMsg.Date.Format("Mon, Jan 2, 2006 at 3:04 PM") + ", " + originalMsg.GetDisplayFrom() + " wrote:",
		"",
	}

	// Quote each line of the original message
	originalLines := strings.Split(originalMsg.Body, "\n")
	for _, line := range originalLines {
		quotedBody = append(quotedBody, "> "+line)
	}

	composer.body = quotedBody
	composer.currentField = BodyField

	return composer
}

// SendMessageMsg represents a message sending result
type SendMessageMsg struct {
	Success bool
	Error   error
}

// Init initializes the composer model
func (m *ComposerModelImpl) Init() tea.Cmd {
	return nil
}

// Update handles composer-specific updates
func (m *ComposerModelImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.sending {
		switch msg := msg.(type) {
		case SendMessageMsg:
			m.sending = false
			if msg.Success {
				m.Sent = true
			} else {
				m.err = msg.Error
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.Cancelled = true
			return m, nil

		case "ctrl+s":
			return m, m.sendMessage()

		case "tab", "shift+tab":
			m.nextField(msg.String() == "shift+tab")

		case "enter":
			return m.handleEnter()

		case "backspace":
			return m.handleBackspace()

		case "left":
			return m.handleLeft()

		case "right":
			return m.handleRight()

		case "up":
			return m.handleUp()

		case "down":
			return m.handleDown()

		case "home":
			m.cursorPos = 0

		case "end":
			m.cursorPos = len(m.getCurrentFieldText())

		default:
			// Handle regular character input
			if len(msg.String()) == 1 {
				return m.handleCharInput(msg.String())
			}
		}
	}

	return m, nil
}

// View renders the composer view
func (m *ComposerModelImpl) View() string {
	if m.sending {
		return m.renderSending()
	}

	if m.err != nil {
		return m.renderError()
	}

	return m.renderComposer()
}

// renderSending renders the sending state
func (m *ComposerModelImpl) renderSending() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render("📤 Sending message...")
}

// renderError renders error state
func (m *ComposerModelImpl) renderError() string {
	errorContent := ErrorStyle.Render("Failed to send message: " + m.err.Error())

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(errorContent)
}

// renderComposer renders the compose form
func (m *ComposerModelImpl) renderComposer() string {
	var sections []string

	// From field (read-only)
	fromSection := m.renderField("From:", m.fromEmail, false, false)
	sections = append(sections, fromSection)

	// To field
	toSection := m.renderField("To:", m.to, m.currentField == ToField, true)
	sections = append(sections, toSection)

	// Subject field
	subjectSection := m.renderField("Subject:", m.subject, m.currentField == SubjectField, true)
	sections = append(sections, subjectSection)

	// Body field
	bodySection := m.renderBodyField()
	sections = append(sections, bodySection)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return BorderStyle.
		Width(m.width-2).
		Height(m.height-1).
		Render(content)
}

// renderField renders a single input field
func (m *ComposerModelImpl) renderField(label, value string, focused, editable bool) string {
	labelStyle := InputLabelStyle.Copy()
	inputStyle := InputStyle.Copy()

	if focused && editable {
		inputStyle = FocusedInputStyle.Copy()
	}

	if !editable {
		inputStyle = inputStyle.Foreground(MutedColor)
	}

	// Show cursor if focused
	displayValue := value
	if focused && editable {
		if m.cursorPos <= len(value) {
			if m.cursorPos == len(value) {
				displayValue = value + "█"
			} else {
				displayValue = value[:m.cursorPos] + "█" + value[m.cursorPos+1:]
			}
		}
	}

	labelRendered := labelStyle.Render(label)
	inputRendered := inputStyle.Width(m.width - 20).Render(displayValue)

	return lipgloss.JoinHorizontal(lipgloss.Top, labelRendered, " ", inputRendered)
}

// renderBodyField renders the multi-line body field
func (m *ComposerModelImpl) renderBodyField() string {
	labelStyle := InputLabelStyle.Copy()

	// Calculate available space
	bodyHeight := m.height - 10 // Account for other fields and borders
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	var lines []string

	// Show a subset of body lines based on current position
	startLine := 0
	if m.bodyLine >= bodyHeight-1 {
		startLine = m.bodyLine - bodyHeight + 2
	}

	for i := 0; i < bodyHeight && startLine+i < len(m.body); i++ {
		lineIndex := startLine + i
		line := m.body[lineIndex]

		// Show cursor on current line if in body field
		if m.currentField == BodyField && lineIndex == m.bodyLine {
			if m.cursorPos <= len(line) {
				if m.cursorPos == len(line) {
					line = line + "█"
				} else {
					line = line[:m.cursorPos] + "█" + line[m.cursorPos+1:]
				}
			}
		}

		lines = append(lines, line)
	}

	// Fill empty lines if needed
	for len(lines) < bodyHeight {
		lines = append(lines, "")
	}

	bodyStyle := InputStyle.Copy()
	if m.currentField == BodyField {
		bodyStyle = FocusedInputStyle.Copy()
	}

	bodyContent := lipgloss.JoinVertical(lipgloss.Left, lines...)
	bodyRendered := bodyStyle.
		Width(m.width - 20).
		Height(bodyHeight).
		Render(bodyContent)

	labelRendered := labelStyle.Render("Message:")

	return lipgloss.JoinHorizontal(lipgloss.Top, labelRendered, " ", bodyRendered)
}

// nextField moves to the next/previous input field
func (m *ComposerModelImpl) nextField(reverse bool) {
	if reverse {
		switch m.currentField {
		case BodyField:
			m.currentField = SubjectField
		case SubjectField:
			m.currentField = ToField
		case ToField:
			m.currentField = BodyField
		}
	} else {
		switch m.currentField {
		case ToField:
			m.currentField = SubjectField
		case SubjectField:
			m.currentField = BodyField
		case BodyField:
			m.currentField = ToField
		}
	}

	m.cursorPos = len(m.getCurrentFieldText())
}

// getCurrentFieldText returns the text of the currently selected field
func (m *ComposerModelImpl) getCurrentFieldText() string {
	switch m.currentField {
	case ToField:
		return m.to
	case SubjectField:
		return m.subject
	case BodyField:
		if m.bodyLine < len(m.body) {
			return m.body[m.bodyLine]
		}
		return ""
	}
	return ""
}

// handleEnter handles the Enter key
func (m *ComposerModelImpl) handleEnter() (*ComposerModelImpl, tea.Cmd) {
	if m.currentField == BodyField {
		// Insert new line in body
		currentLine := ""
		if m.bodyLine < len(m.body) {
			currentLine = m.body[m.bodyLine]
		}

		beforeCursor := currentLine[:m.cursorPos]
		afterCursor := currentLine[m.cursorPos:]

		// Update current line and insert new line
		m.body[m.bodyLine] = beforeCursor
		newLine := afterCursor

		// Insert new line
		m.body = append(m.body[:m.bodyLine+1], append([]string{newLine}, m.body[m.bodyLine+1:]...)...)

		// Move to next line
		m.bodyLine++
		m.cursorPos = 0
	} else {
		// Move to next field
		m.nextField(false)
	}

	return m, nil
}

// handleBackspace handles the Backspace key
func (m *ComposerModelImpl) handleBackspace() (*ComposerModelImpl, tea.Cmd) {
	if m.currentField == BodyField {
		if m.cursorPos > 0 {
			// Remove character in current line
			line := m.body[m.bodyLine]
			m.body[m.bodyLine] = line[:m.cursorPos-1] + line[m.cursorPos:]
			m.cursorPos--
		} else if m.bodyLine > 0 {
			// Join with previous line
			prevLine := m.body[m.bodyLine-1]
			currentLine := m.body[m.bodyLine]

			// Remove current line
			m.body = append(m.body[:m.bodyLine], m.body[m.bodyLine+1:]...)

			// Update previous line
			m.bodyLine--
			m.body[m.bodyLine] = prevLine + currentLine
			m.cursorPos = len(prevLine)
		}
	} else {
		// Handle single-line fields
		text := m.getCurrentFieldText()
		if m.cursorPos > 0 {
			newText := text[:m.cursorPos-1] + text[m.cursorPos:]
			m.setCurrentFieldText(newText)
			m.cursorPos--
		}
	}

	return m, nil
}

// handleLeft handles the Left arrow key
func (m *ComposerModelImpl) handleLeft() (*ComposerModelImpl, tea.Cmd) {
	if m.cursorPos > 0 {
		m.cursorPos--
	}
	return m, nil
}

// handleRight handles the Right arrow key
func (m *ComposerModelImpl) handleRight() (*ComposerModelImpl, tea.Cmd) {
	text := m.getCurrentFieldText()
	if m.cursorPos < len(text) {
		m.cursorPos++
	}
	return m, nil
}

// handleUp handles the Up arrow key
func (m *ComposerModelImpl) handleUp() (*ComposerModelImpl, tea.Cmd) {
	if m.currentField == BodyField && m.bodyLine > 0 {
		m.bodyLine--
		// Adjust cursor position to fit new line
		newLineLen := len(m.body[m.bodyLine])
		if m.cursorPos > newLineLen {
			m.cursorPos = newLineLen
		}
	}
	return m, nil
}

// handleDown handles the Down arrow key
func (m *ComposerModelImpl) handleDown() (*ComposerModelImpl, tea.Cmd) {
	if m.currentField == BodyField && m.bodyLine < len(m.body)-1 {
		m.bodyLine++
		// Adjust cursor position to fit new line
		newLineLen := len(m.body[m.bodyLine])
		if m.cursorPos > newLineLen {
			m.cursorPos = newLineLen
		}
	}
	return m, nil
}

// handleCharInput handles regular character input
func (m *ComposerModelImpl) handleCharInput(char string) (*ComposerModelImpl, tea.Cmd) {
	if m.currentField == BodyField {
		// Insert character in body
		line := m.body[m.bodyLine]
		newLine := line[:m.cursorPos] + char + line[m.cursorPos:]
		m.body[m.bodyLine] = newLine
		m.cursorPos++
	} else {
		// Insert character in single-line field
		text := m.getCurrentFieldText()
		newText := text[:m.cursorPos] + char + text[m.cursorPos:]
		m.setCurrentFieldText(newText)
		m.cursorPos++
	}

	return m, nil
}

// setCurrentFieldText sets the text of the currently selected field
func (m *ComposerModelImpl) setCurrentFieldText(text string) {
	switch m.currentField {
	case ToField:
		m.to = text
	case SubjectField:
		m.subject = text
	case BodyField:
		if m.bodyLine < len(m.body) {
			m.body[m.bodyLine] = text
		}
	}
}

// sendMessage creates a command to send the email
func (m *ComposerModelImpl) sendMessage() tea.Cmd {
	// Validate compose data
	composeData := email.ComposeData{
		To:      strings.TrimSpace(m.to),
		Subject: email.SanitizeSubject(m.subject),
		Body:    strings.Join(m.body, "\n"),
	}

	if err := composeData.Validate(); err != nil {
		return func() tea.Msg {
			return SendMessageMsg{Success: false, Error: err}
		}
	}

	m.sending = true

	return func() tea.Msg {
		err := m.emailClient.SendMessage(context.Background(), composeData.To, composeData.Subject, composeData.Body)
		return SendMessageMsg{
			Success: err == nil,
			Error:   err,
		}
	}
}

// SetSize updates the size of the composer view
func (m *ComposerModelImpl) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetEmailClient sets the email client for sending messages
func (m *ComposerModelImpl) SetEmailClient(client *email.Client) {
	m.emailClient = client
}

// IsSent returns whether the message was sent
func (m *ComposerModelImpl) IsSent() bool {
	return m.Sent
}

// IsCancelled returns whether composition was cancelled
func (m *ComposerModelImpl) IsCancelled() bool {
	return m.Cancelled
}
