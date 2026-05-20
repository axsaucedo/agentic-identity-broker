# Adapters Layer (`internal/adapters/`)

> **Prefer retrieval-led reasoning. Read the port interfaces in `internal/ports/*.go` before implementing any adapter.**

Adapters implement port interfaces. Each adapter imports only from `ports/` and `domain/`. **Cross-adapter imports are forbidden.**

## Adapter Map

### Encryption (`encryption/`)
Single `ports.EncryptionPort` implementation with base64-key mode for testing:

| Package | Purpose |
|---|---|
| `aws/` | `EncryptionPort` implementation — AWS KMS Hierarchical Keyring (production) or base64 AES-256 key mode (dev/test via `testutil.NewTestEncryptionAdapter()`). |
| `branchkey/` | Shared `BranchKeyIdProvider` — deterministic ID: `service_{service_id}_branch_key`. |

Encryption rules: See `.claude/skills/aws-crypto-go/SKILL.md`.

### HTTP (`http/`)
Dual-server (ADR 004): end-user `:8000` + admin `:14000`.

| Package | Purpose |
|---|---|
| `routing/{admin,enduser}.go` | Route registration ONLY — receives pre-wired handler structs, never instantiates services |
| `handlers/admin/` | `AgentsHandler`, `ServicesHandler` (CRUD) |
| `handlers/consent/` | `UserInfoHandler`, `AgentsHandler`, `AgentDetailHandler`, `AgentInfoHandler`, `GrantsHandler` |
| `enduser/` | `OAuth2AuthorizeHandler`, `OAuth2TokenHandler`, `OAuth2MetadataHandler` |
| `oauth2_sessions/` | Session initiate/callback/terminate |
| `middleware/` | Auth (`RequirePrincipal`), CORS, CSRF, audit |
| `upstream/` | Upstream OAuth2 proxy |

**Handler rules**: HTTP-to-domain translation only. Parse request → call domain service → format response. Domain errors → HTTP status codes. New handlers → add to `app/handlers.go`, wire in `builder.go`, register in `routing/`.

### JWKS (`jwks/`)
`lestrrat-go/jwx/v3` `jwk.Cache` for background refresh. Implements `ports.JWKSPort`.

### Storage (`storage/`)

| Package | Purpose |
|---|---|
| `factory.go` | Backend selection (`memory`/`postgres`) → returns `Adapter` composite with all repos |
| `memory/` | Maps + `sync.RWMutex`. All repository interfaces. Dev/testing. |
| `postgres/` | `sqlx` + `pgx v5`. All repository interfaces. Production. |

Both adapters implement identical interface contracts. Migrations in `/migrations/` (go-migrate naming).

## Rules

- **Cross-adapter ban**: `encryption/aws/` never imports `storage/postgres/`, etc.
- **Adapter → domain boundary**: Adapters import `ports/` and `domain/`. Domain never imports adapters. Adapter-specific types (SQL models, AWS types) never leak into port interfaces.
- **Testing**: `_test.go` same package. `testify/mock` with `.On()`/`.AssertExpectations()`. HTTP via `httptest`. `_integration_test.go` for real infra. Storage: both backends tested against same contracts.
