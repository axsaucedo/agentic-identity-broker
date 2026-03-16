package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/branchkey"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

// InMemoryBranchKeyRepository implements BranchKeyRepository and BranchKeyIdProvider interfaces for testing and development.
// It tracks provisioned branch keys in memory using a concurrent-safe map.
// Used when keyring type is not hierarchical or during testing when real DynamoDB access is not available.
type InMemoryBranchKeyRepository struct {
	mu          sync.RWMutex
	provisioned map[string]bool
}

// NewInMemoryBranchKeyRepository creates a new in-memory branch key repository.
func NewInMemoryBranchKeyRepository() *InMemoryBranchKeyRepository {
	return &InMemoryBranchKeyRepository{
		provisioned: make(map[string]bool),
	}
}

// Create creates a branch key for a service (in-memory).
// The branch key ID follows the pattern: service_{service_id}_branch_key
// Returns the generated branch key ID or error if validation fails.
func (r *InMemoryBranchKeyRepository) Create(ctx context.Context, serviceID id.ServiceID) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if serviceID.IsZero() {
		return "", fmt.Errorf("service_id cannot be empty")
	}

	branchKeyID := r.GenerateBranchKeyId(serviceID)
	r.provisioned[branchKeyID] = true
	return branchKeyID, nil
}

// Get retrieves a branch key ID for the given service (in-memory).
// Returns the branch key ID if provisioned, error if not found or validation fails.
func (r *InMemoryBranchKeyRepository) Get(ctx context.Context, serviceID id.ServiceID) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if serviceID.IsZero() {
		return "", fmt.Errorf("service_id cannot be empty")
	}

	branchKeyID := r.GenerateBranchKeyId(serviceID)
	if r.provisioned[branchKeyID] {
		return branchKeyID, nil
	}
	return "", fmt.Errorf("branch key not found for service_id: %s", serviceID)
}

// GenerateBranchKeyId generates a deterministic branch key ID from a service ID.
// Format: service_{service_id}_branch_key
// Example: service_oauth2_branch_key, service_github_branch_key
func (r *InMemoryBranchKeyRepository) GenerateBranchKeyId(serviceID id.ServiceID) string {
	return branchkey.GenerateBranchKeyId(serviceID.String())
}

// ExtractServiceIdFromBranchKey extracts the service ID from a branch key ID.
// This is the inverse operation of GenerateBranchKeyId.
// Format: service_{service_id}_branch_key -> service_id
// Example: service_oauth2_branch_key -> oauth2
// Returns empty string if parsing fails.
func (r *InMemoryBranchKeyRepository) ExtractServiceIdFromBranchKey(branchKeyID string) id.ServiceID {
	serviceID, err := branchkey.ExtractServiceID(branchKeyID)
	if err != nil {
		return id.ServiceID{} // Interface contract: return empty string on parsing failure
	}
	parsed, pErr := id.ParseServiceID(serviceID)
	if pErr != nil {
		return id.ServiceID{}
	}
	return parsed
}
