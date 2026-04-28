# ADR 016: Server-Side Authorization Sessions — Consent Screen Anti-Spoofing via Capability Tokens

**Status**: Accepted
**Date**: 2026-04-28

---

## Context

When an agent uses a CIMD URL-format `client_id`, the OAuth2 authorization flow must display the agent's identity metadata (client name, logo, redirect URIs, scopes) on the consent screen so the user can make an informed delegation decision.

A naïve implementation passes this metadata as URL query parameters on the consent page redirect (e.g., `?client_name=...&logo_uri=...&redirect_uri=...&scope=...`). This creates a consent screen spoofing vector: an attacker can craft a consent URL with misleading metadata — a different display name, a different logo, or inflated scopes — and trick the user into granting permissions they would not otherwise approve. The frontend, which renders whatever the URL parameters say, becomes complicit in the phishing attack.

This is distinct from the SSRF risk addressed by ADR 015 (fetcher architecture). ADR 015 secures the *inbound* fetch of the CIMD document. This ADR secures the *outbound* delivery of that metadata to the consent UI.

---

## Decision

### Server-Side Authorization Session

For every CIMD-based authorization request, the `OAuth2AuthorizationService` creates a server-side `AuthorizationSession` record that captures the full authorization context at the moment the broker processes the `/authorize` request:

- Agent identity (`agent_id`, `client_id`)
- Authorization parameters (`redirect_uri`, `scope`, `state`, PKCE `code_challenge` + `code_challenge_method`)
- The original authorize URL (`original_url`) — used as the redirect target after consent
- A `CIMDMetadataSnapshot` containing the trusted CIMD document fields (`client_name`, `logo_uri`, `redirect_uris`, `auth_method`, `jwks_uri`) as resolved and validated by the CIMD service

The session is persisted in the `authorization_sessions` table. The consent redirect URL contains only an opaque `session_id` parameter — no authorization context leaks into the URL.

### Capability Token Properties

The `session_id` is an opaque capability token with the following mandatory properties:

| Property | Specification | Rationale |
|---|---|---|
| **Entropy** | 256 bits from `crypto/rand`, encoded as 64-character lowercase hex | OWASP and NIST SP 800-63B recommend ≥128 bits for session tokens; 256 bits is the industry standard that future-proofs against advances in attack capability at zero additional cost |
| **TTL** | 10 minutes from creation | Consent decisions are interactive and immediate; 10 minutes accommodates page load delays and user deliberation without creating a wide replay window |
| **Single-use** | Consumed on grant creation (approve/deny), not on read | The user may reload the consent page, navigate away and return, or experience network retries — all must succeed as long as the session is live. Consumption is bound to the *action* (consent submission), not the *observation* (page load) |
| **Agent-bound** | `agent_id` validated on every session access | Prevents a session created for agent A from being presented on agent B's consent page |

### Consumption Semantics

The session lifecycle has three terminal states:

1. **Consumed** — The user submitted a consent decision (approve or deny). The `consumed_at` timestamp is set. The session cannot be reused.
2. **Expired** — The 10-minute TTL elapsed without a consent decision. The session is rejected on any subsequent access.
3. **Abandoned** — The user closed the browser or navigated away. Functionally identical to expired; the TTL provides natural garbage collection.

Between creation and a terminal state, the session is valid for repeated reads. This allows the consent page to fetch authorization context multiple times (initial load, client-side navigation, retry after transient error) without invalidating the session.

### Non-CIMD Flows

For opaque `client_id` flows (UUID-format agent identifiers), the metadata displayed on the consent screen is static and server-controlled (it comes from the pre-registered agent record, not from an external document). These flows continue using URL parameters for backward compatibility, since there is no external-origin metadata to spoof.

### Redirect Target Integrity

After consent submission, the frontend needs the original authorize URL to re-enter the OAuth2 flow. This URL is stored in the session's `OriginalURL` field and returned in the grant creation response. The frontend never constructs or modifies this URL — it receives it as an opaque string from the server.

---

## Consequences

**Benefits**:
- Eliminates consent screen spoofing for CIMD flows — the frontend renders only server-attested metadata
- The `session_id` is the sole URL-visible artifact; no authorization parameters leak into browser history, referrer headers, or server logs of intermediate proxies
- The pattern is well-understood (same mechanism as magic-link tokens, OAuth authorization codes, CSRF tokens) and requires no novel cryptography
- 256-bit tokens provide computational infeasibility of brute-force even against future adversaries

**Trade-offs**:
- Requires server-side storage (one row per in-flight authorization) — bounded by the 10-minute TTL and expected concurrency
- Non-atomic grant creation + session consumption — if `Consume()` fails after a successful grant, the session remains unconsumed until TTL expiry. This is a narrow window (the `Consume` operation is a simple UPDATE) and is logged for observability
- Deep-linking directly to a consent URL without going through `/authorize` is intentionally impossible for CIMD flows — this is a feature, not a limitation
- Expired sessions accumulate until cleaned up; the `DeleteExpired()` repository method exists for this purpose and must be called periodically

**New invariant**: For CIMD flows, all authorization context displayed on the consent screen MUST originate from the server-side `AuthorizationSession`. Any future consent UI path that bypasses session lookup for CIMD agents violates this ADR.
