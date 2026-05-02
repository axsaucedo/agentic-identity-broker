package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Compile-time interface check
var _ ports.AuthorizationSessionRepository = (*AuthorizationSessionRepo)(nil)

// AuthorizationSessionRepo implements AuthorizationSessionRepository using PostgreSQL.
type AuthorizationSessionRepo struct {
	adapter *Adapter
}

// NewAuthorizationSessionRepo creates a new PostgreSQL authorization session repository.
func NewAuthorizationSessionRepo(adapter *Adapter) *AuthorizationSessionRepo {
	return &AuthorizationSessionRepo{adapter: adapter}
}

func (r *AuthorizationSessionRepo) Create(ctx context.Context, session *storage.AuthorizationSession) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("AuthorizationSessionRepo.Create", storage.ErrorKindConnection, nil, "database not initialized")
	}

	var metaJSON []byte
	if session.CIMDMetadata != nil {
		var err error
		metaJSON, err = json.Marshal(session.CIMDMetadata)
		if err != nil {
			return storage.NewStorageError("AuthorizationSessionRepo.Create", storage.ErrorKindUnknown, err, "failed to marshal cimd_metadata")
		}
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	_, err := r.adapter.db.ExecContext(execCtx,
		`INSERT INTO authorization_sessions
		 (session_id, agent_id, principal, client_id, original_url, redirect_uri, scope, state,
		  code_challenge, code_challenge_method, cimd_metadata, created_at, expires_at, consumed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		session.SessionID, session.AgentID, session.Principal, session.ClientID, session.OriginalURL,
		session.RedirectURI, session.Scope, session.State,
		session.CodeChallenge, session.CodeChallengeMethod,
		metaJSON, session.CreatedAt, session.ExpiresAt, session.ConsumedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return storage.NewStorageError("AuthorizationSessionRepo.Create", storage.ErrorKindConflict, err, "authorization session already exists")
		}
		return storage.NewStorageError("AuthorizationSessionRepo.Create", storage.ErrorKindUnknown, err, "failed to create authorization session")
	}
	return nil
}

func (r *AuthorizationSessionRepo) GetBySessionID(ctx context.Context, sessionID string) (*storage.AuthorizationSession, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError("AuthorizationSessionRepo.GetBySessionID", storage.ErrorKindConnection, nil, "database not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	var row struct {
		storage.AuthorizationSession
		CIMDMetadataJSON []byte `db:"cimd_metadata"`
	}
	err := r.adapter.db.GetContext(queryCtx, &row,
		`SELECT session_id, agent_id, principal, client_id, original_url, redirect_uri, scope, state,
		        code_challenge, code_challenge_method, cimd_metadata, created_at, expires_at, consumed_at
		 FROM authorization_sessions WHERE session_id = $1`, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.NewStorageError("AuthorizationSessionRepo.GetBySessionID", storage.ErrorKindNotFound, err, "authorization session not found")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, storage.NewStorageError("AuthorizationSessionRepo.GetBySessionID", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return nil, storage.NewStorageError("AuthorizationSessionRepo.GetBySessionID", storage.ErrorKindConnection, err, "failed to query authorization session")
	}

	session := row.AuthorizationSession
	if len(row.CIMDMetadataJSON) > 0 {
		var meta storage.CIMDMetadataSnapshot
		if err := json.Unmarshal(row.CIMDMetadataJSON, &meta); err != nil {
			return nil, storage.NewStorageError("AuthorizationSessionRepo.GetBySessionID", storage.ErrorKindUnknown, err, "failed to unmarshal cimd_metadata")
		}
		session.CIMDMetadata = &meta
	}

	return &session, nil
}

func (r *AuthorizationSessionRepo) Consume(ctx context.Context, sessionID string) error {
	return r.ConsumeIf(ctx, sessionID, nil)
}

func (r *AuthorizationSessionRepo) ConsumeIf(ctx context.Context, sessionID string, fn ports.AuthorizationSessionMutation) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	tx, err := r.adapter.db.BeginTxx(execCtx, nil)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindTimeout, err, "failed to begin authorization session transaction")
		}
		return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindConnection, err, "failed to begin authorization session transaction")
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	var consumedAt *time.Time
	var expiresAt time.Time
	err = tx.QueryRowContext(execCtx,
		`SELECT consumed_at, expires_at FROM authorization_sessions WHERE session_id = $1 FOR UPDATE`,
		sessionID,
	).Scan(&consumedAt, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindNotFound, nil, "authorization session not found")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindConnection, err, "failed to lock authorization session")
	}

	if expiresAt.Before(now) {
		return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindNotFound, nil, "authorization session has expired")
	}
	if consumedAt != nil {
		return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindConflict, nil, "authorization session has already been consumed")
	}

	if fn != nil {
		if err := fn(contextWithTx(execCtx, tx)); err != nil {
			return err
		}
	}

	result, err := tx.ExecContext(execCtx,
		`UPDATE authorization_sessions SET consumed_at = $1 WHERE session_id = $2`,
		now, sessionID)
	if errors.Is(err, context.DeadlineExceeded) {
		return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindTimeout, err, "operation exceeded timeout")
	}
	if err != nil {
		return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindUnknown, err, "failed to consume authorization session")
	}

	rows, err := rowsAffectedCount(result, "AuthorizationSessionRepo.ConsumeIf")
	if err != nil {
		return err
	}
	if rows != 1 {
		return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindUnknown, nil, "failed to consume authorization session")
	}

	if err := tx.Commit(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindTimeout, err, "failed to commit authorization session transaction")
		}
		return storage.NewStorageError("AuthorizationSessionRepo.ConsumeIf", storage.ErrorKindUnknown, err, "failed to commit authorization session transaction")
	}

	return nil
}

func (r *AuthorizationSessionRepo) DeleteExpired(ctx context.Context) (int, error) {
	if r.adapter.db == nil {
		return 0, storage.NewStorageError("AuthorizationSessionRepo.DeleteExpired", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	now := time.Now()
	result, err := r.adapter.db.ExecContext(execCtx,
		`DELETE FROM authorization_sessions WHERE expires_at < $1`, now)
	if err != nil {
		return 0, storage.NewStorageError("AuthorizationSessionRepo.DeleteExpired", storage.ErrorKindUnknown, err, "failed to delete expired authorization sessions")
	}
	rows, err := rowsAffectedCount(result, "AuthorizationSessionRepo.DeleteExpired")
	if err != nil {
		return 0, err
	}
	return int(rows), nil
}

func rowsAffectedCount(result sql.Result, operation string) (int64, error) {
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, storage.NewStorageError(operation, storage.ErrorKindUnknown, err, "failed to determine rows affected")
	}
	return rows, nil
}
