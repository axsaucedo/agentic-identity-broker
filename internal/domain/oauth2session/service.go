package oauth2session

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwe"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"golang.org/x/oauth2"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// OAuth2SessionService orchestrates OAuth2 authorization flows and session management.
type OAuth2SessionService struct {
	serviceRepo ports.ThirdpartyOAuth2ServiceRepository
	sessionRepo ports.UserSessionRepository
	grantRepo   ports.UserGrantRepository // For dependent agents
	agentRepo   ports.AgentRepository     // For agent display names
	encryption  ports.EncryptionPort
	jweKey      jwk.Key
	config      Config
	logger      *slog.Logger
}

// Config holds configuration for the OAuth2 session service.
type Config struct {
	CallbackBaseURL    string        // e.g., "https://broker.example.com"
	StateTokenTTL      time.Duration // Default: 10 minutes
	PKCEVerifierLength int           // Default: 32 bytes
	MaxRetries         int           // Default: 3
	RetryBaseDelay     time.Duration // Default: 1 second
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		StateTokenTTL:      10 * time.Minute,
		PKCEVerifierLength: 32,
		MaxRetries:         3,
		RetryBaseDelay:     time.Second,
	}
}

// NewConfigFromPorts builds OAuth2SessionService Config from application-wide configuration.
// This ensures the service uses configuration from the application's ConfigPort instead of hardcoded defaults.
// Constitution Principle VII (Configuration-Driven Design) compliance.
func NewConfigFromPorts(portsCfg ports.ThirdPartyOAuth2Config, callbackBaseURL string) Config {
	cfg := Config{
		CallbackBaseURL: callbackBaseURL,
		MaxRetries:      3,           // Not yet in ports config, use default
		RetryBaseDelay:  time.Second, // Not yet in ports config, use default
	}

	// Use configured values if provided, otherwise use defaults
	if portsCfg.StateTokenTTL > 0 {
		cfg.StateTokenTTL = portsCfg.StateTokenTTL
	} else {
		cfg.StateTokenTTL = 10 * time.Minute
	}

	if portsCfg.PKCEVerifierLength > 0 {
		cfg.PKCEVerifierLength = portsCfg.PKCEVerifierLength
	} else {
		cfg.PKCEVerifierLength = 32
	}

	return cfg
}

