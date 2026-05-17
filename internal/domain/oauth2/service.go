package oauth2

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwe"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/cimd"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/servermode"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

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

	// Mode indicates whether the broker operates in "proxy", "local", or "hybrid" mode.
	// In local mode, JWKS and code_challenge_methods are included in metadata.
	Mode servermode.Mode

	// CIMDEnabled indicates whether CIMD-based client_id resolution is enabled.
	// When true, client_id_metadata_document_supported is advertised in metadata.
	CIMDEnabled bool

	// TokenExchangeEnabled indicates whether the RFC 8693 token exchange service is wired.
	// When true, the token-exchange grant type is appended to GrantTypesSupported in metadata.
	TokenExchangeEnabled bool

	// ModeStrategy enforces which agent client modes are permitted in this server mode.
	// When non-nil, every authorization request is checked against the strategy.
	ModeStrategy ModeStrategy
}

// Service implements the OAuth2Service port
type Service struct {
	grantRepo       ports.UserGrantRepository
	sessionRepo     ports.UserSessionRepository
	clientResolver  ports.ClientResolver
	jweTokenService *jwe.TokenService
	config          *OAuth2Config
	logger          *slog.Logger
}

// NewService creates a new OAuth2Service implementation
func NewService(agentRepo ports.AgentRepository, grantRepo ports.UserGrantRepository, config *OAuth2Config) ports.OAuth2Service {
	return &Service{
		grantRepo:      grantRepo,
		clientResolver: NewOpaqueClientResolver(agentRepo),
		config:         config,
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
		grantRepo:      grantRepo,
		sessionRepo:    sessionRepo,
		clientResolver: NewOpaqueClientResolver(agentRepo),
		config:         config,
		logger:         logger,
	}
}

// NewServiceWithClientResolver creates a new OAuth2Service with an explicit ClientResolver strategy.
// Used when CIMD support is enabled (cimd.enabled: true) or when a custom resolver is required.
// Returns *Service so callers can chain WithJWETokenService before assigning to the port interface.
func NewServiceWithClientResolver(
	grantRepo ports.UserGrantRepository,
	sessionRepo ports.UserSessionRepository,
	clientResolver ports.ClientResolver,
	config *OAuth2Config,
	logger *slog.Logger,
) *Service {
	return &Service{
		grantRepo:      grantRepo,
		sessionRepo:    sessionRepo,
		clientResolver: clientResolver,
		config:         config,
		logger:         logger,
	}
}

// WithJWETokenService sets the JWE token service on the service.
// Required for CIMD consent flows that use stateless JWE session tokens.
func (s *Service) WithJWETokenService(ts *jwe.TokenService) *Service {
	s.jweTokenService = ts
	return s
}

// ResolveForTokenGrant resolves the client_id, classifies the agent, and enforces
// mode boundaries for the token endpoint. Follows the same universal resolution as
// HandleAuthorization (FR-003, FR-004, FR-005) but without authorization-specific
// logic (redirect URI, scopes, consent).
func (s *Service) ResolveForTokenGrant(ctx context.Context, rawClientID string) (*ports.TokenGrantResolution, error) {
	resolution, resolveErr := s.clientResolver.ResolveClient(ctx, id.ClientID(rawClientID))
	if resolveErr != nil {
		var clientErr *ports.ClientIDError
		if errors.As(resolveErr, &clientErr) {
			return nil, clientErr
		}
		if s.logger != nil {
			s.logger.ErrorContext(ctx, "unexpected error from client resolver", "error", resolveErr)
		}
		return nil, &ports.ClientIDError{Code: "server_error", Desc: "client resolution failed"}
	}

	agent := resolution.Agent
	mode := agent.ClientMode()

	if s.config.ModeStrategy != nil && !s.config.ModeStrategy.AcceptsClientMode(mode) {
		return nil, &ports.ClientIDError{
			Code: "unauthorized_client",
			Desc: fmt.Sprintf("client mode not supported in %s mode", s.config.ModeStrategy.Mode()),
		}
	}

	return &ports.TokenGrantResolution{Agent: agent, ClientMode: mode}, nil
}

