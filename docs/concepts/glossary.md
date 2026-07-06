---
title: Glossary
description: Precise definitions of the terms used across the Agentic Identity Broker documentation.
---

# Glossary

Definitions of the terms used throughout these docs. Terms are grouped by area. For how they
fit together, see [delegation and consent](/docs/concepts/delegation-and-consent).

## Delegation model

**Principal** — the authenticated end user. The broker does not authenticate users; a trusted
reverse proxy authenticates the request and passes the principal identifier in a header
(default `X-Remote-User`).

**Agent** — an AI agent registered in the broker. Identified by a system-generated UUID (used
as its OAuth2 `client_id`), with a display name, description, and optional governance and
documentation URLs. Declares the services and permission sets it needs, each mandatory or
optional.

**Third-party service** (third-party OAuth2 provider) — an external OAuth2 provider the broker
integrates with, such as GitHub or Google. Holds the provider's client ID, an encrypted client
secret, issuer and endpoints, offered scopes, and optional resource URIs for token exchange.

**Scope** — a single permission string defined by a service (for example `repo` or
`user:email`), paired with a human-readable description.

**Permission set** — an administrator-defined bundle of scopes spanning one or more services.
The unit users consent to, expressed in business terms rather than raw provider scopes.

**Requirement type** — whether an agent's need for a service or permission set is **mandatory**
(authorization is blocked until satisfied) or **optional** (the agent degrades gracefully if
it is not granted).

**Grant** (user grant) — the record of a principal delegating permission sets to an agent. At
most one active grant per user and agent; optionally expires (`valid_until`); revocable at any
time.

**User session** — an authenticated OAuth2 session between a principal and one third-party
service, holding the encrypted access and refresh tokens, expiry, and scopes. Exactly one per
user and service.

**Consent flow** — the process where a user reviews an agent's requested access and grants,
adjusts, or revokes it, through the consent UI.

## OAuth2 and token exchange

**Authorization server** — the broker's OAuth2 surface implementing RFC 6749 (authorization
code) and RFC 8414 (server metadata). Depending on [server mode](/docs/concepts/oauth2-server-modes),
it either proxies an upstream server or issues its own tokens.

**Server mode** — one of `proxy` (forward to an upstream authorization server), `local` (issue
tokens locally), or `hybrid` (both, dispatched per agent).

**PKCE** — Proof Key for Code Exchange (RFC 7636). Always required on the broker's
authorization flows, using the `S256` challenge method only.

**Token exchange** — RFC 8693. A privileged gateway swaps an agent's subject token for the
correct third-party token, scoped to a target resource, after the broker verifies the user's
grant.

**Subject token** — in token exchange, the JWT identifying the user (via a configurable claim,
default `sub`) and the agent (default `azp`) on whose behalf the exchange is requested.

**Client assertion** — in token exchange, the JWT a privileged gateway presents to authenticate
itself to the broker. Validated against the upstream authorization server's keys.

**Resource / protected resource** — a URI identifying the target third-party service for a
token exchange. Matched against the `protected_resources` configured on a service.

**Broker client credential** — OAuth2 client credentials the broker itself issues to an agent
for `local` or `hybrid` mode. The secret is hashed at rest, shown once, and rotatable.

**Signing key** — the asymmetric key (ES256) the broker uses to sign locally issued JWT access
tokens. Private material is encrypted at rest and published in the broker's JWKS for
verification.

**JWKS** — JSON Web Key Set (RFC 7517). The broker publishes public keys at
`/oauth2/jwks.json` so tokens it issues (or republishes) can be verified.

**CIMD** — Client ID Metadata Document. An HTTPS URL an agent can use as its `client_id`; the
broker fetches and validates a JSON document there (with SSRF protection) to show trustworthy
metadata on the consent screen.

**CEL** — Common Expression Language. Used for authorizing which gateways may perform token
exchange, extracting identifiers from tokens, and adding custom claims to locally issued
tokens.

## Encryption

**Envelope encryption** — encrypting data with a fresh data key, then encrypting that data key
with a higher-level key. The broker uses this so a single key-management call protects many
tokens. See [encryption at rest](/docs/concepts/encryption).

**KEK / DEK** — Key Encryption Key (the root key, an AWS KMS customer-managed key in production)
and Data Encryption Key (the per-operation key that actually encrypts a token).

**Branch key** — an intermediate key, cached in DynamoDB, between the KEK and the DEK in the
hierarchical keyring. Reduces calls to KMS.

**Encryption context** — non-secret data bound to ciphertext as additional authenticated data,
constrained to a single subject (the service) so a token encrypted for one service cannot be
decrypted for another.

**Token vault** — the encrypted store of third-party OAuth2 tokens held in user sessions.

## Deployment and operations

**Reverse-proxy pre-authentication** — the model where a trusted proxy authenticates the user
and injects the principal header. The broker's baseline authentication mode.

**JWT pre-authentication** — an optional mode where the broker verifies a signed JWT header
(against a JWKS, or unsigned in trusted environments) and extracts a user profile via CEL.

**ExtProc token-exchange sidecar** — a standalone gRPC service implementing Envoy's External
Processor protocol. Performs transparent token exchange (and optional OPA policy checks) at an
Envoy-based agent gateway.

**OPA** — Open Policy Agent. Used by the ExtProc sidecar to authorize proxied requests and MCP
tool calls, independently of the broker's own token-exchange policy.

**IRSA** — IAM Roles for Service Accounts. The AWS mechanism by which broker pods obtain
permission to use the KMS key and DynamoDB table without static credentials.

**Migration job** — a one-shot container image that applies database schema changes with a
least-privilege database user, run separately from the broker service.
