// internal/ui/composer.go - Beautiful Zen Composer
package ui

import (
	"context"
	"strings"
	"vimail/internal/email"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ComposerField int

const (
	ToField ComposerField = iota
	SubjectField
	BodyField
)

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

func NewComposerModelImpl(fromEmail string, emailClient *email.Client) *ComposerModelImpl {
	return &ComposerModelImpl{
		fromEmail:    fromEmail,
		currentField: ToField,
		body:         []string{""},
		bodyLine:     0,
		cursorPos:    0,
		emailClient:  emailClient,
	}
}

func NewReplyComposerModelImpl(fromEmail string, originalMsg *email.Message, emailClient *email.Client) *ComposerModelImpl {
	composer := NewComposerModelImpl(fromEmail, emailClient)
	composer.to = originalMsg.From
	composer.subject = "Re: " + originalMsg.Subject
	return composer
}

type SendMessageMsg struct {
	Success bool
	Error   error
}

func (m *ComposerModelImpl) Init() tea.Cmd {
	return nil
}

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

		case "tab":
			m.nextField()

		case "enter":
			return m.handleEnter()

		case "backspace":
			return m.handleBackspace()

		case "left":
			if m.cursorPos > 0 {
				m.cursorPos--
			}

		case "right":
			text := m.getCurrentFieldText()
			if m.cursorPos < len(text) {
				m.cursorPos++
			}

		case "up":
			if m.currentField == BodyField && m.bodyLine > 0 {
				m.bodyLine--
				newLineLen := len(m.body[m.bodyLine])
				if m.cursorPos > newLineLen {
					m.cursorPos = newLineLen
				}
			}

		case "down":
			if m.currentField == BodyField && m.bodyLine < len(m.body)-1 {
				m.bodyLine++
				newLineLen := len(m.body[m.bodyLine])
				if m.cursorPos > newLineLen {
					m.cursorPos = newLineLen
				}
			}

		default:
			if len(msg.String()) == 1 {
				return m.handleCharInput(msg.String())
			}
		}
	}

	return m, nil
}

func (m *ComposerModelImpl) View() string {
	if m.sending {
		return lipgloss.NewStyle().
			Foreground(White).
			Align(lipgloss.Center, lipgloss.Center).
			Width(m.width).
			Height(m.height).
			Render("✉ Sending...")
	}

	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(White).
			Padding(2).
			Render("✗ Error: " + m.err.Error() + "\n\nPress Esc to go back")
	}

	var sections []string

	// Clean minimal form
	sections = append(sections, m.renderField("To:", m.to, m.currentField == ToField))
	sections = append(sections, "")
	sections = append(sections, m.renderField("Subject:", m.subject, m.currentField == SubjectField))
	sections = append(sections, "")
	sections = append(sections, m.renderBodyField())
	sections = append(sections, "")

	// Zen help text
	help := lipgloss.NewStyle().
		Foreground(Gray).
		Align(lipgloss.Center).
		Render("Tab: next field • Ctrl+S: send • Esc: cancel")

	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *ComposerModelImpl) renderField(label, value string, focused bool) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(Gray).
		Width(10)

	inputStyle := lipgloss.NewStyle().
		Foreground(White)

	if focused {
		labelStyle = labelStyle.Foreground(Blue)
		inputStyle = inputStyle.
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(Blue)
	}

	// Show cursor if focused
	displayValue := value
	if focused {
		if m.cursorPos <= len(value) {
			if m.cursorPos == len(value) {
				displayValue = value + "█"
			} else {
				displayValue = value[:m.cursorPos] + "█" + value[m.cursorPos+1:]
			}
		}
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		labelStyle.Render(label),
		" ",
		inputStyle.Width(m.width-15).Render(displayValue),
	)
}

