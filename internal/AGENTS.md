# Backend — Hexagonal Architecture (`internal/`)

> **Prefer retrieval-led reasoning. Read source files before making assumptions about types, interfaces, or patterns. Read the relevant child `AGENTS.md` for detailed rules per layer.**

## Child AGENTS.md Index
`internal/domain/AGENTS.md` | Domain layer rules, data models, error types, zero-infra imports
`internal/domain/id/AGENTS.md` | Strongly typed entity IDs, code generation, type catalogue
`internal/ports/AGENTS.md` | Port interface catalogue (read `ports/*.go` directly for contract)
`internal/adapters/AGENTS.md` | Adapter map, cross-adapter ban, storage/encryption/HTTP
`internal/extproc/AGENTS.md` | Standalone ExtProc service (separate binary, no broker imports)

## Hexagonal Architecture Boundaries

Dependency arrows flow **inward only**: `app/` → `adapters/` → `ports/` ← `domain/`

- **Domain** (`domain/`) — Pure business logic. ZERO infrastructure imports. Never imports `adapters/` or `app/`.
- **Ports** (`ports/`) — Interfaces + minimal DTOs only. The hexagonal boundary definitions.
- **Adapters** (`adapters/`) — Infrastructure implementations of ports. May import `ports/` and `domain/`, **never other adapters**.
- **App** (`app/`) — DI wiring only. Composes adapters into domain services via Builder pattern.

Violations of inward-dependency flow break the architecture. When in doubt, read `ports/*.go` for available interfaces before creating new cross-layer dependencies.

## Package Map

```
domain/
  config/          Config domain types + validation (LogLevel, LogFormat)
  consent/         ConsentService — agent delegation & grant management
  encryption/      Encryption domain errors (not implementations)
  oauth2/          OAuth2AuthorizationService — upstream SSO integration
  oauth2session/   OAuth2SessionService — token vault, PKCE, JWE state tokens
  principal/       Principal extraction from X-Remote-User header
  server/          Server lifecycle config types
  thirdparty/      ThirdpartyOAuth2ProviderService — provider CRUD + encryption (domain service)
  storage/         Domain data models (Agent, UserGrant, UserSession) + endpoint discovery
  tokenexchange/   TokenExchangeService + CEL policy evaluation (RFC 8693)

ports/             7 interface files defining ALL hexagonal boundaries
  cel.go           CELCompilerPort
  config.go        ConfigPort + Config struct
  encryption.go    EncryptionPort, BranchKeyRepository, BranchKeyIdProvider, BranchKeyManager
  jwks.go          JWKSPort
  oauth2.go        OAuth2Service interface
  server.go        HealthState
  storage.go       AgentRepository, UserGrantRepository, UserSessionRepository, HealthChecker
  thirdparty_provider.go  ThirdpartyOAuth2ProviderRepository (uses model.ThirdpartyOAuth2ProviderEntity)

adapters/
  encryption/
    aws/           AWS KMS Hierarchical Keyring (production)
    branchkey/     BranchKeyIdProvider implementation
    memory/        In-memory encryption (development/testing)
  http/
    enduser/       End-user OAuth2 handlers (authorize, token, metadata)
    handlers/      SPA handler, admin/ and consent/ handler groups
    middleware/     Authentication, logging, recovery middleware
    oauth2_sessions/  OAuth2 session management handlers
    routing/       SetupAdminRoutes(), SetupEnduserRoutes() — route registration ONLY
    upstream/      Upstream OAuth2 proxy integration
  jwks/            JWKS fetching adapter (lestrrat-go/jwx jwk.Cache)
  storage/
    factory.go     Backend selection (memory vs postgres) from config
    memory/        In-memory storage (maps + sync.RWMutex)
    postgres/      PostgreSQL storage (sqlx + pgx v5)
    (no noop — encryption is mandatory, no fallback)

app/
  builder.go       Builder pattern — all DI wiring
  handlers.go      AdminHandlers + EnduserHandlers struct definitions
```

