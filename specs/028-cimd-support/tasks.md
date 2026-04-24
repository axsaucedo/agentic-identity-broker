# Tasks: Client ID Metadata Document (CIMD) Support

**Input**: Design documents from `/specs/028-cimd-support/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/, quickstart.md

**Tests**: Per Constitution Principle VIII (Test-Driven Development & Automated Testing), automated tests are MANDATORY for all features. Test tasks are included in each user story below and MUST be written before or alongside implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and configuration scaffolding

- [ ] T001 Add `CIMDConfig` struct to `OAuth2AuthServerConfig` in `internal/ports/config.go` with defaults (enabled: false, fetch_timeout: 1s, max_response_bytes: 5120, cache.max_ttl: 1h, cache.min_ttl: 60s)
- [ ] T002 [P] Register Viper defaults for all `oauth2_authorization_server.cimd.*` keys in config initialization
- [ ] T003 [P] Create example YAML config at `examples/config/cimd.yaml` with annotated CIMD configuration
- [ ] T004 [P] Update Helm chart `charts/agentic-identity-broker/values.yaml` with `cimd` block under `oauth2AuthorizationServer`

**Checkpoint**: Configuration scaffolding in place, project compiles

---

## 🔒 Phase 2: Design Preconditions (Blocking Prerequisites)

**Purpose**: Domain model, configuration, API, and database design MUST all be complete before implementation

**⚠️ CRITICAL**: No code implementation can begin until this entire phase is complete

### Phase 2a: Domain Model & Glossary

- [ ] T005 Add CIMD domain terms to ARCHITECTURE.md Glossary: ClientIDMetadataDocument, CIMDCacheEntry, ClientIDMetadataDocumentURL, SSRFBlocklist, BrandPinMismatchDetected, CIMDSecurityFieldChanged
- [ ] T006 [P] Document ClientResolver strategy pattern and CIMDFetcher port in ARCHITECTURE.md

**Checkpoint**: Domain model documented

### Phase 2b: Configuration Design

- [ ] T007 Verify `CIMDConfig` YAML examples committed to `examples/config/cimd.yaml` (from T003)
- [ ] T008 [P] Update `examples/config/README.md` to reference CIMD configuration section

**Checkpoint**: Configuration designed with YAML examples

### Phase 2c: API Design

- [ ] T009 Update `/api/admin/openapi.yaml`: extend AgentRequest/AgentResponse schemas with `client_uris`, `auth_method`, `jwks_uri`
- [ ] T010 [P] Update `/api/enduser/openapi.yaml`: add `client_id_metadata_document_supported` to metadata response, add `cimd_metadata` to consent response, document new error responses for `/oauth2/authorize`
- [ ] T011 Get user/stakeholder confirmation for API design changes (API-001 through API-004)

**Checkpoint**: APIs designed and confirmed

### Phase 2d: Database Design

- [ ] T012 Create migration `migrations/015_add_agent_cimd_fields.up.sql`: add `client_uris TEXT[] NOT NULL DEFAULT '{}'`, `auth_method TEXT`, `jwks_uri TEXT` to agents table
- [ ] T013 [P] Create migration `migrations/015_add_agent_cimd_fields.down.sql`: drop the three columns

**Checkpoint**: Database migrations created

### Phase 2e: Frontend/Design System Review

- [ ] T014 Review `web/src/design-system/docs/INDEX.md` for component selection for CIMD consent components (CS-001–CS-004)
- [ ] T015 [P] Identify semantic tokens for domain badge (trust-deep), localhost warning (warning/danger), advanced details section

**Checkpoint**: Design system usage planned

### Phase 2f: E2E Acceptance Test Design

- [ ] T016 Create CIMD test fixtures in `tests/e2e/fixtures/cimd.go`: `CIMDAgent()`, `ValidCIMDDocument()`, `CIMDConfig()`, mock CIMD HTTPS server helper
- [ ] T017 Write E2E tests in `tests/e2e/cimd_authorization_test.go` for US1 scenarios (5 It blocks: resolve agent from CIMD, client_id mismatch, absent document, redirect_uri mismatch, disabled gate)
- [ ] T018 [P] Write E2E tests in `tests/e2e/cimd_ssrf_test.go` for US2 scenarios (7 It blocks: private IP, loopback, link-local, HTTP scheme, dot segments, oversized, timeout)
- [ ] T019 [P] Write E2E tests in `tests/e2e/cimd_caching_test.go` for US3 scenarios (4 It blocks: cache hit, expired refetch, no error cache, operator TTL override)
- [ ] T020 [P] Write E2E tests in `tests/e2e/cimd_metadata_test.go` for US4 scenarios (2 It blocks: field present when enabled, absent when disabled)
- [ ] T021 [P] Write E2E tests in `tests/e2e/cimd_consent_test.go` for US5 scenarios (4 It blocks: summary+badge+details, localhost warning, expand details, brand mismatch)
- [ ] T022 [P] Write Playwright E2E tests in `tests/e2e/frontend/cimd_consent_test.go` for CS-001–CS-004 with screenshot captures
- [ ] T023 Verify all E2E tests FAIL semantically (red phase): detailed expectations present and failing, no placeholders, no XIt/PIt/Skip markers

**Checkpoint**: E2E tests written and verified to fail before implementation

---

## Phase 2.7: Agent Entity Extension (Boilerplate)

**Purpose**: Extend the existing Agent entity with CIMD fields and add new storage method, isolated from business logic

- [ ] T024 Add `ClientURIs []string`, `AuthMethod *string`, `JwksURI *string` fields to Agent struct in `internal/domain/storage/agent.go`; update `Validate()` and `Copy()`
- [ ] T025 Add `GetByClientURI(ctx context.Context, uri string) (*Agent, error)` to `AgentRepository` interface in `internal/ports/storage.go`
- [ ] T026 [P] Implement `GetByClientURI` on in-memory adapter in `internal/adapters/storage/memory/` with secondary index `map[string]id.AgentID`
- [ ] T027 [P] Implement `GetByClientURI` on postgres adapter in `internal/adapters/storage/postgres/` using `SELECT ... FROM agents WHERE $1 = ANY(client_uris)`
- [ ] T028 Extend `AgentRequest`/`AgentResponse` DTOs with `ClientURIs`, `AuthMethod`, `JwksURI` in `internal/adapters/http/handlers/admin/agents_handler.go`
- [ ] T029 Verify project compiles with all agent entity extensions (`just build`)

**Checkpoint**: Agent entity extended, project compiles, no business logic yet

---

## Phase 2.5: Foundational Infrastructure

**Purpose**: CIMD domain types, port interface, and fetcher adapter that ALL user stories depend on

- [ ] T030 Define `ClientResolver` strategy interface, `CIMDFetcher` port interface, `ClientResolution` DTO, and `CIMDFetchResult` DTO in `internal/ports/cimd.go`
- [ ] T031 [P] Implement `ClientIDMetadataDocumentURL` value object with parse-time validation (scheme, path, fragment, credentials, port, dot-segments) in `internal/domain/cimd/url.go`
- [ ] T032 [P] Implement `SSRFBlocklist` value object with RFC 6890 default ranges + operator extras in `internal/domain/cimd/blocklist.go`
- [ ] T033 [P] Implement `ClientIDMetadataDocument` value object with JSON parsing and validation (client_id match, auth method, redirect URI same-origin, keyword blocklist) in `internal/domain/cimd/document.go`
- [ ] T034 Implement `CIMDCache` (sync.RWMutex-protected map with TTL clamping from HTTP headers) in `internal/domain/cimd/cache.go`
- [ ] T035 Implement SSRF-hardened HTTP fetcher adapter (custom Dialer.Control, no redirects, LimitReader, context timeout) in `internal/adapters/cimd/fetcher.go`
- [ ] T036 [P] Write unit tests for `ClientIDMetadataDocumentURL` in `internal/domain/cimd/url_test.go`
- [ ] T037 [P] Write unit tests for `SSRFBlocklist` in `internal/domain/cimd/blocklist_test.go`
- [ ] T038 [P] Write unit tests for `ClientIDMetadataDocument` in `internal/domain/cimd/document_test.go`
- [ ] T039 [P] Write unit tests for `CIMDCache` in `internal/domain/cimd/cache_test.go`
- [ ] T040 Write integration tests for SSRF-hardened fetcher with injected resolver/dialer and dial-attempt spy in `internal/adapters/cimd/fetcher_test.go`

**Checkpoint**: Foundation ready — all CIMD domain types, port, and adapter in place; user story implementation can begin

---

## Phase 3: User Story 1 — Agent Authenticates with URL-Based Client ID (Priority: P1) 🎯 MVP

**Goal**: An AI agent uses its HTTPS URL as `client_id`; the broker fetches, validates, and presents CIMD metadata on the consent screen

**Independent Test**: Send a valid HTTPS URL `client_id` in an authorization request to a broker backed by a mock CIMD endpoint; verify the consent screen renders CIMD metadata

### Tests for User Story 1

- [ ] T041 [P] [US1] Write unit tests for `CIMDService` orchestration (fetch → validate → cache → audit → return) in `internal/domain/cimd/service_test.go`
- [ ] T042 [P] [US1] Write unit tests for `OpaqueClientResolver` (rejects https:// prefix, resolves UUID) in `internal/domain/oauth2/client_resolver_test.go`
- [ ] T043 [P] [US1] Write unit tests for `CIMDClientResolver` (URL detection → CIMD resolution → fallback to UUID) in `internal/domain/cimd/client_resolver_test.go`

### Implementation for User Story 1

- [ ] T044 [US1] Implement `CIMDService` in `internal/domain/cimd/service.go`: URL validation → cache check → fetch via port → document validation → brand pin check → security field change detection → cache store
- [ ] T045 [US1] Implement `OpaqueClientResolver` in `internal/domain/oauth2/client_resolver.go`: reject `https://`-prefixed client_id with `invalid_client`, parse UUID for opaque IDs
- [ ] T046 [US1] Implement `CIMDClientResolver` in `internal/domain/cimd/client_resolver.go`: detect URL → validate → lookup agent by client URI → fetch/validate CIMD → return resolution with metadata; fall back to UUID for non-URL
- [ ] T047 [US1] Modify `OAuth2AuthorizationService` in `internal/domain/oauth2/service.go` to delegate client resolution to injected `ClientResolver` strategy instead of direct `id.ParseAgentID`
- [ ] T048 [US1] Wire `ClientResolver` strategy in `internal/app/builder.go` based on `cimd.enabled`: `CIMDClientResolver` when true, `OpaqueClientResolver` when false; instantiate CIMDService/fetcher/cache only when enabled
- [ ] T049 [US1] Verify US1 E2E tests in `tests/e2e/cimd_authorization_test.go` turn green

