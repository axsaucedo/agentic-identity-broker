# Phase 0 Research: Agentgateway Native Token Exchange

**Feature**: `045-agentgateway-token-exchange` | **Date**: 2026-09-16
**Input**: [spec.md](./spec.md)

All Technical Context unknowns are resolved below. No `NEEDS CLARIFICATION` marker remains.

---

## R1 — Agentgateway `backendAuth.oauthTokenExchange` contract in v1.5.0

**Decision**: Use the standalone-binary `backendAuth.oauthTokenExchange` policy with the default
RFC 8693 grant, attached to the MCP backend at `routes[].backends[].policies.backendAuth`.
Authenticate with `clientAuth.method: privateKeyJwt`.

**Evidence** (agentgateway `v1.5.0`, verified against the released tag, not documentation alone):

| Fact | Source |
|---|---|
| `oauthTokenExchange` exists in the released standalone schema | `schema/config.json` → `$defs.BackendAuth.oneOf[7].properties.oauthTokenExchange` → `$defs.OAuthTokenExchangeAuth` |
| Token endpoint is a backend reference: use a scheme-bearing `host: https://<authority>` for production TLS (or `http://` only in the E2E fixture), plus separate `path` (defaults `/`) | `$defs.OAuthTokenExchangeAuth` `oneOf` + `properties.path`; `SimpleBackendReferenceWithPolicies` |
| Default grant is RFC 8693 (`grantType` defaults to `tokenExchange`) | `$defs.OAuthTokenExchangeAuth.properties.grantType.default` |
| Form fields emitted: `grant_type`, `subject_token`, `subject_token_type`, `audience`, `scope`, `resource`, `requested_token_type`, `client_id`, `client_assertion_type`, `client_assertion` | `crates/agentgateway/src/http/auth/oauth/transport.rs:284-345`, `mod.rs:158-173` |
| `requested_token_type` is **omitted** unless configured | `mod.rs:69-74`, `transport.rs` |
| Client assertion claims are `iss = sub = clientId`, `aud = assertionAudience`, plus `jti`, `nbf`, `iat`, `exp` | `crates/agentgateway/src/http/auth/oauth/client_auth.rs:544-578` |
| `client_assertion_type` is always the RFC 7523 jwt-bearer URN, and `client_id` is sent alongside | `transport.rs:336-342` |
| `privateKeyJwt` required fields: `method`, `clientId`, `signingKey`, `assertionAudience`; optional `alg` (default `RS256`), `kid`, `certificate`, `certificateHeader` | `$defs.RawOAuthClientAuth.oneOf[2]` |
| Signing algorithms: `RS256`, `RS384`, `RS512`, `PS256`, `ES256`, `ES384` | `$defs.JwtSigningAlg` |
| `signingKey` accepts `{file: <path>}` or inline PEM (`FileOrInline`) | `$defs.RawOAuthClientAuth.oneOf[2].properties.signingKey` |
| Subject token defaults to the `Authorization: Bearer` header with type `urn:ietf:params:oauth:token-type:access_token` | `$defs.TokenSpec` defaults |
| On success the inbound credential is **removed** and the exchanged token is written to `authorizationLocation` (default `Authorization: Bearer`) | `mod.rs:385-392`, `$defs.OAuthTokenExchangeAuth.properties.authorizationLocation.default` |
| `resources[]` maps to the RFC 8707 `resource` parameter; non-absolute URIs log a warning | `mod.rs:66-68, 643-651` |
| Response cache defaults to 8192 entries, 300 s TTL when `expires_in` is absent; `maxEntries: 0` disables | `$defs.OAuthTokenExchangeAuth.properties.cache` |
| A policy sets **at most one** backend-auth method; two methods are rejected at load | agentgateway backend-authn docs, "Combine methods" |
| Standalone spells methods camelCase (`privateKeyJwt`); Kubernetes CRDs use PascalCase (`PrivateKeyJwt`); the wrong case fails validation | agentgateway backend-authn docs, "Method availability and field differences" |

