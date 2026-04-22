package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

func TestAuthorizationCodeRepo_FindByCodeHash_NilDB(t *testing.T) {
	repo := NewAuthorizationCodeRepo(&Adapter{db: nil})
	_, err := repo.FindByCodeHash(context.Background(), "somehash")
	require.Error(t, err)

	var se *storage.StorageError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, storage.ErrorKindConnection, se.Kind)
}
