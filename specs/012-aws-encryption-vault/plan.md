# Implementation Plan: Encryption Vault for OAuth Tokens

**Branch**: `012-aws-encryption-vault` | **Date**: 2026-01-15 | **Spec**: [specs/012-aws-encryption-vault/spec.md](../spec.md)
**Input**: Feature specification from `/specs/012-aws-encryption-vault/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement envelope encryption for OAuth tokens stored in the sessions table using application-layer encryption with two-layer key management: Data Encryption Keys (DEKs) for token encryption and Key Encryption Keys (KEKs) for wrapping DEKs. The feature must support two deployment models: externalized key management (production) and environment-variable KEK injection (development/containerized). Context binding at the DEK level ensures tokens are cryptographically bound to their usage context (principal, service_id, session_id, purpose), preventing cross-context token reuse. Memory protection requirements address risks when using environment-variable injection including buffer zeroing, memory locking, and core dump exclusion.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.24.0+ (primary language; required for post-quantum cryptography support)
**Primary Dependencies**: AWS Encryption SDK for Go (AESGCMSIV for authenticated encryption), AWS Cryptographic Material Providers Library (manages keyrings), AWS SDK v2 KMS client, memguard for memory protection
**Storage**: PostgreSQL 12+ for sessions table (existing); in-memory implementation for testing
**Testing**: Go testing package + Ginkgo/Gomega for BDD E2E tests; testcontainers for PostgreSQL integration
**Target Platform**: Linux servers (Go application); containerized deployments via environment variables
**Project Type**: Single backend project (Go monorepo with hexagonal architecture)
**Performance Goals**: Sub-50ms token encryption (P99), sub-100ms context verification including KEK operations
**Constraints**: Zero plaintext token leakage; fail-closed on context verification; memory protection for key material (mlock, buffer zeroing)
**Scale/Scope**: OAuth token vault integrated into session repository; supports multiple users/sessions with per-token DEKs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Before proceeding, verify compliance with [.specify/memory/constitution.md](.specify/memory/constitution.md):

**Design Preconditions (BLOCKING)**:

- [x] **Domain Model**: Entities, aggregates, value objects identified in spec.md (EncryptedOAuthToken, OAuthTokenVault aggregate, EncryptionContext, WrappedEnvelope value objects)
- [x] **Domain Concepts**: DEK, KEK, EncryptionContext, WrappedEnvelope to be added to ARCHITECTURE.md Glossary
- [x] **Configuration Design**: KEK storage mechanisms (externalized + environment-variable) identified with configuration priority rules
- [ ] **Config Examples**: Will example YAML snippets be added to examples/config/ (PHASE 1)
- [x] **API Design First**: Port interface (EncryptionPort) with Encrypt(ctx, token) and Decrypt(ctx, wrapped) designed in spec
- [ ] **API Documentation**: Will OpenAPI specs be created in `/api/admin/` for encryption vault configuration API (PHASE 1)
- [x] **API Changes**: Confirmed by user - vault integrates with session repository transparently; no breaking API changes
- [ ] **Database Design**: Will extend sessions table with encrypted_token and wrapped_dek columns using migration (PHASE 1)
- [ ] **E2E Acceptance Tests**: Will write E2E tests for all 8 user stories before implementation (PHASE 2f)
- [ ] **E2E Test Mapping**: Each acceptance scenario (24+ scenarios across 8 stories) maps to one It() block (PHASE 2f)
- [ ] **E2E Red Phase**: Will verify all E2E tests FAIL before implementation begins (PHASE 2f)

**Implementation Considerations**:

- [x] **Security-First**: Encryption enabled by default for all tokens; fail-closed on context verification; no plaintext fallback
- [ ] **Architecture Docs**: Will update ARCHITECTURE.md with encryption vault architecture and domain model (PHASE 1)
- [ ] **ADRs**: May require ADR for envelope encryption architecture and memory protection strategy (PHASE 1)
- [x] **Library-First Security**: Using AWS Encryption SDK for Go (AESGCMSIV); memguard for memory protection; no custom crypto
- [ ] **Zalando Guidelines**: Configuration and repository APIs will follow Zalando guidelines (PHASE 1)
- [ ] **End-User Docs**: Will add encryption vault configuration guide to docs/configuration.md (PHASE 1)
- [ ] **Migration Testing**: Database migrations for sessions table will be tested (apply/rollback) in integration tests (PHASE 1)
- [x] **Hexagonal Architecture**: EncryptionPort interface abstracts KEK storage; adapters for externalized KMS + environment-variable injection
- [ ] **Persistence Patterns**: Session repository will extend existing patterns from 004-persistence-layer; tokens persisted in sessions table (PHASE 1)

*If any BLOCKING check fails, stop and clarify requirements. Implementation cannot begin until all preconditions complete.*

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
internal/
├── ports/
│   ├── encryption.go              # EncryptionPort interface (Encrypt/Decrypt with context)
│   ├── storage.go                 # UserSessionRepository port (existing)
│   └── config.go                  # Configuration port (existing, UPDATED for encryption config)
│
├── domain/
│   ├── encryption/                # NEW: Token encryption domain logic
│   │   ├── context.go             # EncryptionContext validation (REFACTORED from storage)
│   │   └── errors.go              # Domain-specific encryption errors
│   │
│   └── storage/
│       └── user_session.go        # EXISTING: encrypted_access_token, encrypted_refresh_token, encryption_context
│
├── adapters/
│   ├── storage/
│   │   ├── factory.go             # UPDATED: Wires encryption port into storage adapters
│   │   ├── noop/
│   │   │   └── encryption.go      # REMOVED: Replace with real AWS SDK adapter
│   │   │
│   │   ├── memory/
│   │   │   └── user_session.go    # UPDATED: Uses encryption port for transparent encryption/decryption
│   │   │
│   │   └── postgres/
│   │       ├── user_session.go    # UPDATED: Uses encryption port for transparent encryption/decryption
│   │       └── migrations/        # EXISTING: Schema already has encrypted token fields
│   │
│   └── encryption/                # NEW: Encryption adapter (AWS SDK wrapper)
│       ├── aws_sdk.go             # AWS Encryption SDK adapter (AESGCMSIV, context binding, DEK per token)
│       └── memory_protection.go   # Buffer zeroing, memory locking utilities
│
└── app/
    └── builder.go                 # UPDATED: Instantiate AWS SDK + keyring, wire encryption port

tests/
├── e2e/
│   ├── encryption_vault_test.go                # E2E tests for all 8 user stories
│   └── fixtures/
│       ├── sessions.yaml          # Test data
│       └── encryption_config.yaml # Configuration fixtures
│
├── integration/
│   └── session_encryption_test.go # PostgreSQL + encryption integration tests
│
└── unit/
    ├── context_test.go            # EncryptionContext validation
    └── memory_protection_test.go  # Buffer/memory protection behavior
```