// HandleAuthorization processes an OAuth2 authorization request
// Returns an AuthorizationDecision with either:
// - proceed: Valid client with active grant — handler decides next step
// - redirect_to_consent: Valid client but no active grant
// - error: Invalid client or server error
func (s *Service) HandleAuthorization(ctx context.Context, req *ports.AuthorizationRequest, principal id.Principal) (*ports.AuthorizationDecision, error) {
	// Resolve the client via the injected strategy (OpaqueClientResolver or CIMDClientResolver).
	var agent *storage.Agent
	var cimdMeta *ports.CIMDMetadataDTO

	resolution, resolveErr := s.clientResolver.ResolveClient(ctx, req.ClientID)
	if resolveErr != nil {
		code := "invalid_client"
		desc := "Client not registered"
		var clientErr *ports.ClientIDError
		if errors.As(resolveErr, &clientErr) {
			code = clientErr.Code
			desc = clientErr.Desc
		}
		return &ports.AuthorizationDecision{
			Action:    "error",
			ErrorCode: code,
			ErrorDesc: desc,
		}, nil
	}
	agent = resolution.Agent
	cimdMeta = resolution.CIMDMetadata

	// Mode enforcement: reject agents whose ClientMode is not permitted in this server mode.
	if s.config.ModeStrategy != nil && !s.config.ModeStrategy.AcceptsClientMode(agent.ClientMode()) {
		return &ports.AuthorizationDecision{
			Action:    "error",
			ErrorCode: "unauthorized_client",
			ErrorDesc: fmt.Sprintf("client mode not supported in %s mode", s.config.ModeStrategy.Mode()),
		}, nil
	}

	// Step 1b: Validate redirect_uri.
	// For CIMD clients, validate against the document's redirect_uris.
	// For opaque clients, validate against the agent's registered redirect_uris.
	// Per RFC 6749 §4.1.2.1, MUST NOT redirect if redirect_uri is unverified.
	var allowedRedirectURIs []string
	if cimdMeta != nil {
		allowedRedirectURIs = cimdMeta.RedirectURIs
	} else {
		allowedRedirectURIs = agent.RedirectURIs
	}

	// CIMD clients: redirect_uri failures are invalid_request (the CIMD contract specifies this).
	// Opaque clients: use the standard invalid_redirect_uri error code.
	redirectURIErrCode := "invalid_redirect_uri"
	if cimdMeta != nil {
		redirectURIErrCode = "invalid_request"
	}

	if len(allowedRedirectURIs) == 0 {
		return &ports.AuthorizationDecision{
			Action:    "error",
			ErrorCode: redirectURIErrCode,
			ErrorDesc: "redirect_uri not registered for this client",
		}, nil
	}
	uriAllowed := false
	for _, allowed := range allowedRedirectURIs {
		if req.RedirectURI == allowed {
			uriAllowed = true
			break
		}
	}
	if !uriAllowed {
		return &ports.AuthorizationDecision{
			Action:    "error",
			ErrorCode: redirectURIErrCode,
			ErrorDesc: "redirect_uri not registered for this client",
		}, nil
	}

	// Step 1b-runtime: Enforce HTTPS for non-loopback hosts even on legacy data.
	// Write-time validation (Agent.Validate/ValidateForCreate) prevents new non-HTTPS
	// registrations, but this guard closes the gap for pre-existing stored URIs.
	if !storage.IsValidRedirectURI(req.RedirectURI) {
		return &ports.AuthorizationDecision{
			Action:    "error",
			ErrorCode: "invalid_redirect_uri",
			ErrorDesc: "redirect_uri must use HTTPS for non-local hosts",
		}, nil
	}

	// Step 1c: Validate requested scopes against agent's allowed scopes.
	// redirect_uri is validated above, so a redirect-with-error is now safe.
	if len(agent.AllowedScopes) > 0 && req.Scope != "" {
		allowedSet := make(map[string]bool, len(agent.AllowedScopes))
		for _, s := range agent.AllowedScopes {
			allowedSet[s] = true
		}
		for _, s := range strings.Fields(req.Scope) {
			if !allowedSet[s] {
				errRedirect, _ := BuildErrorRedirectURL(req.RedirectURI, req.State, "invalid_scope", "requested scope is not permitted")
				return &ports.AuthorizationDecision{
					Action:      "error",
					ErrorCode:   "invalid_scope",
					ErrorDesc:   "requested scope is not permitted",
					RedirectURL: errRedirect,
				}, nil
			}
		}
	}

	// Step 2: Check if user has active grant for this agent
	grant, err := s.grantRepo.FindByPrincipalAndAgent(ctx, principal, agent.ID)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		// redirect_uri is validated above so a redirect-with-error is safe here.
		errRedirect, _ := BuildErrorRedirectURL(req.RedirectURI, req.State, "server_error", "server error")
		return &ports.AuthorizationDecision{
			Action:      "error",
			ErrorCode:   "server_error",
			ErrorDesc:   "Failed to check grant",
			RedirectURL: errRedirect,
		}, nil
	}

	// Step 3: Determine action based on grant status
	if grant == nil || !grant.IsActive() {
		consentURL, buildErr := s.buildConsentURL(ctx, req, principal, agent, cimdMeta)
		if buildErr != nil {
			errRedirect, _ := BuildErrorRedirectURL(req.RedirectURI, req.State, "server_error", "server error")
			return &ports.AuthorizationDecision{
				Action:      "error",
				ErrorCode:   "server_error",
				ErrorDesc:   "Failed to initiate consent session",
				RedirectURL: errRedirect,
			}, nil
		}
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
		expired, err := s.anyDelegatedSessionExpired(ctx, principal, grant.DelegatedOAuth2Tokens)
		if err != nil {
			errRedirect, _ := BuildErrorRedirectURL(req.RedirectURI, req.State, "server_error", "server error")
			return &ports.AuthorizationDecision{
				Action:      "error",
				ErrorCode:   "server_error",
				ErrorDesc:   "Failed to check session status",
				RedirectURL: errRedirect,
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
			consentURL, buildErr := s.buildConsentURL(ctx, req, principal, agent, cimdMeta)
			if buildErr != nil {
				errRedirect, _ := BuildErrorRedirectURL(req.RedirectURI, req.State, "server_error", "server error")
				return &ports.AuthorizationDecision{
					Action:      "error",
					ErrorCode:   "server_error",
					ErrorDesc:   "Failed to initiate consent session",
					RedirectURL: errRedirect,
				}, nil
			}
			return &ports.AuthorizationDecision{
				Action:      "redirect_to_consent",
				RedirectURL: consentURL,
			}, nil
		}
	}

	// Step 5: Validate mandatory service requirements (if session repo available)
	if s.sessionRepo != nil && len(agent.ServiceRequirements) > 0 {
		err := s.validateMandatoryRequirements(ctx, principal.String(), agent)
		if err != nil {
			var storageErr *storage.StorageError
			if errors.As(err, &storageErr) {
				if s.logger != nil {
					s.logger.Error(
						"MandatoryRequirementStorageFailure",
						"agent_id", agent.ID,
						"error", err.Error(),
					)
				}
				errRedirect, _ := BuildErrorRedirectURL(req.RedirectURI, req.State, "server_error", "server error")
				return &ports.AuthorizationDecision{
					Action:      "error",
					ErrorCode:   "server_error",
					ErrorDesc:   "Failed to validate service requirements",
					RedirectURL: errRedirect,
				}, nil
			}
			if s.logger != nil {
				s.logger.Warn(
					"MandatoryRequirementNotMet",
					"agent_id", agent.ID,
					"error", err.Error(),
				)
			}
			consentURL, buildErr := s.buildConsentURL(ctx, req, principal, agent, cimdMeta)
			if buildErr != nil {
				errRedirect, _ := BuildErrorRedirectURL(req.RedirectURI, req.State, "server_error", "server error")
				return &ports.AuthorizationDecision{
					Action:      "error",
					ErrorCode:   "server_error",
					ErrorDesc:   "Failed to initiate consent session",
					RedirectURL: errRedirect,
				}, nil
			}
			return &ports.AuthorizationDecision{
				Action:      "redirect_to_consent",
				RedirectURL: consentURL,
			}, nil
		}
	}

	// Active grant exists and all mandatory requirements satisfied — proceed.
	// For ProxyClient agents with an upstream endpoint, build the redirect URL.
	// For CIMDClient/LocalClient agents (or when no upstream is configured), leave empty
	// so the proceed strategy issues a local authorization code.
	var upstreamURL string
	if s.config.UpstreamAuthorizeEndpoint != "" && agent.ClientID != nil {
		var urlErr error
		upstreamURL, urlErr = s.buildUpstreamAuthorizeURL(req, agent)
		if urlErr != nil {
			errRedirect, _ := BuildErrorRedirectURL(req.RedirectURI, req.State, "server_error", "server error")
			return &ports.AuthorizationDecision{
				Action:      "error",
				ErrorCode:   "server_error",
				ErrorDesc:   "Failed to build upstream authorize URL",
				RedirectURL: errRedirect,
			}, nil
		}
	}
	return &ports.AuthorizationDecision{
		Action:      "proceed",
		RedirectURL: upstreamURL,
		ClientMode:  agent.ClientMode(),
	}, nil
}