**Checkpoint**: User Story 1 fully functional — URL-based client_id resolves agent and presents CIMD metadata

---

## Phase 4: User Story 2 — SSRF-Hardened Fetcher Blocks Malicious URLs (Priority: P1)

**Goal**: The CIMD fetcher proactively rejects URLs targeting private networks, loopback, link-local, and enforces timeout/size limits

**Independent Test**: Send authorization requests with client_id pointing to blocked IP ranges; verify rejection before TCP connection

### Tests for User Story 2

- [ ] T050 [P] [US2] Write additional fetcher adapter tests for each SSRF category (private, loopback, link-local, HTTP scheme, dot-segments, oversized, timeout) with dial-attempt spy in `internal/adapters/cimd/fetcher_test.go`

### Implementation for User Story 2

- [ ] T051 [US2] Verify SSRF enforcement is complete in fetcher adapter (all 14 categories from SC-002); add any missing checks in `internal/adapters/cimd/fetcher.go` and `internal/domain/cimd/url.go`
- [ ] T052 [US2] Add structured audit logging for SSRF blocks (CIMDFetchBlocked events) in `internal/domain/cimd/service.go`
- [ ] T053 [US2] Verify US2 E2E tests in `tests/e2e/cimd_ssrf_test.go` turn green

**Checkpoint**: All SSRF attack categories individually rejected, adapter tests prove no TCP dial for blocked addresses

