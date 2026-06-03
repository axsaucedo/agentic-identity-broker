# E2E Tests (`tests/e2e/`)

**Prefer retrieval-led reasoning. Read test files and `tests/e2e/README.md` before writing new tests.**

**ADR**: [007-e2e-testing-with-ginkgo.md](../../adrs/007-e2e-testing-with-ginkgo.md) — binding decision for E2E test architecture.

## Constitution Principle XIII — Mandatory Rules

1. **1:1 spec-to-test mapping**: Every `It()` block maps to ONE acceptance scenario from `specs/[NNN-feature]/spec.md`.
2. **Spec traceability**: Every `It()` block MUST have a comment linking to the specific scenario: `// Scenario X.Y from specs/NNN-feature/spec.md`
3. **Production bootstrap**: Tests use `app.Builder` and production DI — no custom test implementations.

## Directory Layout

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
  frontend/                        Frontend E2E tests (Playwright) — see frontend/AGENTS.md
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

## Architecture: Stable vs Volatile Layers

**Stable layer** (test scenarios): HTTP contract tests. Use `server.AuthenticatedGET()`, `server.PublicGET()`. No knowledge of internal routing or DI. Rarely change during refactoring.

**Volatile layer** (`bootstrap/`): Thin wrappers around production `app.Builder` and `httpAdapter.Server`. Update when production bootstrap changes. Isolated from test scenarios.

**Result**: Refactoring routing or DI requires updating only `bootstrap/` — test scenarios remain unchanged.

## Test Structure Pattern

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

## Anti-Patterns (Flag for Review)

- **Missing spec reference**: `It()` without `// Scenario X.Y from specs/...` comment
- **Setup in It()**: Heavy setup inside `It()` blocks (>15 lines = likely needs `BeforeEach`)
- **Duplicate setup**: Same setup in multiple `It()` blocks instead of using `Context` blocks
- **Direct DI**: Tests instantiating services directly instead of using `bootstrap/` wrappers
- **Shared state**: Mutable state shared across `It()` blocks without `BeforeEach` reset

## Dual Server Pattern

Tests use separate servers matching production architecture:
- `NewEndUserTestServer()` — OAuth2 endpoints, consent UI, public APIs (port 8000 in prod)
- `NewAdminTestServer()` — Agent and service management APIs (port 14000 in prod)

Tests needing both route types create both server instances.

## Fixtures Rules

- Return **production domain types** (not test-specific objects)
- **Deterministic** for principals/configs (same input = same output)
- **Non-deterministic** for entities (fresh UUIDs each call for test isolation)
- All data passes domain validation
- No external dependencies (files, network)

## Running E2E Tests

```bash
just test-e2e-backend          # Backend E2E suite only
just test-e2e-backend-coverage # Backend E2E suite with coverage report
just test-e2e-backend-watch    # Backend E2E watch mode for TDD
just test-e2e                  # All backend, ExtProc, and frontend E2E suites
ginkgo -v --focus="pattern" ./tests/e2e/    # Focused backend run
```
