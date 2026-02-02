package aws

import (
	"fmt"

	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/branchkey"
)

// BranchKeyIdSupplier implements IBranchKeyIdSupplier for the AWS hierarchical keyring.
// It resolves branch key IDs based on the service_id from the encryption context.
// This is required by the AWS Encryption SDK's hierarchical keyring at runtime.
//
// The supplier delegates to the branchkey package for the actual ID generation,
// ensuring a single source of truth for branch key ID formatting.
// It also implements the BranchKeyIdProvider port interface for cross-adapter compatibility.
type BranchKeyIdSupplier struct{}

// GetBranchKeyId returns a deterministic branch key identifier based on service_id from encryption context.
// Called by the AWS Encryption SDK hierarchical keyring during encryption/decryption.
//
// Parameters:
//   - input.EncryptionContext: Must contain "service_id" key with service identifier
//
// Returns:
//   - BranchKeyId: "service_{service_id}_branch_key" (e.g., "service_oauth2_branch_key")
//   - error: If encryption context is missing or invalid
func (d *BranchKeyIdSupplier) GetBranchKeyId(input mpltypes.GetBranchKeyIdInput) (*mpltypes.GetBranchKeyIdOutput, error) {
	// Extract service_id from encryption context
	ec := input.EncryptionContext
	serviceID, exists := ec["service_id"]

	// Validate service_id is present and non-empty
	if !exists || serviceID == "" {
		return nil, fmt.Errorf("encryption context missing or empty required key 'service_id'")
	}

	// Use deterministic ID generation (single source of truth)
	branchKeyID := branchkey.GenerateBranchKeyId(serviceID)

	return &mpltypes.GetBranchKeyIdOutput{
		BranchKeyId: branchKeyID,
	}, nil
}

// GenerateBranchKeyId generates a deterministic branch key ID from a service ID.
// Format: service_{service_id}_branch_key
// Example: service_oauth2_branch_key, service_github_branch_key
func (d *BranchKeyIdSupplier) GenerateBranchKeyId(serviceID string) string {
	return branchkey.GenerateBranchKeyId(serviceID)
}

// ExtractServiceIdFromBranchKey extracts the service ID from a branch key ID.
// This is the inverse operation of GenerateBranchKeyId.
// Format: service_{service_id}_branch_key -> service_id
// Example: service_oauth2_branch_key -> oauth2
// Returns empty string if parsing fails.
func (d *BranchKeyIdSupplier) ExtractServiceIdFromBranchKey(branchKeyID string) string {
	serviceID, err := branchkey.ExtractServiceID(branchKeyID)
	if err != nil {
		return "" // Interface contract: return empty string on parsing failure
	}
	return serviceID
}
