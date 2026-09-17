# Contract: Agentgateway Direct Route (`backendAuth.oauthTokenExchange`)

**Feature**: `045-agentgateway-token-exchange`
**Applies to**: agentgateway `v1.5.0` (`cr.agentgateway.dev/agentgateway:v1.5.0`)
**Interface kind**: configuration document consumed by an external component (agentgateway)

This is the interface the Broker exposes to gateway operators for the direct integration path. It is
a configuration contract, not an HTTP API: the Broker's HTTP API is unchanged
(see [broker-token-exchange-request.md](./broker-token-exchange-request.md)).

---

## 1. Standalone binary — reference route

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
              host: https://broker.example.com      # Scheme-bearing host enables backend TLS
              path: /oauth2/token
              resources:
              - https://api.example.com             # MUST match a Broker protected-resource mapping
              clientAuth:
                method: privateKeyJwt               # camelCase is mandatory in the standalone binary
                clientId: https://gateway.example.com
                assertionAudience: token-exchange-broker
                signingKey:
                  file: /etc/agentgateway/keys/gateway-signing.key
                alg: RS256
                kid: gateway-2026-09

**Placeholders an operator MUST replace as one unit** (FR-018):
`host`, `resources[0]`, `clientAuth.clientId`, `clientAuth.assertionAudience`,
`clientAuth.signingKey.file`, `clientAuth.kid`, and the matching Broker settings
`token_exchange.client_assertion.issuer_uri`, `token_exchange.client_assertion.jwks_uri`,
`token_exchange.expected_audience`.

**Attachment point**: `routes[].backends[].policies.backendAuth`. The direct route configures one
backend-auth method for its MCP backend.
---

## 2. Kubernetes — equivalent policy

Shape verified against the v1.5.0 controller translator fixtures
(`controller/pkg/agentgateway/translator/testdata/backends/oauth-private-key-jwt.yaml` and
`oauth-token-exchange-full.yaml`). The token endpoint is a backend reference, and the private-key
settings live in a nested `clientAuth.privateKeyJwt` block.

```yaml
apiVersion: agentgateway.dev/v1alpha1
kind: AgentgatewayBackend
metadata:
  namespace: default
  name: identity-broker-token-endpoint
spec:
  static:
    host: broker.example.com
    port: 443
---
apiVersion: v1
kind: Secret
metadata:
  namespace: default
  name: agentgateway-signing-key
stringData:
  signingKey: |
    -----BEGIN PRIVATE KEY-----
    ...
    -----END PRIVATE KEY-----
---
apiVersion: agentgateway.dev/v1alpha1
kind: AgentgatewayBackend
metadata:
  namespace: default
  name: mcp-backend
spec:
  static:
    host: mcp.example.com
    port: 443
  policies:
    auth:
      oauthTokenExchange:
        backendRef:
          group: agentgateway.dev
          kind: AgentgatewayBackend
          name: identity-broker-token-endpoint
        path: /oauth2/token
        grantType: TokenExchange          # PascalCase is mandatory in the CRDs
        resources:
        - https://api.example.com
        clientAuth:
          clientId: https://gateway.example.com
          method: PrivateKeyJwt
          privateKeyJwt:
            signingKeyRef:
              name: agentgateway-signing-key
            alg: RS256
            kid: gateway-2026-09
            assertionAudience: token-exchange-broker
```

The capitalization difference is enforced by both modes: the standalone binary rejects
`PrivateKeyJwt`, the CRDs reject `privateKeyJwt`.

---

## 3. Field contract

The table names standalone fields first; the Kubernetes equivalent follows in parentheses where the
shape differs.

| Field | Required | Contract |
|---|---|---|
| `host` (`backendRef` → an `AgentgatewayBackend` with `static.host`/`static.port`) | yes | Broker end-user server. Port `443` implicitly enables backend TLS |
| `path` | yes | `/oauth2/token` |
| `grantType` | no | MUST be absent, `tokenExchange` (standalone) or `TokenExchange` (Kubernetes). `jwtBearer` is not supported by the Broker (FR-016) |
| `resources` | yes | Exactly one absolute URI without a fragment, matching a Broker protected-resource mapping |
| `clientAuth.method` | yes | MUST be `privateKeyJwt` (standalone) or `PrivateKeyJwt` (Kubernetes) (FR-004) |
| `clientAuth.clientId` | yes | Becomes the assertion `iss` **and** `sub`; MUST equal the Broker `token_exchange.client_assertion.issuer_uri` |
| `clientAuth.assertionAudience` (`clientAuth.privateKeyJwt.assertionAudience`) | yes | Becomes the assertion `aud`; MUST equal the Broker `token_exchange.expected_audience`. It is **not** the Broker token-endpoint URL |
| `clientAuth.signingKey` (`clientAuth.privateKeyJwt.signingKeyRef`) | yes | PEM private key; use `{file: …}` in standalone YAML and a Secret reference on Kubernetes. MUST NOT be inlined in route configuration (FR-011) |
| `clientAuth.alg` (`clientAuth.privateKeyJwt.alg`) | no | `RS256` (default), `RS384`, `RS512`, `PS256`, `ES256`, `ES384`; MUST match the key family |
| `clientAuth.kid` (`clientAuth.privateKeyJwt.kid`) | no | Set when the published JWKS holds more than one key |
| `subjectToken` | no | Default (`Authorization: Bearer`, `…token-type:access_token`) is correct for the Broker |
| `authorizationLocation` (`location`) | no | Default (`Authorization: Bearer`) is correct for the backend |
| `cache` | no | Production: keep the default. Tests: `{maxEntries: 0}` for determinism |
| `audiences`, `scopes` | no | Omit. The Broker ignores `audience` on this path and derives scope from the stored session |
| `actorToken`, `additionalParams`, `requestedTokenType`, `chainedExchange` | no | Omit — out of scope |
| `extProc` (route policy) | — | MUST be absent on a direct route (FR-008) |
| `clientSecret` / `secretRef`, `clientSecretBasic`, `clientSecretPost` | — | MUST NOT be used with the Broker (FR-004, FR-011) |

---

## 4. Behavioural contract

1. The policy reads the inbound credential from `Authorization: Bearer` and sends it as
   `subject_token`. A route-level `jwtAuth`/MCP-auth policy that strips the same header breaks the
   exchange; use `subjectToken.source.expression: jwt.rawToken.unredacted()` or `preserveToken: true`
   when combining them.
2. On success, agentgateway removes the inbound credential and writes the exchanged credential to
   `authorizationLocation`. The protected backend never sees the inbound credential (FR-005).
3. On any Broker rejection, network failure, empty `access_token`, or non-`Bearer` `token_type`,
   agentgateway fails the request. It never falls back to forwarding the inbound credential (FR-007).
4. A policy sets at most one backend-auth method; two methods are rejected at configuration load.

---

## 5. Route selection rule (operator-visible)

- A route uses **either** `extProc` **or** `backendAuth.oauthTokenExchange`, never both.
- Existing ExtProc routes are unchanged and remain supported (FR-009).
- One Broker instance has exactly one client-assertion trust anchor. An ExtProc route presenting an
  upstream-issued access token and a direct route presenting a gateway-signed assertion therefore
  cannot authenticate against the **same** Broker instance unless both assertions come from the same
  issuer. Plan one Broker instance per client-assertion anchor.
