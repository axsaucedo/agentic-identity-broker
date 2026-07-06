---
title: "API overview"
description: "The shared conventions of the Agentic Identity Broker's dual-port API: the reverse-proxy auth model, response envelopes, error codes, and a compact map of every end-user and admin endpoint."
---

# API overview

This page is a hand-authored companion to the two generated OpenAPI contracts. It explains
the conventions the endpoints share — how requests authenticate, how responses are shaped,
and what errors mean — and gives you a compact map of the whole surface.

The full, field-level contracts are generated from the source specifications and rendered at:

- **[End-user API reference](/api/enduser)** — port 8000.
- **[Admin API reference](/api/admin)** — port 14000.

Those two pages are the authoritative contract for request and response schemas. This
overview does not repeat them; it points you to the right group and explains what is common
across all of them.

## Authentication model

The broker relies on a **trusted reverse proxy** for pre-authentication. The proxy
(oauth2-proxy, nginx `auth_request`, a service mesh) authenticates the caller and injects the
principal in a request header. The broker never authenticates end users itself.

- **Principal header** — default **`X-Remote-User`**, carrying the principal (an email,
  username, or opaque ID). The header name is configurable. The broker trusts it only from a
  trusted source; requests carrying it from an untrusted source are rejected.
- **Session cookie** (end-user server only) — after pre-authentication, the broker maintains
  a `session_token` cookie for the session duration. On the end-user server, a request is
  authenticated by **either** the session cookie **or** the principal header.
- **Admin privilege** — every admin endpoint requires the principal header, and
  administrative privilege itself is enforced **at the proxy**, before requests reach the
  admin API.
- **CORS** — enabled for all `/api/*` routes on the end-user server, to support the browser
  consent app.

### Public endpoints

These endpoints require no pre-authentication:

| Endpoint | Server |
|---|---|
| `GET /health` | Both |
| `GET /oauth2/jwks.json` | End-user |
| `GET /.well-known/oauth-authorization-server` | End-user |

The OAuth2 endpoints `GET /oauth2/authorize` and `POST /oauth2/token` are not pre-auth gated
either — they authenticate through their own OAuth2 parameters (agent client credentials or a
`client_assertion`), not the principal header.

See [Configure authentication](/docs/guides/configure-authentication) for how to establish
the proxy trust boundary.

## Response envelopes

Successful and error responses follow a small set of consistent shapes.

| Shape | Used by | Example |
|---|---|---|
| `{"data": <resource-or-array>}` | Most resource responses | `{"data": {"principal": "…"}}` |
| Bare JSON array | Admin list endpoints `GET /api/agents`, `GET /api/services` | `[{"id": "…"}]` |
| `{"items": [ … ]}` | `GET /api/oauth2-server/signing-keys` | `{"items": [{"kid": "…"}]}` |
| `{"error": "<code>", "message": "<text>"}` | Standard errors (end-user and admin) | `{"error": "agent not found", "message": "…"}` |
| `{"error": "<code>", "error_description": "…"}` | OAuth2 endpoints (RFC 6749 / 8693) | `{"error": "access_denied", "error_description": "…"}` |

:::note
The two error shapes are intentionally different. Standard end-user and admin errors use the
`{error, message}` envelope, where both fields are required. The OAuth2 endpoints
(`/oauth2/token`, `/oauth2/jwks.json`, `/.well-known/*`, and authorize) use the RFC-defined
`{error, error_description}` format instead.
:::

## Error codes

The `error` field carries a machine-readable code. The codes differ by surface: end-user and
admin surfaces use human-readable strings, while the OAuth2 endpoints use the snake_case RFC
tokens.

| Surface | Codes |
|---|---|
| End-user consent / session | `session_expired`, `invalid_permission_set`, `forbidden`, `invalid_state`, `service_id_mismatch`, `unauthorized`, `bad request`, `invalid request`, `invalid scopes`, `service not found` |
| Admin | `invalid request body`, `validation failed`, `agent not found`, `service not found`, `conflict`, `last_key`, `current_key` |
| OAuth2 (`/oauth2/token`) | `invalid_request`, `invalid_client`, `invalid_grant`, `invalid_target`, `access_denied`, `server_error` |
| OAuth2 authorize | `invalid_client`, `invalid_redirect_uri` |

## End-user API map (port 8000)

The full request and response schemas for every endpoint below are in the
[end-user API reference](/api/enduser).

