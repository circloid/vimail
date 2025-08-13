package email

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Client wraps the Gmail API client with our application logic
type Client struct {
	service   *gmail.Service
	userEmail string
}

// NewClient creates a new email client with the provided OAuth token
func NewClient(ctx context.Context, token *oauth2.Token, clientID, clientSecret string) (*Client, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	httpClient := config.Client(ctx, token)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	client := &Client{
		service: service,
	}

	// Get user profile to store email address
	profile, err := client.service.Users.GetProfile("me").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	client.userEmail = profile.EmailAddress

	return client, nil
}

// GetUserEmail returns the authenticated user's email address
func (c *Client) GetUserEmail() string {
	return c.userEmail
}

// ListMessages retrieves messages from the specified label (folder)
func (c *Client) ListMessages(ctx context.Context, label string, maxResults int64) ([]*Message, error) {
	query := ""
	if label != "" {
		query = fmt.Sprintf("label:%s", label)
	}

	req := c.service.Users.Messages.List("me").
		Context(ctx).
		Q(query).
		MaxResults(maxResults)

	resp, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	var messages []*Message
	for _, msg := range resp.Messages {
		fullMsg, err := c.service.Users.Messages.Get("me", msg.Id).
			Context(ctx).
			Do()
		if err != nil {
			// Skip messages that can't be retrieved
			continue
		}

		message, err := NewMessageFromGmail(fullMsg)
		if err != nil {
			// Skip messages that can't be parsed
			continue
		}

		messages = append(messages, message)
	}

	return messages, nil
}

// GetMessage retrieves a specific message by ID
func (c *Client) GetMessage(ctx context.Context, messageID string) (*Message, error) {
	gmailMsg, err := c.service.Users.Messages.Get("me", messageID).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	message, err := NewMessageFromGmail(gmailMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	return message, nil
}

// SendMessage sends an email message
func (c *Client) SendMessage(ctx context.Context, to, subject, body string) error {
	message := &gmail.Message{
		Raw: encodeMessage(c.userEmail, to, subject, body),
	}

	_, err := c.service.Users.Messages.Send("me", message).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// GetInboxMessages retrieves messages from the inbox
func (c *Client) GetInboxMessages(ctx context.Context, maxResults int64) ([]*Message, error) {
	return c.ListMessages(ctx, "INBOX", maxResults)
}

// GetSentMessages retrieves sent messages
func (c *Client) GetSentMessages(ctx context.Context, maxResults int64) ([]*Message, error) {
	return c.ListMessages(ctx, "SENT", maxResults)
}

// TestConnection verifies the Gmail API connection
func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.service.Users.GetProfile("me").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("Gmail API connection test failed: %w", err)
	}
	return nil
}
