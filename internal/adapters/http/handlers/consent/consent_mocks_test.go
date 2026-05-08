package consent

import (
	"context"
	"errors"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// mockConsentService implements a mock consent service for testing.
//
//nolint:unused // Used in tests
type mockConsentService struct {
	getAgentConsentInfoFunc       func(ctx context.Context, agentID id.AgentID) (*consent.AgentConsentInfo, error)
	getActiveGrantsFunc           func(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error)
	grantConsentFunc              func(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error)
	revokeConsentFunc             func(ctx context.Context, principal id.Principal, agentID id.AgentID) error
	revokeConsentForPrincipalFunc func(ctx context.Context, principal id.Principal, agentID id.AgentID) error
	getAgentDelegationsFunc       func(ctx context.Context, principal id.Principal) ([]consent.AgentDelegation, error)
	getUserGrantsFunc             func(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error)
}

//nolint:unused // Used in tests
func (m *mockConsentService) GetAgentConsentInfo(ctx context.Context, agentID id.AgentID) (*consent.AgentConsentInfo, error) {
	if m.getAgentConsentInfoFunc != nil {
		return m.getAgentConsentInfoFunc(ctx, agentID)
	}
	return nil, errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) GetActiveGrants(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	if m.getActiveGrantsFunc != nil {
		return m.getActiveGrantsFunc(ctx, principal, agentID)
	}
	return nil, errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) GrantConsent(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
	if m.grantConsentFunc != nil {
		return m.grantConsentFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) RevokeConsent(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	if m.revokeConsentFunc != nil {
		return m.revokeConsentFunc(ctx, principal, agentID)
	}
	return errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) RevokeConsentForPrincipal(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	if m.revokeConsentForPrincipalFunc != nil {
		return m.revokeConsentForPrincipalFunc(ctx, principal, agentID)
	}
	return errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) GetAgentWithServiceRequirements(ctx context.Context, userPrincipal id.Principal, agentID id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
	return nil, nil, errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) GetAgentDelegations(ctx context.Context, principal id.Principal) ([]consent.AgentDelegation, error) {
	if m.getAgentDelegationsFunc != nil {
		return m.getAgentDelegationsFunc(ctx, principal)
	}
	return nil, errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) GetUserGrants(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	if m.getUserGrantsFunc != nil {
		return m.getUserGrantsFunc(ctx, principal, agentID)
	}
	return nil, errors.New("not implemented")
}
