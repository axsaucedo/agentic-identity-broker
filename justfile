# Variable definitions
NAME := "agentic-identity-broker"
VERSION := `git describe --tags --always 2>/dev/null || echo "latest"`


# Default recipe (shown when running `just` with no args)
default:
    @just --list

# =============================================================================
# Go Development Targets
# =============================================================================

# Build the Go binary for the host OS/ARCH
build:
    @echo "Building {{NAME}}..."
    @mkdir -p bin
    go build -ldflags="-s -w" -o bin/{{NAME}} ./cmd/{{NAME}}
    @echo "✓ Built: bin/{{NAME}}"

# Build Linux binary for arm64
build-linux-arm64:
    @echo "Building Linux binary for arm64..."
    @mkdir -p bin/linux/arm64
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/linux/arm64/{{NAME}} ./cmd/{{NAME}}
    @echo "✓ Built: bin/linux/arm64/{{NAME}}"

# Build Linux binary for amd64
build-linux-amd64:
    @echo "Building Linux binary for amd64..."
    @mkdir -p bin/linux/amd64
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/linux/amd64/{{NAME}} ./cmd/{{NAME}}
    @echo "✓ Built: bin/linux/amd64/{{NAME}}"

# Build macOS binary for arm64 (Apple Silicon)
build-darwin-arm64:
    @echo "Building macOS binary for arm64..."
    @mkdir -p bin/darwin/arm64
    GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/darwin/arm64/{{NAME}} ./cmd/{{NAME}}
    @echo "✓ Built: bin/darwin/arm64/{{NAME}}"

# Build Windows binary for amd64
build-windows-amd64:
    @echo "Building Windows binary for amd64..."
    @mkdir -p bin/windows/amd64
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/windows/amd64/{{NAME}}.exe ./cmd/{{NAME}}
    @echo "✓ Built: bin/windows/amd64/{{NAME}}.exe"

# Run all Go tests with verbose output
test:
    @echo "Running Go tests..."
    go test -v -race ./...

# Run all Go tests and generate JUnit XML report for CI/CD
test-junit:
    @echo "Running Go tests with JUnit output..."
    @mkdir -p test-results
    @go test -v -race ./... 2>&1 | tee test-results/go-test-output.txt | go-junit-report -set-exit-code > test-results/junit.xml
    @echo "JUnit report generated at test-results/junit.xml"

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

# Run end-to-end tests with Ginkgo
test-e2e:
    @echo "Running E2E tests..."
    @if command -v ginkgo > /dev/null; then \
        ginkgo -v ./tests/e2e/; \
    else \
        echo "Error: ginkgo is not installed. Install it with: go install github.com/onsi/ginkgo/v2/ginkgo@latest"; \
        exit 1; \
    fi

# Run E2E tests with coverage report
test-e2e-coverage:
    @echo "Running E2E tests with coverage..."
    @mkdir -p coverage
    @if command -v ginkgo > /dev/null; then \
        ginkgo -v --cover --coverprofile=e2e.out --output-dir=coverage ./tests/e2e/; \
        go tool cover -html=coverage/e2e.out -o coverage/e2e.html; \
        echo "E2E coverage report generated at coverage/e2e.html"; \
    else \
        echo "Error: ginkgo is not installed. Install it with: go install github.com/onsi/ginkgo/v2/ginkgo@latest"; \
        exit 1; \
    fi

# Watch E2E tests during development (auto-rerun on changes)
test-e2e-watch:
    @echo "Starting E2E test watch mode..."
    @if command -v ginkgo > /dev/null; then \
        ginkgo watch -v ./tests/e2e/; \
    else \
        echo "Error: ginkgo is not installed. Install it with: go install github.com/onsi/ginkgo/v2/ginkgo@latest"; \
        exit 1; \
    fi

# Build and run the application
run: build
    @echo "Running {{NAME}}..."
    ./bin/{{NAME}}

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
    rm -rf build
    rm -rf coverage
    rm -rf test-results
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

