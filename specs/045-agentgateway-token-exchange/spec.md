# Feature Specification: Agentgateway Native Token Exchange

**Feature Branch**: `045-agentgateway-token-exchange`  
**Feature Directory**: `specs/045-agentgateway-token-exchange`  
**Created**: 2026-09-16  
**Status**: Draft  
**Input**: User description: "Support agentgateway `oauthTokenExchange` policy as an alternative to ExtProc, with documentation and end-to-end tests proving the Broker interface works."

## Existing Integration and Alternative Path

The current local gateway integration uses agentgateway `v1.5.0`, routes MCP traffic through an `extProc` policy at `extproc-token-exchange:50051`, and depends on the ExtProc service in the Compose deployment. This feature adds a separate direct route configuration: it replaces that route's `extProc` policy with `backendAuth.oauthTokenExchange`, sends the exchange request directly to the Broker's `/oauth2/token` endpoint, and retains the MCP backend target. The direct route MUST NOT configure or contact an ExtProc service. Existing ExtProc routes remain unchanged.

The direct route uses RFC 8693's default token-exchange grant. Its `clientAuth` uses the native `privateKeyJwt` method (or the equivalent `PrivateKeyJwt` spelling in Kubernetes configuration) to create the signed `client_assertion` that the Broker requires. Shared-secret client authentication methods are not compatible with this Broker contract and are excluded from the direct integration.
## E2E Signed Assertion Trust Contract

The direct-policy reference configuration and its E2E fixture use one explicit assertion-and-subject trust tuple. The gateway `clientId` is `https://agentgateway-direct-e2e.example.test`, so agentgateway emits that value as `iss` and `sub` in each `privateKeyJwt` assertion. Its `assertionAudience` is `token-exchange-broker`.

The E2E fixture mints the inbound subject JWT with `iss` equal to the existing Broker-configured upstream fixture issuer and `aud=token-exchange-broker`. The Broker configuration sets `token_exchange.client_assertion.issuer_uri` to `https://agentgateway-direct-e2e.example.test`, `token_exchange.client_assertion.jwks_uri` to the fixture HTTPS JWKS endpoint, and `token_exchange.expected_audience` to `token-exchange-broker`. The fixture serves the matching public key. Only agentgateway receives the private key.

The reference configuration uses the same client-assertion values with deployment-specific URLs and key locations. Operators must mint subject tokens with the Broker-configured upstream issuer and expected audience. The configuration MUST NOT use an assertion audience derived from the Broker token endpoint, a missing JWKS URI, or shared-secret client authentication.


## User Scenarios & Testing *(mandatory)*

### User Story 1 - Exchange Tokens Through the Native Gateway Policy (Priority: P1)

An operator configures agentgateway's native `oauthTokenExchange` policy directly on an MCP or API backend. The policy sends the authenticated inbound credential, resource indicator, and a fresh signed client assertion to the Broker, then forwards only the returned downstream credential to the backend without deploying the separate ExtProc service.

**Why this priority**: This supplies the requested simpler deployment path while keeping credentials out of agents.

**Independent Test**: Start real agentgateway v1.5.0 configuration containing `backendAuth.oauthTokenExchange` and no `extProc` policy or ExtProc endpoint. Send an authenticated agent request through it to a production Broker boundary configured to trust the gateway signing key; verify that the protected downstream service receives the Broker-returned token, never the inbound token.

**Acceptance Scenarios**:

1. **Given** a protected backend has the direct `backendAuth.oauthTokenExchange` route, a matching Broker resource mapping, and the documented E2E trust tuple, **When** an agent sends two authenticated requests through agentgateway, **Then** the Broker's `/oauth2/token` endpoint receives one RFC 8693 request for each exchange. Each request contains the inbound subject JWT with `aud=token-exchange-broker` and the configured resource. Each signed `client_assertion` has `iss` and `sub` equal to `https://agentgateway-direct-e2e.example.test` and `aud` equal to `token-exchange-broker`. The two assertions have different `jti` values.
2. **Given** the Broker authorizes that RFC 8693 exchange, **When** agentgateway receives the exchanged token, **Then** it replaces the inbound `Authorization` credential with the exchanged token before forwarding the request to the protected backend and completes the agent request successfully.
3. **Given** a route uses the direct native token-exchange policy, **When** the agent makes a supported request, **Then** agentgateway sends an RFC 8693 exchange request to the Broker, the route has no `extProc` policy, does not contact an ExtProc endpoint, and does not require the ExtProc service.

