---
title: "Token exchange"
description: "The field-level reference for the RFC 8693 token-exchange grant on POST /oauth2/token: request and response parameters, the resource and user-grant model, error cases, and worked examples."
---

# Token exchange

This is the field-level reference for the RFC 8693 token-exchange grant on the broker's token
endpoint. It documents every request and response field, the resource-matching and
user-grant model, and each error case.

:::note
For the concept and the request flow, read [Token exchange](/docs/concepts/token-exchange).
To deploy transparent exchange at an Envoy-based gateway, see the
[token-exchange gateway guide](/docs/guides/token-exchange-gateway). The complete generated
contract for this endpoint lives in the [end-user API reference](/api/enduser).
:::

Token exchange lets a privileged client — an API gateway or reverse proxy — exchange an
agent's token for a third-party OAuth2 token the broker holds encrypted, so the agent never
holds the third-party credential itself. The broker returns the third-party token only after
verifying that the user has an active grant for that agent and service.

## Overview

- **Endpoint:** `POST /oauth2/token` (end-user server, port 8000).
- **Grant type:** `urn:ietf:params:oauth:grant-type:token-exchange`.
- **Content type:** `application/x-www-form-urlencoded`.

The token endpoint is unified: it detects a token-exchange request by its `grant_type` and
processes it per RFC 8693. The same endpoint also serves the `authorization_code` and
`client_credentials` grants, documented in the [API overview](/docs/reference/api) and
[OAuth2 server modes](/docs/concepts/oauth2-server-modes).

## Key concepts

### Tokens

**Subject token** (`subject_token`) — a JWT carrying the user's principal (extracted via a
configurable CEL expression, default the `sub` claim) and the agent identifier (default the
`azp` claim). The privileged client presents it to request the exchange.

**Client assertion** (`client_assertion`) — a JWT identifying the privileged client (gateway
or reverse proxy), validated against the upstream OAuth2 server's JWKS. Its `sub` claim
identifies the gateway for audit logging.

**Third-party token** — the OAuth2 access token the broker holds encrypted in its token vault
for the target service (for example GitHub or Google). This is what a successful exchange
returns.

### Resource parameter

The `resource` parameter identifies which service's token to return. The broker matches it
against the `protected_resources` URIs configured on services. The URI is normalized
(trailing slash removed) before matching, and each resource URI may be configured on only one
service.

```mermaid
flowchart LR
    A["resource=https://api.github.com"] --> B["Broker matches protected_resources"]
    B --> C["Service found: GitHub"]
    C --> D["Return that service's stored token"]
```

Configure `protected_resources` on a service through the [admin API](/api/admin); see
[Manage agents and services](/docs/guides/manage-agents-and-services). A resource URI must be
a valid HTTP or HTTPS URI (scheme required), is stored trailing-slash-normalized, and a
duplicate URI across services is rejected with `409 conflict`.

### User grant

