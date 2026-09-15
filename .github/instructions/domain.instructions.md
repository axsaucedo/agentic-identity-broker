---
applyTo: "internal/domain/**"
---

# Domain Layer — Innermost Hexagonal Ring

Full reference: `internal/domain/AGENTS.md`

## Zero Infrastructure Imports (Mandatory)

Domain packages must **never** import:
- Database drivers (`database/sql`, `pgx`, `sqlx`)
- HTTP frameworks (`chi`, `net/http` — except `oauth2session` for upstream token calls)
- AWS SDKs, encryption libraries
- Any package from `internal/adapters/` or `internal/app/`

Allowed: Go stdlib (non-I/O), `internal/ports/`, other `internal/domain/` packages, declared external libraries (`jwx`, `oauth2`, `cel-go`) where domain logic requires them.

## Package Responsibilities

| Package | Role |
|---|---|
| `config/` | Config domain types + validation (`LogLevel`, `LogFormat`) |
| `consent/` | Consent management business logic |
| `encryption/` | Encryption domain errors (not implementations) |
| `oauth2/` | OAuth2 authorization service (upstream SSO) |
| `oauth2session/` | OAuth2 session lifecycle + token vault (PKCE, JWE state) |
| `principal/` | Authenticated user identity — `WithPrincipal()` / `FromContext()` |
| `server/` | Server lifecycle config types |
| `services/` | Third-party service CRUD + branch key provisioning |
| `storage/` | Domain data models (Agent, ThirdpartyOAuth2Service, UserGrant, UserSession) |
| `tokenexchange/` | RFC 8693 token exchange + CEL evaluation |

## Domain Services Pattern

Constructor injection with port interfaces — never concrete adapters:
```go
type Service struct {
    repo ports.SomeRepository
}
func NewService(repo ports.SomeRepository) *Service {
    return &Service{repo: repo}
}
```
Services instantiated in `app/builder.go` — never self-instantiate.

## Domain Data Models (`storage/`)

These are **domain entities/value objects**, NOT database models. Adapter-specific details belong in adapters.

- **Entities**: `Agent`, `ThirdpartyOAuth2Service`, `UserGrant`, `UserSession`, `User`
- **Value Objects**: `OAuthScope`, `RequirementType`, `ServiceRequirement`, `DelegatedToken`, `EncryptionContext`
- All entities implement `Validate() error`. Validation is pure — no I/O.

## Error Types

Each package defines its own errors:
- `storage.StorageError` — classified by `ErrorKind` (connection, timeout, validation, not_found, conflict)
- `encryption.EncryptionError` — classified by `ErrorKind`. Messages sanitized — **never contain key material**.
- `tokenexchange.TokenExchangeError` — RFC 8693 error. Token values **never** in messages (SR-005).
- `consent.Err*`, `principal.Err*`, `config.Err*` — sentinel errors.

## Testing

- **TDD**: Tests first, must fail before implementation.
- **Files**: `_test.go` co-located, same package (white-box).
- **Mocking**: Hand-rolled mocks (structs with configurable function fields). No `testify/mock` in domain.
- **Table-driven tests** preferred for validation. Security tests in `_security_test.go` files.
