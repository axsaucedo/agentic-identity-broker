package noop

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

// BranchKeyManager is a null-object BranchKeyManager.
// Injected when no real branch key store is configured (memory / raw-AES backend),
// eliminating nil guards in domain services.
type BranchKeyManager struct{}

func (n *BranchKeyManager) Create(_ context.Context, _ id.ServiceID) (string, error) {
	return "", nil
}
