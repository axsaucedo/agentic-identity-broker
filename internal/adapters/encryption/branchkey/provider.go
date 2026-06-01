package branchkey

import domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"

// DefaultProvider implements the ports.BranchKeyIdProvider interface
type DefaultProvider struct{}

func NewDefaultProvider() *DefaultProvider {
	return &DefaultProvider{}
}

func (p *DefaultProvider) GenerateBranchKeyId(subject domainencryption.BranchKeySubject) (string, error) {
	return GenerateBranchKeyId(subject)
}

func (p *DefaultProvider) ExtractSubjectFromBranchKey(branchKeyID string) (domainencryption.BranchKeySubject, error) {
	return ExtractSubject(branchKeyID)
}
