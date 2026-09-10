---
title: Concepts
description: The mental model behind the Agentic Identity Broker — how delegation, consent, server modes, token exchange, and encryption fit together.
---

# Concepts

This section explains how the broker works. It also explains why the broker has this
design. It is for readers who decide whether the broker fits their architecture. It also
serves operators who need a correct model before deployment.

The [delegation and consent model](/docs/concepts/delegation-and-consent) is the core
model. The remaining pages address specific parts of that model.

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

Four facts explain the broker:

1. **Delegation is per-user, per-agent, per-service, and scoped.** A *principal* creates a
   *grant*. The grant lets an *agent* use specific *permission sets* with specific
   third-party services. A user has at most one active grant for each agent. A grant can
   expire and can be revoked.

2. **The broker owns the third-party tokens.** When a user authorizes a third-party service,
   the broker creates a *user session*. The session stores encrypted access and refresh
   tokens. An agent does not receive these tokens.

3. **Access is exchanged, not shared.** At request time, a gateway presents an agent token
   and the target resource. The broker validates the grant and returns the appropriate
   third-party token through [token exchange](/docs/concepts/token-exchange). An agent does
   not hold a provider credential.

4. **Authentication is delegated to the edge.** The broker does not log users in. A trusted
   reverse proxy authenticates each user. It passes the principal in a header. The broker
   then enforces consent and least privilege.

Server modes, encryption backends, the ExtProc sidecar, and CIMD support these facts in a
deployment.
