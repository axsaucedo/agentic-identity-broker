---
title: "Reference"
description: "The API contracts, token-exchange field reference, and configuration reference for the Agentic Identity Broker — how the dual-port API is structured and authenticated."
---

# Reference

This section is the precise, contract-level documentation for operating and integrating
against the Agentic Identity Broker. Use it when you already understand the
[concepts](/docs/concepts) and need exact endpoints, fields, error codes, and configuration
keys.

## What's here

| Reference | What it covers |
|---|---|
| [End-user API](/api/enduser) | The full OpenAPI contract for the end-user server (port 8000): health, user info, consent, third-party sessions, and the OAuth2 server surface. |
| [Admin API](/api/admin) | The full OpenAPI contract for the admin server (port 14000): agents, services, permission sets, client credentials, and signing keys. |
| [API overview](/docs/reference/api) | A hand-authored companion to the two contracts — the auth model, response envelopes, error codes, and a compact endpoint map. |
| [Token exchange](/docs/reference/token-exchange) | The field-level reference for the RFC 8693 grant on `POST /oauth2/token`. |
| [Configuration](/docs/configuration) | Every configuration section and key, and how the configuration sources compose. |

The two API pages ([end-user](/api/enduser) and [admin](/api/admin)) are rendered directly
from the source OpenAPI specifications, so they are always the authoritative contract. The
[API overview](/docs/reference/api) explains the conventions those contracts share.

## How the API is organized

The broker exposes **two independent HTTP servers on separate ports**, so that end-user and
administrative traffic can be routed, firewalled, and exposed independently:

- **End-user API — port 8000.** The consent surface (`/api/consent/*`), third-party sessions
  (`/api/third-party/*`), user info (`/api/me`), and the OAuth2 server endpoints
  (`/oauth2/*`, `/.well-known/*`).
- **Admin API — port 14000.** Registration and lifecycle for agents, services, permission
  sets, per-agent client credentials, and the broker's signing keys.

## How authentication works

The broker does not authenticate users itself. A **trusted reverse proxy** (oauth2-proxy,
nginx `auth_request`, a service mesh) authenticates the caller and injects the principal in a
request header — **`X-Remote-User`** by default, and configurable. The broker trusts that
header only from a trusted source. On the admin server, administrative privilege is enforced
**at the proxy** before requests reach the API.

A small set of endpoints is public and needs no pre-authentication: `GET /health` (both
servers), `GET /oauth2/jwks.json`, and `GET /.well-known/oauth-authorization-server`. The
OAuth2 endpoints `GET /oauth2/authorize` and `POST /oauth2/token` authenticate through their
own OAuth2 parameters rather than the principal header.

See [Configure authentication](/docs/guides/configure-authentication) to set up the proxy
trust boundary, and the [API overview](/docs/reference/api) for the details of every
convention above.
