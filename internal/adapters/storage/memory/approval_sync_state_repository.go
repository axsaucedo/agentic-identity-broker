package memory

import (
	"context"
	"sync/atomic"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// ApprovalSyncStateRepository implements ApprovalSyncStateRepository in-memory.
type ApprovalSyncStateRepository struct {
	version atomic.Int64
}

// NewApprovalSyncStateRepository creates a new in-memory ApprovalSyncStateRepository.
func NewApprovalSyncStateRepository() *ApprovalSyncStateRepository {
	return &ApprovalSyncStateRepository{}
}

var _ ports.ApprovalSyncStateRepository = (*ApprovalSyncStateRepository)(nil)

func (r *ApprovalSyncStateRepository) GetVersion(_ context.Context) (int64, error) {
	return r.version.Load(), nil
}

func (r *ApprovalSyncStateRepository) IncrementVersion(_ context.Context) (int64, error) {
	return r.version.Add(1), nil
}
