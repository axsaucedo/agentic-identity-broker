package consent

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	domotp2 "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// ConsentService defines the interface that handlers depend on.
// This allows for easier testing with mock implementations.
type ConsentService interface {
	GetAgentConsentInfo(ctx context.Context, agentID id.AgentID) (*consent.AgentConsentInfo, error)
	GetAgentWithServiceRequirements(ctx context.Context, userPrincipal id.Principal, agentID id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error)
	GrantConsent(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error)
	RevokeConsent(ctx context.Context, principal id.Principal, agentID id.AgentID) error
	// RevokeConsentForPrincipal is the user-facing revocation entry point for FR-014.
	// Maps storage ErrNotFound → consent.ErrGrantNotFound so the handler can return 404.
	RevokeConsentForPrincipal(ctx context.Context, principal id.Principal, agentID id.AgentID) error
	GetAgentDelegations(ctx context.Context, principal id.Principal) ([]consent.AgentDelegation, error)
	GetUserGrants(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error)
}

// Ensure consent.Service implements ConsentService interface
var _ ConsentService = (*consent.Service)(nil)

// SessionTokenValidator validates JWE authorization session tokens.
type SessionTokenValidator interface {
	ValidateAuthorizationSessionToken(token string, agentID id.AgentID, principal id.Principal) (*domotp2.AuthorizationSessionClaims, error)
}

// Ensure AuthorizationService implements SessionTokenValidator.
var _ SessionTokenValidator = (*domotp2.AuthorizationService)(nil)
