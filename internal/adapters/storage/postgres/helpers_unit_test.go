package postgres

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// mockSQLResult is a minimal database/sql.Result implementation for unit testing.
type mockSQLResult struct {
	rowsAffected    int64
	rowsAffectedErr error
}

func (m mockSQLResult) LastInsertId() (int64, error) { return 0, nil }
func (m mockSQLResult) RowsAffected() (int64, error) { return m.rowsAffected, m.rowsAffectedErr }

func TestCheckRowsAffected(t *testing.T) {
	const op = "Repo.TestOp"
	const notFoundMsg = "entity not found"

	t.Run("driver error returns ErrorKindUnknown", func(t *testing.T) {
		driverErr := errors.New("driver: result not supported")
		err := checkRowsAffected(op, mockSQLResult{rowsAffectedErr: driverErr}, notFoundMsg)
		require.Error(t, err)

		var se *storage.StorageError
		require.True(t, errors.As(err, &se))
		assert.Equal(t, storage.ErrorKindUnknown, se.Kind)
		assert.Equal(t, op, se.Operation)
		assert.ErrorIs(t, err, driverErr)
	})

	t.Run("zero rows returns ErrorKindNotFound", func(t *testing.T) {
		err := checkRowsAffected(op, mockSQLResult{rowsAffected: 0}, notFoundMsg)
		require.Error(t, err)

		var se *storage.StorageError
		require.True(t, errors.As(err, &se))
		assert.Equal(t, storage.ErrorKindNotFound, se.Kind)
		assert.Equal(t, op, se.Operation)
		assert.Equal(t, notFoundMsg, se.Message)
	})

	t.Run("one row returns nil", func(t *testing.T) {
		err := checkRowsAffected(op, mockSQLResult{rowsAffected: 1}, notFoundMsg)
		assert.NoError(t, err)
	})
}
