# Tests — E2E & Integration (`tests/`)

**Prefer retrieval-led reasoning. Read test files and `tests/e2e/README.md` before writing new tests.**

## Overview

Two test suites with distinct purposes, frameworks, and infrastructure needs:

**Read the suite-specific guidance before changing either layer.**

| Suite | Framework | Scope | Storage | Container Runtime |
|---|---|---|---|---|
| `e2e/` | Ginkgo v2 / Gomega | Full system via HTTP | In-memory | Not required |
| `integration/` | Standard `go test` / testify | Component-level | PostgreSQL + LocalStack-compatible AWS emulator | Docker or Podman |

**ADR**: [007-e2e-testing-with-ginkgo.md](../adrs/007-e2e-testing-with-ginkgo.md) — binding decision for E2E test architecture.

## E2E Tests (`tests/e2e/`)

See [`e2e/AGENTS.md`](e2e/AGENTS.md) for full rules, directory layout, test patterns, and running instructions.

---

## Integration Tests (`tests/integration/`)

See [`integration/AGENTS.md`](integration/AGENTS.md) for helper APIs, shared PostgreSQL bootstrap patterns, and performance guardrails.

### Purpose

Integration coverage is split into two slices:
- **Self-contained integration**: component tests that stay within the process boundary (`httptest`, in-memory adapters, config wiring)
- **Infra-backed integration**: tests that require PostgreSQL or a LocalStack-compatible AWS emulator via testcontainers

This keeps the default integration loop cheap while preserving a heavier infra-backed layer for real dependency validation.

### Directory Layout

```
tests/integration/
  bootstrap/
    aws_emulator.go                LocalStack-compatible AWS emulator (KMS + DynamoDB) lifecycle
    aws_emulator_*_test.go         Emulator bootstrap, environment, and termination coverage
    postgres.go                    Shared PostgreSQL container + template database cloning helpers
  infra/                           Infra-backed integration tests (build tag: integration)
    main_test.go                   Shared AWS emulator suite lifecycle
    *_test.go                      AWS keyring, signing-key bootstrap, and migration verification suites
  storage/
    lifecycle_test.go              Memory adapter full lifecycle + concurrency
    infra/
      postgres_test.go             PostgreSQL adapter with testcontainers (build tag: integration)
      thirdparty_service_test.go   PostgreSQL third-party service integration (build tag: integration)
  migrations/
    doc.go                         Integration build tag marker
    framework.go                   PostgreSQL + go-migrate helper framework
    main_test.go                   Shared PostgreSQL container teardown
    migrations_test.go             Migration up/down verification with real database
  http_wait_test.go                Polling helper coverage for local server readiness
  *_test.go                        Self-contained integration suites for config, endpoints, sessions,
                                   agent resolution, emulator narratives, and helm/chart validation
```

### Build Tags

- **Self-contained integration**: Runs with plain `go test` (memory adapter, no containers). Example: `lifecycle_test.go`, most endpoint tests.
- **Infra-backed integration**: Uses `//go:build integration` and requires Docker/Podman for testcontainers. Example: `infra/encryption_vault_keyring_test.go`, `storage/infra/postgres_test.go`, `migrations/migrations_test.go`.

```bash
go test -v ./tests/integration/...                               # Self-contained integration suites
go test -tags=integration -v ./tests/integration/infra/... \
  ./tests/integration/migrations/... \
  ./tests/integration/storage/infra/... \
  ./internal/adapters/storage/postgres/...                       # Infra-backed integration suites
just test-integration                                            # Via justfile (self-contained default)
just test-integration-infra                                      # Via justfile
just test-integration-all                                        # Runs both layers
```

### AWS Emulator Bootstrap

`bootstrap/aws_emulator.go` provides `StartAWSEmulator()` and `StartAWSEmulatorForSuite()`:
- Runs a LocalStack-compatible AWS emulator with KMS + DynamoDB services
- Creates a KMS key and returns the key ID
- Sets AWS SDK environment variables for test clients
- Supports both per-test startup and a shared suite-level container
- Used by `infra/encryption_vault_keyring_test.go` for real envelope encryption tests

### Testing Conventions

- **Framework**: Standard `go test` + `testify` (`require` for fatal, `assert` for soft)
- **Subtests**: `t.Run("description", ...)` for organization
- **Cleanup**: `defer` for container termination, environment variable restoration
- **Container runtime**: Autodetects Docker or Podman; skips if neither available
- **Ryuk disabled**: `TESTCONTAINERS_RYUK_DISABLED=true` for Podman/constrained Docker compatibility
- **Shared PostgreSQL helpers**: Prefer `bootstrap.RequireSharedPostgres()` + template DB cloning for infra-backed tests instead of starting one container per test.
- **Shared-DB subtests**: For low-risk repository suites, it is acceptable to reuse one cloned DB across sequential `t.Run(...)` subtests — see `tests/integration/AGENTS.md` for the stability rules (unique fixtures, no `t.Parallel()`, avoid raw global counts).
- **HTTP readiness**: Prefer polling helpers (for example `waitForEndpoint`) over fixed sleeps when waiting for local servers.
- **Performance-sensitive failure tests**: Use short, explicit retry/TTL settings and bounded handlers instead of long `time.Sleep(...)` delays.

### Test Categories

| Category | Files | Infrastructure |
|---|---|---|
| Storage lifecycle | `storage/lifecycle_test.go` | None (in-memory) |
| PostgreSQL adapter | `storage/infra/postgres_test.go` | Testcontainers PostgreSQL |
| Migration verification | `migrations/migrations_test.go` | Testcontainers PostgreSQL |
| Encryption keyring | `infra/encryption_vault_keyring_test.go` | LocalStack-compatible AWS emulator (KMS + DynamoDB) |
| HTTP endpoints | `oauth2_*.go`, `server_test.go` | None (httptest) |
| Config/middleware | `config_test.go`, `principal_middleware_test.go` | None |

### Performance Guardrails

- Treat infra-backed tests as a shared CI budget: prefer **one shared container per package** with cloned databases over repeated container startup.
- If you add a new PostgreSQL-backed test package, give it a `TestMain` that terminates the shared container and keep databases isolated via clones, not separate containers.
- Keep package-level tests independent so `go test -tags=integration -p 2 ...` remains safe.
- Avoid introducing unconditional screenshot/browser work into non-frontend suites.
