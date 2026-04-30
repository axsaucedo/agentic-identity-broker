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
	if r.adapter.db == nil {
		return storage.NewStorageError("AuthorizationSessionRepo.Consume", storage.ErrorKindConnection, nil, "database not initialized")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	now := time.Now()
	result, err := r.adapter.db.ExecContext(execCtx,
		`UPDATE authorization_sessions SET consumed_at = $1 WHERE session_id = $2 AND consumed_at IS NULL AND expires_at >= $1`,
		now, sessionID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return storage.NewStorageError("AuthorizationSessionRepo.Consume", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return storage.NewStorageError("AuthorizationSessionRepo.Consume", storage.ErrorKindUnknown, err, "failed to consume authorization session")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return storage.NewStorageError("AuthorizationSessionRepo.Consume", storage.ErrorKindUnknown, err, "failed to determine rows affected")
	}
	if rows > 0 {
		return nil
	}

	// 0 rows: session missing, expired, or already consumed — distinguish via SELECT.
	queryCtx, cancel2 := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel2()

	var consumedAt *time.Time
	var expiresAt time.Time
	err = r.adapter.db.QueryRowContext(queryCtx,
		`SELECT consumed_at, expires_at FROM authorization_sessions WHERE session_id = $1`, sessionID).Scan(&consumedAt, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
		return storage.NewStorageError("AuthorizationSessionRepo.Consume", storage.ErrorKindNotFound, nil, "authorization session not found")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return storage.NewStorageError("AuthorizationSessionRepo.Consume", storage.ErrorKindTimeout, err, "timed out checking authorization session state")
	}
	if err != nil {
		return storage.NewStorageError("AuthorizationSessionRepo.Consume", storage.ErrorKindConnection, err, "failed to check authorization session state")
	}
	if expiresAt.Before(now) {
		return storage.NewStorageError("AuthorizationSessionRepo.Consume", storage.ErrorKindNotFound, nil, "authorization session has expired")
	}
	return storage.NewStorageError("AuthorizationSessionRepo.Consume", storage.ErrorKindConflict, nil, "authorization session has already been consumed")
}

func (r *AuthorizationSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
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
	rows, _ := result.RowsAffected()
	return rows, nil
}
