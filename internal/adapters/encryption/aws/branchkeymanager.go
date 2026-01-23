package aws

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
)

// AWSBranchKeyManager implements ports.BranchKeyManager for AWS hierarchical keyring.
// Handles branch key provisioning for services in DynamoDB via the AWS Encryption SDK KeyStore.
//
// Architecture:
// - Create: Provisions a new branch key for a service using GetBranchKeyID() and KeyStore
// - Get: Retrieves branch key ID (uses same deterministic ID generation)
// - ID resolution at runtime: Delegated to BranchKeyIdSupplier for encryption/decryption
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

// Create creates a branch key for the service in DynamoDB.
// Implements BranchKeyRepository.Create()
func (m *AWSBranchKeyManager) Create(ctx context.Context, serviceID string) (string, error) {
	if m == nil || m.keyStore == nil {
		return "", encryption.NewKEKUnavailableError("branch key manager not properly initialized", nil)
	}

	// 1. Validate service_id (non-empty)
	if serviceID == "" {
		return "", encryption.NewKEKUnavailableError("service_id cannot be empty", nil)
	}

	// 2. Generate deterministic branch key ID (using helper to match supplier logic)
	branchKeyID := GetBranchKeyID(serviceID)

	// 3. Provision in DynamoDB via KeyStore
	provisioned, err := m.keyStore.CreateBranchKey(ctx, branchKeyID)
	if err != nil {
		return "", err
	}

	return provisioned, nil
}