// buildUpstreamAuthorizeURL constructs the upstream authorization endpoint URL
// preserving all OAuth2 parameters.
//
// Feature 021: uses agent.ClientID (upstream OAuth2 client ID) instead of req.ClientID
// (which is now the broker's internal agent UUID). When MultiAgentClient.Enabled,
// appends the agent's internal UUID as the configured AgentIDParamName query parameter.
func (s *Service) buildUpstreamAuthorizeURL(req *ports.AuthorizationRequest, agent *storage.Agent) (string, error) {
	u, err := url.Parse(s.config.UpstreamAuthorizeEndpoint)
	if err != nil {
		return "", fmt.Errorf("invalid upstream authorize endpoint URL: %w", err)
	}
	q := u.Query()

	// Use agent.ClientID as upstream client_id (NOT the broker's internal agent UUID)
	if agent.ClientID == nil {
		return "", fmt.Errorf("agent %s has no upstream client_id configured", agent.ID)
	}
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
	return u.String(), nil
}

// buildConsentURL builds the consent redirect URL for a given agent and request.
// For CIMD flows (cimdMeta != nil), seals an AuthorizationSessionClaims JWE and returns
// a URL with ?session_token=<jwe>. For opaque flows, falls back to ?redirect_uri=<OriginalURL>.
func (s *Service) buildConsentURL(_ context.Context, req *ports.AuthorizationRequest, principal id.Principal, agent *storage.Agent, cimdMeta *ports.CIMDMetadataDTO) (string, error) {
	if cimdMeta != nil {
		meta := &cimd.ClientIDMetadataDocument{
			ClientID:     cimdMeta.ClientID,
			ClientName:   cimdMeta.ClientName,
			LogoURI:      cimdMeta.LogoURI,
			RedirectURIs: cimdMeta.RedirectURIs,
		}
		claims, err := NewAuthorizationSessionClaims(
			agent.ID,
			principal,
			req.OriginalURL,
			meta,
		)
		if err != nil {
			return "", fmt.Errorf("failed to create authorization session claims: %w", err)
		}
		token, err := s.CreateAuthorizationSessionToken(claims)
		if err != nil {
			return "", fmt.Errorf("failed to create authorization session token: %w", err)
		}
		return fmt.Sprintf("%s/consent/agent/%s?session_token=%s", s.config.PublicURL, agent.ID, url.QueryEscape(token)), nil
	}
	return fmt.Sprintf("%s/consent/agent/%s?redirect_uri=%s",
		s.config.PublicURL,
		agent.ID,
		url.QueryEscape(req.OriginalURL),
	), nil
}