# Install development tools (air, golangci-lint, go-junit-report)
install-tools:
    @echo "Installing development tools..."
    @echo "Installing air for hot reload..."
    go install github.com/air-verse/air@v1.63.6
    @echo "Installing golangci-lint..."
    @if ! command -v golangci-lint > /dev/null; then \
        bash scripts/golangci-lint-install.sh -b $(shell go env GOPATH)/bin v2.8.0; \
    else \
        echo "golangci-lint is already installed"; \
    fi
    @echo "Installing go-junit-report for CI/CD test reporting..."
    go install github.com/jstemmer/go-junit-report/v2@v2.1.0
    @echo "Tools installation complete"

# Setup git hooks for quality checks
setup-hooks:
    @echo "Setting up git hooks..."
    git config core.hooksPath .githooks
    @echo "✓ Git hooks configured to use .githooks directory"
    @echo "Pre-commit hook will run: fmt, vet, lint"

# Run integration tests (requires Docker for PostgreSQL tests)
test-integration:
    @echo "Running integration tests..."
    go test -tags=integration -v ./test/integration/storage/...

# Run integration tests and generate JUnit XML report for CI/CD
test-integration-junit:
    @echo "Running integration tests with JUnit output..."
    @mkdir -p test-results
    @go test -tags=integration -v ./test/integration/storage/... 2>&1 | tee test-results/integration-test-output.txt | go-junit-report -set-exit-code > test-results/integration-junit.xml
    @echo "JUnit report generated at test-results/integration-junit.xml"

# Run all tests (unit + integration) and generate single JUnit XML report for CI/CD
test-all-junit:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Running all tests (unit + integration) with JUnit output..."
    mkdir -p test-results
    echo ""
    echo "==> Running unit tests..."
    go test -v -race ./... 2>&1 | tee test-results/unit-tests-output.txt
    UNIT_EXIT=${PIPESTATUS[0]}
    echo ""
    echo "==> Running integration tests..."
    go test -v -tags=integration ./test/integration/storage/... 2>&1 | tee test-results/integration-tests-output.txt
    INTEGRATION_EXIT=${PIPESTATUS[0]}
    echo ""
    echo "==> Generating JUnit report..."
    if command -v go-junit-report > /dev/null; then
        cat test-results/unit-tests-output.txt test-results/integration-tests-output.txt | go-junit-report > test-results/all-tests-junit.xml
        echo "✓ JUnit report generated at test-results/all-tests-junit.xml"
    else
        echo "Warning: go-junit-report not found. Run 'just install-tools' to install it."
        echo "✗ JUnit report not generated"
    fi
    echo ""
    if [ $UNIT_EXIT -ne 0 ] || [ $INTEGRATION_EXIT -ne 0 ]; then
        echo "✗ Tests failed (unit exit: $UNIT_EXIT, integration exit: $INTEGRATION_EXIT)"
        exit 1
    fi
    echo "✓ All tests passed"

# Run all unit and integration tests with coverage
test-all: test test-integration
    @echo "All tests completed"

# Run all tests: unit, integration, and E2E (comprehensive test suite)
test-full: test test-integration test-e2e
    @echo "Full test suite completed"

# Run all quality checks (fmt, vet, lint, unit tests)
check: fmt vet lint test
    @echo "All checks passed!"

# Run comprehensive checks: formatting, linting, and all tests (unit, integration, E2E)
check-full: fmt vet lint test test-integration test-e2e
    @echo "Comprehensive checks passed!"

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
web-build: web-install
    @echo "Building web frontend..."
    cd web && npm run build

# Build both Go backend and web frontend in release quality
# Produces artifacts: ./bin/{{NAME}} and ./web/dist/
build-all: build web-build
    @echo "✓ Build complete: Go backend and web frontend"

# =============================================================================
# Docker Targets
# =============================================================================

