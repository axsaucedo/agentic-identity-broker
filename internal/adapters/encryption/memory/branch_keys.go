package memory

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryBranchKeyRepository implements ports.BranchKeyManager for testing and development.
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
func (r *InMemoryBranchKeyRepository) Create(ctx context.Context, serviceID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if serviceID == "" {
		return "", fmt.Errorf("service_id cannot be empty")
	}

	branchKeyID := getBranchKeyId(serviceID)
	r.provisioned[branchKeyID] = true
	return branchKeyID, nil
}

// Get retrieves a branch key ID for the given service (in-memory).
// Returns the branch key ID if provisioned, error if not found or validation fails.
func (r *InMemoryBranchKeyRepository) Get(ctx context.Context, serviceID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if serviceID == "" {
		return "", fmt.Errorf("service_id cannot be empty")
	}

	branchKeyID := getBranchKeyId(serviceID)
	if r.provisioned[branchKeyID] {
		return branchKeyID, nil
	}
	return "", fmt.Errorf("branch key not found for service_id: %s", serviceID)
}

func getBranchKeyId(serviceID string) string {
	return fmt.Sprintf("service_%s_branch_key", serviceID)
}
