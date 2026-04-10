# RFC 8693 OAuth 2.0 Token Exchange

RFC 8693 Token Exchange enables privileged clients (API gateways and reverse proxies) to exchange tokens issued by the Upstream OAuth2 Server for third-party OAuth2 tokens stored in the identity broker's token vault. This enables agents to access third-party services on behalf of users with user-controlled consent.

## Overview

**Endpoint:** `POST /oauth2/token`

**Grant Type:** `urn:ietf:params:oauth:grant-type:token-exchange`

Token exchange integrates seamlessly with the existing `/oauth2/token` endpoint. The endpoint detects token exchange requests by examining the `grant_type` parameter. When `grant_type=urn:ietf:params:oauth:grant-type:token-exchange`, the system processes the request according to RFC 8693 Token Exchange. For other grant types (e.g., `authorization_code`), requests are proxied to the Upstream OAuth2 Server.

## Key Concepts

### Tokens

**Subject Token** (`subject_token`)
- JWT issued by the Upstream OAuth2 Server
- Contains user's principal (`sub` claim) and agent identifier (configurable claim, e.g., `azp`)
- Privileged client presents this token to request exchange

**Client Assertion** (`client_assertion`)
- JWT identifying the privileged client (API gateway or reverse proxy)
- Issued by the Upstream OAuth2 Server
- Privileged client authenticates using this JWT

**Third-Party Token**
- OAuth2 access token stored in the identity broker's token vault
- Issued by a third-party OAuth2 service (e.g., GitHub, Google)
- Returned to privileged client upon successful exchange

### Resource Parameter

The `resource` parameter identifies which third-party service's token should be returned. The system matches the resource parameter against `protected_resources` URIs configured on services:

```
resource=https://api.github.com
  ↓
System finds service with protected_resources containing "https://api.github.com"
  ↓
Returns that service's stored third-party token
```

### User Grant

Before tokens can be exchanged, the system verifies that the user (extracted from subject_token's `sub` claim) has explicitly granted the agent (extracted from subject_token's configurable claim) access to the target service. This enforces the consent model.

## Request Format

### POST /oauth2/token

Token exchange request with RFC 8693 parameters.

**Content-Type:** `application/x-www-form-urlencoded`

**Required Parameters:**

| Parameter | Value | Description |
|-----------|-------|-------------|
| `grant_type` | `urn:ietf:params:oauth:grant-type:token-exchange` | Indicates RFC 8693 token exchange |
| `subject_token` | JWT string | User's token issued by Upstream OAuth2 Server (contains user principal + agent ID) |
| `subject_token_type` | `urn:ietf:params:oauth:token-type:jwt` | Indicates subject_token is a JWT |
| `client_assertion` | JWT string | Gateway's authentication token (issued by Upstream OAuth2 Server) |
| `client_assertion_type` | `urn:ietf:params:oauth:client-assertion-type:jwt-bearer` | Indicates client_assertion is a JWT |
| `resource` | URI string | Target service resource URI (e.g., `https://api.github.com`) |

**Optional Parameters:**

| Parameter | Value | Description |
|-----------|-------|-------------|
| `scope` | Space-separated scopes | Requested scopes (currently unused, for future OAuth2 scope restrictions) |

### Request Example

```bash
curl -X POST http://localhost:8080/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:jwt" \
  -d "client_assertion=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d "client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer" \
  -d "resource=https://api.github.com"
```

## Response Format

### Success Response (200 OK)

RFC 8693 compliant response with exchanged token.

**Content-Type:** `application/json`

```json
{
  "access_token": "ya29.a0AfH6SMBx...",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

**Response Fields:**

| Field | Description |
|-------|-------------|
| `access_token` | The exchanged third-party OAuth2 access token |
| `issued_token_type` | Always `urn:ietf:params:oauth:token-type:access_token` |
| `token_type` | Token type from third-party service (typically `Bearer`) |
| `expires_in` | Seconds until token expires (from third-party service's response) |

### Success Example

```bash
# Gateway exchanges Upstream token for third-party GitHub token
$ curl -X POST http://localhost:8080/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=$UPSTREAM_TOKEN" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:jwt" \
  -d "client_assertion=$GATEWAY_JWT" \
  -d "client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer" \
  -d "resource=https://api.github.com"

