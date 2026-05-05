package app

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

type issueTokenMintingStrategy struct {
	provider *oauth2server.Provider
}

func newIssueTokenMintingStrategy(provider *oauth2server.Provider) *issueTokenMintingStrategy {
	return &issueTokenMintingStrategy{provider: provider}
}

func (s *issueTokenMintingStrategy) HandleClientCredentials(ctx context.Context, clientID id.ClientID, clientSecret, scope string) (*ports.TokenResponse, error) {
	return s.provider.HandleClientCredentials(ctx, string(clientID), clientSecret, scope)
}

func (s *issueTokenMintingStrategy) HandleAuthorizationCodeExchange(ctx context.Context, clientID id.ClientID, clientSecret, code, redirectURI, codeVerifier string) (*ports.TokenResponse, error) {
	return s.provider.HandleAuthorizationCodeExchange(ctx, string(clientID), clientSecret, code, redirectURI, codeVerifier)
}

type issueTokenCodeIssuer struct {
	provider *oauth2server.Provider
}

func newIssueTokenCodeIssuer(provider *oauth2server.Provider) *issueTokenCodeIssuer {
	return &issueTokenCodeIssuer{provider: provider}
}

func (s *issueTokenCodeIssuer) IssueAuthorizationCode(ctx context.Context, req *ports.AuthorizationRequest, principal id.Principal) (string, error) {
	return s.provider.HandleAuthorize(
		ctx,
		string(req.ClientID),
		req.RedirectURI,
		req.ResponseType,
		req.Scope,
		req.State,
		req.CodeChallenge,
		req.CodeChallengeMethod,
		principal,
	)
}
