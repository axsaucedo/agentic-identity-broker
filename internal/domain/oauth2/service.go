package oauth2

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// isNotFoundErr returns true if err represents a not-found condition from any storage adapter.
// Handles both ports.ErrNotFound (used in mocks/tests) and storage.StorageError{Kind: ErrorKindNotFound}
// (used by production adapters).
func isNotFoundErr(err error) bool {
	if errors.Is(err, ports.ErrNotFound) {
		return true
	}
	var storageErr *storage.StorageError
	if errors.As(err, &storageErr) {
		return storageErr.Kind == storage.ErrorKindNotFound
	}
	return false
}

// OAuth2Config contains configuration for the OAuth2 service
type OAuth2Config struct {
	// Upstream OAuth2 server authorization endpoint
	UpstreamAuthorizeEndpoint string

	// Upstream OAuth2 server token endpoint
	UpstreamTokenEndpoint string

	// Broker's public URL (from enduser ServerInstanceConfig.PublicURL)
	PublicURL string

	// Supported response types (default: ["code"])
	SupportedResponseTypes []string

	// Supported grant types (default: ["authorization_code", "refresh_token"])
	SupportedGrantTypes []string

	// MultiAgentClient holds optional multi-agent client sharing configuration.
	// When Enabled, multiple agents may share a single upstream OAuth2 client ID.
	MultiAgentClient ports.MultiAgentClientConfig
}

// Service implements the OAuth2Service port
type Service struct {
	agentRepo   ports.AgentRepository
	grantRepo   ports.UserGrantRepository
	sessionRepo ports.UserSessionRepository
	config      *OAuth2Config
	logger      *slog.Logger
}

// NewService creates a new OAuth2Service implementation
func NewService(agentRepo ports.AgentRepository, grantRepo ports.UserGrantRepository, config *OAuth2Config) ports.OAuth2Service {
	return &Service{
		agentRepo: agentRepo,
		grantRepo: grantRepo,
		config:    config,
	}
}

// NewServiceWithSessions creates a new OAuth2Service implementation with session support
// for mandatory requirement validation
func NewServiceWithSessions(
	agentRepo ports.AgentRepository,
	grantRepo ports.UserGrantRepository,
	sessionRepo ports.UserSessionRepository,
	config *OAuth2Config,
	logger *slog.Logger,
) ports.OAuth2Service {
	return &Service{
		agentRepo:   agentRepo,
		grantRepo:   grantRepo,
		sessionRepo: sessionRepo,
		config:      config,
		logger:      logger,
	}
}