# Response:
{
  "access_token": "ghp_1234567890abcdefghijklmnopqr...",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "token_type": "Bearer",
  "expires_in": 28800
}
```

### Error Responses

Token exchange errors follow RFC 8693 error format: `error=<code>` and optional `error_description=<description>`.

#### 400 Bad Request - Invalid Request

Returned when request is malformed or missing required parameters.

```json
{
  "error": "invalid_request",
  "error_description": "resource parameter is required"
}
```

**Common Causes:**
- Missing required parameter (subject_token, client_assertion, resource, etc.)
- Invalid subject_token format or signature
- Invalid client_assertion format or signature
- Invalid resource URI format
- subject_token missing required claim (e.g., 'sub')

#### 401 Unauthorized - Invalid Client

Returned when client_assertion (privileged client identity) cannot be verified.

```json
{
  "error": "invalid_client",
  "error_description": "client_assertion signature verification failed"
}
```

**Common Causes:**
- client_assertion signature invalid against Upstream OAuth2 JWKS
- client_assertion expired (exp claim in past)
- client_assertion audience does not include broker identifier
- Missing or invalid client_assertion

#### 400 Bad Request - Invalid Target

Returned when resource parameter doesn't match any configured service.

```json
{
  "error": "invalid_target",
  "error_description": "No service configured for the requested resource"
}
```

**Causes:**
- No service has the resource URI in its protected_resources
- Multiple services have the same resource URI (ambiguous configuration)

#### 403 Forbidden - Access Denied

Returned when user authorization checks fail.

```json
{
  "error": "access_denied",
  "error_description": "User has not granted this agent access to the requested service"
}
```

**Common Causes:**
- User has not granted the agent access to the target service (no UserGrant)
- User's grant has been revoked (revocation_timestamp present)
- User's grant has expired (valid_until in past)
- CEL authorization expression evaluated to false
- Gateway not authorized (CEL expression violation)

#### 400 Bad Request - Invalid Grant

Returned when user has no valid session with the target service.

```json
{
  "error": "invalid_grant",
  "error_description": "User has no active session with the requested service"
}
```

**Causes:**
- No UserSession exists for user+service
- Both access_token and refresh_token expired (cannot refresh)
- Token refresh attempt failed (third-party service rejected refresh_token)

#### 500 Internal Server Error

Returned for unexpected server errors.

```json
{
  "error": "server_error",
  "error_description": "CEL evaluation timeout"
}
```

**Common Causes:**
- CEL authorization expression timeout (>100ms)
- Database connection failure
- JWKS fetch timeout
- Unexpected error during token retrieval

## Configuration

Token exchange behavior is configured via the main application configuration file.

### Example Configuration

```yaml
upstream_oauth2:
  issuer: https://auth.example.com
  jwks_uri: https://auth.example.com/.well-known/jwks.json
  client_id: broker-client-id
  client_secret: secret123

token_exchange:
  claim_extraction:
    principal_expression: subject_token.sub
    agent_id_expression: subject_token.azp

  authorization:
    type: cel
    cel:
      expression: "client_assertion.iss == 'trusted-privileged-client' || client_assertion.aud.contains('broker')"

  refresh:
    enabled: true
```

### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `upstream_oauth2.issuer` | string | (required) | Upstream OAuth2 Server issuer URI (shared with other features) |
| `upstream_oauth2.jwks_uri` | string | (required) | JWKS endpoint for JWT validation (shared with other features) |
| `token_exchange.claim_extraction.principal_expression` | string | `subject_token.sub` | CEL expression to extract user principal from subject_token |
| `token_exchange.claim_extraction.agent_id_expression` | string | `subject_token.azp` | CEL expression to extract agent client ID from subject_token |
| `token_exchange.authorization.type` | string | `cel` | Authorization evaluation strategy (currently only `cel` supported) |
| `token_exchange.authorization.cel.expression` | string | `true` | CEL expression for privileged client authorization (evaluates against client_assertion claims) |
| `token_exchange.refresh.enabled` | boolean | `true` | Enable automatic refresh when third-party token expired |

### CEL Expression Context

When evaluating CEL expressions, the following variables are available:

**For Authorization Expression:**
```
client_assertion              // JWT claims from client_assertion as map
  .sub                        // Privileged client subject
  .iss                        // Privileged client issuer
  .aud                        // Privileged client audience (array)
  .exp                        // Privileged client token expiration
  .iat                        // Privileged client token issued at
  .<custom-claim>             // Any custom claims in JWT

request                       // Token exchange request context
  .resource                   // Resource URI from request
  .grant_type                 // Grant type (always token-exchange)
  .scope                      // Optional scopes from request
```

**For Claim Extraction Expressions:**
```
subject_token                 // JWT claims from subject_token as map
  .sub                        // User principal (standard)
  .azp                        // Authorized party (standard)
  .aud                        // Audience (array)
  .iss                        // Issuer
  .exp                        // Expiration
  .iat                        // Issued at
  .<custom-claim>             // Any custom claims in JWT
```

### Example Expressions

**Allow any privileged client (default):**
```yaml
token_exchange:
  authorization:
    cel:
      expression: "true"
```

**Only allow specific privileged client issuer:**
```yaml
token_exchange:
  authorization:
    cel:
      expression: "client_assertion.iss == 'api-gateway.example.com'"