// GenerateMetadata returns RFC 8414 OAuth2 metadata for this broker.
// In local mode, includes JWKS URI and code_challenge_methods.
func (s *Service) GenerateMetadata(ctx context.Context) (*ports.MetadataResponse, error) {
	issuer := s.config.PublicURL

	const tokenExchangeGrant = "urn:ietf:params:oauth:grant-type:token-exchange"
	var grantTypes []string
	if s.config.TokenExchangeEnabled {
		if !slices.Contains(s.config.SupportedGrantTypes, tokenExchangeGrant) {
			grantTypes = append(slices.Clone(s.config.SupportedGrantTypes), tokenExchangeGrant)
		} else {
			grantTypes = s.config.SupportedGrantTypes
		}
	} else {
		grantTypes = slices.DeleteFunc(slices.Clone(s.config.SupportedGrantTypes), func(g string) bool {
			return g == tokenExchangeGrant
		})
		if len(grantTypes) == 0 {
			switch s.config.Mode {
			case servermode.Local, servermode.Hybrid:
				grantTypes = []string{"authorization_code", "client_credentials"}
			default:
				grantTypes = []string{"authorization_code"}
			}
		}
	}

	metadata := &ports.MetadataResponse{
		Issuer:                            issuer,
		AuthorizationEndpoint:             fmt.Sprintf("%s/oauth2/authorize", issuer),
		TokenEndpoint:                     fmt.Sprintf("%s/oauth2/token", issuer),
		ResponseTypesSupported:            s.config.SupportedResponseTypes,
		GrantTypesSupported:               grantTypes,
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post"},
	}

	// In local and hybrid modes, include JWKS URI and code challenge methods (both serve the JWKS endpoint).
	if s.config.Mode == servermode.Local || s.config.Mode == servermode.Hybrid {
		metadata.JWKSURI = fmt.Sprintf("%s/oauth2/jwks.json", issuer)
		metadata.CodeChallengeMethodsSupported = []string{"S256"}
		// client_secret_post is always supported for confidential clients.
		// none is included when CIMD is enabled: CIMD agents are public clients with no pre-registered secret.
		methods := []string{"client_secret_post"}
		if s.config.CIMDEnabled {
			methods = append([]string{"none"}, methods...)
		}
		metadata.TokenEndpointAuthMethodsSupported = methods
	}

	if s.config.CIMDEnabled {
		t := true
		metadata.ClientIDMetadataDocumentSupported = &t
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
			return fmt.Errorf("failed to check session for service %s: %w", req.ServiceID, err)
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
