package aws

import (
	"fmt"

	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
)

// BranchKeyIdSupplier implements IBranchKeyIdSupplier for the AWS hierarchical keyring.
// It resolves branch key IDs based on the service_id from the encryption context.
// This is required by the AWS Encryption SDK's hierarchical keyring at runtime.
//
// The supplier delegates to ResolveDeterministicBranchKeyID() for the actual ID generation,
// ensuring a single source of truth for branch key ID formatting.
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
	branchKeyID := getBranchKeyId(serviceID)

	return &mpltypes.GetBranchKeyIdOutput{
		BranchKeyId: branchKeyID,
	}, nil
}

// ResolveDeterministicBranchKeyID generates a deterministic branch key ID from a service_id.
// Single source of truth for branch key ID generation: service_{service_id}_branch_key
// This is exported for use by other implementations (e.g., in-memory provider for testing).
// Format: service_{service_id}_branch_key
// Example: service_oauth2_branch_key, service_github_branch_key
func getBranchKeyId(serviceID string) string {
	return fmt.Sprintf("service_%s_branch_key", serviceID)
}
