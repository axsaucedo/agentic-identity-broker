# Implementation Plan: Agentgateway Native Token Exchange

**Branch**: `045-agentgateway-token-exchange` | **Date**: 2026-09-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/045-agentgateway-token-exchange/spec.md`

## Summary

Add a second, ExtProc-free gateway integration path: an agentgateway `v1.5.0`
`backendAuth.oauthTokenExchange` route that calls the Broker's existing `POST /oauth2/token`
RFC 8693 endpoint directly, authenticating with a `privateKeyJwt` client assertion, and forwards only
the exchanged credential to the protected backend.

Nothing in the Broker changes, and nothing in the existing ExtProc deployment changes. The work is a
committed operator reference configuration, operator documentation and an ADR, and a new isolated
Ginkgo E2E suite that drives a real pinned agentgateway container — booting that same reference file —
against the production Broker boundary built by `app.Builder`. No Compose stack is added: the file an
operator copies is the file the suite boots and verifies (research R9).

The design rests on facts verified against the agentgateway `v1.5.0` tag rather than documentation
alone (see [research.md](./research.md)): agentgateway's `privateKeyJwt` emits `iss = sub = clientId`
and `aud = assertionAudience`, which is exactly the tuple the Broker's
`token_exchange.client_assertion.issuer_uri` and `token_exchange.expected_audience` validate; and on
success the policy removes the inbound credential before writing the exchanged one, which is what
makes FR-005 a property of the gateway rather than something we must add.

## Technical Context

**Language/Version**: Go 1.26.8 (Broker, E2E suites); agentgateway standalone YAML (`v1.5.0` schema);
Markdown for operator documentation
**Primary Dependencies**: agentgateway `cr.agentgateway.dev/agentgateway:v1.5.0` (pinned, FR-012);
`testcontainers-go` v0.44.0; Ginkgo v2 / Gomega; `lestrrat-go/jwx` v4.4.0; `mark3labs/mcp-go`;
existing Broker stack (chi v5, in-memory storage for E2E)
**Storage**: none added. The E2E suite uses `bootstrap.NewStorageFactory(logger).NewTestStorage()`
(in-memory). No migration, no new repository, no new entity
**Testing**: new Ginkgo suite `tests/e2e/gateway/` (9 specs, one per acceptance scenario);
`just test-e2e-gateway`; existing suites unchanged
**Target Platform**: Linux/macOS developer machines and CI, with a Docker or Podman runtime for the
agentgateway container
**Project Type**: Go backend monorepo. No frontend change, so no Playwright work and no screenshots
**Performance Goals**: none introduced. The suite keeps the existing container budget: a 30 s
listening-port wait and a 120 s suite context, matching `tests/e2e/extproc/agentgateway_e2e_test.go`
**Constraints**: direct routes carry no `extProc` policy and contact no ExtProc endpoint (FR-008);
no shared-secret client authentication (FR-004, FR-011); fail closed on every Broker rejection
(FR-007); the client assertion and the inbound subject JWT both carry `aud=token-exchange-broker`
(FR-003); certificate verification stays enabled everywhere, including the E2E HTTPS JWKS fixture
(research R3)
**Scale/Scope**: 1 new E2E suite package; 1 committed reference gateway configuration; 1 new operator
guide plus edits to 4 existing documents; 0 Compose changes; 0 production Go changes;
0 configuration keys; 0 API changes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Design Preconditions (BLOCKING)**:

- [x] **Domain Model**: Entities identified in [data-model.md](./data-model.md). All five are
  configuration documents or credentials in flight; no new domain entity, aggregate or value object
- [x] **Domain Concepts**: `Native Gateway Exchange Policy` and `Gateway Client Assertion` will be
  added to the `ARCHITECTURE.md` glossary (data-model.md, "Glossary additions")
- [x] **Entity IDs**: N/A — no new entity with a UUID primary key, so ADR 013 does not apply
- [x] **Configuration Design**: No Broker configuration key is added or changed. The existing
  `token_exchange.*` keys the direct path depends on are documented with values in
  [contracts/broker-direct-path-config.md](./contracts/broker-direct-path-config.md)
- [x] **Config Examples**: `examples/config/` gains a direct-path Broker example that sets the
  existing `token_exchange` keys to the gateway trust tuple
- [x] **Helm Chart**: N/A — no configuration parameter added, changed or removed, so
  `charts/agentic-identity-broker/` needs no edit (verified: the direct path uses
  `expectedAudience` and `clientAssertion.{issuerUri,jwksUri}`, already present at
  `values.yaml:597-643`)
- [x] **API Design First**: N/A — no API change. The direct path uses `POST /oauth2/token` exactly as
  specified today in `api/enduser/openapi.yaml`
- [x] **API Documentation**: N/A — no OpenAPI addition. The gateway-facing usage of the existing
  endpoint is recorded in
  [contracts/broker-token-exchange-request.md](./contracts/broker-token-exchange-request.md)
- [x] **API Changes**: None to confirm. The spec's Assumptions state the Broker contract is unchanged
- [x] **Database Design**: N/A — no schema change, no migration
- [x] **E2E Acceptance Tests**: 9 tests, written before implementation, mapped 1:1 below
- [x] **E2E Test Mapping**: see the Scenario Mapping table in Testing Strategy
- [x] **E2E Red Phase**: assertions target concrete form fields, JWT claims, HTTP status codes and
  backend request counts — see [quickstart.md](./quickstart.md) §2
- [x] **Frontend Playwright E2E**: N/A — no React UI change
- [x] **Frontend Screenshots**: N/A — no React UI change

**Implementation Considerations**:

- [x] **Security-First**: Security defaults stay on. The direct path fails closed on every Broker
  rejection; the reference configuration keeps `security.skip_thirdparty_https_validation` at `false`
  and points `client_assertion.jwks_uri` at an HTTPS endpoint. The E2E fixture serves that JWKS over
  real TLS and adds its certificate as an **additional** trusted root for the suite process, never
  `InsecureSkipVerify` and never a scheme bypass (research R3). Three shortcuts were explicitly
  rejected there: widening the dev scheme flag into `InsecureSkipVerify`, adding a CA-bundle
  configuration key, and serving the fixture JWKS over HTTP
- [x] **Architecture Docs**: `ARCHITECTURE.md` gains the two glossary terms and a short note that the
  gateway integration has two alternative paths
- [x] **ADRs**: A new ADR records the decision to support two mutually exclusive gateway integration
  paths and the single-client-assertion-anchor rule (research R8). It supersedes nothing; ADR 011
  (ExtProc standalone binary) and ADR 029 (client-assertion trust anchor) both stay in force
- [x] **Library-First Security**: No cryptography is written. Signing is agentgateway's; verification
  is the Broker's existing `jwx`-based validator; the E2E fixture uses `crypto/tls`, `crypto/x509`
  and `jwx`
- [x] **Zalando Guidelines**: N/A — no API change
- [x] **End-User Docs**: `docs/guides/` gains a direct-path guide; `docs/reference/token-exchange.md`
  and `docs/concepts/token-exchange.md` gain the path-selection rule
- [x] **Migration Testing**: N/A — no migration
- [x] **Hexagonal Architecture**: Untouched. No handler, port, adapter or domain service changes
- [x] **Persistence Patterns**: N/A — no persistence change

*Re-checked after Phase 1 design: no gate moved from pass to fail. Two candidate designs that would
have failed the Security-First gate (widening the dev TLS bypass) or the Configuration-Driven gate
(adding a config key for a test's benefit) were rejected in research R3.*

## Project Structure

### Documentation (this feature)

```text
specs/045-agentgateway-token-exchange/
├── plan.md                                   # This file
├── research.md                               # Phase 0 output (R1–R11)
├── data-model.md                             # Phase 1 output
├── quickstart.md                             # Phase 1 output
├── contracts/                                # Phase 1 output
│   ├── agentgateway-direct-route.md          # Gateway configuration contract (standalone + K8s)
│   ├── broker-token-exchange-request.md      # RFC 8693 wire contract + error/status mapping
│   └── broker-direct-path-config.md          # Broker settings + route-selection rule
├── checklists/requirements.md                # Existing
└── tasks.md                                  # Phase 2 output (/speckit-tasks — not created here)
```

### Source Code (repository root)

```text
tests/e2e/gateway/                            # NEW — native-policy E2E suite (no ExtProc import)
├── gateway_suite_test.go                     # TestGatewayNativeTokenExchange
├── native_token_exchange_e2e_test.go         # US1 + US2 scenarios (6 specs)
├── reference_config_e2e_test.go              # US3 scenarios (3 specs)
├── testdata/
│   └── agentgateway-config-v1.5.0.schema.json  # vendored schema, pinned to the image tag
└── support/                                  # suite-local fixtures
    ├── agentgateway_container.go             # pinned image, config render+validate, HostAccessPorts
    ├── https_jwks_server.go                  # TLS JWKS fixture + additional-root transport swap
    ├── signing_key.go                        # gateway keypair, JWKS, assertion helpers
    ├── downstream_backend.go                 # MCP backend recorder + ExtProc stand-in listener
    └── recording_token_endpoint.go           # pass-through recorder for POST /oauth2/token

