---
title: "Set up native token exchange at agentgateway"
description: Configure agentgateway v1.5.0 to exchange an agent token directly with the Broker by using RFC 8693 and privateKeyJwt.
---

# Set up native token exchange at agentgateway

This guide configures the direct replacement route for agentgateway `v1.5.0`.
The route uses `backendAuth.oauthTokenExchange` on an MCP backend.
It sends RFC 8693 requests directly to the Broker `POST /oauth2/token` endpoint.

This route is an alternative to the ExtProc route. Choose one integration path for each route.
Do not configure `extProc` on a direct route. The direct route does not contact or require `extproc-token-exchange`.

The direct route authenticates the gateway with a signed `privateKeyJwt` assertion.
It does not use a shared secret or a client-credentials grant.

## What you need

- Agentgateway `v1.5.0` in standalone-binary mode.
- A Broker token endpoint at `POST /oauth2/token`.
- An MCP target that the gateway can reach.
- A private signing key stored in a file that agentgateway can read.
- A matching public JWKS that the Broker can fetch over HTTPS.
- A Broker protected-resource mapping for the target `resource` URI.
- A Broker-configured upstream trust anchor for inbound subject JWTs.

## Direct route configuration

Copy [`examples/agentgateway/direct-token-exchange.yaml`](../../examples/agentgateway/direct-token-exchange.yaml).
Then replace the deployment values in this configuration:

```yaml
# yaml-language-server: $schema=https://agentgateway.dev/schema/config
binds:
- port: 4000
  listeners:
  - routes:
    - backends:
      - mcp:
          targets:
          - name: tools
            mcp:
              host: https://mcp.example.com/mcp
        policies:
          backendAuth:
            oauthTokenExchange:
              host: https://broker.example.com
              path: /oauth2/token
              resources:
              - https://api.example.com
              clientAuth:
                method: privateKeyJwt
                clientId: https://gateway.example.com
                assertionAudience: token-exchange-broker
                signingKey:
                  file: /etc/agentgateway/keys/gateway-signing.key
                alg: RS256
                kid: gateway-2026-09
```

Put `policies.backendAuth.oauthTokenExchange` beside `mcp` in the same `backends[]` item.
Do not attach it to the route, listener, or a different backend.

Use a scheme-bearing URI for `host`. Use an `https://` Broker URI in production.
Set `path` separately to `/oauth2/token`. Do not put the token path in `host`.

Set exactly one `resources` value. It must be an absolute URI without a fragment.
This URI must match one protected-resource mapping on a Broker-managed third-party service.
An unmapped URI returns `invalid_target`.

Use `privateKeyJwt` exactly as shown. Standalone YAML uses this camelCase spelling.
Set `signingKey.file` to the mounted private-key file. Do not put private-key PEM in the route configuration or logs.

Set `kid` when the HTTPS JWKS publishes more than one public key. Select an algorithm that matches the private key.
The listed values are `RS256`, `RS384`, `RS512`, `PS256`, `ES256`, and `ES384`.

Do not configure `clientSecretBasic`, `clientSecretPost`, or another shared-secret client-authentication method.
Do not configure `grantType: jwtBearer`, provider-specific on-behalf-of modes, or another non-RFC-8693 exchange mode.
Omit `grantType` for this route. Agentgateway then uses the RFC 8693 `tokenExchange` default.

## Configure the trust tuple

Replace these values as one validation contract. A value that matches only part of this table is not valid.

| Item | Gateway value | Broker value | Requirement |
|---|---|---|---|
| Gateway identity | `clientAuth.clientId` | `token_exchange.client_assertion.issuer_uri` | The values match. Agentgateway sets assertion `iss` and `sub` to this value. |
| Assertion audience | `clientAuth.assertionAudience` | `token_exchange.expected_audience` | The values match. This is the assertion `aud`, not the Broker token-endpoint URL. |
| Gateway key | `clientAuth.signingKey.file` | `token_exchange.client_assertion.jwks_uri` | The file holds the private key. The HTTPS JWKS serves its matching public key. |
| Subject JWT issuer | Not configured by the route | Existing Broker upstream issuer configuration | The inbound subject JWT `iss` matches the Broker upstream trust anchor. |
| Subject JWT audience | Inbound subject JWT `aud` | `token_exchange.expected_audience` | The values match the assertion audience. |

