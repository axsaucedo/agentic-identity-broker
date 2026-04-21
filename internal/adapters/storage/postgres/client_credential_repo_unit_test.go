package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

func TestClientCredentialRepo_GetByAgentID_NilDB(t *testing.T) {
	repo := NewClientCredentialRepo(&Adapter{db: nil})
	_, err := repo.GetByAgentID(context.Background(), id.NewAgentID())
	require.Error(t, err)

	var se *storage.StorageError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, storage.ErrorKindConnection, se.Kind)
}

func TestClientCredentialRepo_GetByClientID_NilDB(t *testing.T) {
	repo := NewClientCredentialRepo(&Adapter{db: nil})
	_, err := repo.GetByClientID(context.Background(), id.NewClientID(id.NewAgentID().String()))
	require.Error(t, err)

	var se *storage.StorageError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, storage.ErrorKindConnection, se.Kind)
}