---

## Phase 5: User Story 5 — CIMD-Enhanced Consent Screen (Priority: P1)

**Goal**: Consent screen displays CIMD-sourced metadata: summary, domain badge, localhost warning, advanced details

**Independent Test**: Render consent screen for a CIMD-based authorization request; assert each UI element (CS-001–CS-004) is present

### Tests for User Story 5

- [ ] T054 [P] [US5] Write unit/component tests for `CIMDConsentSummary` in `web/src/components/consent/CIMDConsentSummary.test.tsx`
- [ ] T055 [P] [US5] Write unit/component tests for `CIMDDomainBadge` in `web/src/components/consent/CIMDDomainBadge.test.tsx`
- [ ] T056 [P] [US5] Write unit/component tests for `CIMDLocalhostWarning` in `web/src/components/consent/CIMDLocalhostWarning.test.tsx`
- [ ] T057 [P] [US5] Write unit/component tests for `CIMDAdvancedDetails` in `web/src/components/consent/CIMDAdvancedDetails.test.tsx`

### Implementation for User Story 5

- [ ] T058 [US5] Extend consent API response with `cimd_metadata` object: pass CIMD metadata through consent session in backend
- [ ] T059 [P] [US5] Add CIMD-related TypeScript types to `web/src/types/consent.ts`
- [ ] T060 [P] [US5] Implement `CIMDConsentSummary` component (CS-001) in `web/src/components/consent/CIMDConsentSummary.tsx`
- [ ] T061 [P] [US5] Implement `CIMDDomainBadge` component (CS-002) in `web/src/components/consent/CIMDDomainBadge.tsx`
- [ ] T062 [P] [US5] Implement `CIMDLocalhostWarning` component (CS-003) in `web/src/components/consent/CIMDLocalhostWarning.tsx`
- [ ] T063 [P] [US5] Implement `CIMDAdvancedDetails` component (CS-004) in `web/src/components/consent/CIMDAdvancedDetails.tsx`
- [ ] T064 [US5] Integrate CIMD consent components into `web/src/pages/ConsentOverviewPage.tsx` (conditional rendering when `cimd_metadata` present)
- [ ] T065 [US5] Verify US5 E2E tests in `tests/e2e/cimd_consent_test.go` turn green
- [ ] T066 [US5] Verify Playwright tests in `tests/e2e/frontend/cimd_consent_test.go` pass with screenshots captured

