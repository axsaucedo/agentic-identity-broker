# Default recipe (shown when running `just` with no args)
default:
    @just --list

# =============================================================================
# Go Development Targets
# =============================================================================

# Build the Go binary to ./bin/identity-broker
build:
    @echo "Building identity-broker..."
    @mkdir -p bin
    go build -o bin/identity-broker ./cmd/identity-broker

# Build with optimizations for production (smaller binary, no debug info)
build-release:
    @echo "Building identity-broker (release mode)..."
    @mkdir -p bin
    go build -ldflags="-s -w" -o bin/identity-broker ./cmd/identity-broker

# Run all Go tests with verbose output
test:
    @echo "Running Go tests..."
    go test -v -race ./...

# Run tests with coverage report
test-coverage:
    @echo "Running tests with coverage..."
    @mkdir -p coverage
    go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html
    @echo "Coverage report generated at coverage/coverage.html"

# Run tests and display coverage percentage
test-coverage-summary:
    @echo "Running tests with coverage summary..."
    @mkdir -p coverage
    go test -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
    go tool cover -func=coverage/coverage.out

# Build and run the application
run: build
    @echo "Running identity-broker..."
    ./bin/identity-broker

# Run with Air for hot-reload development (requires air to be installed)
dev:
    @echo "Starting development server with hot reload..."
    @if command -v air > /dev/null; then \
        air; \
    else \
        echo "Error: air is not installed. Install it with: go install github.com/air-verse/air@latest"; \
        exit 1; \
    fi

# Clean build artifacts and temporary files
clean:
    @echo "Cleaning build artifacts..."
    rm -rf bin
    rm -rf coverage
    @echo "Clean complete"

# Download and tidy Go module dependencies
deps:
    @echo "Tidying Go modules..."
    go mod tidy
    @echo "Downloading Go modules..."
    go mod download
    @echo "Verifying Go modules..."
    go mod verify

# Run golangci-lint if available
lint:
    @echo "Running linter..."
    @if command -v golangci-lint > /dev/null; then \
        golangci-lint run ./...; \
    else \
        echo "Warning: golangci-lint is not installed. Install it from: https://golangci-lint.run/usage/install/"; \
        echo "Running basic go vet instead..."; \
        go vet ./...; \
    fi

# Format Go code with gofmt
fmt:
    @echo "Formatting Go code..."
    gofmt -s -w .
    @echo "Format complete"

# Run go vet for static analysis
vet:
    @echo "Running go vet..."
    go vet ./...

# Install development tools (air, golangci-lint)
install-tools:
    @echo "Installing development tools..."
    @echo "Installing air for hot reload..."
    go install github.com/air-verse/air@latest
    @echo "Installing golangci-lint..."
    @if ! command -v golangci-lint > /dev/null; then \
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
    else \
        echo "golangci-lint is already installed"; \
    fi
    @echo "Tools installation complete"

# Run all quality checks (fmt, vet, lint, test)
check: fmt vet lint test
    @echo "All checks passed!"

# =============================================================================
# Documentation Targets
# =============================================================================

# Install documentation dependencies
docs-install:
    cd assets/docusaurus && npm install

# Build documentation site
docs-build: docs-install
    cd assets/docusaurus && npm run build

# Serve documentation locally (hot reload)
docs-serve:
    cd assets/docusaurus && npm start

# Serve production build locally for testing
docs-preview: docs-build
    cd assets/docusaurus && npm run serve

# Clean build artifacts
docs-clean:
    cd assets/docusaurus && npm run clear
    rm -rf assets/docusaurus/build assets/docusaurus/.docusaurus

# Run markdown linting
docs-lint:
    npx markdownlint-cli2 "docs/**/*.md"

# Check for broken links
docs-check-links: docs-build
    npx broken-link-checker http://localhost:3000

# Combined: install, build, preview
docs: docs-install docs-build docs-preview

# Deploy to GitHub Pages
docs-deploy: docs-build
    cd assets/docusaurus && npm run deploy