```

**Allow multiple trusted privileged clients:**
```yaml
token_exchange:
  authorization:
    cel:
      expression: |
        client_assertion.iss in [
          'api-gateway.example.com',
          'reverse-proxy.example.com'
        ]
```

**Extract agent ID from custom claim:**
```yaml
token_exchange:
  claim_extraction:
    agent_id_expression: "subject_token.agent_id || subject_token.azp"
```

## Admin API - Configure Protected Resources

Administrators configure which resource URIs map to which third-party services using the admin API.

### PUT /api/services/{id}

Update service with protected_resources.

**Headers:**
```
X-Principal: admin@example.com
Content-Type: application/json
```

**Request Body:**
```json
{
  "display_name": "GitHub",
  "protected_resources": [
    "https://api.github.com",
    "https://github.com"
  ]
}
```

**Response:** `200 OK`
```json
{
  "id": "github-service-123",
  "client_id": "github-oauth-client",
  "display_name": "GitHub",
  "protected_resources": [
    "https://api.github.com",
    "https://github.com"
  ],
  "created_at": "2026-01-15T10:00:00Z",
  "updated_at": "2026-01-19T14:30:00Z"
}
```

### POST /api/services

Create new service with protected_resources.

**Headers:**
```
X-Principal: admin@example.com
Content-Type: application/json
```

**Request Body:**
```json
{
  "client_id": "new-oauth-service",
  "display_name": "New OAuth2 Service",
  "protected_resources": [
    "https://api.example.com"
  ]
}
```

**Response:** `201 Created`
```json
{
  "id": "new-service-456",
  "client_id": "new-oauth-service",
  "display_name": "New OAuth2 Service",
  "protected_resources": [
    "https://api.example.com"
  ],
  "created_at": "2026-01-19T14:35:00Z",
  "updated_at": "2026-01-19T14:35:00Z"
}
```

### GET /api/services/{id}

Retrieve service with protected_resources.

**Response:** `200 OK`
```json
{
  "id": "github-service-123",
  "client_id": "github-oauth-client",
  "display_name": "GitHub",
  "protected_resources": [
    "https://api.github.com",
    "https://github.com"
  ],
  "created_at": "2026-01-15T10:00:00Z",
  "updated_at": "2026-01-19T14:30:00Z"
}
```

### Validation Rules

**Protected Resources Validation:**

1. **Must be valid URIs**: Each resource must be a valid HTTP/HTTPS URI
   ```
   Valid:   "https://api.github.com", "https://api.github.com/"
   Invalid: "not-a-uri", "api.github.com" (missing scheme)
   ```

2. **Must have scheme**: HTTP or HTTPS required
   ```
   Valid:   "https://api.github.com"
   Invalid: "api.github.com"
   ```

3. **Trailing slashes normalized**: Both `https://api.github.com` and `https://api.github.com/` are treated identically
   ```
   Input:  ["https://api.github.com/"]
   Stored: ["https://api.github.com"]
   ```

4. **No duplicates across services**: Each resource URI can only be configured on one service
   ```
   Service A: ["https://api.github.com"]
   Service B: ["https://api.github.com"]  // ERROR: 409 Conflict
   ```

### Admin API Error Responses

#### 400 Bad Request - Invalid URI

When resource URI is invalid.

```json
{
  "error": "invalid_request",
  "error_description": "protected_resources[0]: must include a scheme (e.g., https://)"
}
```

#### 409 Conflict - Duplicate Resource

When resource URI already configured on another service.

```json
{
  "error": "conflict",
  "error_description": "resource https://api.github.com already configured on service github-service-123"
}
```

## Complete Workflow Example

### Scenario: Agent Needs GitHub Access

```bash
# Step 1: Admin creates GitHub service with protected resources
curl -X POST http://localhost:8081/api/services \
  -H "X-Principal: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "github-oauth",
    "display_name": "GitHub",
    "protected_resources": ["https://api.github.com"]
  }'

# Response: {"id": "github-svc-1", ...}

# Step 2: Admin creates agent
curl -X POST http://localhost:8081/api/agents \
  -H "X-Principal: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "github-assistant",
    "display_name": "GitHub Assistant"
  }'

# Response: {"id": "agent-1", ...}

# Step 3: User grants consent for agent+service
curl -X POST http://localhost:8080/api/consent/agent/agent-1/grants \
  -H "X-Principal: user@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "delegated_oauth2_tokens": [
      {
        "thirdparty_oauth2_service_id": "github-svc-1",
        "scopes": ["repo"]
      }
    ]
  }'

# Response: {"id": "grant-1", ...}

# Step 4: User authenticates with Upstream OAuth2 Server (via browser)
# This creates a UserSession with stored GitHub tokens

# Step 5: Gateway exchanges token for GitHub access
curl -X POST http://localhost:8080/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=$UPSTREAM_JWT_WITH_USER_SUB_AND_AGENT_AZP" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:jwt" \
  -d "client_assertion=$GATEWAY_JWT" \
  -d "client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer" \
  -d "resource=https://api.github.com"

# Response:
{
  "access_token": "ghp_16C7e42F292c6912E7710c838347Ae178B4a",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "token_type": "Bearer",
  "expires_in": 28800
}

# Step 6: Gateway uses GitHub token to call GitHub API on behalf of user
curl -H "Authorization: Bearer ghp_16C7e42F292c6912E7710c838347Ae178B4a" \
  https://api.github.com/user

# Response: {"login": "octocat", ...}
```

