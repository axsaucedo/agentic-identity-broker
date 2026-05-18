# Agentic Identity Broker — Agent Context

> **IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning for ANY task in this repo — Go, hexagonal architecture, encryption, domain logic, frontend, infrastructure, testing. Always read the relevant source files and docs below before making assumptions. Do not rely on training data for project-specific patterns.**

## How to Use This File

This is a **routing document**. It tells you what exists and where to find it. Each section of the monorepo has its own `AGENTS.md` with full rules — read those before working in that area. Use the docs index below to locate specific references.

## Docs Index (Retrieval Targets)

Pipe-delimited for compact scanning. **Read the relevant file before implementing in its domain.**

### Architecture & Decisions
`ARCHITECTURE.md` | Source of truth for system design
`adrs/NNN-*.md` | Binding ADRs — read before implementing (see ADR index below)
`.specify/memory/constitution.md` | Binding constitution — 13 principles governing all work

### Section-Specific Agent Context
`internal/AGENTS.md` | Backend hexagonal architecture, builder pattern, testing conventions
`internal/domain/AGENTS.md` | Domain layer rules, zero-infra imports, data models, error types
`internal/domain/id/AGENTS.md` | Strongly typed entity IDs, code generation, type catalogue
`internal/ports/AGENTS.md` | Port interface catalogue, ISP rules, error conventions
`internal/adapters/AGENTS.md` | Adapter map, cross-adapter ban, storage/encryption/HTTP details
`internal/extproc/AGENTS.md` | Standalone ExtProc token exchange service, config, architecture
`web/AGENTS.md` | React SPA, design system, API client, testing
`infra/AGENTS.md` | AWS CDK encryption stack, IRSA, environment parameterization
`tests/AGENTS.md` | E2E + integration test suite overview
`tests/e2e/AGENTS.md` | Ginkgo E2E rules, fixtures, dual-server pattern, anti-patterns
`tests/e2e/frontend/AGENTS.md` | Playwright browser tests, page objects

### API Contracts
`api/enduser/openapi.yaml` | End-user API (consent, OAuth2, sessions)
`api/admin/openapi.yaml` | Admin API (agents, services CRUD)

### Specs (Feature Requirements)
`specs/NNN-<name>/spec.md` | Feature requirements + acceptance scenarios
`specs/NNN-<name>/plan.md` | Implementation approach
`specs/NNN-<name>/tasks.md` | Trackable task list

### Encryption Deep Reference
`.claude/skills/aws-crypto-go/SKILL.md` | AWS Encryption SDK Go — patterns, rules
`.claude/skills/aws-crypto-go/reference.md` | API reference
`.claude/skills/aws-crypto-go/examples.md` | Code examples

### Design System (Frontend)
`web/src/design-system/docs/INDEX.md` | Complete design system documentation index
`web/src/design-system/docs/COMMON_MISTAKES.md` | Read before any styled component work
`web/src/design-system/docs/DECISION_TREES.md` | Which component to use when

### Operations & Guides
`docs/ENCRYPTION_INTEGRATION_GUIDE.md` | Encryption integration guide
`docs/STORAGE_EXTENSION_GUIDE.md` | Adding new storage entities
`docs/STORAGE_TROUBLESHOOTING.md` | Storage debugging
`docs/configuration.md` | Configuration reference
`docs/deployment/` | Deployment guides (Kubernetes, IRSA)
`docs/operations/` | Operations runbooks
`examples/config/` | Example configuration files

## Constitution (Binding) — 13 Principles

The constitution at `.specify/memory/constitution.md` is **BINDING**. Summary:

