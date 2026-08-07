# Tool Approval API

The Tool Approval API enables human-in-the-loop authorization for agent tool calls. When an AI agent attempts to invoke a tool that requires human authorization, the ExtProc gateway creates a pending approval record. The user reviews and approves or denies the request through the consent UI.

## Authentication

| Endpoint | Auth Method |
|---|---|
| `POST /api/approvals` | Subject token + client assertion (dual-auth) |
| `GET /api/approvals` (sync) | Client assertion (CEL) |
| `GET /api/approvals/permanent` | `X-Remote-User` header |
| `GET /api/approvals/{id}` | `X-Remote-User` header |
| `POST /api/approvals/{id}/approve` | `X-Remote-User` header |
| `POST /api/approvals/{id}/deny` | `X-Remote-User` header |
| `POST /api/approvals/{id}/consume` | Subject token (Bearer) |
| `POST /api/approvals/{id}/revoke` | `X-Remote-User` header |

## Approval Flow

```
ExtProc ──POST /api/approvals──▶ Broker ──(pending)──▶ User sees in UI
                                                          │
                                              ┌───────────┴───────────┐
                                              ▼                       ▼
                                   POST .../approve            POST .../deny
                                     (once|session|permanent)    (optional: permanent)
                                              │                       │
                                              ▼                       ▼
ExtProc ◀──GET /api/approvals (long-poll)── Broker updates sync state
                                              │
                                              ▼ (if once + approved)
                                   POST .../consume
```

## Endpoints

### Create Pending Approval

```
POST /api/approvals
```

Creates a pending approval record. Idempotent: duplicate requests for the same tool call return the existing record (200) instead of creating a new one (201).

**Rate Limiting**: Limited to `max_requests_per_minute` per (principal, agent) pair. Returns 429 when exceeded.

### Sync Approval State (Long-Poll)

```
GET /api/approvals
```

Returns all active approvals grouped by (principal, agent) pair. Supports long-poll via:

- `If-None-Match`: ETag from previous response (version number)
- `X-Long-Poll-Timeout`: Seconds to wait for changes (1-120, default 30)

Returns 304 Not Modified if no changes occur within the timeout. The `ETag` header contains the current version number.

**Query Parameters**:
- `principal` (optional): Filter results to a specific principal

### Get Approval Detail

```
GET /api/approvals/{id}
```

Returns full details of a single approval record. The acting principal must match the approval's principal.

### Approve

```
POST /api/approvals/{id}/approve
```

Transitions a pending approval to approved state. Requires a `persistence` field:
- `once`: Single-use, must be consumed after use
- `session`: Valid for the agent session duration
- `permanent`: Persists indefinitely, manageable via consent UI

### Deny

```
POST /api/approvals/{id}/deny
```

Transitions a pending approval to denied state. Optionally accepts `persistence: "permanent"` to create a permanent denial. Request body is optional.

### Consume

```
POST /api/approvals/{id}/consume
```

Marks a `once`-persistence approved approval as consumed. Idempotent: consuming an already-consumed approval returns 200. Returns 422 if the approval is not `once`-persistence or not in approved state.

### List Permanent Approvals

```
GET /api/approvals/permanent
```

Returns all permanent approvals and denials for the authenticated user.

### Revoke Permanent Approval

```
POST /api/approvals/{id}/revoke
```

Revokes a permanent approval or denial, transitioning it to denied state. Returns 422 if the approval is not permanent.

## Persistence Scopes

| Scope | Behavior |
|---|---|
| `once` | Single use. Must be consumed via `POST .../consume` after the tool executes. |
| `session` | Valid for the agent session (scoped by `agent_session_id`). |
| `permanent` | Persists indefinitely. Visible in the consent management UI. Revocable. |

## Error Codes

All error responses use the `ApprovalError` schema:

```json
{
  "error": "error_code",
  "message": "Human-readable description"
}
```

| Code | HTTP Status | Description |
|---|---|---|
| `unauthorized` | 401 | Missing or invalid authentication |
| `bad_request` | 400 | Invalid request body or parameters |
| `forbidden` | 403 | Principal does not match approval owner |
| `not_found` | 404 | Approval not found |
| `gone` | 410 | Approval has expired |
| `rate_limit_exceeded` | 429 | Rate limit exceeded |
| `not_consumable` | 422 | Approval cannot be consumed |
| `not_revocable` | 422 | Approval cannot be revoked |
| `internal_error` | 500 | Unexpected server error |

## OpenAPI Specification

See [`/api/enduser/openapi.yaml`](../../api/enduser/openapi.yaml) for the complete OpenAPI 3.0 specification.
