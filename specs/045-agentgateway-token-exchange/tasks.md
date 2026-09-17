---
description: "Task list for Agentgateway Native Token Exchange implementation"
---

# Tasks: Agentgateway Native Token Exchange

**Input**: Design documents from `/specs/045-agentgateway-token-exchange/`

**Prerequisites**: `plan.md` (required), `spec.md` (required), `research.md`, `data-model.md`,
`contracts/`, and `quickstart.md`.

**Tests**: Principle VIII and Principle XIII require tests. Write all nine Ginkgo acceptance specs in
Phase 2f. Make them compile and fail semantically before feature support or artifacts turn them green.

**Organization**: Tasks are grouped by user story. The direct path is a separate E2E package. It has
no `internal/extproc` import, no `extProc` policy, and no Compose dependency.

## Format

- **`[P]`**: This task can run in parallel. It uses different files and has no unfinished dependency.
- **`[USn]`**: This task belongs to user story `n`.
- Every task has an exact file path.

## Scope Notes

- **Phase 0 is skipped.** The feature needs no production-code refactor.
- **Phase 2.7 is skipped.** The feature adds no persisted entity, typed ID, port, repository, handler,
  or migration.
- **No Broker API changes.** Reuse `POST /oauth2/token` exactly as documented in
  `api/enduser/openapi.yaml`.
- **No Compose changes.** Leave `docker-compose.yml`, `mocks/agentgateway/config.yaml`, and all
  ExtProc assets unchanged. Research R9 explains why.
- **No frontend work.** Do not change `web/`, `tests/e2e/frontend/`, or screenshot artifacts.
- **Keep the E2E process isolated.** The gateway suite uses `--procs=1` because it temporarily adds
  the HTTPS-JWKS fixture CA to a cloned `http.DefaultTransport`.

---

## Phase 1: Setup (Shared Test Infrastructure)

**Purpose**: Create the isolated test package, pinned schema input, dependency declaration, and
commands before Phase 2f writes acceptance tests.

- [X] T001 Create `tests/e2e/gateway/gateway_suite_test.go` as an isolated `gateway_test` Ginkgo suite named `Agentgateway Native Token Exchange`; register `Fail`, use `RunSpecs`, and import no `internal/extproc` package.
- [X] T002 [P] Vendor agentgateway `v1.5.0` `schema/config.json` at `tests/e2e/gateway/testdata/agentgateway-config-v1.5.0.schema.json`; preserve SHA-256 `22129d32cb367eab78cf0c4e62cf15758513c0e40aa7aa4ce5401790efc25250`.
- [X] T003 [P] Promote `github.com/santhosh-tekuri/jsonschema/v6` at its pinned v6.0.2 version to a direct test dependency in `go.mod`; update only required entries in `go.sum`.
- [X] T004 [P] Add `test-e2e-gateway`, `test-e2e-gateway-coverage`, and gateway coverage aggregation in `justfile`; run `ginkgo -v --procs=1 ./tests/e2e/gateway/` and add the suite to `test-e2e`, `test-e2e-coverage`, and `test-e2e-watch`.

**Checkpoint**: The repository has a dedicated gateway E2E target. It cannot share the root backend
suite process or the ExtProc suite package.

---

## Phase 2: Design Preconditions (Blocking Prerequisites) [MANDATORY]

**Purpose**: Complete the constitution preconditions before feature implementation.

**Critical rule**: Do not implement the direct route, E2E environment, or operator guide until all
Phase 2 sections are complete.

### Phase 2a: Domain Model & Glossary [MANDATORY]

**Constitution Reference**: Principles II and V.

- [X] T005 Confirm the in-flight configuration and credential model in `specs/045-agentgateway-token-exchange/data-model.md`; record that no domain entity, aggregate, typed ID, storage port, or migration is required.
- [X] T006 [P] Update the gateway and glossary descriptions in `ARCHITECTURE.md` with `Native Gateway Exchange Policy` and `Gateway Client Assertion`; distinguish direct native exchange from the existing ExtProc path.
- [X] T007 [P] Create `adrs/036-agentgateway-native-token-exchange.md` as a Proposed ADR; record mutually exclusive route policies, one client-assertion issuer per Broker instance, the unchanged ExtProc path, and links to ADRs 007, 011, and 029.

