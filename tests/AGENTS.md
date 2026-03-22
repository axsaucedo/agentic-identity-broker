# Tests — E2E & Integration (`tests/`)

**Prefer retrieval-led reasoning. Read test files and `tests/e2e/README.md` before writing new tests.**

## Overview

Two test suites with distinct purposes, frameworks, and infrastructure needs:

| Suite | Framework | Scope | Storage | Container Runtime |
|---|---|---|---|---|
| `e2e/` | Ginkgo v2 / Gomega | Full system via HTTP | In-memory | Not required |
| `integration/` | Standard `go test` / testify | Component-level | PostgreSQL + LocalStack | Docker or Podman |

**ADR**: [007-e2e-testing-with-ginkgo.md](../adrs/007-e2e-testing-with-ginkgo.md) — binding decision for E2E test architecture.

## E2E Tests (`tests/e2e/`)

See [`e2e/AGENTS.md`](e2e/AGENTS.md) for full rules, directory layout, test patterns, and running instructions.

---

## Integration Tests (`tests/integration/`)

### Purpose

Component-level tests that exercise real infrastructure (PostgreSQL, LocalStack KMS/DynamoDB) via testcontainers. Tests validate adapter behavior with actual external dependencies.

### Directory Layout

```
tests/integration/
  bootstrap/
    localstack.go                  LocalStack container (KMS + DynamoDB) lifecycle
  storage/
    lifecycle_test.go              Memory adapter full lifecycle + concurrency
    postgres_test.go               PostgreSQL adapter with testcontainers (build tag: integration)
  migrations/
    framework.go                   Testcontainers PostgreSQL + go-migrate framework
    migrations_test.go             Migration up/down verification with real database
  config_test.go                   Configuration loading integration tests
  server_test.go                   Server startup, port connectivity, atomic failure
  health_test.go                   Health check endpoint integration
  principal_middleware_test.go     X-Remote-User middleware integration
  oauth2_authorize_test.go         Authorization endpoint integration
  oauth2_metadata_test.go          Metadata endpoint integration
  oauth2_token_test.go             Token endpoint integration
  oauth2_sessions_api_test.go      OAuth2 session management API
  user_sessions_test.go            User session lifecycle
  encryption_vault_keyring_test.go AWS KMS hierarchical keyring with LocalStack
  agent_repository_service_requirements_test.go  Agent service requirements (memory + postgres)
  agent_service_requirements_migration_test.go   Migration verification for service requirements
```

### Build Tags

- **No tag**: Tests run with `go test` (memory adapter, no containers). Example: `lifecycle_test.go`, most endpoint tests.
- **`integration` tag**: Requires Docker/Podman for testcontainers. Example: `postgres_test.go`, `migrations_test.go`.

```bash
go test -v ./tests/integration/...                        # Non-container tests only
go test -tags=integration -v ./tests/integration/...      # All including container tests
just test-integration                                      # Via justfile
```

### LocalStack Bootstrap

`bootstrap/localstack.go` provides `StartLocalStack()`:
- Runs LocalStack container with KMS + DynamoDB services
- Creates a KMS key and returns the key ID
- Sets AWS SDK environment variables for test clients
- Used by `encryption_vault_keyring_test.go` for real envelope encryption tests

### Testing Conventions

- **Framework**: Standard `go test` + `testify` (`require` for fatal, `assert` for soft)
- **Subtests**: `t.Run("description", ...)` for organization
- **Cleanup**: `defer` for container termination, environment variable restoration
- **Container runtime**: Autodetects Docker or Podman; skips if neither available
- **Ryuk disabled**: `TESTCONTAINERS_RYUK_DISABLED=true` for Podman/constrained Docker compatibility

### Test Categories

| Category | Files | Infrastructure |
|---|---|---|
| Storage lifecycle | `storage/lifecycle_test.go` | None (in-memory) |
| PostgreSQL adapter | `storage/postgres_test.go` | Testcontainers PostgreSQL |
| Migration verification | `migrations/migrations_test.go` | Testcontainers PostgreSQL |
| Encryption keyring | `encryption_vault_keyring_test.go` | LocalStack (KMS + DynamoDB) |
| HTTP endpoints | `oauth2_*.go`, `server_test.go` | None (httptest) |
| Config/middleware | `config_test.go`, `principal_middleware_test.go` | None |
