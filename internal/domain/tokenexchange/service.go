// Package tokenexchange provides domain types and services for RFC 8693 OAuth 2.0 Token Exchange.
package tokenexchange

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"

	storagedomain "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// TokenExchangeService orchestrates the RFC 8693 token exchange flow.
// It coordinates validation, authorization, and token retrieval across multiple components.
//
// The service implements the complete token exchange flow:
// 1. Validate token exchange request structure
// 2. Validate subject_token JWT signature and claims
// 3. Validate client_assertion JWT signature and claims
// 4. Extract user principal and agent client ID from tokens
// 5. Authorize gateway via CEL expression evaluation
// 6. Lookup target service by resource URI
// 7. Verify user has granted agent access to service
// 8. Retrieve stored third-party tokens for user
// 9. Check if tokens are fully expired
// 10. Refresh expired access token if refresh token available
// 11. Return RFC 8693 compliant response
//
// Thread-safe: All methods are safe for concurrent use. The service maintains
// no mutable state, only dependency references.
type TokenExchangeService struct {
	// jwtValidator validates JWT signatures and claims
	jwtValidator *JWTValidator

	// celEvaluator extracts claims and evaluates authorization policies
	celEvaluator *CELEvaluator

	// serviceRepository provides service discovery by protected resources
	serviceRepository ports.ThirdpartyOAuth2ServiceRepository

	// grantRepository verifies user has authorized agent access to service
	grantRepository ports.UserGrantRepository

	// sessionRepository retrieves stored tokens for user
	sessionRepository ports.UserSessionRepository

	// encryptionPort decrypts stored encrypted tokens
	encryptionPort ports.EncryptionPort

	// config provides token exchange configuration
	config *ports.TokenExchangeConfig
}