**Checkpoint**: The architecture uses one vocabulary for native and ExtProc exchange. The new ADR does
not supersede an accepted ADR or authorize an API change.

### Phase 2b: Configuration Design [MANDATORY]

**Constitution Reference**: Principle VII.

- [X] T008 Examine `internal/ports/config.go` and `specs/045-agentgateway-token-exchange/contracts/broker-direct-path-config.md`; make sure that the direct path uses existing `token_exchange.*` settings and adds no Broker configuration field.
- [X] T009 [P] Create the matching reference configurations in `examples/agentgateway/direct-token-exchange.yaml` and `examples/config/token-exchange-direct-gateway.yaml`: use a v1.5.0 native MCP route with sibling `path: /oauth2/token`, one `resources` URI, `privateKeyJwt`, the explicit E2E trust tuple, `signingKey.file`, and existing Broker `token_exchange.*` settings; omit `extProc`, `grantType`, shared secrets, inline private keys, and HTTP client-assertion JWKS URLs.
- [X] T010 Update `examples/config/README.md` to link `token-exchange-direct-gateway.yaml`; state that it configures existing Broker fields and that operators replace the trust tuple as one unit.
- [X] T011 [P] Examine `charts/agentic-identity-broker/values.yaml`, `values.schema.json`, and `templates/configmap.yaml`; record in the PR that no Helm change is needed because no configuration parameter changes.

**Checkpoint**: Both committed reference configurations exist before Phase 2f. They use existing
Broker settings only and add no TLS bypass or HTTP client-assertion JWKS URI.

### Phase 2c: API Design [MANDATORY]

**Constitution Reference**: Principles IV and X.

- [X] T012 Examine `api/enduser/openapi.yaml` and `specs/045-agentgateway-token-exchange/contracts/broker-token-exchange-request.md`; record in `adrs/036-agentgateway-native-token-exchange.md` and the PR that the feature reuses `POST /oauth2/token` without an API change or new stakeholder approval.

**Checkpoint**: The direct route uses the existing RFC 8693 contract. Do not edit either OpenAPI file.

### Phase 2d: Database Design [MANDATORY]

**Constitution Reference**: Principle IX.

- [X] T013 Confirm in `specs/045-agentgateway-token-exchange/data-model.md` and `adrs/036-agentgateway-native-token-exchange.md` that the direct path reuses protected-resource mappings, grants, and sessions; do not add files under `migrations/` or edit `internal/ports/storage.go`.

**Checkpoint**: The feature creates no database schema, repository, or storage-adapter work.

### Phase 2e: Frontend and Design-System Review [MANDATORY — N/A]

**Constitution Reference**: Principle XI.

- [X] T014 Record in `adrs/036-agentgateway-native-token-exchange.md` that this feature changes no `web/` path; do not add Playwright tests or screenshots under `tests/e2e/frontend/` and `tests/e2e/screenshots/`.

**Checkpoint**: The frontend review is complete. No UI acceptance scenario exists.

### Phase 2f: E2E Acceptance-Test Design [MANDATORY]

**Constitution Reference**: Principle XIII. Write these tests before the direct-route implementation.

- [X] T015 Implement a functional red-phase baseline in `tests/e2e/gateway/support/agentgateway_container.go`: accept the committed reference YAML and documented substitutions, but deliberately render a valid agentgateway v1.5.0 MCP route with neither `backendAuth.oauthTokenExchange` nor `extProc`; start a reachable MCP recorder and forward the inbound bearer so every acceptance spec fails on observable behavior, not setup.
- [X] T016 [P] Write the six Ginkgo `It()` blocks in `tests/e2e/gateway/native_token_exchange_e2e_test.go` for US1-S1 through US2-S3; use `Context` names `Exchange Tokens Through the Native Gateway Policy` and `Preserve Broker Authorization Boundaries`, add a nearby `// USn-Sn from specs/045-agentgateway-token-exchange/spec.md` comment to every block, and use fresh `BeforeEach` state.
- [X] T017 [P] Write the three Ginkgo `It()` blocks in `tests/e2e/gateway/reference_config_e2e_test.go` for US3-S1 through US3-S3; use the `Deploy the Alternative Safely` `Context`, read the committed reference YAML and existing `docs/guides/token-exchange-gateway.md` first, assert its absent direct-path link in the red state, and read `docs/guides/token-exchange-gateway-direct.md` only after that link exists.
- [X] T018 Run `just test-e2e-gateway` from `justfile`; make sure that all nine specs compile and fail semantically because the valid baseline route forwards the inbound bearer, emits no Broker form, or the existing guide lacks direct-path instructions — not because of a harness error, file-load error, compilation error, `Skip`, `XIt`, `PIt`, or placeholder assertion.

