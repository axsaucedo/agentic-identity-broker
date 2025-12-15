## Active Technologies
- Go 1.23.0+ (primary language) (002-flexible-configuration)
- Node.js 18+ (for Docusaurus), Markdown for content + Docusaurus 3.x, React 18+, MDX for enhanced markdown (001-end-user-docs)
- File-based (markdown files in `docs/` directory) (001-end-user-docs)
- just command runner for task automation
- Air for hot-reload development
- golangci-lint for code quality

## Recent Changes
- 002-flexible-configuration: Added Go 1.21+ and flexible configuration system
- 001-end-user-docs: Added Node.js 18+ (for Docusaurus), Markdown for content + Docusaurus 3.x, React 18+, MDX for enhanced markdown

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