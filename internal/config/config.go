package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

// Config represents the application configuration
type Config struct {
	OAuth     OAuthConfig `json:"oauth"`
	UserEmail string      `json:"user_email,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// OAuthConfig holds OAuth 2.0 configuration and tokens
type OAuthConfig struct {
	ClientID     string       `json:"client_id"`
	ClientSecret string       `json:"client_secret"`
	Token        *oauth2.Token `json:"token,omitempty"`
}

const (
	ConfigFileName = "config.json"
	ConfigDirName  = ".terminal-email"
)

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigDirName)
	
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, ConfigFileName), nil
}

// Load reads and parses the configuration file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found at %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save writes the configuration to file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	c.UpdatedAt = time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = c.UpdatedAt
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists checks if a configuration file already exists
func Exists() bool {
	configPath, err := GetConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(configPath)
	return !os.IsNotExist(err)
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
