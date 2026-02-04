# Consent API Endpoints

This document provides comprehensive API documentation for the consent management endpoints used by the consent frontend. These endpoints enable end-users to view, grant, update, and revoke OAuth2 delegations to AI agents.

## Table of Contents

1. [Authentication](#authentication)
2. [CSRF Protection](#csrf-protection)
3. [User Information](#user-information)
4. [Agent Delegations](#agent-delegations)
5. [Agent Details](#agent-details)
6. [User Grants Management](#user-grants-management)
7. [Error Handling](#error-handling)

## Authentication

All consent API endpoints (except public endpoints) require authentication via the `X-Principal` header. This header is set by a trusted reverse proxy after user authentication.

**Header:**
```
X-Principal: user@example.com
```

**Authentication Flow:**
1. User authenticates with reverse proxy (OAuth2, SAML, etc.)
2. Reverse proxy validates credentials and sets `X-Principal` header
3. Go backend extracts principal from header
4. Principal is used to identify user for all operations

**No Authentication Endpoints:**
- `GET /api/consent/agent/:agentId` - Public agent information

**Authentication Required:**
- `GET /api/me` - User information
- `GET /api/consent/agents` - List delegations
- `GET /api/consent/agent/:agentId/grants` - List grants
- `POST /api/consent/agent/:agentId/grants` - Create/update grant

## CSRF Protection

All mutating requests (POST, PUT, DELETE) require CSRF token protection.

**Token Delivery:**
1. Frontend makes initial GET request (e.g., `GET /api/consent/agents`)
2. Backend sets CSRF token in cookie: `csrf_token`
3. Frontend reads token from cookie
4. Frontend includes token in `X-CSRF-Token` header for mutating requests

**Headers:**
```
Cookie: csrf_token=base64-encoded-token
X-CSRF-Token: base64-encoded-token
```

**Token Properties:**
- **Name**: `csrf_token`
- **Length**: 32 bytes (base64-encoded)
- **TTL**: 24 hours
- **Storage**: In-memory on backend, keyed by session ID (principal)
- **SameSite**: Strict
- **HttpOnly**: false (JavaScript needs to read it)
- **Secure**: true (HTTPS only in production)

**Example:**
```bash
# 1. Get CSRF token (automatic from cookie)
curl -c cookies.txt http://localhost:8080/api/consent/agents \
  -H "X-Principal: user@example.com"

# 2. Use token in POST request
CSRF_TOKEN=$(grep csrf_token cookies.txt | awk '{print $7}')
curl -b cookies.txt http://localhost:8080/api/consent/agent/agent-123/grants \
  -H "X-Principal: user@example.com" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"delegated_oauth2_tokens": [...]}'
```

## User Information

### GET /api/me

Retrieves information about the currently authenticated user.

**Authentication:** Required

**Request:**
```http
GET /api/me HTTP/1.1
Host: localhost:8080
X-Principal: alice@example.com
```

**Response: 200 OK**
```json
{
  "data": {
    "principal": "alice@example.com",
    "displayName": "alice@example.com",
    "pictureUrl": null
  }
}
```

**Response Fields:**
- `principal` (string): Unique user identifier
- `displayName` (string): Display name for UI (currently same as principal)
- `pictureUrl` (string|null): User profile picture URL (future feature)

**Error Responses:**
- `401 Unauthorized`: No principal in context
  ```json
  {
    "error": "unauthorized",
    "message": "authentication required"
  }
  ```

**Example:**
```bash
curl http://localhost:8080/api/me \
  -H "X-Principal: alice@example.com"
```

## Agent Delegations

### GET /api/consent/agents

Retrieves all agents that the authenticated user has granted permissions to.

**Authentication:** Required

**Request:**
```http
GET /api/consent/agents HTTP/1.1
Host: localhost:8080
X-Principal: alice@example.com
```

**Response: 200 OK**
```json
{
  "data": [
    {
      "agentId": "550e8400-e29b-41d4-a716-446655440000",
      "clientId": "github-assistant",
      "displayName": "GitHub Assistant",
      "description": "AI assistant for GitHub operations",
      "governanceUrl": "https://example.com/governance",
      "userDocumentationUrl": "https://example.com/docs",
      "agentInterfaceUrl": "https://example.com/chat",
      "hasActiveGrants": true,
      "grantCount": 2,
      "grantExpiresAt": "2026-12-17T12:00:00Z"
    }
  ]
}
```

**Response Fields:**
- `agentId` (string): Unique agent identifier (UUID)
- `clientId` (string): Client ID for the agent
- `displayName` (string): Human-readable agent name
- `description` (string): Agent purpose and capabilities
- `governanceUrl` (string|null): URL to governance documentation
- `userDocumentationUrl` (string|null): URL to user documentation
- `agentInterfaceUrl` (string|null): URL to agent chat interface
- `hasActiveGrants` (boolean): Whether user has active grants
- `grantCount` (number): Number of OAuth2 services granted
- `grantExpiresAt` (string|null): ISO 8601 timestamp of earliest grant expiration

**Notes:**
- Returns empty array if no delegations exist (not an error)
- Only includes agents with active (non-expired) grants
- Results sorted by most recently updated

**Error Responses:**
- `401 Unauthorized`: No principal in context
- `500 Internal Server Error`: Service error

**Example:**
```bash
curl http://localhost:8080/api/consent/agents \
  -H "X-Principal: alice@example.com"
```

## Agent Details

### GET /api/consent/agent/:agentId

Retrieves detailed information about a specific agent and all available OAuth2 services.

**Authentication:** Not required (public endpoint)

**Path Parameters:**
- `agentId` (string): Agent identifier (UUID)

**Request:**
```http
GET /api/consent/agent/550e8400-e29b-41d4-a716-446655440000 HTTP/1.1
Host: localhost:8080
```

**Response: 200 OK**
```json
{
  "data": {
    "agent": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "clientId": "github-assistant",
      "displayName": "GitHub Assistant",
      "description": "AI assistant for GitHub operations",
      "governanceUrl": "https://example.com/governance",
      "userDocumentationUrl": "https://example.com/docs",
      "agentInterfaceUrl": "https://example.com/chat",
      "createdAt": "2025-12-17T12:00:00Z",
      "updatedAt": "2025-12-17T12:00:00Z"
    },
    "services": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440000",
        "displayName": "GitHub",
        "scopes": [
          {
            "scopeValue": "repo",
            "description": "Access to repositories"
          },
          {
            "scopeValue": "user",
            "description": "Access to user profile"
          }
        ]
      },
      {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "displayName": "Google",
        "scopes": [
          {
            "scopeValue": "email",
            "description": "Read email address"
          }
        ]
      }
    ]
  }
}
```

**Response Fields:**

**Agent Object:**
- `id` (string): Unique agent identifier (UUID)
- `clientId` (string): Client ID for OAuth2
- `displayName` (string): Human-readable name
- `description` (string): Agent purpose
- `governanceUrl` (string|null): Governance documentation URL
- `userDocumentationUrl` (string|null): User documentation URL
- `agentInterfaceUrl` (string|null): Agent interface URL
- `createdAt` (string): ISO 8601 creation timestamp
- `updatedAt` (string): ISO 8601 last update timestamp

**Service Object:**
- `id` (string): OAuth2 service identifier (UUID)
- `displayName` (string): Service name (e.g., "GitHub")
- `scopes` (array): Available OAuth2 scopes

**Scope Object:**
- `scopeValue` (string): OAuth2 scope string (e.g., "repo")
- `description` (string): Human-readable scope description

**Notes:**
- All registered OAuth2 services are available to all agents (FR-025)
- Client secrets are never exposed in responses (SR-003)
- Agent metadata helps users make informed consent decisions (FR-009)

**Error Responses:**
- `400 Bad Request`: Invalid agent ID format
  ```json
  {
    "error": "bad request",
    "message": "agent ID is required"
  }
  ```
- `404 Not Found`: Agent doesn't exist
  ```json
  {
    "error": "not found",
    "message": "agent not found"
  }
  ```
- `500 Internal Server Error`: Service error

**Example:**
```bash
curl http://localhost:8080/api/consent/agent/550e8400-e29b-41d4-a716-446655440000
```

## User Grants Management

### GET /api/consent/agent/:agentId/grants

Retrieves all active grants the authenticated user has granted to the specified agent.

**Authentication:** Required

**Path Parameters:**
- `agentId` (string): Agent identifier (UUID)

**Request:**
```http
GET /api/consent/agent/550e8400-e29b-41d4-a716-446655440000/grants HTTP/1.1
Host: localhost:8080
X-Principal: alice@example.com
```

**Response: 200 OK**
```json
{
  "data": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "principal": "alice@example.com",
      "agentId": "550e8400-e29b-41d4-a716-446655440000",
      "validUntil": "2026-12-17T12:00:00Z",
      "delegatedOauth2Tokens": [
        {
          "thirdpartyOauth2ServiceId": "660e8400-e29b-41d4-a716-446655440000",
          "scopes": ["repo", "user"]
        },
        {
          "thirdpartyOauth2ServiceId": "770e8400-e29b-41d4-a716-446655440000",
          "scopes": ["email"]
        }
      ],
      "createdAt": "2025-12-17T12:00:00Z",
      "updatedAt": "2025-12-17T12:00:00Z"
    }
  ]
}
```

**Response Fields:**
- `id` (string): Grant identifier (UUID)
- `principal` (string): User identifier
- `agentId` (string): Agent identifier
- `validUntil` (string|null): ISO 8601 expiration timestamp (null = indefinite)
- `delegatedOauth2Tokens` (array): Delegated service tokens
  - `thirdpartyOauth2ServiceId` (string): OAuth2 service ID
  - `scopes` (string[]): Granted scopes for this service
- `createdAt` (string): ISO 8601 creation timestamp
- `updatedAt` (string): ISO 8601 last update timestamp

**Notes:**
- Returns only active (non-expired) grants
- Returns empty array if no grants exist (not an error per FR-012)
- Each user can have at most one grant per agent (upsert semantics)

**Error Responses:**
- `400 Bad Request`: Invalid agent ID format
- `401 Unauthorized`: No principal in context
- `404 Not Found`: Agent doesn't exist
- `500 Internal Server Error`: Service error

**Example:**
```bash
curl http://localhost:8080/api/consent/agent/550e8400-e29b-41d4-a716-446655440000/grants \
  -H "X-Principal: alice@example.com"
```

### POST /api/consent/agent/:agentId/grants

Creates a new grant or updates an existing one (upsert semantics). Special case: empty `delegatedOauth2Tokens` array revokes the grant.

**Authentication:** Required

**CSRF Protection:** Required

**Path Parameters:**
- `agentId` (string): Agent identifier (UUID)

**Request Headers:**
```
X-Principal: alice@example.com
X-CSRF-Token: base64-encoded-token
Content-Type: application/json
```

**Request Body:**
```json
{
  "validUntil": "2026-12-17T12:00:00Z",
  "delegatedOauth2Tokens": [
    {
      "thirdpartyOauth2ServiceId": "660e8400-e29b-41d4-a716-446655440000",
      "scopes": ["repo", "user"]
    },
    {
      "thirdpartyOauth2ServiceId": "770e8400-e29b-41d4-a716-446655440000",
      "scopes": ["email"]
    }
  ]
}
```

**Request Fields:**
- `validUntil` (string|null, optional): ISO 8601 timestamp. Grant expires after this time. Omit for indefinite grant.
- `delegatedOauth2Tokens` (array, required): Service delegations. Empty array = revoke grant.
  - `thirdpartyOauth2ServiceId` (string): OAuth2 service ID
  - `scopes` (string[]): Scope values to delegate

**Response: 201 Created** (Grant created/updated)
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "principal": "alice@example.com",
  "agentId": "550e8400-e29b-41d4-a716-446655440000",
  "validUntil": "2026-12-17T12:00:00Z",
  "delegatedOauth2Tokens": [
    {
      "thirdpartyOauth2ServiceId": "660e8400-e29b-41d4-a716-446655440000",
      "scopes": ["repo", "user"]
    }
  ],
  "createdAt": "2025-12-17T12:00:00Z",
  "updatedAt": "2025-12-17T12:00:00Z"
}
```

**Response: 204 No Content** (Grant revoked - empty tokens)

**Business Logic:**

**Upsert Semantics (FR-013, FR-015):**
- If no grant exists for (principal, agent), creates new grant
- If grant exists, updates it with new token delegations and expiration
- Each user can have at most one grant per agent

**Scope Validation (FR-018):**
All requested scopes must exist in the service's scope configuration:
```json
// Valid: both scopes defined in service
{
  "thirdpartyOauth2ServiceId": "github",
  "scopes": ["repo", "user"]
}

// Invalid: "admin:org" not defined in service
{
  "thirdpartyOauth2ServiceId": "github",
  "scopes": ["repo", "admin:org"]
}
```

**Grant Expiration (FR-016, FR-019):**
- `validUntil` must be in the future if provided
- Expired grants are filtered out when listing grants
- Indefinite grants have `validUntil: null`

**Revocation (FR-014):**
Two ways to revoke:
1. POST with empty `delegatedOauth2Tokens` array
2. DELETE agent (cascade deletes all grants for that agent)

Revocation is idempotent - revoking a non-existent grant succeeds.

**Error Responses:**

**400 Bad Request** - Invalid request body or validation error
```json
{
  "error": "invalid request",
  "message": "valid_until must be in the future"
}
```

**400 Bad Request** - Invalid scopes
```json
{
  "error": "invalid scopes",
  "message": "delegation 0 (service=github) has invalid scopes: [admin:org]"
}
```

**400 Bad Request** - Service not found
```json
{
  "error": "service not found",
  "message": "service_id=invalid-service-id"
}
```

**401 Unauthorized** - No principal in context
```json
{
  "error": "unauthorized",
  "message": ""
}
```

**403 Forbidden** - Missing or invalid CSRF token
```json
{
  "error": "Forbidden",
  "message": "CSRF token required"
}
```

**404 Not Found** - Agent doesn't exist
```json
{
  "error": "agent not found",
  "message": ""
}
```

**500 Internal Server Error** - Service error

**Examples:**

**Grant Consent:**
```bash
curl -X POST http://localhost:8080/api/consent/agent/550e8400-e29b-41d4-a716-446655440000/grants \
  -H "X-Principal: alice@example.com" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "validUntil": "2026-12-17T12:00:00Z",
    "delegatedOauth2Tokens": [
      {
        "thirdpartyOauth2ServiceId": "660e8400-e29b-41d4-a716-446655440000",
        "scopes": ["repo", "user"]
      }
    ]
  }'
```

**Update Existing Grant:**
```bash
curl -X POST http://localhost:8080/api/consent/agent/550e8400-e29b-41d4-a716-446655440000/grants \
  -H "X-Principal: alice@example.com" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "delegatedOauth2Tokens": [
      {
        "thirdpartyOauth2ServiceId": "660e8400-e29b-41d4-a716-446655440000",
        "scopes": ["repo"]
      }
    ]
  }'
```

**Revoke Grant (Empty Tokens):**
```bash
curl -X POST http://localhost:8080/api/consent/agent/550e8400-e29b-41d4-a716-446655440000/grants \
  -H "X-Principal: alice@example.com" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "delegatedOauth2Tokens": []
  }'
```

## Error Handling

All error responses follow a consistent format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message"
}
```

### HTTP Status Codes

| Status | Meaning | When to Use |
|--------|---------|-------------|
| 200 OK | Success | Successful GET requests |
| 201 Created | Resource created | Successful grant creation/update |
| 204 No Content | Success, no body | Grant revoked successfully |
| 400 Bad Request | Invalid input | Validation errors, invalid JSON |
| 401 Unauthorized | Auth required | Missing or invalid principal |
| 403 Forbidden | CSRF failure | Missing or invalid CSRF token |
| 404 Not Found | Resource not found | Agent or service doesn't exist |
| 500 Internal Server Error | Server error | Database errors, service failures |

### Common Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `unauthorized` | 401 | No principal in context |
| `bad request` | 400 | Invalid request format or parameters |
| `invalid request` | 400 | Request validation failed |
| `invalid scopes` | 400 | Requested scopes not defined in service |
| `service not found` | 400 | OAuth2 service doesn't exist |
| `agent not found` | 404 | Agent doesn't exist |
| `not found` | 404 | Generic resource not found |
| `internal server error` | 500 | Unexpected server error |

### Error Handling Best Practices

**Frontend:**
1. Always check HTTP status code first
2. Parse error response body for details
3. Display user-friendly error messages
4. Retry transient errors (500, network errors) with exponential backoff
5. Don't retry client errors (400, 401, 403, 404)
6. Log errors for debugging

**Example Frontend Error Handler:**
```typescript
try {
  const response = await apiClient.post('/consent/agent/123/grants', data);
  return response.data;
} catch (error) {
  if (error.response) {
    const { status, data } = error.response;

    if (status === 401) {
      // Redirect to login
      window.location.href = '/login';
    } else if (status === 400 && data.error === 'invalid scopes') {
      // Show validation error
      showError(`Invalid scopes: ${data.message}`);
    } else if (status === 500) {
      // Retry or show generic error
      showError('Server error. Please try again.');
    }
  } else {
    // Network error
    showError('Network error. Check your connection.');
  }
}
```

## Complete Workflow Example

```bash
# 1. Admin creates agent (admin port 8081)
curl -X POST http://localhost:8081/api/agents \
  -H "X-Principal: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "my-assistant",
    "display_name": "My Assistant",
    "description": "Personal AI assistant"
  }'
# Response: {"id": "agent-123", ...}

# 2. Admin creates OAuth2 service
curl -X POST http://localhost:8081/api/third-party/oauth2/clients \
  -H "X-Principal: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "github-oauth",
    "client_secret": "secret",
    "display_name": "GitHub",
    "scopes": [
      {"scope_value": "repo", "description": "Repository access"}
    ]
  }'
# Response: {"id": "service-456", ...}

# 3. User views consent info (enduser port 8080)
curl http://localhost:8080/api/consent/agent/agent-123
# Response: {"data": {"agent": {...}, "services": [...]}}

# 4. User gets CSRF token (automatic from first GET)
curl -c cookies.txt http://localhost:8080/api/consent/agents \
  -H "X-Principal: alice@example.com"

# 5. User grants consent
CSRF_TOKEN=$(grep csrf_token cookies.txt | awk '{print $7}')
curl -b cookies.txt -X POST http://localhost:8080/api/consent/agent/agent-123/grants \
  -H "X-Principal: alice@example.com" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "delegatedOauth2Tokens": [
      {
        "thirdpartyOauth2ServiceId": "service-456",
        "scopes": ["repo"]
      }
    ]
  }'
# Response: 201 Created with grant details

# 6. User lists grants
curl http://localhost:8080/api/consent/agent/agent-123/grants \
  -H "X-Principal: alice@example.com"
# Response: [{"id": "grant-789", ...}]

# 7. User updates grant (add scope)
curl -b cookies.txt -X POST http://localhost:8080/api/consent/agent/agent-123/grants \
  -H "X-Principal: alice@example.com" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "delegatedOauth2Tokens": [
      {
        "thirdpartyOauth2ServiceId": "service-456",
        "scopes": ["repo", "user"]
      }
    ]
  }'

# 8. User revokes grant
curl -b cookies.txt -X POST http://localhost:8080/api/consent/agent/agent-123/grants \
  -H "X-Principal: alice@example.com" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"delegatedOauth2Tokens": []}'
# Response: 204 No Content
```

## Security Considerations

1. **Principal Authentication**: All endpoints (except GET agent info) require `X-Principal` header
2. **CSRF Protection**: All mutating requests require CSRF token validation
3. **Secret Redaction**: Client secrets are never exposed in responses (SR-003)
4. **Principal Isolation**: Users can only access their own grants (FR-017)
5. **Scope Validation**: Only defined scopes can be delegated (FR-018)
6. **Cascade Delete**: Deleting agent removes all grants automatically (FR-020)
7. **Service Protection**: Cannot delete service if grants reference it (409 Conflict)
8. **HTTPS Only**: All production traffic must use HTTPS
9. **SameSite Cookies**: CSRF cookies use SameSite=Strict
10. **Rate Limiting**: (Future) Implement rate limiting per principal

## Related Documentation

- [Admin APIs](/docs/api/admin-apis.md) - Agent and service management
- [Consent APIs](/docs/api/consent-apis.md) - Legacy consent documentation
- [Authentication](/docs/api/authentication.md) - Authentication patterns
- [Middleware](/docs/api/middleware.md) - CSRF and principal middleware
- [Error Codes](/docs/api/error-codes.md) - Complete error reference
