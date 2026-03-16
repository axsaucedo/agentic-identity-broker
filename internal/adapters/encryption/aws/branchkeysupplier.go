package aws

import (
	"fmt"

	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/branchkey"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
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
//   - input.EncryptionContext: Must contain "service_id" key with a UUID service identifier
//
// Returns:
//   - BranchKeyId: "service_{service_id}_branch_key" (e.g., "service_550e8400-e29b-41d4-a716-446655440001_branch_key")
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
// Example: service_550e8400-e29b-41d4-a716-446655440001_branch_key
func (d *BranchKeyIdSupplier) GenerateBranchKeyId(serviceID id.ServiceID) string {
	return branchkey.GenerateBranchKeyId(serviceID.String())
}

// ExtractServiceIdFromBranchKey extracts the service ID from a branch key ID.
// This is the inverse operation of GenerateBranchKeyId.
// Format: service_{service_id}_branch_key -> service_id
// Example: service_550e8400-e29b-41d4-a716-446655440001_branch_key -> 550e8400-e29b-41d4-a716-446655440001
// Returns zero value (id.ServiceID{}) if parsing fails.
func (d *BranchKeyIdSupplier) ExtractServiceIdFromBranchKey(branchKeyID string) id.ServiceID {
	serviceID, err := branchkey.ExtractServiceID(branchKeyID)
	if err != nil {
		return id.ServiceID{} // zero value returned when branch key ID cannot be parsed
	}
	parsed, pErr := id.ParseServiceID(serviceID)
	if pErr != nil {
		return id.ServiceID{}
	}
	return parsed
}
