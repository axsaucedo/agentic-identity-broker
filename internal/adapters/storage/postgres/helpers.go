package postgres

import (
	"database/sql"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// checkRowsAffected verifies that a SQL result affected at least one row.
// Returns ErrorKindUnknown if the driver reports an error from RowsAffected(),
// ErrorKindNotFound if zero rows were affected, and nil otherwise.
func checkRowsAffected(op string, result sql.Result, notFoundMsg string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return storage.NewStorageError(op, storage.ErrorKindUnknown, err, "failed to determine rows affected")
	}
	if rows == 0 {
		return storage.NewStorageError(op, storage.ErrorKindNotFound, nil, notFoundMsg)
	}
	return nil
}
