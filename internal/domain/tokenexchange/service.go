// Package tokenexchange provides domain types and services for RFC 8693 OAuth 2.0 Token Exchange.
package tokenexchange

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2session"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
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
// 5. Authorize privileged client via CEL expression evaluation
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

	// providerService provides service discovery by protected resources
	providerService *thirdparty.ThirdpartyOAuth2ProviderService

	// oauth2SessionService handles OAuth2 session lifecycle including token refresh
	oauth2SessionService *oauth2session.OAuth2SessionService

	// consentService verifies user has granted agent access
	consentService *consent.Service

	// agentRepository resolves agent client_id to internal UUID
	agentRepository ports.AgentRepository

	// config provides token exchange configuration
	config *ports.TokenExchangeConfig
}

// NewTokenExchangeService creates a new token exchange service.
// All dependencies are required and must not be nil.
//
// Parameters:
//   - jwtValidator: Validates JWT signatures and claims
//   - celEvaluator: Extracts claims and evaluates authorization policies
//   - providerService: Looks up services by protected resource
//   - oauth2SessionService: Handles OAuth2 session lifecycle including token refresh
//   - consentService: Verifies user grants and agent access
//   - config: Token exchange configuration
//
// Returns error if any dependency is nil.
func NewTokenExchangeService(
	jwtValidator *JWTValidator,
	celEvaluator *CELEvaluator,
	providerService *thirdparty.ThirdpartyOAuth2ProviderService,
	oauth2SessionService *oauth2session.OAuth2SessionService,
	consentService *consent.Service,
	agentRepository ports.AgentRepository,
	config *ports.TokenExchangeConfig,
) (*TokenExchangeService, error) {
	if jwtValidator == nil {
		return nil, fmt.Errorf("jwtValidator cannot be nil")
	}
	if celEvaluator == nil {
		return nil, fmt.Errorf("celEvaluator cannot be nil")
	}
	if providerService == nil {
		return nil, fmt.Errorf("providerService cannot be nil")
	}
	if oauth2SessionService == nil {
		return nil, fmt.Errorf("oauth2SessionService cannot be nil")
	}
	if consentService == nil {
		return nil, fmt.Errorf("consentService cannot be nil")
	}
	if agentRepository == nil {
		return nil, fmt.Errorf("agentRepository cannot be nil")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	return &TokenExchangeService{
		jwtValidator:         jwtValidator,
		celEvaluator:         celEvaluator,
		providerService:      providerService,
		oauth2SessionService: oauth2SessionService,
		consentService:       consentService,
		agentRepository:      agentRepository,
		config:               config,
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
// 6. Authorize privileged client via CEL expression evaluation
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

	// Step 6: Authorize privileged client via CEL expression evaluation
	clientAssertionClaims := jwtToClaims(clientAssertionJWT)
	requestContext := &CELRequestContext{
		Resource:      req.Resource,
		GrantType:     req.GrantType,
		Scope:         req.Scope,
		Principal:     principal,
		AgentClientID: agentClientID,
	}
	authorized, err := s.celEvaluator.AuthorizePrivilegedClient(clientAssertionClaims, subjectTokenClaims, *requestContext)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, NewAccessDeniedError("privileged client authorization failed")
	}

	// Step 7: Normalize resource URI
	normalizedResource := Normalize(req.Resource)

	// Step 8: Lookup service by resource URI
	service, err := s.providerService.FindByProtectedResource(ctx, normalizedResource)
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
	// T060 (Feature 021): agentClientID is the broker-internal agent UUID (resolved by CEL).
	// Parse it as UUID and look up by primary key — no GetByClientID needed.
	agentID, parseErr := id.ParseAgentID(agentClientID)
	if parseErr != nil {
		return nil, NewInvalidRequestError(
			fmt.Sprintf("agent_id %q extracted from subject_token is not a valid agent UUID", agentClientID),
		)
	}
	agent, err := s.agentRepository.Get(ctx, agentID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, NewAccessDeniedErrorWithDetails(
				"user has not granted permission for this agent to access the requested service",
				fmt.Sprintf("agent with id %q not found", agentClientID),
			)
		}
		return nil, NewServerErrorWithCause("failed to lookup agent by agent_id", err)
	}
	// T061-T065: Delegate grant verification to ConsentService using internal agent UUID
	// ConsentService.VerifyAgentAccess checks:
	// - T061: Query UserGrant by principal + agent UUID
	// - T062: Check grant status: active, not revoked, not expired
	// - T063: Return error for missing grant
	// - T064: Return error for revoked grant
	// - T065: Return error for expired grant
	_, err = s.consentService.VerifyAgentAccess(ctx, id.Principal(principal), agent.ID)
	if err != nil {
		// Map ConsentService errors to TokenExchange errors
		if errors.Is(err, consent.ErrAgentAccessDenied) {
			// T063: User has not granted agent access
			return nil, NewAccessDeniedErrorWithDetails(
				"user has not granted permission for this agent to access the requested service",
				err.Error(),
			)
		}
		if errors.Is(err, consent.ErrGrantExpired) {
			// T065: User grant has expired
			return nil, NewAccessDeniedErrorWithDetails(
				"user grant has expired",
				err.Error(),
			)
		}
		// System/repository error
		return nil, NewServerErrorWithCause("failed to verify user grant", err)
	}

	// Step 10: Get valid access token with session metadata (with transparent refresh if needed)
	// This single call handles:
	// - Fetching the session from storage
	// - Checking if access token is expired
	// - Automatically refreshing if refresh token is available
	// - Decrypting and returning valid token along with session metadata
	//
	// Per T075: invalid_grant if session doesn't exist
	// Per T076: invalid_grant if both tokens are expired
	sessionObj, accessToken, err := s.oauth2SessionService.GetValidAccessToken(ctx, id.Principal(principal), service.ID)
	if err != nil {
		// Map oauth2session errors to RFC 8693 token exchange errors
		if errors.Is(err, oauth2session.ErrSessionNotFound) {
			// T075: No session exists for this principal+service combination
			// T077: Include service info and re-auth hint in error_description
			description := fmt.Sprintf(
				"User has no active session with the requested service. Service: %s. Please re-authenticate to %s.",
				service.ID,
				service.DisplayName,
			)
			return nil, NewInvalidGrantError(description)
		}
		if errors.Is(err, oauth2session.ErrSessionExpired) {
			// T076: Both access and refresh tokens are expired
			// T077: Include service info and re-auth hint in error_description
			description := fmt.Sprintf(
				"User session has expired. All tokens are no longer valid. Service: %s. Please re-authenticate to %s.",
				service.ID,
				service.DisplayName,
			)
			return nil, NewInvalidGrantError(description)
		}
		// Other errors (refresh failed, decryption failed, etc)
		return nil, NewServerErrorWithCause("failed to get valid access token", err)
	}

	// Step 11: Build and return RFC 8693 response
	// Calculate expires_in from session's current access token expiration
	now := time.Now().UTC()
	expiresIn := int64(0)
	if sessionObj.AccessTokenExpiresAt != nil && sessionObj.AccessTokenExpiresAt.After(now) {
		expiresIn = int64(sessionObj.AccessTokenExpiresAt.Sub(now).Seconds())
	}

	response := NewTokenExchangeResponseFull(
		accessToken,
		sessionObj.TokenType,
		AccessTokenType, // issued_token_type per RFC 8693
		"",              // refresh token typically not returned in exchange response per RFC 8693
		strings.Join(sessionObj.Scope, " "),
		expiresIn,
	)

	return response, nil
}

// jwtToClaims converts a JWT token to a claims map for CEL evaluation.
// Extracts all claims from the token into a flat map suitable for CEL expressions.
// Uses the lestrrat-go/jwx library API for token introspection via Keys() and Get().
func jwtToClaims(token jwt.Token) map[string]any {
	claims := make(map[string]any)

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
			var value any
			if err := token.Get(key, &value); err == nil {
				claims[key] = value
			}
		}
	}

	return claims
}