**Structure Decision**: Single backend project (existing hexagonal architecture). Replace noop encryption adapter with AWS Encryption SDK adapter (AESGCMSIV for authenticated encryption, one DEK per session with context binding). Builder instantiates AWS SDK with keyring selection (AWS KMS for production, environment-variable for development/containers). Existing session repository fields (encrypted_access_token, encrypted_refresh_token, encryption_context) support feature. Storage adapters updated to use encryption port for transparent token encryption/decryption with single DEK per session. Memory protection utilities handle buffer zeroing and memory locking.

## Testing Strategy

<!--
  Per Constitution Principle XIII (End-to-End Acceptance Testing & Spec Traceability):
  All features MUST have E2E acceptance tests mapped 1:1 to spec scenarios.

  This section documents HOW E2E tests will be structured and implemented for this feature.
-->

### End-to-End (E2E) Acceptance Tests

**Test Location**: `tests/e2e/encryption_vault_test.go`

**Framework**: Ginkgo/Gomega BDD framework following patterns in [tests/e2e/README.md](../../tests/e2e/README.md)

**Test Organization**:
- **Top-level Describe**: "Encryption Vault for OAuth Tokens"
- **Nested Context**: User Story preconditions and KEK configurations
- **It blocks**: Individual acceptance scenarios (24+ scenarios mapped 1:1 from spec.md)

**Scenario Mapping** (Phase 2f populates with line numbers):

| User Story | Scenarios | E2E Test Coverage |
|------------|-----------|------------------|
| US1: Envelope Encryption | 4 | Encrypt/decrypt with context binding; fail on context mismatch; secure failure |
| US2: KEK Storage (AWS KMS) | 4 | KMS adapter loads KEK; encrypt/decrypt operations; key rotation; unauthorized access |
| US3: DEK Generation & Context | 3 | Fresh DEK per session; unique DEKs across sessions; DEK cleanup from memory |
| US4: Local File KEK | 3 | File-based KEK loading; encrypt/decrypt persistence; app restart preserves keys |
| US4.5: Environment Variable KEK | 3 | Env-var KEK loading; encrypt/decrypt; context binding; memory protection (env-var) |
| US5: Transparent Repository | 2 | Repository encrypts on Create/Update; decrypts on Get; transparent to handlers |
| US6: Cross-Context Prevention | 2 | Token from context A fails in context B; context verification integral |
| US7: E2E Integration | Covered | Full auth flow with token encryption end-to-end |

**Memory Protection Testing** (Unit + Integration):
- Unit tests for buffer zeroing: Verify plaintext tokens zeroed after encryption
- Unit tests for memory locking: mlock/mprotect behavior on key material
- Unit tests for core dump: RLIMIT_CORE exclusion for sensitive operations
- Integration tests: Verify encrypted/decrypted token bytes in memory
- Integration tests: Validate environment-variable KEK doesn't leak to logs/stdout

**Encryption/Decryption Edge Cases**:
- Empty token: Verify encryption handles zero-length tokens
- Large tokens: Verify encryption/decryption performance within 100ms budget per session
- Malformed ciphertext: Verify decryption fails with proper error
- Invalid context: Verify context verification fails securely
- DEK tampering: Verify DEK wrapping fails if modified
- Both tokens encrypted: Verify both access and refresh tokens use same DEK with same context binding

