package aws

import (
	"fmt"

	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
)

// DynamicBranchKeySupplier implements IBranchKeyIdSupplier using dynamic service-id based mapping.
// Each service_id from the encryption context maps to a deterministic branch key identifier,
// enabling unlimited services without hardcoded mappings.
//
// Branch Key ID Generation:
//   service_id "oauth2" → branch key "service_oauth2_branch_key"
//   service_id "github" → branch key "service_github_branch_key"
//   service_id "microsoft" → branch key "service_microsoft_branch_key"
//
// This ensures cryptographic isolation between services:
// - Each service gets its own branch key
// - Tokens encrypted with one service's context cannot be decrypted with another's
// - Adding new services requires no code changes, only service creation
//
// Requirements:
// - Encryption context MUST contain "service_id" key with non-empty value
// - Branch key must be pre-created in DynamoDB before first use
//
// Architecture:
// - Encryption context binding: service_id acts as key context
// - Branch key naming: Deterministic and service-specific
// - Scalability: Supports unlimited number of services
type DynamicBranchKeySupplier struct{}

// GetBranchKeyId returns a deterministic branch key identifier based on service_id from encryption context.
//
// Parameters:
//   - input.EncryptionContext: Must contain "service_id" key with service identifier
//
// Returns:
//   - BranchKeyId: "service_{service_id}_branch_key" (e.g., "service_oauth2_branch_key")
//   - error: If encryption context is missing or invalid
//
// Error Cases:
//   - Missing "service_id" key in encryption context
//   - Empty "service_id" value
//
// The branch key ID is deterministic: same service_id always produces same branch key ID.
// This enables proper context binding at both DEK and KEK layers.
func (d *DynamicBranchKeySupplier) GetBranchKeyId(input mpltypes.GetBranchKeyIdInput) (*mpltypes.GetBranchKeyIdOutput, error) {
	// Extract service_id from encryption context
	// This is the cryptographic binding key for the service
	ec := input.EncryptionContext
	serviceID, exists := ec["service_id"]

	// Validate service_id is present and non-empty
	if !exists || serviceID == "" {
		return nil, fmt.Errorf("encryption context missing or empty required key 'service_id'")
	}

	// Generate deterministic branch key identifier
	// Format: service_{service_id}_branch_key
	// Examples: service_oauth2_branch_key, service_github_branch_key
	branchKeyID := fmt.Sprintf("service_%s_branch_key", serviceID)

	return &mpltypes.GetBranchKeyIdOutput{
		BranchKeyId: branchKeyID,
	}, nil
}
