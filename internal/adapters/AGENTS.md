# Adapters Layer (`internal/adapters/`)

**Adapters implement port interfaces. Each adapter imports only from `ports/` and `domain/`. Cross-adapter imports are forbidden.**

Adapters are infrastructure implementations — they bridge external systems (databases, APIs, HTTP, AWS) to the domain through the port interfaces defined in `internal/ports/`.

## Adapter Map

### Encryption (`encryption/`)
Three implementations of `ports.EncryptionPort`:

| Package | Purpose | Key Details |
|---|---|---|
| `aws/` | Production — AWS KMS Hierarchical Keyring | Three-layer envelope encryption (KEK → Branch Key → DEK). Uses AWS Encryption SDK with `RequireEncryptRequireDecrypt` commitment policy. Branch keys cached in DynamoDB. `AWSAdapter` for encrypt/decrypt, `AWSBranchKeyManager` for key provisioning. |
| `memory/` | Development/testing — in-memory encryption | `InMemoryBranchKeyRepository` tracks provisioned keys in a `sync.RWMutex`-protected map. Delegates to `branchkey.Provider` for ID generation. |
| `branchkey/` | Shared branch key ID logic | `Provider` implements `ports.BranchKeyIdProvider` — deterministic ID format: `service_{service_id}_branch_key`. Used by both `aws/` and `memory/`. |

**Encryption rules** (from constitution + ADR 009):
- Commitment policy: Always `RequireEncryptRequireDecrypt`
- Encryption context: Only `{"service_id": "<id>"}` (ADR 008). Never include secrets.
- Hierarchical keyring: Go SDK has no Caching CMM — use Hierarchical Keyring with DynamoDB branch key store
- Explicit wrapping keys: Always specify KMS key ARN. No discovery mode.

### NoOp (`storage/noop/`)
`NoOpEncryption` — pass-through that returns plaintext unchanged. **Development/testing only.** 

### HTTP (`http/`)
Dual-server architecture (ADR 004): end-user `:8000` + admin `:14000`.

| Package | Purpose | Key Details |
|---|---|---|
| `server.go` | HTTP server lifecycle | `Server` struct — generic, not server-type-aware. chi router, atomic health state, graceful shutdown. Route setup via closure injection. |
| `middleware.go` | Common middleware | `LoggingMiddleware`, `RecoveryMiddleware` |
| `middleware/` | Specialized middleware | `RequirePrincipalMiddleware` (auth), `CORSMiddleware`, `CSRFMiddleware`, `OAuth2AuditMiddleware` |
| `routing/admin.go` | Admin route registration | `SetupAdminRoutes(r chi.Router, h *app.AdminHandlers)` — CRUD for agents and services |
| `routing/enduser.go` | End-user route registration | `SetupEnduserRoutes(r chi.Router, h *app.EnduserHandlers, cfg EnduserRouteConfig)` — consent, sessions, OAuth2, SPA |
| `handlers/admin/` | Admin API handlers | `AgentsHandler` (CRUD), `ServicesHandler` (CRUD) |
| `handlers/consent/` | Consent UI API handlers | `UserInfoHandler`, `AgentsHandler`, `AgentDetailHandler`, `AgentGrantsHandler`, `GrantsHandler` + integration tests |
| `handlers/spa.go` | SPA serving | History API fallback for React SPA (ADR 005) |
| `enduser/` | End-user OAuth2 handlers | `OAuth2AuthorizeHandler`, `OAuth2TokenHandler`, `OAuth2MetadataHandler` |
| `oauth2_sessions/` | OAuth2 session management | Session initiate/callback/terminate handlers |
| `upstream/` | Upstream OAuth2 proxy | `OAuth2Client` for upstream token exchange |

**Routing rules**:
- Routing functions receive **pre-wired handler structs** from `app/` — they never instantiate services
- Route registration only — no business logic in routing files
- Admin routes: `POST/GET/PUT/DELETE /api/agents/{agent-id}`, `POST/GET/PUT/DELETE /api/services/{service-id}`
- End-user routes: `/api/me`, `/api/consent/...`, `/api/third-party/...`, `/oauth2/authorize`, `/oauth2/token`, `/.well-known/oauth-authorization-server`

**Handler rules**:
- Handlers call domain services — they never contain business logic
- HTTP-to-domain translation: parse request → call service → format response
- Error mapping: domain errors → HTTP status codes + RFC-compliant error bodies
- New handlers must be added to `AdminHandlers` or `EnduserHandlers` structs in `app/handlers.go`

### JWKS (`jwks/`)
Single adapter implementing `ports.JWKSPort`:
- Uses `lestrrat-go/jwx/v3` `jwk.Cache` for automatic background JWKS refresh
- Configurable min/max refresh intervals
- Each adapter instance maintains its own cached key set

### Storage (`storage/`)
Factory pattern with backend selection:

| Package | Purpose | Key Details |
|---|---|---|
| `factory.go` | Backend selection factory | `NewAdapter(config)` → returns `Adapter` composite with all repositories. Selects `memory` or `postgres` based on config. |
| `memory/` | In-memory storage | Maps + `sync.RWMutex` for thread safety. Implements all 5 repository interfaces. Used in dev/testing. |
| `postgres/` | PostgreSQL storage | `sqlx` + `pgx v5` driver. Implements all 5 repository interfaces. Production backend. |
| `noop/` | NoOp encryption adapter | See encryption section above. |

Each storage package implements these repository interfaces from `ports/storage.go`:
- `AgentRepository` — CRUD + `GetByClientID()`
- `ThirdpartyOAuth2ServiceRepository` — CRUD + `FindByProtectedResource()` + `CountGrantsReferencingService()`
- `UserGrantRepository` — CRUD + `FindByPrincipalAndAgent()` + `ListByPrincipal()` + cascade operations
- `UserSessionRepository` — CRUD + `FindByPrincipalAndService()` + `CountByService()`
- `UserRepository` — basic CRUD (via adapter composite for memory)

**Storage conventions**:
- `Adapter` struct composes all repositories — callers get one object with all storage access
- Memory adapter: all data in Go maps, `sync.RWMutex` for concurrency
- Postgres adapter: `sqlx` for query execution, `pgx/v5` as driver, connection pooling via `sqlx.DB`
- Migrations in `/migrations/` using go-migrate naming convention (`NNN_description.{up,down}.sql`)
- Both adapters must implement identical interface contracts — same error semantics, same behavior under concurrency

## Rules

### Cross-Adapter Import Ban
`encryption/aws/` must **never** import `storage/postgres/`. `http/handlers/` must **never** import `storage/memory/`. Adapters communicate only through port interfaces defined in `ports/`.

### Adapter ↔ Domain Boundary
- Adapters may import from `ports/` and `domain/`
- Adapters must **not** be imported by `domain/` packages
- Adapter-specific types (SQL models, AWS SDK types) must not leak into port interfaces

### Testing in Adapters
- **Files**: `_test.go` co-located, same package (white-box testing)
- **Mocking**: `testify/mock` with `.On()` / `.AssertExpectations()`. Mock helpers in `mocks_test.go`.
- **HTTP tests**: `httptest.NewRecorder()` + `httptest.NewRequest()` with chi router for URL param injection
- **Integration tests**: `_integration_test.go` suffix for tests requiring real infrastructure
- **Storage tests**: Both memory and postgres adapters tested against same interface contracts
