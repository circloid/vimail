package email

import (
	"encoding/base64"
	"fmt"
	"mime"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

// Message represents an email message with parsed content
type Message struct {
	ID        string
	ThreadID  string
	From      string
	To        string
	Subject   string
	Date      time.Time
	Body      string
	Snippet   string
	Labels    []string
	Unread    bool
}

// NewMessageFromGmail creates a Message from a Gmail API message
func NewMessageFromGmail(gmailMsg *gmail.Message) (*Message, error) {
	msg := &Message{
		ID:       gmailMsg.Id,
		ThreadID: gmailMsg.ThreadId,
		Snippet:  gmailMsg.Snippet,
		Labels:   gmailMsg.LabelIds,
		Unread:   contains(gmailMsg.LabelIds, "UNREAD"),
	}

	// Parse headers
	if err := msg.parseHeaders(gmailMsg.Payload.Headers); err != nil {
		return nil, fmt.Errorf("failed to parse headers: %w", err)
	}

	// Parse body
	if err := msg.parseBody(gmailMsg.Payload); err != nil {
		return nil, fmt.Errorf("failed to parse body: %w", err)
	}

	return msg, nil
}

// parseHeaders extracts relevant information from email headers
func (m *Message) parseHeaders(headers []*gmail.MessagePartHeader) error {
	for _, header := range headers {
		switch strings.ToLower(header.Name) {
		case "from":
			m.From = cleanEmailAddress(header.Value)
		case "to":
			m.To = cleanEmailAddress(header.Value)
		case "subject":
			decoded, err := decodeHeader(header.Value)
			if err != nil {
				m.Subject = header.Value // fallback to raw value
			} else {
				m.Subject = decoded
			}
		case "date":
			date, err := parseDate(header.Value)
			if err != nil {
				return fmt.Errorf("failed to parse date: %w", err)
			}
			m.Date = date
		}
	}

	return nil
}

// parseBody extracts plain text content from the email body
func (m *Message) parseBody(part *gmail.MessagePart) error {
	if part.Body != nil && part.Body.Data != "" {
		// Single part message
		decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err != nil {
			return fmt.Errorf("failed to decode body: %w", err)
		}

		contentType := getContentType(part.Headers)
		m.Body = processBodyContent(string(decoded), contentType)
		return nil
	}

	// Multi-part message
	if part.Parts != nil {
		var textParts []string
		for _, subPart := range part.Parts {
			if subPart.Body != nil && subPart.Body.Data != "" {
				decoded, err := base64.URLEncoding.DecodeString(subPart.Body.Data)
				if err != nil {
					continue // skip parts that can't be decoded
				}

				contentType := getContentType(subPart.Headers)
				if isTextContent(contentType) {
					content := processBodyContent(string(decoded), contentType)
					if content != "" {
						textParts = append(textParts, content)
					}
				}
			}
		}

		m.Body = strings.Join(textParts, "\n\n")
	}

	return nil
}

// getContentType extracts content type from headers
func getContentType(headers []*gmail.MessagePartHeader) string {
	for _, header := range headers {
		if strings.ToLower(header.Name) == "content-type" {
			return strings.ToLower(header.Value)
		}
	}
	return ""
}

// isTextContent checks if content type is text-based
func isTextContent(contentType string) bool {
	return strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "text/html")
}

// processBodyContent processes body content based on content type
func processBodyContent(content, contentType string) string {
	if strings.Contains(contentType, "text/html") {
		// Strip HTML tags for plain text display
		return stripHTML(content)
	}
	return content
}

// stripHTML removes HTML tags and returns plain text
func stripHTML(html string) string {
	// Simple HTML tag removal - not perfect but adequate for MVP
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")
	
	// Clean up common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	
	// Clean up whitespace
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	
	return strings.Join(cleanLines, "\n")
}

// cleanEmailAddress extracts email address from "Name <email>" format
func cleanEmailAddress(addr string) string {
	// Extract email from "Name <email@domain.com>" format
	re := regexp.MustCompile(`<([^>]+)>`)
	if matches := re.FindStringSubmatch(addr); len(matches) > 1 {
		return matches[1]
	}
	
	// If no angle brackets, assume it's just the email
	return strings.TrimSpace(addr)
}

// decodeHeader decodes MIME encoded headers
func decodeHeader(header string) (string, error) {
	dec := new(mime.WordDecoder)
	return dec.DecodeHeader(header)
}

// parseDate parses various date formats commonly found in emails
func parseDate(dateStr string) (time.Time, error) {
	// Common email date formats
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"2 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"02 Jan 2006 15:04:05 -0700",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// FormatDate returns a human-readable date string
func (m *Message) FormatDate() string {
	now := time.Now()
	diff := now.Sub(m.Date)

	if diff < 24*time.Hour {
		return m.Date.Format("15:04")
	} else if diff < 7*24*time.Hour {
		return m.Date.Format("Mon 15:04")
	} else {
		return m.Date.Format("Jan 02")
	}
}

// GetDisplayFrom returns a formatted sender string
func (m *Message) GetDisplayFrom() string {
	if m.From == "" {
		return "Unknown Sender"
	}
	
	// Extract name part if present
	parts := strings.Split(m.From, "<")
	if len(parts) > 1 {
		name := strings.TrimSpace(parts[0])
		if name != "" && name != `""` {
			// Remove quotes if present
			name = strings.Trim(name, `"`)
			return name
		}
	}
	
	return m.From
}

// GetShortSubject returns a truncated subject for list display
func (m *Message) GetShortSubject(maxLength int) string {
	if len(m.Subject) <= maxLength {
		return m.Subject
	}
	return m.Subject[:maxLength-3] + "..."
}

// IsUnread returns whether the message is unread
func (m *Message) IsUnread() bool {
	return m.Unread
}
