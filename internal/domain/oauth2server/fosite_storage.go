package oauth2server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
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
	pkceRepo  ports.PKCESessionRepository
	agentRepo ports.AgentRepository
	credRepo  ports.BrokerClientCredentialRepository
	logger    *slog.Logger
}

// NewFositeStorage creates a new FositeStorage wrapping our repositories.
func NewFositeStorage(
	codeRepo ports.AuthorizationCodeRepository,
	pkceRepo ports.PKCESessionRepository,
	agentRepo ports.AgentRepository,
	credRepo ports.BrokerClientCredentialRepository,
	logger *slog.Logger,
) *FositeStorage {
	return &FositeStorage{
		codeRepo:  codeRepo,
		pkceRepo:  pkceRepo,
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
	agentID, err := extractAgentID(req.GetClient())
	if err != nil {
		return fmt.Errorf("CreateAuthorizeCodeSession: %w", err)
	}
	session := req.GetSession()
	authCode := &storage.AuthorizationCode{
		ID:            id.NewAuthorizationCodeID(),
		CodeHash:      code, // code is already the fosite signature (sha256Hex of raw code)
		AgentID:       agentID,
		ClientID:      id.NewClientID(req.GetRequestForm().Get("client_id")),
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
	authCode, err := s.codeRepo.FindByCodeHash(ctx, code)
	if err != nil {
		return nil, s.mapStorageError(ctx, err)
	}

	// Look up the client (agent) — needed even for invalidated codes so fosite can revoke tokens.
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

	// fosite requires a non-nil requester alongside ErrInvalidatedAuthorizeCode so it can
	// revoke associated tokens (RFC 6749 §4.1.2 — code replay must revoke previous grants).
	if authCode.UsedAt != nil {
		return req, fosite.ErrInvalidatedAuthorizeCode
	}

	// Expiry is intentionally not checked here. fosite inspects session.GetExpiresAt(AuthorizeCode)
	// and returns ErrTokenExpired itself. Returning ErrInvalidatedAuthorizeCode for expired codes
	// would incorrectly trigger fosite's replay-attack revocation path.
	return req, nil
}

// InvalidateAuthorizeCodeSession marks an authorization code as used.
func (s *FositeStorage) InvalidateAuthorizeCodeSession(ctx context.Context, code string) error {
	authCode, err := s.codeRepo.FindByCodeHash(ctx, code)
	if err != nil {
		return s.mapStorageError(ctx, err)
	}
	if err := s.codeRepo.MarkUsed(ctx, authCode.ID); err != nil {
		var storageErr *storage.StorageError
		if errors.As(err, &storageErr) && storageErr.Kind == storage.ErrorKindNotFound {
			return fosite.ErrInvalidatedAuthorizeCode
		}
		return err
	}
	return nil
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

// RevokeRefreshToken is a no-op (refresh tokens not supported).
func (s *FositeStorage) RevokeRefreshToken(_ context.Context, _ string) error {
	return nil
}

// CreatePKCERequestSession stores the PKCE code challenge in a dedicated store keyed by code signature.
func (s *FositeStorage) CreatePKCERequestSession(ctx context.Context, signature string, req fosite.Requester) error {
	session := &storage.PKCESession{
		Signature:           signature,
		CodeChallenge:       req.GetRequestForm().Get("code_challenge"),
		CodeChallengeMethod: req.GetRequestForm().Get("code_challenge_method"),
		ExpiresAt:           req.GetSession().GetExpiresAt(fosite.AuthorizeCode),
		CreatedAt:           time.Now(),
	}
	if err := session.Validate(); err != nil {
		return err
	}
	return s.pkceRepo.Create(ctx, session)
}

// GetPKCERequestSession retrieves the PKCE challenge for verifying the code verifier at the token endpoint.
func (s *FositeStorage) GetPKCERequestSession(ctx context.Context, signature string, _ fosite.Session) (fosite.Requester, error) {
	session, err := s.pkceRepo.FindBySignature(ctx, signature)
	if err != nil {
		return nil, s.mapStorageError(ctx, err)
	}
	return &fosite.Request{
		Form: url.Values{
			"code_challenge":        {session.CodeChallenge},
			"code_challenge_method": {session.CodeChallengeMethod},
		},
	}, nil
}

// DeletePKCERequestSession removes the PKCE session after successful token exchange (one-shot).
func (s *FositeStorage) DeletePKCERequestSession(ctx context.Context, signature string) error {
	return s.pkceRepo.Delete(ctx, signature)
}

// GetClient satisfies the fosite.Storage interface.
// This broker uses fosite in library mode: Provider pre-resolves the client via buildClient or
// Authenticate before constructing the fosite request with req.Client already set. fosite's
// grant handlers (AuthorizeExplicitGrantHandler, pkce.Handler, ClientCredentialsGrantHandler)
// therefore never call GetClient — they consume the pre-populated req.Client or the client
// embedded in GetAuthorizeCodeSession's return value. GetClient would only be called in
// fosite's framework mode (fosite.Provider.NewAuthorizeRequest / NewAccessRequest), which
// this code does not use.
// clientID is expected to be the agent UUID, consistent with brokerClient.GetID().
func (s *FositeStorage) GetClient(ctx context.Context, clientID string) (fosite.Client, error) {
	agentID, err := id.ParseAgentID(clientID)
	if err != nil {
		return nil, fosite.ErrNotFound
	}
	client, err := s.buildClient(ctx, agentID)
	if err != nil {
		return nil, s.mapStorageError(ctx, err)
	}
	return client, nil
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

func extractAgentID(client fosite.Client) (id.AgentID, error) {
	bc, ok := client.(*brokerClient)
	if !ok {
		return id.AgentID{}, fmt.Errorf("expected *brokerClient, got %T", client)
	}
	if bc == nil {
		return id.AgentID{}, fmt.Errorf("brokerClient is nil")
	}
	if bc.agent == nil {
		return id.AgentID{}, fmt.Errorf("brokerClient.agent is nil")
	}
	if bc.agent.ID.IsZero() {
		return id.AgentID{}, fmt.Errorf("brokerClient.agent.ID is zero")
	}
	return bc.agent.ID, nil
}
