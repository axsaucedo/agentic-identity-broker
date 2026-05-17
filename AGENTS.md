# Agentic Identity Broker — Agent Context

**IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning for any Go, hexagonal architecture, encryption, or domain-specific tasks. Always read project files before making assumptions.**

## Constitution (Binding)

The constitution at `.specify/memory/constitution.md` is **BINDING**. Key principles:

1. **Security-First** — Security enabled by default, fail closed, no bypasses. Signature validation never optional.
2. **ADRs are Binding** — `ARCHITECTURE.md` is the source of truth. Accepted ADRs in `adrs/` must be followed; deviation requires a superseding ADR.
3. **Library-First Security** — No custom crypto. Use Go `crypto/*`, `golang.org/x/crypto`, or AWS Encryption SDK. Escalate if insufficient.
4. **OpenAPI Transparency** — APIs documented in `/api/enduser/openapi.yaml` and `/api/admin/openapi.yaml` before implementation. Follow Zalando RESTful API Guidelines.
5. **Domain-Driven Design** — Ubiquitous language enforced. New domain concepts must be added to `ARCHITECTURE.md` glossary.
6. **Hexagonal Architecture** — Domain depends on ports (interfaces), never adapters. Adapters implement ports. See `internal/` structure. Violation catalog:
   - **Port bypass**: Handlers must go through domain services. Calling a port interface directly from a handler is still a violation, even if the call target is an interface rather than a concrete type.
   - **Anemic domain**: If removing a domain service and having the handler call the port directly would lose no business rules, the service is anemic and adds no value. Domain services must enforce invariants or orchestrate logic.
   - **Domain logic leakage**: Conditional logic beyond input parsing and error mapping belongs in domain services, not handlers.
   - **Domain packaging**: A new `domain/X/` package must be a genuinely independent bounded context. If its primary interactions are with a single existing domain service, it likely belongs as a sub-package — not a peer package. Separate packages imply independent lifecycles.
7. **Configuration-Driven** — All config via `internal/ports/config.go` port. No ad-hoc config loading.
8. **TDD** — Red-green-refactor. Tests written first, must compile and fail semantically before implementation.
9. **Persistence Consistency** — ISP repositories in `internal/ports/storage.go`. sqlx for PostgreSQL. Migrations in `/migrations/` with go-migrate naming. Both in-memory and postgres adapters required.
10. **API-First** — Design APIs before implementation. API changes require stakeholder confirmation, even for bug fixes.
11. **Design System Compliance** — Refined Trust Architecture aesthetic, WCAG 2.1 AA, semantic tokens. See `web/src/design-system/`.
12. **DI via Builder** — All wiring in `internal/app/builder.go`. Routing functions in `internal/adapters/http/routing/` receive pre-wired deps, never instantiate services.
13. **E2E Acceptance Tests** — 1:1 spec-to-test mapping in `tests/e2e/`. Ginkgo/Gomega BDD. Written before implementation (red-green). A test that hits only one endpoint of a multi-step flow is a **segment test**, not a true e2e test. True e2e tests exercise a complete user journey where each step's output feeds the next step's input.

Full constitution: `.specify/memory/constitution.md`

## Monorepo Structure

| Section | Technology | Role |
|---|---|---|
| `internal/` | Go 1.25.6 | Backend — hexagonal architecture (domain → ports → adapters → app) |
| `web/` | React 19 + TypeScript + Vite 7 + Tailwind | Consent frontend SPA |
| `infra/cdk/` | AWS CDK (Go) | Encryption infrastructure (KMS, DynamoDB, IAM) |
| `tests/` | Ginkgo/Gomega (e2e), Go testing (integration) | E2E and integration test suites |

Each section has (or will have) its own `AGENTS.md` with section-specific rules.

### Backend Package Map (`internal/`)

```
domain/                          # Pure business logic, NO infrastructure imports
  config/                        # Config domain types + validation (LogLevel, LogFormat)
  consent/                       # ConsentService
  encryption/                    # Encryption domain errors
  id/                            # Strongly typed entity IDs — see internal/domain/id/AGENTS.md
  oauth2/                        # OAuth2AuthorizationService
  oauth2session/                 # oauth2session.Service (token vault, PKCE, JWE state)
  principal/                     # User identity context (from X-Remote-User header)
  server/                        # Server lifecycle config
  services/                      # ThirdpartyServiceManager
  storage/                       # Storage domain types (Agent, UserGrant, ThirdpartyOAuth2Service)
  tokenexchange/                 # TokenExchangeService + CEL policy evaluation
ports/                           # Interface definitions ONLY (7 files)
  cel.go, config.go, encryption.go, jwks.go, oauth2.go, server.go, storage.go
adapters/                        # Infrastructure implementations
  encryption/{aws/,branchkey/,memory/}  — EncryptionPort implementations
  http/{handlers/,middleware/,routing/,upstream/,oauth2_sessions/}  — chi router, dual-server
  jwks/                          # JWKS fetching adapter (lestrrat-go/jwx)
  storage/{memory/,postgres/,noop/}  — StoragePort implementations, factory.go
app/                             # DI wiring: Builder pattern (builder.go)
```

