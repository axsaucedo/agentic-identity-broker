# OAuth2 Server Mode

## Overview

The identity broker can operate as a standalone OAuth2 authorization server using `issue_token` mode. In this mode, the broker mints its own JWT access tokens signed with managed asymmetric keys, supports both `client_credentials` and `authorization_code` (with PKCE) grant types, and exposes RFC 8414 discovery and JWKS endpoints.

## Configuration

Set `mode: issue_token` in the OAuth2 authorization server configuration:

```yaml
oauth2:
  auth_server:
    mode: "issue_token"
    issuer_uri: "https://broker.example.com"
    token_ttl: "1h"
    token_claims_expression: '{"team": agent.display_name}'
```

### Configuration Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `mode` | No | `proxy` | Operating mode: `proxy` or `issue_token` |
| `issuer_uri` | Yes (issue_token) | — | Issuer identifier for JWTs and discovery |
| `token_ttl` | No | `1h` | Access token validity period |
| `token_claims_expression` | No | `""` | CEL expression for custom JWT claims |

## Admin API

### Client Credentials Management

Generate broker-issued credentials for an agent:
```bash
# Generate credentials
curl -X POST http://localhost:14000/api/agents/{agent_id}/client-credentials

# Get credential metadata
curl http://localhost:14000/api/agents/{agent_id}/client-credentials

# Revoke credentials
curl -X DELETE http://localhost:14000/api/agents/{agent_id}/client-credentials
```

### Signing Key Management

```bash
# Add a new signing key (becomes current)
curl -X POST http://localhost:14000/api/oauth2-server/signing-keys \
  -H "Content-Type: application/json" \
  -d '{"algorithm": "ES256"}'

# List active signing keys
curl http://localhost:14000/api/oauth2-server/signing-keys

# Promote a key to current
curl -X PUT http://localhost:14000/api/oauth2-server/signing-keys/{kid}/current

# Remove a signing key
curl -X DELETE http://localhost:14000/api/oauth2-server/signing-keys/{kid}
```

## OAuth2 Endpoints (End-User)

### Discovery
```
GET /.well-known/oauth-authorization-server
```

### JWKS
```
GET /oauth2/jwks.json
```

### Token Endpoint

**Client Credentials Grant:**
```bash
curl -X POST http://localhost:8000/oauth2/token \
  -d "grant_type=client_credentials" \
  -d "client_id=broker_xxxx" \
  -d "client_secret=yyyy" \
  -d "scope=read write"
```

**Authorization Code Grant (with PKCE):**
```bash
# 1. Authorization request (browser redirect)
GET /oauth2/authorize?response_type=code&client_id=broker_xxxx&redirect_uri=http://app/callback&code_challenge=xxxx&code_challenge_method=S256&state=random

# 2. Exchange code for token
curl -X POST http://localhost:8000/oauth2/token \
  -d "grant_type=authorization_code" \
  -d "client_id=broker_xxxx" \
  -d "client_secret=yyyy" \
  -d "code=zzzz" \
  -d "redirect_uri=http://app/callback" \
  -d "code_verifier=original_verifier"
```

## Security

- **PKCE**: Always enforced (S256 only). No plaintext challenge method.
- **Client secrets**: Hashed with Argon2id. Never stored in plaintext.
- **Signing keys**: Private material encrypted at rest via EncryptionPort.
- **Authorization codes**: Single-use, 60-second TTL, stored as SHA-256 hash.
- **Fail-closed**: Signing key decryption failure prevents token issuance.

## Custom Token Claims

Use CEL expressions to add custom claims to issued tokens:

```yaml
token_claims_expression: '{"department": "engineering", "env": "production"}'
```

Available variables in CEL:
- `agent` — the Agent struct (display_name, description, etc.)
- `principal` — the authenticated user identifier (string)
- `request` — request context (map)

Base claims (`iss`, `sub`, `exp`, `iat`, `jti`, `kid`, `agent_id`, `scope`) cannot be overridden.