**Acceptance-spec content for T016**:

- **US1-S1**: Record two real `POST /oauth2/token` forms. Assert the RFC 8693 grant, `subject_token`, access-token type, configured `resource`, jwt-bearer assertion type, decoded assertion `iss == sub == https://agentgateway-direct-e2e.example.test`, `aud == token-exchange-broker`, and a distinct `jti` for each exchange.
- **US1-S2**: Assert successful MCP completion. The downstream recorder sees the exchanged token and never the inbound token.
- **US1-S3**: Assert the direct route sends an RFC 8693 request to the Broker, has no `extProc` policy, and gives an exposed ExtProc stand-in listener zero connections.
- **US2-S1**: Keep one `It()` and use subcases for an untrusted signing key, mismatched assertion issuer, mismatched assertion audience, a subject-token issuer different from the configured upstream fixture issuer, and wrong subject-token audience. Assert the documented status and zero downstream requests. Assert an unsupported algorithm as gateway configuration-load rejection, not assertion tampering.
- **US2-S2**: Use a valid assertion and subject JWT but omit the active `UserGrant`. Assert the Broker returns `access_denied` with 403, the MCP route returns agent-visible 500, and the backend sees no request.
- **US2-S3**: Use valid direct routes with a missing resource, an unmapped resource, no stored session, insufficient stored-session scope, unavailable client-assertion JWKS, and an unavailable Broker endpoint. Assert the Broker returns `invalid_request` with 400 for the missing resource, `invalid_target` with 400 for the unmapped resource, and `invalid_grant` with 400 for each stored-session condition. Assert the Broker returns `server_error` with 500 for unavailable JWKS. The MCP route returns agent-visible 500 for every case, including an unavailable Broker. Assert zero backend requests and no inbound-token fallback for every subcase.

**Acceptance-spec content for T017**:

- **US3-S1**: Read `examples/agentgateway/direct-token-exchange.yaml` and the existing
  `mocks/agentgateway/config.yaml`. First assert that `docs/guides/token-exchange-gateway.md` lacks
  the direct-path link; after the link exists, read
  `docs/guides/token-exchange-gateway-direct.md` and assert that the paths are alternatives,
  `privateKeyJwt` is required, and shared-secret methods are excluded. This sequence keeps the red
  failure observable rather than turning it into a file-loading error.
- **US3-S2**: Pass the committed direct reference configuration to the red-phase renderer after only
  documented substitutions. Assert an exchange succeeds, `signingKey.file` supplies the key, and
  neither a client secret nor inline PEM reaches the rendered configuration or logs. The valid baseline
  deliberately serves its no-exchange route, so the backend observes the original token and this
  assertion fails on behavior rather than a source-file or setup error.
- **US3-S3**: Assert the expected guide verification result. The downstream credential must differ from
  the inbound credential and the ExtProc stand-in must observe zero connections. In the red state the
  valid baseline forwards the inbound credential.

**Checkpoint**: All nine acceptance scenarios have one executable `It()`. The tests have realistic
assertions and an observed semantic red phase.

---

## Phase 2.5: Foundational E2E Infrastructure [BLOCKING]

**Purpose**: Replace the functional red-phase baseline with test-only infrastructure that exercises the
production Broker boundary and a real agentgateway container.

