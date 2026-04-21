package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Compile-time interface check
var _ ports.BrokerClientCredentialRepository = (*BrokerClientCredentialRepo)(nil)

// BrokerClientCredentialRepo implements BrokerClientCredentialRepository using PostgreSQL.
type BrokerClientCredentialRepo struct {
	adapter *Adapter
}

// NewBrokerClientCredentialRepo creates a new PostgreSQL broker client credential repository.
func NewBrokerClientCredentialRepo(adapter *Adapter) *BrokerClientCredentialRepo {
	return &BrokerClientCredentialRepo{adapter: adapter}
}

func (r *BrokerClientCredentialRepo) Create(ctx context.Context, credential *storage.BrokerClientCredential) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("BrokerClientCredentialRepo.Create", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	_, err := r.adapter.db.ExecContext(execCtx,
		`INSERT INTO client_credentials (id, client_id, secret_hash, created_at, rotated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		credential.ID, credential.ClientID,
		credential.SecretHash, credential.CreatedAt, credential.RotatedAt,
	)
	if err != nil {
		return storage.NewStorageError("BrokerClientCredentialRepo.Create", storage.ErrorKindUnknown, err, "failed to create credential")
	}
	return nil
}

// GetByAgentID retrieves the credential for the given agent.
// Since client_id == agentID.String(), this delegates to GetByClientID.
func (r *BrokerClientCredentialRepo) GetByAgentID(ctx context.Context, agentID id.AgentID) (*storage.BrokerClientCredential, error) {
	return r.GetByClientID(ctx, id.ClientID(agentID.String()))
}

func (r *BrokerClientCredentialRepo) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.BrokerClientCredential, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError("BrokerClientCredentialRepo.GetByClientID", storage.ErrorKindConnection, nil, "database not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	var cred storage.BrokerClientCredential
	err := r.adapter.db.GetContext(queryCtx, &cred,
		`SELECT id, client_id, secret_hash, created_at, rotated_at
		 FROM client_credentials WHERE client_id = $1`, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.NewStorageError("BrokerClientCredentialRepo.GetByClientID", storage.ErrorKindNotFound, err, "credential not found")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, storage.NewStorageError("BrokerClientCredentialRepo.GetByClientID", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return nil, storage.NewStorageError("BrokerClientCredentialRepo.GetByClientID", storage.ErrorKindConnection, err, "failed to query credential")
	}
	return &cred, nil
}

func (r *BrokerClientCredentialRepo) Delete(ctx context.Context, agentID id.AgentID) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("BrokerClientCredentialRepo.Delete", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	result, err := r.adapter.db.ExecContext(execCtx,
		`DELETE FROM client_credentials WHERE client_id = $1`, agentID.String())
	if err != nil {
		return storage.NewStorageError("BrokerClientCredentialRepo.Delete", storage.ErrorKindUnknown, err, "failed to delete credential")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return storage.NewStorageError("BrokerClientCredentialRepo.Delete", storage.ErrorKindNotFound, nil, "credential not found")
	}
	return nil
}

func (r *BrokerClientCredentialRepo) Rotate(ctx context.Context, agentID id.AgentID, newCredential *storage.BrokerClientCredential) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("BrokerClientCredentialRepo.Rotate", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	tx, err := r.adapter.db.BeginTxx(execCtx, nil)
	if err != nil {
		return storage.NewStorageError("BrokerClientCredentialRepo.Rotate", storage.ErrorKindUnknown, err, "failed to begin transaction")
	}
	defer func() { _ = tx.Rollback() }()

	// DELETE before INSERT: client_id is UNIQUE, so we must remove the old row first.
	// The transaction guarantees atomicity: if the INSERT fails the DELETE rolls back.
	result, err := tx.ExecContext(execCtx,
		`DELETE FROM client_credentials WHERE client_id = $1`, agentID.String())
	if err != nil {
		return storage.NewStorageError("BrokerClientCredentialRepo.Rotate", storage.ErrorKindUnknown, err, "failed to delete old credential")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return storage.NewStorageError("BrokerClientCredentialRepo.Rotate", storage.ErrorKindNotFound, nil,
			"no existing credential found for agent")
	}

	_, err = tx.ExecContext(execCtx,
		`INSERT INTO client_credentials (id, client_id, secret_hash, created_at, rotated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		newCredential.ID, newCredential.ClientID,
		newCredential.SecretHash, newCredential.CreatedAt, newCredential.RotatedAt,
	)
	if err != nil {
		return storage.NewStorageError("BrokerClientCredentialRepo.Rotate", storage.ErrorKindUnknown, err, "failed to insert new credential")
	}

	if err := tx.Commit(); err != nil {
		return storage.NewStorageError("BrokerClientCredentialRepo.Rotate", storage.ErrorKindUnknown, err, "failed to commit rotation transaction")
	}
	return nil
}