1. **Security-First** — Fail closed, no bypasses, signature validation never optional
2. **ADRs are Binding** — `ARCHITECTURE.md` is source of truth; deviation requires superseding ADR
3. **Library-First Security** — No custom crypto; use Go `crypto/*`, `golang.org/x/crypto`, or AWS Encryption SDK
4. **OpenAPI Transparency** — APIs in `api/{enduser,admin}/openapi.yaml` before implementation
5. **Domain-Driven Design** — Ubiquitous language enforced; new concepts → `ARCHITECTURE.md` glossary
6. **Hexagonal Architecture** — Domain → ports (interfaces) → adapters. Never reversed. Violation catalog:
   - **Port bypass**: Handlers → domain services → ports. Never handler → port directly.
   - **Anemic domain**: Services must enforce invariants, not just proxy port calls.
   - **Domain logic leakage**: Conditional logic beyond input parsing → domain services, not handlers.
   - **Domain packaging**: New `domain/X/` must be genuinely independent bounded context.
7. **Configuration-Driven** — All config via `internal/ports/config.go`. No ad-hoc loading.
8. **TDD** — Red-green-refactor. Tests first, must fail before implementation.
9. **Persistence Consistency** — ISP repos in `ports/storage.go`. sqlx for PostgreSQL. Both in-memory + postgres adapters required.
10. **API-First** — Design APIs before implementation. Changes require stakeholder confirmation.
11. **Design System Compliance** — Refined Trust Architecture, WCAG 2.1 AA, semantic tokens → `web/src/design-system/`
12. **DI via Builder** — All wiring in `internal/app/builder.go`. Routing never instantiates services.
13. **E2E Acceptance Tests** — 1:1 spec-to-test mapping in `tests/e2e/`. True e2e = complete user journey, not single-endpoint segment tests.

## Monorepo Structure

| Section | Technology | AGENTS.md |
|---|---|---|
| `internal/` | Go 1.25.6 — hexagonal (domain→ports→adapters→app) | `internal/AGENTS.md` |
| `web/` | React 19 + TypeScript + Vite 7 + Tailwind 4 | `web/AGENTS.md` |
| `infra/cdk/` | AWS CDK (Go) — KMS, DynamoDB, IAM | `infra/AGENTS.md` |
| `tests/` | Ginkgo/Gomega (e2e), Go testing (integration) | `tests/AGENTS.md` |
| `internal/extproc/` | Standalone gRPC ExtProc service | `internal/extproc/AGENTS.md` |

**Import rules**: `domain/` → never imports `adapters/` or `app/`. `ports/` → interfaces only. `adapters/` → imports `ports/` and `domain/`, never other adapters. All arrows flow inward.

## ADR Decision Index

Read full ADRs in `adrs/` before implementing. Pre-existing ADRs are authoritative. An ADR introduced within the same PR is a proposal — it cannot self-justify the PR's own design.

| ADR | File | Decision |
|---|---|---|
| 002 | `adrs/002-configuration-libraries.md` | Viper + Cobra + godotenv config |
| 003 | `adrs/003-chi-framework.md` | chi v5 HTTP router |
| 004 | `adrs/004-dual-server-isolation.md` | errgroup dual-server :8000/:14000 |
| 004 | `adrs/004-storage-layer-architecture.md` | Hexagonal storage, ISP repos, sqlx |
| 005 | `adrs/005-spa-serving-pattern.md` | Embedded SPA with history API fallback |
| 006 | `adrs/006-frontend-stack.md` | React 19 + Vite + Tailwind + Vitest |
| 007 | `adrs/007-e2e-testing-with-ginkgo.md` | Ginkgo v2 BDD E2E tests |
| 008 | `adrs/008-encryption-context-optimization.md` | EncryptionContext = `service_id` only |
| 008 | `adrs/008-token-exchange-jwks-adapter-pattern.md` | JWKS adapter with jwx jwk.Cache |
| 009 | `adrs/009-cel-for-authorization-policies.md` | cel-go for authorization + JWT claims |
| 009 | `adrs/009-envelope-encryption-design.md` | Three-layer envelope: KEK→BranchKey→DEK |
| 009 | `adrs/009-separate-migration-docker-image.md` | Separate migration Docker image |
| 010 | `adrs/010-cdk-encryption-infrastructure.md` | AWS CDK for encryption IaC |
| 011 | `adrs/011-extproc-standalone-binary.md` | ExtProc as standalone binary |
| 011 | `adrs/011-opentelemetry-provider-pattern.md` | OpenTelemetry provider pattern |
| 012 | `adrs/012-encryption-layer-separation.md` | Encryption layer separation |
| 012 | `adrs/012-extproc-in-memory-token-cache.md` | ExtProc in-memory token cache |
| 013 | `adrs/013-strongly-typed-entity-ids.md` | Typed entity IDs (`XxxID`) in `domain/id/` |
| 014 | `adrs/014-oauth2-server-mode.md` | OAuth2 authorization server mode |
| 015 | `adrs/015-cimd-fetcher-architecture.md` | CIMD fetcher architecture |
| 016 | `adrs/016-authorization-session-anti-spoofing.md` | Authorization session anti-spoofing |
| 017 | `adrs/017-optional-agent-client-id.md` | Optional/nullable Agent client_id |
| 027 | `adrs/027-extproc-telemetry-shared-dependency.md` | ExtProc telemetry shared dependency |