- [X] T019 Extend `tests/e2e/bootstrap/test_server.go` with an opt-in container-reachable builder mode: bind `0.0.0.0:0`, expose `http://host.testcontainers.internal:<port>` separately from host `BaseURL()`, and accept an outer handler wrapper; keep `NewTestServerBuilder` defaults unchanged, use LSP references before the edit, and make the existing callers in `tests/e2e/oauth2_metadata_test.go` stay green.
- [X] T020 [P] Implement `tests/e2e/gateway/support/https_jwks_server.go` with `httptest.NewTLSServer`, a gateway public JWKS endpoint, and request accounting; clone `http.DefaultTransport`, add only the fixture CA to `RootCAs`, keep `InsecureSkipVerify` false, and restore the original transport in cleanup.
- [X] T021 [P] Implement `tests/e2e/gateway/support/signing_key.go` with a fresh RSA signing keypair, JWKS generation, 0600 temporary private-key files, and subject-token helpers; reuse `tests/e2e/helpers/jwt_helpers.go` and `lestrrat-go/jwx/v4` rather than custom crypto.
- [X] T022 [P] Implement `tests/e2e/gateway/support/downstream_backend.go` with a `0.0.0.0:0` streamable-MCP `whoami` backend that records each Authorization header and a separate TCP ExtProc stand-in that records connection attempts without importing `internal/extproc`.
- [X] T023 [P] Implement `tests/e2e/gateway/support/recording_token_endpoint.go` as an outer handler wrapper that copies and restores the `POST /oauth2/token` body, records parsed form values, and delegates unchanged to the production Broker router.
- [X] T024 [P] Replace the red-phase baseline route in `tests/e2e/gateway/support/agentgateway_container.go` with the pinned v1.5.0 testcontainers helper: honor `AGENTGATEWAY_IMAGE`, validate rendered YAML with `testdata/agentgateway-config-v1.5.0.schema.json`, mount the key file, register Broker/MCP/stand-in ports through `HostAccessPorts`, wait for port 4000, and prove route liveness.
- [X] T025 Replace the red-phase baseline in `tests/e2e/gateway/support/agentgateway_container.go` with per-`It()` direct-route orchestration: use `bootstrap.NewStorageFactory(logger).NewTestStorage()`, production `app.Builder`, `fixtures.OAuth2ConfigWithTokenExchange`, `ValidAgent`, `GitHubService`, `SeedPlaceholderGrantData`, `ActiveGrant`, and `GitHubSessionForPrincipal`; configure the existing upstream fixture issuer and HTTPS client-assertion JWKS trust tuple; disable only the E2E gateway cache; and close servers, listeners, containers, temporary keys, storage, and the transport override in cleanup.

**Checkpoint**: Every test starts a fresh real Broker boundary, real agentgateway v1.5.0 container,
HTTPS JWKS fixture, and downstream recorder. The foundation contains no ExtProc service, mock Broker,
or residual passthrough baseline.

---

## Phase 3: User Story 1 — Exchange Tokens Through the Native Gateway Policy (Priority: P1)

**Goal**: An operator can configure a direct native route that exchanges an inbound agent credential at
the Broker and forwards only the exchanged credential to an MCP backend.

**Independent Test**: Operate the three US1 specs in `tests/e2e/gateway/native_token_exchange_e2e_test.go`.
The real `v1.5.0` gateway sends the expected form and the backend receives only the exchanged token.

### Acceptance Tests for User Story 1

The US1 specs were written in T016 and observed red in T018. Do not weaken their form, token-replacement,
or no-ExtProc assertions.

### Implementation for User Story 1

- [X] T026 [US1] Implement reference-config loading and static invariant checks in `tests/e2e/gateway/support/agentgateway_container.go`; read the committed `examples/agentgateway/direct-token-exchange.yaml`, require the v1.5.0 native policy shape, and reject `extProc`, shared-secret, JWT-bearer, and inline-private-key variants before substitution.
- [X] T027 [US1] Complete source-config rendering in `tests/e2e/gateway/support/agentgateway_container.go`: substitute only documented endpoint, resource, key-file, `kid`, and test-host placeholders into `examples/agentgateway/direct-token-exchange.yaml`; add `cache.maxEntries: 0` only to the E2E rendering (as required by the pinned v1.5.0 schema) and retain production cache defaults in the committed reference.
- [X] T028 [US1] Run `ginkgo -v --procs=1 --focus "Exchange Tokens Through the Native Gateway Policy" ./tests/e2e/gateway/` from `tests/e2e/gateway/`; make US1-S1, US1-S2, and US1-S3 green with the recorded form, downstream-token replacement, and zero ExtProc-contact assertions.

**Checkpoint**: US1 works through the native path. The gateway directly calls the Broker and never
starts, configures, or contacts ExtProc.

---

## Phase 4: User Story 2 — Preserve Broker Authorization Boundaries (Priority: P1)