**Rationale**: FR-002/FR-003 require exactly this shape. The claim construction in `client_auth.rs`
is what makes the spec's assertion-and-subject trust tuple work: setting `clientId` to
`https://agentgateway-direct-e2e.example.test` produces `iss` and `sub` equal to that value, which is
precisely what the Broker's `token_exchange.client_assertion.issuer_uri` check requires.

**Alternatives considered**:
- `grantType: jwtBearer` (RFC 7523) — rejected: the Broker's third-party exchange path only accepts
  `urn:ietf:params:oauth:grant-type:token-exchange` (`internal/domain/tokenexchange/request.go:120-145`),
  and FR-016 declares JWT-bearer/OBO out of scope.
- `clientSecretBasic` / `clientSecretPost` — rejected by FR-004/FR-011; the Broker validates a signed
  client assertion only and has no shared-secret client authentication for this endpoint.
- Keeping `extProc` alongside the native policy — rejected by FR-008; a direct route carries no
  ExtProc policy at all.

---

## R2 — Gateway failure semantics, and the HTTP status an agent observes

**Decision**: Assert these exact status codes in the E2E failure scenarios.

Agentgateway's MCP route wraps each direct-exchange upstream error as an HTTP 500 response:

| Broker outcome | Broker HTTP status | MCP route handling | Status the agent sees |
|---|---|---|---|
| `invalid_client` (bad/untrusted client assertion) | `401` | Wraps the upstream exchange error | **500** |
| `invalid_grant` (subject-token issuer/audience/expiry, stale permission set, no session) | `400` | Wraps the upstream exchange error | **500** |
| `access_denied` (no/expired user delegation, CEL policy false) | `403` | Wraps the upstream exchange error | **500** |
| `invalid_target` (no protected-resource mapping) | `400` | Wraps the upstream exchange error | **500** |
| `server_error` (CEL failure, JWKS fetch failure) | `500` | Wraps the upstream exchange error | **500** |
| Broker unreachable / connection refused | — | Wraps the upstream exchange error | **500** |

Broker status mapping is `internal/adapters/http/enduser/oauth2_token.go:395-406`
(`invalid_client` → 401, `access_denied` → 403, `server_error` → 500, and the remaining validation errors → 400).

**Rationale**: FR-007 requires fail-closed behaviour; the table turns "fails closed" into concrete,
non-vacuous assertions. In every row the exchange fails before agentgateway replaces the credential,
so the protected backend receives no request at all — which is what SC-002 measures.

**Which US2-S1 variants a real gateway can actually produce**: the spec's edge cases list untrusted
key, wrong issuer, wrong audience, **invalid algorithm** and expired claims. Four of those are
runtime Broker rejections and are driven by rendering a gateway configuration whose `clientId`,
`assertionAudience` or `signingKey` does not match the Broker trust tuple. The invalid-algorithm case
is different: `clientAuth.alg` is a closed enum of asymmetric algorithms
(`RS256`, `RS384`, `RS512`, `PS256`, `ES256`, `ES384`) validated at configuration load, so a real
agentgateway cannot emit `alg: none` or an unlisted algorithm at all. It is therefore covered as a
**gateway configuration-load rejection** — the container fails to serve the route — and not as a
Broker runtime path. Claiming native runtime coverage for it would be false, and tampering with the
assertion in flight would mean no longer testing a real gateway (FR-012).

**Alternatives considered**: asserting only "not 200" — rejected as a weak assertion that would pass
for the wrong reason (for example a routing mistake).

---

## R3 — Serving the gateway's public JWKS to the Broker over HTTPS in E2E

**Decision**: The E2E fixture serves the gateway public JWKS from `httptest.NewTLSServer`. The suite
installs the fixture's certificate as an **additional** root in a clone of `http.DefaultTransport`
for the duration of the suite, and restores the original transport afterwards. Certificate
verification stays fully enabled; no production code, no configuration key, and no verification
bypass is involved.

**Why this works**: `app.Builder` builds its upstream HTTP client as `&http.Client{Timeout: …}` with a
**nil** `Transport` (`internal/app/builder.go:557-559`), so the client resolves `http.DefaultTransport`
at request time. The client-assertion JWKS adapter receives that same client
(`internal/app/builder.go:677-683`), so the fixture CA is honoured by the real, production JWKS
fetch path. The gateway suite is its own Go package and therefore its own test binary and process,
so the swap is contained; it runs with `--procs=1` to keep it single-process.

