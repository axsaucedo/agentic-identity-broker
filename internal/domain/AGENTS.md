# Domain Layer (`internal/domain/`)

**This is the innermost hexagonal ring. Zero infrastructure dependencies. Never import `adapters/`, `app/`, or any external I/O library.**

Domain packages depend on `ports/` interfaces for infrastructure needs. All dependency arrows point outward to ports, never inward to adapters.

## Package Responsibilities

| Package | Role | Key Types |
|---|---|---|
| `config/` | Configuration domain types + validation | `LogLevel`, `LogFormat` — validated enums with `Valid()` methods |
| `consent/` | Consent management business logic | `Service` — orchestrates agent grants using port interfaces |
| `encryption/` | Encryption domain errors (not implementations) | `EncryptionError` — typed error with `ErrorKind` classification |
| `oauth2/` | OAuth2 authorization service (upstream SSO) | `OAuth2AuthorizationService` — handles authorize requests |
| `oauth2session/` | OAuth2 session lifecycle + token vault | `OAuth2SessionService` — PKCE, JWE state tokens, encrypted token storage |
| `principal/` | Authenticated user identity context | `WithPrincipal()` / `FromContext()` — context-based principal propagation |
| `server/` | Server lifecycle configuration types | `Config`, lifecycle helpers |
| `thirdparty/` | Third-party OAuth2 provider management | `ThirdpartyOAuth2ProviderService` — CRUD + encryption/decryption of provider secrets |
| `storage/` | Domain data models (entities + value objects) | `Agent`, `UserGrant`, `UserSession`, `OAuthScope`, etc. |
| `tokenexchange/` | RFC 8693 token exchange + CEL evaluation | `TokenExchangeService` — JWT validation, CEL authorization, token retrieval |

## Critical Rules

### Zero Infrastructure Imports
Domain packages must **never** import:
- Database drivers (`database/sql`, `pgx`, `sqlx`)
- HTTP frameworks (`chi`, `net/http` — except `oauth2session` which uses `http.Client` for upstream token calls)
- AWS SDKs, encryption libraries
- Any package from `internal/adapters/` or `internal/app/`

Allowed imports: Go stdlib (non-I/O), `internal/ports/`, other `internal/domain/` packages, and declared external libraries (e.g., `jwx`, `oauth2`, `cel-go`) where domain logic requires them.

### Domain Services Pattern
Domain services follow a consistent constructor injection pattern:
```go
type Service struct {
    repo     ports.SomeRepository  // port interface, never concrete adapter
    otherDep ports.AnotherPort
}

func NewService(repo ports.SomeRepository, ...) *Service {
    return &Service{repo: repo, ...}
}
```

Services are instantiated in `app/builder.go` — never self-instantiate or resolve dependencies.

### Cross-Domain Dependencies
Some domain packages depend on each other:
- `tokenexchange/` → `consent/`, `oauth2session/` (orchestration)
- `oauth2session/` → `storage/` (data models)
- `consent/` → `storage/` (data models)
- `services/` → `storage/` (data models)
- All services → `ports/` (interface dependencies)

These dependencies are acceptable because they stay within the domain ring. The direction is always toward more fundamental packages.

## Domain Data Models (`storage/` package)

The `storage/` package defines **domain entities** and **value objects** — these are NOT database models. Adapter-specific details (SQL column mappings, serialization) belong in adapters.

### Entities (with identity and lifecycle)
| Type | Description | Key Invariants |
|---|---|---|
| `Agent` | AI agent registered in the broker | `ClientID` unique, `DisplayName` required, URLs validated for HTTP(S) |
| `UserGrant` | User delegating scopes to an agent | One grant per (principal, agent) pair (upsert), must have ≥1 `DelegatedToken` |
| `UserSession` | Authenticated OAuth2 session | One per (principal, service_id), tokens encrypted at rest, `EncryptionContext` simplified to `service_id` only (ADR 008) |
| `User` | Basic user entity | ID + email, timestamps |

### Value Objects (identity-less, validated)
| Type | Description |
|---|---|
| `OAuthScope` | Permission scope: `ScopeValue` + `Description`, both required |
| `RequirementType` | Enum: `"mandatory"` or `"optional"` — with `Valid()`, `IsMandatory()`, `IsOptional()` |
| `ServiceRequirement` | Agent's declared need: `ServiceID` + `RequirementType` + `RequiredScopes[]` |
| `DelegatedToken` | Grant component: `ThirdpartyOAuth2ServiceID` + `Scopes[]` |
| `EncryptionContext` | AAD metadata for token encryption: `ServiceID` only (ADR 008). Implements `driver.Valuer`/`sql.Scanner` for JSONB |
| `ConnectionParameters` | PostgreSQL connection details with `Redacted()` for safe logging |
| `StorageBackend` | Enum: `"memory"` or `"postgres"` |
| `DiscoveryConfig` | OAuth2 endpoint discovery settings |
| `OAuth2Endpoints` | Token + authorize endpoint URLs |

### Validation Pattern
All entities implement `Validate() error` for domain constraint checking. Some add `ValidateForCreate()` / `ValidateForUpdate()` for lifecycle-specific rules. Validation is pure — no I/O, no database lookups.

### Error Types
Each domain package defines its own error types:
- `storage.StorageError` — Classified by `ErrorKind` (connection, timeout, validation, not_found, conflict, unknown). Wraps adapter errors without exposing implementation details.
- `encryption.EncryptionError` — Classified by `ErrorKind` (encryption_failed, decryption_failed, context_mismatch, integrity_violation, kek_unavailable). Messages sanitized — **never contain key material or tokens**.
- `tokenexchange.TokenExchangeError` — RFC 8693 error with `code`, `description`, `httpStatus`. Token values **never** appear in error messages (SR-005).
- `consent.Err*` — Sentinel errors (`ErrAgentNotFound`, `ErrServiceNotFound`, `ErrInvalidScopes`, `ErrAgentAccessDenied`, `ErrGrantExpired`).
- `principal.Err*` — Sentinel errors for missing/invalid principal.
- `config.Err*` — Configuration validation errors.

### Storage Discovery
`storage/discovery.go` provides `DiscoverOAuth2Endpoints()` for RFC 8414 OAuth2 metadata discovery. This fetches endpoints from an issuer's `.well-known` URL. HTTPS required in production; HTTP allowed for localhost in dev mode.

## Testing in Domain

- **TDD**: Write tests first. Tests must compile and fail before implementation.
- **Files**: `_test.go` co-located, same package (white-box testing).
- **Mocking**: Hand-rolled mocks — structs implementing port interfaces with configurable function fields. No `testify/mock` in domain (that pattern is for adapters).
- **Table-driven tests**: Preferred for validation and parameterized scenarios.
- **Security tests**: `_security_test.go` files for security-sensitive behavior (e.g., `connection_security_test.go`, `state_token_security_test.go`).
