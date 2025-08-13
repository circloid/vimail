# Veloci Mail Makefile

# Application name
APP_NAME = veloci_mail

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# Binary names
BINARY_NAME = $(APP_NAME)
BINARY_UNIX = $(BINARY_NAME)_unix
BINARY_WINDOWS = $(BINARY_NAME).exe
BINARY_DARWIN = $(BINARY_NAME)_darwin

# Build flags
LDFLAGS = -ldflags "-w -s"
BUILD_FLAGS = $(LDFLAGS)

# Default target
.PHONY: all
all: deps test build

# Setup dependencies (fix common issues)
.PHONY: setup
setup:
	@echo "Setting up dependencies..."
	$(GOMOD) tidy
	$(GOGET) golang.org/x/oauth2@v0.30.0
	$(GOGET) google.golang.org/api/gmail/v1@v0.247.0
	$(GOGET) github.com/charmbracelet/bubbletea@v1.3.6
	$(GOGET) github.com/charmbracelet/lipgloss@v1.1.0
	$(GOGET) cloud.google.com/go/compute/metadata@v0.8.0
	$(GOGET) github.com/googleapis/gax-go/v2@v2.14.1
	$(GOGET) cloud.google.com/go/auth@v0.15.0
	$(GOGET) cloud.google.com/go/auth/credentials@v0.15.0
	$(GOGET) cloud.google.com/go/auth/oauth2adapt@v0.2.7
	$(GOGET) github.com/google/s2a-go@v0.1.8
	$(GOGET) github.com/googleapis/enterprise-certificate-proxy@v0.3.4
	$(GOGET) google.golang.org/grpc@v1.72.0
	$(GOGET) github.com/google/uuid@v1.6.0
	$(GOGET) golang.org/x/net/http2@v0.35.0
	$(GOGET) go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@v0.58.0
	$(GOMOD) tidy
	@echo "Dependencies setup complete!"

# Build the binary
.PHONY: build
build:
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) -v .

# Test all packages
.PHONY: test
test:
	$(GOTEST) -v ./...

# Test with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_WINDOWS)
	rm -f $(BINARY_DARWIN)
	rm -f coverage.out
	rm -f coverage.html

# Run the application
.PHONY: run
run: build
	./$(BINARY_NAME)

# Install dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) verify

# Update dependencies
.PHONY: deps-update
deps-update:
	$(GOMOD) tidy
	$(GOGET) -u ./...

# Cross compilation
.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_UNIX) -v .

.PHONY: build-windows
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_WINDOWS) -v .

.PHONY: build-darwin
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_DARWIN) -v .

.PHONY: build-all
build-all: build-linux build-windows build-darwin

# Development helpers
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

.PHONY: vet
vet:
	$(GOCMD) vet ./...

.PHONY: lint
lint:
	golangci-lint run

# Check code quality
.PHONY: check
check: fmt vet test

# Setup development environment
.PHONY: dev-setup
dev-setup:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install the binary to GOPATH/bin
.PHONY: install
install:
	$(GOCMD) install $(BUILD_FLAGS) .

# Create release archive
.PHONY: release
release: build-all
	mkdir -p releases
	tar -czf releases/$(APP_NAME)_linux_amd64.tar.gz $(BINARY_UNIX)
	zip -r releases/$(APP_NAME)_windows_amd64.zip $(BINARY_WINDOWS)
	tar -czf releases/$(APP_NAME)_darwin_amd64.tar.gz $(BINARY_DARWIN)

# Fix common dependency issues
.PHONY: fix-deps
fix-deps:
	$(GOCMD) clean -modcache
	$(GOMOD) tidy
	$(GOGET) -u all
	$(GOMOD) tidy

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  setup         - Setup all dependencies (run this first!)"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Build and run the application"
	@echo "  deps          - Download dependencies"
	@echo "  deps-update   - Update dependencies"
	@echo "  build-linux   - Build for Linux"
	@echo "  build-windows - Build for Windows"
	@echo "  build-darwin  - Build for macOS"
	@echo "  build-all     - Build for all platforms"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run golangci-lint"
	@echo "  check         - Run fmt, vet, and test"
	@echo "  dev-setup     - Setup development tools"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  release       - Create release archives"
	@echo "  fix-deps      - Fix dependency issues"
	@echo "  help          - Show this help message"