**Rationale**: FR-003 and FR-013 require the Broker's `jwks_uri` to be the fixture's **HTTPS**
endpoint. The Broker's own config validator already demands HTTPS for
`token_exchange.client_assertion.jwks_uri` unless a development bypass is enabled
(`internal/config/validator.go:655-688`), so an HTTPS fixture also keeps the E2E on the
production-default validation path.

**Alternatives considered and rejected**:
- **HTTP JWKS + `security.skip_thirdparty_https_validation: true`** (the pattern used by
  `tests/e2e/hybrid_token_exchange_e2e_test.go:326-333`) — rejected: it does not satisfy FR-003's
  HTTPS requirement and it removes the scheme validation the direct path depends on.
- **Making `security.skip_thirdparty_https_validation` set `InsecureSkipVerify` on the shared
  upstream client** — rejected: that client also serves OAuth2 session token calls
  (`builder.go:573-583`) and the shared upstream JWKS adapter (`builder.go:612-618`), so the change
  would silently disable certificate verification across production paths. Principle I forbids it.
- **A new `token_exchange.client_assertion.ca_bundle_path` config key** — rejected: it expands a
  documentation-and-tests feature into a broker schema, loader, validator, Helm and docs contract
  change for something the test can achieve without any production change.
- **`SSL_CERT_FILE`** — rejected: Go ignores it on darwin (`root_unix.go` is built for
  `unix && !darwin`), so the suite would behave differently on developer machines and in CI.

---

## R4 — Making the Broker reachable from the agentgateway container

**Decision**: Add a container-reachable variant of the E2E broker bootstrap that binds `0.0.0.0:0`
and advertises `http://host.testcontainers.internal:<port>` to the gateway, while tests continue to
call the Broker directly over the host-mapped address. Register the port through
`testcontainers.ContainerRequest.HostAccessPorts`.

**Rationale**: `tests/e2e/bootstrap/test_server.go:865-877` binds `127.0.0.1:0`, which the
testcontainers host-access tunnel cannot reach. The existing agentgateway suite solves the same
problem for its host fixtures by binding `0.0.0.0:0`
(`tests/e2e/extproc/agentgateway_e2e_test.go:300-302, 357-358`) and wiring
`host.testcontainers.internal` into the generated gateway config
(`tests/e2e/extproc/agentgateway_e2e_test.go:378-410`). This is a `tests/e2e/bootstrap` change only —
no production API is added for tests (repository rule: no test-only production APIs).

**Note on `PublicURL`**: `TestServerBuilderImpl.Build` overwrites
`config.Server.EndUser.PublicURL` with the listener address. The container-reachable variant must set
it to the `host.testcontainers.internal` URL so that Broker-minted URLs stay consistent with the
address the gateway actually dials.

**Alternatives considered**:
- Running the Broker as a container in the same Docker network — rejected: Principle XIII requires
  E2E tests to boot the Broker through production bootstrap code (`app.Builder`) in
  `tests/e2e/bootstrap/`, and a container would also require building and pinning a Broker image for
  every E2E run.
- Reusing the ExtProc suite's mock Broker — rejected explicitly by FR-012.

---

## R5 — Observing the exact RFC 8693 form the gateway sends

**Decision**: Wrap the production end-user router in a recording `http.Handler` that copies the
`POST /oauth2/token` form values and then delegates, unchanged, to the production handler chain.
Assertions read the recorded form.

