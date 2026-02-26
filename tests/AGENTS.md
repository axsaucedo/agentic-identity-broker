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

### Constitution Principle XIII — Mandatory Rules

1. **1:1 spec-to-test mapping**: Every `It()` block maps to ONE acceptance scenario from `specs/[NNN-feature]/spec.md`.
2. **Spec traceability**: Every `It()` block MUST have a comment linking to the specific scenario: `// Scenario X.Y from specs/NNN-feature/spec.md`
3. **Production bootstrap**: Tests use `app.Builder` and production DI — no custom test implementations.

### Directory Layout

```
tests/e2e/
  e2e_suite_test.go               Ginkgo suite entrypoint
  bootstrap/                       Server + storage initialization (volatile layer)
    test_server.go                 TestServer wrapper — AuthenticatedGET/POST, PublicGET
    server_factory.go              Builds production app.App via DI
    storage.go                     In-memory storage initialization
    playwright.go                  Playwright browser lifecycle for frontend tests
  fixtures/                        Deterministic test data factories
    agents.go                      ValidAgent(), AgentWithClientID(), AgentWithURLs()
    grants.go                      ActiveGrant(), ExpiredGrant(), IndefiniteGrant()
    principals.go                  DefaultPrincipal(), AnotherPrincipal(), AdminPrincipal()
    config.go                      DefaultOAuth2Config(), OAuth2ConfigWithUpstream()
    services.go                    Third-party OAuth2 service fixtures
    sessions.go                    User session fixtures
    encryption.go                  Encryption-related fixtures
  helpers/                         Test utilities
    http_helpers.go                Response body parsing, status assertions
    jwt_helpers.go                 JWT creation/validation for tests
    mock_upstream.go               Mock OAuth2 upstream server
  matchers/                        Custom Gomega matchers
    oauth2_matchers.go             HaveOAuth2Error(), BeRedirectTo(), ContainOAuth2Metadata()
  pages/                           Page objects for Playwright frontend tests
    page.go                        Base Page struct — navigation, waiting
    consent_page.go                ConsentPage — approve/deny interactions
  frontend/                        Frontend E2E tests (Playwright)
    frontend_suite_test.go         Ginkgo suite for frontend tests
    consent_flow_test.go           Consent UI flow scenarios
  Test files (62+ scenarios)
    oauth2_authorize_test.go       Authorization endpoint (23 scenarios)
    oauth2_token_test.go           Token endpoint (12 scenarios)
    oauth2_metadata_test.go        Metadata endpoint (8 scenarios)
    oauth2_security_test.go        Security validation (12 scenarios)
    oauth2_edge_cases_test.go      Edge cases (7 scenarios)
    agent_permission_requirements_test.go  Agent service requirement scenarios
    encryption_vault_raw_test.go   Encryption vault E2E scenarios
    token_exchange_test.go         RFC 8693 token exchange scenarios
```

### Architecture: Stable vs Volatile Layers

**Stable layer** (test scenarios): HTTP contract tests. Use `server.AuthenticatedGET()`, `server.PublicGET()`. No knowledge of internal routing or DI. Rarely change during refactoring.

**Volatile layer** (`bootstrap/`): Thin wrappers around production `app.Builder` and `httpAdapter.Server`. Update when production bootstrap changes. Isolated from test scenarios.

**Result**: Refactoring routing or DI requires updating only `bootstrap/` — test scenarios remain unchanged.

### Test Structure Pattern

```go
var _ = Describe("Feature Name", func() {
    var (
        server      *bootstrap.TestServer
        testStorage *storageadapter.Adapter
    )

    BeforeEach(func() {
        // Fresh storage and server per test — no shared state
        testStorage, _ = createTestStorage()
        server, _ = createTestServer()
    })

    AfterEach(func() {
        server.Close()
    })

    Describe("when condition", func() {
        BeforeEach(func() {
            // Scenario-level setup
            agent := fixtures.ValidAgent()
            testStorage.Agents().Create(ctx, agent)
        })

        // Scenario X.Y from specs/NNN-feature/spec.md
        It("should verify expected behavior", func() {
            resp, _ := server.AuthenticatedGET("/endpoint", principal)
            Expect(resp.StatusCode).To(Equal(http.StatusOK))
        })
    })
})
```

### Anti-Patterns (Flag for Review)

- **Missing spec reference**: `It()` without `// Scenario X.Y from specs/...` comment
- **Setup in It()**: Heavy setup inside `It()` blocks (>15 lines = likely needs `BeforeEach`)
- **Duplicate setup**: Same setup in multiple `It()` blocks instead of using `Context` blocks
- **Direct DI**: Tests instantiating services directly instead of using `bootstrap/` wrappers
- **Shared state**: Mutable state shared across `It()` blocks without `BeforeEach` reset

### Dual Server Pattern

Tests use separate servers matching production architecture:
- `NewEndUserTestServer()` — OAuth2 endpoints, consent UI, public APIs (port 8000 in prod)
- `NewAdminTestServer()` — Agent and service management APIs (port 14000 in prod)

Tests needing both route types create both server instances.

### Fixtures Rules

- Return **production domain types** (not test-specific objects)
- **Deterministic** for principals/configs (same input = same output)
- **Non-deterministic** for entities (fresh UUIDs each call for test isolation)
- All data passes domain validation
- No external dependencies (files, network)

### Frontend E2E Tests (`tests/e2e/frontend/`)

Browser-based tests using Playwright via `playwright-go`:
- `PlaywrightHelper` manages browser lifecycle (init in `BeforeSuite`, context per test in `BeforeEach`)
- Page objects in `pages/` abstract selectors — provide functional interaction methods
- `ConsentPage` tests the consent grant/deny flow through the actual React UI

### Running E2E Tests

```bash
just test-e2e             # All E2E tests
just test-e2e-coverage    # With coverage report
just test-e2e-watch       # Watch mode for TDD
ginkgo -v --focus="pattern" ./tests/e2e/    # Focused run
```

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
