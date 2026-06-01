package noop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

func TestBranchKeyManagerCreate(t *testing.T) {
	manager := &BranchKeyManager{}
	subject := domainencryption.NewServiceBranchKeySubject(id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"))

	branchKeyID, err := manager.Create(context.Background(), subject)
	require.NoError(t, err)
	assert.Empty(t, branchKeyID)
}
