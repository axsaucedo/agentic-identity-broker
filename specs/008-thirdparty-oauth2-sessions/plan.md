# Implementation Plan: Third-Party OAuth2 Session Management

**Branch**: `008-thirdparty-oauth2-sessions` | **Date**: 2025-12-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-thirdparty-oauth2-sessions/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Users need to authenticate with third-party OAuth2 services through the identity broker. The system will manage OAuth2 authorization code flows with PKCE, securely store encrypted tokens, and provide a UI for users to view and manage their sessions. Technical approach uses the standard Go OAuth2 library for flow orchestration, lestrrat-go/jwx for JWE state tokens, and follows hexagonal architecture with a new domain service (OAuth2SessionService) coordinating the flows.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**:
- `golang.org/x/oauth2` - OAuth2 authorization code flow orchestration with PKCE
- `github.com/lestrrat-go/jwx/v3` - JWE state token creation and validation
- `github.com/go-chi/chi/v5` - HTTP routing (existing)
- `github.com/jmoiron/sqlx` - PostgreSQL adapter (existing pattern)
- Viper/Cobra - Configuration (existing pattern)

**Storage**: PostgreSQL (new `user_sessions` table for encrypted OAuth2 tokens)  
**Testing**: Go standard library `testing`, testify for assertions, testcontainers for PostgreSQL integration tests  
**Target Platform**: Linux server (Docker/Kubernetes deployment)  
**Project Type**: Web backend (Go) + SPA frontend (React/TypeScript)  
**Performance Goals**: See spec.md Success Criteria (SC-001 through SC-007)  
**Constraints**: <200ms p95 for session status API, encrypted tokens at rest, JWE state tokens with 10-minute TTL  
**Scale/Scope**: Multi-user system, one session per user per service, audit logging for security events

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Before proceeding, verify compliance with [.specify/memory/constitution.md](../../.specify/memory/constitution.md):

**Design Preconditions (BLOCKING)**:

- [x] **Domain Model**: Entities identified: `UserSession` (encrypted OAuth2 tokens + metadata), `OAuth2StateToken` (JWE for flow binding). New domain service: `OAuth2SessionService` for flow orchestration.
- [x] **Domain Concepts**: New terms for ARCHITECTURE.md Glossary: UserSession, OAuth2StateToken, OAuth2SessionService, PKCE Flow, JWE State Binding
- [x] **Configuration Design**: Config requirements identified: `third_party_oauth2.jwe_signing_key`, `third_party_oauth2.state_token_ttl`, `third_party_oauth2.pkce_verifier_length`
- [x] **Config Examples**: YAML snippets will be added to `examples/config/third-party-oauth2.yaml`
- [x] **API Design First**: APIs designed in `contracts/` before implementation (authorize, callback, sessions list, terminate)
- [x] **API Documentation**: OpenAPI specs to be added to `/api/enduser/openapi.yaml`
- [x] **API Changes**: API additions confirmed by user in feature spec (new endpoints for OAuth2 session management)
- [x] **Database Design**: Migration `004_create_user_sessions.up.sql` planned with go-migrate naming

**Implementation Considerations**:

- [x] **Security-First**: JWE authenticated encryption for state tokens, PKCE mandatory, encrypted token storage, fail-closed validation
- [x] **Architecture Docs**: ARCHITECTURE.md will be updated with new domain service and glossary terms
- [ ] **ADRs**: May require ADR for OAuth2 library choices (golang.org/x/oauth2 + lestrrat-go/jwx)
- [x] **Library-First Security**: Using vetted libraries - golang.org/x/oauth2 for flows, lestrrat-go/jwx for JWE/JWS, no custom crypto
- [x] **Zalando Guidelines**: APIs follow Zalando RESTful guidelines (existing pattern)
- [x] **End-User Docs**: API documentation in `docs/api/` with OAuth2 flow examples
- [x] **Migration Testing**: Migrations tested (apply/rollback) in PostgreSQL integration tests
- [x] **Hexagonal Architecture**: OAuth2SessionService (domain) uses ports: EncryptionPort (existing), OAuth2SessionRepository (new), StateTokenPort (new)
- [x] **Persistence Patterns**: Following `specs/004-persistence-layer/quickstart.md` patterns for UserSessionRepository

*All BLOCKING checks pass. Proceeding to Phase 0.*

## Project Structure

### Documentation (this feature)

```text
specs/008-thirdparty-oauth2-sessions/
├── plan.md              # This file (/speckit.plan command output)
├── spec.md              # Feature specification
├── research.md          # Phase 0 output - library patterns and decisions
├── data-model.md        # Phase 1 output - domain entities and relationships
├── quickstart.md        # Phase 1 output - implementation guide
├── contracts/           # Phase 1 output - OpenAPI specifications
│   └── oauth2-sessions.yaml
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Backend (Go)
internal/
├── domain/
│   ├── storage/
│   │   ├── user_session.go         # UserSession entity (new)
│   │   └── user_session_test.go    # Entity validation tests
│   └── oauth2session/
│       ├── service.go              # OAuth2SessionService (new domain service)
│       ├── service_test.go         # Domain service unit tests
│       ├── state_token.go          # State token logic (uses jwx)
│       └── errors.go               # Domain-specific errors
├── ports/
│   ├── storage.go                  # + UserSessionRepository interface
│   └── oauth2.go                   # + StateTokenPort interface (new)
├── adapters/
│   ├── http/
│   │   └── oauth2_session_handlers.go  # HTTP handlers for OAuth2 endpoints
│   ├── storage/
│   │   ├── memory/
│   │   │   └── user_session.go     # In-memory adapter
│   │   └── postgres/
│   │       └── user_session.go     # PostgreSQL adapter
│   └── oauth2/
│       ├── state_token.go          # JWE state token adapter (uses jwx)
│       └── http_client.go          # OAuth2 HTTP client adapter

migrations/
├── 004_create_user_sessions.up.sql
└── 004_create_user_sessions.down.sql

# Frontend (React/TypeScript)
web/src/
├── pages/
│   └── ThirdPartySessionsPage.tsx  # Sessions management page
├── components/
│   └── sessions/
│       ├── ServiceSessionCard.tsx  # Service card with session status
│       ├── SessionStatusBadge.tsx  # Status indicator uses StatusIndicators from design system
│       └── TerminateDialog.tsx     # Confirmation dialog
├── hooks/
│   ├── useThirdPartySessions.ts    # Fetch sessions list
│   └── useTerminateSession.ts      # Terminate session mutation
└── services/api/
    └── sessions.ts                 # API client for session endpoints

# Tests
tests/
├── integration/
│   └── oauth2_session_test.go      # End-to-end OAuth2 flow tests
└── unit/
    └── oauth2_session_service_test.go
```

**Structure Decision**: Web application pattern (backend + frontend). Backend follows existing hexagonal architecture with new `oauth2session` domain service. Frontend extends existing React SPA with new sessions page.

## Complexity Tracking

> No Constitution Check violations requiring justification. All design patterns follow established ADRs and quickstart guides.
