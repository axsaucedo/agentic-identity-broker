package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	pgx "github.com/jackc/pgx/v5"

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
		`INSERT INTO signing_keys (id, kid, algorithm, private_key_encrypted, is_current, activates_at, created_at, removed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		key.ID, key.KID, key.Algorithm, key.PrivateKeyEncrypted,
		key.IsCurrent, key.ActivatesAt, key.CreatedAt, key.RemovedAt,
	)
	if err != nil {
		return classifySigningKeyRepoError("SigningKeyRepo.Create", err, "failed to create signing key")
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
		return classifySigningKeyRepoError("SigningKeyRepo.CreateAndSetCurrent", err, "failed to begin transaction")
	}
	defer tx.Rollback() //nolint:errcheck

	// Demote all existing current keys before inserting the new one.
	_, err = tx.ExecContext(execCtx, `UPDATE signing_keys SET is_current = false WHERE is_current = true`)
	if err != nil {
		return classifySigningKeyRepoError("SigningKeyRepo.CreateAndSetCurrent", err, "failed to demote existing keys")
	}

	_, err = tx.ExecContext(execCtx,
		`INSERT INTO signing_keys (id, kid, algorithm, private_key_encrypted, is_current, activates_at, created_at, removed_at)
		 VALUES ($1, $2, $3, $4, true, $5, $6, $7)`,
		key.ID, key.KID, key.Algorithm, key.PrivateKeyEncrypted,
		key.ActivatesAt, key.CreatedAt, key.RemovedAt,
	)
	if err != nil {
		return classifySigningKeyRepoError("SigningKeyRepo.CreateAndSetCurrent", err, "failed to insert signing key")
	}

	if err := tx.Commit(); err != nil {
		return classifySigningKeyRepoError("SigningKeyRepo.CreateAndSetCurrent", err, "failed to commit transaction")
	}
	return nil
}

func (r *SigningKeyRepo) GetByKID(ctx context.Context, kid id.KeyID) (*storage.SigningKey, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError("SigningKeyRepo.GetByKID", storage.ErrorKindConnection, nil, "database not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	var key storage.SigningKey
	err := r.adapter.db.GetContext(queryCtx, &key,
		`SELECT id, kid, algorithm, private_key_encrypted, is_current, activates_at, created_at, removed_at
		 FROM signing_keys WHERE kid = $1 AND removed_at IS NULL`, kid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.NewStorageError("SigningKeyRepo.GetByKID", storage.ErrorKindNotFound, err, "signing key not found")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, storage.NewStorageError("SigningKeyRepo.GetByKID", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return nil, storage.NewStorageError("SigningKeyRepo.GetByKID", storage.ErrorKindConnection, err, "failed to query signing key")
	}
	return &key, nil
}

// GetCurrent returns the signing key to use for token issuance. It prefers the key flagged
// as is_current provided its activates_at has passed. If the current key is still in its
// grace period, it falls back to the most recently activated key, keeping token issuance
// uninterrupted while JWKS caches learn about the new key.
func (r *SigningKeyRepo) GetCurrent(ctx context.Context) (*storage.SigningKey, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError("SigningKeyRepo.GetCurrent", storage.ErrorKindConnection, nil, "database not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	var key storage.SigningKey
	err := r.adapter.db.GetContext(queryCtx, &key,
		`SELECT id, kid, algorithm, private_key_encrypted, is_current, activates_at, created_at, removed_at
		 FROM signing_keys
		 WHERE removed_at IS NULL AND activates_at <= NOW()
		 ORDER BY
		   CASE WHEN is_current THEN 0 ELSE 1 END,
		   activates_at DESC
		 LIMIT 1`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.NewStorageError("SigningKeyRepo.GetCurrent", storage.ErrorKindNotFound, err, "no current signing key")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, storage.NewStorageError("SigningKeyRepo.GetCurrent", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return nil, storage.NewStorageError("SigningKeyRepo.GetCurrent", storage.ErrorKindConnection, err, "failed to query current signing key")
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
		`SELECT id, kid, algorithm, private_key_encrypted, is_current, activates_at, created_at, removed_at
		 FROM signing_keys WHERE removed_at IS NULL ORDER BY created_at DESC`)
	if err != nil {
		return nil, classifySigningKeyRepoError("SigningKeyRepo.ListActive", err, "failed to list signing keys")
	}
	return keys, nil
}

// SetCurrent promotes a key to be the current signing key and sets activates_at = NOW()
// because an explicit admin promotion targets a key already present in the JWKS.
func (r *SigningKeyRepo) SetCurrent(ctx context.Context, kid id.KeyID) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("SigningKeyRepo.SetCurrent", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	tx, err := r.adapter.db.BeginTxx(execCtx, nil)
	if err != nil {
		return classifySigningKeyRepoError("SigningKeyRepo.SetCurrent", err, "failed to begin transaction")
	}
	defer tx.Rollback() //nolint:errcheck

	// Demote all current keys
	_, err = tx.ExecContext(execCtx, `UPDATE signing_keys SET is_current = false WHERE is_current = true`)
	if err != nil {
		return classifySigningKeyRepoError("SigningKeyRepo.SetCurrent", err, "failed to demote keys")
	}

	// Promote the target key and reset activates_at to NOW() so it starts signing immediately.
	result, err := tx.ExecContext(execCtx,
		`UPDATE signing_keys SET is_current = true, activates_at = NOW() WHERE kid = $1 AND removed_at IS NULL`, kid)
	if err != nil {
		return classifySigningKeyRepoError("SigningKeyRepo.SetCurrent", err, "failed to promote key")
	}
	if err := checkRowsAffected("SigningKeyRepo.SetCurrent", result, "signing key not found"); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return classifySigningKeyRepoError("SigningKeyRepo.SetCurrent", err, "failed to commit transaction")
	}
	return nil
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
		return classifySigningKeyRepoError("SigningKeyRepo.Delete", err, "failed to delete signing key")
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
		if errors.Is(err, context.DeadlineExceeded) {
			return 0, storage.NewStorageError("SigningKeyRepo.CountActive", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return 0, storage.NewStorageError("SigningKeyRepo.CountActive", storage.ErrorKindConnection, err, "failed to count signing keys")
	}
	return count, nil
}

func classifySigningKeyRepoError(operation string, err error, message string) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return storage.NewStorageError(operation, storage.ErrorKindTimeout, err, "operation exceeded timeout")
	}
	return storage.NewStorageError(operation, storage.ErrorKindConnection, err, message)
}
