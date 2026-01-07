package oauth2

import (
	"context"
	"errors"
	"fmt"
	"net/url"

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
}

// Service implements the OAuth2Service port
type Service struct {
	agentRepo ports.AgentRepository
	grantRepo ports.UserGrantRepository
	config    *OAuth2Config
}

// NewService creates a new OAuth2Service implementation
func NewService(agentRepo ports.AgentRepository, grantRepo ports.UserGrantRepository, config *OAuth2Config) ports.OAuth2Service {
	return &Service{
		agentRepo: agentRepo,
		grantRepo: grantRepo,
		config:    config,
	}
}

// HandleAuthorization processes an OAuth2 authorization request
// Returns an AuthorizationDecision with either:
// - redirect_to_upstream: Valid client with active grant
// - redirect_to_consent: Valid client but no active grant
// - error: Invalid client or server error
func (s *Service) HandleAuthorization(ctx context.Context, req *ports.AuthorizationRequest, principal string) (*ports.AuthorizationDecision, error) {
	// Step 1: Validate client_id against registered agents
	agent, err := s.agentRepo.GetByClientID(ctx, req.ClientID)
	if err != nil {
		// Check if it's a not found error
		if storageErr, ok := err.(*storage.StorageError); ok && storageErr.Kind == storage.ErrorKindNotFound {
			// Invalid client_id - build error redirect
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
	grant, err := s.grantRepo.FindByPrincipalAndAgent(ctx, principal, agent.ID)
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

	// Active grant exists - redirect to upstream OAuth2 server
	upstreamURL := s.buildUpstreamAuthorizeURL(req)
	return &ports.AuthorizationDecision{
		Action:      "redirect_to_upstream",
		RedirectURL: upstreamURL,
	}, nil
}

// buildUpstreamAuthorizeURL constructs the upstream authorization endpoint URL
// preserving all OAuth2 parameters
func (s *Service) buildUpstreamAuthorizeURL(req *ports.AuthorizationRequest) string {
	u, _ := url.Parse(s.config.UpstreamAuthorizeEndpoint)
	q := u.Query()

	// Add required parameters
	q.Set("client_id", req.ClientID)
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
