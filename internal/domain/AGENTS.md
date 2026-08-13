# Domain Layer (`internal/domain/`)

> **Prefer retrieval-led reasoning. Read source files before making assumptions about types, interfaces, or patterns.**

Innermost hexagonal ring. Zero infrastructure dependencies. Never import `adapters/`, `app/`, or external I/O libraries. Domain packages depend on `ports/` interfaces only.

## Package Responsibilities

| Package | Role | Key Types |
|---|---|---|
| `config/` | Config domain types + validation | `LogLevel`, `LogFormat` (validated enums) |
| `consent/` | Consent management business logic | `Service` — orchestrates agent grants |
| `encryption/` | Encryption domain errors (not implementations) | `EncryptionError` with `ErrorKind` |
| `oauth2/` | OAuth2 authorization service (upstream SSO) | `OAuth2AuthorizationService` |
| `oauth2session/` | OAuth2 session lifecycle + token vault | `OAuth2SessionService` (PKCE, JWE state, encrypted tokens) |
| `principal/` | Authenticated user identity context | `WithPrincipal()` / `FromContext()` |
| `server/` | Server lifecycle config types | `Config`, lifecycle helpers |
| `thirdparty/` | Third-party OAuth2 provider management | `ThirdpartyOAuth2ProviderService` (CRUD + encryption) |
| `storage/` | Domain data models (entities + value objects) | `Agent`, `UserGrant`, `UserSession`, `OAuthScope`, etc. |
| `tokenexchange/` | RFC 8693 token exchange + CEL evaluation | `TokenExchangeService` (JWT validation, CEL authz) |

## Critical Rules

### Zero Infrastructure Imports
Forbidden: `database/sql`, `pgx`, `sqlx`, `chi`, AWS SDKs, `internal/adapters/`, `internal/app/`. `net/http` is forbidden by default and allowed only in domain packages with established protocol/domain needs: `oauth2session`, `storage` (RFC 8414 discovery), and `oauth2server` (HTTP status code mapping).
Allowed: Go stdlib (non-I/O by default), `internal/ports/`, other `internal/domain/`, declared external libs (`jwx`, `oauth2`, `cel-go`), plus the explicit `net/http` exceptions above.

### Domain Services Pattern
Constructor injection with port interfaces. Services instantiated in `app/builder.go` — never self-instantiate. Read existing services for patterns.

### Cross-Domain Dependencies
Acceptable within domain ring: `tokenexchange/` → `consent/`, `oauth2session/`. `consent/`, `thirdparty/`, `oauth2session/` → `storage/` (data models). All → `ports/`.

## Domain Data Models (`storage/`)

**Domain entities and value objects** — NOT database models. Adapter-specific details belong in adapters.

### Entities
| Type | Key Invariants |
|---|---|
| `Agent` | `ClientID` nullable (ADR 017), `DisplayName` required, URLs validated for HTTP(S) |
| `UserGrant` | One per (principal, agent) pair (upsert); delegated tokens may be empty when no delegation is required |
| `UserSession` | One per (principal, service_id), tokens encrypted, user-session `EncryptionContext` uses `service_id` while other encrypted assets may use a different single subject such as `kid` (ADR 008 amendment) |
| `User` | ID + email, timestamps |

### Value Objects
- `OAuthScope` — permission scope (`ScopeValue` + `Description`)
- `RequirementType` — enum: mandatory/optional
- `ServiceRequirement` — agent's declared need: `ServiceID` + `RequirementType` + `RequiredScopes[]` ceiling, or `RequireAllScopes` to consume the full granted permission-set scope union
- `DelegatedToken` — grant component: `ThirdpartyOAuth2ServiceID` + `Scopes[]`
- `EncryptionContext` — AAD metadata (JSONB, implements `driver.Valuer`/`sql.Scanner`); exactly one approved subject key such as `service_id` or `kid`
- `BranchKeySubject` — typed encryption namespace wrapper selecting either the `service` or `signing_key` branch-key subject
- `ConnectionParameters`, `StorageBackend` (memory/postgres), `DiscoveryConfig`, `OAuth2Endpoints`

### Validation
All entities implement `Validate() error`. Some add `ValidateForCreate()`/`ValidateForUpdate()`. Validation is pure — no I/O.

### Error Types
- `storage.StorageError` — `ErrorKind`: connection, timeout, validation, not_found, conflict, unknown
- `encryption.EncryptionError` — `ErrorKind`: encryption_failed, decryption_failed, context_mismatch, integrity_violation, kek_unavailable. **Never contains key material.**
- `tokenexchange.TokenExchangeError` — RFC 8693 error. Token values **never** in messages (SR-005).
- Sentinel errors: `consent.Err*`, `principal.Err*`, `config.Err*`

### Storage Discovery
`storage/discovery.go` — `DiscoverOAuth2Endpoints()` for RFC 8414 metadata. HTTPS required in prod; HTTP allowed for localhost in dev.

## Testing

- **TDD**: Tests first, must fail before implementation
- **Files**: `_test.go` co-located, same package (white-box)
- **Mocking**: Hand-rolled mocks (structs with function fields). No `testify/mock` in domain.
- **Table-driven tests** for validation. Security tests in `_security_test.go` files.
