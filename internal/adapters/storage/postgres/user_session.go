package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// PostgresUserSessionRepository is a PostgreSQL implementation of UserSessionRepository.
type PostgresUserSessionRepository struct {
	adapter *Adapter
}

// NewUserSessionRepository creates a new PostgreSQL user session repository.
func NewUserSessionRepository(adapter *Adapter) ports.UserSessionRepository {
	return &PostgresUserSessionRepository{adapter: adapter}
}

// Create creates a new user session with upsert semantics.
func (r *PostgresUserSessionRepository) Create(ctx context.Context, session *storage.UserSession) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}
	if err := session.Validate(); err != nil {
		return err
	}

	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	query := `
		INSERT INTO user_sessions (
			id, principal, service_id, encrypted_access_token, encrypted_refresh_token,
			token_type, access_token_expires_at, refresh_token_expires_at, scope,
			encryption_context, initiated_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (principal, service_id) DO UPDATE SET
			encrypted_access_token = EXCLUDED.encrypted_access_token,
			encrypted_refresh_token = EXCLUDED.encrypted_refresh_token,
			token_type = EXCLUDED.token_type,
			access_token_expires_at = EXCLUDED.access_token_expires_at,
			refresh_token_expires_at = EXCLUDED.refresh_token_expires_at,
			scope = EXCLUDED.scope,
			encryption_context = EXCLUDED.encryption_context,
			updated_at = NOW()
	`

	_, err := r.adapter.db.ExecContext(ctx, query,
		session.ID, session.Principal, session.ServiceID,
		session.EncryptedAccessToken, session.EncryptedRefreshToken,
		session.TokenType, session.AccessTokenExpiresAt, session.RefreshTokenExpiresAt,
		session.Scope, session.EncryptionContext,
		session.InitiatedAt, session.CreatedAt, session.UpdatedAt,
	)

	if err != nil {
		return r.wrapError(err, "Create")
	}
	return nil
}

// Get retrieves a session by ID.
func (r *PostgresUserSessionRepository) Get(ctx context.Context, id string) (*storage.UserSession, error) {
	if id == "" {
		return nil, errors.New("session ID cannot be empty")
	}

	var session storage.UserSession
	query := `SELECT * FROM user_sessions WHERE id = $1`

	err := r.adapter.db.GetContext(ctx, &session, query, id)
	if err == sql.ErrNoRows {
		return nil, storage.NewStorageError("Get", storage.ErrorKindNotFound, err, "session not found")
	}
	if err != nil {
		return nil, r.wrapError(err, "Get")
	}
	return &session, nil
}

// FindByPrincipalAndService retrieves the session for a principal and service.
func (r *PostgresUserSessionRepository) FindByPrincipalAndService(ctx context.Context, principal, serviceID string) (*storage.UserSession, error) {
	if principal == "" || serviceID == "" {
		return nil, errors.New("principal and serviceID required")
	}

	var session storage.UserSession
	query := `SELECT * FROM user_sessions WHERE principal = $1 AND service_id = $2`

	err := r.adapter.db.GetContext(ctx, &session, query, principal, serviceID)
	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}
	if err != nil {
		return nil, r.wrapError(err, "FindByPrincipalAndService")
	}
	return &session, nil
}

// ListByPrincipal retrieves all sessions for a principal.
func (r *PostgresUserSessionRepository) ListByPrincipal(ctx context.Context, principal string) ([]*storage.UserSession, error) {
	if principal == "" {
		return nil, errors.New("principal required")
	}

	var sessions []*storage.UserSession
	query := `SELECT * FROM user_sessions WHERE principal = $1 ORDER BY created_at DESC`

	err := r.adapter.db.SelectContext(ctx, &sessions, query, principal)
	if err != nil && err != sql.ErrNoRows {
		return nil, r.wrapError(err, "ListByPrincipal")
	}
	return sessions, nil
}

// Delete deletes a session by ID.
func (r *PostgresUserSessionRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("session ID cannot be empty")
	}

	query := `DELETE FROM user_sessions WHERE id = $1`
	_, err := r.adapter.db.ExecContext(ctx, query, id)
	if err != nil {
		return r.wrapError(err, "Delete")
	}
	return nil
}

// DeleteByPrincipalAndService deletes the session for a principal and service.
func (r *PostgresUserSessionRepository) DeleteByPrincipalAndService(ctx context.Context, principal, serviceID string) error {
	if principal == "" || serviceID == "" {
		return errors.New("principal and serviceID required")
	}

	query := `DELETE FROM user_sessions WHERE principal = $1 AND service_id = $2`
	_, err := r.adapter.db.ExecContext(ctx, query, principal, serviceID)
	if err != nil {
		return r.wrapError(err, "DeleteByPrincipalAndService")
	}
	return nil
}

// CountByService counts sessions referencing a service.
func (r *PostgresUserSessionRepository) CountByService(ctx context.Context, serviceID string) (int, error) {
	if serviceID == "" {
		return 0, errors.New("serviceID required")
	}

	var count int
	query := `SELECT COUNT(*) FROM user_sessions WHERE service_id = $1`

	err := r.adapter.db.GetContext(ctx, &count, query, serviceID)
	if err != nil {
		return 0, r.wrapError(err, "CountByService")
	}
	return count, nil
}

// Helper function to wrap errors into StorageError
func (r *PostgresUserSessionRepository) wrapError(err error, operation string) error {
	if err == sql.ErrNoRows {
		return storage.NewStorageError(operation, storage.ErrorKindNotFound, err, "not found")
	}
	if err.Error() == "context deadline exceeded" {
		return storage.NewStorageError(operation, storage.ErrorKindTimeout, err, "timeout")
	}
	return storage.NewStorageError(operation, storage.ErrorKindUnknown, err, err.Error())
}
