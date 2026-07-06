---
title: "Operate the OAuth2 server modes"
description: "Configure the broker's proxy, local, and hybrid OAuth2 authorization-server modes, manage ES256 signing keys, and mint custom token claims with CEL."
---

# Operate the OAuth2 server modes

The broker exposes an OAuth2 authorization-server surface on the end-user port (8000):
`/oauth2/authorize`, `/oauth2/token`, `/oauth2/jwks.json`, and
`/.well-known/oauth-authorization-server`. It runs in exactly one **mode**, which decides
where authorization requests and tokens come from:

- **proxy** (default) — forward `/oauth2/authorize` and `/oauth2/token` to an upstream OAuth2
  server, while still serving broker metadata and a JWKS that republishes upstream keys.
- **local** — the broker is a standalone authorization server that mints its own JWT access
  tokens signed with managed ES256 keys.
- **hybrid** — both, dispatched per request based on how each agent is registered.

This is the operator how-to companion to the [OAuth2 server modes](/docs/concepts/oauth2-server-modes)
concept page. It shows what to set for each mode, how to manage signing keys through the
admin API, and how to shape custom claims with CEL.

## What you'll need

- A decision on which mode fits your deployment. If you already run a corporate OAuth2 server
  and want the broker to defer to it, choose **proxy**. If you want the broker to issue its
  own tokens for agents, choose **local**. To support both kinds of agent in one deployment,
  choose **hybrid**.
- For proxy or hybrid: the upstream server's issuer, authorization, and token endpoints.
- For local or hybrid: an encryption backend configured — the broker encrypts signing-key
  private material at rest and refuses to start in these modes without one. See
  [configure encryption](/docs/guides/configure-encryption).
- Access to the broker's YAML configuration and the admin API on port 14000.

All modes require PKCE `S256` on the authorization endpoint, and the `client_id` on
`/oauth2/authorize` is the agent's **UUID** (`agent.id`) — not the upstream OAuth2 client id.

## Proxy mode

Set `mode: proxy` and configure the upstream endpoints under `proxy`. The broker forwards
authorization and token requests upstream after checking the user's grant, and republishes
the upstream signing keys through its own JWKS. The `upstream_jwks_min_refresh` and
`upstream_jwks_max_refresh` bounds cap how often the broker re-fetches those keys.

```yaml
oauth2_authorization_server:
  mode: "proxy"
  proxy:
    upstream_issuer_uri: "https://auth.example.com"
    upstream_authorize_endpoint: "https://auth.example.com/authorize"
    upstream_token_endpoint: "https://auth.example.com/token"
    upstream_timeout: 30s
    upstream_jwks_min_refresh: "15m"
    upstream_jwks_max_refresh: "1h"
  supported_response_types:
    - code
  supported_grant_types:
    - authorization_code
```

The broker's own public URL — used as the `issuer` in the discovery metadata — comes from
`server.enduser.public_url`, not from this section.

## Local mode

Set `mode: local` and configure token issuance under `local`. In this mode the broker mints
its own JWT access tokens, so there are no upstream endpoints. It supports the
`client_credentials` and `authorization_code` (with PKCE) grants.

```yaml
oauth2_authorization_server:
  mode: "local"
  local:
    token_ttl: "1h"
    token_claims_expression: '{"team": agent.display_name}'
    signing_keys:
      bootstrap_timeout: 90s
```

- `token_ttl` — how long issued access tokens are valid (a Go duration such as `30m`, `1h`,
  `1h30m`).
