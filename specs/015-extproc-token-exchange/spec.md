# Feature Specification: Envoy ExtProc Token Exchange Service

**Feature Branch**: `015-extproc-token-exchange`  
**Created**: 2026-02-23  
**Status**: Draft  
**Input**: User description: "support for token exchange via Envoy's ExtProc interface. I want to implement a new application in this codebase (resulting in a new process defined in cmd/) in this project that implements the ExtProc interface that Envoy exposes. The new cmd should use the same configuration _structure_ but with completely different content. It is designed to be usable with agentgateway so that an agent can call an MCP tool or another agent with transparent token exchange. The new application should only share the infrastructure of the configuration but should not share any structures, i.e. it has its own schema. The ExtProc interface should be implemented to transparently exchange a Bearer token of the request using the identity broker's token exchange capability as specified in specs/013-token-exchange/spec.md and specs/013-token-exchange/quickstart.md. The resource that is required as part of the token exchange should be the request URI that the ExtProc interface receives. The exchanged token should be cached until the lifetime of the token expires. A simple in memory cache is sufficient. The configuration should include: - server and port for the grpc interface - the OAuth2 Authorization Server that supports the token exhange RFC - ClientID + Secret to obtain an id token to be used as the client assertion in the token exchange. We do not need to introduce new domain objects as this is a separate application. Also this must define completely separate e2e tests, the existing test harness cannot be used."

## Clarifications

### Session 2026-02-23

- Q: Behavior when Authorization header is missing or not Bearer? → A: Pass through unchanged (no exchange attempt)
- Q: Behavior when token exchange times out or returns non-200? → A: Return a 500 response via ExtProc and log the failure
- Q: When exchanged token response lacks an expiry, what TTL should the cache use? → A: Use extproc.cache.default_ttl
- Q: What response should ExtProc return when the request URI is empty or invalid for resource? → A: Reject with 503 (service unavailable)
- Q: How should concurrent requests refresh an expired cached token? → A: Use single-flight refresh (one exchange, others wait)

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.

  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently

  E2E ACCEPTANCE TESTING (Constitution Principle XIII):
  Each acceptance scenario below MUST have a corresponding end-to-end (E2E) test in tests/e2e/.
  - Each scenario maps 1:1 to one It() block in E2E tests
  - E2E tests MUST be written BEFORE implementation begins (red-green development)
  - E2E tests MUST FAIL initially, proving they test actual functionality
  - E2E tests change minimally during implementation (fixture adjustments only)
  - E2E tests turn GREEN when implementation satisfies acceptance criteria
  - Use Ginkgo/Gomega framework following patterns in tests/e2e/README.md
  - Test file naming: tests/e2e/[feature]_test.go
  - Include comment references to spec scenarios in E2E test files
-->

### User Story 1 - Transparent Token Exchange for Agent Requests (Priority: P1)

An operator deploys an Envoy External Processing (ExtProc) service so that incoming requests from agents can have their Bearer tokens exchanged for downstream service tokens without changes to the agent’s call pattern.

**Why this priority**: This is the core capability enabling transparent token exchange for agentgateway and downstream MCP tooling.

**Independent Test**: Send a request with a Bearer token through ExtProc and verify the outgoing Authorization header contains the exchanged token tied to the request URI.

**Acceptance Scenarios**:

1. **Given** an incoming request with an Authorization header using the Bearer scheme, **When** ExtProc receives request headers, **Then** it requests a token exchange using the incoming token as the subject token and the request URI as the resource
2. **Given** a successful token exchange response, **When** ExtProc responds to Envoy, **Then** the Authorization header is replaced with the exchanged token and the request continues
3. **Given** the token exchange request fails validation or authorization, **When** ExtProc processes the headers, **Then** the request is rejected with a failure response and the original token is not forwarded

---

### User Story 2 - Token Exchange Cache for Repeated Calls (Priority: P2)

An operator wants repeated requests with the same Bearer token and resource to avoid unnecessary exchanges while the exchanged token is still valid.

**Why this priority**: Reduces latency and load on the authorization server while preserving security.

**Independent Test**: Perform two requests with identical tokens and resource and verify the second request reuses the cached token without a new exchange call.

**Acceptance Scenarios**:

1. **Given** a valid exchanged token stored in cache, **When** a new request arrives with the same subject token and resource, **Then** ExtProc uses the cached token without calling token exchange
2. **Given** a cached exchanged token is expired, **When** a new request arrives, **Then** ExtProc performs a fresh token exchange and updates the cache

---

### User Story 3 - Operable Configuration and Startup Validation (Priority: P3)

An operator configures the ExtProc service with gRPC settings and OAuth2 authorization server details so it starts predictably and fails fast on invalid configuration.

**Why this priority**: Reliable startup and validation are required to deploy the service safely.

**Independent Test**: Launch the service with valid configuration and verify it starts; launch with invalid configuration and verify startup fails with a clear error.

**Acceptance Scenarios**:

1. **Given** valid configuration for gRPC and token exchange settings, **When** the service starts, **Then** it binds to the configured host/port and logs a startup summary
2. **Given** missing or invalid required configuration values, **When** the service starts, **Then** it exits with a configuration validation error

---

### Edge Cases

