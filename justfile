# Variable definitions
IMAGE := "identity-broker"
VERSION := `git describe --tags --always 2>/dev/null || echo "latest"`

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

# Build static Linux binary for amd64
build-linux-static:
    @echo "Building static Linux binary for amd64..."
    @mkdir -p build/linux/static
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-extldflags=-static -s -w" -o build/linux/static/identity-broker ./cmd/identity-broker
    @echo "✓ Built: build/linux/static/identity-broker"

# Build Linux binary for arm64
build-linux-arm64:
    @echo "Building Linux binary for arm64..."
    @mkdir -p build/linux/arm64
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/linux/arm64/identity-broker ./cmd/identity-broker
    @echo "✓ Built: build/linux/arm64/identity-broker"

# Build Linux binary for amd64
build-linux-amd64:
    @echo "Building Linux binary for amd64..."
    @mkdir -p build/linux/amd64
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/linux/amd64/identity-broker ./cmd/identity-broker
    @echo "✓ Built: build/linux/amd64/identity-broker"

# Build macOS binary for arm64 (Apple Silicon)
build-darwin-arm64:
    @echo "Building macOS binary for arm64..."
    @mkdir -p build/darwin/arm64
    GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/darwin/arm64/identity-broker ./cmd/identity-broker
    @echo "✓ Built: build/darwin/arm64/identity-broker"

# Build Windows binary for amd64
build-windows-amd64:
    @echo "Building Windows binary for amd64..."
    @mkdir -p build/windows/amd64
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/windows/amd64/identity-broker.exe ./cmd/identity-broker
    @echo "✓ Built: build/windows/amd64/identity-broker.exe"

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
    rm -rf web/node_modules web/dist
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

# Run integration tests (requires Docker for PostgreSQL tests)
test-integration:
    @echo "Running integration tests..."
    go test -tags=integration -v ./test/integration/storage/...

# Run all unit tests with coverage
test-all: test test-integration
    @echo "All tests completed"

# Run all quality checks (fmt, vet, lint, test)
check: fmt vet lint test
    @echo "All checks passed!"

# =============================================================================
# Web Development Targets
# =============================================================================

# Install web dependencies
web-install:
    @echo "Installing web dependencies..."
    cd web && npm install

# Start web development server
web-dev:
    @echo "Starting web development server..."
    cd web && npm run dev

# Build web frontend
web-build:
    @echo "Building web frontend..."
    cd web && npm run build

# Build both Go backend and web frontend in release quality
# Produces artifacts for Docker build: ./bin/identity-broker and ./web/dist/
build-all: build-release web-build
    @echo "✓ Build complete: Go backend (release) and web frontend"

# =============================================================================
# Docker Targets
# =============================================================================

# Build Docker image from pre-built artifacts
# Requires: ./bin/identity-broker and ./web/dist/ to exist
docker-build:
    @echo "Building Docker image..."
    @if ! command -v docker > /dev/null; then \
        echo "Error: Docker is not installed. Please install Docker or Docker Desktop."; \
        exit 1; \
    fi
    @if [ ! -f ./bin/identity-broker ]; then \
        echo "Error: Backend binary not found at ./bin/identity-broker"; \
        echo "Run 'just build-all' to build artifacts first."; \
        exit 1; \
    fi
    @if [ ! -d ./web/dist ]; then \
        echo "Error: Frontend assets not found at ./web/dist/"; \
        echo "Run 'just build-all' to build artifacts first."; \
        exit 1; \
    fi
    docker build -t identity-broker:latest .
    @echo "✓ Docker image built: identity-broker:latest"
    docker images | grep identity-broker

# Run Docker image locally
# Exposes backend on port 8000 (customizable with DOCKER_PORT environment variable)
docker-run:
    @echo "Starting Docker container..."
    @if ! command -v docker > /dev/null; then \
        echo "Error: Docker is not installed. Please install Docker or Docker Desktop."; \
        exit 1; \
    fi
    @PORT=$${DOCKER_PORT:-8000}; \
    echo "Starting identity-broker on http://localhost:$$PORT"; \
    docker run -p $$PORT:8000 identity-broker:latest

# Create and push multi-architecture Docker images to registry
# Builds for linux/amd64 and linux/arm64 using docker buildx
# Requires: /etc/cdp-buildkitd.toml configuration for buildx
push-multiarch:
    @echo "Building and pushing multi-architecture Docker images: $(IMAGE):$(VERSION) (amd64, arm64)..."
    @if ! command -v docker > /dev/null; then \
        echo "Error: Docker is not installed."; \
        exit 1; \
    fi
    docker buildx create --config /etc/cdp-buildkitd.toml --driver-opt network=host --bootstrap --use
    docker buildx build --rm -t "$(IMAGE):$(VERSION)" --platform linux/amd64,linux/arm64 --push .
    @echo "✓ Multi-architecture images pushed: $(IMAGE):$(VERSION)"

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
