## Active Technologies
- Go 1.23.0+ (primary language) (002-flexible-configuration)
- Node.js 18+ (for Docusaurus), Markdown for content + Docusaurus 3.x, React 18+, MDX for enhanced markdown (001-end-user-docs)
- File-based (markdown files in `docs/` directory) (001-end-user-docs)
- sqlx v1.3.5+ (PostgreSQL adapter) (004-persistence-layer)
- pgx v5 (PostgreSQL driver) (004-persistence-layer)
- In-memory (maps with sync.RWMutex), PostgreSQL 12+ (004-persistence-layer)
- just command runner for task automation
- Air for hot-reload development
- golangci-lint for code quality

## Recent Changes
- 002-flexible-configuration: Added Go 1.21+ and flexible configuration system
- 001-end-user-docs: Added Node.js 18+ (for Docusaurus), Markdown for content + Docusaurus 3.x, React 18+, MDX for enhanced markdown
- 004-persistence-layer: Added sqlx v1.3.5+, pgx v5, in-memory and PostgreSQL storage layer

## Development Workflow

### Building and Running
Use the justfile targets for all development tasks. Run `just --list` to see all available commands.

**Core Commands:**
- `just build` - Build the Go binary to ./bin/identity-broker
- `just build-release` - Build optimized binary for production (30% smaller, no debug info)
- `just run` - Build and run the application
- `just dev` - Start hot-reload development server with Air (auto-rebuilds on file changes)
- `just clean` - Remove build artifacts (./bin and ./coverage directories)

### Testing
- `just test` - Run all tests with verbose output and race detection
- `just test-coverage` - Generate HTML coverage report at coverage/coverage.html
- `just test-coverage-summary` - Display coverage percentages in terminal

### Code Quality
Before committing code, ensure quality checks pass:
- `just fmt` - Format code with gofmt (applies -s for simplification)
- `just vet` - Run go vet for static analysis
- `just lint` - Run golangci-lint (falls back to go vet if not installed)
- `just check` - Run all quality checks: fmt, vet, lint, test

### Dependencies
- `just deps` - Run go mod tidy, download, and verify
- `just install-tools` - Install Air and golangci-lint

### Documentation
- `just docs-serve` - Start documentation server locally with hot reload
- `just docs-build` - Build documentation site
- `just docs-deploy` - Deploy documentation to GitHub Pages

### Frontend Development (Consent UI)

The project includes a React-based consent frontend in the `web/` directory.

**Quick Start (Recommended Workflow):**

```bash
# Terminal 1: Start Go backend
just run

# Terminal 2: Start frontend dev server with hot reload
just web-dev
```

Access frontend at http://localhost:3000 (Vite dev server)
API requests automatically proxy to http://localhost:8000 (Go backend)

**Web Commands:**
- `just web-install` - Install npm dependencies (run once after cloning)
- `just web-dev` - Start Vite dev server with HMR on port 3000
- `just web-build` - Build production bundle to web/dist/consent/
- `just build-all` - Build both Go backend and frontend

**How the Vite Proxy Works:**

The Vite dev server (port 3000) proxies API requests to the Go backend (port 8000):

```
Frontend (port 3000) → Vite Proxy (adds X-Remote-User: dev@example.com) → Backend (port 8000) → API Response
```

The proxy automatically injects the `X-Remote-User` header to simulate authentication during development. In production, this header is set by the upstream authentication proxy (oauth2-proxy, nginx, etc.).

This provides:
- Hot Module Replacement (HMR) for instant frontend updates
- No CORS issues (same-origin requests from browser perspective)
- Full access to backend API during development
- Simulated authentication via X-Remote-User header injection (all requests appear as dev@example.com)
- React DevTools for debugging

**Alternative Workflow (No HMR):**

```bash
# Build frontend and run from Go backend
just web-build && just run

# Access at http://localhost:8000/consent
```

## Pre-commit Checklist

Before committing code, run:
```bash
just check
```

This runs:
1. `just fmt` - Format code
2. `just vet` - Static analysis
3. `just lint` - Linting
4. `just test` - All tests with race detection

All checks must pass before submitting a PR.