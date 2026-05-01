package postgres

import (
	"errors"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rowsAffectedErrorResult struct{}

func (rowsAffectedErrorResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (rowsAffectedErrorResult) RowsAffected() (int64, error) {
	return 0, errors.New("rows affected unavailable")
}

func TestRowsAffectedCount_ReturnsStorageError(t *testing.T) {
	rows, err := rowsAffectedCount(rowsAffectedErrorResult{}, "AuthorizationSessionRepo.DeleteExpired")
	require.Error(t, err)
	assert.Zero(t, rows)

	var storageErr *storage.StorageError
	require.ErrorAs(t, err, &storageErr)
	assert.Equal(t, storage.ErrorKindUnknown, storageErr.Kind)
	assert.Contains(t, err.Error(), "failed to determine rows affected")
}