// NewOAuth2SessionService creates a new OAuth2SessionService.
func NewOAuth2SessionService(
	serviceRepo ports.ThirdpartyOAuth2ServiceRepository,
	sessionRepo ports.UserSessionRepository,
	grantRepo ports.UserGrantRepository,
	agentRepo ports.AgentRepository,
	encryption ports.EncryptionPort,
	jweKey jwk.Key,
	config Config,
	logger *slog.Logger,
) *OAuth2SessionService {
	if config.StateTokenTTL == 0 {
		config.StateTokenTTL = 10 * time.Minute
	}
	if config.PKCEVerifierLength == 0 {
		config.PKCEVerifierLength = 32
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryBaseDelay == 0 {
		config.RetryBaseDelay = time.Second
	}

	return &OAuth2SessionService{
		serviceRepo: serviceRepo,
		sessionRepo: sessionRepo,
		grantRepo:   grantRepo,
		agentRepo:   agentRepo,
		encryption:  encryption,
		jweKey:      jweKey,
		config:      config,
		logger:      logger,
	}
}

// GetCallbackBaseURL returns the configured callback base URL.
// This is used for validating OAuth2 redirect URIs and constructing callback endpoints.
// Constitution Principle VII: Configuration-Driven Design
func (s *OAuth2SessionService) GetCallbackBaseURL() string {
	return s.config.CallbackBaseURL
}

// InitiateFlowResult contains the data needed to redirect user to authorization.
type InitiateFlowResult struct {
	AuthorizationURL string // Full URL to redirect user to
	StateToken       string // JWE-encrypted state token for storage/form hidden field
}

// HandleCallbackRequest contains parameters from the OAuth2 callback.
type HandleCallbackRequest struct {
	ServiceID string // From URL path
	Code      string // Authorization code from query
	State     string // JWE state token from query
	Error     string // OAuth2 error code (optional)
	ErrorDesc string // OAuth2 error description (optional)
}

// HandleCallbackResult contains the result of processing an OAuth2 callback.
type HandleCallbackResult struct {
	Session     *storage.UserSession
	RedirectURI string // Original redirect_uri from state token claims
}

// AgentInfo represents an agent with display information.
type AgentInfo struct {
	ID          string `json:"id"`           // Agent's unique identifier
	DisplayName string `json:"display_name"` // Agent's display name for UI presentation
}

// SessionWithAgents represents a session with information about dependent agents.
type SessionWithAgents struct {
	Session         *storage.UserSession
	DependentAgents []AgentInfo // Agents that use this session with display names
}

// =============================================================================
// GROUP 1: Crypto Foundations
// =============================================================================

// CreateStateToken encrypts the state token claims into a JWE string.
func (s *OAuth2SessionService) CreateStateToken(claims *OAuth2StateTokenClaims) (string, error) {
	// Validate claims
	if err := claims.Validate(); err != nil {
		return "", fmt.Errorf("invalid state token claims: %w", err)
	}

	// Marshal claims to JSON
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Encrypt JSON using JWE with symmetric key (A256GCMKW for key wrap + A256GCM for content)
	encrypted, err := jwe.Encrypt(payload,
		jwe.WithKey(jwa.A256GCMKW(), s.jweKey),
		jwe.WithContentEncryption(jwa.A256GCM()),
		jwe.WithCompact())
	if err != nil {
		return "", fmt.Errorf("failed to encrypt state token: %w", err)
	}

	return string(encrypted), nil
}

// ValidateStateToken decrypts and validates a state token.
func (s *OAuth2SessionService) ValidateStateToken(
	tokenString string,
	currentPrincipal string,
	expectedServiceID string,
) (*OAuth2StateTokenClaims, error) {
	// Decrypt JWE
	decrypted, err := jwe.Decrypt([]byte(tokenString), jwe.WithKey(jwa.A256GCMKW(), s.jweKey))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt state token: %w", ErrInvalidStateToken)
	}

	// Unmarshal JSON to claims
	var claims OAuth2StateTokenClaims
	if err := json.Unmarshal(decrypted, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state token claims: %w", err)
	}

	// Validate claims structure
	if err := claims.Validate(); err != nil {
		return nil, fmt.Errorf("invalid state token claims: %w", err)
	}

	// Check expiration
	if claims.IsExpired() {
		// Audit log: state token expired
		s.logger.Warn("oauth2_state_token_expired",
			"event", "session.oauth2.state_expired",
			"principal", claims.Principal,
			"service_id", claims.ServiceID,
			"expired_at", claims.ExpiresAt.Unix(),
			"timestamp", time.Now().Unix())
		return nil, ErrStateTokenExpired
	}

	// Check principal match (CSRF protection)
	if claims.Principal != currentPrincipal {
		// Audit log: principal mismatch (CSRF attack detection)
		s.logger.Error("oauth2_state_validation_failed",
			"event", "session.oauth2.state_validation_failed",
			"reason", "principal_mismatch",
			"expected_principal", currentPrincipal,
			"actual_principal", claims.Principal,
			"service_id", expectedServiceID,
			"timestamp", time.Now().Unix())
		return nil, ErrPrincipalMismatch
	}

	// Check service ID match
	if claims.ServiceID != expectedServiceID {
		s.logger.Error("oauth2_state_validation_failed",
			"event", "session.oauth2.state_validation_failed",
			"reason", "service_id_mismatch",
			"expected_service_id", expectedServiceID,
			"actual_service_id", claims.ServiceID,
			"principal", currentPrincipal,
			"timestamp", time.Now().Unix())
		return nil, fmt.Errorf("state token validation failed: %w", ErrServiceIDMismatch)
	}

	return &claims, nil
}

// =============================================================================
// GROUP 2: OAuth2 Helpers
// =============================================================================