// NewTokenExchangeService creates a new token exchange service.
// All dependencies are required and must not be nil.
//
// Parameters:
//   - jwtValidator: Validates JWT signatures and claims
//   - celEvaluator: Extracts claims and evaluates authorization policies
//   - serviceRepository: Looks up services by protected resource
//   - grantRepository: Verifies user grants
//   - sessionRepository: Retrieves stored tokens
//   - config: Token exchange configuration
//
// Returns error if any dependency is nil.
func NewTokenExchangeService(
	jwtValidator *JWTValidator,
	celEvaluator *CELEvaluator,
	serviceRepository ports.ThirdpartyOAuth2ServiceRepository,
	grantRepository ports.UserGrantRepository,
	sessionRepository ports.UserSessionRepository,
	encryptionPort ports.EncryptionPort,
	config *ports.TokenExchangeConfig,
) (*TokenExchangeService, error) {
	if jwtValidator == nil {
		return nil, fmt.Errorf("jwtValidator cannot be nil")
	}
	if celEvaluator == nil {
		return nil, fmt.Errorf("celEvaluator cannot be nil")
	}
	if serviceRepository == nil {
		return nil, fmt.Errorf("serviceRepository cannot be nil")
	}
	if grantRepository == nil {
		return nil, fmt.Errorf("grantRepository cannot be nil")
	}
	if sessionRepository == nil {
		return nil, fmt.Errorf("sessionRepository cannot be nil")
	}
	if encryptionPort == nil {
		return nil, fmt.Errorf("encryptionPort cannot be nil")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	return &TokenExchangeService{
		jwtValidator:      jwtValidator,
		celEvaluator:      celEvaluator,
		serviceRepository: serviceRepository,
		grantRepository:   grantRepository,
		sessionRepository: sessionRepository,
		encryptionPort:    encryptionPort,
		config:            config,
	}, nil
}

// Exchange processes a complete token exchange request.
// PHASE 5+: Full implementation deferred to Phase 5+.
// For Phase 4, resource-based service discovery is implemented in HTTP handler.
// This is the main orchestration method that will implement the full token exchange flow.
//
// Flow:
// 1. Validate request structure (grant_type, required parameters)
// 2. Validate subject_token JWT (signature, issuer, audience, expiration)
// 3. Validate client_assertion JWT (signature, issuer, audience, expiration)
// 4. Extract principal from subject_token via CEL
// 5. Extract agent_client_id from subject_token via CEL
// 6. Authorize gateway via CEL expression evaluation
// 7. Normalize resource URI (remove trailing slashes)
// 8. Lookup service by resource URI
// 9. Verify user has granted agent access to service (BEFORE session check per T078)
// 10. Retrieve user's session and stored tokens for service
// 11. Check if tokens are fully expired (T076)
// 12. Refresh expired tokens if refresh_token available
// 13. Return RFC 8693 compliant response
//
// Per spec FR-048, token exchange endpoint detection happens in HTTP layer
// (checking grant_type parameter). This service assumes it's processing
// a token exchange request and will fail if given non-token-exchange requests.
//
// Error handling:
// - InvalidRequest: malformed request, missing required parameters
// - InvalidClient: invalid/missing client_assertion
// - InvalidGrant: invalid/expired subject_token, no session, or all tokens expired (T075, T076)
// - InvalidTarget: no service matches resource URI
// - AccessDenied: user has not granted agent access, or CEL authorization failed
// - ServerError: configuration/system errors
//
// Per SR-005 (Security Rule): Token values are never included in error messages
// or logs. Only metadata (issuer, audience) and request context.
//
// Per SR-058 (Audit Logging): Caller is responsible for logging token exchange
// success/failure with context (principal, service, agent). This method does not log.
//
// PHASE 8 NOTES (T075-T078):
// - T075: Return invalid_grant when no UserSession exists for principal+service
// - T076: Return invalid_grant when both access_token and refresh_token expired
// - T077: Include service_id and re-auth hint in error_description
// - T078: CRITICAL - Grant check MUST occur BEFORE session check to prevent information leakage
func (s *TokenExchangeService) Exchange(ctx context.Context, req *TokenExchangeRequest) (*TokenExchangeResponse, error) {
	// Step 1: Validate request structure
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Step 2: Validate subject_token JWT
	subjectTokenJWT, err := s.jwtValidator.ValidateSubjectToken(ctx, req.SubjectToken)
	if err != nil {
		return nil, err
	}

	// Step 3: Validate client_assertion JWT
	clientAssertionJWT, err := s.jwtValidator.ValidateClientAssertion(ctx, req.ClientAssertion)
	if err != nil {
		return nil, err
	}

	// Step 4: Extract principal from subject_token via CEL
	subjectTokenClaims := jwtToClaims(subjectTokenJWT)
	principal, err := s.celEvaluator.ExtractPrincipal(subjectTokenClaims)
	if err != nil {
		return nil, err
	}

	// Step 5: Extract agent_client_id from subject_token via CEL
	agentClientID, err := s.celEvaluator.ExtractAgentClientID(subjectTokenClaims)
	if err != nil {
		return nil, err
	}

	// Step 6: Authorize gateway via CEL expression evaluation
	clientAssertionClaims := jwtToClaims(clientAssertionJWT)
	requestContext := &CELRequestContext{
		Resource:      req.Resource,
		GrantType:     req.GrantType,
		Scope:         req.Scope,
		Principal:     principal,
		AgentClientID: agentClientID,
	}
	authorized, err := s.celEvaluator.AuthorizeGateway(clientAssertionClaims, subjectTokenClaims, *requestContext)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, NewAccessDeniedError("gateway authorization failed")
	}

	// Step 7: Normalize resource URI
	normalizedResource := Normalize(req.Resource)

	// Step 8: Lookup service by resource URI
	service, err := s.serviceRepository.FindByProtectedResource(ctx, normalizedResource)
	if err != nil {
		if IsTokenExchangeError(err) {
			// InvalidTarget error from repository
			return nil, err
		}
		return nil, NewServerErrorWithCause("failed to lookup service by resource URI", err)
	}

	// Step 9: Verify user has granted agent access to service (T059-T065)
	// CRITICAL (T078): Grant verification MUST occur BEFORE session check
	// This prevents information leakage: if grant is missing, return access_denied (no permission).
	// Only if grant exists but session is missing do we return invalid_grant (no session).
	//
	// T059: Agent client ID extracted from subject_token (already done in Step 5)
	// T060: Look up Agent by agent_client_id (agent lookup deferred to future phase)
	// T061: Query UserGrant by principal + agent_client_id
	// T062: Check grant status: active, not revoked, not expired
	// T063: Return 403 access_denied for missing grant
	// T064: Return 403 access_denied for revoked grant
	// T065: Return 403 access_denied for expired grant
	grant, err := s.grantRepository.FindByPrincipalAndAgent(ctx, principal, agentClientID)
	if err != nil {
		// T063: NotFound error - user has not granted agent access to any service
		// Per Constitution Principle I (Security-First), fail closed with access_denied
		if err == ports.ErrNotFound {
			errorMsg := fmt.Sprintf(
				"user has not granted permission for agent (principal: %s, agent: %s)",
				principal, agentClientID,
			)
			return nil, NewAccessDeniedErrorWithDetails(
				"user has not granted permission for this agent to access the requested service",
				errorMsg,
			)
		}
		return nil, NewServerErrorWithCause("failed to verify user grant", err)
	}

	// T062a: Check grant is active (not expired)
	// T065: Return 403 access_denied with expiration info in error_description if expired
	if grant.ValidUntil != nil && grant.ValidUntil.Before(time.Now().UTC()) {
		expiredMsg := fmt.Sprintf(
			"user grant expired at %s (principal: %s, agent: %s)",
			grant.ValidUntil.Format(time.RFC3339),
			principal, agentClientID,
		)
		return nil, NewAccessDeniedErrorWithDetails(
			fmt.Sprintf("user grant has expired (expired at %s)", grant.ValidUntil.Format(time.RFC3339)),
			expiredMsg,
		)
	}

	// T062b: Check grant is not revoked
	// T064: Return 403 access_denied for revoked grant
	// TODO (T064): Implement revocation check when revocation status is added to UserGrant entity
	// For now, assume no revocation field exists. When revoked field is added:
	// if grant.Revoked {
	//   return nil, NewAccessDeniedErrorWithDetails(
	//     "user grant has been revoked",
	//     fmt.Sprintf("grant revoked (principal: %s, agent: %s)", principal, agentClientID),
	//   )
	// }

	// Step 10: Retrieve user's session and stored tokens for service
	// At this point, we know grant exists, so if session is missing, it's invalid_grant (T075)
	sessionObj, err := s.sessionRepository.FindByPrincipalAndService(ctx, principal, service.ID)
	if err != nil {
		if err == ports.ErrNotFound {
			// T075: No session exists for this principal+service combination
			// T077: Include service info and re-auth hint in error_description
			description := fmt.Sprintf(
				"User has no active session with the requested service. Service: %s. Please re-authenticate to %s.",
				service.ID,
				service.DisplayName,
			)
			return nil, NewInvalidGrantError(description)
		}
		return nil, NewServerErrorWithCause("failed to retrieve user session", err)
	}

	// Verify session exists and has valid tokens
	if sessionObj == nil {
		// T075: Session is nil even though no error occurred (defensive check)
		description := fmt.Sprintf(
			"User has no active session with the requested service. Service: %s. Please re-authenticate to %s.",
			service.ID,
			service.DisplayName,
		)
		return nil, NewInvalidGrantError(description)
	}

	// Step 11: Check if tokens are fully expired
	// T076: Return invalid_grant when both access_token and refresh_token are expired
	now := time.Now().UTC()
	accessTokenExpired := sessionObj.AccessTokenExpiresAt == nil || sessionObj.AccessTokenExpiresAt.Before(now)
	refreshTokenExpired := sessionObj.RefreshTokenExpiresAt == nil || sessionObj.RefreshTokenExpiresAt.Before(now)

	if accessTokenExpired && refreshTokenExpired {
		// Both tokens are expired - user needs to re-authenticate
		// T076: Return 400 invalid_grant with re-auth hint
		// T077: Include service info in error_description
		description := fmt.Sprintf(
			"User session has expired. All tokens are no longer valid. Service: %s. Please re-authenticate to %s.",
			service.ID,
			service.DisplayName,
		)
		return nil, NewInvalidGrantError(description)
	}

	// Step 12: Refresh expired tokens if refresh_token available
	// If access_token has expired but refresh_token is valid, attempt refresh
	// TODO: Implement token refresh using refresh endpoint
	// Per Phase 5+ requirements, refresh logic will be implemented when refresh flow is needed

	// Step 13: Build and return RFC 8693 response
	// Use stored token's type and expiration time
	// Calculate expires_in from access token expiration time
	expiresIn := int64(0)
	if sessionObj.AccessTokenExpiresAt != nil {
		if sessionObj.AccessTokenExpiresAt.After(now) {
			expiresIn = int64(sessionObj.AccessTokenExpiresAt.Sub(now).Seconds())
		}
	}

	// Step 12: Decrypt stored encrypted tokens
	// The session contains encrypted access and refresh tokens
	encContext := map[string]string{
		"principal":  principal,
		"service_id": service.ID,
		"session_id": sessionObj.ID,
	}

	// Decrypt access token
	accessToken, err := s.encryptionPort.Decrypt(ctx, sessionObj.EncryptedAccessToken, encContext)
	if err != nil {
		return nil, NewServerErrorWithCause("failed to decrypt access token", err)
	}

	// Decrypt refresh token (may be nil or empty)
	var refreshToken string
	if len(sessionObj.EncryptedRefreshToken) > 0 {
		decryptedRefresh, err := s.encryptionPort.Decrypt(ctx, sessionObj.EncryptedRefreshToken, encContext)
		if err != nil {
			return nil, NewServerErrorWithCause("failed to decrypt refresh token", err)
		}
		refreshToken = string(decryptedRefresh)
	}

	response := NewTokenExchangeResponseFull(
		string(accessToken),
		sessionObj.TokenType,
		AccessTokenType, // issued_token_type per RFC 8693
		refreshToken,
		strings.Join(sessionObj.Scope, " "),
		expiresIn,
	)

	return response, nil
}

