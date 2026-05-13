# Tasks: Hybrid OAuth Server Modes

**Input**: Design documents from `/specs/030-hybrid-oauth-modes/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Per Constitution Principle VIII (Test-Driven Development & Automated Testing), automated tests are MANDATORY for all features. Test tasks are included in each user story below and MUST be written before or alongside implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 0: Pre-implementation Refactoring — Rename `issue_token` → `local`

**Purpose**: Rename all references from `issue_token` to `local` across config values, strategy types, builder branches, tests, docs, and examples. No behavior change.

- [ ] T001 Rename `issue_token` mode value to `local` in internal/ports/config.go (config enum, validation functions)
- [ ] T002 [P] Rename issueToken strategy types to local in internal/adapters/http/enduser/proceed_strategy.go
- [ ] T003 [P] Rename issueToken strategy types to local in internal/adapters/http/enduser/token_grant_strategy.go
- [ ] T004 Rename issueToken builder branch to local in internal/app/builder.go
- [ ] T005 [P] Update config examples in examples/config/ to use `mode: local` instead of `mode: issue_token`
- [ ] T006 [P] Update tests referencing `issue_token` in tests/e2e/mode_configuration_e2e_test.go
- [ ] T007 [P] Update any other test files referencing `issue_token` mode name
- [ ] T008 Verify all existing tests pass after renaming (zero behavior changes)

**Checkpoint**: Refactoring complete, all existing tests pass, no behavior changes introduced

---

## Phase 1: Setup

**Purpose**: No new dependencies needed. Skip — existing project structure is sufficient.

---

## 🔒 Phase 2: Design Preconditions (Blocking Prerequisites)

**Purpose**: Domain model, configuration, API, and database design MUST all be complete before implementation

**⚠️ CRITICAL**: No code implementation can begin until this entire phase is complete

### Phase 2a: Domain Model & Glossary

- [ ] T009 Add `OAuthServerMode`, `ClientMode`, `ModeStrategy` domain terms to ARCHITECTURE.md Glossary section
- [ ] T010 [P] Document mode strategy pattern (proxy/local/hybrid acceptance rules) in ARCHITECTURE.md

**Checkpoint**: Domain model documented

### Phase 2b: Configuration Design

- [ ] T011 Create example YAML for hybrid mode in examples/config/oauth2-hybrid-mode.yaml
- [ ] T012 [P] Update examples/config/oauth2-server-mode.yaml to use `mode: local` with nested sections
- [ ] T013 [P] Update examples/config/README.md to reference new hybrid mode configuration
- [ ] T014 [P] Update charts/agentic-identity-broker/values.yaml with nested proxy/local/cimd config structure
- [ ] T014a [P] Update charts/agentic-identity-broker/templates/configmap.yaml to render nested mode config
- [ ] T014b [P] Update charts/agentic-identity-broker/values.schema.json to validate new mode structure (if exists)
- [ ] T014c [P] Update charts/agentic-identity-broker/README.md with new config parameter documentation

**Checkpoint**: Configuration designed with YAML examples and Helm chart updated

### Phase 2c: API Design

- [ ] T015 Update /api/enduser/openapi.yaml to document mode-specific behavior (JWKS availability, metadata differences per mode)
- [ ] T016 [P] Get user/stakeholder confirmation for API behavioral changes (hybrid mode accepting all client modes)

**Checkpoint**: APIs designed and confirmed

### Phase 2d: Database Design

- [ ] T017 Confirm no database changes needed (FR-017 — no new entity fields, classification from existing properties)

**Checkpoint**: No migrations required — confirmed

### Phase 2f: E2E Acceptance Test Design

- [ ] T018 Write E2E acceptance tests for US1 (6 scenarios) in tests/e2e/hybrid_oauth_modes_e2e_test.go
- [ ] T019 [P] Write E2E acceptance tests for US2 (11 scenarios) in tests/e2e/hybrid_oauth_modes_e2e_test.go
- [ ] T020 [P] Write E2E acceptance tests for US3 (4 scenarios) in tests/e2e/hybrid_oauth_modes_e2e_test.go
- [ ] T021 [P] Write unit/architecture tests for US4 (2 scenarios — structural verification: no runtime mode checks) in internal/app/builder_test.go
- [ ] T022 Create test fixtures: proxy agent, local agent, CIMD agent, hybrid mode config in tests/e2e/fixtures/
- [ ] T023 Verify E2E tests FAIL semantically (red phase): detailed expectations present and failing

**Checkpoint**: E2E acceptance tests written and verified to fail semantically before implementation

---

## Phase 2.5: Foundational Infrastructure

**Purpose**: Core domain types and interfaces that ALL user stories depend on

- [ ] T024 Define `ClientMode` type and constants (ProxyClient, CIMDClient, LocalClient) in internal/domain/storage/agent.go
- [ ] T025 Implement `Agent.ClientMode() ClientMode` method on Agent entity in internal/domain/storage/agent.go
- [ ] T026 [P] Write unit tests for Agent.ClientMode() classification logic in internal/domain/storage/agent_test.go
- [ ] T027 Define `OAuthServerMode` type (proxy, local, hybrid) in internal/ports/config.go
- [ ] T028 Define `ModeStrategy` interface (AcceptsClientMode(ClientMode) bool) in internal/domain/oauth2/mode_strategy.go
- [ ] T029 [P] Implement proxyModeStrategy in internal/domain/oauth2/mode_strategy.go
- [ ] T030 [P] Implement localModeStrategy in internal/domain/oauth2/mode_strategy.go
- [ ] T031 [P] Implement hybridModeStrategy in internal/domain/oauth2/mode_strategy.go
- [ ] T032 Write unit tests for all three ModeStrategy implementations in internal/domain/oauth2/mode_strategy_test.go

**Checkpoint**: Foundation ready — agent classification and mode strategies available for all user stories

---

## Phase 3: User Story 1 — Symmetric Mode Naming and Configuration (Priority: P1) 🎯 MVP

**Goal**: Operators configure broker using three symmetric mode names with self-documenting nested config sections. Invalid configs rejected with actionable errors.

**Independent Test**: Configure each mode in isolation, verify startup success with valid config and rejection with clear errors for invalid config.

### Tests for User Story 1

- [ ] T033 [P] [US1] Write unit tests for config validation: proxy mode accepts proxy section, rejects local section in internal/ports/config_test.go
- [ ] T034 [P] [US1] Write unit tests for config validation: local mode accepts local section, rejects proxy section in internal/ports/config_test.go
- [ ] T035 [P] [US1] Write unit tests for config validation: hybrid mode requires both sections in internal/ports/config_test.go
- [ ] T036 [P] [US1] Write unit test for `issue_token` deprecation error in internal/ports/config_test.go

### Implementation for User Story 1

- [ ] T037 [US1] Restructure OAuth2AuthServerConfig with nested Proxy, Local, CIMD sections in internal/ports/config.go
- [ ] T038 [US1] Implement mode validation: proxy requires Proxy section, forbids Local/CIMD in internal/ports/config.go
- [ ] T039 [US1] Implement mode validation: local requires Local section, forbids Proxy in internal/ports/config.go
- [ ] T040 [US1] Implement mode validation: hybrid requires both Proxy and Local sections in internal/ports/config.go
- [ ] T041 [US1] Implement `issue_token` deprecation error with clear message in internal/ports/config.go
- [ ] T042 [US1] Implement cross-mode field rejection with actionable error messages in internal/ports/config.go

**Checkpoint**: US1 fully functional — three modes configurable, invalid configs rejected with clear errors

---

## Phase 4: User Story 2 — Universal Agent Resolution, Classification, and Mode Enforcement (Priority: P1)

**Goal**: Every incoming request goes through universal resolution → classification → mode enforcement. Each mode accepts/rejects client modes.

**Independent Test**: Register agents of each client mode, send requests per mode, verify correct acceptance/rejection.

### Tests for User Story 2

- [ ] T043 [P] [US2] Write unit tests for universal client resolver: URL → GetByClientURI, UUID → Agent.ID, other → reject in internal/domain/oauth2/client_resolver_test.go
- [ ] T044 [P] [US2] Write unit test for CIMD agent UUID rejection (FR-005) in internal/domain/oauth2/client_resolver_test.go
- [ ] T045 [P] [US2] Write unit tests for mode enforcement: proxy rejects local/CIMD, local rejects proxy agents in internal/domain/oauth2/mode_strategy_test.go
- [ ] T045a [P] [US2] Write unit test for JWKS endpoint availability: served in local/hybrid, not served in proxy mode
- [ ] T045b [P] [US2] Write unit test for metadata endpoint: hybrid mode reflects union of capabilities

### Implementation for User Story 2

- [ ] T046 [US2] Implement universal client resolver with format detection (URL/UUID/invalid) in internal/domain/oauth2/client_resolver.go
- [ ] T047 [US2] Add CIMD agent UUID rejection: after resolving by UUID, reject if agent has client_uris in internal/domain/oauth2/client_resolver.go
- [ ] T048 [US2] Integrate ModeStrategy.AcceptsClientMode() check after resolution and classification in the authorization flow (/authorize path)
- [ ] T048a [US2] Integrate ModeStrategy.AcceptsClientMode() check in the token flow (/token path): client credentials grant and authorization code exchange
- [ ] T048b [P] [US2] Write unit tests for mode enforcement on /token path: client credentials and authorization code exchange rejected in wrong mode, accepted in correct mode
- [ ] T049 [US2] Return actionable error messages for mode rejection (local not supported in proxy, proxy not supported in local)
- [ ] T050 [US2] Ensure JWKS endpoint served in local and hybrid modes, not in proxy mode
- [ ] T051 [US2] Ensure OAuth2 metadata endpoint reflects active mode capabilities (hybrid = union)

**Checkpoint**: US2 fully functional — universal resolution, classification, mode enforcement working

---

## Phase 5: User Story 3 — Mode-Specific Feature Gating (Priority: P2)

**Goal**: CIMD available in local and hybrid modes only, rejected in proxy mode at config validation time.

**Independent Test**: Attempt to enable CIMD in each mode — proxy fails, local/hybrid succeed.

### Tests for User Story 3

- [ ] T052 [P] [US3] Write unit test: CIMD enabled in proxy mode → config rejection in internal/ports/config_test.go
- [ ] T053 [P] [US3] Write unit test: CIMD enabled in local/hybrid mode → config accepted in internal/ports/config_test.go

### Implementation for User Story 3

- [ ] T054 [US3] Add CIMD-in-proxy-mode rejection to config validation in internal/ports/config.go
- [ ] T055 [US3] Verify hybrid mode with CIMD: proxy agents unaffected by CIMD (no behavioral change needed — classification handles this)

**Checkpoint**: US3 fully functional — CIMD gated by mode

---

## Phase 6: User Story 4 — Mode as First-Class Architectural Strategy (Priority: P2)

**Goal**: Mode selection drives entire wiring at startup via builder. No runtime `if mode ==` checks in handlers.

**Independent Test**: Inspect builder — each mode produces self-contained strategy set without runtime mode checks in handlers.

### Tests for User Story 4

- [ ] T056 [P] [US4] Write unit test verifying builder produces correct strategy set for each mode in internal/app/builder_test.go

### Implementation for User Story 4

- [ ] T057 [US4] Implement dispatching AuthorizationProceedStrategy for hybrid mode in internal/adapters/http/enduser/proceed_strategy.go
- [ ] T058 [US4] Implement dispatching TokenGrantStrategy for hybrid mode in internal/adapters/http/enduser/token_grant_strategy.go
- [ ] T059 [US4] Update builder to wire hybrid mode: create both proxy+local strategies, wrap in dispatching strategies in internal/app/builder.go
- [ ] T060 [US4] Verify no runtime `if mode ==` checks exist in handler or service code (structural review)

**Checkpoint**: US4 fully functional — mode wiring is pure strategy pattern

---

## 🔒 Phase N: Constitution Compliance & Polish

**Purpose**: Verify constitution requirements and final polish

### 🔒 Constitution Compliance Verification

#### Design Phase Verification

- [ ] T061 Verify domain model design documented in ARCHITECTURE.md Glossary (Principle V)
- [ ] T062 Verify configuration design YAML examples exist in examples/config/ (Principle VII)
- [ ] T063 Verify examples/config/README.md references new config (Principle VII)
- [ ] T064 Verify API designs documented in /api/enduser/openapi.yaml (Principles IV, X)
- [ ] T065 Verify user/stakeholder confirmed API behavioral changes (Principle X)
- [ ] T066 Verify no DB changes needed confirmed (Principle IX)
- [ ] T067 Verify E2E acceptance tests in tests/e2e/ for all 21 behavioral spec scenarios (Principle XIII — US4 structural scenarios verified via unit/review, not E2E)
- [ ] T068 Verify E2E tests failed semantically before implementation (Principle XIII)

#### Implementation Phase Verification

- [ ] T069 [P] Verify API implementation matches confirmed OpenAPI specification
- [ ] T070 Update ARCHITECTURE.md with mode strategy pattern and ClientMode classification
- [ ] T071 [P] Verify configuration uses unified config port (not custom loading)
- [ ] T072 [P] Verify Helm chart updated with new config structure
- [ ] T073 Verify security: mode enforcement is strict, fail closed, no bypasses (Principle I)
- [ ] T074 Verify domain logic uses ports/interfaces — ModeStrategy is domain (AcceptsClientMode), dispatching is adapter (Principle VI)
- [ ] T075 Verify unit tests written first and failed before implementation (Principle VIII)
- [ ] T076 Verify E2E tests map 1:1 to the 21 behavioral spec scenarios (US1-US3); US4's 2 structural scenarios verified via unit tests in builder_test.go (Principle XIII)
- [ ] T077 Run full E2E test suite: `ginkgo -v ./tests/e2e/` (all tests must pass)
- [ ] T078 Run full unit test suite: `just test` (all tests must pass)

### Additional Polish

- [ ] T079 Update docs/configuration.md with new mode names and hybrid mode documentation
- [ ] T080 Remove any remaining references to `issue_token` in documentation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 0 (Rename)**: Can start immediately. MUST complete before Phase 2 begins.
- **Phase 2 (Design Preconditions)**: Depends on Phase 0 completion. BLOCKS all implementation.
- **Phase 2.5 (Foundational Infrastructure)**: Depends on Phase 2 completion. BLOCKS all user stories.
- **Phase 3 (US1)**: Depends on Phase 2.5 completion.
- **Phase 4 (US2)**: Depends on Phase 2.5 completion. Can run in parallel with US1.
- **Phase 5 (US3)**: Depends on US1 (config validation infrastructure).
- **Phase 6 (US4)**: Depends on US2 (resolution + classification must exist for dispatching).
- **Phase N (Polish)**: Depends on all user stories complete.

### User Story Dependencies

- **US1 (P1)**: Independent — config validation only
- **US2 (P1)**: Independent — can run parallel with US1
- **US3 (P2)**: Depends on US1 (extends config validation)
- **US4 (P2)**: Depends on US2 (dispatching uses classification)

### Parallel Opportunities

- Phase 0: T002, T003, T005, T006, T007 can run in parallel
- Phase 2.5: T026, T029, T030, T031 can run in parallel
- US1 tests: T033, T034, T035, T036 can run in parallel
- US2 tests: T043, T044, T045 can run in parallel
- US1 and US2 implementation can proceed in parallel (different files)

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests in parallel:
Task: "Unit test proxy mode config validation in internal/ports/config_test.go"
Task: "Unit test local mode config validation in internal/ports/config_test.go"
Task: "Unit test hybrid mode config validation in internal/ports/config_test.go"
Task: "Unit test issue_token deprecation in internal/ports/config_test.go"

# Then implement sequentially (same file):
Task: "Restructure OAuth2AuthServerConfig in internal/ports/config.go"
Task: "Implement validation rules in internal/ports/config.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 + 2)

1. Complete Phase 0: Rename `issue_token` → `local` (separate PR)
2. Complete Phase 2: Design Preconditions (all areas)
3. Complete Phase 2.5: Foundational Infrastructure (ClientMode, ModeStrategy)
4. Complete Phase 3: User Story 1 — config validation
5. Complete Phase 4: User Story 2 — universal resolution + mode enforcement
6. **STOP and VALIDATE**: Test US1 + US2 independently
7. Deploy/demo if ready

### Incremental Delivery

1. Phase 0 → Rename PR (reviewable independently)
2. Phase 2 + 2.5 → Foundation ready
3. US1 → Config works → Deploy
4. US2 → Resolution works → Deploy
5. US3 + US4 → Feature gating + architectural strategy → Deploy
6. Phase N → Polish and verify
