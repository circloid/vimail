//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🔍 Verifying Veloci Mail setup...")

	// Check if all required files exist
	requiredFiles := []string{
		"main.go",
		"internal/auth/oauth.go",
		"internal/auth/credentials.go",
		"internal/config/config.go",
		"internal/config/storage.go",
		"internal/email/client.go",
		"internal/email/message.go",
		"internal/email/compose.go",
		"internal/ui/app.go",
		"internal/ui/inbox.go",
		"internal/ui/reader.go",
		"internal/ui/composer.go",
		"internal/ui/styles.go",
	}

	missing := []string{}
	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			missing = append(missing, file)
		}
	}

	if len(missing) > 0 {
		fmt.Printf("❌ Missing files:\n")
		for _, file := range missing {
			fmt.Printf("   - %s\n", file)
		}
		return
	}

	fmt.Println("✅ All required files present")

	// Check for syntax errors
	fmt.Println("🔍 Checking syntax...")

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "verify.go") {
			return nil
		}

		fset := token.NewFileSet()
		_, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			fmt.Printf("❌ Syntax error in %s: %v\n", path, parseErr)
			return parseErr
		}

		return nil
	})

	if err != nil {
		fmt.Printf("❌ Syntax check failed: %v\n", err)
		return
	}

	fmt.Println("✅ Syntax check passed")

	// Check go.mod
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("❌ go.mod missing")
		return
	}

	fmt.Println("✅ go.mod present")

	fmt.Println("\n🚀 Setup verification complete!")
	fmt.Println("Run 'go mod tidy && go run .' to start the application")
}

func checkGoFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Basic AST check
	ast.Inspect(node, func(n ast.Node) bool {
		return true
	})

	return nil
}