**Import rules**: `domain/` → never imports `adapters/` or `app/`. `ports/` → interfaces + DTOs only. `adapters/` → may import `ports/` and `domain/`, never other adapters. All dependency arrows flow inward.

### Typed Entity IDs (`internal/domain/id/`)

Full rules: [`internal/domain/id/AGENTS.md`](internal/domain/id/AGENTS.md) | ADR: [`adrs/013-strongly-typed-entity-ids.md`](adrs/013-strongly-typed-entity-ids.md)

The `id` package provides one named type per entity with a UUID primary key (`AgentID`, `ServiceID`, `GrantID`, `SessionID`, `UserID`) and named string types for non-UUID identifiers (`ClientID`, `ExternalID`, `Principal`). Named types give compile-time safety — the compiler rejects passing a `ServiceID` where an `AgentID` is expected. **Key rules**:
- Use `ParseXxxID` (returns error) in all production code; `MustParseXxxID` (panics) in test fixtures only.
- Check principal (401) *before* UUID format validation (400) in every HTTP handler.
- Every new domain entity with a UUID PK must add its type to `gen_ids.go` and document it in `internal/domain/id/AGENTS.md`.

## ADR Decision Index

Read full ADRs in `adrs/` before implementing in their domain. Pre-existing ADRs are authoritative. An ADR introduced within the same PR is a proposal — it cannot self-justify the PR's own design choices.

| ADR | Decision |
|---|---|
| 002 | Use Viper + Cobra + godotenv for multi-source config with clear precedence |
| 003 | chi v5 as lightweight, stdlib-compatible HTTP router |
| 004 (dual-server) | errgroup atomic startup: end-user :8000, admin :14000 |
| 004 (storage) | Hexagonal storage with swappable in-memory/PostgreSQL backends, ISP repositories, sqlx |
| 005 | Serve React SPA from Go backend via embedded static files with history API fallback |
| 006 | React 19 + react-router-dom 7 + TypeScript + Vite + Tailwind + Headless UI + Vitest |
| 007 | Ginkgo v2/Gomega BDD E2E tests exercising full production bootstrap stack |
| 008 (encryption) | Reduce EncryptionContext to `service_id` only for performance |
| 008 (JWKS) | Dedicated JWKS adapter with lestrrat-go/jwx jwk.Cache behind hexagonal port |
| 009 (CEL) | cel-go for sandboxed gateway authorization policy + JWT claim extraction |
| 009 (encryption) | Three-layer envelope encryption: KEK → Branch Key → DEK via AWS KMS Hierarchical Keyring |
| 009 (migration) | Separate migration Docker image to minimize production attack surface |
| 010 | AWS CDK (Go) for KMS keys, DynamoDB key store, IAM roles as IaC |
| 013 | Strongly typed entity IDs (`type XxxID uuid.UUID`) in `internal/domain/id/` — compile-time cross-entity ID safety |

## Domain Glossary

| Term | Definition |
|---|---|
| **Agent** | AI agent with unique `client_id` when set (canonical identifier, used in OAuth2/consent flows via `GetByClientID`; optional/nullable in the Admin API, and omitted values remain `NULL` per ADR 017) and optional `client_uris` (CIMD URLs, resolved via `GetByClientURI`). Both resolve to the same entity; `client_id` is primary, client_uris are supplementary discovery handles. Has display name, optional service requirements (mandatory/optional) |
| **ThirdpartyOAuth2Service** | External OAuth2 provider (GitHub, Google, etc.) with client credentials and scopes |
| **UserGrant** | User (principal) delegating specific OAuth2 scopes to an agent. One grant per user-agent pair (upsert) |
| **DelegatedToken** | Component of a grant: {service_id, scopes[]} |
| **ConsentSession** | User's review and permission flow for an agent's requested delegations |
| **UserSession** | Authenticated OAuth2 session between user and third-party service (encrypted tokens) |
| **Principal** | Authenticated user ID from upstream proxy header (e.g., `X-Remote-User`) |
| **Secret** | Value object with two mutually exclusive states: plaintext or encrypted. Type-level encryption safety |
| **ServiceRequirement** | Agent's declared need for a third-party service: mandatory (blocks auth) or optional (degrades gracefully) |
| **BranchKey** | Cached key in DynamoDB, sits between KMS KEK and per-operation DEK in the key hierarchy |
| **EncryptionContext** | `map[string]string` AAD bound to ciphertext. Contains `service_id` only (ADR 008). Never store secrets in it |
| **OAuth2StateToken** | JWE-encrypted ephemeral token binding OAuth2 callback to initiating request (10 min TTL) |
| **ResourceURI** | Normalized URI on ThirdpartyOAuth2Service identifying protected resources for RFC 8693 token exchange |
| **CEL Expression** | Common Expression Language policy for privileged client authorization and JWT claim extraction |

