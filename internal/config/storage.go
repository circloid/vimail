package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0700)
}

// BackupConfig creates a backup of the current config file
func BackupConfig() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // No config to backup
	}

	backupPath := configPath + ".backup"

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config for backup: %w", err)
	}

	err = os.WriteFile(backupPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// RestoreConfig restores config from backup
func RestoreConfig() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	backupPath := configPath + ".backup"

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup file found")
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	return nil
}

// RemoveConfig removes the config file
func RemoveConfig() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // Already doesn't exist
	}

	return os.Remove(configPath)
}
