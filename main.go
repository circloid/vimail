package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"veloci_mail/internal/auth"
	"veloci_mail/internal/config"
	"veloci_mail/internal/email"
	"veloci_mail/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Setup context
	ctx := context.Background()

	// Check if config exists, if not run setup
	if !config.Exists() {
		if err := runSetup(ctx); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Refresh token if needed
	newToken, err := auth.RefreshTokenIfNeeded(cfg.OAuth.Token, cfg.OAuth.ClientID, cfg.OAuth.ClientSecret)
	if err != nil {
		log.Fatalf("Failed to refresh token: %v", err)
	}

	// Update config if token was refreshed
	if newToken != cfg.OAuth.Token {
		cfg.OAuth.Token = newToken
		if err := cfg.Save(); err != nil {
			log.Printf("Warning: Failed to save updated token: %v", err)
		}
	}

	// Create email client
	emailClient, err := email.NewClient(ctx, cfg.OAuth.Token, cfg.OAuth.ClientID, cfg.OAuth.ClientSecret)
	if err != nil {
		log.Fatalf("Failed to create email client: %v", err)
	}

	// Test connection
	if err := emailClient.TestConnection(ctx); err != nil {
		log.Fatalf("Gmail API connection failed: %v", err)
	}

	// Update user email in config
	if cfg.UserEmail != emailClient.GetUserEmail() {
		cfg.UserEmail = emailClient.GetUserEmail()
		if err := cfg.Save(); err != nil {
			log.Printf("Warning: Failed to save user email: %v", err)
		}
	}

	// Create and run TUI application
	model := ui.NewModel(ctx, emailClient, cfg)

	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := program.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

// runSetup handles the initial OAuth setup
func runSetup(ctx context.Context) error {
	fmt.Println("🔧 Terminal Email Client Setup")
	fmt.Println("==============================")
	fmt.Println()

	// Get OAuth credentials from user
	var clientID, clientSecret string

	fmt.Println("To use this email client, you need to create a Google OAuth application.")
	fmt.Println("Visit: https://console.developers.google.com/")
	fmt.Println("1. Create a new project or select existing")
	fmt.Println("2. Enable Gmail API")
	fmt.Println("3. Create OAuth 2.0 credentials (Desktop application)")
	fmt.Println("4. Download the credentials and enter them below")
	fmt.Println()

	fmt.Print("Enter Client ID: ")
	if _, err := fmt.Scanln(&clientID); err != nil {
		return fmt.Errorf("failed to read client ID: %w", err)
	}

	fmt.Print("Enter Client Secret: ")
	if _, err := fmt.Scanln(&clientSecret); err != nil {
		return fmt.Errorf("failed to read client secret: %w", err)
	}

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("client ID and secret are required")
	}

	fmt.Println()
	fmt.Println("🔐 Starting OAuth authentication...")

	// Perform OAuth flow
	oauthFlow := auth.NewOAuthFlow(clientID, clientSecret)
	token, err := oauthFlow.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("OAuth authentication failed: %w", err)
	}

	fmt.Println("✅ Authentication successful!")

	// Create and save configuration
	cfg := config.NewConfig()
	cfg.OAuth.ClientID = clientID
	cfg.OAuth.ClientSecret = clientSecret
	cfg.OAuth.Token = token

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Test the connection
	fmt.Println("🧪 Testing Gmail API connection...")
	emailClient, err := email.NewClient(ctx, token, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("failed to create email client: %w", err)
	}

	if err := emailClient.TestConnection(ctx); err != nil {
		return fmt.Errorf("Gmail API connection test failed: %w", err)
	}

	// Update config with user email
	cfg.UserEmail = emailClient.GetUserEmail()
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save user email: %w", err)
	}

	fmt.Printf("✅ Setup complete! Connected as: %s\n", emailClient.GetUserEmail())
	fmt.Println()
	fmt.Println("🚀 Starting Terminal Email Client...")
	fmt.Println()

	return nil
}

// showUsage displays usage information
func showUsage() {
	fmt.Println("Terminal Email Client")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("A minimal CLI email client for Gmail")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  terminal-email-client          # Start the client")
	fmt.Println("  terminal-email-client --help   # Show this help")
	fmt.Println("  terminal-email-client --setup  # Re-run setup")
	fmt.Println()
	fmt.Println("First time setup:")
	fmt.Println("  1. Run the application")
	fmt.Println("  2. Follow OAuth setup instructions")
	fmt.Println("  3. Authenticate with your Google account")
	fmt.Println()
	fmt.Println("Navigation:")
	fmt.Println("  ↑/↓ or j/k    Navigate message list")
	fmt.Println("  Enter         Read selected message")
	fmt.Println("  c             Compose new message")
	fmt.Println("  r             Reply to message")
	fmt.Println("  q             Quit application")
	fmt.Println("  Esc           Go back/cancel")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Config file: ~/.terminal-email/config.json")
	fmt.Println()
}

func init() {
	// Check for help flag
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			showUsage()
			os.Exit(0)
		}
		if arg == "--setup" {
			// Force re-setup by removing config
			configPath, err := config.GetConfigPath()
			if err == nil {
				os.Remove(configPath)
			}
		}
	}
}
