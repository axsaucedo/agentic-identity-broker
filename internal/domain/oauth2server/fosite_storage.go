package oauth2server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ory/fosite"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// FositeStorage wraps project repositories to implement fosite storage interfaces.
type FositeStorage struct {
	codeRepo  ports.AuthorizationCodeRepository
	agentRepo ports.AgentRepository
	credRepo  ports.BrokerClientCredentialRepository
	logger    *slog.Logger
}

// NewFositeStorage creates a new FositeStorage wrapping our repositories.
func NewFositeStorage(
	codeRepo ports.AuthorizationCodeRepository,
	agentRepo ports.AgentRepository,
	credRepo ports.BrokerClientCredentialRepository,
	logger *slog.Logger,
) *FositeStorage {
	return &FositeStorage{
		codeRepo:  codeRepo,
		agentRepo: agentRepo,
		credRepo:  credRepo,
		logger:    logger,
	}
}

// mapStorageError translates a storage error for fosite consumption.
// Not-found errors become fosite.ErrNotFound so fosite maps them to invalid_grant/invalid_client.
// Infrastructure errors (connection, timeout) are logged and returned as-is so fosite surfaces
// them as server_error rather than silently treating them as "not found".
func (s *FositeStorage) mapStorageError(ctx context.Context, err error) error {
	var se *storage.StorageError
	if errors.As(err, &se) {
		if se.Kind == storage.ErrorKindNotFound {
			return fosite.ErrNotFound
		}
		s.logger.ErrorContext(ctx, "storage failure in OAuth2 flow",
			"storage_op", se.Operation,
			"storage_kind", string(se.Kind),
		)
		return err
	}
	s.logger.ErrorContext(ctx, "unexpected error type in OAuth2 flow", "error", err.Error())
	return err
}

// CreateAuthorizeCodeSession stores an authorization code issued by the authorization endpoint.
func (s *FositeStorage) CreateAuthorizeCodeSession(ctx context.Context, code string, req fosite.Requester) error {
	session := req.GetSession()
	authCode := &storage.AuthorizationCode{
		ID:            id.NewAuthorizationCodeID(),
		CodeHash:      sha256Hex(code),
		AgentID:       extractAgentID(req.GetClient()),
		Principal:     id.NewPrincipal(session.GetSubject()),
		RedirectURI:   req.GetRequestForm().Get("redirect_uri"),
		CodeChallenge: req.GetRequestForm().Get("code_challenge"),
		Scope:         strings.Join(req.GetRequestedScopes(), " "),
		ExpiresAt:     session.GetExpiresAt(fosite.AuthorizeCode),
		CreatedAt:     time.Now(),
	}
	return s.codeRepo.Create(ctx, authCode)
}

// GetAuthorizeCodeSession retrieves an authorization code session by code signature.
func (s *FositeStorage) GetAuthorizeCodeSession(ctx context.Context, code string, _ fosite.Session) (fosite.Requester, error) {
	codeHash := sha256Hex(code)
	authCode, err := s.codeRepo.FindByCodeHash(ctx, codeHash)
	if err != nil {
		return nil, s.mapStorageError(ctx, err)
	}

	if authCode.UsedAt != nil {
		return nil, fosite.ErrInvalidatedAuthorizeCode
	}

	if time.Now().After(authCode.ExpiresAt) {
		return nil, fosite.ErrInvalidatedAuthorizeCode
	}

	// Look up the client (agent)
	client, err := s.buildClient(ctx, authCode.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to look up client: %w", err)
	}

	// Reconstruct the fosite request
	session := &fosite.DefaultSession{
		Subject: authCode.Principal.String(),
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AuthorizeCode: authCode.ExpiresAt,
		},
	}

	req := &fosite.Request{
		ID:             authCode.ID.String(),
		Client:         client,
		Session:        session,
		RequestedScope: strings.Split(authCode.Scope, " "),
		GrantedScope:   strings.Split(authCode.Scope, " "),
		Form: map[string][]string{
			"redirect_uri":          {authCode.RedirectURI},
			"code_challenge":        {authCode.CodeChallenge},
			"code_challenge_method": {"S256"},
		},
		RequestedAt: authCode.CreatedAt,
	}

	return req, nil
}

