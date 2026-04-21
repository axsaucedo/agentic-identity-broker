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
// PKCE enforcement, response_type validation, and scope checking are delegated to the Provider
// which in turn delegates to fosite handlers.
func (s *IssueTokenCodeIssuer) IssueAuthorizationCode(ctx context.Context, req *ports.AuthorizationRequest, principal id.Principal) (string, error) {
	agentID, err := id.ParseAgentID(string(req.ClientID))
	if err != nil {
		return "", fmt.Errorf("%w: invalid client_id: %v", oauth2server.ErrInvalidClient, err)
	}
	return s.provider.HandleAuthorize(
		ctx,
		agentID,
		req.RedirectURI,
		req.ResponseType,
		req.Scope,
		req.State,
		req.CodeChallenge,
		req.CodeChallengeMethod,
		principal,
	)
}
