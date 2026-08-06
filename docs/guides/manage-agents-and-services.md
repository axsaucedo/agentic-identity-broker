---
title: "Manage agents and services"
description: "Register third-party OAuth2 services, define permission sets, register agents, and issue broker client credentials through the Agentic Identity Broker admin API on port 14000."
---

# Manage agents and services

This guide walks an administrator through the objects that make delegation possible, using the
**admin API on port 14000**. You register a third-party service, bundle its scopes into
permission sets, register an agent that requests those sets, and — for the broker's local and
hybrid server modes — issue the agent a set of broker client credentials.

The order matters: a permission set references a service, and an agent references permission
sets, so you create them in that sequence.

## What you'll need

- **Admin access through your proxy.** The broker does not authenticate users. A trusted
  reverse proxy in front of the admin port authenticates the caller, enforces admin
  privilege, and injects the principal header (`X-Remote-User` by default). The broker trusts
  that header only from a trusted source. The examples below send the header directly, which
  works against a local development stack; in production your proxy sets it.
- **The admin base URL.** These examples use `http://localhost:14000`, the admin port on a
  local stack. In a real deployment this is the internal admin endpoint your proxy fronts —
  never the public, end-user port `8000`.

Every request carries the principal header:

```bash
-H "X-Remote-User: admin@example.com"
```

## Register a third-party OAuth2 service

A third-party service is the external OAuth2 provider the broker integrates with — GitHub,
Google, an internal API. Create one with `POST /api/services`.

The core fields:

- `display_name` — the human-readable name shown in the admin surface.
- `client_id` / `client_secret` — the OAuth2 client credentials the provider issued you. The
  secret is **encrypted at rest** (AES-256-GCM) and returned as `"REDACTED"` on every read.
- `issuer_uri` — the provider's issuer, which must be an **HTTPS** URL.
- `discovery.enable_discovery` — when `true`, the broker fetches the provider's endpoints from
  `{issuer_uri}/.well-known/oauth-authorization-server`. When `false`, you supply
  `endpoints.token_endpoint` and `endpoints.authorize_endpoint` explicitly.
- `scopes` — optional scopes the provider offers, each as `{scope_value, description}`.
  Omit the field or use an empty list for a provider that does not use OAuth2 scopes; the
  broker then omits `scope` from its upstream authorization request.