---

### User Story 2 - Preserve Broker Authorization Boundaries (Priority: P1)

A security operator needs the direct gateway path to preserve the Broker's signed-client-assertion validation, token validation, client authorization, user-delegation, and resource-authorization controls.

**Why this priority**: A simpler deployment path must not weaken the trust boundary or expose an inbound credential to a downstream service.

**Independent Test**: Exercise invalid client assertion, denied delegation, and Broker-unavailable cases through a real agentgateway direct-policy route without ExtProc; verify that each request stops before the protected backend receives the inbound credential.

**Acceptance Scenarios**:

1. **Given** an invalid credential condition, **When** the direct policy requests exchange, **Then** the Broker rejects the request before resource or delegation checks. The condition can be an untrusted key, an invalid assertion issuer or audience, or an inbound subject JWT with an invalid issuer or audience. The protected backend receives no request carrying the inbound credential.
2. **Given** the gateway presents a valid signed client assertion but the agent has no active delegation for the requested protected resource, **When** agentgateway requests exchange, **Then** the Broker denies the exchange and agentgateway does not forward the inbound credential.
3. **Given** the Broker returns a validation, authorization, or availability failure, **When** agentgateway handles the direct-policy failure, **Then** the agent receives a failure response and the protected backend receives no request.

---

### User Story 3 - Deploy the Alternative Safely (Priority: P2)

An operator needs clear documentation and a reference configuration that explicitly replaces the current ExtProc route with the direct agentgateway policy, rather than combining both paths or copying an incompatible shared-secret sample.

**Why this priority**: The integration is only useful when operators can deliberately select the direct path, configure its signed gateway identity correctly, and retain ExtProc only where they choose it.

**Independent Test**: Follow the documented direct-policy setup using the supplied reference configuration, a private signing key whose public key is trusted by the Broker, and no ExtProc service; complete a successful token-exchange request.

**Acceptance Scenarios**:

1. **Given** an operator is choosing a gateway integration, **When** they read the deployment guide, **Then** they can see that the current ExtProc route and the direct `backendAuth.oauthTokenExchange` route are alternatives, not policies to combine, and can identify the appropriate path for each route.
2. **Given** an operator follows the direct-policy setup, **When** they provide the Broker token endpoint, matching protected resource, `privateKeyJwt` gateway client identity, signing-key source, a JWKS URI, and matching issuer and audience values for the gateway and Broker, **Then** they can configure a working exchange path without embedding a client secret or private key in the route configuration or logs.
3. **Given** an operator needs to verify a direct deployment, **When** they follow the guide's validation steps, **Then** they can confirm the direct route does not use ExtProc and that a downstream service receives an exchanged token rather than the inbound credential.

### Edge Cases