// buildOAuth2Config creates an oauth2.Config from third-party service details.
func (s *OAuth2SessionService) buildOAuth2Config(
	service *storage.ThirdpartyOAuth2Service,
	callbackURL string,
) *oauth2.Config {
	// Extract scopes from service
	scopes := make([]string, 0, len(service.Scopes))
	for _, scope := range service.Scopes {
		scopes = append(scopes, scope.ScopeValue)
	}

	// Create OAuth2 config
	return &oauth2.Config{
		ClientID:     service.ClientID,
		ClientSecret: service.ClientSecret,
		RedirectURL:  callbackURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  service.Endpoints.AuthorizeEndpoint,
			TokenURL: service.Endpoints.TokenEndpoint,
		},
	}
}

// exchangeCodeWithRetry exchanges authorization code for tokens with exponential backoff retry.
func (s *OAuth2SessionService) exchangeCodeWithRetry(
	ctx context.Context,
	config *oauth2.Config,
	code string,
	verifier string,
) (*oauth2.Token, error) {
	var lastErr error

	for attempt := 0; attempt < s.config.MaxRetries; attempt++ {
		// Try to exchange code for token
		token, err := config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err == nil {
			return token, nil
		}

		lastErr = err

		// If this was the last attempt, break
		if attempt == s.config.MaxRetries-1 {
			break
		}

		// Calculate exponential backoff delay
		delay := s.config.RetryBaseDelay * time.Duration(math.Pow(2, float64(attempt)))
		s.logger.Warn("token exchange failed, retrying",
			"attempt", attempt+1,
			"delay", delay,
			"error", err)

		// Sleep with context cancellation support
		select {
		case <-time.After(delay):
			// Continue to next retry
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("token exchange failed after %d attempts: %w", s.config.MaxRetries, lastErr)
}

// =============================================================================
// GROUP 3: Service Methods
// =============================================================================

// InitiateOAuth2Flow starts the OAuth2 authorization code flow.
// Creates PKCE verifier/challenge, generates JWE state token, returns auth URL.
func (s *OAuth2SessionService) InitiateOAuth2Flow(
	ctx context.Context,
	principal string,
	serviceID string,
	redirectURI string,
) (*InitiateFlowResult, error) {
	s.logger.Info("initiating OAuth2 flow", "principal", principal, "service_id", serviceID)

	// Fetch the service
	service, err := s.serviceRepo.Get(ctx, serviceID)
	if err != nil {
		s.logger.Error("service not found", "service_id", serviceID, "err", err)
		return nil, fmt.Errorf("failed to initiate OAuth2 flow: %w", ErrServiceNotFound)
	}

	// Generate PKCE
	verifier := oauth2.GenerateVerifier()

	// Create state token claims
	now := time.Now()
	claims := &OAuth2StateTokenClaims{
		Principal:    principal,
		ServiceID:    serviceID,
		PKCEVerifier: verifier,
		RedirectURI:  redirectURI,
		IssuedAt:     now,
		ExpiresAt:    now.Add(s.config.StateTokenTTL),
	}

	// Encrypt state token
	stateToken, err := s.CreateStateToken(claims)
	if err != nil {
		s.logger.Error("failed to create state token", "err", err)
		return nil, fmt.Errorf("failed to create state token: %w", err)
	}

	// Build OAuth2 config with callback URL
	callbackURL := s.config.CallbackBaseURL + "/api/third-party/" + serviceID + "/oauth2/callback"
	cfg := s.buildOAuth2Config(service, callbackURL)

	// Generate authorization URL with PKCE
	authURL := cfg.AuthCodeURL(stateToken, oauth2.S256ChallengeOption(verifier))

	// Audit log: OAuth2 flow initiated successfully
	s.logger.Info("oauth2_flow_initiated",
		"event", "session.oauth2.flow_initiated",
		"principal", principal,
		"service_id", serviceID,
		"timestamp", now.Unix())

	return &InitiateFlowResult{
		AuthorizationURL: authURL,
		StateToken:       stateToken,
	}, nil
}

// createSession creates and stores a new user session from OAuth2 tokens.
// Uses upsert semantics: if session exists for (principal, service_id), updates tokens while preserving CreatedAt.
func (s *OAuth2SessionService) createSession(
	ctx context.Context,
	principal string,
	serviceID string,
	token *oauth2.Token,
	scope []string,
) (*storage.UserSession, error) {
	// Check if session already exists for this principal+service
	existingSession, err := s.sessionRepo.FindByPrincipalAndService(ctx, principal, serviceID)
	if err != nil {
		s.logger.Error("failed to query existing session", "principal", principal, "service_id", serviceID, "err", err)
		return nil, fmt.Errorf("failed to query existing session: %w", err)
	}

	// Encrypt access token
	encryptedAccess, err := s.encryption.Encrypt(ctx, []byte(token.AccessToken), nil)
	if err != nil {
		s.logger.Error("failed to encrypt access token", "err", err)
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}

	// Encrypt refresh token if present
	var encryptedRefresh []byte
	if token.RefreshToken != "" {
		encryptedRefresh, err = s.encryption.Encrypt(ctx, []byte(token.RefreshToken), nil)
		if err != nil {
			s.logger.Error("failed to encrypt refresh token", "err", err)
			return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	}

	// Determine access token expiration
	var accessTokenExpiresAt *time.Time
	if !token.Expiry.IsZero() {
		accessTokenExpiresAt = &token.Expiry
	}

	// Determine token type (default to Bearer if not specified)
	tokenType := token.TokenType
	if tokenType == "" {
		tokenType = "Bearer"
	}

	// Create or update session
	now := time.Now()
	var sessionID string
	var createdAt time.Time
	var initiatedAt time.Time

	if existingSession != nil {
		// Update existing session: preserve ID and CreatedAt
		sessionID = existingSession.ID
		createdAt = existingSession.CreatedAt
		initiatedAt = existingSession.InitiatedAt
	} else {
		// Create new session
		sessionID = uuid.New().String()
		createdAt = now
		initiatedAt = now
	}

	session := &storage.UserSession{
		ID:                    sessionID,
		Principal:             principal,
		ServiceID:             serviceID,
		EncryptedAccessToken:  encryptedAccess,
		EncryptedRefreshToken: encryptedRefresh,
		TokenType:             tokenType,
		Scope:                 scope,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		InitiatedAt:           initiatedAt,
		CreatedAt:             createdAt,
		UpdatedAt:             now,
	}

	// Store session (will upsert if already exists)
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error("failed to create session", "principal", principal, "service_id", serviceID, "err", err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info("session created", "session_id", sessionID, "principal", principal, "service_id", serviceID)
	return session, nil
}

// HandleCallback processes the OAuth2 callback, exchanges code for tokens, stores session.
func (s *OAuth2SessionService) HandleCallback(
	ctx context.Context,
	principal string,
	req *HandleCallbackRequest,
) (*HandleCallbackResult, error) {
	s.logger.Debug("processing OAuth2 callback", "service_id", req.ServiceID, "principal", principal)

	// Check for OAuth2 error response
	if req.Error != "" {
		if req.ErrorDesc != "" {
			s.logger.Warn("OAuth2 authorization failed", "error", req.Error, "description", req.ErrorDesc)
			return nil, fmt.Errorf("OAuth2 authorization failed: %s: %s", req.Error, req.ErrorDesc)
		}
		s.logger.Warn("OAuth2 authorization failed", "error", req.Error)
		return nil, fmt.Errorf("OAuth2 authorization failed: %s", req.Error)
	}

	// Validate state token (checks expiration, principal mismatch, tampering)
	claims, err := s.ValidateStateToken(req.State, principal, req.ServiceID)
	if err != nil {
		s.logger.Error("state token validation failed", "err", err)
		return nil, fmt.Errorf("state token validation failed: %w", err)
	}

	// Fetch the service
	service, err := s.serviceRepo.Get(ctx, req.ServiceID)
	if err != nil {
		s.logger.Error("service not found during callback", "service_id", req.ServiceID, "err", err)
		return nil, fmt.Errorf("service not found: %w", err)
	}

	// Build OAuth2 config
	callbackURL := s.config.CallbackBaseURL + "/api/third-party/" + req.ServiceID + "/oauth2/callback"
	cfg := s.buildOAuth2Config(service, callbackURL)

	// Exchange authorization code for tokens (with retry)
	s.logger.Info("exchanging authorization code for token",
		"service_id", req.ServiceID,
		"callback_url", callbackURL,
		"token_endpoint", cfg.Endpoint.TokenURL,
		"client_id", cfg.ClientID)
	token, err := s.exchangeCodeWithRetry(ctx, cfg, req.Code, claims.PKCEVerifier)
	if err != nil {
		// Audit log: PKCE validation failure (token exchange failure typically indicates PKCE error)
		s.logger.Error("oauth2_pkce_validation_failed",
			"event", "session.oauth2.pkce_validation_failed",
			"principal", principal,
			"service_id", req.ServiceID,
			"error", err.Error(),
			"timestamp", time.Now().Unix())
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Extract scopes from token response or fall back to service scopes
	scopes := make([]string, 0)
	if scopeVal := token.Extra("scope"); scopeVal != nil {
		if scopeStr, ok := scopeVal.(string); ok {
			scopes = strings.Split(scopeStr, " ")
		}
	}
	if len(scopes) == 0 {
		// Fall back to service scopes
		for _, scope := range service.Scopes {
			scopes = append(scopes, scope.ScopeValue)
		}
	}

	// Create and store session
	session, err := s.createSession(ctx, principal, req.ServiceID, token, scopes)
	if err != nil {
		s.logger.Error("failed to create session from token", "err", err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Audit log: OAuth2 session established successfully
	s.logger.Info("oauth2_session_established",
		"event", "session.oauth2.session_established",
		"principal", principal,
		"service_id", req.ServiceID,
		"session_id", session.ID,
		"timestamp", time.Now().Unix())

	return &HandleCallbackResult{
		Session:     session,
		RedirectURI: claims.RedirectURI,
	}, nil
}

// ListUserSessions returns all sessions for a principal with summary info.
func (s *OAuth2SessionService) ListUserSessions(
	ctx context.Context,
	principal string,
) ([]*storage.UserSessionSummary, error) {
	// Fetch all sessions for principal
	sessions, err := s.sessionRepo.ListByPrincipal(ctx, principal)
	if err != nil {
		s.logger.Error("failed to list sessions", "principal", principal, "err", err)
		return nil, err
	}

	// Convert to summaries with agent counts
	summaries := make([]*storage.UserSessionSummary, 0, len(sessions))
	for _, session := range sessions {
		// Fetch service details
		service, err := s.serviceRepo.Get(ctx, session.ServiceID)
		if err != nil {
			s.logger.Warn("service not found", "service_id", session.ServiceID, "err", err)
			continue
		}

		// Count dependent agents
		agentCount, err := s.grantRepo.CountAgentsByServiceID(ctx, session.ServiceID)
		if err != nil {
			s.logger.Warn("failed to count agents", "service_id", session.ServiceID, "err", err)
			agentCount = 0
		}

		summary := storage.NewUserSessionSummary(session, service.DisplayName, agentCount)
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// TerminateSession deletes a session and its encrypted tokens.
func (s *OAuth2SessionService) TerminateSession(
	ctx context.Context,
	principal string,
	serviceID string,
) error {
	// Audit log: Session termination initiated
	s.logger.Info("oauth2_session_termination_initiated",
		"event", "session.oauth2.termination_initiated",
		"principal", principal,
		"service_id", serviceID,
		"timestamp", time.Now().Unix())

	// Step 1: Fetch session to verify ownership
	session, err := s.sessionRepo.FindByPrincipalAndService(ctx, principal, serviceID)
	if err != nil {
		// Audit log: Session not found
		s.logger.Warn("oauth2_session_termination_failed",
			"event", "session.oauth2.session_not_found",
			"principal", principal,
			"service_id", serviceID,
			"reason", "session_not_found",
			"error", err.Error(),
			"timestamp", time.Now().Unix())
		return fmt.Errorf("session not found: %w", ErrSessionNotFound)
	}

	if session == nil {
		// Audit log: Session not found (nil case)
		s.logger.Warn("oauth2_session_termination_failed",
			"event", "session.oauth2.session_not_found",
			"principal", principal,
			"service_id", serviceID,
			"reason", "session_not_found",
			"timestamp", time.Now().Unix())
		return ErrSessionNotFound
	}

	// Step 2: Verify principal ownership (authorization check)
	if session.Principal != principal {
		// Audit log: Unauthorized termination attempt (SECURITY INCIDENT)
		s.logger.Error("oauth2_session_termination_failed",
			"event", "session.oauth2.termination_unauthorized",
			"principal", principal,
			"expected_principal", session.Principal,
			"service_id", serviceID,
			"session_id", session.ID,
			"reason", "principal_mismatch",
			"timestamp", time.Now().Unix())
		return fmt.Errorf("termination denied: %w", ErrUnauthorized)
	}

	// Step 3: Delete session from repository
	err = s.sessionRepo.DeleteByPrincipalAndService(ctx, principal, serviceID)
	if err != nil {
		// Audit log: Repository error during termination
		s.logger.Error("oauth2_session_termination_failed",
			"event", "session.oauth2.termination_repository_error",
			"principal", principal,
			"service_id", serviceID,
			"session_id", session.ID,
			"reason", "repository_error",
			"error", err.Error(),
			"timestamp", time.Now().Unix())
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Audit log: Session terminated successfully
	s.logger.Info("oauth2_session_terminated",
		"event", "session.oauth2.session_terminated",
		"principal", principal,
		"service_id", serviceID,
		"session_id", session.ID,
		"initiated_at", session.InitiatedAt,
		"timestamp", time.Now().Unix())

	return nil
}

// GetSessionWithAgents returns session details including list of dependent agents.
func (s *OAuth2SessionService) GetSessionWithAgents(
	ctx context.Context,
	principal string,
	serviceID string,
) (*SessionWithAgents, error) {
	s.logger.Debug("retrieving session with agents",
		"principal", principal,
		"service_id", serviceID)

	// Step 1: Fetch session to verify it exists and ownership
	session, err := s.sessionRepo.FindByPrincipalAndService(ctx, principal, serviceID)
	if err != nil {
		s.logger.Error("failed to fetch session", "principal", principal, "service_id", serviceID, "err", err)
		return nil, fmt.Errorf("session details retrieval failed: %w", ErrSessionNotFound)
	}

	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Step 2: Verify principal ownership (authorization)
	if session.Principal != principal {
		s.logger.Error("principal mismatch in GetSessionWithAgents",
			"expected_principal", principal,
			"session_principal", session.Principal,
			"service_id", serviceID)
		return nil, fmt.Errorf("session details access denied: %w", ErrUnauthorized)
	}

	// Step 3: Query dependent agent IDs using grant repository
	// Get list of actual agent IDs that have delegated tokens for this service
	agentIDs, err := s.grantRepo.ListByServiceID(ctx, serviceID)
	if err != nil {
		s.logger.Warn("failed to list dependent agent IDs", "service_id", serviceID, "err", err)
		agentIDs = []string{} // Return empty list on error
	}

	// Step 4: Fetch agent display names for each dependent agent
	dependentAgents := make([]AgentInfo, 0, len(agentIDs))
	for _, agentID := range agentIDs {
		agent, err := s.agentRepo.Get(ctx, agentID)
		if err != nil {
			s.logger.Warn("failed to fetch agent details", "agent_id", agentID, "err", err)
			// Fall back to using agent ID if agent not found
			dependentAgents = append(dependentAgents, AgentInfo{
				ID:          agentID,
				DisplayName: agentID, // Use ID as fallback
			})
			continue
		}

		dependentAgents = append(dependentAgents, AgentInfo{
			ID:          agent.ID,
			DisplayName: agent.DisplayName,
		})
	}

	s.logger.Debug("retrieved session with agents",
		"principal", principal,
		"service_id", serviceID,
		"dependent_agent_count", len(dependentAgents))

	return &SessionWithAgents{
		Session:         session,
		DependentAgents: dependentAgents,
	}, nil
}
