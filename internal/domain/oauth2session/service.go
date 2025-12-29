package oauth2session

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// OAuth2SessionService orchestrates OAuth2 authorization flows and session management.
type OAuth2SessionService struct {
	serviceRepo ports.ThirdpartyOAuth2ServiceRepository
	sessionRepo ports.UserSessionRepository
	grantRepo   ports.UserGrantRepository // For counting dependent agents
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

// NewOAuth2SessionService creates a new OAuth2SessionService.
func NewOAuth2SessionService(
	serviceRepo ports.ThirdpartyOAuth2ServiceRepository,
	sessionRepo ports.UserSessionRepository,
	grantRepo ports.UserGrantRepository,
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
		encryption:  encryption,
		jweKey:      jweKey,
		config:      config,
		logger:      logger,
	}
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

// SessionWithAgents represents a session with information about dependent agents.
type SessionWithAgents struct {
	Session         *storage.UserSession
	DependentAgents []string // Agent IDs that use this session
}

// Stub methods - will be implemented in User Story tasks
// These exist to allow code to compile before implementation

// InitiateOAuth2Flow starts the OAuth2 authorization code flow.
// Creates PKCE verifier/challenge, generates JWE state token, returns auth URL.
func (s *OAuth2SessionService) InitiateOAuth2Flow(
	ctx context.Context,
	principal string,
	serviceID string,
	redirectURI string,
) (*InitiateFlowResult, error) {
	s.logger.Info("stub: InitiateOAuth2Flow")
	return nil, fmt.Errorf("not implemented")
}

// HandleCallback processes the OAuth2 callback, exchanges code for tokens, stores session.
func (s *OAuth2SessionService) HandleCallback(
	ctx context.Context,
	principal string,
	req *HandleCallbackRequest,
) (*storage.UserSession, error) {
	s.logger.Info("stub: HandleCallback")
	return nil, fmt.Errorf("not implemented")
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
	s.logger.Info("stub: TerminateSession")
	return fmt.Errorf("not implemented")
}

// GetSessionWithAgents returns session details including list of dependent agents.
func (s *OAuth2SessionService) GetSessionWithAgents(
	ctx context.Context,
	principal string,
	serviceID string,
) (*SessionWithAgents, error) {
	s.logger.Info("stub: GetSessionWithAgents")
	return nil, fmt.Errorf("not implemented")
}