// HandleAuthorization processes an OAuth2 authorization request
// Returns an AuthorizationDecision with either:
// - redirect_to_upstream: Valid client with active grant
// - redirect_to_consent: Valid client but no active grant
// - error: Invalid client or server error
//
// Feature 021: client_id MUST be the agent's internal UUID (agent.id), NOT agent.client_id.
func (s *Service) HandleAuthorization(ctx context.Context, req *ports.AuthorizationRequest, principal string) (*ports.AuthorizationDecision, error) {
	// Step 1: Validate client_id as a UUID (agent.id) and resolve the agent.
	// Feature 021: client_id is now the broker's internal agent UUID, not the upstream client_id.
	agentID, err := id.ParseAgentID(string(req.ClientID))
	if err != nil {
		// client_id is not a valid UUID → invalid_client
		redirectURL, _ := buildErrorRedirectURL(req.RedirectURI, req.State, "invalid_client", "client_id must be a valid agent UUID")
		return &ports.AuthorizationDecision{
			Action:      "error",
			ErrorCode:   "invalid_client",
			ErrorDesc:   "client_id must be a valid agent UUID",
			RedirectURL: redirectURL,
		}, nil
	}

	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		if isNotFoundErr(err) {
			// Agent UUID not registered
			redirectURL, _ := buildErrorRedirectURL(req.RedirectURI, req.State, "invalid_client", "Client not registered")
			return &ports.AuthorizationDecision{
				Action:      "error",
				ErrorCode:   "invalid_client",
				ErrorDesc:   "Client not registered",
				RedirectURL: redirectURL,
			}, nil
		}
		// Other error (connection, timeout)
		return &ports.AuthorizationDecision{
			Action:    "error",
			ErrorCode: "server_error",
			ErrorDesc: "Failed to validate client",
		}, nil
	}

	// Step 2: Check if user has active grant for this agent
	grant, err := s.grantRepo.FindByPrincipalAndAgent(ctx, id.Principal(principal), agent.ID)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		// Only treat as error if it's not a "not found" condition (which is valid)
		return &ports.AuthorizationDecision{
			Action:    "error",
			ErrorCode: "server_error",
			ErrorDesc: "Failed to check grant",
		}, nil
	}

	// Step 3: Determine action based on grant status
	if grant == nil || !grant.IsActive() {
		// No active grant or expired - redirect to consent UI
		consentURL := fmt.Sprintf("%s/consent/agent/%s?redirect_uri=%s",
			s.config.PublicURL,
			agent.ID,
			url.QueryEscape(req.OriginalURL),
		)
		return &ports.AuthorizationDecision{
			Action:      "redirect_to_consent",
			RedirectURL: consentURL,
		}, nil
	}

	// Step 4: Check that sessions for all delegated services are not expired.
	// Even if the grant is active, a token exchange will fail when the underlying
	// third-party session has expired. Redirect to the consent screen early so
	// the user can re-authenticate with the affected service instead of getting
	// a cryptic error later.
	if s.sessionRepo != nil {
		expired, err := s.anyDelegatedSessionExpired(ctx, id.Principal(principal), grant.DelegatedOAuth2Tokens)
		if err != nil {
			return &ports.AuthorizationDecision{
				Action:    "error",
				ErrorCode: "server_error",
				ErrorDesc: "Failed to check session status",
			}, nil
		}
		if expired {
			if s.logger != nil {
				s.logger.Warn(
					"DelegatedSessionExpired",
					"agent_id", agent.ID,
					"principal", principal,
				)
			}
			consentURL := fmt.Sprintf("%s/consent/agent/%s?redirect_uri=%s",
				s.config.PublicURL,
				agent.ID,
				url.QueryEscape(req.OriginalURL),
			)
			return &ports.AuthorizationDecision{
				Action:      "redirect_to_consent",
				RedirectURL: consentURL,
			}, nil
		}
	}

	// Step 5: Validate mandatory service requirements (if session repo available)
	if s.sessionRepo != nil && len(agent.ServiceRequirements) > 0 {
		err := s.validateMandatoryRequirements(ctx, principal, agent)
		if err != nil {
			// Mandatory requirement not met - redirect to consent screen
			if s.logger != nil {
				s.logger.Warn(
					"MandatoryRequirementValidationFailed",
					"agent_id", agent.ID,
					"error", err.Error(),
				)
			}
			consentURL := fmt.Sprintf("%s/consent/agent/%s?redirect_uri=%s",
				s.config.PublicURL,
				agent.ID,
				url.QueryEscape(req.OriginalURL),
			)
			return &ports.AuthorizationDecision{
				Action:      "redirect_to_consent",
				RedirectURL: consentURL,
			}, nil
		}
	}

	// Active grant exists and all mandatory requirements satisfied - redirect to upstream OAuth2 server
	upstreamURL := s.buildUpstreamAuthorizeURL(req, agent)
	return &ports.AuthorizationDecision{
		Action:      "redirect_to_upstream",
		RedirectURL: upstreamURL,
	}, nil
}

