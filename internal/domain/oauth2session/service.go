package oauth2session

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
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
	httpClient  *http.Client // For upstream OAuth2 token endpoint calls
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
	httpClient *http.Client,
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
		httpClient:  httpClient,
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
// The token must be a valid JWE compact serialization with A256GCMKW key wrapping
// and A256GCM content encryption. Any tampering with the token will cause decryption
// to fail due to GCM authentication tag verification.
func (s *OAuth2SessionService) ValidateStateToken(
	tokenString string,
	currentPrincipal string,
	expectedServiceID string,
) (*OAuth2StateTokenClaims, error) {
	// Input validation
	if tokenString == "" {
		return nil, fmt.Errorf("state token is empty: %w", ErrInvalidStateToken)
	}

	// Decrypt JWE with explicit content encryption algorithm for enhanced security
	// Specifying the algorithm during decryption provides defense-in-depth:
	// - Validates algorithm match
	// - Prevents algorithm confusion attacks
	// - Ensures consistent decryption behavior
	decrypted, err := jwe.Decrypt(
		[]byte(tokenString),
		jwe.WithKey(jwa.A256GCMKW(), s.jweKey),
	)
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

	var encryptedAccess []byte
	var encryptedRefresh []byte

	// Create encryption context for this session bound to service (cryptographic service isolation)
	// This ensures tokens encrypted for one service cannot be decrypted with another service's context
	encryptionContext := map[string]string{
		"service_id": serviceID,
	}

	// Encrypt access token
	encryptedAccess, err = s.encryption.Encrypt(ctx, []byte(token.AccessToken), encryptionContext)
	if err != nil {
		s.logger.Error("failed to encrypt access token", "err", err)
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}

	// Encrypt refresh token if present
	if token.RefreshToken != "" {
		encryptedRefresh, err = s.encryption.Encrypt(ctx, []byte(token.RefreshToken), encryptionContext)
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

// RefreshAccessToken calls the upstream OAuth2 service's token endpoint to refresh an expired access token.
// Uses the provided refresh token to obtain a new access token from the service.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - service: The ThirdpartyOAuth2Service configuration containing token endpoint and credentials
//   - refreshToken: The valid refresh token from the stored session
//
// Returns:
//   - *oauth2.Token with new access_token, optional refresh_token, and expiry
//   - error if the refresh request fails (network error, invalid response, or upstream error)
//
// Per RFC 6749 Section 6, sends a POST request to the token endpoint with:
//   - grant_type=refresh_token
//   - refresh_token=<the provided refresh token>
//   - client_id=<from service config>
//   - client_secret=<from service config>
func (s *OAuth2SessionService) RefreshAccessToken(
	ctx context.Context,
	service *storage.ThirdpartyOAuth2Service,
	refreshToken string,
) (*oauth2.Token, error) {
	if service == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}

	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token cannot be empty")
	}

	// Prepare refresh token request per RFC 6749 Section 6
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", service.ClientID)
	data.Set("client_secret", service.ClientSecret)

	// Create POST request to token endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, service.Endpoints.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token request: %w", err)
	}

	// Set standard OAuth2 headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call upstream token endpoint for refresh: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Decode response
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
		Scope        string `json:"scope,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode upstream token response: %w", err)
	}

	// Check for HTTP error status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upstream token endpoint returned error status %d", resp.StatusCode)
	}

	// Validate required fields in response
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("upstream token response missing access_token")
	}

	// Determine token expiry
	var expiry time.Time
	if tokenResp.ExpiresIn > 0 {
		expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	// Build oauth2.Token
	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
		Expiry:       expiry,
	}

	s.logger.Info("access token refreshed",
		"service_id", service.ID,
		"token_endpoint", service.Endpoints.TokenEndpoint)

	return token, nil
}

// UpdateSessionTokens updates an existing session with refreshed tokens.
// Encrypts tokens using encryption context binding and persists the updated session.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - principal: The user principal (for encryption context)
//   - session: The UserSession to update (will be modified in-place)
//   - newToken: The new oauth2.Token from refresh operation
//
// Returns:
//   - error if encryption or persistence fails
//
// The session is modified in-place and persisted with upsert semantics.
// Encryption context binds tokens to principal/service/session for additional security.
func (s *OAuth2SessionService) UpdateSessionTokens(
	ctx context.Context,
	principal string,
	session *storage.UserSession,
	newToken *oauth2.Token,
) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if newToken == nil {
		return fmt.Errorf("new token cannot be nil")
	}

	// Build encryption context for this session
	encContext := map[string]string{
		"principal":  principal,
		"service_id": session.ServiceID,
		"session_id": session.ID,
	}

	// Encrypt new access token
	encryptedAccess, err := s.encryption.Encrypt(ctx, []byte(newToken.AccessToken), encContext)
	if err != nil {
		return fmt.Errorf("failed to encrypt refreshed access token: %w", err)
	}

	// Update session with new access token
	session.EncryptedAccessToken = encryptedAccess

	// Update access token expiration time
	if !newToken.Expiry.IsZero() {
		session.AccessTokenExpiresAt = &newToken.Expiry
	} else {
		// If no expiry provided, assume token doesn't expire
		session.AccessTokenExpiresAt = nil
	}

	// Update refresh token if provided in response
	if newToken.RefreshToken != "" {
		encryptedRefresh, err := s.encryption.Encrypt(ctx, []byte(newToken.RefreshToken), encContext)
		if err != nil {
			return fmt.Errorf("failed to encrypt new refresh token: %w", err)
		}
		session.EncryptedRefreshToken = encryptedRefresh
	}

	// Update timestamp
	session.UpdatedAt = time.Now()

	// Persist updated session (upsert semantics)
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return fmt.Errorf("failed to update session with refreshed tokens: %w", err)
	}

	s.logger.Info("session tokens updated",
		"principal", principal,
		"service_id", session.ServiceID,
		"session_id", session.ID)

	return nil
}

// DecryptAccessToken retrieves and decrypts the stored access token from a session.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - principal: The user principal (for encryption context)
//   - session: The UserSession containing encrypted token
//
// Returns:
//   - string: The decrypted access token value
//   - error if decryption fails
//
// The encryption context uses principal/service/session for verification.
func (s *OAuth2SessionService) DecryptAccessToken(
	ctx context.Context,
	principal string,
	session *storage.UserSession,
) (string, error) {
	if session == nil {
		return "", fmt.Errorf("session cannot be nil")
	}

	encContext := map[string]string{
		"principal":  principal,
		"service_id": session.ServiceID,
		"session_id": session.ID,
	}

	accessToken, err := s.encryption.Decrypt(ctx, session.EncryptedAccessToken, encContext)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt access token: %w", err)
	}

	return string(accessToken), nil
}

// DecryptRefreshToken retrieves and decrypts the stored refresh token from a session.
// Returns empty string (not error) if no refresh token stored.
func (s *OAuth2SessionService) DecryptRefreshToken(
	ctx context.Context,
	principal string,
	session *storage.UserSession,
) (string, error) {
	if session == nil {
		return "", fmt.Errorf("session cannot be nil")
	}

	// Return empty if no refresh token stored
	if len(session.EncryptedRefreshToken) == 0 {
		return "", nil
	}

	encContext := map[string]string{
		"principal":  principal,
		"service_id": session.ServiceID,
		"session_id": session.ID,
	}

	refreshToken, err := s.encryption.Decrypt(ctx, session.EncryptedRefreshToken, encContext)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	return string(refreshToken), nil
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

// GetValidAccessToken retrieves a valid, non-expired access token for a user at a service.
// This method transparently handles token refresh if the access token has expired but
// a valid refresh token is available.
//
// Usage pattern for callers:
//
//	token, err := s.GetValidAccessToken(ctx, principal, serviceID)
//	if err != nil {
//	    // Handle error (session not found, all tokens expired, etc)
//	}
//	// Use token - it's guaranteed valid and non-expired
//
// Encapsulated logic:
// 1. Fetch session from repository
// 2. Check if access token is expired
// 3. If expired, refresh using refresh token (if available)
// 4. Update session with new tokens
// 5. Decrypt and return valid access token
//
// Error cases:
//   - ErrSessionNotFound: No session exists for principal+service (T075)
//   - ErrSessionExpired: Both tokens expired, user must re-authenticate (T076)
//   - ErrRefreshFailed: Upstream provider rejected refresh request
//   - fmt.Errorf wraps: Other retrieval/decryption/storage errors
//
// Per SR-005 (Security Rule): Token values are never included in error messages.
// Only metadata (service, expiration times) is included.
func (s *OAuth2SessionService) GetValidAccessToken(
	ctx context.Context,
	principal string,
	serviceID string,
) (string, error) {
	// Step 1: Fetch session from repository
	// This is the ONLY repository access in this flow.
	// All other operations use session aggregate methods.
	session, err := s.sessionRepo.FindByPrincipalAndService(ctx, principal, serviceID)
	if err != nil {
		if err == ports.ErrNotFound {
			// T075: Session doesn't exist
			return "", fmt.Errorf("%w: principal=%s, service=%s", ErrSessionNotFound, principal, serviceID)
		}
		// Other repository errors (connection, timeout, etc)
		return "", fmt.Errorf("failed to retrieve session: %w", err)
	}

	// Defensive: ensure session is not nil (should not happen given err is nil)
	if session == nil {
		return "", fmt.Errorf("%w: session is nil (principal=%s, service=%s)", ErrSessionNotFound, principal, serviceID)
	}

	// Step 2: Check if access token has expired
	if session.HasValidAccessToken() {
		// Access token is still valid - just decrypt and return
		accessToken, err := s.DecryptAccessToken(ctx, principal, session)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt access token: %w", err)
		}
		return accessToken, nil
	}

	// Access token has expired - check if we can refresh

	// Step 3: Check if refresh token is available and valid
	if !session.CanRefresh() {
		// T076: Both tokens are expired
		return "", fmt.Errorf("%w: principal=%s, service=%s", ErrSessionExpired, principal, serviceID)
	}

	// Step 4: Fetch service details for refresh operation
	service, err := s.serviceRepo.Get(ctx, serviceID)
	if err != nil {
		s.logger.Error("failed to fetch service for token refresh",
			"principal", principal,
			"service_id", serviceID,
			"err", err)
		return "", fmt.Errorf("failed to fetch service for token refresh: %w", err)
	}

	if service == nil {
		return "", fmt.Errorf("service not found for refresh: service_id=%s", serviceID)
	}

	// Step 5: Decrypt refresh token
	refreshToken, err := s.DecryptRefreshToken(ctx, principal, session)
	if err != nil {
		s.logger.Error("failed to decrypt refresh token",
			"principal", principal,
			"service_id", serviceID,
			"err", err)
		return "", fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	// Defensive: ensure refresh token is not empty
	if refreshToken == "" {
		return "", fmt.Errorf("refresh token is empty: principal=%s, service=%s", principal, serviceID)
	}

	// Step 6: Call upstream OAuth2 provider to refresh tokens
	newToken, err := s.RefreshAccessToken(ctx, service, refreshToken)
	if err != nil {
		s.logger.Error("oauth2_refresh_failed",
			"event", "session.oauth2.refresh_failed",
			"principal", principal,
			"service_id", serviceID,
			"error", err.Error(),
			"timestamp", time.Now().Unix())
		return "", fmt.Errorf("%w: %v", ErrRefreshFailed, err)
	}

	if newToken == nil {
		return "", fmt.Errorf("upstream provider returned nil token: principal=%s, service=%s", principal, serviceID)
	}

	// Step 7: Update session with new tokens from refresh
	if err := s.UpdateSessionTokens(ctx, principal, session, newToken); err != nil {
		s.logger.Error("failed to update session with refreshed tokens",
			"principal", principal,
			"service_id", serviceID,
			"err", err)
		return "", fmt.Errorf("failed to update session with refreshed tokens: %w", err)
	}

	// Audit log: Token refresh succeeded
	s.logger.Info("oauth2_token_refreshed",
		"event", "session.oauth2.token_refreshed",
		"principal", principal,
		"service_id", serviceID,
		"timestamp", time.Now().Unix())

	// Step 8: Decrypt and return the new access token
	accessToken, err := s.DecryptAccessToken(ctx, principal, session)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt refreshed access token: %w", err)
	}

	return accessToken, nil
}

// GetSessionWithValidToken retrieves session metadata plus a valid access token.
// This is useful when the caller needs to build a response that includes session
// metadata (scope, token_type, expires_in, etc) along with the token itself.
//
// Internally uses GetValidAccessToken to ensure token is valid (with auto-refresh).
//
// Usage pattern for RFC 8693 token exchange response:
//
//	session, token, err := s.GetSessionWithValidToken(ctx, principal, serviceID)
//	if err != nil {
//	    // Handle error
//	}
//	// Build response using both:
//	response.AccessToken = token
//	response.Scope = strings.Join(session.Scope, " ")
//	response.ExpiresIn = calculateExpiresIn(session)
//
// Returns:
//   - session: Current session state (may have been updated by refresh)
//   - token: Valid, non-expired access token
//   - error: Same error cases as GetValidAccessToken
//
// Note: The session returned here may have been modified by refresh operation.
// The session's AccessTokenExpiresAt and UpdatedAt fields reflect the latest state.
func (s *OAuth2SessionService) GetSessionWithValidToken(
	ctx context.Context,
	principal string,
	serviceID string,
) (*storage.UserSession, string, error) {
	// Step 1: Get valid token (this handles all refresh logic transparently)
	token, err := s.GetValidAccessToken(ctx, principal, serviceID)
	if err != nil {
		// Return error as-is; no additional wrapping needed
		return nil, "", err
	}

	// Step 2: Re-fetch session to get updated metadata
	// This is necessary because GetValidAccessToken may have refreshed the token,
	// updating session.AccessTokenExpiresAt and session.UpdatedAt
	session, err := s.sessionRepo.FindByPrincipalAndService(ctx, principal, serviceID)
	if err != nil {
		s.logger.Error("failed to fetch session metadata after token retrieval",
			"principal", principal,
			"service_id", serviceID,
			"err", err)
		return nil, "", fmt.Errorf("failed to fetch session metadata: %w", err)
	}

	if session == nil {
		// This should not happen since GetValidAccessToken just succeeded
		s.logger.Error("session became nil after successful token retrieval",
			"principal", principal,
			"service_id", serviceID)
		return nil, "", fmt.Errorf("session disappeared after token retrieval: principal=%s, service=%s", principal, serviceID)
	}

	return session, token, nil
}