tests/e2e/bootstrap/
└── test_server.go                            # EXTEND — container-reachable server variant (0.0.0.0)

examples/agentgateway/direct-token-exchange.yaml   # NEW — committed operator reference route
examples/config/token-exchange-direct-gateway.yaml # NEW — matching Broker configuration example
justfile                                      # EXTEND — test-e2e-gateway (+ coverage, aggregate)

mocks/agentgateway/config.yaml                # UNCHANGED (ExtProc route, FR-009)
docker-compose.yml                            # UNCHANGED (research R9)
config.extproc.docker.yaml                    # UNCHANGED
internal/**                                   # UNCHANGED — no production Go change
charts/**, migrations/**, api/**              # UNCHANGED — no config key, no schema, no API change

docs/guides/token-exchange-gateway.md         # EXTEND — name the path it documents (ExtProc)
docs/guides/token-exchange-gateway-direct.md  # NEW — direct-path operator guide
docs/concepts/token-exchange.md               # EXTEND — two alternative paths
docs/reference/token-exchange.md              # EXTEND — gateway-emitted form fields, error mapping
ARCHITECTURE.md                               # EXTEND — glossary + integration-path note
adrs/036-agentgateway-native-token-exchange.md # NEW — two alternative paths; one anchor per instance
```

**Structure Decision**: No `internal/` change and no Compose change. This feature is a committed
reference configuration, operator documentation, and acceptance coverage over an unchanged Broker.

Two placement choices carry weight:

- **The E2E suite is its own package, `tests/e2e/gateway/`.** "No ExtProc dependency" becomes a
  property of the test binary — the package imports nothing from `internal/extproc` — rather than
  something a reader has to confirm by scanning YAML (research R10). It also keeps the root backend
  E2E suite free of a container-runtime dependency, and it isolates the process-global transport swap
  described below.
- **The reference configuration is committed and executed, not illustrated.** US3-S2 boots
  `examples/agentgateway/direct-token-exchange.yaml` itself, with only documented placeholders
  substituted, so the file operators are told to copy is the file under test. An earlier draft added
  a `direct-gateway` Compose profile with a second Broker, a key generator, an HTTPS JWKS server and
  a seeding script; it was cut because Compose profiles do not gate the unconditional
  `extproc-token-exchange` service, because the second Broker would start with empty in-memory
  storage, and because it proved nothing the E2E does not prove with assertions (research R9).

## Implementation Phase Overview

| Phase | Purpose | Required? |
|-------|---------|-----------|
| **Phase 0** | Pre-implementation refactoring | Skipped |
| **Phase 1** | Setup — suite package, justfile recipes, pinned image selector | Included |
| **Phase 2** | Design Preconditions (domain model, config, API, DB, E2E tests) | **MANDATORY** |
| **Phase 2.7** | Entity Boilerplate | Skipped |
| **Phase 2.5** | Foundational Infrastructure — E2E only: container-reachable Broker bootstrap, HTTPS JWKS fixture and scoped CA trust, gateway signing-key helpers, pinned-schema config validation and route-liveness check, token-endpoint recorder, downstream and ExtProc stand-in recorders | Included |
| **Phase 3+** | User Stories — US1 (P1), US2 (P1), US3 (P2) | Included |
| **Phase N** | Constitution Compliance verification | **MANDATORY** |

- [x] Phase 0 (refactoring): **skip** — no existing code is renamed, moved or restructured. The one
  existing-file change, a container-reachable variant in `tests/e2e/bootstrap/test_server.go`, is
  additive and cannot break current suites
- [x] Phase 2.7 (entity boilerplate): **skip** — no new domain entity, port, repository or handler

## Testing Strategy

### End-to-End (E2E) Acceptance Tests

**Test Location**: `tests/e2e/gateway/native_token_exchange_e2e_test.go` and
`tests/e2e/gateway/reference_config_e2e_test.go`

**Framework**: Ginkgo/Gomega, following `tests/e2e/AGENTS.md` and `tests/e2e/README.md`

**Architecture under test**:

```text
MCP client (mcp-go)
  -> agentgateway v1.5.0 container, backendAuth.oauthTokenExchange, no extProc policy
       -> POST /oauth2/token -> Broker built by production app.Builder (host, 0.0.0.0)
            -> HTTPS JWKS fixture serving the gateway public key
       -> forwarded request -> mock MCP backend that records the Authorization header
  (plus an ExtProc stand-in listener that MUST record zero connections)
```

**Test Organization**:

- **Top-level Describe**: "Agentgateway Native Token Exchange"
- **Nested Context**: preconditions — "when the direct policy is configured with the documented trust
  tuple", "when the gateway assertion is not trusted", "when the agent has no delegation", "when the
  Broker is unavailable", "when an operator follows the reference configuration"
- **It blocks**: one per acceptance scenario, nine in total

**Scenario Mapping**:

| Spec Scenario | E2E Test Location | Test Description |
|---------------|-------------------|------------------|
| US1, Scenario 1 | `native_token_exchange_e2e_test.go` | `It("sends two RFC 8693 requests with documented claims and distinct client-assertion jti values", …)` |
| US1, Scenario 2 | `native_token_exchange_e2e_test.go` | `It("replaces the inbound credential with the exchanged token before reaching the backend", …)` |
| US1, Scenario 3 | `native_token_exchange_e2e_test.go` | `It("sends an exchange request with no extProc policy or ExtProc endpoint contact", …)` |
| US2, Scenario 1 | `native_token_exchange_e2e_test.go` | `It("rejects an untrusted assertion or a mismatched assertion or subject-token issuer or audience before resource and delegation checks", …)` |
| US2, Scenario 2 | `native_token_exchange_e2e_test.go` | `It("denies the exchange when the agent has no active delegation for the resource", …)` |
| US2, Scenario 3 | `native_token_exchange_e2e_test.go` | `It("fails closed for missing or unmapped resources and an unavailable Broker", …)` |
| US3, Scenario 1 | `reference_config_e2e_test.go` | `It("ships the ExtProc and direct configurations as alternatives, never combined", …)` |
| US3, Scenario 2 | `reference_config_e2e_test.go` | `It("completes an exchange from the reference configuration without any embedded secret", …)` |
| US3, Scenario 3 | `reference_config_e2e_test.go` | `It("verifies a direct deployment through the documented steps", …)` |

*Line numbers are filled in during Phase 2f when the tests are written.*

**US2, Scenario 1 stays one `It()` with real subcases.** The test uses independently rendered gateway environments for:

- An untrusted signing key.
- A `clientId` that does not match `client_assertion.issuer_uri`.
- An `assertionAudience` that does not match `expected_audience`.
- A subject JWT with an issuer different from the configured upstream fixture issuer or an invalid `aud`.

Each subcase asserts an agent-visible MCP status of 500 and zero backend requests. An unsupported
`clientAuth.alg` causes gateway configuration-load rejection. A real gateway cannot emit that algorithm at runtime. The test must not tamper with an assertion in flight.

**US2, Scenario 3 stays one `It()` with real subcases.** It covers a missing `resource`, an unmapped `resource`, a missing or insufficient stored session, client-assertion JWKS failure, and an unavailable Broker. The Broker returns `invalid_request` with 400, `invalid_target` with 400, and `invalid_grant` with 400. The JWKS failure returns `server_error` with 500. The unavailable Broker has no Broker response. The MCP route returns agent-visible 500 for every error. Every subcase asserts zero backend requests.

**Red Phase Requirements**:

- Tests compile and fail on realistic assertions: exact form field values, decoded assertion claims (`iss`, `sub`, `aud`), HTTP status codes from [contracts/broker-token-exchange-request.md](./contracts/broker-token-exchange-request.md) §5, and a recorded backend-request count of zero on every failure path.
- No `XIt`, `PIt`, `XDescribe`, `PDescribe`, `XContext`, `PContext` or `Skip()`.
- No red-phase comments.

**Test Data Strategy**:

- Reuse `tests/e2e/fixtures`: `ValidAgent`, `GitHubService` (supplies the protected-resource URI), `SeedPlaceholderGrantData` + `ActiveGrant` for delegation, `GitHubSessionForPrincipal` for the stored downstream token, and `OAuth2ConfigWithTokenExchange` as the configuration base. Use its configured upstream fixture issuer to mint valid subject JWTs and a different issuer for the rejection subcase.
- Reuse `tests/e2e/helpers`: `GenerateTestRSAKeyPair`, `SignTestJWT`, `GenerateJWKSFromPublicKey`
- New suite-local fixtures (`tests/e2e/gateway/support/`): HTTPS JWKS server, gateway signing-key
  material written to a file for the container mount, agentgateway container helper, and the
  pass-through token-endpoint recorder
- US2, Scenario 2 removes the delegation rather than the agent, so the failure is `access_denied` and
  not `invalid_grant`

**Test Execution Flow**:

1. Phase 2f: write all nine specs with detailed expectations
2. Verify red: `just test-e2e-gateway` fails on assertions, not on compilation
3. Implement the reference configuration and the documentation
4. Verify green: all nine pass
5. Only fixture adjustments during implementation, never test logic

**Bootstrap Strategy**:

- The Broker is built by production `app.Builder` through `tests/e2e/bootstrap`. A container-reachable
  variant binds `0.0.0.0:0` and advertises `http://host.testcontainers.internal:<port>` so the gateway
  container can dial it; the test itself keeps calling the Broker over the host-mapped address
  (research R4)
- **Per-scenario environment.** Each scenario gets its own Broker, storage, key material, JWKS fixture, gateway container, and downstream recorder. `DeferCleanup` tears down every resource. No scenario shares an `Ordered` environment or depends on another scenario's state. This satisfies the isolation required by `tests/e2e/README.md`. It also budgets container startup per scenario.
- **Before asserting behaviour**, each environment validates its rendered configuration against the
  vendored `v1.5.0` schema and waits for the route to serve, so a drifted field or a route that failed
  to load fails as setup rather than as a misleading routing assertion (research R12)
- The suite runs with `--procs=1` and only in its own package: it installs a clone of
  `http.DefaultTransport` carrying the fixture CA before `app.Builder.Build()` and restores the
  original in `DeferCleanup`, which must never happen inside the parallel root backend suite
  (research R3)

**Helper Utilities**:

- Reuse `tests/e2e/matchers`: `HaveStatusCode`, `HaveOAuth2Error`
- New suite-local helpers only; nothing test-only is added to production packages
- The mock MCP backend follows `startAgentgwMCPServer` in the ExtProc suite: `0.0.0.0:0`, a `whoami`
  tool that returns the observed `Authorization` header, plus a counter of received requests

### Frontend Playwright E2E Tests

Not applicable. This feature changes no React UI, so no Playwright test and no screenshot is required
under Principle XIII.

### Unit & Integration Tests

**Unit Tests**: none required. No production Go code changes. One small exception is permitted if the
suite adds a non-trivial pure helper (for example rendering the gateway configuration from a template);
such a helper gets a table-driven test next to it.

**Integration Tests**: none. No repository, adapter or migration changes.

**Test Coverage Goals**:

- Unit: unchanged
- Integration: unchanged
- E2E: 100% of the nine acceptance scenarios (mandatory, Principle XIII)
- Frontend E2E: not applicable

## Complexity Tracking

> Filled because two design choices are worth justifying on review.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Swapping `http.DefaultTransport` for the duration of the gateway E2E suite | FR-003 requires the Broker to fetch the fixture JWKS over HTTPS, and `app.Builder` gives its JWKS adapter a client with a nil `Transport` (`builder.go:557-559, 677-683`), so this is the only injection point that keeps verification strict and adds no production code (research R3). It is contained by giving the suite its own package, process and `--procs=1`, and restoring the transport in `DeferCleanup` | Serving the JWKS over HTTP with `security.skip_thirdparty_https_validation` fails FR-003 and drops scheme validation. Widening that flag into `InsecureSkipVerify` would disable verification on shared production paths (`builder.go:573-583, 612-618`). Adding a CA-bundle config key would grow the Broker schema, loader, validator, Helm chart and docs contract for a test's benefit |
| A container-reachable variant of the E2E Broker bootstrap | `TestServerBuilderImpl.Build` binds `127.0.0.1:0` (`test_server.go:865-877`), which the testcontainers host-access tunnel cannot reach, so the gateway container could not dial the real Broker at all (research R4) | Running the Broker as a container would abandon the production `app.Builder` bootstrap that Principle XIII requires and would add an image build to every E2E run. Reusing the ExtProc suite's mock Broker is excluded by FR-012 |

**Scope deliberately not taken**: a local Compose stack for the direct path. It was designed, then
cut — it could not prove FR-008 because Compose profiles do not gate the unconditional
`extproc-token-exchange` service, its second Broker would start with empty in-memory storage, and it
would have required a bespoke seeding path plus a key generator and a TLS JWKS server to demonstrate
what the E2E already asserts. Full reasoning and the conditions under which it would become
worthwhile are in research R9.
