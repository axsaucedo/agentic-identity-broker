# Feature Specification: ExtProc MCP URL Elicitation for Session Errors

**Feature Branch**: `023-extproc-mcp-elicitation`
**Created**: 2026-03-26
**Status**: Implemented

## Overview

When an AI agent's delegated session has expired, the ExtProc token exchange service must signal that the user must re-authenticate — not silently fail. This feature closes two gaps:

1. The identity broker includes a re-auth URL (`error_uri`) in its RFC 6749 error responses when a session is missing or expired.
2. ExtProc translates that `error_uri` into a JSON-RPC `-32042` `URLElicitationRequiredError` returned directly to the calling agent, following the MCP 2025-11-05 specification.

As a prerequisite improvement, JWT validation error messages are also made more informative for operators by including expected `iss`/`aud` values.

---

## User Scenarios & Testing

### User Story 1 — Agent Receives Actionable Re-Auth Signal (Priority: P1)

An AI agent makes an MCP tool call via agentgateway. The user's delegated OAuth2 session to the third-party service has expired. Instead of receiving a generic failure, the agent receives a structured error containing the URL the user must visit to re-authorise.

**Why this priority**: Without this, expired sessions silently block agents with no recovery path. The MCP `URLElicitationRequiredError` is the standard mechanism for agents to request user action.

**Independent Test**: Can be tested end-to-end by triggering a token exchange against a mock broker that returns `error_uri`, and verifying the mcp-go client returns `*mcp.URLElicitationRequiredError` with a non-empty `elicitations[0].url`.

**Acceptance Scenarios**:

1. **Given** an agent sends an MCP request through agentgateway with a valid Bearer token, **when** the identity broker returns `{"error":"invalid_grant","error_uri":"https://..."}` from the token exchange endpoint, **then** the agent receives a JSON-RPC error with code `-32042`, `data.elicitations[0].mode = "url"`, and `data.elicitations[0].url` equal to the broker's `error_uri`.

2. **Given** an agent sends an MCP request through agentgateway, **when** the broker returns a token exchange error *without* `error_uri`, **then** ExtProc returns a standard `500 Internal Server Error` (no elicitation).

3. **Given** an agent sends an MCP request through agentgateway, **when** ExtProc returns a JSON-RPC `-32042` elicitation response, **then** the HTTP status code of the agentgateway response is `200 OK` (per JSON-RPC 2.0 — errors travel over HTTP 200).

---

### User Story 2 — Operator Diagnoses JWT Validation Failures (Priority: P2)

An operator troubleshooting a token exchange failure wants to know precisely *why* a JWT was rejected — not just that validation failed. Error messages must include the expected issuer and audience values so the operator can identify misconfiguration without digging into source code.

**Why this priority**: Operational visibility; does not block agent functionality but significantly reduces mean time to resolution.

**Independent Test**: Triggering a token exchange with a JWT signed for the wrong issuer yields an error description containing the expected issuer value (e.g., `expected iss="https://auth.example.com"`).

**Acceptance Scenarios**:

1. **Given** a JWT with an incorrect `iss` claim, **when** the token exchange service validates it, **then** the error description includes `expected iss=` followed by the configured issuer value.

2. **Given** a JWT with an incorrect `aud` claim, **when** the token exchange service validates it, **then** the error description includes `expected aud=` followed by the configured audience value.

3. **Given** a JWT that has expired, **when** validation runs, **then** the error description states the token has expired (without leaking raw token content).

---

### Edge Cases

- **Elicitation from headers phase only**: agentgateway 0.12.0 commits the upstream request to the backend immediately upon receiving any `RequestHeaders` response. `ImmediateResponse` can therefore only be returned from the headers phase — a body-phase `ImmediateResponse` is never reached. The JSON-RPC `id` is set to `null` (per JSON-RPC 2.0 §5) because the request body is unavailable at headers phase.
- **Broker error without `error_uri`**: broker errors that lack `error_uri` (e.g., `access_denied`) must not produce an elicitation — they fall through to the existing 500/503 path.
- **"Got" claim value not exposed**: jwx error types do not expose the actual claim value (only the expected value). Client-facing messages include only the expected value; the full error detail (containing the actual value) is available in operator logs via `WithCause`.

---

## Requirements

### Functional Requirements

- **FR-001**: The identity broker token exchange endpoint MUST include `error_uri` in RFC 6749 §5.2 error responses when a user session is not found or has expired, pointing to the service-specific re-authorisation URL.

- **FR-002**: ExtProc MUST parse the `error_uri` field from non-2xx token exchange responses and, when present, return an HTTP 200 response with a JSON-RPC `-32042` error body containing an `elicitations` array with one entry of `mode="url"` and `url` set to the `error_uri` value.

- **FR-003**: The JSON-RPC `-32042` elicitation response MUST include a unique `elicitationId` (UUID v4) per response to allow clients to correlate elicitation flows.

- **FR-004**: The JSON-RPC `id` field in the elicitation response MUST be `null` when ExtProc cannot determine the request id (i.e., when responding from the headers phase before the body is available).

- **FR-005**: JWT validation error descriptions MUST include the expected `iss` or `aud` value when the corresponding claim does not match, using typed error matching rather than string inspection.

- **FR-006**: ExtProc MUST return `ImmediateResponse` from the **request headers phase** (not body phase) when an elicitation is triggered. This is required for compatibility with agentgateway 0.12.0, which forwards the request to the backend immediately after processing a headers-phase response.

- **FR-007**: When `error_uri` is absent from a broker error response, ExtProc MUST fall through to existing error handling (500 for unclassified errors, 503 for assertion-expired errors).

### Security Requirements

- **SR-001**: The `error_uri` value MUST be sourced from the broker response and must not be constructed from user-supplied input.
- **SR-002**: JWT validation MUST use typed error matching against the JWT library's error hierarchy — `strings.Contains` on error messages is prohibited as a validation mechanism.
- **SR-003**: Token values (subject tokens, access tokens) MUST NOT appear in client-facing error messages.
- **SR-004**: The `error_uri` re-auth URL MUST use the broker's configured `PublicURL` as the base, ensuring it resolves to the correct public endpoint.

---

## Success Criteria

### Measurable Outcomes

- **SC-001**: An agent client receives `*mcp.URLElicitationRequiredError` (JSON-RPC `-32042`) with a non-empty `elicitations[0].url` when the broker returns `error_uri` on token exchange failure, verified end-to-end through a real agentgateway Docker container.
- **SC-002**: JWT issuer mismatch errors include the expected issuer value in the error description, verifiable in unit tests without running a full server.
- **SC-003**: Broker errors without `error_uri` continue to produce 500 responses (no regression in existing error handling).
- **SC-004**: All unit and E2E tests pass with the race detector enabled (`go test -race`).

---

## Assumptions

- agentgateway version is 0.12.0. Future versions may process body-phase `ImmediateResponse` correctly; this spec is valid for 0.12.0 behaviour.
- The MCP `URLElicitationRequiredError` (-32042) is as defined in the MCP specification 2025-11-05.
- The broker's public base URL for re-auth links is already configured via `ServerInstanceConfig.PublicURL` — no new configuration field is needed.
- RFC 6749 §5.2 `error_uri` is an optional extension field; existing clients that do not inspect it are unaffected.
