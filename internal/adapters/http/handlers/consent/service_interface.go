package consent

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// ConsentService defines the interface that handlers depend on.
// This allows for easier testing with mock implementations.
type ConsentService interface {
	GetAgentConsentInfo(ctx context.Context, agentID string) (*consent.AgentConsentInfo, error)
	GrantConsent(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error)
	RevokeConsent(ctx context.Context, principal string, agentID string) error
	GetActiveGrants(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error)
	GetAgentDelegations(ctx context.Context, principal string) ([]consent.AgentDelegation, error)
	GetAgentDetail(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error)
	GetUserGrants(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error)
}

// Ensure consent.Service implements ConsentService interface
var _ ConsentService = (*consent.Service)(nil)
