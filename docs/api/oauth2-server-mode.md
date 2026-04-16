# OAuth2 Server Mode API Reference

This document describes the API endpoints available when the broker operates in `issue_token` mode. For configuration details, see [OAuth2 Server Mode Feature Documentation](../features/oauth2-server-mode.md). For general configuration, see [Configuration Guide](../configuration.md).

## End-User Endpoints (Port 8000)

### OAuth2 Authorization Server Discovery

**Endpoint:** `GET /.well-known/oauth-authorization-server`

Returns RFC 8414 authorization server metadata including supported grant types, token endpoint, JWKS URI, and supported signing algorithms.

**Response:** `200 OK`
```json
{
  "issuer": "https://broker.example.com",
  "authorization_endpoint": "https://broker.example.com/oauth2/authorize",
  "token_endpoint": "https://broker.example.com/oauth2/token",
  "jwks_uri": "https://broker.example.com/oauth2/jwks.json",
  "grant_types_supported": ["client_credentials", "authorization_code"],
  "response_types_supported": ["code"],
  "code_challenge_methods_supported": ["S256"],
  "token_endpoint_auth_methods_supported": ["client_secret_post"]
}
```

### JWKS Endpoint

**Endpoint:** `GET /oauth2/jwks.json`

Returns the JSON Web Key Set containing the broker's public signing keys. Clients use these keys to verify JWT access tokens issued by the broker.

**Response:** `200 OK`
```json
{
  "keys": [
    {
      "kty": "EC",
      "crv": "P-256",
      "kid": "key-id-1",
      "use": "sig",
      "alg": "ES256",
      "x": "...",
      "y": "..."
    }
  ]
}
```

### Authorization Endpoint

**Endpoint:** `GET /oauth2/authorize`

Initiates an authorization code grant flow. The user is redirected to the consent UI to approve or deny the request.

**Query Parameters:**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `response_type` | Yes | Must be `code` |
| `client_id` | Yes | Agent UUID (e.g. `550e8400-e29b-41d4-a716-446655440000`) |
| `redirect_uri` | Yes | Registered callback URI |
| `code_challenge` | Yes | PKCE code challenge (S256) |
| `code_challenge_method` | Yes | Must be `S256` |
| `state` | Recommended | Opaque value for CSRF protection |
| `scope` | No | Space-delimited requested scopes |

**Success Response:** `302 Found` redirect to `redirect_uri` with `code` and `state` query parameters.

**Error Responses:**
- `400 Bad Request` — Missing or invalid parameters
- `401 Unauthorized` — Unknown client_id
- `403 Forbidden` — User denied consent

### Token Endpoint

**Endpoint:** `POST /oauth2/token`

Issues JWT access tokens. Supports `client_credentials` and `authorization_code` grant types.

**Content-Type:** `application/x-www-form-urlencoded`

#### Client Credentials Grant

| Parameter | Required | Description |
|-----------|----------|-------------|
| `grant_type` | Yes | `client_credentials` |
| `client_id` | Yes | Agent UUID (e.g. `550e8400-e29b-41d4-a716-446655440000`) |
| `client_secret` | Yes | Broker-issued client secret |
| `scope` | No | Space-delimited requested scopes |

#### Authorization Code Grant

| Parameter | Required | Description |
|-----------|----------|-------------|
| `grant_type` | Yes | `authorization_code` |
| `client_id` | Yes | Agent UUID (e.g. `550e8400-e29b-41d4-a716-446655440000`) |
| `client_secret` | Yes | Broker-issued client secret |
| `code` | Yes | Authorization code from authorize endpoint |
| `redirect_uri` | Yes | Must match the original authorization request |
| `code_verifier` | Yes | PKCE code verifier |

**Success Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJFUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "read write"
}
```

**Error Responses:**
- `400 Bad Request` — Invalid grant, missing parameters, or PKCE verification failure
- `401 Unauthorized` — Invalid client credentials
- `403 Forbidden` — Authorization code expired or already used

## Admin Endpoints (Port 14000)

### Client Credentials Management

#### Generate Client Credentials

**Endpoint:** `POST /api/agents/{agent_id}/client-credentials`

Generates a new `broker_client_id` and `client_secret` for the specified agent. The secret is returned once in plaintext and stored as an Argon2id hash. At runtime, agents use `agent.id` (UUID) as `client_id` — `broker_client_id` is internal metadata used for rotation safety.

**Response:** `201 Created`
```json
{
  "broker_client_id": "broker_a1b2c3d4",
  "client_secret": "bsec_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "created_at": "2025-01-15T10:00:00Z"
}
```

**Notes:**
- The `client_secret` is only returned at creation time. Store it securely.
- Generating new credentials revokes any previous credentials for the agent.
- Use `agent.id` (UUID) as `client_id` at the token and authorization endpoints, not `broker_client_id`.

#### Get Client Credentials Metadata

**Endpoint:** `GET /api/agents/{agent_id}/client-credentials`

Returns metadata about the agent's broker-issued credentials (without the secret).

**Response:** `200 OK`
```json
{
  "broker_client_id": "broker_a1b2c3d4",
  "created_at": "2025-01-15T10:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` — Agent has no broker-issued credentials

#### Revoke Client Credentials

**Endpoint:** `DELETE /api/agents/{agent_id}/client-credentials`

Revokes the agent's broker-issued credentials. Tokens already issued remain valid until expiry.

**Response:** `204 No Content`

### Signing Key Management

#### Add Signing Key

**Endpoint:** `POST /api/oauth2-server/signing-keys`

Generates a new asymmetric signing key pair. The new key becomes the current signing key.

**Request Body:**
```json
{
  "algorithm": "ES256"
}
```

**Response:** `201 Created`
```json
{
  "kid": "key-2025-01-15-abc123",
  "algorithm": "ES256",
  "is_current": true,
  "created_at": "2025-01-15T10:00:00Z"
}
```

#### List Signing Keys

**Endpoint:** `GET /api/oauth2-server/signing-keys`

Returns all active signing keys with metadata.

**Response:** `200 OK`
```json
{
  "keys": [
    {
      "kid": "key-2025-01-15-abc123",
      "algorithm": "ES256",
      "is_current": true,
      "created_at": "2025-01-15T10:00:00Z"
    }
  ]
}
```

#### Promote Signing Key

**Endpoint:** `PUT /api/oauth2-server/signing-keys/{kid}/current`

Promotes the specified key to the current signing key. Previously current key remains active for verification.

**Response:** `200 OK`

**Error Responses:**
- `404 Not Found` — Key ID not found

#### Remove Signing Key

**Endpoint:** `DELETE /api/oauth2-server/signing-keys/{kid}`

Removes a signing key. Cannot remove the current signing key — promote another key first.

**Response:** `204 No Content`

**Error Responses:**
- `404 Not Found` — Key ID not found
- `409 Conflict` — Cannot delete the current signing key