- `protected_resources` — optional resource URIs used to route
  [token exchange](/docs/concepts/delegation-and-consent#using-a-delegation) to this service.
- `authorization_params` — optional static provider parameters added by the broker to upstream authorization requests. For Zalando Platform, use `{ "business_partner_id": "12345" }`. These are administrator configuration, never browser-supplied values; omit the field on update to preserve it, or use `{}` to clear it.

With endpoint discovery enabled:

```bash
curl -X POST http://localhost:14000/api/services \
  -H "X-Remote-User: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "GitHub",
    "client_id": "Iv1.1234567890abcdef",
    "client_secret": "ghp_secretkey1234567890abcdef",
    "issuer_uri": "https://github.com",
    "discovery": { "enable_discovery": true },
    "scopes": [
      { "scope_value": "repo", "description": "Full control of private repositories" },
      { "scope_value": "read:org", "description": "Read organization membership" }
    ],
    "protected_resources": [ "https://api.github.com" ]
  }'
```

For a provider without discovery, set `enable_discovery` to `false` and pass the endpoints:

```bash
curl -X POST http://localhost:14000/api/services \
  -H "X-Remote-User: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Corporate SSO",
    "client_id": "corp-sso-client-123",
    "client_secret": "secret-value-never-exposed",
    "issuer_uri": "https://sso.corp.example.com",
    "discovery": { "enable_discovery": false },
    "endpoints": {
      "token_endpoint": "https://sso.corp.example.com/oauth/token",
      "authorize_endpoint": "https://sso.corp.example.com/oauth/authorize"
    },
    "scopes": [
      { "scope_value": "profile", "description": "Basic user profile" }
    ]
  }'
```

For an IdP that does not accept scopes, omit `scopes` entirely:

```json
{
  "display_name": "Corporate session IdP",
  "client_id": "corporate-session-client",
  "client_secret": "secret-value-never-exposed",
  "issuer_uri": "https://idp.corp.example.com",
  "discovery": { "enable_discovery": false },
  "endpoints": {
    "token_endpoint": "https://idp.corp.example.com/oauth/token",
    "authorize_endpoint": "https://idp.corp.example.com/oauth/authorize"
  }
}
```

Include that service in a permission set with `"scopes": []`. This still requires a user
session when included as mandatory, but it does not require a scope match.

The response returns the created service with a system-generated `id` (a UUID) and the
`client_secret` shown as `"REDACTED"`. Record the `id` — permission sets reference it.

:::note
Deleting a service is blocked while grants still reference it: `DELETE /api/services/{service-id}`
returns **409 `conflict`** so you never leave dangling delegations. Revoke or migrate the
dependent grants first.
:::

## Define permission sets

A permission set is a business-readable bundle of scopes spanning one or more services. This
is the unit users actually consent to — they approve "Read repositories," not raw provider
scope strings. Create one with `POST /api/permission-sets`.

The fields:

- `name` — a unique, human-readable name (up to 255 characters).
- `description` — what capabilities the set grants, in the language you want users to see.
- `service_scopes` — one or more entries, each `{service_id, scopes[], requirement_type}`,
  where `service_id` is a service UUID from the previous section. `scopes` may be omitted or
  empty for a scope-less service; otherwise its non-empty values must match the service's
  configured scopes. `requirement_type` is `mandatory` or `optional`.

```bash
curl -X POST http://localhost:14000/api/permission-sets \
  -H "X-Remote-User: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "GitHub Read Access",
    "description": "Read repository contents and user profile from GitHub",
    "service_scopes": [
      {
        "service_id": "550e8400-e29b-41d4-a716-446655440001",
        "scopes": [ "repo", "user:email" ],
        "requirement_type": "mandatory"
      }
    ]
  }'
```

The response returns the permission set with its own `id` (UUID) plus `created_at` and
`updated_at`. An agent references this `id`. Changing a set's `service_scopes` later takes
effect on the next token exchange for any grant that references it.

## Register an agent

An agent is the AI agent that requests delegated access. Create one with `POST /api/agents`.

The fields:

- `display_name` and `description` — required; shown to users at consent time.
- `permission_sets` — required, at least one entry, each
  `{permission_set_id, requirement_type}`. `mandatory` sets block authorization until granted;
  `optional` sets let the agent degrade gracefully. See
  [delegation and consent](/docs/concepts/delegation-and-consent#mandatory-vs-optional-requirements).
- `client_id` — optional. An upstream OAuth2 client id, only for agents that use one. Omit it
  for agents resolved by other means.
- `governance_url`, `user_documentation_url`, `agent_interface_url` — optional URLs surfaced
  to users on the consent screen.

```bash
curl -X POST http://localhost:14000/api/agents \
  -H "X-Remote-User: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Research Assistant",
    "description": "AI assistant that reads repositories and datasets on your behalf",
    "governance_url": "https://example.com/agents/research-assistant/governance",
    "user_documentation_url": "https://example.com/docs/research-assistant",
    "permission_sets": [
      {
        "permission_set_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "requirement_type": "mandatory"
      }
    ]
  }'
```

The response returns the agent with a system-generated `id` (a UUID). **That `id` is the
agent's canonical identifier and the `client_id` you use on the broker's `/oauth2/authorize`
endpoint** — not the optional upstream `client_id` field. The broker does not auto-generate an
upstream `client_id`.

## Issue broker client credentials

When the broker acts as its own OAuth2 authorization server — in `local` or `hybrid`
[server mode](/docs/guides/operate-oauth2-server-modes) — an agent authenticates to the token
endpoint with credentials the broker issues. Generate them per agent with
`POST /api/agents/{agent-id}/client-credentials`, using the agent's UUID in the path.

```bash
curl -X POST http://localhost:14000/api/agents/550e8400-e29b-41d4-a716-446655440000/client-credentials \
  -H "X-Remote-User: admin@example.com"
```

The response returns:

```json
{
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_secret": "brk_sec_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "created_at": "2025-12-19T10:30:00Z"
}
```

Note the shape:

- `client_id` equals the agent's UUID — one credential set per agent.
- `client_secret` is prefixed `brk_sec_` and is **shown only in this response**. Store it
  securely; you cannot read it back. `GET …/client-credentials` returns metadata only, never
  the secret.
- Calling `POST` again **rotates** the credential: it returns `200` with a new secret and a
  `previous_invalidated_at` timestamp. Tokens issued under the previous secret stay valid until
  they expire.

## Provider-specific configuration hints

### Google

Google requires two `authorization_params` to support token refresh:

```json
"authorization_params": {
  "access_type": "offline",
  "prompt": "consent"
}
```

- **`access_type: offline`** tells Google to issue a refresh token alongside the access token.
  Without it, Google returns only a short-lived access token and the broker cannot refresh the
  session when it expires.
- **`prompt: consent`** forces the consent screen on every authorization. Google only returns a
  refresh token on the **first** consent grant for a given user–client pair. If the user has
  previously authorized the same OAuth2 client (even outside the broker), Google silently skips
  the refresh token in subsequent responses. Adding `prompt: consent` ensures a fresh refresh
  token is issued every time.

If either parameter is missing, the broker stores the session without a refresh token. Once the
access token expires (typically after one hour), the broker returns `invalid_grant` with the
message *"User session has expired. All tokens are no longer valid"* and the user must
re-authenticate.

:::tip
After adding or changing `authorization_params`, existing sessions are unaffected. Users with
sessions that were created without a refresh token must re-authenticate to pick up the new
parameters.
:::

## Related

- **[/api/admin](/api/admin)** — the full admin OpenAPI reference: every field, response
  schema, and error code.
- **[Delegation and consent](/docs/concepts/delegation-and-consent)** — how the objects you
  create here become grants and sessions.
- **[Operate OAuth2 server modes](/docs/guides/operate-oauth2-server-modes)** — when an agent
  needs broker client credentials, and how proxy, local, and hybrid modes differ.
