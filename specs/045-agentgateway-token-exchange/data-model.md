# Phase 1 Data Model: Agentgateway Native Token Exchange

**Feature**: `045-agentgateway-token-exchange` | **Date**: 2026-09-16
**Input**: [spec.md](./spec.md), [research.md](./research.md)

## Scope statement

This feature adds **no persisted domain entity**, no aggregate, no value object in
`internal/domain/`, no typed entity ID in `internal/domain/id/` (ADR 013 does not apply), and no
database migration. Every entity in the specification is either a **configuration document** owned by
agentgateway, a **credential in flight**, or an **existing Broker domain concept** that the direct
path reuses unchanged.

The model below therefore describes shapes, ownership, validation rules and state transitions, so
that the implementation and the E2E suite have one authoritative reference.

---

## E1. Native Gateway Exchange Policy

The `backendAuth.oauthTokenExchange` block on a direct route. Owned by agentgateway; expressed in the
agentgateway standalone YAML configuration.

**Location in the document**: `binds[].listeners[].routes[].backends[].policies.backendAuth.oauthTokenExchange`.

| Field | Type | Required | Value for this feature |
|---|---|---|---|
| `host` | `string` (`host:port`) | yes | Broker end-user server authority. Port `443` implicitly enables backend TLS |
| `path` | `string` | yes here | `/oauth2/token` |
| `grantType` | enum | no | omitted — defaults to `tokenExchange` (RFC 8693) |
| `resources` | `[]string` | yes here | exactly one absolute protected-resource URI known to the Broker |
| `clientAuth` | object | yes here | see **E3** |
| `subjectToken` | object | no | omitted — defaults to `Authorization: Bearer`, type `…:token-type:access_token` |
| `authorizationLocation` | object | no | omitted — defaults to `Authorization: Bearer` |
| `cache` | object | no | default in the operator reference; `{inMemory: {maxEntries: 0}}` in the E2E fixture (research R6) |
| `audiences`, `scopes` | `[]string` | no | omitted — the Broker ignores `audience` on this path and derives scope from the stored session |
| `actorToken`, `additionalParams`, `requestedTokenType` | — | no | omitted — out of scope (FR-016) |

**Validation rules**

- V1 — the route MUST NOT contain an `extProc` policy anywhere (FR-008).
- V2 — `clientAuth.method` MUST be `privateKeyJwt`; `clientSecretBasic` and `clientSecretPost` are
  forbidden (FR-004, FR-011).
- V3 — `grantType` MUST be absent or `tokenExchange`; `jwtBearer` is out of scope (FR-016).
- V4 — exactly one `resources` entry, an absolute URI without a fragment. Agentgateway warns on other
  values; the Broker rejects an unmapped resource with `invalid_target`.
- V5 — agentgateway rejects a policy that sets more than one backend-auth method.
- V6 — the standalone binary requires camelCase method names; Kubernetes CRDs require PascalCase.

**Relationships**: one policy references exactly one **E2 Protected Resource Mapping** and owns
exactly one **E3 Gateway Client Assertion** identity.

---

## E2. Protected Resource Mapping

The association between the route's `resource` indicator and the Broker-managed service that can
supply a downstream credential. **Existing Broker concept — unchanged by this feature.**

| Aspect | Value |
|---|---|
| Producer | the direct route's `resources[0]` |
| Consumer | Broker protected-resource lookup (ADR 030 normalized child records) |
| Normalization | trailing slash normalized before lookup |
| Failure | no mapping → OAuth2 `invalid_target`, HTTP 400, agent sees 400 |

**Relationships**: one mapping resolves to one `ThirdpartyOAuth2Service`, which is what the
`UserGrant` delegation and permission-set authorization are evaluated against.

---

## E3. Gateway Client Assertion

A fresh JWT minted by agentgateway per token request, proving the gateway's identity to the Broker.
**In flight only — never stored by either side.**

**Producer**: agentgateway `privateKeyJwt` (`client_auth.rs:544-578`).

| Claim / header | Source | Value in the E2E trust tuple |
|---|---|---|
| `iss` | `clientAuth.clientId` | `https://agentgateway-direct-e2e.example.test` |
| `sub` | `clientAuth.clientId` | `https://agentgateway-direct-e2e.example.test` |
| `aud` | `clientAuth.assertionAudience` | `token-exchange-broker` |
| `jti` | fresh UUID per request | — |
| `iat`, `nbf`, `exp` | short lifetime with clock-skew allowance | — |
| `alg` header | `clientAuth.alg` | `RS256` (default) |
| `kid` header | `clientAuth.kid` | set when the JWKS publishes more than one key |

**Transport**: form fields `client_assertion`, `client_assertion_type`
(`urn:ietf:params:oauth:client-assertion-type:jwt-bearer`), plus `client_id`.

**Validation rules (Broker side)**