// calculateExpiresIn calculates the time-to-live for an access token in seconds.
// Returns 0 if the session has no expiration or the token has expired.
// This value is included in the RFC 8693 token exchange response to indicate
// how long the returned access token is valid.
func (s *TokenExchangeService) calculateExpiresIn(session *storagedomain.UserSession) int64 {
	if session == nil || session.AccessTokenExpiresAt == nil {
		return 0
	}

	now := time.Now().UTC()
	expiresAt := *session.AccessTokenExpiresAt

	// If already expired, return 0
	if expiresAt.Before(now) {
		return 0
	}

	// Calculate remaining seconds
	return int64(expiresAt.Sub(now).Seconds())
}

// jwtToClaims converts a JWT token to a claims map for CEL evaluation.
// Extracts all claims from the token into a flat map suitable for CEL expressions.
// Uses the lestrrat-go/jwx library API for token introspection via Keys() and Get().
func jwtToClaims(token jwt.Token) map[string]interface{} {
	claims := make(map[string]interface{})

	// Extract standard claims
	if iss, _ := token.Issuer(); iss != "" {
		claims["iss"] = iss
	}
	if sub, _ := token.Subject(); sub != "" {
		claims["sub"] = sub
	}
	if aud, _ := token.Audience(); len(aud) > 0 {
		if len(aud) == 1 {
			claims["aud"] = aud[0]
		} else {
			claims["aud"] = aud
		}
	}
	if exp, _ := token.Expiration(); !exp.IsZero() {
		claims["exp"] = exp.Unix()
	}
	if iat, _ := token.IssuedAt(); !iat.IsZero() {
		claims["iat"] = iat.Unix()
	}
	if nbf, _ := token.NotBefore(); !nbf.IsZero() {
		claims["nbf"] = nbf.Unix()
	}
	if jti, _ := token.JwtID(); jti != "" {
		claims["jti"] = jti
	}

	// Extract all custom claims from the token
	// Use Keys() to get all claim names and Get() to retrieve each value
	for _, key := range token.Keys() {
		// Avoid overwriting standard claims that were already extracted
		if _, exists := claims[key]; !exists {
			// Try to get the claim value
			var value interface{}
			if err := token.Get(key, &value); err == nil {
				claims[key] = value
			}
		}
	}

	return claims
}
