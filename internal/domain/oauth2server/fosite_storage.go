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
	codeRepo       ports.AuthorizationCodeRepository
	pkceRepo       ports.PKCESessionRepository
	credRepo       ports.ClientCredentialRepository
	clientResolver ports.ClientResolver
	logger         *slog.Logger
}

// NewFositeStorage creates a new FositeStorage wrapping our repositories.
func NewFositeStorage(
	codeRepo ports.AuthorizationCodeRepository,
	pkceRepo ports.PKCESessionRepository,
	credRepo ports.ClientCredentialRepository,
	clientResolver ports.ClientResolver,
	logger *slog.Logger,
) *FositeStorage {
	return &FositeStorage{
		codeRepo:       codeRepo,
		pkceRepo:       pkceRepo,
		credRepo:       credRepo,
		clientResolver: clientResolver,
		logger:         logger,
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

	// Look up the client using the stored ClientID (the original client_id from the authorize
	// request). For CIMD clients this is the metadata URL; for opaque clients it's the UUID.
	// Using AgentID.String() would fail for CIMD clients whose resolver rejects UUID lookups.
	clientLookupID := authCode.ClientID.String()
	if clientLookupID == "" {
		clientLookupID = authCode.AgentID.String()
	}
	client, err := s.GetClient(ctx, clientLookupID)
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
		RequestedScope: splitScope(authCode.Scope),
		GrantedScope:   splitScope(authCode.Scope),
		Form: map[string][]string{
			"redirect_uri":   {authCode.RedirectURI},
			"code_challenge": {authCode.CodeChallenge},
			// code_challenge_method is hardcoded to S256 because EnablePKCEPlainChallengeMethod
			// is false in provider config, so only S256 can reach this point. The actual PKCE
			// verification uses GetPKCERequestSession which stores the method independently.
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
		mapped := s.mapStorageError(ctx, err)
		if errors.Is(mapped, fosite.ErrNotFound) {
			return fosite.ErrInvalidatedAuthorizeCode
		}
		return mapped
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
	if err := s.pkceRepo.Delete(ctx, signature); err != nil {
		return s.mapStorageError(ctx, err)
	}
	return nil
}

// GetClient satisfies the fosite.Storage interface.
// Delegates to ClientResolver which handles both UUID-format and URL-format client_id values.
func (s *FositeStorage) GetClient(ctx context.Context, clientID string) (fosite.Client, error) {
	resolution, err := s.clientResolver.ResolveClient(ctx, id.ClientID(clientID))
	if err != nil {
		var clientErr *ports.ClientIDError
		if errors.As(err, &clientErr) {
			if clientErr.Code == "server_error" {
				s.logger.ErrorContext(ctx, "infrastructure error during client resolution",
					"client_id", clientID, "error", clientErr.Desc)
				return nil, err
			}
			return nil, fosite.ErrNotFound
		}
		return nil, err
	}

	if resolution.CIMDMetadata != nil {
		return &publicClient{clientID: clientID, agent: resolution.Agent, redirectURIs: resolution.CIMDMetadata.RedirectURIs}, nil
	}

	// LocalClient agents are public clients — no credential is registered for them.
	if resolution.Agent.ClientType() == storage.LocalClient {
		return &publicClient{clientID: clientID, agent: resolution.Agent, redirectURIs: resolution.Agent.RedirectURIs}, nil
	}

	cred, err := s.credRepo.GetByAgentID(ctx, resolution.Agent.ID)
	if err != nil {
		return nil, s.mapStorageError(ctx, err)
	}
	return &confidentialClient{clientID: clientID, agent: resolution.Agent, credential: cred}, nil
}

// ClientAssertionJWTValid rejects all JWT assertions. This broker uses client_secret_basic only;
// private_key_jwt and client_secret_jwt are not supported. Returning ErrJTIKnown fails closed:
// any accidental invocation rejects the assertion rather than silently accepting a replay.
func (s *FositeStorage) ClientAssertionJWTValid(_ context.Context, _ string) error {
	return fosite.ErrJTIKnown
}

// SetClientAssertionJWT is a no-op. JWT client assertions are not supported;
// ClientAssertionJWTValid always rejects, so JTIs never need to be recorded.
func (s *FositeStorage) SetClientAssertionJWT(_ context.Context, _ string, _ time.Time) error {
	return nil
}