Before returning a token, the broker verifies that the user (from the subject token's
principal claim) has an active [grant](/docs/concepts/delegation-and-consent) allowing the
agent (from the subject token's agent claim) to use the target service. Without a current,
non-expired grant, the exchange is denied. This is where the consent model is enforced at
request time.

## Request format

`POST /oauth2/token` with `Content-Type: application/x-www-form-urlencoded`.

### Required parameters

| Parameter | Value | Description |
|---|---|---|
| `grant_type` | `urn:ietf:params:oauth:grant-type:token-exchange` | Selects RFC 8693 token exchange. |
| `subject_token` | JWT string | The user's token (contains user principal + agent identifier). |
| `subject_token_type` | `urn:ietf:params:oauth:token-type:access_token` | The only accepted subject-token type. |
| `client_assertion` | JWT string | The gateway's authentication token, validated against upstream JWKS. |
| `client_assertion_type` | `urn:ietf:params:oauth:client-assertion-type:jwt-bearer` | The client authentication method. |
| `resource` | URI string | Target service resource URI (for example `https://api.github.com`). |

### Optional parameters

| Parameter | Value | Description |
|---|---|---|
| `scope` | Space-delimited scopes | Requested scopes. |
| `audience` | String | Intended audience for the issued token. |

### Request example

```bash
curl -X POST http://localhost:8000/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_assertion=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d "client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer" \
  -d "resource=https://api.github.com"
```

## Response format

A successful exchange returns `200 OK` with `Content-Type: application/json`.

```json
{
  "access_token": "ghu_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "token_type": "Bearer",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "expires_in": 3600,
  "scope": "read:user repo"
}
```

### Response fields

| Field | Presence | Description |
|---|---|---|
| `access_token` | Always | The third-party OAuth2 access token from the token vault. |
| `token_type` | Always | Token type, passed through from the stored token (typically `Bearer`). |
| `issued_token_type` | Always | `urn:ietf:params:oauth:token-type:access_token`. |
| `expires_in` | Optional | Remaining lifetime in seconds, when the stored token carries expiry. |
| `scope` | Optional | Scopes associated with the token, when they differ from the request. |
| `refresh_token` | Optional | A refresh token — typically absent, since the broker manages refresh. |
| `granted_permission_sets` | Optional | Map of permission-set UUID → array of service UUIDs, when the exchange is scoped to specific permission sets. |

### Success example

```bash
curl -X POST http://localhost:8000/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=$UPSTREAM_TOKEN" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_assertion=$GATEWAY_JWT" \
  -d "client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer" \
  -d "resource=https://api.github.com"
```

```json
{
  "access_token": "ghu_1234567890abcdefghijklmnopqrstuvwxyz",
  "token_type": "Bearer",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "expires_in": 28800
}
```

## Error responses

Token-exchange errors use the OAuth2 error format: an `error` code and an optional
`error_description`.

| HTTP status | `error` | When it occurs |
|---|---|---|
| 400 | `invalid_request` | The request is malformed or missing a required parameter. |
| 400 | `invalid_target` | The `resource` matches no configured service. |
| 400 | `invalid_grant` | The user has no active session with the target service. |
| 401 | `invalid_client` | The `client_assertion` (or broker credentials) cannot be verified. |
| 403 | `access_denied` | The user has not granted the agent access, or a CEL policy denied it. |
| 500 | `server_error` | An unexpected server error. |

### 400 invalid_request

```json
{
  "error": "invalid_request",
  "error_description": "subject_token is required"
}
```

Common causes: a missing required parameter (`subject_token`, `client_assertion`, `resource`);
an invalid subject-token or client-assertion format or signature; an invalid resource URI; or
a subject token missing a required claim such as `sub`.

### 400 invalid_target

```json
{
  "error": "invalid_target",
  "error_description": "No service configured for the requested resource"
}
```

No service has the requested URI in its `protected_resources`.

### 400 invalid_grant

```json
{
  "error": "invalid_grant",
  "error_description": "User has no active session with the requested service"
}
```

The user has no session for the service, or the stored access and refresh tokens have both
expired and cannot be refreshed.

### 401 invalid_client

```json
{
  "error": "invalid_client",
  "error_description": "Invalid client_assertion signature"
}
```

The `client_assertion` signature failed verification against the upstream JWKS, the assertion
has expired, or its audience does not include the broker.

### 403 access_denied

```json
{
  "error": "access_denied",
  "error_description": "User has not granted this agent access to the requested service"
}
```

The user has no grant for this agent and service, the grant was revoked or has expired, or a
CEL authorization expression evaluated to false.

### 500 server_error

```json
{
  "error": "server_error",
  "error_description": "An unexpected error occurred"
}
```

## Configuration

Token exchange is configured under `token_exchange` in the broker configuration. Two settings
matter most; the [Configuration reference](/docs/configuration) covers the rest.

- **Claim extraction** — CEL expressions that pull the principal and agent identifier out of
  the subject token.
- **Authorization** — a CEL expression, evaluated against the `client_assertion` claims, that
  decides whether a privileged client may perform the exchange.

```yaml
token_exchange:
  claim_extraction:
    principal_expression: subject_token.sub
    agent_id_expression: subject_token.azp
  authorization:
    cel:
      expression: "client_assertion.iss == 'api-gateway.example.com'"
```

The authorization expression can allow any client (`"true"`), pin a single issuer as above, or
accept several:

```yaml
token_exchange:
  authorization:
    cel:
      expression: |
        client_assertion.iss in [
          'api-gateway.example.com',
          'reverse-proxy.example.com'
        ]
```

## Worked example

This walks a GitHub delegation end to end. Administrative calls go to the admin server
(port 14000); consent and token exchange go to the end-user server (port 8000).

```bash
# 1. Admin registers the GitHub service with a protected resource (admin API, :14000)
curl -X POST http://localhost:14000/api/services \
  -H "X-Remote-User: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "GitHub",
    "client_id": "github-oauth-client",
    "issuer_uri": "https://github.com",
    "protected_resources": ["https://api.github.com"]
  }'
# → {"id": "aa0e8400-e29b-41d4-a716-446655440000", ...}

# 2. Admin creates a permission set bundling the GitHub scopes (admin API, :14000)
curl -X POST http://localhost:14000/api/permission-sets \
  -H "X-Remote-User: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Read repositories",
    "description": "Read access to GitHub repositories",
    "service_scopes": [
      {
        "service_id": "aa0e8400-e29b-41d4-a716-446655440000",
        "scopes": ["repo"],
        "requirement_type": "mandatory"
      }
    ]
  }'
# → {"id": "770e8400-e29b-41d4-a716-446655440001", ...}

# 3. Admin registers the agent (admin API, :14000)
curl -X POST http://localhost:14000/api/agents \
  -H "X-Remote-User: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{"display_name": "GitHub Assistant"}'
# → {"id": "550e8400-e29b-41d4-a716-446655440000", ...}

# 4. User grants the permission set for that service (end-user API, :8000)
curl -X POST http://localhost:8000/api/consent/agents/550e8400-e29b-41d4-a716-446655440000/grants \
  -H "X-Remote-User: user@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "granted_permission_sets": {
      "770e8400-e29b-41d4-a716-446655440001": [
        "aa0e8400-e29b-41d4-a716-446655440000"
      ]
    }
  }'
# → 201 {"data": {"agent_id": "550e8400-...", "granted_permission_sets": {...}}}

# 5. User completes the GitHub authorization flow in the browser,
#    which stores an encrypted third-party session.

# 6. The gateway exchanges the agent's token for the GitHub token (end-user API, :8000)
curl -X POST http://localhost:8000/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=$UPSTREAM_JWT_WITH_USER_SUB_AND_AGENT_AZP" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_assertion=$GATEWAY_JWT" \
  -d "client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer" \
  -d "resource=https://api.github.com"
# → 200 {"access_token": "ghu_...", "token_type": "Bearer", ...}

# 7. The gateway calls GitHub with the exchanged token
curl -H "Authorization: Bearer ghu_..." https://api.github.com/user
```

## Security considerations

- **Both JWTs are verified.** The `client_assertion` and `subject_token` signatures are
  validated against the upstream OAuth2 server's JWKS; expired or wrongly-audienced tokens are
  rejected.
- **A user grant is always required.** The broker returns a token only when the principal has
  an active, non-expired grant for the agent and service.
- **Token values are never logged.** The returned `access_token` does not appear in logs;
  audit records reference the principal, agent, service, and resource, not the token.
- **Transport is HTTPS.** Terminate TLS in front of the broker; tokens must not cross the
  network in the clear.

## Related

- [Token exchange (concept)](/docs/concepts/token-exchange) — the model and request flow.
- [Token-exchange gateway guide](/docs/guides/token-exchange-gateway) — transparent exchange
  at an Envoy-based gateway.
- [End-user API reference](/api/enduser) — the generated `/oauth2/token` contract.
- [RFC 8693 — OAuth 2.0 Token Exchange](https://www.rfc-editor.org/rfc/rfc8693).
