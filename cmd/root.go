package cmd

import (
	"fmt"
	"os"
)

// This file is optional for the MVP
// We keep the main logic in main.go for simplicity
// This could be used later for more complex CLI commands

// Execute is the main entry point for the cobra CLI
// Currently unused but kept for future expansion
func Execute() {
	fmt.Println("Veloci Mail - Terminal Email Client")
	fmt.Println("Run 'go run .' from the project root to start the application")
	os.Exit(0)
}

// RootCmd would be used if we implement cobra CLI
// For MVP, we keep it simple with direct main.go execution
var RootCmd = struct {
	Use   string
	Short string
}{
	Use:   "veloci_mail",
	Short: "A fast terminal email client",
}