// buildUpstreamAuthorizeURL constructs the upstream authorization endpoint URL
// preserving all OAuth2 parameters.
//
// Feature 021: uses agent.ClientID (upstream OAuth2 client ID) instead of req.ClientID
// (which is now the broker's internal agent UUID). When MultiAgentClient.Enabled,
// appends the agent's internal UUID as the configured AgentIDParamName query parameter.
func (s *Service) buildUpstreamAuthorizeURL(req *ports.AuthorizationRequest, agent *storage.Agent) string {
	u, _ := url.Parse(s.config.UpstreamAuthorizeEndpoint)
	q := u.Query()

	// Use agent.ClientID as upstream client_id (NOT the broker's internal agent UUID)
	q.Set("client_id", agent.ClientID.String())
	q.Set("redirect_uri", req.RedirectURI)
	q.Set("response_type", req.ResponseType)

	// Add optional parameters if present
	if req.Scope != "" {
		q.Set("scope", req.Scope)
	}
	if req.State != "" {
		q.Set("state", req.State)
	}
	if req.CodeChallenge != "" {
		q.Set("code_challenge", req.CodeChallenge)
	}
	if req.CodeChallengeMethod != "" {
		q.Set("code_challenge_method", req.CodeChallengeMethod)
	}

	// Feature 021 — multi-agent client sharing: inject agent UUID param so upstream
	// can embed it as a claim in the returned token (for MultiAgentTokenVerifier).
	if s.config.MultiAgentClient.Enabled {
		q.Set(s.config.MultiAgentClient.AgentIDParamName, agent.ID.String())
		if s.logger != nil {
			s.logger.Info("AgentIDParamInjected",
				"agent_id", agent.ID.String(),
				"param_name", s.config.MultiAgentClient.AgentIDParamName,
				"upstream_url", s.config.UpstreamAuthorizeEndpoint,
			)
		}
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// GenerateMetadata returns RFC 8414 OAuth2 metadata for this broker
func (s *Service) GenerateMetadata(ctx context.Context) (*ports.MetadataResponse, error) {
	metadata := &ports.MetadataResponse{
		Issuer:                            s.config.PublicURL,
		AuthorizationEndpoint:             fmt.Sprintf("%s/oauth2/authorize", s.config.PublicURL),
		TokenEndpoint:                     fmt.Sprintf("%s/oauth2/token", s.config.PublicURL),
		ResponseTypesSupported:            s.config.SupportedResponseTypes,
		GrantTypesSupported:               s.config.SupportedGrantTypes,
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post", "client_secret_basic"},
	}

	return metadata, nil
}

// anyDelegatedSessionExpired returns true if any existing session for a delegated service is
// expired. A missing session (nil) is not treated as expired — absence means the user simply
// has not logged into that service yet, which is handled separately by mandatory requirement
// validation. Only an existing session whose refresh token has expired triggers a consent redirect.
// Returns an error only on unexpected storage failures.
func (s *Service) anyDelegatedSessionExpired(ctx context.Context, principal id.Principal, tokens []storage.DelegatedToken) (bool, error) {
	for _, token := range tokens {
		session, err := s.sessionRepo.FindByPrincipalAndService(ctx, principal, token.ThirdpartyOAuth2ServiceID)
		if err != nil {
			return false, fmt.Errorf("failed to check session status: %w", err)
		}
		if session != nil && session.IsExpired() {
			return true, nil
		}
	}
	return false, nil
}

// validateMandatoryRequirements checks that user has active sessions for all mandatory services
// with required scopes. Returns error if any mandatory requirement is not satisfied.
// Optional requirements are ignored and never block authorization.
func (s *Service) validateMandatoryRequirements(
	ctx context.Context,
	principal string,
	agent *storage.Agent,
) error {
	// If agent has no service requirements, nothing to validate
	if len(agent.ServiceRequirements) == 0 {
		return nil
	}

	// Check each mandatory requirement
	for _, req := range agent.ServiceRequirements {
		// Skip optional requirements
		if !req.IsMandatory() {
			continue
		}

		// Verify user has active session for this service.
		// FindByPrincipalAndService returns (nil, nil) when no session exists — that is
		// not an error condition per the port contract; check nil separately.
		session, err := s.sessionRepo.FindByPrincipalAndService(ctx, id.Principal(principal), req.ServiceID)
		if err != nil {
			if s.logger != nil {
				s.logger.Warn(
					"MandatoryRequirementNotMet",
					"agent_id", agent.ID,
					"service_id", req.ServiceID,
					"reason", "session_lookup_error",
				)
			}
			return fmt.Errorf("session_required: user does not have required session for service %s", req.ServiceID)
		}
		if session == nil {
			if s.logger != nil {
				s.logger.Warn(
					"MandatoryRequirementNotMet",
					"agent_id", agent.ID,
					"service_id", req.ServiceID,
					"reason", "session_not_found",
				)
			}
			return fmt.Errorf("session_required: user does not have required session for service %s", req.ServiceID)
		}

		// Check if session is expired
		if session.IsExpired() {
			if s.logger != nil {
				s.logger.Warn(
					"MandatoryRequirementNotMet",
					"agent_id", agent.ID,
					"service_id", req.ServiceID,
					"reason", "session_expired",
				)
			}
			return fmt.Errorf("session_expired: user session has expired for service %s", req.ServiceID)
		}

		// Check if session scopes include all required scopes (superset check)
		if !s.hasRequiredScopes(session.Scope, req.RequiredScopes) {
			if s.logger != nil {
				s.logger.Warn(
					"MandatoryRequirementNotMet",
					"agent_id", agent.ID,
					"service_id", req.ServiceID,
					"reason", "scope_mismatch",
					"required_scopes", req.RequiredScopes,
					"actual_scopes", session.Scope,
				)
			}
			return fmt.Errorf("scope_mismatch: user session lacks required scopes for service %s", req.ServiceID)
		}
	}

	// All mandatory requirements satisfied
	return nil
}

// hasRequiredScopes checks if sessionScopes is a superset of requiredScopes (case-sensitive).
// Returns true if sessionScopes contains all scopes in requiredScopes, allowing for extra scopes.
// Returns true if requiredScopes is empty or nil (no requirements).
func (s *Service) hasRequiredScopes(sessionScopes []string, requiredScopes []string) bool {
	// If no required scopes, always pass
	if len(requiredScopes) == 0 {
		return true
	}

	// If required scopes exist but session has none, fail
	if len(sessionScopes) == 0 {
		return false
	}

	// Build map of session scopes for efficient lookup
	sessionScopeMap := make(map[string]bool)
	for _, scope := range sessionScopes {
		sessionScopeMap[scope] = true
	}

	// Check that all required scopes exist in session scopes
	for _, required := range requiredScopes {
		if !sessionScopeMap[required] {
			return false
		}
	}

	return true
}