Configure the Broker with [`examples/config/token-exchange-direct-gateway.yaml`](../../examples/config/token-exchange-direct-gateway.yaml) as the reference.
Keep `security.skip_thirdparty_https_validation: false`. Set `token_exchange.client_assertion.jwks_uri` to an HTTPS URL.

The Broker validates the gateway assertion signature against this dedicated HTTPS JWKS.
It also validates the separate inbound subject JWT against its configured upstream trust anchor.
The public JWKS must contain the public key that matches `signingKey.file` and `kid`.
Only the gateway host stores the private key.

The Broker has one broker-wide client-assertion trust anchor. An ExtProc route presents an upstream-issued assertion.
A direct route presents a gateway-signed assertion. These assertions normally have different issuers.

Use a dedicated Broker instance for the direct route, or move every route of one Broker instance to one assertion issuer.
Existing ExtProc routes remain unchanged. Do not combine their policy with this direct route.

## What agentgateway sends

When it needs an exchange, agentgateway sends an `application/x-www-form-urlencoded` request to the Broker.
The default subject token type is `urn:ietf:params:oauth:token-type:access_token`.

| Form field | Value |
|---|---|
| `grant_type` | `urn:ietf:params:oauth:grant-type:token-exchange` |
| `subject_token` | The inbound agent credential |
| `subject_token_type` | `urn:ietf:params:oauth:token-type:access_token` |
| `resource` | The configured `resources[0]` URI |
| `client_id` | `clientAuth.clientId` |
| `client_assertion_type` | `urn:ietf:params:oauth:client-assertion-type:jwt-bearer` |
| `client_assertion` | A fresh signed assertion |

The assertion has `iss` and `sub` equal to `clientAuth.clientId`.
Its `aud` equals `clientAuth.assertionAudience`. Agentgateway creates a fresh `jti` for each request.

The Broker validates the subject token signature, issuer, audience, and time claims.
It validates the client assertion signature, issuer, audience, and time claims.
The Broker then applies CEL authorization, protected-resource authorization, delegation, and stored-session checks.

## Fail-closed results

Agentgateway forwards a request only after the Broker returns a non-empty `access_token` and `token_type: Bearer`.
It replaces the inbound credential before it forwards the request.
The protected backend receives exactly one `Authorization: Bearer <exchanged credential>` header.
It never receives the inbound credential in a header, query parameter, or request body.

If exchange fails, the MCP route returns HTTP 500 to the agent. It sends no request to the protected backend.
It does not forward the inbound credential as a fallback.

| Condition | Broker result | Agent result | Protected backend |
|---|---|---|---|
| Untrusted key, wrong assertion issuer, or wrong assertion audience | `invalid_client` (401) | 500 | No request |
| Invalid subject-token issuer or audience | `invalid_grant` (400) | 500 | No request |
| Missing resource | `invalid_request` (400) | 500 | No request |
| Unmapped resource | `invalid_target` (400) | 500 | No request |
| Delegation or CEL authorization denies exchange | `access_denied` (403) | 500 | No request |
| Client-assertion JWKS is unavailable, or the Broker is unavailable | `server_error` (500) or no Broker response | 500 | No request |

## Verify the route

1. Make sure that the direct route contains `backendAuth.oauthTokenExchange` and contains no `extProc` policy.
2. Send an MCP request with a subject JWT from the Broker-configured upstream issuer and audience.
3. Record the `Authorization` header at the MCP backend.
4. Make sure that this header is `Bearer <exchanged credential>` and does not equal the inbound bearer token.
5. Make sure that the ExtProc service receives no connection from this route.

If the Broker rejects the assertion, examine the client ID, assertion audience, HTTPS JWKS URL, key ID, and key pair together.
If the Broker rejects the subject JWT, examine its issuer and audience against the Broker upstream trust anchor.

## Related

- [ExtProc token-exchange gateway](/docs/guides/token-exchange-gateway) — the existing alternative for routes that use ExtProc.
- [Token exchange](/docs/concepts/token-exchange) — Broker authorization and protected-resource behavior.
- [Token exchange reference](/docs/reference/token-exchange) — RFC 8693 request and response fields.