# Create and push multi-architecture Docker images to registry
# Builds for linux/amd64 and linux/arm64 using docker buildx
# Optional: set BUILDKIT_CONFIG to a buildx config file path (defaults to /etc/cdp-buildkitd.toml if present) 
docker-push: build-linux-amd64 build-linux-arm64 web-build
    @echo "Building and pushing multi-architecture Docker images: {{NAME}}:{{VERSION}} (amd64, arm64)..."
    @BUILDKIT_CONFIG="$${BUILDKIT_CONFIG:-/etc/cdp-buildkitd.toml}"; \
    if [ -f "$$BUILDKIT_CONFIG" ]; then \
        echo "Using buildx config: $$BUILDKIT_CONFIG"; \
        docker buildx create --config "$$BUILDKIT_CONFIG" --driver-opt network=host --bootstrap --use 2>/dev/null || true; \
    else \
        echo "Note: buildx config not found at $$BUILDKIT_CONFIG"; \
        docker buildx create --driver-opt network=host --bootstrap --use 2>/dev/null || true; \
    fi; \
    docker buildx build --rm -t "{{NAME}}:{{VERSION}}" --build-arg VERSION="{{VERSION}}" --platform linux/amd64,linux/arm64 --push .
    @echo "✓ Multi-architecture images pushed: {{NAME}}:{{VERSION}}"

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

# =============================================================================
# Mock Third-Party OAuth2 Service Targets (Manual Testing)
# =============================================================================

# Start mock third-party OAuth2 service (port 9000)
mock-third-party-oauth2-start:
    @echo "Starting mock third-party OAuth2 service..."
    cd mocks/third-party-service && go run cmd/mock-oauth2-server/main.go

# Build mock third-party OAuth2 service binary
mock-third-party-oauth2-build:
    @echo "Building mock third-party OAuth2 service..."
    @mkdir -p bin
    cd mocks/third-party-service && go build -o ../../bin/mock-oauth2-server cmd/mock-oauth2-server/main.go
    @echo "✓ Binary built: ./bin/mock-oauth2-server"

# Register mock third-party service with broker admin API
mock-third-party-oauth2-register:
    @echo "Registering mock third-party OAuth2 service with broker..."
    @bash mocks/third-party-service/scripts/register-with-broker.sh

# Full setup: build, start (background), register
mock-third-party-oauth2-setup: mock-third-party-oauth2-build
    @echo "Setting up mock third-party OAuth2 testing environment..."
    @echo "1. Starting mock third-party OAuth2 server (background)..."
    @./bin/mock-oauth2-server &
    @sleep 2
    @echo "2. Checking mock server health..."
    @curl -s -f http://localhost:9000/health || (echo "Mock server failed to start"; exit 1)
    @echo "   ✓ Mock server healthy"
    @echo "3. Registering mock service with broker..."
    @just mock-third-party-oauth2-register
    @echo ""
    @echo "========================================="
    @echo "Mock Third-Party OAuth2 Setup Complete!"
    @echo "========================================="
    @echo "Mock OAuth2 Server: http://localhost:9000"
    @echo "Broker Consent UI: http://localhost:8000/consent/sessions"
    @echo ""
    @echo "To stop: pkill -f mock-oauth2-server"

# Stop and clean mock third-party OAuth2 artifacts
mock-third-party-oauth2-clean:
    @echo "Stopping mock third-party OAuth2 server..."
    @pkill -f mock-oauth2-server || true
    @rm -f bin/mock-oauth2-server
    @echo "✓ Mock cleanup complete"

# =============================================================================
# Mock Upstream OAuth2 Server Targets (Manual Testing)
# =============================================================================

# Start mock upstream OAuth2 server (port 9001) - visually distinctive
mock-upstream-oauth2-start:
    @echo "Starting mock upstream OAuth2 server..."
    cd mocks/upstream-oauth2-server && go run cmd/mock-upstream-oauth2-server/main.go

# Build mock upstream OAuth2 server binary
mock-upstream-oauth2-build:
    @echo "Building mock upstream OAuth2 server..."
    @mkdir -p bin
    cd mocks/upstream-oauth2-server && go build -o ../../bin/mock-upstream-oauth2-server cmd/mock-upstream-oauth2-server/main.go
    @echo "✓ Binary built: ./bin/mock-upstream-oauth2-server"

# Run tests for upstream OAuth2 mock server
mock-upstream-oauth2-test:
    @echo "Running tests for upstream OAuth2 mock server..."
    cd mocks/upstream-oauth2-server && go test -v ./internal/handlers/...

