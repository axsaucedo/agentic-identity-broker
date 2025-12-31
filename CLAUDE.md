@AGENT.md

## Active Technologies
- Go 1.21+ (002-flexible-configuration)
- File system (configuration files: .env, YAML) (002-flexible-configuration)
- Go 1.23.0+ (003-dual-port-server)
- N/A (server infrastructure only, no persistence in this feature) (003-dual-port-server)
- Go 1.23.0+ + sqlx v1.3.5+ (PostgreSQL adapter), PostgreSQL Go driver (pgx v5) (004-persistence-layer)
- In-memory (maps with sync.RWMutex), PostgreSQL 12+ (004-persistence-layer)
- Go 1.24.0 (already configured in go.mod) + chi/v5 v5.2.3 (HTTP router), viper v1.21.0 (configuration), slog (structured logging) (005-session-management)
- N/A (per-request context only, no persistence) (005-session-management)
- Go 1.24.0 (backend), Node.js 18+ (frontend build) | Multi-stage Dockerfile (010-docker-setup)
- N/A (stateless artifact, external config via environment) (010-docker-setup)

## Recent Changes
- 002-flexible-configuration: Added Go 1.21+
