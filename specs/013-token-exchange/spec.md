# Feature Specification: RFC 8693 OAuth 2.0 Token Exchange

**Feature Branch**: `013-token-exchange`  
**Created**: 2026-01-15  
**Status**: Draft  
**Input**: User description: "Implement RFC 8693 OAuth 2.0 Token Exchange via /oauth2/token endpoint enabling exchange of tokens issued by the Upstream OAuth2 Server to third-party OAuth2 tokens. Both subject_token and client_assertion are issued by the Upstream OAuth2 Server. Token retrieval from token vault with principal from subject_token 'sub' claim. Use resource parameter to identify target service. Allow adding protected resources URIs to services via admin API. Configuration should contain authorization options to validate the pre-validated client assertion with CEL support and future OPA support."

## Clarifications

### Session 2026-01-16

- Q: Who initiates the token exchange request? → A: A gateway (e.g., API gateway, reverse proxy) triggers the token exchange on behalf of agents, not agents directly
- Q: Which endpoint handles token exchange? → A: The existing `/oauth2/token` endpoint, detecting token exchange via grant_type parameter
- Q: What authorization checks are required beyond session existence? → A: System must verify user has granted the specific service to the specific agent (UserGrant check)
- Q: What is the naming for resource URIs on services? → A: `protected_resources` (not expected_resources)
- Q: What token_type value should be returned in RFC 8693 response? → A: Return the stored token_type from the third-party service's original token response (pass-through)
- Q: What level of detail for principal in audit logs? → A: Log full principal in plaintext (enables full audit trail and incident investigation)
- Q: How are gateway and agent identities represented in tokens? → A: client_assertion identifies the gateway only; subject_token contains both user principal ('sub' claim) and agent identifier (configurable claim mapping to agent's client_id)
- Q: How are principal and agent claims extracted from subject_token? → A: Both extractions are configurable via CEL expressions (e.g., `subject_token.sub` for principal, `subject_token.azp` or custom claim for agent_client_id)
- Q: Should token_exchange feature be toggleable? → A: No, token exchange is always enabled; no configuration toggle needed
- Q: How should upstream OAuth2 config be handled? → A: Reuse existing root-level upstream_oauth2 configuration directly (no duplication in token_exchange section)
- Q: Should rate limiting be implemented? → A: Deferred to future iteration
- Q: How should resource URIs be stored and compared? → A: Normalize URIs (remove trailing slashes) before storing and comparison

## User Scenarios & Testing *(mandatory)*

<!--
  E2E ACCEPTANCE TESTING (Constitution Principle XIII):
  Each acceptance scenario below MUST have a corresponding end-to-end (E2E) test in tests/e2e/.
  - Each scenario maps 1:1 to one It() block in E2E tests
  - E2E tests MUST be written BEFORE implementation begins (red-green development)
  - E2E tests MUST FAIL initially, proving they test actual functionality
  - Test file naming: tests/e2e/token_exchange_test.go
-->

### User Story 1 - Gateway Exchanges Upstream Token for Third-Party Token (Priority: P1)

A gateway (API gateway or reverse proxy) needs to exchange a token issued by the Upstream OAuth2 Server for a third-party OAuth2 token stored in the token vault. The gateway authenticates itself using client_assertion (JWT) and presents a subject_token (JWT) that contains both the user principal ('sub' claim) and the agent identifier (configurable claim, e.g., 'azp' or custom claim mapping to agent's client_id). The system validates both tokens, extracts user and agent identifiers via configurable CEL expressions, verifies the user has granted the agent access to the target service, looks up the third-party service by the resource parameter, retrieves the user's stored third-party token from the vault, and returns it to the gateway.

**Why this priority**: This is the core capability that enables gateways to obtain delegated third-party tokens for agents. Without this, agents cannot access third-party services on behalf of users.

**Independent Test**: Gateway sends token exchange request with valid subject_token (containing user's sub claim and agent's client_id in configurable claim) and client_assertion (identifying gateway) to `/oauth2/token`, system validates tokens, extracts principal and agent_client_id via CEL, verifies user grant exists for agent+service, looks up service by resource URI, retrieves stored third-party token, and returns the token in RFC 8693 response format.

**Acceptance Scenarios**:

1. **Given** a gateway with valid client_assertion JWT (identifying the gateway) and subject_token JWT (containing user principal and agent identifier) both issued by Upstream OAuth2 Server, **When** gateway sends POST to `/oauth2/token` with grant_type=urn:ietf:params:oauth:grant-type:token-exchange, **Then** system detects token exchange request and processes it
2. **Given** valid token exchange request with resource parameter, **When** system processes request, **Then** system looks up ThirdpartyOAuth2Service by matching resource URI against configured protected_resources
3. **Given** service is found, user has active session with stored tokens, and user has granted the agent access to this service, **When** system retrieves tokens, **Then** system returns the third-party access_token in RFC 8693 response format with token_type and issued_token_type fields
4. **Given** stored third-party access_token has expired but refresh_token is valid, **When** system processes exchange, **Then** system automatically refreshes the token using the refresh_token, stores the new token, and returns the fresh access_token
5. **Given** gateway sends request without valid client_assertion, **When** system validates request, **Then** system returns 401 Unauthorized with error=invalid_client
6. **Given** gateway sends request with invalid or expired subject_token, **When** system validates request, **Then** system returns 400 Bad Request with error=invalid_request and error_description explaining the issue

---

### User Story 2 - Resource-Based Service Discovery (Priority: P1)

The system needs to determine which third-party service to exchange tokens for based on the resource parameter in the token exchange request. Administrators configure protected resource URIs on each service, and the system matches incoming resource parameters against these configurations.

**Why this priority**: This is essential for the token exchange to work - the system must know which service's token to return. Without resource-based lookup, the endpoint cannot route requests.

**Independent Test**: Administrator configures service with protected_resources URIs via admin API, gateway sends token exchange request with matching resource parameter, system correctly identifies the target service.

**Acceptance Scenarios**:

1. **Given** admin creates/updates a service with protected_resources array containing URIs, **When** service is saved, **Then** system stores the protected_resources and makes them available for lookup
2. **Given** token exchange request contains resource parameter, **When** system processes request, **Then** system normalizes the resource URI (removes trailing slashes) and searches services where protected_resources contains the normalized URI
3. **Given** multiple services exist but only one matches the resource, **When** system performs lookup, **Then** system returns tokens for the matching service
4. **Given** no service matches the resource parameter, **When** system processes request, **Then** system returns 400 Bad Request with error=invalid_target and error_description="No service configured for the requested resource"
5. **Given** multiple services match the same resource URI (misconfiguration), **When** system performs lookup, **Then** system returns 400 Bad Request with error=invalid_target and error_description indicating ambiguous configuration

---

### User Story 3 - User Grant Verification (Priority: P1)

The system must verify that the user (identified by subject_token 'sub' claim) has explicitly granted the agent (identified by configurable claim in subject_token, e.g., 'azp') permission to access the target third-party service. This prevents agents from accessing services the user has not authorized.

**Why this priority**: This is a critical security control. Without grant verification, any agent could potentially access any user's third-party tokens, violating the consent model.

**Independent Test**: User has granted Agent A access to GitHub but not Agent B. Agent A's gateway can exchange for GitHub tokens. Agent B's gateway receives access_denied error.

**Acceptance Scenarios**:

1. **Given** user has an active UserGrant for the agent and target service, **When** gateway requests token exchange, **Then** system proceeds with token retrieval
2. **Given** user has NOT granted the agent access to the target service, **When** gateway requests token exchange, **Then** system returns 403 Forbidden with error=access_denied and error_description="User has not granted this agent access to the requested service"
3. **Given** user's grant for the agent+service has been revoked, **When** gateway requests token exchange, **Then** system returns 403 Forbidden with error=access_denied
4. **Given** user's grant exists but has expired, **When** gateway requests token exchange, **Then** system returns 403 Forbidden with error=access_denied and error_description="User grant has expired"

---

### User Story 4 - Gateway Authorization via CEL (Priority: P2)

Administrators need to configure authorization rules that validate client_assertion JWTs beyond basic signature verification. The system uses Common Expression Language (CEL) to express authorization policies that determine whether a gateway is authorized to perform token exchange. Note: The client_assertion identifies the gateway; the agent is identified via claims in the subject_token.

**Why this priority**: This provides fine-grained access control for which gateways can exchange tokens. It depends on P1 (basic exchange must work) and adds security controls.

**Independent Test**: Administrator configures CEL authorization expression, gateway sends token exchange request, system evaluates CEL expression against client_assertion claims, and request is allowed or denied based on expression result.

**Acceptance Scenarios**:

1. **Given** configuration contains CEL authorization expression, **When** token exchange request is received, **Then** system compiles and evaluates CEL expression with client_assertion claims as input context
2. **Given** CEL expression evaluates to true, **When** system processes authorization, **Then** token exchange proceeds normally
3. **Given** CEL expression evaluates to false, **When** system processes authorization, **Then** system returns 403 Forbidden with error=access_denied and error_description from CEL context
4. **Given** CEL expression contains syntax error, **When** system starts up, **Then** system fails to start with clear error message indicating invalid CEL expression
5. **Given** CEL expression references client_assertion claims, **When** system builds context, **Then** system provides claims.sub, claims.aud, claims.iss, claims.exp, claims.iat, and any custom claims
6. **Given** CEL expression references request context, **When** system builds context, **Then** system provides request.resource, request.grant_type, and request.scope

---

### User Story 5 - No Valid Session Returns Appropriate Error (Priority: P2)

When a user has not established a session with a third-party service (no tokens in vault), or their session has expired (both access and refresh tokens expired), the system must return an appropriate error that gateways/agents can handle programmatically to trigger re-authentication flows.

**Why this priority**: This enables agents to detect when users need to re-authenticate with third-party services. It depends on P1 (basic exchange) and provides essential error handling.

**Independent Test**: Gateway requests token exchange for a user who has no session with the target service, system returns error indicating no valid session exists.

**Acceptance Scenarios**:

1. **Given** user has no stored session for the target service, **When** gateway requests token exchange, **Then** system returns 400 Bad Request with error=invalid_grant and error_description="User has no active session with the requested service"
2. **Given** user's stored tokens have all expired (both access and refresh), **When** gateway requests token exchange, **Then** system returns 400 Bad Request with error=invalid_grant and error_description="User session has expired, re-authentication required"
3. **Given** error response is returned, **When** gateway receives response, **Then** response includes sufficient information for agent to redirect user to appropriate re-authentication flow

---

### User Story 6 - Admin Configures Protected Resources on Services (Priority: P1)

Administrators need to configure which resource URIs map to which third-party services. This is done by adding a `protected_resources` field to the service configuration via the admin API.

**Why this priority**: This is required for resource-based service discovery (P1). Without this, the system cannot map resource parameters to services.

**Independent Test**: Administrator updates service via PUT /api/services/{id} with protected_resources array, system stores the configuration, subsequent token exchange requests with matching resource parameter route to this service.

**Acceptance Scenarios**:

1. **Given** admin sends PUT request to update service, **When** request includes protected_resources array, **Then** system validates URIs, normalizes them (removes trailing slashes), and stores them with the service
2. **Given** admin provides invalid URI in protected_resources, **When** system validates request, **Then** system returns 400 Bad Request with validation error
3. **Given** admin provides duplicate URI that exists on another service, **When** system validates request, **Then** system returns 409 Conflict indicating resource URI already mapped to another service
4. **Given** admin sends GET request for service, **When** system returns service, **Then** response includes protected_resources array
5. **Given** admin creates new service via POST, **When** request includes protected_resources, **Then** system stores the resources with the new service

---

### Edge Cases

- **Token refresh during exchange fails**: When access_token is expired and refresh attempt fails (invalid_grant from third-party), system returns error=invalid_grant with description indicating refresh failed and re-authentication required
- **Subject token missing required claim**: When subject_token JWT is missing the claim targeted by principal_expression or agent_client_id_expression, system returns error=invalid_request with description indicating which claim extraction failed
- **Resource parameter missing**: When token exchange request omits resource parameter, system returns error=invalid_request with description "resource parameter is required"
- **Multiple resources in request**: RFC 8693 allows multiple resource parameters; system validates that all resources map to the same service, otherwise returns error=invalid_target
- **Client assertion signature verification fails**: When client_assertion JWT signature cannot be verified against Upstream OAuth2 Server's JWKS, system returns error=invalid_client
- **Resource URI with trailing slash**: When resource parameter contains trailing slash, system normalizes it before lookup (e.g., `https://api.github.com/` becomes `https://api.github.com`)
- **Concurrent refresh attempts**: When multiple exchange requests trigger refresh for same user/service, system uses database locking to ensure only one refresh occurs
- **CEL expression timeout**: CEL evaluation has 100ms timeout; if exceeded, system returns error=server_error with description "Authorization evaluation timeout"
- **Grant exists but session missing**: When user has granted agent access but has no session with the service, system returns error=invalid_grant (not access_denied) to indicate missing session rather than missing permission
- **Non-token-exchange grant_type**: When `/oauth2/token` receives a grant_type other than token-exchange, system passes through to existing proxy behavior

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST handle POST requests to `/oauth2/token` endpoint and detect token exchange via grant_type parameter
- **FR-002**: System MUST accept requests with Content-Type: application/x-www-form-urlencoded per RFC 8693
- **FR-003**: System MUST detect token exchange when grant_type equals `urn:ietf:params:oauth:grant-type:token-exchange`
- **FR-004**: System MUST validate subject_token parameter is a valid JWT issued by the Upstream OAuth2 Server
- **FR-005**: System MUST extract user principal from subject_token using a configurable CEL expression (default: `subject_token.sub`)
- **FR-005a**: System MUST extract agent_client_id from subject_token using a configurable CEL expression (default: `subject_token.azp`)
- **FR-006**: System MUST validate client_assertion parameter is a valid JWT issued by the Upstream OAuth2 Server; the client_assertion identifies the **gateway** (not the agent)
- **FR-007**: System MUST validate client_assertion_type equals `urn:ietf:params:oauth:client-assertion-type:jwt-bearer`
- **FR-008**: System MUST require resource parameter to identify target third-party service
- **FR-009**: System MUST look up ThirdpartyOAuth2Service by matching resource URI against protected_resources field (after normalization)
- **FR-009a**: System MUST normalize resource URIs by removing trailing slashes before storage and comparison
- **FR-010**: System MUST verify user has an active UserGrant for the agent (agent_client_id extracted from subject_token) and target service before retrieving tokens
- **FR-011**: System MUST retrieve stored third-party OAuth2 tokens from token vault using principal and service ID
- **FR-012**: System MUST return RFC 8693 compliant response with access_token, token_type (pass-through from stored third-party token response), issued_token_type, and optional expires_in
- **FR-013**: System MUST set issued_token_type to the original token type from the third-party service (typically urn:ietf:params:oauth:token-type:access_token)
- **FR-014**: System MUST automatically refresh expired access_tokens if refresh_token is available and valid
- **FR-015**: System MUST return appropriate RFC 8693 error responses (invalid_request, invalid_client, invalid_grant, invalid_target, access_denied)
- **FR-016**: System MUST support CEL-based authorization for client_assertion (gateway) validation
- **FR-017**: System MUST validate CEL expressions at startup and fail fast on syntax errors
- **FR-018**: System MUST validate claim extraction CEL expressions (principal_expression, agent_client_id_expression) at startup

### Admin API Requirements

- **API-001**: PUT /api/services/{service-id} MUST accept optional `protected_resources` array field
- **API-002**: POST /api/services MUST accept optional `protected_resources` array field
- **API-003**: GET /api/services and GET /api/services/{service-id} MUST return `protected_resources` field
- **API-004**: protected_resources values MUST be valid URI strings
- **API-004a**: protected_resources URIs MUST be normalized (trailing slashes removed) before storage
- **API-005**: protected_resources MUST be unique across all services (no duplicate URIs after normalization)
- **API-006**: protected_resources validation MUST return 400 for invalid URIs, 409 for duplicate URIs (checked after normalization)

### Configuration Requirements

**Configuration Parameters**:
- **token_exchange.claim_extraction.principal_expression**: (string) CEL expression to extract user principal from subject_token. Default: `subject_token.sub`
- **token_exchange.claim_extraction.agent_client_id_expression**: (string) CEL expression to extract agent identifier from subject_token. Default: `subject_token.azp`
- **token_exchange.authorization.type**: (string) Authorization method: "cel" or "opa" (OPA reserved for future). Default: "cel"
- **token_exchange.authorization.cel.expression**: (string) CEL expression for gateway authorization. Default: "true" (allow all valid gateways)
- **token_exchange.refresh.enabled**: (boolean) Enable automatic token refresh. Default: true

**Upstream OAuth2 Configuration**: Token exchange reuses the existing root-level `upstream_oauth2` configuration (issuer, jwks_uri, audience). No separate token_exchange-specific upstream config is needed.

**Note**: Token exchange is always enabled. There is no toggle to disable this feature.

**Example YAML Configuration**:
```yaml
# Root-level upstream OAuth2 config (shared, already exists)
upstream_oauth2:
  issuer: "https://upstream-oauth2.example.com"
  jwks_uri: "https://upstream-oauth2.example.com/.well-known/jwks.json"
  audience: "agentic-identity-broker"

# Token exchange specific configuration
token_exchange:
  claim_extraction:
    # CEL expression to extract user principal from subject_token
    principal_expression: "subject_token.sub"
    # CEL expression to extract agent identifier from subject_token
    agent_client_id_expression: "subject_token.azp"
  authorization:
    type: cel
    cel:
      # CEL expression for gateway authorization (client_assertion identifies gateway)
      expression: |
        client_assertion.iss == "https://upstream-oauth2.example.com" &&
        "token-exchange" in client_assertion.scope
  refresh:
    enabled: true
```

**Configuration Location**: Will be added to `examples/config/token-exchange.yaml` and referenced in `examples/config/README.md`

### API Requirements

Per Constitution Principle IV and X, API design precedes implementation:

- **API-007**: Token exchange endpoint MUST be documented in `/api/enduser/openapi.yaml`
- **API-008**: Admin API changes (protected_resources) MUST be documented in `/api/admin/openapi.yaml`
- **API-009**: APIs MUST follow Zalando RESTful API Guidelines
- **API-010**: Error responses MUST follow RFC 8693 Section 5.2 error response format

### Database Requirements

- **DB-001**: Migration MUST add `protected_resources` column (TEXT[] or JSONB) to thirdparty_services table
- **DB-002**: Migration file MUST follow naming convention: `005_add_service_protected_resources.up.sql` (adjust number as needed)
- **DB-003**: Migration MUST include corresponding down migration
- **DB-004**: Index MUST be created on protected_resources for efficient lookup
- **DB-005**: Unique constraint SHOULD be enforced at application level (checking across all services)

### Security Requirements

- **SR-001**: JWT signature verification MUST use Upstream OAuth2 Server's published JWKS
- **SR-002**: JWKS MUST be cached with configurable TTL and automatic refresh
- **SR-003**: All token exchange requests MUST be logged for audit including full principal (plaintext), agent_id, service_id, resource, outcome, and timestamp (excluding token values)
- **SR-004**: CEL expressions MUST NOT allow access to system resources (sandboxed execution)
- **SR-005**: Token values MUST never appear in logs (redact access_token, refresh_token)
- **SR-006**: System MUST fail closed: if JWT validation fails, deny the request
- **SR-007**: System MUST verify UserGrant exists before returning tokens (consent enforcement)
- **SR-008**: CEL expressions for claim extraction MUST be validated at startup and fail fast on errors

### Domain Model

**Actors**:

- **Gateway**: An API gateway or reverse proxy that intercepts agent requests and triggers token exchange on behalf of agents. The gateway authenticates using client_assertion JWT.

- **Agent**: An AI agent that needs access to third-party services. The agent does not directly call the token exchange endpoint; instead, the gateway acts on its behalf.

- **User**: The end-user (principal) whose third-party tokens are being exchanged. Identified by subject_token 'sub' claim.

**Entities**:

- **TokenExchangeRequest**: Represents an incoming RFC 8693 token exchange request. Contains grant_type, subject_token, subject_token_type, client_assertion, client_assertion_type, resource, and optional scope.

- **TokenExchangeResponse**: RFC 8693 compliant response containing access_token, issued_token_type, token_type, and optional expires_in, scope, refresh_token.

**Value Objects**:

- **ClientAssertion**: Validated client_assertion JWT with extracted claims (iss, sub, aud, exp, custom claims). Identifies the **gateway** making the token exchange request. The gateway's identity is used for authorization and audit logging. Immutable after validation.

- **SubjectToken**: Validated subject_token JWT containing both the user principal and agent identifier. The principal (extracted via configurable CEL, default: `sub` claim) identifies the user. The agent_client_id (extracted via configurable CEL, default: `azp` claim) identifies which agent is requesting access. Both values are extracted using configurable CEL expressions. Immutable after validation.

- **ResourceURI**: URI identifying the target resource/service for token exchange. Must be valid URI format. Normalized (trailing slashes removed) before storage and comparison. Matched against service protected_resources.

**Domain Events**:

- **TokenExchangeSucceeded**: Fired when token exchange completes successfully. Contains principal, service_id, agent_client_id (extracted from subject_token), gateway_id (from client_assertion sub), timestamp.

- **TokenExchangeFailed**: Fired when token exchange fails. Contains principal (if available), error_code, error_description, agent_client_id (if available), gateway_id (if available), timestamp.

- **TokenRefreshed**: Fired when automatic token refresh occurs during exchange. Contains principal, service_id, timestamp.

### Key Entities

- **ThirdpartyOAuth2Service** (extended): Existing entity extended with `protected_resources []string` field for resource-based lookup
- **UserSession**: Existing entity providing stored OAuth2 tokens for principals per service
- **UserGrant**: Existing entity representing user's consent for an agent to access a service. Must be checked during token exchange.
- **TokenExchangeRequest**: New value object representing the incoming exchange request parameters
- **TokenExchangeResponse**: New value object representing the outgoing RFC 8693 response

*Domain concepts should be added to ARCHITECTURE.md Glossary (per Constitution Principle V)*

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Gateways can successfully exchange Upstream OAuth2 tokens for third-party tokens in under 500ms (excluding token refresh scenarios)
- **SC-002**: Token exchange requests with automatic refresh complete in under 2 seconds
- **SC-003**: System correctly rejects 100% of invalid token exchange requests with appropriate RFC 8693 error codes
- **SC-004**: Administrators can configure protected_resources on services via admin API without system restart
- **SC-005**: CEL authorization expressions are evaluated in under 100ms
- **SC-006**: System handles 100 concurrent token exchange requests without degradation
- **SC-007**: All token exchange operations are logged for audit with sufficient detail for security review (excluding sensitive token values)
- **SC-008**: System correctly enforces UserGrant verification - 100% of requests without valid grants are rejected with access_denied

## Assumptions

- The Upstream OAuth2 Server is a trusted identity provider already integrated with the system
- Both subject_token and client_assertion JWTs use RS256 or ES256 algorithms with keys available via JWKS endpoint
- The Upstream OAuth2 Server's JWKS endpoint is reliably available
- Third-party services store refresh_tokens alongside access_tokens (established in 008-thirdparty-oauth2-sessions)
- Gateways are pre-registered and can obtain valid client_assertion JWTs from the Upstream OAuth2 Server
- Resource URIs are globally unique and consistently used by gateways (no ambiguity in resource-to-service mapping)
- The subject_token contains both the user principal (typically `sub` claim) and agent identifier (typically `azp` claim), both extractable via configurable CEL expressions
- The client_assertion `sub` claim contains the gateway identifier for authorization and audit purposes
- An existing `/oauth2/token` endpoint exists or will be created to handle multiple grant types

## Out of Scope

- OPA (Open Policy Agent) integration for authorization (reserved for future feature)
- Actor token support (RFC 8693 actor_token parameter) - may be added in future iteration
- Token introspection endpoint (RFC 7662)
- Token revocation propagation to third-party services
- Multi-tenant configurations with different upstream OAuth2 servers
- Non-token-exchange grant types on `/oauth2/token` (proxy passthrough behavior)
- Rate limiting per gateway (deferred to future iteration)
