# Ports Layer (`internal/ports/`)

**Ports are interfaces only, plus minimal DTOs needed by those interfaces. No business logic. No implementations.**

This layer defines all hexagonal architecture boundaries. Domain services depend on these interfaces; adapters implement them. Ports must remain thin — they are the contract, not the behavior.

## Port Catalog

7 files defining every cross-boundary interface in the system:

### `cel.go` — CEL Expression Compilation & Evaluation
| Interface | Methods | Purpose |
|---|---|---|
| `CELCompilerPort` | `CompileExpression()`, `ValidateExpression()` | Compile/validate CEL authorization expressions |
| `CELProgram` | `Eval()` | Execute compiled CEL expression with variable bindings |

DTOs: `CELAuthorizationContext`, `CELRequestContext`, `CELClaimExtractionContext`

### `config.go` — Application Configuration
| Interface | Methods | Purpose |
|---|---|---|
| `ConfigPort` | `GetConfig()`, `GetSources()`, `Reload()` | Load, validate, and access typed configuration |

DTOs: `Config`, `ServerConfig`, `ServerInstanceConfig`, `AuthenticationConfig`, `PreauthConfig`, `ShutdownConfig`, `StorageConfig`, `LogConfig`, `ThirdPartyOAuth2Config`, `OAuth2AuthServerConfig`, `TokenExchangeConfig`, `SecurityConfig`, `EncryptionConfig`, and nested config types. This is the largest port file — it defines the entire configuration schema.

### `encryption.go` — Encryption & Branch Key Management
| Interface | Methods | Purpose |
|---|---|---|
| `EncryptionPort` | `Encrypt()`, `Decrypt()` | Envelope encryption with AAD context binding |
| `BranchKeyRepository` | `Create()` | Provision service-specific branch keys in DynamoDB |
| `BranchKeyIdProvider` | `GenerateBranchKeyId()`, `ExtractServiceIdFromBranchKey()` | Deterministic branch key ID resolution (`service_{id}_branch_key`) |

Type alias: `BranchKeyManager = BranchKeyRepository`

### `jwks.go` — JSON Web Key Set Fetching
| Interface | Methods | Purpose |
|---|---|---|
| `JWKSPort` | `GetKeySet()`, `GetKey()` | Fetch and cache JWKS from upstream OAuth2 server |

Depends on: `github.com/lestrrat-go/jwx/v3/jwk` for `jwk.Set` and `jwk.Key` return types.

### `oauth2.go` — OAuth2 Authorization Service
| Interface | Methods | Purpose |
|---|---|---|
| `OAuth2Service` | `HandleAuthorization()`, `GenerateMetadata()` | Process OAuth2 authorization requests, generate RFC 8414 metadata |

DTOs: `AuthorizationRequest`, `AuthorizationDecision`, `MetadataResponse`

### `server.go` — Server Health State
| Type | Values | Purpose |
|---|---|---|
| `HealthState` (enum) | `Starting`, `Healthy`, `ShuttingDown`, `Unhealthy` | Server lifecycle state reporting |

### `storage.go` — All Storage Repository Interfaces
| Interface | Key Methods | Purpose |
|---|---|---|
| `HealthChecker` | `HealthCheck()` | Storage backend health verification |
| `UserRepository` | CRUD + `ListUsers()` | User entity persistence |
| `AgentRepository` | CRUD + `GetByClientID()` | AI agent persistence, client_id lookup |
| `ThirdpartyOAuth2ServiceRepository` | CRUD + `FindByProtectedResource()`, `CountGrantsReferencingService()` | OAuth2 provider config with encrypted secrets |
| `UserGrantRepository` | CRUD + `FindByPrincipalAndAgent()`, `ListByPrincipal()`, `DeleteByAgent()`, `CountAgentsByServiceID()`, `ListByServiceID()` | User delegation grants |
| `UserSessionRepository` | CRUD + `FindByPrincipalAndService()`, `ListByPrincipal()`, `DeleteByPrincipalAndService()`, `CountByService()` | Encrypted OAuth2 token sessions |

DTOs defined here: `User`, `UserFilter`
DTOs from domain: Repositories reference `storage.Agent`, `storage.ThirdpartyOAuth2Service`, `storage.UserGrant`, `storage.UserSession` from `domain/storage/`.

Sentinel error: `ErrNotFound`

## Rules

### Interface Segregation Principle (ISP)
Repositories expose **only the operations their consumers need**. Each repository is a focused interface — not a monolithic data access layer. New operations must justify their existence by a specific consumer need.

### No Business Logic
Ports must **never** contain:
- Validation logic (that belongs in domain)
- Orchestration (that belongs in domain services)
- Error handling beyond sentinel definitions
- Conditional logic or branching

### Adding New Ports
- New port interfaces require an ADR if they introduce a **new architectural boundary**
- New methods on existing ports require consumer justification
- DTOs in ports should be minimal — prefer domain types from `domain/storage/` where possible
- Port DTOs exist only when the domain type doesn't fit the interface contract

### Encryption Context Convention
`EncryptionPort.Encrypt/Decrypt` accept `map[string]string` encryption context (AAD). By convention (ADR 008), this contains only `{"service_id": "<id>"}`. Never store secrets in encryption context.

### Error Conventions
Storage repositories return `storage.StorageError` (from domain) with appropriate `ErrorKind`. Port-level sentinel `ErrNotFound` exists for quick identity checks. Encryption operations return `encryption.EncryptionError` (from domain).