**Checkpoint**: Consent screen renders all four CIMD UX elements (CS-001–CS-004)

---

## Phase 6: User Story 3 — CIMD Response Caching Reduces Latency (Priority: P2)

**Goal**: Repeated authorization requests for the same URL-based client_id use cached documents; cache respects HTTP semantics and operator TTL bounds

**Independent Test**: Issue two sequential authorization requests for the same URL-based client_id; verify mock server receives only one request

### Tests for User Story 3

- [ ] T067 [P] [US3] Write unit tests for cache TTL computation (HTTP header parsing, min/max clamping, no-store/no-cache ignored) in `internal/domain/cimd/cache_test.go`

### Implementation for User Story 3

- [ ] T068 [US3] Implement HTTP cache header parsing (Cache-Control max-age, Expires, ETag) and TTL computation with operator min/max clamping in `internal/domain/cimd/cache.go`
- [ ] T069 [US3] Integrate cache TTL computation into `CIMDService` fetch flow in `internal/domain/cimd/service.go`
- [ ] T070 [US3] Verify US3 E2E tests in `tests/e2e/cimd_caching_test.go` turn green

**Checkpoint**: Caching functional — second request within TTL serves from cache

---

## Phase 7: User Story 4 — Authorization Server Advertises CIMD Support (Priority: P3)

**Goal**: The `/.well-known/oauth-authorization-server` metadata endpoint includes `client_id_metadata_document_supported: true` when CIMD is enabled

**Independent Test**: Fetch metadata endpoint and assert field presence when enabled, absence when disabled

### Implementation for User Story 4

