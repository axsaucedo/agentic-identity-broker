# ADR 017: Optional Agent client_id with UUID Auto-Generation

**Status**: Accepted
**Date**: 2026-05-10

---

## Context

The `client_id` field on Agent was introduced when all agents proxied to an upstream OAuth2 provider — the field held the credential the upstream provider recognizes. With the addition of local token minting mode (ADR 014) and CIMD support (spec 028), many agents never use `client_id` for upstream proxying:

- **Local token minting agents** use the agent's UUID as the OAuth2 `client_id` parameter at runtime (`OpaqueClientResolver` parses the request `client_id` as a UUID and calls `repo.Get()`). The stored `ClientID` field is unused in this flow.
- **CIMD agents** are resolved via `GetByClientURI` against pre-registered `client_uris`. The stored `ClientID` is again not consulted during authorization.

Despite this, the Admin API requires operators to supply a `client_id` for every agent. For agents that don't proxy upstream, operators must invent a meaningless value.

## Decision

Make `client_id` optional in the Admin API `AgentCreateRequest`. When omitted, auto-generate it as the agent's UUID string (e.g. `550e8400-e29b-41d4-a716-446655440000`).

When explicitly provided, use the operator-supplied value unchanged (for upstream proxy scenarios).

### Why UUID, not a prefixed format

- The `OpaqueClientResolver` already treats the OAuth2 request `client_id` as the agent's UUID. Setting the stored field to the same value makes the stored and runtime identifiers consistent.
- The `resolveAgentIdByClientId` CEL helper (used when `multi_agent_client.enabled = false`) calls `GetByClientID` — a UUID-valued `ClientID` makes this a valid (if tautological) lookup.
- No prefix avoids introducing a new naming convention to document and parse.

## Consequences

- Operators no longer need to supply `client_id` for local or CIMD agents.
- Auto-generated values are inherently unique (UUID-based), so existing uniqueness checks need no modification.
- Forward-compatible with hybrid setups: operators who need a real upstream `client_id` alongside `client_uris` can still supply one explicitly.
- No database migration required — the `client_id` column remains NOT NULL; the default is applied at the application layer before domain validation.
- Existing agents are unaffected; no data migration needed.
- On update, omitting `client_id` preserves the existing value.
