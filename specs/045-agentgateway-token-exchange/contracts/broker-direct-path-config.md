# Contract: Broker configuration for the direct gateway path

**Feature**: `045-agentgateway-token-exchange`
**Interface kind**: operator configuration (`internal/ports/config.go` schema, YAML/env/Helm)

> **No configuration key is added, changed or removed by this feature.** Principle VII therefore
> requires no Helm chart change. This document records the existing keys the direct path depends on,
> the values they must take, and the one operational rule that follows from them.

---

## 1. Required Broker settings

```yaml
token_exchange:
  # MUST equal the gateway clientAuth.assertionAudience AND the aud of the inbound subject JWT.
  expected_audience: token-exchange-broker

  client_assertion:
    # MUST equal the gateway clientAuth.clientId (agentgateway emits it as assertion iss and sub).
    issuer_uri: https://gateway.example.com
    # MUST be HTTPS and MUST serve the public half of clientAuth.signingKey.
    jwks_uri: https://gateway.example.com/.well-known/jwks.json
    jwks_min_refresh: 15m
    jwks_max_refresh: 1h

  claim_extraction:
    principal_expression: "subject_token.sub"
    # Resolves the upstream client_id claim to the broker agent.id UUID.
    # Use "subject_token.<agent_id_claim_name>" only when multi_agent_client.enabled is true.
    agent_id_expression: "resolveAgentIdByClientId(subject_token.azp)"

  authorization:
    type: cel
    cel:
      # CEL context exposes client_assertion, subject_token and request.
      expression: 'client_assertion.iss == "https://gateway.example.com"'
      evaluation_timeout: 100ms
```

| Key | Existing | Contract for the direct path |
|---|---|---|
| `token_exchange.expected_audience` | yes (default `token-exchange-broker`) | Validated against **both** the client assertion `aud` and the subject token `aud` |
| `token_exchange.client_assertion.issuer_uri` | yes | MUST be set explicitly to the gateway client ID. When empty it falls back to the upstream issuer, which is the ExtProc arrangement, not this one |
| `token_exchange.client_assertion.jwks_uri` | yes | MUST be HTTPS unless `security.skip_thirdparty_https_validation` is enabled; the validator rejects `http://` otherwise (`internal/config/validator.go:655-688`) |
| `token_exchange.client_assertion.jwks_min_refresh` / `jwks_max_refresh` | yes | Defaults 15m / 1h are applied in `internal/app/builder.go:665-676` |
| `token_exchange.claim_extraction.principal_expression` | yes | Required; compiled at startup. Extracts the delegating user from the subject token |
| `token_exchange.claim_extraction.agent_id_expression` | yes | Required; MUST return the broker `agent.id` UUID. Use `resolveAgentIdByClientId(subject_token.azp)` when `multi_agent_client.enabled` is `false`, or the configured agent-ID claim when it is `true`. A bare `subject_token.azp` is an upstream client ID, not an agent ID, and fails resolution |
| `token_exchange.authorization.*` | yes | `cel` is the implemented type. The CEL context exposes `client_assertion`, `subject_token` and `request` (`internal/domain/tokenexchange/cel_evaluator.go:173-177, 351-355`); pin the privileged gateway with `client_assertion.iss` or `client_assertion.sub` |
| `security.skip_thirdparty_https_validation` | yes | MUST stay `false` for the direct path. It relaxes only **URL-scheme** validation, so it can admit an `http://` URL; it never relaxes certificate verification. Development stacks that point at an HTTP mock upstream may set it, but `client_assertion.jwks_uri` MUST remain an HTTPS URL in every environment |

TLS trust for the JWKS fetch uses the process certificate roots; there is no per-adapter CA
override. A deployment using a private PKI must install its CA in the Broker container's trust store,
for example by mounting the CA and setting `SSL_CERT_FILE`.

---

## 2. Protected resource

The gateway's `resources[0]` MUST match a protected-resource record of a Broker-managed
`ThirdpartyOAuth2Service` (ADR 030 normalized child records). An unmapped resource yields
`invalid_target`.

---

## 3. Route-selection rule (binding, operator-facing)

`token_exchange.client_assertion` is a **single, broker-wide** trust anchor
(`internal/ports/config.go:769-789`, `internal/app/builder.go:635-696`).

- The ExtProc path presents an upstream-issued access token as `client_assertion`, which validates
  against the upstream issuer.
- The direct path presents a gateway-self-signed assertion whose issuer is the gateway client ID.

These are different issuers. Consequently:

> **One Broker instance serves one client-assertion issuer.** Run a dedicated Broker instance for the
> direct path, or move every route of a given Broker instance to the same client-assertion issuer.
> Existing ExtProc routes are unaffected and keep their current configuration (FR-009).

This is also why the local Compose stack is left alone: its Broker anchors the ExtProc path, and
bolting a direct route onto the same instance would break it. See research R9 for why no second local
stack ships with this feature.

---

## 4. What this feature does not change

- `api/enduser/openapi.yaml` and `api/admin/openapi.yaml` — unchanged.
- `internal/ports/config.go` — unchanged.
- `charts/agentic-identity-broker/` — unchanged (no new configuration parameter).
- `migrations/` — unchanged (no schema change).
- `internal/extproc/` and every ExtProc route, image, or configuration — unchanged (FR-009).
- `docker-compose.yml`, `mocks/agentgateway/config.yaml`, `config.docker.yaml` and
  `config.extproc.docker.yaml` — unchanged (research R9).