- [ ] T071 [US4] Modify OAuth2 metadata handler in `internal/adapters/http/handlers/enduser/oauth2_metadata_handler.go` to include `client_id_metadata_document_supported` field based on `cimd.enabled` config
- [ ] T072 [US4] Verify US4 E2E tests in `tests/e2e/cimd_metadata_test.go` turn green

**Checkpoint**: Metadata advertisement functional

---

## 🔒 Phase N: Constitution Compliance & Polish

**Purpose**: Verify constitution requirements and final polish

### 🔒 Constitution Compliance Verification

#### Design Phase Verification

- [ ] T073 Verify domain model design documented in ARCHITECTURE.md Glossary (Principle V)
- [ ] T074 Verify configuration design YAML examples exist in `examples/config/cimd.yaml` (Principle VII)
- [ ] T075 [P] Verify `examples/config/README.md` references CIMD configuration (Principle VII)
- [ ] T076 Verify API designs documented in `/api/admin/openapi.yaml` and `/api/enduser/openapi.yaml` (Principles IV, X)
- [ ] T077 Verify user/stakeholder confirmed API designs (Principle X)
- [ ] T078 Verify database migration 015 documented and tested (Principle IX)
- [ ] T079 Verify design system review completed for CIMD consent components (Principle XI)
- [ ] T080 Verify E2E acceptance tests in `tests/e2e/` cover all 22 spec scenarios (Principle XIII)
- [ ] T081 Verify E2E tests were verified to FAIL before implementation (red phase) (Principle XIII)
- [ ] T082 Verify Playwright E2E tests in `tests/e2e/frontend/` pass with screenshots (Principle XIII)

#### Implementation Phase Verification

**API & Documentation** (Principles IV, X):
- [ ] T083 [P] Verify API implementation matches confirmed OpenAPI specification exactly
- [ ] T084 [P] Update `docs/api/` with CIMD-specific API documentation

**Architecture & Documentation** (Principle II):
- [ ] T085 Update ARCHITECTURE.md with CIMD architectural changes (new port, new domain package, flow description)
- [ ] T086 [P] Create ADR 015 in `adrs/` for CIMD fetcher architecture (SSRF-hardened HTTP client, in-process caching, hexagonal port)

**Configuration** (Principle VII):
- [ ] T087 [P] Verify configuration uses unified config port (no custom loading)
- [ ] T088 Verify Helm chart updated with CIMD config block

**Database & Persistence** (Principle IX):
- [ ] T089 [P] Verify migration 015 follows sequential numbering
- [ ] T090 [P] Write integration tests for migration 015 apply/rollback in `internal/adapters/storage/postgres/agent_integration_test.go`
- [ ] T091 [P] Verify postgres adapter tested with new fields and `GetByClientURI`

**Security** (Principles I, III):
- [ ] T092 Verify SSRF protection enabled by default and cannot be fully disabled
- [ ] T093 [P] Verify no custom cryptography used (Go stdlib only)
- [ ] T094 [P] Verify structured audit logging for all security-critical operations (SSRF blocks, brand mismatch, security field changes)

**Architecture Patterns** (Principle VI):
- [ ] T095 Verify domain logic depends on ports only (no adapter imports in `domain/cimd/`)

**Testing** (Principle VIII):
- [ ] T096 Verify unit tests written first and failed before implementation (red-green TDD)
- [ ] T097 Verify automated tests included (unit, integration, E2E)

**E2E Acceptance Testing** (Principle XIII):
- [ ] T098 Verify each It() block maps to exactly one acceptance scenario from spec.md
- [ ] T099 Verify E2E tests turned GREEN as implementation satisfied acceptance criteria
- [ ] T100 Run full E2E test suite: `ginkgo -v ./tests/e2e/` (all tests must pass)
- [ ] T101 Run frontend E2E suite: `ginkgo -v ./tests/e2e/frontend/` (all tests must pass)

**Frontend** (Principle XI):
- [ ] T102 Verify CIMD consent components use design system primitives and semantic tokens
- [ ] T103 [P] Verify WCAG 2.1 AA accessibility compliance for CIMD consent components

### Additional Polish

