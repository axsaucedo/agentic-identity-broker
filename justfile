# Variable definitions
NAME := "agentic-identity-broker"
IMAGE_NAME := env_var_or_default("IMAGE_NAME", "agentic-identity-broker")
VERSION := `git describe --tags --always 2>/dev/null || echo "latest"`

# Determine container runtime (docker or podman)
# Prefer docker over podman when both are available for better multi-arch support
CONTAINER_RUNTIME := `if [ -n "${CONTAINER_RUNTIME:-}" ]; then echo "$CONTAINER_RUNTIME"; elif command -v docker >/dev/null 2>&1; then echo "docker"; elif command -v podman >/dev/null 2>&1; then echo "podman"; else echo "Error: no container runtime found. Please install podman or docker, or set CONTAINER_RUNTIME." >&2; exit 1; fi`

# Determine compose command (docker compose or podman-compose)
COMPOSE_CMD := `if [ -n "${COMPOSE_CMD:-}" ]; then echo "$COMPOSE_CMD"; elif [ -n "${COMPOSE_TOOL:-}" ]; then echo "$COMPOSE_TOOL"; elif command -v podman-compose >/dev/null 2>&1; then echo "podman-compose"; elif command -v docker >/dev/null 2>&1; then echo "docker compose"; else echo "Error: no compose tool found. Please install podman-compose or Docker, or set COMPOSE_CMD." >&2; exit 1; fi`


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
test-backend-e2e:
    @echo "Running E2E tests..."
    @if command -v ginkgo > /dev/null; then \
        ginkgo -v ./tests/e2e/; \
    else \
        echo "Error: ginkgo is not installed. Install it with: go install github.com/onsi/ginkgo/v2/ginkgo@latest"; \
        exit 1; \
    fi

# Run E2E tests with coverage report
test-backend-e2e-coverage:
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
test-backend-e2e-watch:
    @echo "Starting E2E test watch mode..."
    @if command -v ginkgo > /dev/null; then \
        ginkgo watch -v ./tests/e2e/; \
    else \
        echo "Error: ginkgo is not installed. Install it with: go install github.com/onsi/ginkgo/v2/ginkgo@latest"; \
        exit 1; \
    fi

# Run frontend E2E tests with pre-built frontend (production-like)
test-frontend-e2e:
    #!/usr/bin/env bash
    set -e
    just web-build
    E2E_FRONTEND_MODE=built ginkgo -v ./tests/e2e/frontend/

# Run frontend E2E tests with Vite dev server (hot reload)
# NOTE: Requires 'just web-dev' running in another terminal
test-frontend-e2e-dev:
    #!/usr/bin/env bash
    E2E_FRONTEND_MODE=dev ginkgo -v ./tests/e2e/frontend/

# Run frontend E2E tests with coverage report
test-frontend-e2e-coverage:
    #!/usr/bin/env bash
    set -e
    just web-build
    E2E_FRONTEND_MODE=built ginkgo -v --cover ./tests/e2e/frontend/

# Run all E2E tests: backend + frontend
test-e2e-full:
    #!/usr/bin/env bash
    set -e
    just web-build
    ginkgo -v ./tests/e2e/ --skip="Frontend"
    just test-frontend-e2e

# Build and run the application
run: build
    @echo "Running {{NAME}}..."
    IDENTITY_BROKER_JWE_SIGNING_KEY=`./scripts/generate-jwe-key.sh` ./bin/{{NAME}}

# Run with Air for hot-reload development (requires air to be installed)
dev:
    IDENTITY_BROKER_JWE_SIGNING_KEY=`./scripts/generate-jwe-key.sh`
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
        bash scripts/golangci-lint-install.sh -b /usr/local/bin v2.8.0; \
    else \
        echo "golangci-lint is already installed"; \
    fi
    @echo "Installing go-junit-report for CI/CD test reporting..."
    go install github.com/jstemmer/go-junit-report/v2@v2.1.0
    @echo "Installing ginkgo for E2E testing..."
    go install github.com/onsi/ginkgo/v2/ginkgo@v2.27.3
    @echo "Installing Playwright Go binary..."
    go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5200.1 install --with-deps
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
    go test -tags=integration -v ./tests/integration/storage/...

