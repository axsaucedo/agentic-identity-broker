# Consent APIs

Consent APIs allow end-users to view agent information, grant permissions, and manage their OAuth2 token delegations. These APIs are exposed on the enduser server port.

## Agent Consent Information

### Get Agent Consent Info

Retrieves agent metadata and all available third-party OAuth2 services. This endpoint provides information needed for users to make informed consent decisions.

**Endpoint:** `GET /api/consent/agent/:agent-id`

**Request Headers:**
- `X-Principal: <principal-identifier>` (optional for this endpoint)

**Response:** `200 OK`
```json
{
  "agent": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "client_id": "github-assistant",
    "display_name": "GitHub Assistant",
    "description": "AI assistant for GitHub operations",
    "governance_url": "https://example.com/governance",
    "user_documentation_url": "https://example.com/docs",
    "agent_interface_url": "https://example.com/chat",
    "created_at": "2025-12-17T12:00:00Z",
    "updated_at": "2025-12-17T12:00:00Z"
  },
  "requested_services": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "display_name": "GitHub",
      "scopes": [
        {
          "scope_value": "repo",
          "description": "Access to repositories"
        },
        {
          "scope_value": "user",
          "description": "Access to user profile"
        }
      ]
    },
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "display_name": "Google",
      "scopes": [
        {
          "scope_value": "email",
          "description": "Read email address"
        }
      ]
    }
  ]
}
```

**Notes:**
- All registered OAuth2 services are available to all agents (FR-025)
- Client secrets are never exposed in responses (SR-003)
- Agent metadata helps users understand what they're granting access to (FR-009)

**Example:**
```bash
curl http://localhost:8080/api/consent/agent/550e8400-e29b-41d4-a716-446655440000
```

**Error Responses:**
- `404 Not Found` - Agent doesn't exist
- `500 Internal Server Error` - Service error

## User Grant Management

### List User Grants

Retrieves all active grants for the authenticated user and specified agent.

**Endpoint:** `GET /api/consent/agent/:agent-id/grants`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Response:** `200 OK`
```json
[
  {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "principal": "user@example.com",
    "agent_id": "550e8400-e29b-41d4-a716-446655440000",
    "valid_until": "2026-12-17T12:00:00Z",
    "delegated_oauth2_tokens": [
      {
        "thirdparty_oauth2_service_id": "660e8400-e29b-41d4-a716-446655440000",
        "scopes": ["repo", "user"]
      }
    ],
    "created_at": "2025-12-17T12:00:00Z",
    "updated_at": "2025-12-17T12:00:00Z"
  }
]
```

**Notes:**
- Returns only active grants (not expired)
- Returns empty array if no grants exist (not an error per FR-012)
- Each user can have at most one grant per agent (upsert semantics)

**Example:**
```bash
curl http://localhost:8080/api/consent/agent/550e8400-e29b-41d4-a716-446655440000/grants \
  -H "X-Principal: user@example.com"
```

**Error Responses:**
- `401 Unauthorized` - No principal in context
- `404 Not Found` - Agent doesn't exist
- `500 Internal Server Error` - Service error

### Create or Update Grant

Creates a new grant or updates an existing one (upsert semantics). Special case: empty `delegated_oauth2_tokens` array revokes the grant.

**Endpoint:** `POST /api/consent/agent/:agent-id/grants`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Request Body:**
```json
{
  "valid_until": "2026-12-17T12:00:00Z",
  "delegated_oauth2_tokens": [
    {
      "thirdparty_oauth2_service_id": "660e8400-e29b-41d4-a716-446655440000",
      "scopes": ["repo", "user"]
    },
    {
      "thirdparty_oauth2_service_id": "770e8400-e29b-41d4-a716-446655440000",
      "scopes": ["email"]
    }
  ]
}
```

**Field Details:**
- `valid_until` (optional): ISO 8601 timestamp. Grant expires after this time. Omit for indefinite grant.
- `delegated_oauth2_tokens`: Array of service delegations. Empty array = revoke grant.
- `thirdparty_oauth2_service_id`: ID of the OAuth2 service
- `scopes`: Array of scope values to delegate

**Response:** `201 Created`
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "principal": "user@example.com",
  "agent_id": "550e8400-e29b-41d4-a716-446655440000",
  "valid_until": "2026-12-17T12:00:00Z",
  "delegated_oauth2_tokens": [
    {
      "thirdparty_oauth2_service_id": "660e8400-e29b-41d4-a716-446655440000",
      "scopes": ["repo", "user"]
    },
    {
      "thirdparty_oauth2_service_id": "770e8400-e29b-41d4-a716-446655440000",
      "scopes": ["email"]
    }
  ],
  "created_at": "2025-12-17T12:00:00Z",
  "updated_at": "2025-12-17T12:00:00Z"
}
```

**Example: Grant Consent**
```bash
curl -X POST http://localhost:8080/api/consent/agent/550e8400-e29b-41d4-a716-446655440000/grants \
  -H "X-Principal: user@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "valid_until": "2026-12-17T12:00:00Z",
    "delegated_oauth2_tokens": [
      {
        "thirdparty_oauth2_service_id": "660e8400-e29b-41d4-a716-446655440000",
        "scopes": ["repo", "user"]
      }
    ]
  }'