**Rationale**: Acceptance scenario US1.1 asserts on the *content* of the exchange request
(`subject_token` audience, `resource`, `client_assertion` `iss`/`sub`/`aud`). With a real Broker there
is no mock to inspect, and decoding the assertion from a log line would be brittle. A pass-through
recorder observes the request without replacing or altering the Broker boundary, so FR-012 ("not a
replacement, not a mocked Broker exchange interface") still holds.

**Alternatives considered**:
- Asserting only the downstream token — rejected: it cannot prove the `iss`/`sub`/`aud` tuple that
  FR-003, FR-013 and SC-003 demand.
- Parsing gateway debug logs — rejected: agentgateway redacts secrets and the log format is not a
  contract.

---

## R6 — Deterministic exchanges: the gateway token cache

**Decision**: Set `cache: { maxEntries: 0 }` on the direct route in the E2E gateway
configuration. Keep the default cache in the operator reference configuration and document it.

**Rationale**: The default cache keys on subject token plus grant parameters with a 300 s TTL
(`$defs.OAuthTokenExchangeAuth.properties.cache`). Several scenarios reuse the same subject token
across requests; a cached token would make a later scenario pass without contacting the Broker,
which would silently void the assertions in FR-014 and FR-017. Production deployments want the cache,
so only the test configuration disables it.

**Alternatives considered**: minting a unique subject token per scenario — kept as an additional
safeguard for the failure scenarios, but not sufficient on its own because the success scenarios must
prove the Broker is contacted each time.

---

## R7 — The Broker side of the contract (no API or schema change)

**Decision**: The direct path reuses the existing `POST /oauth2/token` third-party token-exchange
endpoint as-is. No OpenAPI change, no domain change, no storage change, no migration.

**Verified Broker behaviour**:

| Aspect | Behaviour | Source |
|---|---|---|
| Route | `POST /oauth2/token` | `internal/adapters/http/routing/enduser.go:141-159` |
| Accepted content type | `application/x-www-form-urlencoded` only | `internal/adapters/http/enduser/oauth2_token.go:56-59` |
| Fields read on the third-party path | `grant_type`, `subject_token`, `subject_token_type`, `client_assertion`, `client_assertion_type`, `resource`, `scope` | `internal/adapters/http/enduser/oauth2_token.go:134-171` |
| Required fields | non-empty `subject_token`, `client_assertion`, `resource`; `resource` must be an absolute URI | `internal/domain/tokenexchange/request.go:120-145` |
| `subject_token_type` / `client_assertion_type` | defaulted when empty, explicit values not re-validated on this path | `internal/domain/tokenexchange/request.go:97-103` |
| Ignored on this path | `audience`, `actor_token`, `requested_token_type` | `internal/domain/tokenexchange/request.go`, handler dispatch |
| Validation order | subject JWT → client assertion → CEL principal/agent extraction and privileged-client authorization → protected-resource lookup → agent access → permission-set authorization → session token | `internal/domain/tokenexchange/service.go:170-345` |
| Audience check | both credentials validated against `token_exchange.expected_audience` (default `token-exchange-broker`) | `internal/domain/tokenexchange/jwt_validator.go:128-176`, `internal/config/loader.go:256-264` |
| Client-assertion trust anchor | dedicated external issuer + JWKS via `internal/adapters/jwks.Adapter` (httprc), fail-closed when stale | `internal/app/builder.go:635-696`, `internal/adapters/jwks/adapter.go` |
| Resource lookup | normalized protected-resource child records; unmapped resource → `invalid_target` | ADR 030; `internal/adapters/storage/{memory,postgres}/thirdparty_provider.go` |
| Delegation | missing or expired `UserGrant` → `access_denied` | `internal/domain/consent/service.go:639-681` |

**Consequence for FR-013**: the Broker validates the subject token *before* the client assertion.
That still satisfies the spec, which only requires both credentials to be accepted **before** resource
and delegation authorization. The E2E must therefore not assert a client-assertion-first ordering.

**Alternatives considered**: adding a dedicated gateway endpoint — rejected: the spec's Assumptions
state the existing Broker contract is the source of truth and is not altered.

---

## R8 — One Broker instance cannot anchor both integration paths

**Finding**: `token_exchange.client_assertion` is a single, broker-wide trust anchor
(`internal/ports/config.go:769-789`, `internal/app/builder.go:635-696`). The existing ExtProc
deployment presents an upstream-issued access token as its `client_assertion`
(`config.extproc.docker.yaml`, `ClientAssertionType: access_token`), which validates against the
upstream issuer fallback (`internal/app/builder.go:449-451`). The direct path presents a
gateway-self-signed assertion whose issuer is the gateway `clientId`. These are different issuers, so
one Broker instance cannot anchor both at the same time.

**Decision**: Treat this as an operator-facing route-selection rule required by FR-010 and FR-018,
and state it in the guide and in
[contracts/broker-direct-path-config.md](./contracts/broker-direct-path-config.md) §3: plan one
Broker instance per client-assertion issuer. The local Compose stack is **not** changed (see R9).

**Rationale**: This is a real constraint that an operator will hit on day one. Documenting it is
cheaper and safer than discovering it as an `invalid_client` in production.

**Alternatives considered**:
- Multiple client-assertion anchors in the Broker — rejected: that is a Broker contract change, which
  the spec's Assumptions exclude, and it would need its own ADR superseding ADR 029.
- Silently changing `config.docker.yaml` to the gateway anchor — rejected: it breaks the existing
  ExtProc route and violates FR-009.

---

## R9 — No second local Compose stack; ship a reference configuration the E2E executes

**Decision**: Leave `docker-compose.yml` and every ExtProc asset untouched. Ship the direct path as
a committed operator reference configuration, `examples/agentgateway/direct-token-exchange.yaml`,
plus the operator guide. The runnable proof of that exact file is User Story 3, Scenario 2, which
boots a real pinned agentgateway container **from the committed reference file** with only the
documented placeholders substituted, and drives a successful exchange against the production Broker.

**Rationale**: an earlier draft added a `direct-gateway` Compose profile with a second Broker, a key
generator, a TLS JWKS server and a seeding script. That design failed on four counts:

1. **It could not prove FR-008.** Compose profiles only gate services that *declare* a profile.
   `extproc-token-exchange` declares none (`docker-compose.yml:199-221`), so it starts regardless,
   and `agentgateway` depends on it. A stack that still runs ExtProc cannot demonstrate "no ExtProc
   service dependency". A standalone Compose file could avoid that, but it would duplicate or
   `extends`-inherit the mock services, and `extends` merges fields such as `depends_on` that would
   then have to be audited by hand for every future edit.
2. **The second Broker would start empty.** Both Broker services run
   `IDENTITY_BROKER_STORAGE_BACKEND=memory` (`docker-compose.yml:53`) and share nothing. A direct
   instance has no agent, service, protected-resource mapping, delegation or stored session, so the
   documented flow would need a bespoke seeding path through the admin and consent APIs before it
   could return anything but `invalid_target` or `access_denied`.
3. **It could not honour the HTTPS contract end to end.** The reused mock upstream is
   `http://upstream-oauth2:9001`, and the gateway-to-Broker hop would be plain HTTP, so any claim
   that the direct path is TLS-protected in the local stack would be false for most hops.
4. **It proves nothing the E2E does not prove better.** The E2E already exercises the real Broker,
   a real pinned gateway, a genuinely HTTPS JWKS endpoint with strict verification, and the absence
   of ExtProc — with assertions instead of prose.

**How SC-004 is still met**: SC-004 asks that an operator can follow one documented setup from the
ExtProc baseline through downstream-token verification. The guide provides that setup, and the
committed reference configuration it tells the operator to copy is the same file the E2E boots and
verifies (US3-S2) and compares against the ExtProc configuration (US3-S1). The documented path is
therefore continuously tested rather than merely demonstrated once by hand.

**Consequences**: no `docker-compose.yml` change, no second Broker, no key-generation container, no
seeding script, no `SSL_CERT_FILE` wiring, no new Compose volumes. FR-009 holds trivially, because
nothing in the ExtProc deployment is touched.

**Alternatives considered**:
- A standalone `docker-compose.direct.yml` with a fully specified service graph — rejected for this
  feature: it is a second deployment topology to maintain and seed for a path the E2E already proves.
  It remains a reasonable follow-up if operators ask for a local sandbox.
- Committing a development keypair so the stack needs no generator — rejected: committed private keys
  are a standing liability and set a bad example in an operator-facing reference configuration.
- Inline PEM in the reference configuration — rejected by FR-011; the reference uses
  `signingKey: {file: …}`.

---

## R10 — Test suite placement and execution

**Decision**: New Ginkgo suite package `tests/e2e/gateway/` (suite `TestGatewayNativeTokenExchange`),
with `just test-e2e-gateway` and inclusion in the aggregate `just test-e2e`.

**Rationale**: FR-008 requires the direct proof to have no ExtProc dependency. Placing the suite in
`tests/e2e/extproc/` would link it against `internal/extproc/...` and make "no ExtProc" a matter of
reading the config rather than a property of the binary. A separate package makes the absence
structural. It also keeps the root `tests/e2e/` suite free of a container-runtime dependency.

**Container runtime**: reuse the existing Docker/Podman availability check pattern from the ExtProc
agentgateway tests and the pinned image selector (`defaultAgentgatewayImage`,
`AGENTGATEWAY_IMAGE` override) so the image stays pinned to `v1.5.0` per FR-012.

**Alternatives considered**: adding the tests to the root backend E2E suite — rejected: it would make
every backend E2E run depend on a container runtime and a network image pull.

---

## R11 — Making the three User Story 3 scenarios executable

**Decision**: US3 scenarios assert on the shipped artifacts and on observable direct-path behaviour,
not on prose.

| Scenario | Executable assertion |
|---|---|
| US3.1 — the two paths are alternatives | Parse both committed gateway configurations. `examples/agentgateway/direct-token-exchange.yaml` contains `backendAuth.oauthTokenExchange` and no `extProc` key anywhere; `mocks/agentgateway/config.yaml` contains `extProc` and no `oauthTokenExchange` key. Also assert the committed operator guide names both paths as alternatives, names `privateKeyJwt` as the compatible method, and states that shared-secret methods are excluded |
| US3.2 — documented setup works without embedded secrets | Boot agentgateway from `examples/agentgateway/direct-token-exchange.yaml` with only the documented placeholders substituted; assert the exchange succeeds, and that the rendered configuration contains no `clientSecret`, no `BEGIN … PRIVATE KEY`, and sources the key through `signingKey.file` |
| US3.3 — documented verification steps | Run the guide's verification steps: the downstream service sees a token different from the inbound one, and a listener standing in for an ExtProc endpoint records zero connections for the whole scenario |

**Rationale**: Principle XIII forbids placeholder assertions. These are concrete, fail-for-the-right-
reason checks that also protect the documentation from drifting away from the code. Because US3.2
boots the committed reference file rather than a test-local copy, the file an operator is told to
copy is the file that is proven to work.

**Alternatives considered**: a manual documentation checklist — rejected: Principle VIII forbids
manual validation as the acceptance mechanism, and SC-003 requires nine independently executable
tests.

---

## R12 — Guarding against agentgateway schema drift

**Decision**: Two guards, both cheap:

1. **Schema-pinned validation.** Vendor the `v1.5.0` `schema/config.json` under
   `tests/e2e/gateway/testdata/agentgateway-config-v1.5.0.schema.json` and validate every committed
   and rendered gateway configuration against it in the suite. This is the same artifact the
   `# yaml-language-server: $schema=…` comment points editors at, pinned to the image tag the tests
   run.
2. **Startup and route liveness check.** After the container reports its listening port, assert the
   route actually serves before asserting exchange behaviour, so a configuration the gateway silently
   downgraded or a route that failed to load surfaces as a setup failure rather than as a misleading
   routing assertion.

**Rationale**: the field names differ between sources. The launch blog shows `tokenEndpointPath`,
while the released standalone schema uses `path`; the standalone binary requires camelCase method
names while the CRDs require PascalCase; and `latest` documentation tracks `main`, not `v1.5.0`.
Without a pinned schema check, a copy-paste from newer documentation could silently select a
different grant or drop a field, and the suite would fail somewhere far from the cause — or worse,
pass while testing something else.

**Alternatives considered**: relying on the gateway rejecting bad configuration at load — partly
effective, but agentgateway logs rather than fails for some mismatches, for example a `certificate`
whose leaf public key does not match `signingKey`.