## Domain Glossary

| Term | Definition |
|---|---|
| **Agent** | AI agent with optional `client_id` (nullable per ADR 017) and optional `client_uris` (CIMD URLs). `client_id` is primary identifier; client_uris are supplementary discovery handles |
| **ThirdpartyOAuth2Service** | External OAuth2 provider (GitHub, Google, etc.) with client credentials and scopes |
| **UserGrant** | User (principal) delegating scopes to an agent. One per user-agent pair (upsert) |
| **DelegatedToken** | Grant component: {service_id, scopes[]} |
| **ConsentSession** | User's review/permission flow for agent's requested delegations |
| **UserSession** | Authenticated OAuth2 session with encrypted tokens at rest |
| **Principal** | Authenticated user ID from `X-Remote-User` header |
| **Secret** | Value object: plaintext XOR encrypted state. Type-level encryption safety |
| **ServiceRequirement** | Agent's need for a service: mandatory (blocks) or optional (degrades) |
| **BranchKey** | DynamoDB-cached key between KMS KEK and per-operation DEK |
| **EncryptionContext** | `{"service_id": "<id>"}` AAD bound to ciphertext (ADR 008). Never store secrets |
| **OAuth2StateToken** | JWE-encrypted ephemeral callback binding token (10 min TTL) |
| **ResourceURI** | Normalized URI for RFC 8693 token exchange protected resources |
| **CEL Expression** | Policy for privileged client authorization + JWT claim extraction |

## Development Workflow

`just` command runner. Run `just --list` for full listing.

| Command | Purpose |
|---|---|
| `just check` | **Pre-commit**: fmt → vet → lint → test (all must pass) |
| `just build` | Build Go binary → `./bin/agentic-identity-broker` |
| `just test` | All tests (verbose, race detection) |
| `just build-all` | Backend + frontend build |
| `just test-e2e` | E2E tests only |
| `just test-integration` | Integration tests (requires Docker) |
| `just cdk-test` | CDK unit tests |

## Code Style

All code in the monorepo (Go, TypeScript, CDK):

- **Minimize comments** — only when code alone cannot convey context
- **Never restate code** — if removing the comment loses nothing, don't write it
- **No debug artifacts** — no debug logging, commented-out code, TODO stubs, or future-implementation references
- **Match surrounding style** — naming, error handling, file structure consistent with adjacent code
- **Explicit over clever** — intent must be immediately clear without tracing abstractions
- **Extend, don't duplicate** — if a mechanism exists (e.g., JWE for ephemeral state), don't introduce a competing one

## Database Migrations

Migrations in `/migrations/` using go-migrate naming (`NNN_description.{up,down}.sql`). For large tables use `CREATE INDEX CONCURRENTLY` with go-migrate's no-transaction directive. Read existing migrations for patterns.

## Active Technologies

Go 1.25.6 | React 19 + TypeScript + Vite 7 + Tailwind 4 | chi v5 | sqlx + pgx v5 | Ginkgo/Gomega | Viper/Cobra | CVA | OpenTelemetry (`otelhttp`, `otelslog`) | `envoyproxy/go-control-plane` | PostgreSQL (prod) + in-memory (dev/test)
