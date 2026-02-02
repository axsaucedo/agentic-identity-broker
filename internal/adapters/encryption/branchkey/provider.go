package branchkey

// DefaultProvider implements the ports.BranchKeyIdProvider interface
// using the shared branch key ID domain logic from id.go.
// This provides a concrete implementation that can be used by any adapter
// that needs branch key ID generation and parsing capabilities.
type DefaultProvider struct{}

// NewDefaultProvider creates a new DefaultProvider instance.
// This provider uses the standard branch key ID format defined in this package.
func NewDefaultProvider() *DefaultProvider {
	return &DefaultProvider{}
}

// GenerateBranchKeyId generates a deterministic branch key ID from a service ID.
// Implements ports.BranchKeyIdProvider.GenerateBranchKeyId()
//
// Format: service_{service_id}_branch_key
// Example: service_oauth2_branch_key, service_github_branch_key
//
// Parameters:
//   - serviceID: The service identifier
//
// Returns:
//   - string: The formatted branch key ID
func (p *DefaultProvider) GenerateBranchKeyId(serviceID string) string {
	return GenerateBranchKeyId(serviceID)
}

// ExtractServiceIdFromBranchKey extracts the service ID from a branch key ID.
// Implements ports.BranchKeyIdProvider.ExtractServiceIdFromBranchKey()
// This is the inverse operation of GenerateBranchKeyId.
//
// Format: service_{service_id}_branch_key -> service_id
// Example: service_oauth2_branch_key -> oauth2
//
// Parameters:
//   - branchKeyID: The branch key ID to parse
//
// Returns:
//   - string: The extracted service ID, or empty string if parsing fails
func (p *DefaultProvider) ExtractServiceIdFromBranchKey(branchKeyID string) string {
	serviceID, err := ExtractServiceID(branchKeyID)
	if err != nil {
		return "" // Interface contract: return empty string on parsing failure
	}
	return serviceID
}
