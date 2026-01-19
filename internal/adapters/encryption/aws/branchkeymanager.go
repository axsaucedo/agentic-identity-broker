package aws

import (
	"context"
	"fmt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
)

// AWSBranchKeyManager implements ports.BranchKeyManager for AWS hierarchical keyring.
// Consolidates branch key provisioning (Create/Get) and ID resolution (ResolveBranchKeyID)
// into a single cohesive abstraction, separating these concerns from EncryptionPort.
//
// Architecture:
// - Branch key provisioning: Manages DynamoDB cache entries via KeyStore
// - ID resolution: Deterministic mapping from service_id to branch key identifier
// - Single source of truth: ResolveDeterministicBranchKeyID() helper ensures consistent ID generation
type AWSBranchKeyManager struct {
	keyStore *KeyStore
}

// NewAWSBranchKeyManager creates a new AWS branch key manager for the given KeyStore.
// The manager handles both provisioning and ID resolution for branch keys.
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
	branchKeyID := getBranchKeyId(serviceID)

	// 3. Provision in DynamoDB via KeyStore
	provisioned, err := m.keyStore.CreateBranchKey(ctx, branchKeyID)
	if err != nil {
		return "", err
	}

	return provisioned, nil
}

// Get retrieves a branch key ID for the given service.
// Implements BranchKeyRepository.Get()
// Returns error if branch key does not exist or query fails.
func (m *AWSBranchKeyManager) Get(ctx context.Context, serviceID string) (string, error) {
	if m == nil || m.keyStore == nil {
		return "", encryption.NewKEKUnavailableError("branch key manager not properly initialized", nil)
	}

	if serviceID == "" {
		return "", encryption.NewKEKUnavailableError("service_id cannot be empty", nil)
	}

	branchKeyID := getBranchKeyId(serviceID)

	// Query DynamoDB via KeyStore.GetKey() or similar
	// If key exists, return branchKeyID; if not found, return error; if error, return error
	// TODO: Implement actual query using KeyStore or DynamoDB client
	// For now: placeholder that returns the expected branch key ID
	// (assumes key exists since it was created by Create())
	return branchKeyID, nil
}

// ResolveBranchKeyID returns a deterministic branch key identifier based on service_id.
// Implements BranchKeyManager.ResolveBranchKeyID()
// Used by the hierarchical keyring during encryption/decryption to resolve the correct branch key.
// This is a pure function - same service_id always returns the same branch key ID.
func (m *AWSBranchKeyManager) ResolveBranchKeyID(serviceID string) (string, error) {
	if serviceID == "" {
		return "", fmt.Errorf("service_id cannot be empty")
	}

	return getBranchKeyId(serviceID), nil
}