## Builder Pattern (DI Wiring)

All service instantiation happens in `app/builder.go` via `NewBuilder().With*().Build()`. Three phases:

1. **Encryption** — Resolve encryption adapter (AWS KMS, memory, or no-op) + branch key manager
2. **Domain services** — Create ConsentService, OAuth2Service, OAuth2SessionService, TokenExchangeService from port interfaces
3. **Handlers** — Wire AdminHandlers and EnduserHandlers with pre-built services

**Rules**:
- Routing functions (`SetupAdminRoutes`, `SetupEnduserRoutes`) receive pre-wired handler structs — they **never instantiate services**
- New handlers must be added to `AdminHandlers` or `EnduserHandlers` in `app/handlers.go`, instantiated in `builder.go`, then registered in `routing/`
- All adapters depend on ports (interfaces), not concrete implementations

### Handler Structs

**AdminHandlers**: `Agents` (`*admin.AgentsHandler`), `Services` (`*admin.ServicesHandler`)

**EnduserHandlers**: `UserInfo`, `Agents`, `AgentDetail`, `AgentGrants`, `Grants`, `OAuth2Sessions`, `OAuth2Authorize`, `OAuth2Token`, `OAuth2Metadata`, `SPA`

### Routing Signatures

```go
func SetupAdminRoutes(r chi.Router, h *app.AdminHandlers)
func SetupEnduserRoutes(r chi.Router, h *app.EnduserHandlers, cfg EnduserRouteConfig)
```

## Testing Conventions

- **TDD**: Red-green-refactor. Tests written first, must compile and fail semantically before implementation.
- **Files**: `_test.go` co-located in same package (white-box testing). Same package name (not `_test` suffix).
- **Assertions**: `testify` — `require` for preconditions (fatal), `assert` for checks (soft).
- **Subtests**: `t.Run("description", ...)` is the dominant pattern. Table-driven (`[]struct{...}` loop) used for validation/parameterized scenarios.
- **Mocking**:
  - **Adapters layer**: `testify/mock` with `.On()` / `.AssertExpectations()`. Mock helpers in `mocks_test.go`.
  - **Domain layer**: Hand-rolled mocks — structs implementing port interfaces with configurable function fields.
- **HTTP tests**: `httptest.NewRecorder()` + `httptest.NewRequest()` with chi router for URL param injection.

## Key Storage Types (domain/storage/)

These are **domain data models**, NOT database models. Adapter-specific records (e.g., postgres row structs) stay in adapter packages.

| Type | Location | Key Fields |
|---|---|---|
| `Agent` | `domain/storage/` | ID, ClientID, DisplayName, Description, ServiceRequirements |
| `ThirdpartyOAuth2ProviderEntity` | `domain/model/` | ID, DisplayName, ClientID, Secret (value object), Scopes, ProtectedResources |
| `UserGrant` | `domain/storage/` | ID, Principal, AgentID, DelegatedTokens, ValidUntil |
| `UserSession` | `domain/storage/` | ID, Principal, ServiceID, AccessToken (encrypted), RefreshToken (encrypted), ExpiresAt |

## Port Interfaces Quick Reference

| Port | Interface | File |
|---|---|---|
| Encryption | `EncryptionPort` | `ports/encryption.go` |
| Branch Key | `BranchKeyManager` (= `BranchKeyRepository`) | `ports/encryption.go` |
| Config | `ConfigPort` | `ports/config.go` |
| Storage | `AgentRepository`, `UserGrantRepository`, `UserSessionRepository` | `ports/storage.go` |
| Provider Storage | `ThirdpartyOAuth2ProviderRepository` | `ports/thirdparty_provider.go` |
| JWKS | `JWKSPort` | `ports/jwks.go` |
| OAuth2 | `OAuth2Service` | `ports/oauth2.go` |
| CEL | `CELCompilerPort` | `ports/cel.go` |
| Health | `HealthChecker` | `ports/storage.go` |