# Full setup: build, start (background), health check
mock-upstream-oauth2-setup: mock-upstream-oauth2-build
    @echo "Setting up mock upstream OAuth2 testing environment..."
    @echo "1. Starting mock upstream OAuth2 server (background)..."
    @./bin/mock-upstream-oauth2-server &
    @sleep 2
    @echo "2. Checking mock server health..."
    @curl -s -f http://localhost:9001/health || (echo "Mock server failed to start"; exit 1)
    @echo "   ✓ Mock server healthy"
    @echo ""
    @echo "========================================="
    @echo "Mock Upstream OAuth2 Setup Complete!"
    @echo "========================================="
    @echo "Mock Upstream OAuth2 Server: http://localhost:9001"
    @echo "Consent Page: http://localhost:9001/oauth/authorize?client_id=upstream-oauth2-client&response_type=code&redirect_uri=http://localhost/callback&scope=openid&state=test123"
    @echo ""
    @echo "Note: You'll see DISTINCTIVE TEAL background (#00d4aa)"
    @echo "      with 🌐 UPSTREAM OAUTH2 badge"
    @echo ""
    @echo "To stop: pkill -f mock-upstream-oauth2-server"

# Stop and clean mock upstream OAuth2 artifacts
mock-upstream-oauth2-clean:
    @echo "Stopping mock upstream OAuth2 server..."
    @pkill -f mock-upstream-oauth2-server || true
    @rm -f bin/mock-upstream-oauth2-server
    @echo "✓ Mock cleanup complete"

# =============================================================================
# Mock Sample OAuth2 Client Agent Targets (End-to-End Testing)
# =============================================================================

# Start mock sample OAuth2 client agent (port 8001)
mock-sample-agent-start:
    @echo "Starting mock sample OAuth2 client agent..."
    cd mocks/sample-agent && go run cmd/sample-agent/main.go

# Build mock sample OAuth2 client agent binary
mock-sample-agent-build:
    @echo "Building mock sample OAuth2 client agent..."
    @mkdir -p bin
    cd mocks/sample-agent && go build -o ../../bin/sample-agent cmd/sample-agent/main.go
    @echo "✓ Binary built: ./bin/sample-agent"

# Run tests for sample agent
mock-sample-agent-test:
    @echo "Running tests for sample agent..."
    cd mocks/sample-agent && go test -v ./...

# Full setup: build, start (background), health check
mock-sample-agent-setup: mock-sample-agent-build
    @echo "Setting up mock sample OAuth2 client testing environment..."
    @echo "1. Starting mock sample OAuth2 client (background)..."
    @./bin/sample-agent &
    @sleep 2
    @echo "2. Checking mock client health..."
    @curl -s -f http://localhost:8001/health || (echo "Sample client failed to start"; exit 1)
    @echo "   ✓ Sample client healthy"
    @echo ""
    @echo "========================================="
    @echo "Sample OAuth2 Client Setup Complete!"
    @echo "========================================="
    @echo ""
    @echo "Three-Tier OAuth2 Setup:"
    @echo "  Tier 1: Sample Agent ................ http://localhost:8001"
    @echo "  Tier 2: Identity Broker ............ http://localhost:8000"
    @echo "  Tier 3: Upstream OAuth2 Server .... http://localhost:9001"
    @echo ""
    @echo "Testing the complete end-to-end flow:"
    @echo "  1. Start this: just mock-sample-agent-setup"
    @echo "  2. Start broker: just dev"
    @echo "  3. Start upstream: just mock-upstream-oauth2-start"
    @echo "  4. Browser: http://localhost:8001"
    @echo "  5. Click Login with OAuth2"
    @echo "  6. Approve consent (notice TEAL background on port 9001)"
    @echo "  7. See user information page"
    @echo ""
    @echo "To stop: pkill -f 'sample-agent|bin/sample-agent'"

# Stop and clean mock sample agent artifacts
mock-sample-agent-clean:
    @echo "Stopping mock sample OAuth2 client agent..."
    @pkill -f 'sample-agent|bin/sample-agent' || true
    @rm -f bin/sample-agent
    @echo "✓ Mock cleanup complete"
