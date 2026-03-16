package branchkey

import "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"

// DefaultProvider implements the ports.BranchKeyIdProvider interface
type DefaultProvider struct{}

func NewDefaultProvider() *DefaultProvider {
	return &DefaultProvider{}
}

func (p *DefaultProvider) GenerateBranchKeyId(serviceID id.ServiceID) string {
	return GenerateBranchKeyId(serviceID.String())
}

func (p *DefaultProvider) ExtractServiceIdFromBranchKey(branchKeyID string) id.ServiceID {
	serviceID, err := ExtractServiceID(branchKeyID)
	if err != nil {
		return id.ServiceID{}
	}
	parsed, err := id.ParseServiceID(serviceID)
	if err != nil {
		return id.ServiceID{}
	}
	return parsed
}