func (m *ComposerModelImpl) renderBodyField() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(Gray).
		Width(10)

	if m.currentField == BodyField {
		labelStyle = labelStyle.Foreground(Blue)
	}

	bodyHeight := m.height - 10
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	var lines []string
	for i := 0; i < bodyHeight && i < len(m.body); i++ {
		line := m.body[i]

		// Show cursor on current line if in body field
		if m.currentField == BodyField && i == m.bodyLine {
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

	// Fill empty lines
	for len(lines) < bodyHeight {
		lines = append(lines, "")
	}

	bodyStyle := lipgloss.NewStyle().
		Foreground(White).
		Width(m.width - 15).
		Height(bodyHeight)

	if m.currentField == BodyField {
		bodyStyle = bodyStyle.
			Border(lipgloss.NormalBorder()).
			BorderForeground(Blue)
	}

	bodyContent := lipgloss.JoinVertical(lipgloss.Left, lines...)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		labelStyle.Render("Message:"),
		" ",
		bodyStyle.Render(bodyContent),
	)
}

func (m *ComposerModelImpl) nextField() {
	switch m.currentField {
	case ToField:
		m.currentField = SubjectField
	case SubjectField:
		m.currentField = BodyField
	case BodyField:
		m.currentField = ToField
	}
	m.cursorPos = len(m.getCurrentFieldText())
}

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

func (m *ComposerModelImpl) handleEnter() (*ComposerModelImpl, tea.Cmd) {
	if m.currentField == BodyField {
		currentLine := ""
		if m.bodyLine < len(m.body) {
			currentLine = m.body[m.bodyLine]
		}

		beforeCursor := currentLine[:m.cursorPos]
		afterCursor := currentLine[m.cursorPos:]

		m.body[m.bodyLine] = beforeCursor
		newLine := afterCursor

		m.body = append(m.body[:m.bodyLine+1], append([]string{newLine}, m.body[m.bodyLine+1:]...)...)
		m.bodyLine++
		m.cursorPos = 0
	} else {
		m.nextField()
	}

	return m, nil
}

func (m *ComposerModelImpl) handleBackspace() (*ComposerModelImpl, tea.Cmd) {
	if m.currentField == BodyField {
		if m.cursorPos > 0 {
			line := m.body[m.bodyLine]
			m.body[m.bodyLine] = line[:m.cursorPos-1] + line[m.cursorPos:]
			m.cursorPos--
		} else if m.bodyLine > 0 {
			prevLine := m.body[m.bodyLine-1]
			currentLine := m.body[m.bodyLine]

			m.body = append(m.body[:m.bodyLine], m.body[m.bodyLine+1:]...)
			m.bodyLine--
			m.body[m.bodyLine] = prevLine + currentLine
			m.cursorPos = len(prevLine)
		}
	} else {
		text := m.getCurrentFieldText()
		if m.cursorPos > 0 {
			newText := text[:m.cursorPos-1] + text[m.cursorPos:]
			m.setCurrentFieldText(newText)
			m.cursorPos--
		}
	}

	return m, nil
}

func (m *ComposerModelImpl) handleCharInput(char string) (*ComposerModelImpl, tea.Cmd) {
	if m.currentField == BodyField {
		line := m.body[m.bodyLine]
		newLine := line[:m.cursorPos] + char + line[m.cursorPos:]
		m.body[m.bodyLine] = newLine
		m.cursorPos++
	} else {
		text := m.getCurrentFieldText()
		newText := text[:m.cursorPos] + char + text[m.cursorPos:]
		m.setCurrentFieldText(newText)
		m.cursorPos++
	}

	return m, nil
}

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

func (m *ComposerModelImpl) sendMessage() tea.Cmd {
	composeData := email.ComposeData{
		To:      strings.TrimSpace(m.to),
		Subject: strings.TrimSpace(m.subject),
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

func (m *ComposerModelImpl) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ComposerModelImpl) IsSent() bool {
	return m.Sent
}

func (m *ComposerModelImpl) IsCancelled() bool {
	return m.Cancelled
}