## Encryption Critical Rules

These 4 rules are **mandatory** for all encryption work (from `.claude/skills/aws-crypto-go/SKILL.md`):

1. **Commitment Policy**: Always use `RequireEncryptRequireDecrypt`. Never use `ForbidEncryptAllowDecrypt`.
2. **Encryption Context**: Always provide `encryptionContext` map (service_id). Never include secrets. Verify context on decrypt.
3. **Hierarchical Keyring**: Go SDK has no Caching CMM — use AWS KMS Hierarchical Keyring with DynamoDB branch key store.
4. **Explicit Wrapping Keys**: Always specify KMS key ARN explicitly. Avoid discovery mode.

Deep reference: `.claude/skills/aws-crypto-go/SKILL.md` (and `reference.md`, `examples.md` alongside it).

## Spec-Driven Development

Specifications in `specs/NNN-<short name>/` as markdown files:
- `spec.md` — the WHAT (feature requirements, acceptance scenarios)
- `plan.md` — the HOW (implementation approach)
- `tasks.md` — trackable task list (must follow constitution task template)

## Development Workflow

Use `just` command runner for all tasks. Run `just --list` for full listing.

### Build & Run
- `just build` — Build Go binary to `./bin/agentic-identity-broker`
- `just run` — Build and run
- `just dev` — Hot-reload dev server with Air

### Testing
- `just test` — All tests (verbose, race detection)
- `just test-coverage` — HTML coverage report

### Code Quality
- `just check` — Run all checks: `fmt` → `vet` → `lint` → `test`
- `just fmt` / `just vet` / `just lint` — Individual checks

### Frontend (Consent UI)
- `just web-install` — Install npm deps (once after clone)
- `just web-dev` — Vite dev server :3000 (proxies API to Go :8000, injects `X-Remote-User: dev@example.com`)
- `just web-build` — Production bundle to `web/dist/consent/`
- `just build-all` — Build Go backend + frontend

### Pre-commit
```bash
just check  # fmt → vet → lint → test — all must pass
```

## Code Style

These rules apply to all code in the monorepo (Go, TypeScript, CDK).

- **Minimize comments.** Add one only when it provides context that the code itself cannot convey.
- **Comments must never restate what the code already says.** If removing the comment would not confuse a future reader, do not write it.
- **No debug artifacts in committed code.** Never leave debug logging, commented-out code, TODO stubs, or references to future implementations unless explicitly asked.
- **Match the style of surrounding code.** Naming conventions, error handling patterns, and file structure must be consistent with adjacent code.
- **Prefer explicit over clever.** Reviewers (human and AI) must understand intent immediately without tracing abstractions.
- **Prefer extending over introducing parallel mechanisms.** If a problem is already solved (e.g., stateless JWE tokens for ephemeral state), do not introduce a competing mechanism. New mechanisms require strong justification beyond "this feature needed it."

## Database Migrations

Migrations live in `/migrations/` and use the go-migrate naming convention (`NNN_description.{up,down}.sql`).

### PostgreSQL Index Creation in Migrations

Use `CREATE INDEX CONCURRENTLY` for production-safe index creation on large tables. Without `CONCURRENTLY`, PostgreSQL holds an `ACCESS EXCLUSIVE` lock for the duration of the build, blocking all reads and writes.

**Important constraint**: `CONCURRENTLY` cannot be used inside a transaction.

go-migrate wraps migrations in transactions by default. To use `CONCURRENTLY`, the migration must opt out of the transaction wrapper via a no-transaction directive at the top of the file:

```sql
-- migrate:no-transaction
```

(Verify the exact directive for the go-migrate version used in this project by checking `cmd/migrate/` or the go-migrate documentation.)

**Migration 008 note**: Migration 008 (`008_drop_agent_client_id_unique`) used non-concurrent index recreation (`CREATE INDEX IF NOT EXISTS`) without `CONCURRENTLY` because the `agents` table is small at migration time and the transactional safety outweighed the lock duration concern.

**Future guidance**: Migrations that create or recreate indexes on tables expected to be large in production should use no-transaction migrations with `CREATE INDEX CONCURRENTLY` to avoid downtime.

## Active Technologies
- Go 1.25.6 (backend), React 19 + TypeScript + Vite 7 (frontend) + chi v5 (router), sqlx (database), Ginkgo/Gomega (E2E), Viper/Cobra (config), Tailwind CSS v4 + CVA (UI) — no new dependencies required (028-cimd-support)
- PostgreSQL (production) + in-memory (dev/test) — Agent entity extension requires migration 015 (028-cimd-support)

## Recent Changes
- 028-cimd-support: Added Go 1.25.6 (backend), React 19 + TypeScript + Vite 7 (frontend) + chi v5 (router), sqlx (database), Ginkgo/Gomega (E2E), Viper/Cobra (config), Tailwind CSS v4 + CVA (UI) — no new dependencies required