- [ ] T104 Run `just check` (fmt → vet → lint → test) — all must pass
- [ ] T105 Verify zero regression in existing authorization flows (SC-004) — full existing E2E suite green

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Design Preconditions (Phase 2)**: Can proceed in parallel with Phase 1; BLOCKS all implementation
  - Phase 2a–2f can proceed in parallel, but all must complete before Phase 2.7
- **Entity Boilerplate (Phase 2.7)**: Depends on Phase 2 completion; BLOCKS Phase 2.5
- **Foundational Infrastructure (Phase 2.5)**: Depends on Phase 2 + Phase 2.7; BLOCKS all user stories
- **User Stories (Phase 3–7)**: All depend on Phase 2.5 completion
  - US1 (Phase 3): No dependencies on other stories
  - US2 (Phase 4): No dependencies on other stories (SSRF infra built in Phase 2.5)
  - US5 (Phase 5): Depends on US1 (needs CIMD metadata flowing to consent API)
  - US3 (Phase 6): No dependencies on other stories (cache infra built in Phase 2.5)
  - US4 (Phase 7): No dependencies on other stories
- **Polish (Phase N)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Independent — can start after Phase 2.5
- **US2 (P1)**: Independent — can start after Phase 2.5 (parallel with US1)
- **US5 (P1)**: Depends on US1 (needs CIMD metadata in consent session)
- **US3 (P2)**: Independent — can start after Phase 2.5 (parallel with US1/US2)
- **US4 (P3)**: Independent — can start after Phase 2.5 (parallel with all)

### Parallel Opportunities

- Phase 1 setup tasks T001–T004 (different files)
- Phase 2a–2f design tasks (different documents)
- Phase 2.7 memory/postgres adapters T026/T027 (different files)
- Phase 2.5 domain types T031/T032/T033 (different files) and their tests T036/T037/T038/T039
- US1/US2/US3/US4 can proceed in parallel after Phase 2.5 (US5 waits for US1)
- Frontend components T060–T063 (different files)

---

## Parallel Example: Phase 2.5 Foundation

```bash
# Launch all domain types in parallel (different files):
Task: "Implement ClientIDMetadataDocumentURL in internal/domain/cimd/url.go"
Task: "Implement SSRFBlocklist in internal/domain/cimd/blocklist.go"
Task: "Implement ClientIDMetadataDocument in internal/domain/cimd/document.go"

# Launch all unit tests in parallel (different test files):
Task: "Write unit tests in internal/domain/cimd/url_test.go"
Task: "Write unit tests in internal/domain/cimd/blocklist_test.go"
Task: "Write unit tests in internal/domain/cimd/document_test.go"
Task: "Write unit tests in internal/domain/cimd/cache_test.go"
```

## Parallel Example: User Story 5 Components

```bash
# Launch all frontend components in parallel (different files):
Task: "Implement CIMDConsentSummary in web/src/components/consent/CIMDConsentSummary.tsx"
Task: "Implement CIMDDomainBadge in web/src/components/consent/CIMDDomainBadge.tsx"
Task: "Implement CIMDLocalhostWarning in web/src/components/consent/CIMDLocalhostWarning.tsx"
Task: "Implement CIMDAdvancedDetails in web/src/components/consent/CIMDAdvancedDetails.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (config scaffolding)
2. Complete Phase 2: Design Preconditions (all 6 sub-phases in parallel)
3. Complete Phase 2.7: Agent entity extension (boilerplate)
4. Complete Phase 2.5: Foundational Infrastructure (domain types, port, adapter)
5. Complete Phase 3: User Story 1 (core CIMD authorization flow)
6. **STOP and VALIDATE**: Test US1 independently — URL-based client_id resolves and presents metadata
7. Deploy/demo if ready

### Incremental Delivery

1. Setup → Design → Entity Extension → Foundation ready
2. Add US1 (core flow) → Test → Deploy/Demo (MVP!)
3. Add US2 (SSRF hardening) + US5 (consent screen) → Test → Deploy/Demo
4. Add US3 (caching) → Test → Deploy/Demo
5. Add US4 (metadata advertisement) → Test → Deploy/Demo
6. Each story adds value without breaking previous stories