- A direct route without a configured resource receives Broker `invalid_request`; the route does not forward the inbound credential.
- A direct route whose resource does not map to a Broker protected resource receives Broker `invalid_target`; the route does not forward the inbound credential.
- The gateway rejects an unsupported `clientAuth.alg` during configuration load. The route does not serve.
- The Broker rejects an assertion with an untrusted key, invalid issuer, or invalid audience. It also rejects an inbound subject JWT with an invalid issuer or audience before resource or delegation checks. The Broker does not disclose the private key, assertion, inbound token, or exchanged token.
- The Broker returns an authorization, validation, or availability failure: the gateway returns a failure to the agent and does not fall back to forwarding the original credential.
- A route has both `extProc` and direct native token-exchange configured: this is not a supported direct integration; the documented configurations select exactly one path for a route.
- A deployment continues to use ExtProc on another route: the native and ExtProc paths remain separately selectable; this feature does not remove or change ExtProc behavior.
- A user attempts to use client-secret gateway authentication or an agentgateway token-exchange mode other than RFC 8693: it is outside this feature's Broker-compatible direct path and is not documented or tested as such.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a direct agentgateway `backendAuth.oauthTokenExchange` route configuration as an alternative to the existing `extProc` route, which currently targets `extproc-token-exchange:50051` with agentgateway v1.5.0.
- **FR-002**: The direct route MUST send an RFC 8693 form request to the existing Broker `/oauth2/token` endpoint with the inbound agent credential as `subject_token`, the configured protected-resource mapping as `resource`, and `client_assertion` as a fresh private-key-signed JWT generated through agentgateway's `privateKeyJwt` capability. It MUST send `client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer`.
- **FR-003**: The direct reference configuration and E2E fixture MUST set agentgateway `clientId` to `https://agentgateway-direct-e2e.example.test` and `assertionAudience` to `token-exchange-broker`. The fixture MUST mint the inbound subject JWT with `iss` equal to the existing Broker-configured upstream fixture issuer and `aud=token-exchange-broker`. The Broker configuration MUST set `token_exchange.client_assertion.issuer_uri` to the gateway client ID, `token_exchange.client_assertion.jwks_uri` to the fixture HTTPS public-JWKS endpoint, and `token_exchange.expected_audience` to `token-exchange-broker`.
- **FR-004**: The direct route MUST NOT use `clientSecretBasic`, `clientSecretPost`, or any shared-secret client-authentication method for Broker token exchange.
- **FR-005**: After a successful exchange, the direct route MUST replace the inbound authorization credential with only the exchanged credential before forwarding the request to the protected backend.
- **FR-006**: The direct route MUST preserve the Broker as the authority for signed client-assertion validation, subject-token validation, privileged-client authorization, user delegation, and protected-resource authorization.
- **FR-007**: If the Broker rejects or cannot complete an exchange, the direct route MUST fail closed: it MUST return a gateway failure to the agent and MUST NOT forward the inbound credential or the request to the protected backend.
- **FR-008**: The direct route MUST have no `extProc` policy, ExtProc endpoint, or ExtProc service dependency. Its end-to-end proof MUST fail if a direct-policy request is routed through ExtProc.
- **FR-009**: Existing ExtProc routes, configuration, and behavior MUST remain supported and unchanged; operators MUST be able to choose exactly one integration path independently for each route.
- **FR-010**: Operator documentation and reference configuration MUST identify the direct-route replacement in the current gateway setup; explain the Broker endpoint, required RFC 8693 form fields including the signed `client_assertion`, protected-resource mapping, `privateKeyJwt` signing key, gateway client ID, assertion audience, Broker issuer URI, Broker JWKS URI, Broker expected audience, route-selection rules, and fail-closed errors; and show how to verify the downstream credential.
- **FR-011**: Reference configuration and documentation MUST source the private signing key without embedding it in route configuration or logs, and MUST NOT recommend `clientSecretBasic`, `clientSecretPost`, or another shared-secret authentication method for this integration.
- **FR-012**: End-to-end acceptance coverage MUST use a real agentgateway v1.5.0 instance configured with the direct native policy and the Broker's implemented token-exchange boundary, rather than ExtProc, a replacement, or a mocked Broker exchange interface.
- **FR-013**: End-to-end acceptance coverage MUST configure agentgateway with `privateKeyJwt`, configure its `clientId` and `assertionAudience` to the E2E trust tuple, and mint the inbound subject JWT with the configured upstream fixture issuer and expected audience. It MUST publish the matching gateway public key at the configured Broker JWKS URI and prove that the Broker accepts both credentials before resource and delegation authorization.
- **FR-014**: End-to-end acceptance coverage MUST prove the downstream service receives the exchanged credential on success and that no downstream request occurs for invalid assertion, denied delegation, or Broker-unavailable failures.
- **FR-015**: End-to-end acceptance coverage MUST map one test to each acceptance scenario in User Stories 1 through 3.
- **FR-016**: This feature MUST cover the default RFC 8693 token-exchange grant only. Agentgateway JWT-bearer and provider-specific on-behalf-of exchange modes are out of scope.
- **FR-017**: The E2E fixture MUST prove that the Broker rejects a client assertion with invalid issuer or audience values. It MUST also prove that the Broker rejects an inbound subject JWT with an issuer other than the configured upstream fixture issuer or an invalid audience before the protected backend receives a request.
- **FR-018**: The direct reference configuration and documentation MUST require operators to replace the client-assertion issuer and audience, subject-token issuer and audience, JWKS URI, and key locations as one validation contract. It MUST identify the gateway and Broker values for the client assertion and the Broker-configured upstream issuer for the subject token.