**Goal**: The direct route preserves signed client-assertion validation, subject-token validation,
resource authorization, delegation checks, and fail-closed behavior.

**Independent Test**: Operate the US2 specs in `tests/e2e/gateway/native_token_exchange_e2e_test.go`.
Each failure reaches no protected backend, and the MCP route returns an agent-visible 500 response.

### Acceptance Tests for User Story 2

The US2 specs were written in T016 and observed red in T018. Keep US2-S1 as one `It()` with subcases.

### Implementation for User Story 2

- [X] T029 [US2] Extend `tests/e2e/gateway/support/agentgateway_container.go` and `tests/e2e/gateway/support/signing_key.go` with per-case rendering for an untrusted signer, wrong client ID, wrong assertion audience, a subject issuer different from the configured upstream fixture issuer, wrong subject audience, missing resource, unmapped resource, absent `UserGrant`, no stored session, insufficient stored-session scope, unavailable client-assertion JWKS, unavailable Broker host, and unsupported `clientAuth.alg` startup rejection; preserve the real Broker for every runtime case.
- [X] T030 [US2] Wire the existing US2-S1 table and US2-S2/US2-S3 scenarios in `tests/e2e/gateway/native_token_exchange_e2e_test.go` to those options. Retain one `It()` per spec scenario. Assert the Broker returns `invalid_request` with 400 for a missing resource, `invalid_target` with 400 for an unmapped resource, and `invalid_grant` with 400 for missing or insufficient stored-session scope. Assert the Broker returns `access_denied` with 403 for no delegation and `server_error` with 500 for unavailable client-assertion JWKS. Assert the MCP route returns agent-visible 500 for every error, including an unavailable Broker. Assert zero downstream requests and no inbound-token fallback.
- [X] T031 [US2] Run `ginkgo -v --procs=1 --focus "Preserve Broker Authorization Boundaries" ./tests/e2e/gateway/`; the focused US2 run observed green for the credential-validation, resource, no-delegation, stored-session, JWKS, and Broker-unavailable scenarios without a mock Broker or an ExtProc fallback.

**Checkpoint**: US2 rejects invalid or unavailable exchanges before a protected backend receives a
request. The direct path stays fail-closed.

---

## Phase 5: User Story 3 — Deploy the Alternative Safely (Priority: P2)

**Goal**: Operators can choose the direct route deliberately, configure its signed identity correctly,
and prove the backend receives an exchanged credential without ExtProc.

**Independent Test**: Operate the three US3 specs in `tests/e2e/gateway/reference_config_e2e_test.go`.
The test reads the committed guidance and boots the same reference configuration.

### Acceptance Tests for User Story 3

The US3 specs were written in T017 and observed red in T018. Keep artifact assertions specific to
committed paths and actual configuration values.

### Implementation for User Story 3

- [X] T032 [P] [US3] Create `docs/guides/token-exchange-gateway-direct.md`; document the direct replacement route, v1.5.0 `path` and `resources` fields, RFC 8693 form, `privateKeyJwt` key source, exact trust tuple, HTTPS JWKS requirement, one-anchor route selection, fail-closed results, and downstream-token validation steps.
- [X] T033 [P] [US3] Update `docs/guides/token-exchange-gateway.md` to identify its ExtProc sidecar path as the existing alternative; link `docs/guides/token-exchange-gateway-direct.md`, prohibit combining policies on one route, and preserve all ExtProc deployment instructions.
- [X] T034 [P] [US3] Update `docs/concepts/token-exchange.md` to describe direct native exchange and ExtProc exchange as separate paths; state which credential validates at the Broker and that neither path forwards the inbound credential after a rejected exchange.
- [X] T035 [P] [US3] Update `docs/reference/token-exchange.md` with the native gateway form fields, `privateKeyJwt` assertion claims, `resources` mapping, direct-path error status behavior, and links to `examples/agentgateway/direct-token-exchange.yaml` and `examples/config/token-exchange-direct-gateway.yaml`.
- [X] T036 [US3] Run `ginkgo -v --procs=1 --focus "Deploy the Alternative Safely" ./tests/e2e/gateway/`; make US3-S1, US3-S2, and US3-S3 green by reading the committed guides, booting the committed reference YAML, and observing an exchanged downstream credential with zero ExtProc contacts.

