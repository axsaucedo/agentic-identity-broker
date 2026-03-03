# agentic-identity-broker Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-01-16

## Active Technologies
- Helm 3.x (Go templates), Dockerfile modifications + Kubernetes 1.25+, golang-migrate, Zalando PostgreSQL Operator (optional) (014-helm-charts)
- PostgreSQL 12+ (optional, in-memory fallback) (014-helm-charts)
- Go 1.24.0+ (matching project go.mod) + google.golang.org/grpc, envoyproxy/go-control-plane (ExtProc v3), spf13/cobra, spf13/viper, golang.org/x/sync/singlefligh (015-extproc-token-exchange)
- In-memory (sync.RWMutex + map) — no database required (015-extproc-token-exchange)

- Go 1.24.0 (per go.mod) (013-token-exchange)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.24.0 (per go.mod)

## Code Style

Go 1.24.0 (per go.mod): Follow standard conventions

## Recent Changes
- 015-extproc-token-exchange: Added Go 1.24.0+ (matching project go.mod) + google.golang.org/grpc, envoyproxy/go-control-plane (ExtProc v3), spf13/cobra, spf13/viper, golang.org/x/sync/singlefligh
- 014-helm-charts: Added Helm 3.x (Go templates), Dockerfile modifications + Kubernetes 1.25+, golang-migrate, Zalando PostgreSQL Operator (optional)
- 014-helm-charts: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
