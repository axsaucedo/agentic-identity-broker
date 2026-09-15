package noop

import (
	"context"

	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Injected when no real branch key store is configured (memory / raw-AES backend),
// eliminating nil guards in domain services.
type BranchKeyManager struct{}

var _ ports.BranchKeyManager = (*BranchKeyManager)(nil)

func (n *BranchKeyManager) Create(_ context.Context, _ domainencryption.BranchKeySubject) (string, error) {
	return "", nil
}
