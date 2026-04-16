package oauth2

import (
	"context"
	"fmt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// IssueTokenMintingStrategy implements TokenMintingStrategy by delegating to
// the local oauth2server.Provider for token minting. Used in issue_token mode.
type IssueTokenMintingStrategy struct {
	provider *oauth2server.Provider
}

// NewIssueTokenMintingStrategy creates a new strategy wrapping the Provider.
func NewIssueTokenMintingStrategy(provider *oauth2server.Provider) *IssueTokenMintingStrategy {
	return &IssueTokenMintingStrategy{provider: provider}
}

// HandleClientCredentials processes a client_credentials grant locally.
func (s *IssueTokenMintingStrategy) HandleClientCredentials(ctx context.Context, agentID id.AgentID, clientSecret, scope string) (*ports.TokenResponse, error) {
	return s.provider.HandleClientCredentials(ctx, agentID, clientSecret, scope)
}

// HandleAuthorizationCodeExchange processes an authorization_code exchange locally.
func (s *IssueTokenMintingStrategy) HandleAuthorizationCodeExchange(ctx context.Context, agentID id.AgentID, clientSecret, code, redirectURI, codeVerifier string) (*ports.TokenResponse, error) {
	return s.provider.HandleAuthorizationCodeExchange(ctx, agentID, clientSecret, code, redirectURI, codeVerifier)
}

// IssueTokenCodeIssuer implements AuthorizationCodeIssuer by delegating to
// the local oauth2server.Provider. Used in issue_token mode.
type IssueTokenCodeIssuer struct {
	provider *oauth2server.Provider
}

// NewIssueTokenCodeIssuer creates a new code issuer wrapping the Provider.
func NewIssueTokenCodeIssuer(provider *oauth2server.Provider) *IssueTokenCodeIssuer {
	return &IssueTokenCodeIssuer{provider: provider}
}

// IssueAuthorizationCode issues a local authorization code via the Provider.
func (s *IssueTokenCodeIssuer) IssueAuthorizationCode(ctx context.Context, req *ports.AuthorizationRequest, principal id.Principal) (string, error) {
	codeChallengeMethod := req.CodeChallengeMethod
	if codeChallengeMethod == "" {
		return "", fmt.Errorf("%w: code_challenge_method is required (PKCE mandatory)", oauth2server.ErrInvalidRequest)
	}

	// Validate PKCE is provided (mandatory in issue_token mode)
	if req.CodeChallenge == "" {
		return "", fmt.Errorf("%w: code_challenge is required (PKCE mandatory)", oauth2server.ErrInvalidRequest)
	}

	// Validate only S256 is accepted
	if codeChallengeMethod != "S256" {
		return "", fmt.Errorf("%w: only S256 code_challenge_method is supported", oauth2server.ErrInvalidRequest)
	}

	// Validate response_type
	if req.ResponseType != "code" {
		return "", fmt.Errorf("%w: only 'code' response_type is supported", oauth2server.ErrUnsupportedResponseType)
	}

	// req.ClientID is already id.AgentID (parsed at the HTTP handler boundary).
	code, err := s.provider.HandleAuthorize(
		ctx,
		req.ClientID,
		req.RedirectURI,
		req.ResponseType,
		req.Scope,
		req.State,
		req.CodeChallenge,
		codeChallengeMethod,
		principal,
	)
	if err != nil {
		// Provider already uses sentinel errors — just propagate them
		return "", err
	}

	return code, nil
}
