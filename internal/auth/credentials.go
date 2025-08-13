package auth

import (
	"fmt"
	"golang.org/x/oauth2"
	"time"
)

// Credentials represents stored OAuth credentials
type Credentials struct {
	ClientID     string       `json:"client_id"`
	ClientSecret string       `json:"client_secret"`
	Token        *oauth2.Token `json:"token,omitempty"`
}

// IsValid checks if the credentials are valid
func (c *Credentials) IsValid() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

// HasToken checks if there's a stored token
func (c *Credentials) HasToken() bool {
	return c.Token != nil && c.Token.AccessToken != ""
}

// IsTokenExpired checks if the token is expired
func (c *Credentials) IsTokenExpired() bool {
	if c.Token == nil {
		return true
	}
	return c.Token.Expiry.Before(time.Now())
}

// ValidateCredentials performs basic validation on OAuth credentials
func ValidateCredentials(clientID, clientSecret string) error {
	if clientID == "" {
		return fmt.Errorf("client ID cannot be empty")
	}
	if clientSecret == "" {
		return fmt.Errorf("client secret cannot be empty")
	}
	if len(clientID) < 10 {
		return fmt.Errorf("client ID appears to be invalid (too short)")
	}
	if len(clientSecret) < 10 {
		return fmt.Errorf("client secret appears to be invalid (too short)")
	}
	return nil
}
