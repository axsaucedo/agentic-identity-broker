package postgres

import (
	"context"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Compile-time interface check
var _ ports.SigningKeyRepository = (*SigningKeyRepo)(nil)

// SigningKeyRepo implements SigningKeyRepository using PostgreSQL.
type SigningKeyRepo struct {
	adapter *Adapter
}

// NewSigningKeyRepo creates a new PostgreSQL signing key repository.
func NewSigningKeyRepo(adapter *Adapter) *SigningKeyRepo {
	return &SigningKeyRepo{adapter: adapter}
}

func (r *SigningKeyRepo) Create(ctx context.Context, key *storage.SigningKey) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("SigningKeyRepo.Create", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	_, err := r.adapter.db.ExecContext(execCtx,
		`INSERT INTO signing_keys (id, kid, algorithm, private_key_encrypted, is_current, created_at, removed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		key.ID, key.KID, key.Algorithm, key.PrivateKeyEncrypted,
		key.IsCurrent, key.CreatedAt, key.RemovedAt,
	)
	if err != nil {
		return storage.NewStorageError("SigningKeyRepo.Create", storage.ErrorKindUnknown, err, "failed to create signing key")
	}
	return nil
}

func (r *SigningKeyRepo) CreateAndSetCurrent(ctx context.Context, key *storage.SigningKey) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("SigningKeyRepo.CreateAndSetCurrent", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	tx, err := r.adapter.db.BeginTxx(execCtx, nil)
	if err != nil {
		return storage.NewStorageError("SigningKeyRepo.CreateAndSetCurrent", storage.ErrorKindUnknown, err, "failed to begin transaction")
	}
	defer tx.Rollback() //nolint:errcheck

	// Demote all existing current keys before inserting the new one.
	_, err = tx.ExecContext(execCtx, `UPDATE signing_keys SET is_current = false WHERE is_current = true`)
	if err != nil {
		return storage.NewStorageError("SigningKeyRepo.CreateAndSetCurrent", storage.ErrorKindUnknown, err, "failed to demote existing keys")
	}

	_, err = tx.ExecContext(execCtx,
		`INSERT INTO signing_keys (id, kid, algorithm, private_key_encrypted, is_current, created_at, removed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		key.ID, key.KID, key.Algorithm, key.PrivateKeyEncrypted,
		true, key.CreatedAt, key.RemovedAt,
	)
	if err != nil {
		return storage.NewStorageError("SigningKeyRepo.CreateAndSetCurrent", storage.ErrorKindUnknown, err, "failed to insert signing key")
	}

	return tx.Commit()
}

func (r *SigningKeyRepo) GetByKID(ctx context.Context, kid id.KeyID) (*storage.SigningKey, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError("SigningKeyRepo.GetByKID", storage.ErrorKindConnection, nil, "database not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	var key storage.SigningKey
	err := r.adapter.db.GetContext(queryCtx, &key,
		`SELECT id, kid, algorithm, private_key_encrypted, is_current, created_at, removed_at
		 FROM signing_keys WHERE kid = $1 AND removed_at IS NULL`, kid)
	if err != nil {
		return nil, storage.NewStorageError("SigningKeyRepo.GetByKID", storage.ErrorKindNotFound, err, "signing key not found")
	}
	return &key, nil
}

func (r *SigningKeyRepo) GetCurrent(ctx context.Context) (*storage.SigningKey, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError("SigningKeyRepo.GetCurrent", storage.ErrorKindConnection, nil, "database not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	var key storage.SigningKey
	err := r.adapter.db.GetContext(queryCtx, &key,
		`SELECT id, kid, algorithm, private_key_encrypted, is_current, created_at, removed_at
		 FROM signing_keys WHERE is_current = true AND removed_at IS NULL`)
	if err != nil {
		return nil, storage.NewStorageError("SigningKeyRepo.GetCurrent", storage.ErrorKindNotFound, err, "no current signing key")
	}
	return &key, nil
}

func (r *SigningKeyRepo) ListActive(ctx context.Context) ([]*storage.SigningKey, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError("SigningKeyRepo.ListActive", storage.ErrorKindConnection, nil, "database not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	var keys []*storage.SigningKey
	err := r.adapter.db.SelectContext(queryCtx, &keys,
		`SELECT id, kid, algorithm, private_key_encrypted, is_current, created_at, removed_at
		 FROM signing_keys WHERE removed_at IS NULL ORDER BY created_at DESC`)
	if err != nil {
		return nil, storage.NewStorageError("SigningKeyRepo.ListActive", storage.ErrorKindUnknown, err, "failed to list signing keys")
	}
	return keys, nil
}

func (r *SigningKeyRepo) SetCurrent(ctx context.Context, kid id.KeyID) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("SigningKeyRepo.SetCurrent", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	tx, err := r.adapter.db.BeginTxx(execCtx, nil)
	if err != nil {
		return storage.NewStorageError("SigningKeyRepo.SetCurrent", storage.ErrorKindUnknown, err, "failed to begin transaction")
	}
	defer tx.Rollback() //nolint:errcheck

	// Demote all current keys
	_, err = tx.ExecContext(execCtx, `UPDATE signing_keys SET is_current = false WHERE is_current = true`)
	if err != nil {
		return storage.NewStorageError("SigningKeyRepo.SetCurrent", storage.ErrorKindUnknown, err, "failed to demote keys")
	}

	// Promote the target key
	result, err := tx.ExecContext(execCtx,
		`UPDATE signing_keys SET is_current = true WHERE kid = $1 AND removed_at IS NULL`, kid)
	if err != nil {
		return storage.NewStorageError("SigningKeyRepo.SetCurrent", storage.ErrorKindUnknown, err, "failed to promote key")
	}
	if err := checkRowsAffected("SigningKeyRepo.SetCurrent", result, "signing key not found"); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SigningKeyRepo) Delete(ctx context.Context, kid id.KeyID) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("SigningKeyRepo.Delete", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	now := time.Now()
	result, err := r.adapter.db.ExecContext(execCtx,
		`UPDATE signing_keys SET removed_at = $1 WHERE kid = $2 AND removed_at IS NULL`, now, kid)
	if err != nil {
		return storage.NewStorageError("SigningKeyRepo.Delete", storage.ErrorKindUnknown, err, "failed to delete signing key")
	}
	return checkRowsAffected("SigningKeyRepo.Delete", result, "signing key not found")
}

func (r *SigningKeyRepo) CountActive(ctx context.Context) (int, error) {
	if r.adapter.db == nil {
		return 0, storage.NewStorageError("SigningKeyRepo.CountActive", storage.ErrorKindConnection, nil, "database not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	var count int
	err := r.adapter.db.GetContext(queryCtx, &count,
		`SELECT COUNT(*) FROM signing_keys WHERE removed_at IS NULL`)
	if err != nil {
		return 0, storage.NewStorageError("SigningKeyRepo.CountActive", storage.ErrorKindUnknown, err, "failed to count signing keys")
	}
	return count, nil
}
