# `internal/domain/id` — Strongly Typed Entity ID Package

Full reference: `internal/domain/AGENTS.md` (domain layer rules apply here too).

**ADR**: `adrs/013-strongly-typed-entity-ids.md` — binding decision for all typed ID usage.

## Purpose

This package defines per-entity ID types that the compiler enforces. Without typed IDs, the compiler cannot distinguish passing a `ServiceID` where an `AgentID` is expected — both would be `string`. With typed IDs, such a mixup is a compile-time error.

## Type Catalogue

### UUID-backed types (backed by `uuid.UUID = [16]byte`)

| Type | Entity | Database Column |
|---|---|---|
| `AgentID` | AI agent | `agents.id` |
| `AuthorizationCodeID` | OAuth2 authorization code | `authorization_codes.id` |
| `CredentialID` | Broker client credential | `broker_client_credentials.id` |
| `GrantID` | User grant | `user_grants.id` |
| `ServiceID` | Third-party OAuth2 service | `thirdparty_oauth2_services.id`, `service_requirements.service_id`, `delegated_tokens.service_id`, `user_sessions.service_id` |
| `SessionID` | OAuth2 user session | `user_sessions.id` |
| `SigningKeyID` | JWT signing key | `signing_keys.id` |
| `UserID` | User account | `users.id` |

Each UUID type exposes:

```go
func NewXxxID() XxxID                          // random UUID
func ParseXxxID(s string) (XxxID, error)       // safe parse, returns error
func MustParseXxxID(s string) XxxID            // panics on invalid — test fixtures ONLY
func (id XxxID) String() string
func (id XxxID) IsZero() bool
func (id XxxID) MarshalJSON() ([]byte, error)              // marshals as JSON string "xxxxxxxx-xxxx-..."
func (id *XxxID) UnmarshalJSON(b []byte) error             // unmarshals JSON string → uuid.Parse
func (id XxxID) Value() (driver.Value, error)              // for database/sql / sqlx
func (id *XxxID) Scan(src interface{}) error               // for database/sql / sqlx (delegates to uuid.UUID.Scan)
func (id XxxID) MarshalText() ([]byte, error)              // for encoding.TextMarshaler (delegates to uuid.UUID)
func (id *XxxID) UnmarshalText(b []byte) error             // for encoding.TextUnmarshaler (delegates to uuid.UUID)
```

**Implementation note**: `MarshalJSON`/`UnmarshalJSON` use explicit JSON string encoding (`json.Marshal`/`json.Unmarshal` into `string`, then `uuid.Parse`) rather than the `(*uuid.UUID)(id)` pointer cast. This is because the standard `uuid.UUID` JSON methods encode as a quoted string and the explicit approach ensures correct JSON output. `Scan`, `MarshalText`, and `UnmarshalText` do use the identity cast `(*uuid.UUID)(id)`, which is safe because `XxxID` has identical memory layout to `uuid.UUID` (`[16]byte`).

### String-backed types (backed by `string`)

| Type | Purpose | Constraint |
|---|---|---|
| `ClientID` | Broker-issued OAuth2 client identifier | `broker_` prefix + 22 chars; `VARCHAR(255)`, not a UUID |
| `ExternalID` | Optional external governance ID | `VARCHAR(255)`, not a UUID |
| `KeyID` | JWT Key ID (`kid` claim) | UUID format |
| `Principal` | Authenticated user identity (email, subject) | From `X-Remote-User` header |

String types expose only `String()`, `IsZero()`, and a `New*()` constructor.

## Code Generation

UUID type implementations are generated from a template — do not edit `uuid_ids_gen.go` by hand:

```bash
go generate ./internal/domain/id/
```

The generator is `gen_ids.go` (build tag `ignore`). It produces `uuid_ids_gen.go`. To add a new UUID type, add an entry to `uuidTypes` in `gen_ids.go` and re-run `go generate`.

## Rules

1. **`ParseXxxID` in production code** — never `MustParseXxxID` in handlers or domain services. Malformed UUIDs must return an error (HTTP 400), not a panic.
2. **`MustParseXxxID` in tests only** — acceptable in test fixtures and table-driven test cases where the UUID string is a compile-time constant.
3. **Handler ordering** — authentication (principal check) must happen *before* UUID format validation. Unauthenticated requests must receive 401, not 400.
4. **No plain `string` for entity IDs** — whenever an entity ID is stored, passed, or returned, use the typed ID. Do not convert back to `string` unless calling an external API that requires it.
5. **`EncryptionContext` uses `ServiceID.String()`** — the encryption context map `{"service_id": "<uuid-string>"}` must use the `.String()` form of `ServiceID`. This is validated by the encryption adapter on every decrypt call.

## Adding a New Entity ID

When a new domain entity with a UUID primary key is introduced:

1. Add an entry to `uuidTypes` in `gen_ids.go`:
   ```go
   {"FooID", "foo"},
   ```
2. Run `go generate ./internal/domain/id/` to regenerate `uuid_ids_gen.go`.
3. Update the type catalogue table in this file.
4. Update `ARCHITECTURE.md` glossary with the new entity and its ID type.
5. Update `AGENTS.md` (root) ADR Decision Index to cross-reference ADR 013.
