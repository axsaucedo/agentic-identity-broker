# OAuth2 Authorization Server Integration Guide

This guide provides step-by-step instructions for integrating with the OAuth2 Authorization Server.

## Table of Contents

1. [Overview](#overview)
2. [Authentication Flow](#authentication-flow)
3. [API Endpoints](#api-endpoints)
4. [Configuration](#configuration)
5. [Integration Examples](#integration-examples)
6. [Error Handling](#error-handling)
7. [Security Considerations](#security-considerations)
8. [Troubleshooting](#troubleshooting)

## Overview

The OAuth2 Authorization Server is a broker that handles OAuth2 authorization code flow for delegated access to third-party OAuth2 services. Key features:

- **RFC 6749 Compliant**: Implements OAuth2 authorization code flow
- **RFC 8414 Compliant**: Provides server metadata discovery endpoint
- **PKCE Support**: Secure authorization code flow for public clients
- **User Consent Management**: Tracks user grants for third-party services
- **Upstream Proxying**: Proxies token requests to upstream OAuth2 providers

## Authentication Flow

The OAuth2 Authorization Server uses the Authorization Code flow:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Authorization Code Flow                       │
└─────────────────────────────────────────────────────────────────┘

1. User clicks "Connect Service"
2. Client redirects to /oauth2/authorize with parameters:
   - client_id: Agent identifier
   - redirect_uri: Where to send the user after authorization
   - response_type: "code"
   - scope: Requested permissions

3. Authorization Server checks user's grants:
   a. If active grant exists → redirect to upstream provider
   b. If no grant → redirect to consent UI for user approval

4. User completes upstream authorization (or grants consent)

5. Upstream provider sends user back with authorization code

6. Client exchanges code for access token at /oauth2/token:
   - grant_type: "authorization_code"
   - code: Authorization code from step 5
   - client_id: Client identifier
   - client_secret: Client secret (for confidential clients)
   - redirect_uri: Must match original request

7. Authorization Server proxies request to upstream token endpoint

8. Client receives access token from authorization server
```

## API Endpoints

### 1. Authorization Endpoint

**Endpoint**: `GET /oauth2/authorize`

**Purpose**: Initiates the OAuth2 authorization code flow

**Request Parameters**:

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `client_id` | Yes | String | OAuth2 client/agent identifier |
| `redirect_uri` | Yes | URI | Where to send the user after authorization |
| `response_type` | Yes | String | Must be `code` |
| `scope` | No | String | Space-separated list of scopes |
| `state` | No | String | Opaque value for CSRF protection |
| `code_challenge` | No | String | PKCE code challenge (base64url-encoded SHA256) |
| `code_challenge_method` | No | String | PKCE method: `S256` (recommended) or `plain` |

**Response**:
- **302 Found**: Redirect to upstream authorize endpoint or consent UI
- **400 Bad Request**: Invalid parameters
- **403 Forbidden**: User not authenticated
- **500 Internal Error**: Server error

**Example Request**:

```bash
GET /oauth2/authorize?client_id=agent-1&redirect_uri=https://client.example.com/callback&response_type=code&scope=openid+profile&state=state-123 HTTP/1.1
Host: broker.example.com
X-Remote-User: user@example.com
```

**Example Error Response**:

```json
{
  "error": "invalid_client",
  "error_description": "Client not found"
}
```

### 2. Token Endpoint

**Endpoint**: `POST /oauth2/token`

**Purpose**: Exchange authorization code for access token

**Request Format**: `application/x-www-form-urlencoded`

**Request Parameters**:

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `grant_type` | Yes | String | `authorization_code` or `refresh_token` |
| `code` | Yes* | String | Authorization code (for authorization_code grant) |
| `redirect_uri` | Yes* | URI | Must match original authorize request |
| `client_id` | Yes | String | Client identifier |
| `client_secret` | Yes** | String | Client secret (for confidential clients) |
| `refresh_token` | Yes* | String | Refresh token (for refresh_token grant) |
| `scope` | No | String | Requested scopes (for refresh_token grant) |

\* Required for specific grant types
\*\* Required unless client type is public

**Response**:
- **200 OK**: Access token response
- **400 Bad Request**: Invalid parameters
- **401 Unauthorized**: Client authentication failed
- **500 Internal Error**: Server error

**Example Request** (Authorization Code):

```bash
POST /oauth2/token HTTP/1.1
Host: broker.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code=auth_code_123&redirect_uri=https://client.example.com/callback&client_id=client-1&client_secret=secret-xyz
```

**Example Response**:

```json
{
  "access_token": "slAV32hkKG",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refreshtoken123",
  "scope": "openid profile"
}
```

**Example Request** (Refresh Token):

```bash
POST /oauth2/token HTTP/1.1
Host: broker.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&refresh_token=refreshtoken123&client_id=client-1&client_secret=secret-xyz
```

### 3. Metadata Endpoint

**Endpoint**: `GET /.well-known/oauth-authorization-server`

**Purpose**: Discover authorization server capabilities and endpoints

**Request**: No parameters required

**Response**:
- **200 OK**: Server metadata (JSON)
- **500 Internal Error**: Server error

**Example Response**:

```json
{
  "issuer": "https://broker.example.com",
  "authorization_endpoint": "https://broker.example.com/oauth2/authorize",
  "token_endpoint": "https://broker.example.com/oauth2/token",
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code", "refresh_token"],
  "token_endpoint_auth_methods_supported": ["client_secret_basic", "client_secret_post"],
  "response_modes_supported": ["query", "fragment"],
  "scopes_supported": ["openid", "profile", "email"],
  "code_challenge_methods_supported": ["S256", "plain"]
}
```

## Configuration

### Server Setup

The OAuth2 Authorization Server requires configuration before startup:

```go
// Create server instance
server := http.NewServer("enduser",
    ports.ServerInstanceConfig{
        Port: 8000,
        Bind: "127.0.0.1",
        Authentication: ports.AuthenticationConfig{
            PrincipalHeader: "X-Remote-User",
        },
    },
    logger,
)

// Set required repositories
server.SetAgentRepository(agentRepo)        // Stores OAuth2 client configurations
server.SetServiceRepository(serviceRepo)    // Stores third-party service configs
server.SetGrantRepository(grantRepo)        // Stores user grants

// Configure OAuth2
oauth2Config := &oauth2.OAuth2Config{
    UpstreamAuthorizeEndpoint: "https://upstream.example.com/authorize",
    UpstreamTokenEndpoint:     "https://upstream.example.com/token",
    PublicBaseURL:             "https://broker.example.com",
    SupportedResponseTypes:    []string{"code"},
    SupportedGrantTypes:       []string{"authorization_code", "refresh_token"},
}
server.SetOAuth2Config(oauth2Config)

// Listen and serve
listener, _ := server.Listen()
server.Serve(context.Background(), listener)
```

### Environment Variables

Configure via environment or configuration file:

```bash
# Upstream OAuth2 Provider
OAUTH2_AUTHORIZE_ENDPOINT=https://provider.example.com/authorize
OAUTH2_TOKEN_ENDPOINT=https://provider.example.com/token

# Public URL (for metadata and redirects)
PUBLIC_BASE_URL=https://broker.example.com

# Authentication
PRINCIPAL_HEADER=X-Remote-User

# Server Binding
SERVER_PORT=8000
SERVER_BIND=127.0.0.1
```

## Integration Examples

### Example 1: React Client Integration

```javascript
// 1. Discovery: Get server metadata
async function discoverEndpoints() {
  const response = await fetch('https://broker.example.com/.well-known/oauth-authorization-server');
  return await response.json();
}

// 2. Authorization: Redirect user to authorize endpoint
function initiateAuthorization() {
  const metadata = await discoverEndpoints();
  const params = new URLSearchParams({
    client_id: 'my-client-id',
    redirect_uri: window.location.href,
    response_type: 'code',
    scope: 'openid profile email',
    state: generateRandomString(32), // CSRF protection
  });

  window.location.href = `${metadata.authorization_endpoint}?${params}`;
}

// 3. Callback: Handle redirect with authorization code
async function handleAuthorizationCallback(code, state) {
  // Verify state parameter
  if (state !== sessionStorage.getItem('oauth_state')) {
    throw new Error('State mismatch');
  }

  // Exchange code for token
  const response = await fetch('https://broker.example.com/oauth2/token', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code: code,
      client_id: 'my-client-id',
      client_secret: 'my-client-secret',
      redirect_uri: window.location.origin + window.location.pathname,
    }).toString(),
  });

  const tokenData = await response.json();

  // Store access token securely
  localStorage.setItem('access_token', tokenData.access_token);
  localStorage.setItem('token_type', tokenData.token_type);

  return tokenData;
}

// 4. Refresh: Get new access token
async function refreshAccessToken(refreshToken) {
  const response = await fetch('https://broker.example.com/oauth2/token', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({
      grant_type: 'refresh_token',
      refresh_token: refreshToken,
      client_id: 'my-client-id',
      client_secret: 'my-client-secret',
    }).toString(),
  });

  return await response.json();
}
```

### Example 2: cURL Integration

```bash
# 1. Get server metadata
curl -s https://broker.example.com/.well-known/oauth-authorization-server | jq .

# 2. Start authorization flow (interactive - opens browser)
# Copy this URL and open in browser:
# https://broker.example.com/oauth2/authorize?client_id=client-1&redirect_uri=https://localhost:3000/callback&response_type=code&scope=openid+profile&state=state-123

# 3. Exchange authorization code for token (after user approves)
curl -X POST https://broker.example.com/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=auth_code_123" \
  -d "redirect_uri=https://localhost:3000/callback" \
  -d "client_id=client-1" \
  -d "client_secret=secret-xyz"

# 4. Refresh access token
curl -X POST https://broker.example.com/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=refresh_token_abc" \
  -d "client_id=client-1" \
  -d "client_secret=secret-xyz"
```

### Example 3: Python Integration

```python
import requests
from urllib.parse import urlencode, parse_qs
import secrets

class OAuth2Client:
    def __init__(self, client_id, client_secret, broker_url, redirect_uri):
        self.client_id = client_id
        self.client_secret = client_secret
        self.broker_url = broker_url
        self.redirect_uri = redirect_uri
        self.session = requests.Session()

    def get_authorization_url(self, scope="openid profile"):
        """Generate authorization URL"""
        state = secrets.token_urlsafe(32)

        params = {
            'client_id': self.client_id,
            'redirect_uri': self.redirect_uri,
            'response_type': 'code',
            'scope': scope,
            'state': state,
        }

        auth_url = f"{self.broker_url}/oauth2/authorize?{urlencode(params)}"
        return auth_url, state

    def exchange_code_for_token(self, code):
        """Exchange authorization code for access token"""
        data = {
            'grant_type': 'authorization_code',
            'code': code,
            'client_id': self.client_id,
            'client_secret': self.client_secret,
            'redirect_uri': self.redirect_uri,
        }

        response = self.session.post(
            f"{self.broker_url}/oauth2/token",
            data=data,
            headers={'Content-Type': 'application/x-www-form-urlencoded'}
        )
        response.raise_for_status()
        return response.json()

    def refresh_token(self, refresh_token):
        """Refresh access token"""
        data = {
            'grant_type': 'refresh_token',
            'refresh_token': refresh_token,
            'client_id': self.client_id,
            'client_secret': self.client_secret,
        }

        response = self.session.post(
            f"{self.broker_url}/oauth2/token",
            data=data,
            headers={'Content-Type': 'application/x-www-form-urlencoded'}
        )
        response.raise_for_status()
        return response.json()

# Usage
client = OAuth2Client(
    client_id='my-client',
    client_secret='secret',
    broker_url='https://broker.example.com',
    redirect_uri='https://myapp.example.com/callback'
)

# Get authorization URL
auth_url, state = client.get_authorization_url()
print(f"Open this URL: {auth_url}")

# After user authorizes, exchange code
token_data = client.exchange_code_for_token(code='auth_code_123')
print(f"Access Token: {token_data['access_token']}")

# Refresh token
new_token = client.refresh_token(token_data['refresh_token'])
print(f"New Access Token: {new_token['access_token']}")
```

## Error Handling

### Common Error Responses

#### Invalid Client Error

```json
{
  "error": "invalid_client",
  "error_description": "Client not found or not registered"
}
```

**Solution**: Verify `client_id` matches a registered agent.

#### Invalid Grant Error

```json
{
  "error": "invalid_grant",
  "error_description": "Authorization code has expired or been revoked"
}
```

**Solution**: Re-initiate authorization flow to get a new code.

#### Unsupported Grant Type Error

```json
{
  "error": "unsupported_grant_type",
  "error_description": "Grant type not supported by server"
}
```

**Solution**: Use only supported grant types: `authorization_code` or `refresh_token`.

#### Invalid Request Error

```json
{
  "error": "invalid_request",
  "error_description": "Missing required parameter: redirect_uri"
}
```

**Solution**: Include all required parameters in request.

### Handling Errors in Code

```javascript
try {
  const response = await fetch('https://broker.example.com/oauth2/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: formData,
  });

  const data = await response.json();

  if (!response.ok) {
    // Handle OAuth2 error response
    console.error(`OAuth2 Error: ${data.error}`);
    console.error(`Description: ${data.error_description}`);

    // Handle specific errors
    switch (data.error) {
      case 'invalid_grant':
        // Redirect user to authorization
        initiateAuthorization();
        break;
      case 'invalid_client':
        // Check client configuration
        throw new Error('Client configuration invalid');
      default:
        throw new Error(data.error_description);
    }
  }

  return data;
} catch (error) {
  console.error('Token exchange failed:', error);
  throw error;
}
```

## Security Considerations

### 1. Use HTTPS Only

Always use HTTPS in production:
- All endpoints require HTTPS
- Redirect URIs must use HTTPS (except localhost)
- No sensitive data over HTTP

### 2. Implement PKCE

For public clients (SPAs, mobile apps), use PKCE:

```javascript
// Generate PKCE parameters
const codeVerifier = generateRandomString(128);
const codeChallenge = base64url(sha256(codeVerifier));

// In authorization request:
// code_challenge=<codeChallenge>
// code_challenge_method=S256

// In token request:
// code_verifier=<codeVerifier>
```

### 3. Validate State Parameter

Always validate state parameter to prevent CSRF:

```javascript
const state = generateRandomString(32);
sessionStorage.setItem('oauth_state', state);

// Include in authorization request
// state=<state>

// Validate in callback
if (callbackState !== sessionStorage.getItem('oauth_state')) {
  throw new Error('CSRF attack detected');
}
```

### 4. Secure Token Storage

- Store tokens in secure, HTTPOnly cookies (preferred)
- Never store in localStorage (vulnerable to XSS)
- Use secure session storage for temporary data

### 5. Client Secret Management

- Never expose client secret in frontend code
- Use backend servers to store secrets
- Rotate secrets regularly
- Use environment variables or secure vaults

### 6. Redirect URI Validation

- Register all valid redirect URIs upfront
- Exact match required (no wildcards)
- Validate redirect URIs in token requests

## Troubleshooting

### Issue: "Authorization endpoint not responding"

**Causes**:
- Upstream OAuth2 provider is down
- Network connectivity issue
- Invalid upstream endpoint configuration

**Solution**:
```bash
# Test connectivity
curl https://upstream.example.com/authorize -v

# Check configuration
echo $OAUTH2_AUTHORIZE_ENDPOINT
```

### Issue: "Access token invalid or expired"

**Causes**:
- Token has expired
- Token was revoked
- Token format incorrect

**Solution**:
- Refresh token using refresh token endpoint
- Re-initiate authorization flow
- Check token format (should be "Bearer <token>")

### Issue: "Redirect URI mismatch"

**Causes**:
- Redirect URI in callback doesn't match authorization request
- Case sensitivity issue
- URL encoding issue

**Solution**:
```javascript
// Use exact same redirect_uri in both requests
const redirectUri = window.location.origin + '/callback'; // consistent

// Authorization
// redirect_uri=https://myapp.example.com/callback

// Token exchange
// redirect_uri=https://myapp.example.com/callback (identical)
```

### Issue: "Cross-origin request blocked"

**Cause**: CORS policy preventing client-side requests

**Solution**: Token endpoint requests must be made from backend, not frontend.

```javascript
// ❌ WRONG - Frontend makes token request
fetch('https://broker.example.com/oauth2/token', {
  method: 'POST',
  body: data,
});

// ✅ CORRECT - Backend makes token request
app.post('/api/auth/callback', async (req, res) => {
  const token = await axios.post(
    'https://broker.example.com/oauth2/token',
    { grant_type: 'authorization_code', code: req.query.code, ... }
  );
  res.json(token);
});
```

## URL-Based Client IDs (CIMD)

When `oauth2_authorization_server.cimd.enabled: true`, agents may identify themselves using an HTTPS URL as `client_id`. The broker fetches and validates a Client ID Metadata Document from that URL before processing the authorization request.

### Authorization Request

```
GET /oauth2/authorize
  ?response_type=code
  &client_id=https%3A%2F%2Fagent.example.com%2F.well-known%2Fagent.json
  &redirect_uri=https%3A%2F%2Fagent.example.com%2Fcallback
  &scope=openid
  &code_challenge=...
  &code_challenge_method=S256
```

The `client_id` must be registered in the agent's `client_uris` array (see Admin APIs).

### CIMD Document Requirements

The broker fetches `https://agent.example.com/.well-known/agent.json` and requires:

1. `client_id` field in the document must exactly match the URL used as `client_id`
2. `redirect_uris` array must be non-empty; the requested `redirect_uri` must be same-origin with the document URL
3. `token_endpoint_auth_method` must not be a client-secret variant (`client_secret_basic`, `client_secret_post`, `client_secret_jwt`)
4. `client_name` must not match the operator `client_name_blocklist`

### Consent Screen

When the CIMD document is successfully fetched, the consent screen displays:
- Agent domain badge (the document's host)
- Localhost warning (if the URL is a loopback address)
- Advanced details section (auth method, redirect URIs, JWKS URI)

### Authorization Server Metadata

When CIMD is enabled, the `/.well-known/oauth-authorization-server` response includes:

```json
{
  "client_id_metadata_document_supported": true
}
```

### SSRF Protection

The broker blocks CIMD fetches to private, loopback, and link-local IP ranges (RFC 6890). SSRF protection is enforced at TCP-connect time after DNS resolution, preventing DNS rebinding attacks. This cannot be disabled.

## Support

For additional help:
- Check logs: `docker logs identity-broker`
- Enable debug logging: `LOG_LEVEL=debug`
- Review RFC 6749: https://tools.ietf.org/html/rfc6749
- Review RFC 8414: https://tools.ietf.org/html/rfc8414