- V7 — signature verified against `token_exchange.client_assertion.jwks_uri`.
- V8 — `iss` MUST equal `token_exchange.client_assertion.issuer_uri`.
- V9 — `aud` MUST contain `token_exchange.expected_audience`.
- V10 — time claims validated with the configured acceptable skew.
- V11 — any failure → `invalid_client`, HTTP 401, agent sees 500 (research R2). The request never
  reaches resource or delegation authorization, and the protected backend receives nothing.

**State transitions**: `minted (per request) → transmitted → validated | rejected → discarded`.
No caching, no reuse: a new `jti` is generated for every token request.

---

## E4. Inbound Subject JWT

The agent credential that the direct route exchanges.

**Producer**: the agent's identity provider (in E2E, the fixture signing key).
**Read from**: the `Authorization: Bearer` header of the inbound request.
**Transport**: form fields `subject_token`, `subject_token_type`.

| Claim | Constraint | Value in the E2E trust tuple |
|---|---|---|
| `aud` | MUST contain `token_exchange.expected_audience` | `token-exchange-broker` |
| `iss` | MUST match the Broker's configured subject-token issuer | fixture issuer |
| `sub` | feeds `claim_extraction.principal_expression` | the delegating principal |
| `azp` (or configured claim) | feeds `claim_extraction.agent_id_expression` | the agent identity |
| `exp`, `nbf`, `iat` | validated with acceptable skew | — |

**Validation rules (Broker side)**

- V12 — signature, issuer, audience and time validated **before** the client assertion
  (`internal/domain/tokenexchange/service.go:170-190`). Tests MUST NOT assert a client-assertion-first
  ordering.
- V13 — issuer, audience or expiry failure → `invalid_grant`, HTTP 400, agent sees 400.
- V14 — malformed token or bad signature → `invalid_request`, HTTP 400, agent sees 400.

**State transitions**: `presented by agent → read by policy → sent as subject_token → removed from
the request on success`. On success agentgateway removes the inbound credential before writing the
exchanged one (`mod.rs:385-392`), which is what FR-005 requires.

---

## E5. Exchanged Credential

The downstream credential the Broker returns and the gateway forwards.

| Aspect | Value |
|---|---|
| Producer | Broker RFC 8693 success response (`access_token`, `token_type: Bearer`, `issued_token_type`, `expires_in`, `scope`) |
| Placement | `authorizationLocation`, default `Authorization: Bearer` |
| Visibility | the protected backend sees this token and never the inbound one |
| Caching | agentgateway in-memory cache, TTL capped by the subject token `exp`; disabled in the E2E fixture |
| Rejection rules | agentgateway rejects an empty `access_token` or a non-`Bearer` `token_type` and fails the request |

**Invariant I1** — on success the backend request carries exactly one `Authorization` header, whose
value is the exchanged credential. The inbound credential value MUST NOT appear anywhere in the
backend request.

**Invariant I2** — on any failure, the protected backend receives **zero** requests. There is no
fallback to forwarding the inbound credential (FR-007, SC-002).

---

## E6. Trust Tuple (derived configuration invariant)

Not an entity in the runtime sense, but the single constraint that ties E3 and E4 to Broker
configuration. It is one unit: changing one member without the others breaks the path.

| Gateway setting | Broker setting | E2E value |
|---|---|---|
| `clientAuth.clientId` (⇒ assertion `iss`, `sub`) | `token_exchange.client_assertion.issuer_uri` | `https://agentgateway-direct-e2e.example.test` |
| `clientAuth.assertionAudience` (⇒ assertion `aud`) | `token_exchange.expected_audience` | `token-exchange-broker` |
| subject JWT `aud` (minted by the IdP) | `token_exchange.expected_audience` | `token-exchange-broker` |
| public half of `clientAuth.signingKey` | `token_exchange.client_assertion.jwks_uri` | fixture HTTPS JWKS endpoint |
| `resources[0]` | protected-resource mapping of a Broker service | service fixture protected-resource URI |

**Validation rules**

- V15 — FR-018: the reference configuration MUST require operators to replace the issuer, assertion
  audience, subject-token audience, JWKS URI and key locations **together**.
- V16 — R8: a single Broker instance has exactly one client-assertion trust anchor, so an ExtProc
  route and a direct route cannot both authenticate against the same Broker instance unless they
  present assertions from the same issuer.

---

## Glossary additions (ARCHITECTURE.md)

Two terms are new to the ubiquitous language and MUST be added to the `ARCHITECTURE.md` glossary
(Principle V):

| Term | Definition |
|---|---|
| **Native Gateway Exchange Policy** | An agentgateway `backendAuth.oauthTokenExchange` route policy that performs the RFC 8693 exchange against the Broker directly, without ExtProc. |
| **Gateway Client Assertion** | A short-lived JWT signed by the gateway's private key, with `iss` and `sub` equal to the gateway client ID and `aud` equal to the Broker expected audience, used as RFC 7523 client authentication at the Broker token endpoint. |

`Exchanged Credential`, `Protected Resource Mapping`/`ResourceURI` and `Inbound Subject JWT` map onto
existing glossary entries and need no new terms.