- `token_claims_expression` — an optional CEL expression that adds custom claims (see
  [Custom token claims with CEL](#custom-token-claims-with-cel)). Empty means no custom
  claims.
- `signing_keys.bootstrap_timeout` — the startup budget for bringing up the first signing key.

Local mode needs an encryption backend so it can protect signing-key private material:

```yaml
encryption:
  memory:
    raw_key: "base64-encoded-32-byte-key" # use aws_kms for production
```

## Hybrid mode

Set `mode: hybrid` and configure **both** the `proxy` and `local` sections — both are
required. The broker classifies each agent by how it was registered and routes accordingly:
agents with an upstream client id go through the proxy path; agents registered as local (or
by CIMD) get locally issued tokens.

```yaml
oauth2_authorization_server:
  mode: "hybrid"
  proxy:
    upstream_issuer_uri: "https://auth.example.com"
    upstream_authorize_endpoint: "https://auth.example.com/authorize"
    upstream_token_endpoint: "https://auth.example.com/token"
    upstream_timeout: 30s
    upstream_jwks_min_refresh: "15m"
    upstream_jwks_max_refresh: "1h"
  local:
    token_ttl: "1h"
    token_claims_expression: ""
    signing_keys:
      bootstrap_timeout: 90s
  supported_response_types:
    - code
  supported_grant_types:
    - authorization_code
    - client_credentials
```

As with local mode, hybrid mode requires an encryption backend for signing-key material.

## Manage signing keys

Local and hybrid modes sign tokens with **ES256** keys the broker manages. The private
material is encrypted at rest; the public keys are published at `/oauth2/jwks.json` so
verifiers can validate tokens. Rotate keys through the admin API on port 14000. Every request
carries the principal header your proxy injects (see
[configure authentication](/docs/guides/configure-authentication)); administrative privilege
is enforced at the proxy.

### Add a key

`POST /api/oauth2-server/signing-keys` creates a key. `ES256` is the default and the only
supported algorithm.

```bash
curl -X POST https://broker.internal:14000/api/oauth2-server/signing-keys \
  -H "X-Remote-User: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{"algorithm": "ES256"}'
```

A new key is published to the JWKS immediately and marked as current, but it does **not**
start signing tokens until its `activates_at` timestamp — a grace period (twice the JWKS
cache lifetime, 600 seconds) that lets verifiers pick up the new public key before any token
is signed with it. The response includes `kid`, `algorithm`, `is_current`, `activates_at`,
and `created_at`.

### List keys

`GET /api/oauth2-server/signing-keys` returns the keys, newest first, wrapped in an `items`
array.

```bash
curl https://broker.internal:14000/api/oauth2-server/signing-keys \
  -H "X-Remote-User: admin@example.com"
```

### Promote a key to current

`PUT /api/oauth2-server/signing-keys/{kid}/current` promotes an existing key. The previously
current key becomes non-current but stays valid for verification, so tokens it already signed
keep working until they expire.

```bash
curl -X PUT https://broker.internal:14000/api/oauth2-server/signing-keys/{kid}/current \
  -H "X-Remote-User: admin@example.com"
```

### Retire a key

`DELETE /api/oauth2-server/signing-keys/{kid}` soft-deletes a key. The broker refuses to
delete the last remaining key (`409 last_key`) or the current key (`409 current_key`) — add
or promote a replacement first.

```bash
curl -X DELETE https://broker.internal:14000/api/oauth2-server/signing-keys/{kid} \
  -H "X-Remote-User: admin@example.com"
```

:::tip
Rotate by adding a new key, waiting out its activation grace period so verifiers have the new
public key, promoting it to current, and only then retiring the old one once its outstanding
tokens have expired.
:::

## Custom token claims with CEL

In local and hybrid modes you can add claims to the JWTs the broker issues with
`local.token_claims_expression`. The expression is CEL and must return a map of claim names
to values. It has three variables available:

| Variable | Type | Contents |
|---|---|---|
| `agent` | struct | The agent the token is issued for (for example `agent.display_name`). |
| `principal` | string | The authenticated user the token acts for. |
| `request` | map | The token request context. |

```yaml
oauth2_authorization_server:
  local:
    token_claims_expression: '{"team": agent.display_name}'
```

The base OAuth2 claims (`iss`, `sub`, `exp`, and the other standard registered claims) are
set by the broker and **cannot be overridden** by this expression — it only adds claims
alongside them.

## Discovery and JWKS

Regardless of mode, the broker publishes two public endpoints on the end-user port that
clients and gateways use to discover the server and verify tokens:

- `GET /.well-known/oauth-authorization-server` — RFC 8414 metadata (issuer, endpoints,
  supported response and grant types, `code_challenge_methods_supported: [S256]`).
- `GET /oauth2/jwks.json` — the aggregated public JWK set. In proxy mode it republishes the
  upstream keys; in local and hybrid modes it publishes the ES256 keys you manage above.

Point your agents and any token-verifying gateway at these endpoints rather than hard-coding
keys or URLs, so key rotation is transparent to them.

## Related

- [OAuth2 server modes](/docs/concepts/oauth2-server-modes) — the concept behind these modes.
- [Manage agents and services](/docs/guides/manage-agents-and-services) — register the agents
  that authenticate against this server.
- [End-user API reference](/api/enduser) — the authorize, token, JWKS, and discovery
  contracts.
