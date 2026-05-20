package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// verifyPermissionSetExistenceInTx validates that all requested permission set IDs
// exist within an active transaction, acquiring a share lock (FOR SHARE) on each row.
// The share lock prevents concurrent DELETE transactions from removing referenced PSes
// before the caller's write commits, closing the write/delete TOCTOU race window.
// Returns ErrorKindConflict with a descriptive error if any IDs are missing.
func verifyPermissionSetExistenceInTx(ctx context.Context, tx *sql.Tx, psIDs []id.PermissionSetID) error {
	if len(psIDs) == 0 {
		return nil
	}
	idStrings := make([]string, len(psIDs))
	for i, psID := range psIDs {
		idStrings[i] = psID.String()
	}
	rows, err := tx.QueryContext(ctx,
		`SELECT id FROM permission_sets WHERE id = ANY($1::uuid[]) FOR SHARE`,
		pq.Array(idStrings),
	)
	if err != nil {
		return storage.NewStorageError("verifyPermissionSetExistence", storage.ErrorKindUnknown, err, "failed to lock permission set rows")
	}
	defer func() { _ = rows.Close() }()
	found := make(map[string]bool, len(psIDs))
	for rows.Next() {
		var foundID string
		if err := rows.Scan(&foundID); err != nil {
			return storage.NewStorageError("verifyPermissionSetExistence", storage.ErrorKindUnknown, err, "failed to scan permission set row")
		}
		found[foundID] = true
	}
	if err := rows.Err(); err != nil {
		return storage.NewStorageError("verifyPermissionSetExistence", storage.ErrorKindUnknown, err, "error iterating permission set rows")
	}
	for _, psID := range psIDs {
		if !found[psID.String()] {
			return storage.NewStorageError(
				"verifyPermissionSetExistence",
				storage.ErrorKindConflict,
				fmt.Errorf("permission set %s not found", psID),
				fmt.Sprintf("permission set %s was deleted concurrently", psID),
			)
		}
	}
	return nil
}

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