### Key Entities

- **Native Gateway Exchange Policy**: The `backendAuth.oauthTokenExchange` route configuration that sends an inbound agent credential to the Broker and forwards the Broker-returned credential to a protected backend.
- **Protected Resource Mapping**: The association between the direct route's resource indicator and the Broker-managed service that may provide a downstream credential.
- **Gateway Client Assertion**: A fresh JWT signed by the gateway's private key to identify the privileged gateway client to the Broker. In the E2E contract, its `iss` and `sub` equal `https://agentgateway-direct-e2e.example.test`, and its `aud` equals `token-exchange-broker`. The Broker validates it through its matching issuer, audience, and external JWKS configuration.
- **Inbound Subject JWT**: The agent credential exchanged by the direct route. In the E2E contract, its `iss` equals the existing Broker-configured upstream fixture issuer and its `aud` equals `token-exchange-broker`. The Broker rejects it before resource or delegation checks when either value is invalid.
- **Exchanged Credential**: The downstream credential returned after Broker authorization and supplied to the protected backend in place of the inbound credential.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In 100% of documented successful direct-policy acceptance requests, a real agentgateway v1.5.0 route with no ExtProc configuration reaches the protected backend with the exchanged credential and never the inbound credential.
- **SC-002**: In every documented invalid-assertion, denied-delegation, and Broker-unavailable acceptance request, the protected backend receives zero requests containing the inbound credential.
- **SC-003**: End-to-end coverage contains one independently executable test for each of the nine acceptance scenarios in the three user stories. These tests demonstrate the matching gateway client ID and issuer, client-assertion audience, subject-token issuer and audience, and JWKS values. They reject mismatches and prove direct-path bypass of ExtProc.
- **SC-004**: An operator can follow one documented direct-policy setup from the current ExtProc baseline through downstream-token verification without deploying, configuring, or contacting ExtProc for that route.
- **SC-005**: Documentation clearly distinguishes the two supported gateway integration paths, identifies `privateKeyJwt` as the compatible gateway-authentication method, excludes shared-secret and non-RFC-8693 modes, and enables an operator to select the appropriate path without relying on undocumented behavior.

## Assumptions

- The existing Broker token-exchange capability and authorization contract remain the source of truth. This feature does not add or alter the Broker API contract.
- Agentgateway v1.5.0 provides the native `oauthTokenExchange` capability described in the issue. Its `privateKeyJwt` method emits the configured `clientId` as assertion `iss` and `sub`, and its configured `assertionAudience` as `aud`.
- The Broker validates the direct-policy client assertion and inbound subject JWT against `token_exchange.expected_audience`. The fixture uses `token-exchange-broker` for both audiences. It mints subject tokens with the existing Broker-configured upstream fixture issuer. The Broker validates the assertion against the matching `token_exchange.client_assertion.issuer_uri` and `token_exchange.client_assertion.jwks_uri`.
- The existing container-based acceptance environment can run a real agentgateway and the production Broker boundary for deterministic end-to-end verification.
- The feature adds an alternative integration path only. It does not remove ExtProc, migrate existing ExtProc routes, add non-RFC-8693 grant types, or change the Broker's authorization rules.