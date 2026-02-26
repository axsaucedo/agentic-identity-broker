---
applyTo: "internal/**"
---

# Backend — Hexagonal Architecture

Full reference: `internal/AGENTS.md`

## Architecture Boundaries (Mandatory)

Dependency arrows flow **inward only**: `app/` → `adapters/` → `ports/` ← `domain/`

- **Domain** (`domain/`) — Pure business logic. ZERO infrastructure imports. Never imports `adapters/` or `app/`.
- **Ports** (`ports/`) — Interfaces + minimal DTOs only. The hexagonal boundary definitions.
- **Adapters** (`adapters/`) — Infrastructure implementations of ports. May import `ports/` and `domain/`, **never other adapters**.
- **App** (`app/`) — DI wiring only. Composes adapters into domain services via Builder pattern.

Violations of inward-dependency flow break the architecture. Read `ports/*.go` for available interfaces before creating new cross-layer dependencies.

## Package Map

```
domain/
  config/          Config domain types + validation
  consent/         ConsentService — agent delegation & grant management
  encryption/      Encryption domain errors (not implementations)
  oauth2/          OAuth2AuthorizationService — upstream SSO integration
  oauth2session/   OAuth2SessionService — token vault, PKCE, JWE state tokens
  principal/       Principal extraction from X-Remote-User header
  server/          Server lifecycle config types
  services/        ThirdpartyOAuth2ServiceProvider — service CRUD + encryption
  storage/         Domain data models (Agent, UserGrant, ThirdpartyOAuth2Service, UserSession)
  tokenexchange/   TokenExchangeService + CEL policy evaluation (RFC 8693)
ports/             7 interface files: cel, config, encryption, jwks, oauth2, server, storage
adapters/
  encryption/      aws/ (production KMS), memory/ (dev/test), branchkey/ (shared ID logic)
  http/            handlers/, middleware/, routing/, enduser/, oauth2_sessions/, upstream/
  jwks/            JWKS fetching (lestrrat-go/jwx)
  storage/         factory.go, memory/ (maps + RWMutex), postgres/ (sqlx + pgx v5), noop/
app/
  builder.go       Builder pattern — all DI wiring (3 phases: encryption → services → handlers)
  handlers.go      AdminHandlers + EnduserHandlers struct definitions
```

## Builder Pattern (DI Wiring)

All service instantiation in `app/builder.go` via `NewBuilder().With*().Build()`.

- Routing functions (`SetupAdminRoutes`, `SetupEnduserRoutes`) receive pre-wired handler structs — they **never instantiate services**
- New handlers: add to `AdminHandlers`/`EnduserHandlers` in `handlers.go`, instantiate in `builder.go`, register in `routing/`

## Testing Conventions

- **TDD**: Red-green-refactor. Tests first.
- **Files**: `_test.go` co-located, same package (white-box).
- **Assertions**: `testify` — `require` for preconditions (fatal), `assert` for checks (soft).
- **Subtests**: `t.Run("description", ...)` dominant. Table-driven for validation/parameterized scenarios.
- **Mocking**: `testify/mock` in adapters, hand-rolled mocks in domain.
- **HTTP tests**: `httptest.NewRecorder()` + `httptest.NewRequest()` with chi router.

## Port Interfaces Quick Reference

| Port | Interface | File |
|---|---|---|
| Encryption | `EncryptionPort` | `ports/encryption.go` |
| Config | `ConfigPort` | `ports/config.go` |
| Storage | `AgentRepository`, `ThirdpartyOAuth2ServiceRepository`, `UserGrantRepository`, `UserSessionRepository` | `ports/storage.go` |
| JWKS | `JWKSPort` | `ports/jwks.go` |
| OAuth2 | `OAuth2Service` | `ports/oauth2.go` |
| CEL | `CELCompilerPort` | `ports/cel.go` |
| Health | `HealthChecker` | `ports/storage.go` |
