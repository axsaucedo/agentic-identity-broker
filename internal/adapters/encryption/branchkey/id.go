package branchkey

import "fmt"

// IDFormat defines the deterministic format for branch key IDs.
// This is the single source of truth for branch key ID generation across all adapters.
const IDFormat = "service_%s_branch_key"

const (
	prefix = "service_"
	suffix = "_branch_key"
)

// GenerateBranchKeyId generates a deterministic branch key ID from a service ID.
// This function provides the single source of truth for branch key ID generation
// used by both provisioning and runtime phases.
//
// Format: service_{service_id}_branch_key
// Example: service_oauth2_branch_key, service_github_branch_key
//
// Parameters:
//   - serviceID: The service identifier (must be non-empty)
//
// Returns:
//   - string: The formatted branch key ID
func GenerateBranchKeyId(serviceID string) string {
	return fmt.Sprintf(IDFormat, serviceID)
}

// ExtractServiceID extracts the service ID from a branch key ID.
// This is the inverse operation of GenerateBranchKeyId, ensuring symmetric operations.
//
// Format: service_{service_id}_branch_key -> service_id
// Example: service_oauth2_branch_key -> oauth2
//
// Parameters:
//   - branchKeyID: The branch key ID to parse
//
// Returns:
//   - string: The extracted service ID, or empty string if parsing fails
//   - error: Non-nil if the branch key ID format is invalid
func ExtractServiceID(branchKeyID string) (string, error) {
	// Validate format and extract service ID
	if len(branchKeyID) > len(prefix)+len(suffix) &&
		branchKeyID[:len(prefix)] == prefix &&
		branchKeyID[len(branchKeyID)-len(suffix):] == suffix {
		serviceID := branchKeyID[len(prefix) : len(branchKeyID)-len(suffix)]
		if serviceID == "" {
			return "", fmt.Errorf("invalid branch key ID: empty service ID in %q", branchKeyID)
		}
		return serviceID, nil
	}

	// Special case: if the length is exactly what we'd expect for an empty service ID,
	// check if it matches the empty service ID format
	if branchKeyID == fmt.Sprintf(IDFormat, "") {
		return "", fmt.Errorf("invalid branch key ID: empty service ID in %q", branchKeyID)
	}

	return "", fmt.Errorf("invalid branch key ID format: %q (expected format: %s)", branchKeyID, IDFormat)
}

// ValidateBranchKeyID validates that a branch key ID follows the expected format.
// This is used for input validation in adapters and tests.
//
// Parameters:
//   - branchKeyID: The branch key ID to validate
//
// Returns:
//   - bool: true if the ID format is valid, false otherwise
func ValidateBranchKeyID(branchKeyID string) bool {
	_, err := ExtractServiceID(branchKeyID)
	return err == nil
}