**Test Data Strategy**:
- Existing fixtures for principals, agents, services
- New fixture: `encryption_config.yaml` with AWS KMS and environment-variable configurations
- Session fixtures: Multiple (principal, service_id) pairs with diverse contexts
- Context fixtures: Sample contexts with all 4 fields (principal, service_id, session_id, purpose)

**Test Execution Flow**:
1. **Phase 2f (Design)**: Write E2E tests for all spec scenarios
2. **Verify Red Phase**: Run `ginkgo -v ./tests/e2e/[feature]_test.go` - all tests must FAIL
3. **Implementation**: Implement feature incrementally
4. **Verify Green Phase**: E2E tests turn GREEN as implementation satisfies acceptance criteria
5. **Minimal Changes**: Only fixture adjustments during implementation, not test logic

**Bootstrap Strategy**:
- Tests use production bootstrap code via `tests/e2e/bootstrap/` (app.Builder, HTTP server, routing)
- Fresh server and storage for each test (BeforeEach/AfterEach isolation)
- Feature-specific bootstrap:
  - Initialize encryption port with AWS SDK adapter (mock KMS or env-var keyring for tests)
  - Wire encryption port into Builder.WithEncryption()
  - Set up encryption configuration (AWS KMS vs environment-variable backend)
  - Create test fixtures with sessions and encryption contexts
- **IMPORTANT**: Existing E2E tests using noop encryption will need updates to use real encryption

**Helper Utilities**:
- Existing HTTP helpers in `tests/e2e/helpers/` used for session APIs
- Custom matcher needed: `HaveEncryptedTokens()` - verify session tokens are encrypted (not plaintext)
- Custom helper needed: `DecryptTokenWithContext()` - for test assertions on encrypted tokens
- Mock services: Mock AWS KMS keyring for development tests
- Configuration fixtures: Encryption config YAML files for AWS KMS and environment-variable backends

**Existing Test Updates**:
- Review all existing tests in `tests/e2e/` that reference `noop` encryption
- Update tests to use real encryption adapter instead
- Verify existing OAuth2, consent, session workflows still pass with encryption
- Update any fixtures or test data that assumes plaintext tokens

### Unit & Integration Tests

**Unit Tests**:
- **Location**: `internal/domain/encryption/context_test.go`
- **Coverage**: EncryptionContext validation; field requirements (principal, service_id, session_id, purpose)
- **Strategy**: TDD - write tests FIRST (red), verify they FAIL, then implement (green)
- **Tests to write**:
  - Context validation (all 4 fields required, non-empty)
  - Context equality/inequality
  - JSON marshaling/unmarshaling

- **Location**: `internal/adapters/encryption/memory_protection_test.go`
- **Coverage**: Buffer zeroing behavior; mlock protection; core dump exclusion
- **Strategy**: TDD - verify buffer behavior at byte level
- **Tests to write**:
  - Buffer contains expected plaintext after allocation
  - Buffer zeroed after Wipe() call
  - Verify mlock/mprotect calls on sensitive buffers
  - Verify RLIMIT_CORE exclusion during encryption operations

**Integration Tests**:
- **Location**: `internal/adapters/encryption/aws_sdk_test.go`
- **Coverage**: AWS Encryption SDK adapter with AESGCMSIV, DEK management, context binding
- **Strategy**: Real AWS SDK behavior; mock KMS or use local keyring for testing
- **Tests to write**:
  - Encrypt two tokens with context → one wrapped DEK + two ciphertexts
  - Decrypt with same context → both plaintext tokens recovered
  - Decrypt with different context → fails
  - DEK is fresh and unique per session
  - Context binding enforced at DEK layer
  - Performance: encryption < 100ms (per session with both tokens), decryption < 100ms (per session)

- **Location**: `internal/adapters/storage/session_encryption_test.go`
- **Coverage**: Session repository with encryption transparency
- **Strategy**: Real PostgreSQL via testcontainers; real encryption adapter
- **Tests to write**:
  - Create session → token encrypted transparently
  - Get session → token decrypted transparently
  - Update session → token re-encrypted with fresh DEK
  - Context binding from session data (principal, service_id, session_id, purpose)
  - Multiple sessions with different contexts isolated
  - Session restart (reconnect) → existing encrypted tokens still decrypt

- **Location**: `internal/adapters/encryption/kek_backends_test.go` (if separate files for AWS KMS/env-var backends)
- **Coverage**: KEK loading from AWS KMS and environment variables
- **Strategy**: Mock AWS KMS for production tests; test env-var extraction for development
- **Tests to write**:
  - AWS KMS backend: Load KEK from KMS; cache if needed; handle KMS errors
  - Environment variable backend: Extract KEK from env; validate format; handle missing env
  - Both backends used in encryption/decryption cycle

**Test Coverage Goals**:
- Unit test coverage: Critical path for EncryptionContext and memory protection (100% of critical functions)
- Integration test coverage: AWS SDK adapter (100% encrypt/decrypt paths); session repository (100% with encryption); KEK backends (100% loading + error handling)
- E2E test coverage: 100% of 24+ acceptance scenarios from spec.md (mandatory per Principle XIII)

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