## Security Considerations

### Authentication & Authorization

1. **Client Assertion Required**: Gateway must provide valid JWT identifying itself
2. **Subject Token Required**: User's upstream token must be provided
3. **Signature Verification Mandatory**: Both JWTs signatures verified against Upstream JWKS
4. **User Grant Required**: User must explicitly grant agent access before exchange
5. **Clock Skew Tolerance**: 30-second tolerance for token expiration (configurable)

### Token Handling

1. **Token Values Never Logged**: Returned access_token never appears in logs
2. **Tokens Transmitted Over HTTPS Only**: Configure TLS termination appropriately
3. **Token Storage**: Third-party tokens stored securely in configured backend (PostgreSQL/memory)
4. **Token Refresh**: Automatic refresh only if refresh_token valid (not manually triggered)

### CEL Expression Safety

1. **Sandboxed Evaluation**: CEL expressions run in sandbox with no system access
2. **Timeout Protection**: CEL evaluation limited to 100ms (returns server_error on timeout)
3. **Expression Validation**: Expressions validated at startup (fail-fast on syntax errors)

### Audit Logging

All token exchange operations are audited:

```
- Token exchange success: principal, agent_client_id, service_id, resource
- Token exchange failure: error_code, error_description, available identifiers
- Token refresh success: principal, service_id
- Token refresh failure: error_code, error_description
```

Token values are NEVER included in logs.

## Troubleshooting

### Gateway receives "invalid_client"

**Cause**: client_assertion JWT invalid or expired

**Solution**:
1. Verify client_assertion JWT is validly signed by Upstream OAuth2 Server
2. Check JWT expiration time (exp claim must be in future)
3. Verify JWT audience includes broker identifier
4. Confirm Upstream JWKS endpoint is accessible

### Gateway receives "invalid_request" for subject_token

**Cause**: subject_token JWT invalid, expired, or missing required claims

**Solution**:
1. Verify subject_token JWT is validly signed by Upstream OAuth2 Server
2. Check JWT expiration time (exp claim must be in future)
3. Verify subject_token contains 'sub' claim (user principal)
4. Verify subject_token contains claim used for agent_client_id extraction (e.g., 'azp')
5. Check CEL claim extraction expressions in configuration

### Gateway receives "invalid_target"

**Cause**: Resource URI doesn't match any configured service, or multiple services match

**Solution**:
1. Verify resource URI matches exactly (case-sensitive) with protected_resources
2. Check for trailing slash differences: normalize to no trailing slash
3. Confirm service was created/updated with correct protected_resources
4. Verify no duplicate resource URIs across services
5. Check database/storage backend for service data

### Gateway receives "access_denied"

**Cause**: User hasn't granted agent access, grant expired, or CEL denied

**Solution**:
1. Verify user granted consent for agent+service combination
2. Check grant valid_until timestamp (must be in future)
3. Confirm grant not revoked (revocation_timestamp)
4. Check CEL authorization expression in configuration
5. Verify client_assertion claims match CEL expression

### Gateway receives "invalid_grant"

**Cause**: User has no session or all tokens expired

**Solution**:
1. User needs to authenticate with Upstream OAuth2 Server (consent flow) to create session
2. Verify user was directed through consent flow and granted service access
3. Check if session exists in storage backend (PostgreSQL/memory)
4. For expired tokens: user must re-authenticate to refresh third-party tokens

## Performance

Token exchange operations are designed to complete quickly:

- **Without refresh**: <500ms p95 (JWT validation + service lookup + token retrieval)
- **With refresh**: <2000ms p95 (includes third-party token refresh)
- **CEL evaluation**: <100ms per expression

Performance depends on:
- JWKS caching (20-minute default, avoids repeated HTTP calls)
- Database query efficiency (GIN index on protected_resources)
- Concurrent request handling (connection pooling)

## References

- [RFC 8693 - OAuth 2.0 Token Exchange](https://tools.ietf.org/html/rfc8693)
- [Upstream OAuth2 Configuration](./oauth2-integration-guide.md)
- [Consent APIs](./consent-apis.md)
- [Admin APIs](./admin-apis.md)