**Checkpoint**: The committed guide and configuration identify the two paths as alternatives. The source
file operators copy is the source file the real gateway E2E suite operates.

---

## Phase N: Constitution Compliance and Polish [MANDATORY COMPLIANCE SECTION]

**Purpose**: Make sure that the implemented feature and its evidence meet every applicable constitution
requirement before merge.

### Design-Phase Verification [MANDATORY]

- [X] T037 Make sure that `ARCHITECTURE.md`, `adrs/036-agentgateway-native-token-exchange.md`, and `specs/045-agentgateway-token-exchange/data-model.md` agree on the direct-route model and one-anchor rule (Principles II and V).
- [X] T038 Make sure that `examples/config/token-exchange-direct-gateway.yaml` and `examples/config/README.md` document only existing Broker settings; make sure that `charts/agentic-identity-broker/` has no required change (Principle VII).
- [X] T039 Make sure that `api/enduser/openapi.yaml`, `api/admin/openapi.yaml`, and `migrations/` remain unchanged; record the unchanged API and database contract in `adrs/036-agentgateway-native-token-exchange.md` (Principles IV, IX, and X).
- [X] T040 Make sure that `tests/e2e/gateway/native_token_exchange_e2e_test.go` and `tests/e2e/gateway/reference_config_e2e_test.go` contain exactly nine traced `It()` blocks, one per US1-S1 through US3-S3, and preserve the T018 semantic-red evidence (Principles VIII and XIII).
- [X] T041 Make sure that no `web/`, `tests/e2e/frontend/`, or `tests/e2e/screenshots/` path changed; record the backend-only scope in `adrs/036-agentgateway-native-token-exchange.md` (Principle XI).

### Implementation-Phase Verification [MANDATORY]

- [X] T042 Audit `examples/agentgateway/direct-token-exchange.yaml` and `tests/e2e/gateway/support/agentgateway_container.go` against the vendored v1.5.0 schema; require `path`, default RFC 8693 grant, `resources`, `privateKeyJwt`, key file source, no shared secret, no `extProc`, and no JWT-bearer/OBO mode (Principles I and VII).
- [X] T043 Audit `tests/e2e/gateway/support/https_jwks_server.go` and `tests/e2e/gateway/support/signing_key.go`; require HTTPS JWKS, an additive fixture CA, restored default transport, `InsecureSkipVerify == false`, no custom cryptography, and no logged key or token value (Principles I and III).
- [X] T044 Audit `tests/e2e/bootstrap/test_server.go` and `tests/e2e/gateway/support/recording_token_endpoint.go`; retain production `app.Builder` and routing, preserve existing builder callers, and keep recording as pass-through observation rather than a mock Broker (Principles VI, XII, and XIII).
- [X] T045 Audit `tests/e2e/gateway/`; remove the T015 passthrough baseline, test-only temporary files, debug logging, commented-out code, `TODO` markers, skipped specs, and red-phase annotations before merge (Principles VIII and XIII).
- [X] T046 Run `just check` from `justfile`, then resolve every formatting, vet, and lint problem introduced by the feature.
- [X] T047 Run `just test` from `justfile` after T046. Make sure that the fast Go/package suite stays green.
- [X] T048 Run `just test-e2e-gateway` from `justfile`. Make sure that all nine native gateway acceptance specs pass with `--procs=1`.
- [X] T049 Run `just test-e2e-extproc` from `justfile`. Make sure that the existing ExtProc route and behavior stay green and unchanged (FR-009).
- [X] T050 Run `just verify` from `justfile`. Make sure that the repository verification gate, including the aggregate E2E suite, passes.

**Checkpoint**: The direct path, documentation, and proof satisfy the specification without changing the
Broker API, Broker configuration schema, storage schema, ExtProc deployment, or frontend.

---

## Dependencies and Execution Order

### Phase Dependencies

```text
Phase 1 Setup
    -> Phase 2 Design Preconditions (2a through 2f)
        -> Phase 2.5 E2E Foundation
            -> US1 Native Exchange (MVP)
                -> US2 Failure Boundaries
            -> US3 Safe Deployment Documentation
                -> Phase N Compliance and Verification
```

