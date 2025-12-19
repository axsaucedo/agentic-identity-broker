# Admin APIs

Admin APIs provide management interfaces for agents and third-party OAuth2 services. These APIs are exposed on the admin server port and require principal authentication.

## Agent Management

### Create Agent

Creates a new agent registration.

**Endpoint:** `POST /api/agents`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Request Body:**
```json
{
  "client_id": "my-agent",
  "display_name": "My AI Agent",
  "description": "An AI agent for task automation",
  "governance_url": "https://example.com/governance",
  "user_documentation_url": "https://example.com/docs",
  "agent_interface_url": "https://example.com/chat"
}
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "client_id": "my-agent",
  "display_name": "My AI Agent",
  "description": "An AI agent for task automation",
  "governance_url": "https://example.com/governance",
  "user_documentation_url": "https://example.com/docs",
  "agent_interface_url": "https://example.com/chat",
  "created_at": "2025-12-17T12:00:00Z",
  "updated_at": "2025-12-17T12:00:00Z"
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/api/agents \
  -H "X-Principal: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "github-assistant",
    "display_name": "GitHub Assistant",
    "description": "AI assistant for GitHub operations"
  }'
```

### Get Agent

Retrieves agent details by ID.

**Endpoint:** `GET /api/agents/:agent-id`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "client_id": "my-agent",
  "display_name": "My AI Agent",
  "description": "An AI agent for task automation",
  "governance_url": "https://example.com/governance",
  "user_documentation_url": "https://example.com/docs",
  "agent_interface_url": "https://example.com/chat",
  "created_at": "2025-12-17T12:00:00Z",
  "updated_at": "2025-12-17T12:00:00Z"
}
```

**Example:**
```bash
curl http://localhost:8081/api/agents/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-Principal: admin@example.com"
```

### Update Agent

Updates an existing agent's metadata.

**Endpoint:** `PUT /api/agents/:agent-id`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Request Body:**
```json
{
  "display_name": "Updated Agent Name",
  "description": "Updated description",
  "governance_url": "https://example.com/governance-v2"
}
```

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "client_id": "my-agent",
  "display_name": "Updated Agent Name",
  "description": "Updated description",
  "governance_url": "https://example.com/governance-v2",
  "user_documentation_url": "https://example.com/docs",
  "agent_interface_url": "https://example.com/chat",
  "created_at": "2025-12-17T12:00:00Z",
  "updated_at": "2025-12-17T13:00:00Z"
}
```

**Example:**
```bash
curl -X PUT http://localhost:8081/api/agents/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-Principal: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "GitHub Assistant Pro",
    "description": "Enhanced AI assistant for GitHub"
  }'
```

### Delete Agent

Deletes an agent and all associated user grants (cascade delete).

**Endpoint:** `DELETE /api/agents/:agent-id`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Response:** `204 No Content`

**Example:**
```bash
curl -X DELETE http://localhost:8081/api/agents/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-Principal: admin@example.com"
```

### List Agents

Retrieves all registered agents.

**Endpoint:** `GET /api/agents`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Response:** `200 OK`
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "client_id": "my-agent",
    "display_name": "My AI Agent",
    "description": "An AI agent for task automation",
    "created_at": "2025-12-17T12:00:00Z",
    "updated_at": "2025-12-17T12:00:00Z"
  }
]
```

**Example:**
```bash
curl http://localhost:8081/api/agents \
  -H "X-Principal: admin@example.com"
```

## Third-Party OAuth2 Service Management

### Create OAuth2 Service

Registers a new third-party OAuth2 service.

**Endpoint:** `POST /api/third-party/oauth2/clients`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Request Body:**
```json
{
  "client_id": "github-oauth",
  "client_secret": "secret123",
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
}
```

**Response:** `201 Created`
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "client_id": "github-oauth",
  "client_secret": "[REDACTED]",
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
  ],
  "created_at": "2025-12-17T12:00:00Z",
  "updated_at": "2025-12-17T12:00:00Z"
}
```

**Note:** Client secrets are redacted in GET responses for security (SR-003).

**Example:**
```bash
curl -X POST http://localhost:8081/api/third-party/oauth2/clients \
  -H "X-Principal: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "github-oauth",
    "client_secret": "ghp_secret123",
    "display_name": "GitHub",
    "scopes": [
      {"scope_value": "repo", "description": "Repository access"},
      {"scope_value": "user", "description": "User profile"}
    ]
  }'
```

### Get OAuth2 Service

Retrieves service details by ID (client secret redacted).

**Endpoint:** `GET /api/third-party/oauth2/clients/:client-id`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Response:** `200 OK`
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "client_id": "github-oauth",
  "client_secret": "[REDACTED]",
  "display_name": "GitHub",
  "scopes": [
    {
      "scope_value": "repo",
      "description": "Access to repositories"
    }
  ],
  "created_at": "2025-12-17T12:00:00Z",
  "updated_at": "2025-12-17T12:00:00Z"
}
```

**Example:**
```bash
curl http://localhost:8081/api/third-party/oauth2/clients/660e8400-e29b-41d4-a716-446655440000 \
  -H "X-Principal: admin@example.com"
```

### Update OAuth2 Service

Updates service configuration.

**Endpoint:** `PUT /api/third-party/oauth2/clients/:client-id`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Request Body:**
```json
{
  "display_name": "GitHub Enterprise",
  "scopes": [
    {
      "scope_value": "repo",
      "description": "Repository access"
    },
    {
      "scope_value": "admin:org",
      "description": "Organization admin"
    }
  ]
}
```

**Response:** `200 OK`

**Example:**
```bash
curl -X PUT http://localhost:8081/api/third-party/oauth2/clients/660e8400-e29b-41d4-a716-446655440000 \
  -H "X-Principal: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "GitHub Enterprise",
    "scopes": [
      {"scope_value": "repo", "description": "Repository access"}
    ]
  }'
```

### Delete OAuth2 Service

Deletes an OAuth2 service. Returns 409 Conflict if active grants reference this service.

**Endpoint:** `DELETE /api/third-party/oauth2/clients/:client-id`

**Request Headers:**
- `X-Principal: <principal-identifier>` (required)

**Response:**
- `204 No Content` - Service deleted successfully
- `409 Conflict` - Service has active grants (cannot delete)

**Example:**
```bash
curl -X DELETE http://localhost:8081/api/third-party/oauth2/clients/660e8400-e29b-41d4-a716-446655440000 \
  -H "X-Principal: admin@example.com"
```

## Error Responses

All endpoints may return these error responses:

**400 Bad Request**
```json
{
  "error": "invalid request",
  "message": "display_name is required"
}
```

**401 Unauthorized**
```json
{
  "error": "unauthorized",
  "message": "X-Principal header is required"
}
```

**404 Not Found**
```json
{
  "error": "not found",
  "message": "agent with ID '...' not found"
}
```

**409 Conflict**
```json
{
  "error": "conflict",
  "message": "agent with client_id 'my-agent' already exists"
}
```

**500 Internal Server Error**
```json
{
  "error": "internal server error",
  "message": ""
}
```
