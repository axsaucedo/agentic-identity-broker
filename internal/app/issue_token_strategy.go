package app

import (
	"context"
	"fmt"

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

func (s *issueTokenMintingStrategy) HandleClientCredentials(ctx context.Context, agentID id.AgentID, clientSecret, scope string) (*ports.TokenResponse, error) {
	return s.provider.HandleClientCredentials(ctx, agentID, clientSecret, scope)
}

func (s *issueTokenMintingStrategy) HandleAuthorizationCodeExchange(ctx context.Context, agentID id.AgentID, clientSecret, code, redirectURI, codeVerifier string) (*ports.TokenResponse, error) {
	return s.provider.HandleAuthorizationCodeExchange(ctx, agentID, clientSecret, code, redirectURI, codeVerifier)
}

type issueTokenCodeIssuer struct {
	provider *oauth2server.Provider
}

func newIssueTokenCodeIssuer(provider *oauth2server.Provider) *issueTokenCodeIssuer {
	return &issueTokenCodeIssuer{provider: provider}
}

func (s *issueTokenCodeIssuer) IssueAuthorizationCode(ctx context.Context, req *ports.AuthorizationRequest, principal id.Principal) (string, error) {
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