### Health

| Method | Path | Purpose |
|---|---|---|
| GET | `/health` | Server health and lifecycle status (public). |

### User info

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/me` | The current authenticated user's profile. |

### Consent

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/consent/agents` | List agents that have active delegations for the user. |
| GET | `/api/consent/agents/{agent-id}` | Agent detail with its requested third-party services. |
| GET | `/api/consent/agents/{agent-id}/grants` | The user's grants for an agent. |
| POST | `/api/consent/agents/{agent-id}/grants` | Create or update a grant; optionally resume an OAuth2 flow. |
| DELETE | `/api/consent/agents/{agent-id}/grants` | Revoke all of the agent's permissions. |

### Third-party sessions

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/third-party/sessions` | List third-party services and per-user session status. |
| GET | `/api/third-party/{serviceId}/oauth2/authorize` | Start an authorization-code + PKCE flow to the third party. |
| GET | `/api/third-party/{serviceId}/oauth2/callback` | Handle the third-party OAuth2 callback. |
| GET | `/api/third-party/{serviceId}/session` | Session detail and the agents that depend on it. |
| DELETE | `/api/third-party/{serviceId}/session` | Terminate the session and delete its stored tokens. |
| GET | `/api/third-party/{serviceId}/session/affected-agents` | Agents that would lose access if the session ends. |

### OAuth2 server

| Method | Path | Purpose |
|---|---|---|
| GET | `/oauth2/authorize` | RFC 6749 authorization endpoint. |
| POST | `/oauth2/token` | Token exchange, authorization-code, or client-credentials grant. |
| GET | `/oauth2/jwks.json` | Aggregated public JWK set (public). |
| GET | `/.well-known/oauth-authorization-server` | RFC 8414 authorization-server metadata (public). |

For the `/oauth2/token` grant modes and the RFC 8693 field reference, see
[Token exchange](/docs/reference/token-exchange).

## Admin API map (port 14000)

Every admin endpoint requires the `X-Remote-User` header, with administrative privilege
enforced at the proxy. `GET /health` is public. Full schemas are in the
[admin API reference](/api/admin).

### Agents

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/agents` | List all agents (bare array). |
| POST | `/api/agents` | Register an agent. |
| GET | `/api/agents/{agent-id}` | Get an agent. |
| PUT | `/api/agents/{agent-id}` | Update an agent. |
| DELETE | `/api/agents/{agent-id}` | Delete an agent (irreversible). |

### Services

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/services` | List services (bare array; secrets redacted). |
| POST | `/api/services` | Register a third-party OAuth2 service. |
| GET | `/api/services/{service-id}` | Get a service (secret redacted). |
| PUT | `/api/services/{service-id}` | Update a service. |
| DELETE | `/api/services/{service-id}` | Delete a service (`409 conflict` if grants reference it). |

### Permission sets

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/permission-sets` | Create a permission set. |
| GET | `/api/permission-sets` | List permission sets (optionally filtered by service). |
| GET | `/api/permission-sets/{permission-set-id}` | Get a permission set. |
| PUT | `/api/permission-sets/{permission-set-id}` | Replace a permission set. |
| DELETE | `/api/permission-sets/{permission-set-id}` | Delete a permission set. |

### Client credentials

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/agents/{agent-id}/client-credentials` | Generate or rotate an agent's broker credentials. |
| GET | `/api/agents/{agent-id}/client-credentials` | Credential metadata (never the secret). |
| DELETE | `/api/agents/{agent-id}/client-credentials` | Revoke the agent's credentials (irreversible). |

### Signing keys

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/oauth2-server/signing-keys` | Add a signing key. |
| GET | `/api/oauth2-server/signing-keys` | List signing keys (`{items}`, newest first). |
| PUT | `/api/oauth2-server/signing-keys/{kid}/current` | Promote a key to current. |
| DELETE | `/api/oauth2-server/signing-keys/{kid}` | Soft-delete a key (`409 last_key` / `current_key`). |

## Related

- [End-user API reference](/api/enduser) and [Admin API reference](/api/admin) — the full
  generated contracts.
- [Token exchange](/docs/reference/token-exchange) — the RFC 8693 field reference.
- [Configuration](/docs/configuration) — every configuration key.
- [Configure authentication](/docs/guides/configure-authentication) — the proxy trust
  boundary.
