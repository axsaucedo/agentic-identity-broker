---
title: Changelog
description: Notable changes, breaking changes, and migration guidance for the Agentic Identity Broker.
---

# Changelog

## [NEXT VERSION] — Breaking Changes

### Root-Mounted Consent SPA (038)

- **BREAKING**: The canonical consent browser paths start at `/`. The sessions page is `/sessions`. Former `/consent` paths do not receive server-side redirects.
- **Confirmation**: The user approved ADR 035 in this conversation.

### Protected Resource Subresources (035)

- **BREAKING**: Full-service `PUT /api/services/{id}` preserves protected resources when `protected_resources` is omitted. Supplying the field, including an empty set, replaces the set only with a current strong `If-Match` ETag; requests without it receive `428`, and stale ETags receive `412` without changing the set.
- **NEW**: Administrators can add protected resources through `POST /api/services/{id}/protected-resources` with a `resource_uri` body field or the retained idempotent member-addressed `PUT`; they can remove, rename, and list resources through dedicated service subresource endpoints.
- **Confirmation**: Stakeholder approved these API changes, including the body-based POST add operation, in this conversation before implementation.

### Multi-Agent OAuth2 Client Delegation (021)

- **BREAKING**: The `client_id` parameter in OAuth2 authorize/token requests now resolves to `agent.id` (UUID), not `agent.client_id` (upstream client ID). Clients must update their `client_id` values from the upstream OAuth2 client ID string to the broker-internal agent UUID.
- **BREAKING**: Token exchange CEL expression `agent_id_expression` must be updated from `subject_token.azp` to `resolveAgentIdByClientId(subject_token.azp)` (or a claim-based expression when the feature is enabled). The previous expression resolved agents by `client_id`; token exchange now resolves by `agent.id` (UUID).
- **NEW**: `multi_agent_client` configuration block under `oauth2_authorization_server` enables multiple agents to share one upstream OAuth2 client ID. See `examples/config/oauth2-authorization-server.yaml` and `docs/configuration.md` for configuration details.
- **NEW**: CEL helper function `resolveAgentIdByClientId(clientId string) string` is available in token exchange CEL policies when `multi_agent_client.enabled = false`. It maps an upstream `client_id` string to the corresponding broker `agent.id` UUID.

#### Migration Guide

**Step 1: Update OAuth2 `client_id` values**

OAuth2 clients integrating with the broker must update the `client_id` parameter in authorize and token requests from the upstream OAuth2 client ID string to the agent's internal UUID:

```
# Before
client_id=my-upstream-app

# After
client_id=a1b2c3d4-e5f6-7890-abcd-ef1234567890  # agent.id UUID
```

**Step 2: Update token exchange CEL expression**

Operators must update `token_exchange.claim_extraction.agent_id_expression` in their configuration:

```yaml
# Before (no longer valid — resolves by client_id, not agent.id)
agent_id_expression: "subject_token.azp"

# After — feature disabled (default): resolve upstream client_id → agent.id UUID
agent_id_expression: "resolveAgentIdByClientId(subject_token.azp)"

# After — feature enabled: use the agent ID claim injected by the broker
agent_id_expression: "subject_token.x_agent_id"  # use your configured claim name
```

**Step 3 (optional): Enable multi-agent client sharing**

If you want multiple agents to share one upstream OAuth2 client ID, add the following to your configuration under `oauth2_authorization_server`:

```yaml
multi_agent_client:
  enabled: true
  agent_id_param_name: "x_agent_id"   # query param appended to upstream authorize redirect
  agent_id_claim_name: "x_agent_id"   # JWT claim in upstream token carrying the agent's internal UUID
```

No action is required if you leave `multi_agent_client.enabled = false` (default). Existing per-agent `client_id` uniqueness enforcement continues to apply.
