---
title: Authorization
description: Authorization features
---

# Authorization

Authorization in this repository is split across two independent, conjunctive policy gates on the same request chain.

| Gate | Service / Boundary | Question Answered | Inputs | Config Surface | Default |
|------|--------------------|-------------------|--------|----------------|---------|
| ExtProc OPA | `extproc-token-exchange` request path | May this proxied request or MCP tool call proceed? | `OPAInput` built from ExtProc metadata, headers, and optional request body | `authorization.*` in ExtProc config | Disabled |
| Broker CEL | Broker `POST /oauth2/token` token-exchange boundary | May this gateway perform token exchange for this resource? | `CELAuthorizationContext` built from client assertion claims and RFC 8693 request fields | `token_exchange.authorization.cel.*` in broker config | CEL expression defaults to `true` |

The gates compose as fail-closed AND:

- ExtProc OPA can deny a proxied request before or after token exchange depending on whether the request is header-only or body-bearing.
- Broker CEL can still deny RFC 8693 token exchange even when ExtProc OPA would allow the proxied request.
- Neither layer replaces the other because they inspect different inputs at different trust boundaries.

**Body-bearing requests**: ExtProc exchanges the token in `RequestHeaders`, then evaluates OPA in `RequestBody`.

**Header-only requests**: ExtProc evaluates OPA before any broker call and exchanges the token only on allow.

`context.granted_permission_sets` in the ExtProc OPA input is token-bound snapshot data, not a live second authority. For body-bearing requests it comes from the RFC 8693 exchange response and is cached alongside the exchanged access token, so it inherits the same `cache.max_ttl` revocation window. For header-only requests, OPA runs before token exchange and therefore sees `context.granted_permission_sets = {}` during that pre-exchange evaluation.
