package auth

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	// OAuth 2.0 redirect URI for installed applications
	RedirectURI = "http://localhost:8080/callback"
	// OAuth 2.0 scopes required for Gmail access
	GmailReadScope = gmail.GmailReadonlyScope
	GmailSendScope = gmail.GmailSendScope
)

// OAuthFlow handles the OAuth 2.0 authentication flow
type OAuthFlow struct {
	config *oauth2.Config
	server *http.Server
	token  *oauth2.Token
	done   chan struct{}
	err    error
}

// NewOAuthFlow creates a new OAuth flow with the given client credentials
func NewOAuthFlow(clientID, clientSecret string) *OAuthFlow {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  RedirectURI,
		Scopes: []string{
			GmailReadScope,
			GmailSendScope,
		},
		Endpoint: google.Endpoint,
	}

	return &OAuthFlow{
		config: config,
		done:   make(chan struct{}),
	}
}

// Authenticate starts the OAuth flow and returns a token
func (o *OAuthFlow) Authenticate(ctx context.Context) (*oauth2.Token, error) {
	// Start local server for callback
	if err := o.startCallbackServer(); err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer o.stopCallbackServer()

	// Generate authorization URL
	state := generateRandomState()
	authURL := o.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))

	fmt.Printf("Opening browser for authentication...\n")
	fmt.Printf("If browser doesn't open, visit this URL manually:\n%s\n\n", authURL)

	// Open browser
	if err := o.openBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser automatically: %v\n", err)
	}

	// Wait for callback or context cancellation
	select {
	case <-o.done:
		if o.err != nil {
			return nil, o.err
		}
		return o.token, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authentication timeout")
	}
}

// startCallbackServer starts the local HTTP server for OAuth callback
func (o *OAuthFlow) startCallbackServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", o.handleCallback)

	o.server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := o.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			o.err = err
			close(o.done)
		}
	}()

	return nil
}

// stopCallbackServer stops the local HTTP server
func (o *OAuthFlow) stopCallbackServer() {
	if o.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		o.server.Shutdown(ctx)
	}
}

// handleCallback processes the OAuth callback
func (o *OAuthFlow) handleCallback(w http.ResponseWriter, r *http.Request) {
	defer close(o.done)

	// Parse query parameters
	query := r.URL.Query()
	code := query.Get("code")
	if code == "" {
		o.err = fmt.Errorf("no authorization code received")
		http.Error(w, "Authorization failed", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := o.config.Exchange(context.Background(), code)
	if err != nil {
		o.err = fmt.Errorf("token exchange failed: %w", err)
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	o.token = token

	// Send success response
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Authentication Successful</title>
			<style>
				body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
				.success { color: #4CAF50; font-size: 24px; }
			</style>
		</head>
		<body>
			<div class="success">✓ Authentication Successful!</div>
			<p>You can close this window and return to the terminal.</p>
		</body>
		</html>
	`))
}

// openBrowser attempts to open the URL in the default browser
func (o *OAuthFlow) openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// RefreshTokenIfNeeded refreshes the token if it's expired
func RefreshTokenIfNeeded(token *oauth2.Token, clientID, clientSecret string) (*oauth2.Token, error) {
	if token.Valid() {
		return token, nil
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
	}

	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

// generateRandomState generates a random state parameter for OAuth
func generateRandomState() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ValidateToken checks if a token is valid and not expired
func ValidateToken(token *oauth2.Token) error {
	if token == nil {
		return fmt.Errorf("token is nil")
	}

	if token.AccessToken == "" {
		return fmt.Errorf("access token is empty")
	}

	if !token.Valid() {
		return fmt.Errorf("token is expired")
	}

	return nil
}
