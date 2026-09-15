package aws

import (
	"context"
	"fmt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
)

// AWSBranchKeyManager implements ports.BranchKeyManager for the AWS hierarchical keyring.
// It provisions branch keys for typed subjects in DynamoDB via the AWS Encryption SDK KeyStore.
type AWSBranchKeyManager struct {
	keyStore *KeyStore
}

// NewAWSBranchKeyManager creates a new AWS branch key manager for the given KeyStore.
// The manager handles branch key provisioning during service creation.
func NewAWSBranchKeyManager(keyStore *KeyStore) *AWSBranchKeyManager {
	return &AWSBranchKeyManager{
		keyStore: keyStore,
	}
}

// Create creates a branch key for the subject in DynamoDB.
// Implements BranchKeyRepository.Create().
func (m *AWSBranchKeyManager) Create(ctx context.Context, subject encryption.BranchKeySubject) (string, error) {
	if m == nil || m.keyStore == nil {
		return "", encryption.NewKEKUnavailableError("branch key manager not properly initialized", nil)
	}

	if err := subject.Validate(); err != nil {
		return "", encryption.NewKEKUnavailableError(fmt.Sprintf("invalid branch key subject: %v", err), err)
	}

	provisioned, err := m.keyStore.CreateBranchKey(ctx, subject)
	if err != nil {
		return "", err
	}

	return provisioned, nil
}
