# Contract: Broker RFC 8693 Token Exchange, as driven by the direct gateway route

**Feature**: `045-agentgateway-token-exchange`
**Endpoint**: `POST /oauth2/token` (end-user server) — **existing, unchanged**
**OpenAPI**: `api/enduser/openapi.yaml` — **no change**; this document records how the direct path
uses the endpoint that is already specified there.

> No API addition, modification or removal is proposed by this feature. Principles IV and X are
> satisfied by the existing, already-confirmed contract.

---

## 1. Request the gateway emits

`Content-Type: application/x-www-form-urlencoded`

| Form field | Emitted by agentgateway | Value |
|---|---|---|
| `grant_type` | always | `urn:ietf:params:oauth:grant-type:token-exchange` |
| `subject_token` | always | the inbound agent credential |
| `subject_token_type` | always | `urn:ietf:params:oauth:token-type:access_token` (default) |
| `resource` | when `resources` set | the configured protected-resource URI |
| `client_id` | with `privateKeyJwt` | the gateway `clientId` |
| `client_assertion_type` | with `privateKeyJwt` | `urn:ietf:params:oauth:client-assertion-type:jwt-bearer` |
| `client_assertion` | with `privateKeyJwt` | freshly signed JWT (see §2) |
| `audience`, `scope` | only if configured | omitted by the reference configuration |
| `requested_token_type` | only if configured | omitted (RFC 8693 makes it optional) |
| `actor_token`, `actor_token_type` | only if configured | omitted |

Broker-side reading of this form: `internal/adapters/http/enduser/oauth2_token.go:134-171`.
Broker-side required fields: non-empty `subject_token`, `client_assertion`, and an absolute-URI
`resource` (`internal/domain/tokenexchange/request.go:120-145`). `audience`, `actor_token` and
`requested_token_type` are ignored on this path.

---

## 2. Client assertion the gateway signs

```json
{
  "iss": "<clientAuth.clientId>",
  "sub": "<clientAuth.clientId>",
  "aud": "<clientAuth.assertionAudience>",
  "jti": "<fresh uuid per request>",
  "iat": 0, "nbf": 0, "exp": 0
}
```

Header: `alg` from `clientAuth.alg` (default `RS256`), optional `kid`, optional `x5c`/`x5t#S256`
when a certificate is configured.

The Broker requires `iss == token_exchange.client_assertion.issuer_uri` and
`aud ∋ token_exchange.expected_audience`, verifying the signature against
`token_exchange.client_assertion.jwks_uri`.

---

## 3. Broker validation order

`internal/domain/tokenexchange/service.go:170-345`

1. Request shape (`grant_type`, non-empty `subject_token` / `client_assertion`, absolute `resource`).
2. **Subject token**: signature, issuer, `aud ∋ expected_audience`, time claims.
3. **Client assertion**: signature via the client-assertion JWKS, `iss`, `aud`, time claims.
4. CEL claim extraction (principal, agent ID) and privileged-client authorization.
5. Protected-resource normalization and lookup (ADR 030).
6. Agent access verification.
7. Permission-set authorization.
8. Stored session token retrieval and scope coverage.

Steps 2 and 3 both complete before steps 5–7. Tests MUST NOT assert a client-assertion-first
ordering.

---

## 4. Success response

```json
{
  "access_token": "<downstream credential>",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "<scope of the stored session>"
}
```

Agentgateway rejects a response with an empty `access_token` or a `token_type` other than `Bearer`,
and fails the request.

---

## 5. Error contract and what the agent observes

| Failure | Broker `error` | Broker status | Agent-visible status | Backend requests |
|---|---|---|---|---|
| Untrusted key, wrong assertion `iss`, wrong assertion `aud`, expired assertion | `invalid_client` | 401 | **500** | 0 |
| Subject token wrong `aud`, wrong issuer, expired | `invalid_grant` | 400 | **500** | 0 |
| Subject token malformed or bad signature | `invalid_request` | 400 | **500** | 0 |
| `resource` missing | `invalid_request` | 400 | **500** | 0 |
| `resource` not mapped to a service | `invalid_target` | 400 | **500** | 0 |
| No user delegation, expired delegation, CEL policy denies | `access_denied` | 403 | **500** | 0 |
| No stored session, session expired, scope not covered, stale permission set | `invalid_grant` | 400 | **500** | 0 |
| CEL evaluation failure, JWKS unavailable | `server_error` | 500 | **500** | 0 |
| Broker unreachable | — | — | **500** | 0 |

Broker status mapping: `internal/adapters/http/enduser/oauth2_token.go:395-406` (RFC 6749 §5.2).
Gateway MCP handling: the MCP route wraps every Broker-originated direct-exchange error as HTTP 500.
The Broker status remains visible only at the Broker boundary.

**Invariant**: in every row the protected backend receives zero requests, and the inbound credential
is never forwarded (FR-007, SC-002).

---

## 6. Downstream request contract

On success the request forwarded to the protected backend carries exactly one `Authorization` header
whose value is `Bearer <exchanged credential>`. The inbound credential value MUST NOT appear in any
header, query parameter or body of the forwarded request (FR-005, SC-001).
