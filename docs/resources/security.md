---
title: "Security posture"
description: "How the Agentic Identity Broker protects delegated credentials: mandatory encryption at rest, redacted secrets, PKCE, single-use codes, consent anti-spoofing, SSRF hardening, the reverse-proxy trust boundary, and how to report a vulnerability."
---

# Security posture

This page summarizes the security controls an operator should know about before running the
broker. The broker's job is to hold third-party credentials on a user's behalf, so its
defaults are built to protect those credentials — security controls are enabled by default,
not optional.

## Encryption at rest

Encryption at rest is **mandatory**. The broker refuses to start if no encryption backend is
configured — there is no plaintext fallback.

- **Envelope encryption.** A root key wraps intermediate keys, which wrap a fresh
  per-operation data key that encrypts each token. In production this is an AWS KMS
  customer-managed key over DynamoDB-cached branch keys; for development it is a raw AES-256
  key held in memory.
- **Per-service context binding.** Every layer binds an encryption context (additional
  authenticated data) tied to a single subject — the `service_id` for service tokens and
  secrets. A ciphertext encrypted for one service cannot be decrypted for another.
- **Fail-closed.** If encryption or decryption cannot be performed correctly, the operation
  fails rather than falling back to plaintext.

See [Encryption](/docs/concepts/encryption) for the full model and
[Configure encryption](/docs/guides/configure-encryption) for setup.

## Secret handling

- Provider client secrets are **encrypted before storage** and are treated as a distinct
  value type that is either plaintext or ciphertext — the type system enforces that a secret
  is encrypted before it reaches storage.
- Secrets are **always redacted in API responses.** Reading a service returns `REDACTED` in
  place of its client secret; the value is never echoed back over the API.
- Broker-issued agent credentials are shown **once**, at creation or rotation, and are stored
  hashed. The API can return credential metadata afterward, never the secret.

## OAuth2 flow protections

- **PKCE is always required, S256 only.** The broker's authorization endpoint requires a
  `code_challenge_method` of `S256`; the `plain` method is not accepted.
- **Authorization codes are single-use and short-lived.** A code is valid for one exchange
  and expires quickly.
- **`client_id` is the agent's internal identifier.** On the authorization endpoint,
  `client_id` is the agent's system-generated UUID — not a reusable upstream OAuth2 client id.

## Consent anti-spoofing

When a consent flow is entered mid-authorization, its context — which agent, which principal,
the original authorization URL, and any client-metadata document — is **sealed server-side in
a short-lived encrypted (JWE) session token** rather than passed as browser-visible
parameters. The consent screen cannot be spoofed with attacker-supplied values, and the
broker rejects a session whose principal does not match the authenticated user.

## SSRF hardening

Agents may identify themselves by an HTTPS URL pointing to a Client ID Metadata Document
(CIMD), which the broker fetches to display trustworthy metadata on the consent screen. These
outbound fetches are **SSRF-hardened** — constrained by a configurable blocklist — so an
agent-supplied URL cannot be used to reach internal network targets.

## Trust boundary and transport

- **Reverse-proxy pre-authentication.** The broker does not authenticate users. A trusted
  reverse proxy authenticates the caller and injects the principal in the `X-Remote-User`
  header (configurable). The broker trusts that header only from a trusted source. Deploy the
  broker so its ports are reachable only through that proxy.
- **Admin privilege at the proxy.** Administrative authorization is enforced at the proxy in
  front of the admin server (port 14000), before requests reach the API.
- **HTTPS expectations.** Service issuer URIs must be HTTPS, and tokens must only cross the
  network over TLS — terminate TLS in front of the broker. A development-only flag exists to
  skip third-party HTTPS validation; it must never be enabled in production.

See [Configure authentication](/docs/guides/configure-authentication) for how to establish the
proxy trust boundary.

## Reporting a vulnerability

Please disclose security issues responsibly rather than opening a public issue.

The project participates in **Zalando's responsible-disclosure and bug-bounty program**. To
report a vulnerability — with or without joining the bug-bounty program — submit it through
the Zalando vulnerability-reporting form:

- **[Report a vulnerability](https://corporate.zalando.com/en/about-us/report-vulnerability)**

The maintainers acknowledge that any code may contain security issues and aim to provide
patches as quickly as possible. Please allow time for a fix before public disclosure.
