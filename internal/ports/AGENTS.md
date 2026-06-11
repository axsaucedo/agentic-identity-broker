# Ports Layer (`internal/ports/`)

> **Prefer retrieval-led reasoning. Read the port `.go` files directly before implementing — they are the authoritative contract.**

Ports are **interfaces only** + minimal DTOs. No business logic. No implementations. Domain services depend on these interfaces; adapters implement them.

## Port Catalog

| File | Interface(s) | Purpose |
|---|---|---|
| `cel.go` | `CELCompilerPort`, `CELProgram` | CEL expression compilation/evaluation. DTOs: `CELAuthorizationContext`, `CELRequestContext`, `CELClaimExtractionContext` |
| `cimd.go` | See `cimd.go` for authoritative interface/error/DTO names | CIMD client resolution/fetching ports. Read the file directly for the exact contract surface. |
| `config.go` | `ConfigPort` | Configuration loading/access. DTOs: `Config` + all nested config types (largest port file) |
| `encryption.go` | `EncryptionPort`, `BranchKeyRepository`, `BranchKeyIdProvider` | Envelope encryption with AAD + branch key management. Alias: `BranchKeyManager = BranchKeyRepository` |
| `jwks.go` | `JWKSPort`, `JWKSHealthPort` | JWKS retrieval with embedded health reporting plus the health-only facet |
| `oauth2.go` | `OAuth2Service` | OAuth2 authorization + RFC 8414 metadata. DTOs: `AuthorizationRequest`, `AuthorizationDecision`, `MetadataResponse` |
| `oauth2server.go` | `SigningKeyManager`, `SigningKeyBootstrapCoordinator`, `CredentialGenerator` | OAuth2 server signing key lifecycle, initial-key bootstrap coordination, and credential generation ports. Read the file directly for the exact contract surface. |
| `server.go` | `HealthState` (enum) | Server lifecycle: `Starting`, `Healthy`, `ShuttingDown`, `Unhealthy` |
| `storage.go` | `HealthChecker`, `UserRepository`, `AgentRepository`, `UserGrantRepository`, `UserSessionRepository` | All storage repos. Sentinel: `ErrNotFound` |
| `thirdparty_provider.go` | `ThirdpartyOAuth2ProviderRepository` | Provider config storage. Uses `model.ThirdpartyOAuth2ProviderEntity` (encrypted `Secret`). Adapters never decrypt. |

## Rules

- **ISP**: Repositories expose only operations their consumers need. Justify new operations.
- **No business logic**: No validation, orchestration, conditionals, or error handling beyond sentinels.
- **New ports**: New architectural boundaries require an ADR. New methods require consumer justification.
- **Minimal DTOs**: Prefer domain types from `internal/domain/storage/`; port DTOs only when domain types don't fit.
- **Encryption context**: `Encrypt/Decrypt` accept exactly one approved subject key per ciphertext namespace: `{"service_id": "<id>"}` for service-scoped secrets or `{"kid": "<id>"}` for signing-key private material (ADR 008 amendment). Never secrets, and never both keys together.
- **Error conventions**: Storage → `storage.StorageError` (domain). Encryption → `encryption.EncryptionError` (domain). Quick identity check: `ports.ErrNotFound`.