- **Phase 1** creates the test package, command, schema, and test dependency.
- **Phase 2** blocks all feature work. T016 and T017 write all acceptance specs before T019 through
  T025 implement their environment.
- **Phase 2.5** blocks every user story. It creates the real-Broker, real-gateway E2E path.
- **US1** must complete before US2 because they share the direct reference configuration and its
  renderer.
- **US3** depends on the US1 reference configuration and the Phase 2 Broker example. It can proceed in
  parallel with US2 after US1 completes.
- **Phase N** follows all user stories.

### User Story Dependencies

| Story | Depends on | Why |
|---|---|---|
| US1 — Native Exchange | Phase 1, Phase 2, Phase 2.5 | Provides the direct route and successful exchange baseline. |
| US2 — Broker Boundaries | US1 | Reuses the direct configuration and per-test container environment. |
| US3 — Safe Deployment | US1, Phase 2b | Tests the committed direct reference file and matching Broker example. It can run with US2. |

### Parallel Opportunities

- **Setup**: T002, T003, and T004 can proceed with T001.
- **Design**: T006 and T007 can proceed after T005. T009 and T011 can proceed after T008. T012 through
  T014 are independent design confirmations.
- **E2E foundation**: T020 through T024 touch separate files and can proceed together after T018.
- **US3**: T032 through T035 touch separate documentation files and can proceed together after T026.
- **Do not parallelize** T025, T027 through T031, T036, or T042 through T050. They integrate shared
  test infrastructure or form the verification chain.

## Parallel Example: User Story 3

```text
Task: "Create the direct operator guide in docs/guides/token-exchange-gateway-direct.md"
Task: "Update the ExtProc guide in docs/guides/token-exchange-gateway.md"
Task: "Update the concept guide in docs/concepts/token-exchange.md"
Task: "Update the field reference in docs/reference/token-exchange.md"
```

Each task changes a different document. Start T036 only after all four documentation tasks complete.

## Implementation Strategy

### MVP First: User Story 1

1. Complete Phase 1 and every Phase 2 precondition.
2. Write and record the semantic-red acceptance tests in T015 through T018.
3. Complete the E2E foundation in T019 through T025.
4. Complete T026 through T028.
5. Operate the three US1 tests independently. Stop if a test exposes ExtProc contact or inbound-token
   forwarding.

### Incremental Delivery

1. Deliver US1: real native exchange with direct Broker form inspection and downstream credential
   replacement.
2. Deliver US2: invalid assertion, denied delegation, and unavailable Broker all fail closed.
3. Deliver US3: committed configuration and guidance match the proven direct route.
4. Complete the Phase N verification chain before merge.

### Parallel Team Strategy

After Phase 2.5, one engineer can complete US2 while another completes US3. Keep the US1 owner on the
shared direct configuration and test renderer until T028 is green.

## Task Summary

| Area | Tasks |
|---|---:|
| Setup | 4 |
| Design Preconditions | 14 |
| Foundational E2E Infrastructure | 7 |
| US1 — Native Exchange | 3 |
| US2 — Broker Boundaries | 3 |
| US3 — Safe Deployment | 5 |
| Constitution Compliance and Polish | 14 |
| **Total** | **50** |

**Story-owned task count**: US1 = 3, US2 = 3, US3 = 5. Phase 2f contains two shared test-authoring
tasks that create all nine scenario-specific `It()` blocks.

## Requirement Traceability

| Requirement | Primary tasks |
|---|---|
| FR-001, FR-004, FR-016 | T009, T026, T027, T042 |
| FR-002, FR-003, FR-005 | T016, T020–T025, T027, T028 |
| FR-006, FR-007, FR-014, FR-017 | T016, T019–T025, T029–T031, T043 |
| FR-008, FR-009 | T016, T022, T026, T033, T049 |
| FR-010, FR-011, FR-018 | T009, T010, T032–T036, T042 |
| FR-012, FR-013, FR-015 | T001–T004, T015–T025, T048 |
| SC-001 | T016, T022, T027–T028, T036 |
| SC-002 | T016, T029–T031 |
| SC-003 | T016–T018, T040, T048 |
| SC-004 | T009–T010, T032–T033, T036 |
| SC-005 | T009–T010, T032–T035 |

Every acceptance scenario maps to one `It()` in T016 or T017. T040 makes the final one-to-one mapping
reviewable before merge.
