package email

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

// ComposeData holds the data for composing an email
type ComposeData struct {
	To      string
	Subject string
	Body    string
}

// Validate checks if the compose data is valid
func (c *ComposeData) Validate() error {
	if c.To == "" {
		return fmt.Errorf("recipient (To) is required")
	}

	// Validate email address format
	if _, err := mail.ParseAddress(c.To); err != nil {
		return fmt.Errorf("invalid email address: %s", c.To)
	}

	if c.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if c.Body == "" {
		return fmt.Errorf("message body is required")
	}

	return nil
}

// encodeMessage creates a base64-encoded email message for Gmail API
func encodeMessage(from, to, subject, body string) string {
	// Create email headers
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
	}

	// Combine headers and body
	message := strings.Join(headers, "\r\n") + body

	// Encode as base64 URL-safe
	return base64.URLEncoding.EncodeToString([]byte(message))
}

// FormatBodyForDisplay prepares body text for display in the compose view
func FormatBodyForDisplay(body string) string {
	// Ensure proper line endings and formatting
	lines := strings.Split(body, "\n")
	var formattedLines []string

	for _, line := range lines {
		// Trim trailing whitespace but preserve leading spaces
		line = strings.TrimRight(line, " \t")
		formattedLines = append(formattedLines, line)
	}

	return strings.Join(formattedLines, "\n")
}

// WrapText wraps text to specified width for better display
func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		if len(line) <= width {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		// Wrap long lines
		words := strings.Fields(line)
		if len(words) == 0 {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				wrappedLines = append(wrappedLines, currentLine)
				currentLine = word
			}
		}
		wrappedLines = append(wrappedLines, currentLine)
	}

	return strings.Join(wrappedLines, "\n")
}

// ValidateEmailAddress checks if an email address is valid
func ValidateEmailAddress(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email address cannot be empty")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email address format: %w", err)
	}

	return nil
}

// SanitizeSubject removes or replaces characters that might cause issues
func SanitizeSubject(subject string) string {
	// Remove control characters and normalize whitespace
	subject = strings.ReplaceAll(subject, "\n", " ")
	subject = strings.ReplaceAll(subject, "\r", " ")
	subject = strings.ReplaceAll(subject, "\t", " ")
	
	// Normalize multiple spaces to single space
	for strings.Contains(subject, "  ") {
		subject = strings.ReplaceAll(subject, "  ", " ")
	}
	
	return strings.TrimSpace(subject)
}

// PrepareReplySubject prepares a subject line for a reply
func PrepareReplySubject(originalSubject string) string {
	subject := strings.TrimSpace(originalSubject)
	
	// Don't add Re: if it's already there
	if strings.HasPrefix(strings.ToLower(subject), "re:") {
		return subject
	}
	
	return "Re: " + subject
}

// PrepareForwardSubject prepares a subject line for forwarding
func PrepareForwardSubject(originalSubject string) string {
	subject := strings.TrimSpace(originalSubject)
	
	// Don't add Fwd: if it's already there
	if strings.HasPrefix(strings.ToLower(subject), "fwd:") ||
		strings.HasPrefix(strings.ToLower(subject), "fw:") {
		return subject
	}
	
	return "Fwd: " + subject
}