// InvalidateAuthorizeCodeSession marks an authorization code as used.
func (s *FositeStorage) InvalidateAuthorizeCodeSession(ctx context.Context, code string) error {
	codeHash := sha256Hex(code)
	authCode, err := s.codeRepo.FindByCodeHash(ctx, codeHash)
	if err != nil {
		return s.mapStorageError(ctx, err)
	}
	return s.codeRepo.MarkUsed(ctx, authCode.ID)
}

// CreateAccessTokenSession is a no-op for stateless JWT tokens.
func (s *FositeStorage) CreateAccessTokenSession(_ context.Context, _ string, _ fosite.Requester) error {
	return nil // JWT tokens are stateless — no storage needed
}

// GetAccessTokenSession is not needed for stateless JWT tokens.
func (s *FositeStorage) GetAccessTokenSession(_ context.Context, _ string, _ fosite.Session) (fosite.Requester, error) {
	return nil, fosite.ErrNotFound // JWT tokens are stateless
}

// DeleteAccessTokenSession is a no-op for stateless JWT tokens.
func (s *FositeStorage) DeleteAccessTokenSession(_ context.Context, _ string) error {
	return nil // JWT tokens are stateless
}

// CreateRefreshTokenSession is not supported (no refresh tokens in this mode).
func (s *FositeStorage) CreateRefreshTokenSession(_ context.Context, _ string, _ string, _ fosite.Requester) error {
	return nil
}

// GetRefreshTokenSession is not supported.
func (s *FositeStorage) GetRefreshTokenSession(_ context.Context, _ string, _ fosite.Session) (fosite.Requester, error) {
	return nil, fosite.ErrNotFound
}

// DeleteRefreshTokenSession is not supported.
func (s *FositeStorage) DeleteRefreshTokenSession(_ context.Context, _ string) error {
	return nil
}

// RotateRefreshToken is not supported.
func (s *FositeStorage) RotateRefreshToken(_ context.Context, _ string, _ string) error {
	return nil
}

// RevokeAccessToken is a no-op for stateless JWT tokens.
func (s *FositeStorage) RevokeAccessToken(_ context.Context, _ string) error {
	return nil
}

// CreatePKCERequestSession stores PKCE data (stored alongside the authorization code).
func (s *FositeStorage) CreatePKCERequestSession(_ context.Context, _ string, _ fosite.Requester) error {
	return nil // PKCE data is stored in the authorization code record
}

// GetPKCERequestSession retrieves PKCE data.
func (s *FositeStorage) GetPKCERequestSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return s.GetAuthorizeCodeSession(ctx, signature, session)
}

// DeletePKCERequestSession deletes PKCE data.
func (s *FositeStorage) DeletePKCERequestSession(_ context.Context, _ string) error {
	return nil // Cleaned up with authorization code
}

// GetClient retrieves a client from the credential repository.
func (s *FositeStorage) GetClient(ctx context.Context, clientID string) (fosite.Client, error) {
	cred, err := s.credRepo.GetByBrokerClientID(ctx, id.NewBrokerClientID(clientID))
	if err != nil {
		return nil, s.mapStorageError(ctx, err)
	}
	agent, err := s.agentRepo.Get(ctx, cred.AgentID)
	if err != nil {
		return nil, s.mapStorageError(ctx, err)
	}
	return &brokerClient{agent: agent, credential: cred}, nil
}

// ClientAssertionJWTValid checks for JWT assertion replay — not supported.
func (s *FositeStorage) ClientAssertionJWTValid(_ context.Context, _ string) error {
	return nil
}

// SetClientAssertionJWT records a JWT assertion — not supported.
func (s *FositeStorage) SetClientAssertionJWT(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (s *FositeStorage) buildClient(ctx context.Context, agentID id.AgentID) (fosite.Client, error) {
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return nil, err
	}
	cred, err := s.credRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return nil, err
	}
	return &brokerClient{agent: agent, credential: cred}, nil
}

func extractAgentID(client fosite.Client) id.AgentID {
	bc, ok := client.(*brokerClient)
	if !ok {
		panic(fmt.Sprintf("extractAgentID: expected *brokerClient, got %T", client))
	}
	return bc.agent.ID
}
