---
title: Concepts
description: The mental model behind the Agentic Identity Broker — how delegation, consent, server modes, token exchange, and encryption fit together.
---

# Concepts

This section explains how the broker works and why it is shaped the way it is. It is written
for readers deciding whether the broker fits their architecture, and for operators who need
an accurate mental model before deploying it.

Start with the [delegation and consent model](/docs/concepts/delegation-and-consent) — it is
the core of everything else. Then read whichever of the remaining pages match your interest.

## The pages

- **[Architecture](/docs/concepts/architecture)** — the components, the dual-port topology,
  how a request flows, and where the broker fits alongside your gateway and identity
  provider.
- **[Delegation and consent](/docs/concepts/delegation-and-consent)** — principals, agents,
  services, permission sets, grants, and sessions: the objects users and administrators
  actually work with.
- **[OAuth2 server modes](/docs/concepts/oauth2-server-modes)** — how the broker can *proxy*
  an existing authorization server, *issue* its own tokens locally, or run both at once.
- **[Token exchange](/docs/concepts/token-exchange)** — how an agent's token becomes the
  right third-party token at request time, using RFC 8693.
- **[Encryption at rest](/docs/concepts/encryption)** — how third-party tokens and secrets
  are sealed with envelope encryption bound to each service.
- **[Glossary](/docs/concepts/glossary)** — precise definitions of every term used across
  these docs.

## The one-minute model

Four ideas carry most of the weight:

1. **Delegation is per-user, per-agent, per-service, and scoped.** A *principal* (the user)
   creates a *grant* that lets an *agent* use specific *permission sets* on specific
   third-party services. At most one active grant exists per user and agent; it is revocable
   and can expire.

2. **The broker owns the third-party tokens.** When a user authorizes a third-party service,
   the resulting access and refresh tokens are stored **encrypted** as a *user session* —
   one per user and service. Agents never receive these tokens directly.

3. **Access is exchanged, not shared.** At request time, a gateway presents an agent's token
   and the target resource; the broker verifies the grant and returns the correct third-party
   token via [token exchange](/docs/concepts/token-exchange). The agent holds no provider
   credential.

4. **Authentication is delegated to the edge.** The broker does not log users in. A trusted
   reverse proxy authenticates the user and passes the principal in a header. The broker
   enforces consent and least privilege from there.

Everything else — server modes, encryption backends, the ExtProc sidecar, CIMD — is a way to
make those four ideas work in a real deployment.
