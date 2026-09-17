# ADR 036: Agentgateway Native Token Exchange

**Status**: Proposed
**Date**: 2026-09-17
**Feature**: 045-agentgateway-token-exchange

---

## Context

The current gateway integration sends MCP traffic through an `extProc` policy to the ExtProc token-exchange service. That route remains supported under ADR 011.

Agentgateway v1.5.0 also provides `backendAuth.oauthTokenExchange`. This native policy can send an RFC 8693 request directly to the existing Broker `POST /oauth2/token` endpoint. It authenticates the gateway with a fresh `privateKeyJwt` client assertion.

The direct assertion and the ExtProc assertion normally have different issuers. The direct policy uses the gateway client ID as the assertion issuer. The ExtProc path uses an upstream-issued access token. The Broker has one broker-wide `token_exchange.client_assertion` trust anchor, as proposed by ADR 029.

A direct route must not configure or contact ExtProc. Combining `extProc` and `backendAuth.oauthTokenExchange` on one route makes the integration ambiguous and is not supported.

The existing Broker endpoint, authorization rules, configuration contract, storage model, and schema already support this exchange. This feature adds no web user interface work.

## Decision

A gateway route selects exactly one token-exchange policy:

- **ExtProc path**: The route uses its existing `extProc` policy and the ExtProc service.
- **Direct native path**: The route uses `backendAuth.oauthTokenExchange` with the default RFC 8693 token-exchange grant and `privateKeyJwt` client authentication. It contains no `extProc` policy, ExtProc endpoint, or ExtProc service dependency.

The direct native path sends the inbound credential as `subject_token` and its protected resource as `resource`. It sends a fresh signed `client_assertion` to the existing Broker token endpoint. The Broker remains the authority for client-assertion validation, subject-token validation, privileged-client authorization, user delegation, and protected-resource authorization.

The direct policy uses a private signing key from a file or platform secret reference. It does not use shared-secret client authentication, including `clientSecretBasic` and `clientSecretPost`. On a successful exchange, agentgateway forwards only the exchanged credential. On a Broker rejection or exchange error, it fails closed and does not forward the request or inbound credential.

One Broker instance has exactly one client-assertion issuer and trust anchor. A direct route whose gateway assertion issuer differs from the ExtProc assertion issuer must use a different Broker instance. Both paths can use one Broker instance only when their assertions use the same trusted issuer.

The feature preserves the existing ExtProc routes, configuration, images, and behavior without modification. It reuses the existing Broker `POST /oauth2/token` API, `token_exchange` settings, protected-resource mappings, `UserGrant` delegations, and stored user sessions. It adds no Broker API, configuration schema, storage, database schema, migration, or production code changes. The scope is backend integration, operator artifacts, and backend E2E coverage only. It makes no change under `web/`.

## Consequences

### Positive

- Operators can select a direct gateway route without deploying or contacting ExtProc for that route.
- Existing ExtProc deployments remain supported and unchanged.
- The Broker retains all existing token-exchange validation and authorization boundaries.
- The direct route has no shared client secret. The gateway private key remains outside the route configuration.
- The direct integration reuses the existing Broker API and data model without a new configuration contract or persistence work.

### Negative

- Operators must select one token-exchange policy for each route. They cannot combine the policies on the same route.
- A Broker instance cannot trust client assertions from different issuers. Operators need a separate Broker instance when the direct gateway and ExtProc paths use different assertion issuers.
- Direct-path operators must manage the gateway signing key, matching public JWKS, issuer, audience, and protected-resource mapping as one trust contract.
- The direct path adds a second documented gateway deployment choice that operators must understand.

---

## References

- Feature specification: [045-agentgateway-token-exchange](../specs/045-agentgateway-token-exchange/spec.md)
- Implementation plan: [045-agentgateway-token-exchange](../specs/045-agentgateway-token-exchange/plan.md)
- Research: [R7 — Broker contract](../specs/045-agentgateway-token-exchange/research.md#r7--the-broker-side-of-the-contract-no-api-or-schema-change) and [R8 — one Broker trust anchor](../specs/045-agentgateway-token-exchange/research.md#r8--one-broker-instance-cannot-anchor-both-integration-paths)
- Direct-route contract: [agentgateway-direct-route.md](../specs/045-agentgateway-token-exchange/contracts/agentgateway-direct-route.md)
- Broker configuration contract: [broker-direct-path-config.md](../specs/045-agentgateway-token-exchange/contracts/broker-direct-path-config.md)
- [ADR 007: E2E Testing with Ginkgo and Production Bootstrap](007-e2e-testing-with-ginkgo.md)
- [ADR 011: ExtProc Token Exchange as Standalone Binary with Separate Configuration Schema](011-extproc-standalone-binary.md)
- [ADR 029: Token Exchange Client-Assertion Trust Anchor](029-token-exchange-client-assertion-trust-anchor.md)
- Constitution: `.specify/memory/constitution.md` — Principles I, II, V, and XIII
