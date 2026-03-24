// Package postgres implements PostgreSQL storage adapters.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/jackc/pgx/v5/pgconn"
)

// UserGrantRepository implements ports.UserGrantRepository using PostgreSQL.
type UserGrantRepository struct {
	adapter *Adapter
}

// NewUserGrantRepository creates a new PostgreSQL user grant repository.
// The adapter must be initialized before use.
func NewUserGrantRepository(adapter *Adapter) *UserGrantRepository {
	return &UserGrantRepository{
		adapter: adapter,
	}
}

// Create creates a new user grant or updates existing grant for same principal+agent (upsert).
// The grant ID should be generated before calling this method.
// Uses ON CONFLICT to implement upsert semantics (one grant per principal-agent pair).
func (r *UserGrantRepository) Create(ctx context.Context, grant *storage.UserGrant) error {
	ctx, span := otel.Tracer("storage").Start(ctx, "storage.upsert.userGrant")
	defer span.End()
	span.SetAttributes(
		semconv.DBSystemKey.String("postgresql"),
		attribute.String("db.operation", "UpsertUserGrant"),
	)

	if r.adapter.db == nil {
		return storage.NewStorageError(
			"CreateUserGrant",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if grant == nil {
		return storage.NewStorageError(
			"CreateUserGrant",
			storage.ErrorKindValidation,
			nil,
			"grant cannot be nil",
		)
	}

	// Generate ID if not provided
	if grant.ID.IsZero() {
		grant.ID = id.NewGrantID()
	}

	// Validate before storing
	if err := grant.ValidateForCreate(); err != nil {
		return storage.NewStorageError(
			"CreateUserGrant",
			storage.ErrorKindValidation,
			err,
			"grant validation failed",
		)
	}

	// Marshal delegated tokens to JSONB
	tokensJSON, err := json.Marshal(grant.DelegatedOAuth2Tokens)
	if err != nil {
		return storage.NewStorageError(
			"CreateUserGrant",
			storage.ErrorKindUnknown,
			err,
			"failed to marshal delegated tokens",
		)
	}

	// Create context with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	// Use ON CONFLICT to implement upsert semantics
	query := `
		INSERT INTO user_grants (
			id, principal, agent_id, valid_until, delegated_oauth2_tokens, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (principal, agent_id)
		DO UPDATE SET
			valid_until = EXCLUDED.valid_until,
			delegated_oauth2_tokens = EXCLUDED.delegated_oauth2_tokens,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	var returnedID id.GrantID
	err = r.adapter.db.QueryRowContext(
		ctxTimeout,
		query,
		grant.ID,
		grant.Principal,
		grant.AgentID,
		grant.ValidUntil,
		tokensJSON,
		grant.CreatedAt,
		grant.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		return r.handlePostgresError("CreateUserGrant", err)
	}

	// Update grant ID if it was changed by upsert
	grant.ID = returnedID

	return nil
}

// Get retrieves a user grant by ID.
// Returns StorageError with Kind=NotFound if grant not found.
func (r *UserGrantRepository) Get(ctx context.Context, grantID id.GrantID) (*storage.UserGrant, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"GetUserGrant",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	// Create context with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `
		SELECT id, principal, agent_id, valid_until, delegated_oauth2_tokens, created_at, updated_at
		FROM user_grants
		WHERE id = $1
	`

	var grant storage.UserGrant
	var tokensJSON []byte

	err := r.adapter.db.QueryRowContext(ctxTimeout, query, grantID).Scan(
		&grant.ID,
		&grant.Principal,
		&grant.AgentID,
		&grant.ValidUntil,
		&tokensJSON,
		&grant.CreatedAt,
		&grant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.NewStorageError(
				"GetUserGrant",
				storage.ErrorKindNotFound,
				ports.ErrNotFound,
				"user grant not found",
			)
		}
		return nil, r.handlePostgresError("GetUserGrant", err)
	}

	// Unmarshal JSONB tokens
	if err := json.Unmarshal(tokensJSON, &grant.DelegatedOAuth2Tokens); err != nil {
		return nil, storage.NewStorageError(
			"GetUserGrant",
			storage.ErrorKindUnknown,
			err,
			"failed to unmarshal delegated tokens",
		)
	}

	return grant.Copy(), nil
}

// Update updates an existing user grant.
// Returns StorageError with Kind=NotFound if grant not found.
func (r *UserGrantRepository) Update(ctx context.Context, grant *storage.UserGrant) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"UpdateUserGrant",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if grant == nil {
		return storage.NewStorageError(
			"UpdateUserGrant",
			storage.ErrorKindValidation,
			nil,
			"grant cannot be nil",
		)
	}

	// Validate before updating
	if err := grant.Validate(); err != nil {
		return storage.NewStorageError(
			"UpdateUserGrant",
			storage.ErrorKindValidation,
			err,
			"grant validation failed",
		)
	}

	// Marshal delegated tokens to JSONB
	tokensJSON, err := json.Marshal(grant.DelegatedOAuth2Tokens)
	if err != nil {
		return storage.NewStorageError(
			"UpdateUserGrant",
			storage.ErrorKindUnknown,
			err,
			"failed to marshal delegated tokens",
		)
	}

	// Create context with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `
		UPDATE user_grants
		SET valid_until = $2,
		    delegated_oauth2_tokens = $3,
		    updated_at = $4
		WHERE id = $1
	`

	result, err := r.adapter.db.ExecContext(
		ctxTimeout,
		query,
		grant.ID,
		grant.ValidUntil,
		tokensJSON,
		grant.UpdatedAt,
	)

	if err != nil {
		return r.handlePostgresError("UpdateUserGrant", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return storage.NewStorageError(
			"UpdateUserGrant",
			storage.ErrorKindUnknown,
			err,
			"failed to get rows affected",
		)
	}

	if rowsAffected == 0 {
		return storage.NewStorageError(
			"UpdateUserGrant",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"user grant not found",
		)
	}

	return nil
}

// Delete deletes a user grant by ID.
// Idempotent: returns nil if grant doesn't exist.
func (r *UserGrantRepository) Delete(ctx context.Context, grantID id.GrantID) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"DeleteUserGrant",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	// Create context with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `DELETE FROM user_grants WHERE id = $1`

	_, err := r.adapter.db.ExecContext(ctxTimeout, query, grantID)
	if err != nil {
		return r.handlePostgresError("DeleteUserGrant", err)
	}

	// Idempotent: success even if no rows deleted
	return nil
}

// ListByPrincipalAndAgent retrieves all grants for a principal and specific agent.
// Includes expired grants (filtering happens in service layer).
// Returns empty slice if no grants exist (not an error).
func (r *UserGrantRepository) ListByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"ListUserGrants",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	// Create context with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `
		SELECT id, principal, agent_id, valid_until, delegated_oauth2_tokens, created_at, updated_at
		FROM user_grants
		WHERE principal = $1 AND agent_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.adapter.db.QueryContext(ctxTimeout, query, principal, agentID)
	if err != nil {
		return nil, r.handlePostgresError("ListUserGrants", err)
	}
	defer func() { _ = rows.Close() }()

	var grants []*storage.UserGrant

	for rows.Next() {
		var grant storage.UserGrant
		var tokensJSON []byte

		err := rows.Scan(
			&grant.ID,
			&grant.Principal,
			&grant.AgentID,
			&grant.ValidUntil,
			&tokensJSON,
			&grant.CreatedAt,
			&grant.UpdatedAt,
		)
		if err != nil {
			return nil, storage.NewStorageError(
				"ListUserGrants",
				storage.ErrorKindUnknown,
				err,
				"failed to scan grant row",
			)
		}

		// Unmarshal JSONB tokens
		if err := json.Unmarshal(tokensJSON, &grant.DelegatedOAuth2Tokens); err != nil {
			return nil, storage.NewStorageError(
				"ListUserGrants",
				storage.ErrorKindUnknown,
				err,
				"failed to unmarshal delegated tokens",
			)
		}

		grants = append(grants, grant.Copy())
	}

	if err := rows.Err(); err != nil {
		return nil, storage.NewStorageError(
			"ListUserGrants",
			storage.ErrorKindUnknown,
			err,
			"error iterating grant rows",
		)
	}

	return grants, nil
}

// FindByPrincipalAndAgent retrieves the grant for a principal and agent pair.
// Returns StorageError with Kind=NotFound if grant not found.
func (r *UserGrantRepository) FindByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) (*storage.UserGrant, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"FindUserGrant",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	// Create context with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `
		SELECT id, principal, agent_id, valid_until, delegated_oauth2_tokens, created_at, updated_at
		FROM user_grants
		WHERE principal = $1 AND agent_id = $2
		LIMIT 1
	`

	var grant storage.UserGrant
	var tokensJSON []byte

	err := r.adapter.db.QueryRowContext(ctxTimeout, query, principal, agentID).Scan(
		&grant.ID,
		&grant.Principal,
		&grant.AgentID,
		&grant.ValidUntil,
		&tokensJSON,
		&grant.CreatedAt,
		&grant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.NewStorageError(
				"FindUserGrant",
				storage.ErrorKindNotFound,
				ports.ErrNotFound,
				"user grant not found",
			)
		}
		return nil, r.handlePostgresError("FindUserGrant", err)
	}

	// Unmarshal JSONB tokens
	if err := json.Unmarshal(tokensJSON, &grant.DelegatedOAuth2Tokens); err != nil {
		return nil, storage.NewStorageError(
			"FindUserGrant",
			storage.ErrorKindUnknown,
			err,
			"failed to unmarshal delegated tokens",
		)
	}

	return grant.Copy(), nil
}

// DeleteByAgent deletes all grants associated with an agent.
// Used during cascade deletion when agent is deleted (FR-021).
// Idempotent: returns nil if agent has no grants.
func (r *UserGrantRepository) DeleteByAgent(ctx context.Context, agentID id.AgentID) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"DeleteGrantsByAgent",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	// Create context with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `DELETE FROM user_grants WHERE agent_id = $1`

	_, err := r.adapter.db.ExecContext(ctxTimeout, query, agentID)
	if err != nil {
		return r.handlePostgresError("DeleteGrantsByAgent", err)
	}

	// Idempotent: success even if no rows deleted
	return nil
}

// ListByPrincipal retrieves all active grants for a principal across all agents.
// Filters expired grants (valid_until < NOW()).
// Returns empty slice if no active grants exist (not an error).
func (r *UserGrantRepository) ListByPrincipal(ctx context.Context, principal id.Principal) ([]storage.UserGrant, error) {
	ctx, span := otel.Tracer("storage").Start(ctx, "storage.list.userGrants")
	defer span.End()
	span.SetAttributes(
		semconv.DBSystemKey.String("postgresql"),
		attribute.String("db.operation", "ListUserGrantsByPrincipal"),
	)

	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"ListByPrincipal",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	// Create context with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `
		SELECT id, principal, agent_id, valid_until, delegated_oauth2_tokens, created_at, updated_at
		FROM user_grants
		WHERE principal = $1
		  AND (valid_until IS NULL OR valid_until > NOW())
		ORDER BY updated_at DESC
	`

	rows, err := r.adapter.db.QueryContext(ctxTimeout, query, principal)
	if err != nil {
		return nil, r.handlePostgresError("ListByPrincipal", err)
	}
	defer func() { _ = rows.Close() }()

	var grants []storage.UserGrant

	for rows.Next() {
		var grant storage.UserGrant
		var tokensJSON []byte

		err := rows.Scan(
			&grant.ID,
			&grant.Principal,
			&grant.AgentID,
			&grant.ValidUntil,
			&tokensJSON,
			&grant.CreatedAt,
			&grant.UpdatedAt,
		)
		if err != nil {
			return nil, storage.NewStorageError(
				"ListByPrincipal",
				storage.ErrorKindUnknown,
				err,
				"failed to scan grant row",
			)
		}

		// Unmarshal JSONB tokens
		if err := json.Unmarshal(tokensJSON, &grant.DelegatedOAuth2Tokens); err != nil {
			return nil, storage.NewStorageError(
				"ListByPrincipal",
				storage.ErrorKindUnknown,
				err,
				"failed to unmarshal delegated tokens",
			)
		}

		grants = append(grants, *grant.Copy())
	}

	if err := rows.Err(); err != nil {
		return nil, storage.NewStorageError(
			"ListByPrincipal",
			storage.ErrorKindUnknown,
			err,
			"error iterating grant rows",
		)
	}

	return grants, nil
}

// CountAgentsByServiceID counts how many agents have delegated OAuth2 tokens for a given service.
// This is used to show dependent agent count when terminating a session.
// Returns the count of distinct agents with delegated_oauth2_tokens JSONB entries for the service.
func (r *UserGrantRepository) CountAgentsByServiceID(ctx context.Context, serviceID id.ServiceID) (int, error) {
	if r.adapter.db == nil {
		return 0, storage.NewStorageError(
			"CountAgentsByServiceID",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	// Query to count distinct agents that have delegated tokens for this service
	// Uses jsonb_array_elements to unnest the delegated_oauth2_tokens array
	// and filters by thirdparty_oauth2_service_id
	query := `
		SELECT COUNT(DISTINCT agent_id)
		FROM user_grants,
		     jsonb_array_elements(delegated_oauth2_tokens) AS token
		WHERE token->>'thirdparty_oauth2_service_id' = $1
	`

	var count int
	err := r.adapter.db.GetContext(ctx, &count, query, serviceID)
	if err != nil {
		return 0, r.handlePostgresError("CountAgentsByServiceID", err)
	}

	return count, nil
}

// ListByServiceID retrieves all agent IDs that have delegated OAuth2 tokens for a given service.
// This is used to show the actual dependent agents when terminating a session.
// Returns the list of distinct agent IDs with delegated_oauth2_tokens entries for the service.
// Returns empty slice if no agents have delegated tokens for the service.
func (r *UserGrantRepository) ListByServiceID(ctx context.Context, serviceID id.ServiceID) ([]id.AgentID, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"ListByServiceID",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	// Query to get distinct agent IDs that have delegated tokens for this service
	// Uses jsonb_array_elements to unnest the delegated_oauth2_tokens array
	// and filters by thirdparty_oauth2_service_id
	query := `
		SELECT DISTINCT agent_id
		FROM user_grants,
		     jsonb_array_elements(delegated_oauth2_tokens) AS token
		WHERE token->>'thirdparty_oauth2_service_id' = $1
		ORDER BY agent_id
	`

	// Create context with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	rows, err := r.adapter.db.QueryContext(ctxTimeout, query, serviceID)
	if err != nil {
		return nil, r.handlePostgresError("ListByServiceID", err)
	}
	defer func() { _ = rows.Close() }()

	var agentIDs []id.AgentID

	for rows.Next() {
		var agentID id.AgentID
		err := rows.Scan(&agentID)
		if err != nil {
			return nil, storage.NewStorageError(
				"ListByServiceID",
				storage.ErrorKindUnknown,
				err,
				"failed to scan agent ID row",
			)
		}
		agentIDs = append(agentIDs, agentID)
	}

	if err := rows.Err(); err != nil {
		return nil, storage.NewStorageError(
			"ListByServiceID",
			storage.ErrorKindUnknown,
			err,
			"error iterating agent ID rows",
		)
	}

	return agentIDs, nil
}

// DeleteByPrincipalAndAgentID deletes the grant owned by principal for the given agent.
// Returns StorageError wrapping ports.ErrNotFound when no grant exists for the pair.
// This is NOT idempotent: absence of a grant is an error (revocation semantics FR-014).
func (r *UserGrantRepository) DeleteByPrincipalAndAgentID(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"DeleteByPrincipalAndAgentID",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `DELETE FROM user_grants WHERE principal = $1 AND agent_id = $2`

	result, err := r.adapter.db.ExecContext(ctxTimeout, query, principal, agentID)
	if err != nil {
		return r.handlePostgresError("DeleteByPrincipalAndAgentID", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return storage.NewStorageError(
			"DeleteByPrincipalAndAgentID",
			storage.ErrorKindUnknown,
			err,
			"failed to get rows affected",
		)
	}

	if rowsAffected == 0 {
		return storage.NewStorageError(
			"DeleteByPrincipalAndAgentID",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"no active grant exists for this principal and agent",
		)
	}

	return nil
}

// handlePostgresError converts PostgreSQL errors to StorageError.
func (r *UserGrantRepository) handlePostgresError(operation string, err error) error {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23505": // unique_violation
			return storage.NewStorageError(
				operation,
				storage.ErrorKindConflict,
				err,
				fmt.Sprintf("conflict: %s", pgErr.Detail),
			)
		case "23503": // foreign_key_violation
			return storage.NewStorageError(
				operation,
				storage.ErrorKindNotFound,
				err,
				fmt.Sprintf("foreign key violation: %s", pgErr.Detail),
			)
		}
	}

	// Check for context errors in the wrapped error
	if err == context.DeadlineExceeded {
		return storage.NewStorageError(
			operation,
			storage.ErrorKindTimeout,
			err,
			"operation timed out",
		)
	}

	// Default to connection error
	return storage.NewStorageError(
		operation,
		storage.ErrorKindConnection,
		err,
		"PostgreSQL operation failed",
	)
}