- What happens when the Authorization header is missing or not a Bearer token?
- How does the service handle token exchange timeouts or non-200 responses from the authorization server? Return a 500 response and log the failure.
- What happens when the request URI is empty or cannot be parsed into a resource value? Reject with 503.
- What happens when the exchanged token response lacks an expiration time? Use the default cache TTL.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST run as a standalone ExtProc gRPC service that can be deployed alongside Envoy
- **FR-002**: System MUST process Envoy ExtProc request headers and inspect the Authorization header
- **FR-003**: System MUST treat the incoming Bearer token as the subject_token for RFC 8693 token exchange
- **FR-004**: System MUST use the request URI (the `:path` pseudo-header from Envoy, which MUST be a non-empty absolute URI with `http` or `https` scheme) as the resource parameter in the token exchange request. **Implementation note**: Agentgateway populates `:path` with the full backend target URL (e.g., `http://mcp-server:9003/mcp`) when routing to MCP backends via HTTP streamable transport, so the absolute URI requirement is satisfied by the gateway. If `:path` contains only a relative path (e.g., `/mcp`), the request MUST be rejected per FR-013.
- **FR-005**: System MUST call the identity broker’s token exchange capability as defined in `specs/013-token-exchange/spec.md`
- **FR-006**: System MUST obtain a client assertion by performing a `client_credentials` grant (with `scope=openid`) against the configured authorization server and extracting the `id_token` from the response to use as the client assertion (per RFC 7523)
- **FR-007**: System MUST send token exchange requests with the client assertion and subject token, and MUST replace the Authorization header with the exchanged access token on success
- **FR-008**: System MUST reject requests when token exchange fails or is unauthorized, without forwarding the original token
- **FR-009**: System MUST pass through requests without an Authorization Bearer token unchanged and without performing token exchange
- **FR-010**: System MUST return a 500 response via ExtProc and log failures when token exchange times out or returns non-200
- **FR-011**: System MUST cache exchanged tokens keyed by subject token and resource until the exchanged token’s expiration time
- **FR-012**: System MUST use extproc.cache.default_ttl when the exchanged token response lacks an expiration time
- **FR-013**: System MUST reject requests with empty or invalid request URIs by returning a 503 response via ExtProc
- **FR-014**: System MUST ensure only one token exchange occurs per subject token/resource when refreshing expired cache entries
- **FR-015**: System MUST evict expired cache entries before reuse and MUST periodically sweep expired entries via a background goroutine to bound memory growth
- **FR-018**: System MUST cap cached token TTL at `extproc.cache.max_ttl` regardless of the `expires_in` value from the token exchange response
- **FR-016**: System MUST allow configuration to be loaded via the shared configuration infrastructure while using its own schema
- **FR-017**: System MUST provide a standalone E2E test harness for ExtProc scenarios that does not reuse the existing E2E harness

### Configuration Requirements *(if applicable - document before implementation)*

**Configuration Parameters**:
- **extproc.grpc.bind**: string, gRPC bind address, default `0.0.0.0`
- **extproc.grpc.port**: integer, gRPC port, default `50051`
- **extproc.grpc.max_concurrent_streams**: integer, maximum concurrent gRPC streams, default `100`
- **extproc.oauth2.token_endpoint**: string (URL), OAuth2 token endpoint that supports RFC 8693 token exchange
- **extproc.oauth2.issuer**: string (URL), issuer identifier for the authorization server
- **extproc.oauth2.client_id**: string, client identifier for obtaining a client assertion
- **extproc.oauth2.client_secret**: string, client secret for obtaining a client assertion
- **extproc.oauth2.client_credentials_endpoint**: string (URL, optional), explicit token endpoint for client_credentials grant; defaults to `{issuer}/oauth/token`
- **extproc.oauth2.exchange_timeout**: duration, timeout for outbound token exchange HTTP calls, default `5s`
- **extproc.oauth2.tls.insecure_skip_verify**: boolean, skip TLS certificate verification (DEV ONLY), default `false`
- **extproc.oauth2.tls.ca_bundle_path**: string, path to custom CA bundle file, default `""`
- **extproc.oauth2.tls.allow_http**: boolean, allow HTTP endpoints (DEV ONLY; disables SR-002 enforcement), default `false`
- **extproc.cache.default_ttl**: duration, fallback TTL when exchanged token lacks expiry information, default `5m`
- **extproc.cache.max_ttl**: duration, maximum cache TTL cap regardless of `expires_in`, default `1h`
- **extproc.log.level**: string, log level (debug/info/warn/error), default `info`
- **extproc.log.format**: string, log format (text/json), default `text`

> **Note**: The complete configuration schema including validation rules and environment variable mapping is specified in [contracts/configuration.md](contracts/configuration.md).

**Example YAML Configuration**:
```yaml
grpc:
  bind: "0.0.0.0"
  port: 50051
oauth2:
  token_endpoint: "https://identity-broker.example.com/oauth2/token"
  issuer: "https://identity-broker.example.com"
  client_id: "extproc-gateway"
  client_secret: "${EXTPROC_CLIENT_SECRET}"
  tls:
    allow_http: false  # Set to true for local development only
cache:
  default_ttl: "5m"
  max_ttl: "1h"
log:
  level: "info"
  format: "text"
```

**Configuration Location**: Will be added to `examples/config/extproc-token-exchange.yaml` and referenced in `examples/config/README.md`

### Security Requirements *(mandatory for security-critical features)*

- **SR-001**: Security controls MUST be enabled by default and fail closed when token exchange or validation fails
- **SR-002**: Token exchange requests MUST use secure transport and verify TLS certificates by default
- **SR-003**: Client secrets and exchanged tokens MUST be treated as sensitive and redacted from logs
- **SR-004**: Security-critical operations MUST emit structured audit logs

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Requests with invalid or unauthorized tokens are rejected 100% of the time
- **SC-002**: The ExtProc service starts with valid configuration in under 10 seconds and fails within 2 seconds on invalid configuration