# Run integration tests and generate JUnit XML report for CI/CD
test-integration-junit:
    @echo "Running integration tests with JUnit output..."
    @mkdir -p test-results
    @go test -tags=integration -v ./tests/integration/storage/... 2>&1 | tee test-results/integration-test-output.txt | go-junit-report -set-exit-code > test-results/integration-junit.xml
    @echo "JUnit report generated at test-results/integration-junit.xml"

# Run all tests (unit, integration, E2E, E2E frontend) and generate consolidated JUnit XML report
test-all-junit:
    #!/usr/bin/env bash
    set +e  # Don't exit on errors; we'll handle them at the end

    echo "Running all tests (unit + integration + E2E) with JUnit output..."
    mkdir -p test-results

    # Initialize exit code tracking
    UNIT_EXIT=0
    INTEGRATION_EXIT=0
    E2E_EXIT=0
    E2E_FRONTEND_EXIT=0
    MERGER_EXIT=0

    # ===== UNIT TESTS =====
    echo ""
    echo "==> Running unit tests (cmd/ and internal/)..."
    if go test -v -race ./cmd/... ./internal/... 2>&1 | tee test-results/unit-tests-output.txt | go-junit-report -set-exit-code > test-results/unit-junit.xml; then
        echo "✓ Unit tests passed"
    else
        UNIT_EXIT=$?
        echo "✗ Unit tests failed (exit code: $UNIT_EXIT)"
    fi

    # ===== INTEGRATION TESTS =====
    echo ""
    echo "==> Running integration tests..."
    if go test -v ./tests/integration/... 2>&1 | tee test-results/integration-tests-output.txt | go-junit-report -set-exit-code > test-results/integration-junit.xml; then
        echo "✓ Integration tests passed"
    else
        INTEGRATION_EXIT=$?
        echo "✗ Integration tests failed (exit code: $INTEGRATION_EXIT)"
    fi

    # Run storage-specific integration tests with PostgreSQL containers
    echo ""
    echo "==> Running storage integration tests (PostgreSQL)..."
    if go test -v -tags=integration ./tests/integration/storage/... 2>&1 | tee test-results/storage-tests-output.txt | go-junit-report -set-exit-code > test-results/storage-junit.xml; then
        echo "✓ Storage integration tests passed"
    else
        STORAGE_EXIT=$?
        echo "✗ Storage integration tests failed (exit code: $STORAGE_EXIT)"
        INTEGRATION_EXIT=$STORAGE_EXIT
    fi

    # ===== E2E TESTS =====
    echo ""
    echo "==> Running E2E tests (Ginkgo)..."
    if ! command -v ginkgo > /dev/null; then
        echo "✗ ginkgo not installed"
        echo "  Run 'just install-tools' to install required development tools"
        E2E_EXIT=1
    else
        if ginkgo run -v --junit-report=test-results/e2e-junit.xml ./tests/e2e/; then
            echo "✓ E2E tests passed"
        else
            E2E_EXIT=$?
            echo "✗ E2E tests failed (exit code: $E2E_EXIT)"
        fi
    fi

    # ===== E2E FRONTEND TESTS =====
    echo ""
    echo "==> Running E2E frontend tests (Ginkgo)..."

    if ! command -v ginkgo > /dev/null; then
        echo "✗ ginkgo not installed"
        echo "  Run 'just install-tools' to install required development tools"
        E2E_FRONTEND_EXIT=1
    else
        if ginkgo run -v --junit-report=test-results/e2e-frontend-junit.xml ./tests/e2e/frontend/; then
            echo "✓ E2E frontend tests passed"
        else
            E2E_FRONTEND_EXIT=$?
            echo "✗ E2E frontend tests failed (exit code: $E2E_FRONTEND_EXIT)"
        fi
    fi

    # ===== MERGE JUNIT REPORTS =====
    echo ""
    echo "==> Merging JUnit reports..."
    if command -v npx > /dev/null; then
        if npx -y junit-report-merger@9.0.3 test-results/all-tests-junit.xml test-results/*-junit.xml; then
            echo "✓ Merged report generated at test-results/all-tests-junit.xml"
        else
            MERGER_EXIT=$?
            echo "⚠ junit-report-merger failed (exit code: $MERGER_EXIT)"
        fi
    else
        echo "⚠ npx not found, junit-report-merger not available"
        echo "  Install Node.js to use junit-report-merger, or reports will not be merged"
    fi

    # ===== FINAL SUMMARY =====
    echo ""
    echo "=== Test Summary ==="
    echo "Unit tests exit code: $UNIT_EXIT"
    echo "Integration tests exit code: $INTEGRATION_EXIT"
    echo "E2E tests exit code: $E2E_EXIT"
    echo "E2E frontend tests exit code: $E2E_FRONTEND_EXIT"
    echo "JUnit XML merger exit code: $MERGER_EXIT"
    echo ""

    # Fail if any test suite failed
    if [ $UNIT_EXIT -ne 0 ] || [ $INTEGRATION_EXIT -ne 0 ] || [ $E2E_EXIT -ne 0 ] || [ $E2E_FRONTEND_EXIT -ne 0 ] || [ $MERGER_EXIT -ne 0 ]; then
        echo "✗ Some tests failed"
        exit 1
    fi

    echo "✓ All tests passed"

# Run all unit and integration tests with coverage
test-all: test test-integration
    @echo "All tests completed"

# Run all tests: unit, integration, and E2E (comprehensive test suite)
test-full: test test-integration test-backend-e2e test-frontend-e2e
    @echo "Full test suite completed"

# Run all quality checks (fmt, vet, lint)
check: fmt vet lint
    @echo "All checks passed!"

# =============================================================================
# Web Development Targets
# =============================================================================

# Install web dependencies
web-install:
    @echo "Installing web dependencies..."
    cd web && npm install

# Install web dependencies for CI with strict engine checking
web-ci:
    @echo "Installing web dependencies for CI..."
    npm config set engine-strict true
    cd web && npm ci

# Start web development server
web-dev:
    @echo "Starting web development server..."
    cd web && npm run dev

# Build web frontend
web-build: web-install
    @echo "Building web frontend..."
    cd web && npm run build

# Run web frontend tests
web-test: web-install
    @echo "Running web frontend tests..."
    cd web && npm test -- --run

# Run web frontend tests with coverage
web-test-coverage: web-install
    @echo "Running web frontend tests with coverage..."
    cd web && npm run test:coverage

# Build both Go backend and web frontend in release quality
# Produces artifacts: ./bin/{{NAME}} and ./web/dist/
build-all: build web-build
    @echo "✓ Build complete: Go backend and web frontend"

# =============================================================================
# Docker Targets
# =============================================================================

# Create and push multi-architecture Docker images to registry
# Builds broker and migrate images for linux/amd64 and linux/arm64 using docker buildx or podman build
# Optional: set BUILDKIT_CONFIG to a buildx config file path (defaults to /etc/cdp-buildkitd.toml if present)
docker-push: build-linux-amd64 build-linux-arm64 web-build
    @echo "Building and pushing multi-architecture images using {{CONTAINER_RUNTIME}}..."
    @if [ "{{CONTAINER_RUNTIME}}" = "docker" ]; then \
        BUILDKIT_CONFIG="$${BUILDKIT_CONFIG:-/etc/cdp-buildkitd.toml}"; \
        if [ -f "$$BUILDKIT_CONFIG" ]; then \
            echo "Using buildx config: $$BUILDKIT_CONFIG"; \
            docker buildx create --config "$$BUILDKIT_CONFIG" --driver-opt network=host --name cdpbuildx --bootstrap --use || true; \
        else \
            echo "Note: buildx config not found at $$BUILDKIT_CONFIG"; \
            docker buildx create --driver-opt network=host --name cdpbuildx --bootstrap --use || true; \
        fi; \
        echo "Building broker image: {{IMAGE_NAME}}:{{VERSION}}..."; \
        docker_buildx build --rm -t "{{IMAGE_NAME}}:{{VERSION}}" --build-arg VERSION="{{VERSION}}" --platform linux/amd64,linux/arm64 --push .; \
        echo "Building migrate image: {{IMAGE_NAME}}-migrate:{{VERSION}}..."; \
        docker_buildx build --rm -t "{{IMAGE_NAME}}-migrate:{{VERSION}}" --build-arg VERSION="{{VERSION}}" --platform linux/amd64,linux/arm64 --file Dockerfile.migrate --push .; \
    else \
        echo "Building broker image: {{IMAGE_NAME}}:{{VERSION}}..."; \
        podman rmi "{{IMAGE_NAME}}:{{VERSION}}" 2>/dev/null || true; \
        podman manifest rm "{{IMAGE_NAME}}:{{VERSION}}" 2>/dev/null || true; \
        podman manifest create "{{IMAGE_NAME}}:{{VERSION}}"; \
        podman build --rm --build-arg VERSION="{{VERSION}}" --platform linux/amd64 --manifest "{{IMAGE_NAME}}:{{VERSION}}" .; \
        podman build --rm --build-arg VERSION="{{VERSION}}" --platform linux/arm64 --manifest "{{IMAGE_NAME}}:{{VERSION}}" .; \
        podman manifest push --all "{{IMAGE_NAME}}:{{VERSION}}" "docker://{{IMAGE_NAME}}:{{VERSION}}"; \
        echo "Building migrate image: {{IMAGE_NAME}}-migrate:{{VERSION}}..."; \
        podman rmi "{{IMAGE_NAME}}-migrate:{{VERSION}}" 2>/dev/null || true; \
        podman manifest rm "{{IMAGE_NAME}}-migrate:{{VERSION}}" 2>/dev/null || true; \
        podman manifest create "{{IMAGE_NAME}}-migrate:{{VERSION}}"; \
        podman build --rm --build-arg VERSION="{{VERSION}}" --platform linux/amd64 --manifest "{{IMAGE_NAME}}-migrate:{{VERSION}}" --file Dockerfile.migrate .; \
        podman build --rm --build-arg VERSION="{{VERSION}}" --platform linux/arm64 --manifest "{{IMAGE_NAME}}-migrate:{{VERSION}}" --file Dockerfile.migrate .; \
        podman manifest push --all "{{IMAGE_NAME}}-migrate:{{VERSION}}" "docker://{{IMAGE_NAME}}-migrate:{{VERSION}}"; \
    fi
    @echo "✓ Multi-architecture images pushed:"
    @echo "  - {{IMAGE_NAME}}:{{VERSION}}"
    @echo "  - {{IMAGE_NAME}}-migrate:{{VERSION}}"

docker-promote:
    @echo "Promoting docker images to production channel..."
    cdp-promote-image {{IMAGE_NAME}}:{{VERSION}}
    cdp-promote-image {{IMAGE_NAME}}-migrate:{{VERSION}}

# Build multi-architecture migrate Docker image (validates both platforms, no output)
docker-build-migrate:
    @echo "Building migrate image: {{IMAGE_NAME}}-migrate:{{VERSION}} using {{CONTAINER_RUNTIME}}..."
    @if [ "{{CONTAINER_RUNTIME}}" = "docker" ]; then \
        BUILDKIT_CONFIG="$${BUILDKIT_CONFIG:-/etc/cdp-buildkitd.toml}"; \
        if [ -f "$$BUILDKIT_CONFIG" ]; then \
            echo "Using buildx config: $$BUILDKIT_CONFIG"; \
            docker buildx create --config "$$BUILDKIT_CONFIG" --driver-opt network=host --name cdpbuildx --bootstrap --use || true; \
        else \
            docker buildx create --driver-opt network=host --name cdpbuildx --bootstrap --use || true; \
        fi; \
        docker_buildx build --rm -t "{{IMAGE_NAME}}-migrate:{{VERSION}}" --build-arg VERSION="{{VERSION}}" --platform linux/amd64,linux/arm64 --file Dockerfile.migrate .; \
    else \
        podman rmi "{{IMAGE_NAME}}-migrate:{{VERSION}}" 2>/dev/null || true; \
        podman manifest rm "{{IMAGE_NAME}}-migrate:{{VERSION}}" 2>/dev/null || true; \
        podman manifest create "{{IMAGE_NAME}}-migrate:{{VERSION}}"; \
        podman build --rm --build-arg VERSION="{{VERSION}}" --platform linux/amd64 --manifest "{{IMAGE_NAME}}-migrate:{{VERSION}}" --file Dockerfile.migrate .; \
        podman build --rm --build-arg VERSION="{{VERSION}}" --platform linux/arm64 --manifest "{{IMAGE_NAME}}-migrate:{{VERSION}}" --file Dockerfile.migrate .; \
    fi
    @echo "✓ Migrate image validated: {{IMAGE_NAME}}-migrate:{{VERSION}}"

# Build multi-architecture broker Docker image (validates both platforms, no output)
docker-build-broker: build-linux-amd64 build-linux-arm64 web-build
    @echo "Building broker image: {{IMAGE_NAME}}:{{VERSION}} using {{CONTAINER_RUNTIME}}..."
    @if [ "{{CONTAINER_RUNTIME}}" = "docker" ]; then \
        BUILDKIT_CONFIG="$${BUILDKIT_CONFIG:-/etc/cdp-buildkitd.toml}"; \
        if [ -f "$$BUILDKIT_CONFIG" ]; then \
            echo "Using buildx config: $$BUILDKIT_CONFIG"; \
            docker buildx create --config "$$BUILDKIT_CONFIG" --driver-opt network=host --name cdpbuildx --bootstrap --use || true; \
        else \
            docker buildx create --driver-opt network=host --name cdpbuildx --bootstrap --use || true; \
        fi; \
        docker_buildx build --rm -t "{{IMAGE_NAME}}:{{VERSION}}" --build-arg VERSION="{{VERSION}}" --platform linux/amd64,linux/arm64 .; \
    else \
        podman rmi "{{IMAGE_NAME}}:{{VERSION}}" 2>/dev/null || true; \
        podman manifest rm "{{IMAGE_NAME}}:{{VERSION}}" 2>/dev/null || true; \
        podman manifest create "{{IMAGE_NAME}}:{{VERSION}}"; \
        podman build --rm --build-arg VERSION="{{VERSION}}" --platform linux/amd64 --manifest "{{IMAGE_NAME}}:{{VERSION}}" .; \
        podman build --rm --build-arg VERSION="{{VERSION}}" --platform linux/arm64 --manifest "{{IMAGE_NAME}}:{{VERSION}}" .; \
    fi
    @echo "✓ Broker image validated: {{IMAGE_NAME}}:{{VERSION}}"

# Build both broker and migrate images (validates both platforms, no output)
docker-build-all: docker-build-broker docker-build-migrate
    @echo "✓ All Docker images validated for linux/amd64,linux/arm64"

# =============================================================================
# Docker Compose - Development (Hot Reload)
# =============================================================================

# Create .env.compose from .env template if it doesn't exist
compose-env:
    @if [ -f .env.compose ]; then \
        echo ".env.compose already exists"; \
    else \
        cp .env .env.compose; \
        echo "✓ Created: .env.compose (customize as needed)"; \
    fi

# Validate docker-compose.yml syntax
compose-validate:
    @echo "Validating docker-compose.yml..."
    @{{COMPOSE_CMD}} -f docker-compose.yml config > /dev/null && echo "✓ Syntax valid" || echo "✗ Syntax error"

# Start all services with logs streaming (foreground)
compose-up: compose-env
    @echo "Generating JWE signing key..."
    @echo "Starting docker-compose services with hot reload..."
    @echo "Services:"
    @echo "  - Identity Broker (8000, 14000) with Air hot reload"
    @echo "  - Frontend (3000) with Vite HMR"
    @echo "  - Upstream OAuth2 (9001)"
    @echo "  - Third-Party OAuth2 (9000)"
    @echo "  - Sample Agent (9002)"
    @echo "  - Seed data will auto-run once broker is healthy"
    @echo ""
    @echo "Press Ctrl+C to stop"
    IDENTITY_BROKER_JWE_SIGNING_KEY=`./scripts/generate-jwe-key.sh` {{COMPOSE_CMD}} -f docker-compose.yml up

# Start all services in background
compose-up-detached: compose-env
    @echo "Generating JWE signing key..."
    @echo "Starting docker-compose services in background..."
    @IDENTITY_BROKER_JWE_SIGNING_KEY=`./scripts/generate-jwe-key.sh` {{COMPOSE_CMD}} -f docker-compose.yml up -d
    @sleep 2
    @just compose-health
    @echo ""
    @echo "Service URLs:"
    @echo "  - Broker (end-user): http://localhost:8000"
    @echo "  - Broker (admin): http://localhost:14000"
    @echo "  - Frontend (consent UI): http://localhost:3000"
    @echo "  - Sample OAuth2 client: http://localhost:9002/oauth2/authorize"
    @echo ""
    @echo "View logs: just compose-logs"
    @echo "Stop services: just compose-down"

# Stop all services
compose-down:
    @echo "Stopping docker-compose services..."
    {{COMPOSE_CMD}} -f docker-compose.yml down

# Stop all services and remove volumes
compose-down-volumes:
    @echo "Stopping docker-compose services and removing volumes..."
    {{COMPOSE_CMD}} -f docker-compose.yml down -v

# View logs from all services (tail -f)
compose-logs:
    {{COMPOSE_CMD}} -f docker-compose.yml logs -f

# View logs from backend only
compose-logs-backend:
    {{COMPOSE_CMD}} -f docker-compose.yml logs -f identity-broker

# View logs from frontend only
compose-logs-frontend:
    {{COMPOSE_CMD}} -f docker-compose.yml logs -f frontend

# View logs from specific service
compose-logs-service SERVICE:
    {{COMPOSE_CMD}} -f docker-compose.yml logs -f {{SERVICE}}

# Restart backend service (after code changes)
compose-restart-backend:
    @echo "Restarting identity-broker service..."
    {{COMPOSE_CMD}} -f docker-compose.yml restart identity-broker

# Restart frontend service (after code changes)
compose-restart-frontend:
    @echo "Restarting frontend service..."
    {{COMPOSE_CMD}} -f docker-compose.yml restart frontend

# Show service status and connectivity
compose-health:
    @echo "Checking service health..."
    @{{COMPOSE_CMD}} -f docker-compose.yml ps
    @echo ""
    @echo "Testing connectivity..."
    @{{COMPOSE_CMD}} -f docker-compose.yml exec -T identity-broker curl -s http://localhost:8000/health && echo "✓ Backend health OK" || echo "✗ Backend not ready"
    @{{COMPOSE_CMD}} -f docker-compose.yml exec -T frontend curl -s http://localhost:3000 > /dev/null && echo "✓ Frontend responding" || echo "✗ Frontend not ready"
    @echo ""

# Clean up: stop containers, remove volumes, clean tmp directories
compose-clean: compose-down-volumes
    @echo "Cleaning up build and temporary directories..."
    @rm -rf tmp/ coverage/ bin/ web/dist web/node_modules
    @echo "✓ Cleanup complete"

# =============================================================================
# Helm Chart Targets
# =============================================================================

# Lint Helm chart for syntax and best practices
helm-lint:
    @echo "Linting Helm chart..."
    @helm lint charts/agentic-identity-broker
    @echo "✓ Helm chart lint passed"

# Validate Helm chart deployment in Kind cluster (full E2E test)
helm-validate:
    @echo "Validating Helm chart in Kind cluster..."
    @./scripts/validate-helm-chart.sh

# Render Helm templates (dry-run)
helm-template:
    @echo "Rendering Helm templates..."
    @helm template broker ./charts/agentic-identity-broker

# Render Helm templates with custom values
helm-template-values VALUES_FILE:
    @echo "Rendering Helm templates with {{VALUES_FILE}}..."
    @helm template broker ./charts/agentic-identity-broker -f {{VALUES_FILE}}

# Install Helm chart to local Kind cluster
helm-install-kind RELEASE_NAME="broker":
    @echo "Installing Helm chart to Kind cluster..."
    @kind create cluster --name helm-test 2>/dev/null || echo "Kind cluster already exists"
    @helm install {{RELEASE_NAME}} ./charts/agentic-identity-broker --wait
    @echo "✓ Chart installed as {{RELEASE_NAME}}"
    @echo ""
    @echo "Check status: kubectl get pods"
    @echo "Uninstall: helm uninstall {{RELEASE_NAME}}"

# Uninstall Helm chart from Kind cluster
helm-uninstall-kind RELEASE_NAME="broker":
    @echo "Uninstalling Helm chart..."
    @helm uninstall {{RELEASE_NAME}}
    @echo "✓ Chart uninstalled"

# Delete Kind test cluster
helm-kind-delete:
    @echo "Deleting Kind test cluster..."
    @kind delete cluster --name helm-test
    @echo "✓ Kind cluster deleted"

# Package Helm chart for distribution
helm-package:
    @echo "Packaging Helm chart..."
    @mkdir -p dist
    @helm package charts/agentic-identity-broker -d dist
    @echo "✓ Chart packaged to dist/"

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
    @bash scripts/register-mock-thirdparty-service.sh

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