```

**Example: Update Existing Grant**
```bash
curl -X POST http://localhost:8080/api/consent/agent/550e8400-e29b-41d4-a716-446655440000/grants \
  -H "X-Principal: user@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "delegated_oauth2_tokens": [
      {
        "thirdparty_oauth2_service_id": "660e8400-e29b-41d4-a716-446655440000",
        "scopes": ["repo"]
      }
    ]
  }'
```

**Example: Revoke Grant (Empty Tokens)**
```bash
curl -X POST http://localhost:8080/api/consent/agent/550e8400-e29b-41d4-a716-446655440000/grants \
  -H "X-Principal: user@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "delegated_oauth2_tokens": []
  }'
```

**Response for Revocation:** `204 No Content`

**Error Responses:**
- `400 Bad Request` - Invalid request body, validation error, or invalid scopes
  ```json
  {
    "error": "invalid request",
    "message": "valid_until must be in the future"
  }
  ```
  ```json
  {
    "error": "invalid scopes",
    "message": "delegation 0 (service=github) has invalid scopes: [admin:org]"
  }
  ```
  ```json
  {
    "error": "service not found",
    "message": "service_id=invalid-service-id"
  }
  ```
- `401 Unauthorized` - No principal in context
- `404 Not Found` - Agent doesn't exist
- `500 Internal Server Error` - Service error

## Business Logic

### Upsert Semantics (FR-013, FR-015)

The grant endpoint implements upsert semantics:
- If no grant exists for (principal, agent), creates new grant
- If grant exists, updates it with new token delegations and expiration
- Each user can have at most one grant per agent

### Scope Validation (FR-018)

All requested scopes must exist in the service's scope configuration:
```json
// Valid: both scopes defined in service
{
  "thirdparty_oauth2_service_id": "github",
  "scopes": ["repo", "user"]
}

// Invalid: "admin:org" not defined in service
{
  "thirdparty_oauth2_service_id": "github",
  "scopes": ["repo", "admin:org"]
}
```

### Grant Expiration (FR-016, FR-019)

- `valid_until` must be in the future if provided
- Expired grants are filtered out when listing grants
- Indefinite grants have `valid_until: null`

### Revocation (FR-014)

Two ways to revoke:
1. POST with empty `delegated_oauth2_tokens` array
2. DELETE agent (cascade deletes all grants for that agent)

Revocation is idempotent - revoking a non-existent grant succeeds.

### Principal Isolation (FR-017)

Each user's grants are isolated by principal:
- User can only view/modify their own grants
- Principal extracted from `X-Principal` header
- No cross-user access possible

## Complete Workflow Example

```bash
# 1. Admin creates agent
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

# 3. User views consent info
curl http://localhost:8080/api/consent/agent/agent-123

# Response: {"agent": {...}, "requested_services": [{...}]}

# 4. User grants consent
curl -X POST http://localhost:8080/api/consent/agent/agent-123/grants \
  -H "X-Principal: user@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "delegated_oauth2_tokens": [
      {
        "thirdparty_oauth2_service_id": "service-456",
        "scopes": ["repo"]
      }
    ]
  }'

# Response: {"id": "grant-789", ...}

# 5. User lists grants
curl http://localhost:8080/api/consent/agent/agent-123/grants \
  -H "X-Principal: user@example.com"

# Response: [{"id": "grant-789", ...}]

# 6. User updates grant (add scope)
curl -X POST http://localhost:8080/api/consent/agent/agent-123/grants \
  -H "X-Principal: user@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "delegated_oauth2_tokens": [
      {
        "thirdparty_oauth2_service_id": "service-456",
        "scopes": ["repo", "user"]
      }
    ]
  }'

# 7. User revokes grant
curl -X POST http://localhost:8080/api/consent/agent/agent-123/grants \
  -H "X-Principal: user@example.com" \
  -H "Content-Type: application/json" \
  -d '{"delegated_oauth2_tokens": []}'

# Response: 204 No Content

# 8. Admin deletes agent (cascade deletes all grants)
curl -X DELETE http://localhost:8081/api/agents/agent-123 \
  -H "X-Principal: admin@example.com"

# 9. Try to delete service with grants (fails)
curl -X DELETE http://localhost:8081/api/third-party/oauth2/clients/service-456 \
  -H "X-Principal: admin@example.com"

# Response: 409 Conflict
```

## Security Considerations

1. **Principal Authentication**: All endpoints (except GET agent consent info) require `X-Principal` header
2. **Secret Redaction**: Client secrets are redacted in GET responses (SR-003)
3. **Principal Isolation**: Users can only access their own grants (FR-017)
4. **Scope Validation**: Only defined scopes can be delegated (FR-018)
5. **Cascade Delete**: Deleting agent removes all grants automatically (FR-020)
6. **Service Protection**: Cannot delete service if grants reference it (409 Conflict)
