// Package postgres implements PostgreSQL storage adapters.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/jackc/pgx/v5/pgconn"
)

// emptyIfNil converts a nil string slice to an empty slice.
// pq.Array returns SQL NULL for a nil slice, which violates NOT NULL constraints.
func emptyIfNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// AgentRepository implements ports.AgentRepository using PostgreSQL.
type AgentRepository struct {
	adapter *Adapter
}

// NewAgentRepository creates a new PostgreSQL agent repository.
// The adapter must be initialized before use.
func NewAgentRepository(adapter *Adapter) *AgentRepository {
	return &AgentRepository{
		adapter: adapter,
	}
}

// Create creates a new agent entity in PostgreSQL.
// Generates a UUID for the agent if ID is empty.
// Returns StorageError with Kind=Conflict if agent ID or client_id already exists.
func (r *AgentRepository) Create(ctx context.Context, agent *storage.Agent) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"CreateAgent",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if agent == nil {
		return storage.NewStorageError(
			"CreateAgent",
			storage.ErrorKindValidation,
			nil,
			"agent cannot be nil",
		)
	}

	// Validate before storing
	if err := agent.ValidateForCreate(); err != nil {
		return storage.NewStorageError(
			"CreateAgent",
			storage.ErrorKindValidation,
			err,
			"agent validation failed",
		)
	}

	// Generate ID if not provided
	if agent.ID.IsZero() {
		agent.ID = id.NewAgentID()
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	// Marshal service_requirements to JSON (NULL if empty/nil)
	var serviceReqsJSON []byte
	var err error
	if len(agent.ServiceRequirements) > 0 {
		serviceReqsJSON, err = json.Marshal(agent.ServiceRequirements)
		if err != nil {
			return storage.NewStorageError(
				"CreateAgent",
				storage.ErrorKindValidation,
				err,
				"failed to marshal service_requirements to JSON",
			)
		}
	}

	query := `
		INSERT INTO agents (
			id, client_id, external_id, display_name, description,
			governance_url, user_documentation_url, agent_interface_url,
			service_requirements, redirect_uris, allowed_scopes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = r.adapter.db.ExecContext(
		execCtx,
		query,
		agent.ID,
		agent.ClientID,
		agent.ExternalID,
		agent.DisplayName,
		agent.Description,
		agent.GovernanceURL,
		agent.UserDocumentationURL,
		agent.AgentInterfaceURL,
		serviceReqsJSON, // NULL if empty
		pq.Array(emptyIfNil(agent.RedirectURIs)),
		pq.Array(emptyIfNil(agent.AllowedScopes)),
		agent.CreatedAt,
		agent.UpdatedAt,
	)

	if err != nil {
		// Check for timeout
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return storage.NewStorageError(
				"CreateAgent",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}

		// Check for duplicate key violation (PostgreSQL error code 23505)
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "pkey") {
				return storage.NewStorageError(
					"CreateAgent",
					storage.ErrorKindConflict,
					err,
					"agent with this ID already exists",
				)
			}
			if strings.Contains(pgErr.ConstraintName, "client_id") {
				return storage.NewStorageError(
					"CreateAgent",
					storage.ErrorKindConflict,
					err,
					"agent with this client_id already exists",
				)
			}
		}

		return storage.NewStorageError(
			"CreateAgent",
			storage.ErrorKindConnection,
			err,
			"failed to create agent",
		)
	}

	return nil
}

// Get retrieves an agent entity by ID from PostgreSQL.
// Returns StorageError with Kind=NotFound if agent not found.
func (r *AgentRepository) Get(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
	ctx, span := otel.Tracer("storage").Start(ctx, "storage.get.agent")
	defer span.End()
	span.SetAttributes(
		semconv.DBSystemKey.String("postgresql"),
		attribute.String("db.operation", "GetAgent"),
	)

	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"GetAgent",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if agentID.IsZero() {
		return nil, storage.NewStorageError(
			"GetAgent",
			storage.ErrorKindValidation,
			nil,
			"agent ID cannot be empty",
		)
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	var (
		serviceReqsJSON []byte
	)

	query := `
		SELECT id, client_id, external_id, display_name, description,
		       governance_url, user_documentation_url, agent_interface_url,
		       service_requirements, redirect_uris, allowed_scopes, created_at, updated_at
		FROM agents
		WHERE id = $1
	`

	agent := &storage.Agent{}
	err := r.adapter.db.QueryRowContext(queryCtx, query, agentID).Scan(
		&agent.ID,
		&agent.ClientID,
		&agent.ExternalID,
		&agent.DisplayName,
		&agent.Description,
		&agent.GovernanceURL,
		&agent.UserDocumentationURL,
		&agent.AgentInterfaceURL,
		&serviceReqsJSON,
		pq.Array(&agent.RedirectURIs),
		pq.Array(&agent.AllowedScopes),
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.NewStorageError(
				"GetAgent",
				storage.ErrorKindNotFound,
				ports.ErrNotFound,
				"agent not found",
			)
		}
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, storage.NewStorageError(
				"GetAgent",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}
		return nil, storage.NewStorageError(
			"GetAgent",
			storage.ErrorKindConnection,
			err,
			"failed to get agent",
		)
	}

	// Unmarshal service_requirements from JSON (if not NULL)
	if len(serviceReqsJSON) > 0 {
		if err := json.Unmarshal(serviceReqsJSON, &agent.ServiceRequirements); err != nil {
			return nil, storage.NewStorageError(
				"GetAgent",
				storage.ErrorKindValidation,
				err,
				"failed to unmarshal service_requirements from JSON",
			)
		}
	}

	// Return deep copy to prevent external mutation
	return agent.Copy(), nil
}

// Update updates an existing agent entity in PostgreSQL.
// Returns StorageError with Kind=NotFound if agent ID not found.
func (r *AgentRepository) Update(ctx context.Context, agent *storage.Agent) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"UpdateAgent",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if agent == nil {
		return storage.NewStorageError(
			"UpdateAgent",
			storage.ErrorKindValidation,
			nil,
			"agent cannot be nil",
		)
	}

	// Validate before updating
	if err := agent.Validate(); err != nil {
		return storage.NewStorageError(
			"UpdateAgent",
			storage.ErrorKindValidation,
			err,
			"agent validation failed",
		)
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	// Marshal service_requirements to JSON (NULL if empty/nil)
	var serviceReqsJSON []byte
	var err error
	if len(agent.ServiceRequirements) > 0 {
		serviceReqsJSON, err = json.Marshal(agent.ServiceRequirements)
		if err != nil {
			return storage.NewStorageError(
				"UpdateAgent",
				storage.ErrorKindValidation,
				err,
				"failed to marshal service_requirements to JSON",
			)
		}
	}

	query := `
		UPDATE agents
		SET client_id = $2,
		    external_id = $3,
		    display_name = $4,
		    description = $5,
		    governance_url = $6,
		    user_documentation_url = $7,
		    agent_interface_url = $8,
		    service_requirements = $9,
		    redirect_uris = $10,
		    allowed_scopes = $11,
		    updated_at = $12
		WHERE id = $1
	`

	result, err := r.adapter.db.ExecContext(
		execCtx,
		query,
		agent.ID,
		agent.ClientID,
		agent.ExternalID,
		agent.DisplayName,
		agent.Description,
		agent.GovernanceURL,
		agent.UserDocumentationURL,
		agent.AgentInterfaceURL,
		serviceReqsJSON, // NULL if empty
		pq.Array(emptyIfNil(agent.RedirectURIs)),
		pq.Array(emptyIfNil(agent.AllowedScopes)),
		agent.UpdatedAt,
	)

	if err != nil {
		// Check for timeout
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return storage.NewStorageError(
				"UpdateAgent",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}

		// Check for duplicate client_id violation
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "client_id") {
				return storage.NewStorageError(
					"UpdateAgent",
					storage.ErrorKindConflict,
					err,
					"agent with this client_id already exists",
				)
			}
		}

		return storage.NewStorageError(
			"UpdateAgent",
			storage.ErrorKindConnection,
			err,
			"failed to update agent",
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return storage.NewStorageError(
			"UpdateAgent",
			storage.ErrorKindConnection,
			err,
			"failed to get affected rows",
		)
	}

	if rows == 0 {
		return storage.NewStorageError(
			"UpdateAgent",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"agent not found",
		)
	}

	return nil
}

// Delete deletes an agent entity by ID from PostgreSQL.
// Idempotent: returns nil if agent doesn't exist.
// Associated grants are CASCADE deleted per FR-021.
func (r *AgentRepository) Delete(ctx context.Context, agentID id.AgentID) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"DeleteAgent",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if agentID.IsZero() {
		return storage.NewStorageError(
			"DeleteAgent",
			storage.ErrorKindValidation,
			nil,
			"agent ID cannot be empty",
		)
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `DELETE FROM agents WHERE id = $1`

	_, err := r.adapter.db.ExecContext(execCtx, query, agentID)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return storage.NewStorageError(
				"DeleteAgent",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}
		return storage.NewStorageError(
			"DeleteAgent",
			storage.ErrorKindConnection,
			err,
			"failed to delete agent",
		)
	}

	return nil
}

// List retrieves all agent entities from PostgreSQL.
// Returns empty slice if no agents exist (not an error).
func (r *AgentRepository) List(ctx context.Context) ([]*storage.Agent, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"ListAgents",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `
		SELECT id, client_id, external_id, display_name, description,
		       governance_url, user_documentation_url, agent_interface_url,
		       service_requirements, redirect_uris, allowed_scopes, created_at, updated_at
		FROM agents
		ORDER BY created_at DESC
	`

	rows, err := r.adapter.db.QueryContext(queryCtx, query)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, storage.NewStorageError("ListAgents", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return nil, storage.NewStorageError("ListAgents", storage.ErrorKindConnection, err, "failed to list agents")
	}
	defer func() { _ = rows.Close() }()

	var result []*storage.Agent
	for rows.Next() {
		agent := &storage.Agent{}
		var serviceReqsJSON []byte
		if err := rows.Scan(
			&agent.ID, &agent.ClientID, &agent.ExternalID,
			&agent.DisplayName, &agent.Description,
			&agent.GovernanceURL, &agent.UserDocumentationURL, &agent.AgentInterfaceURL,
			&serviceReqsJSON,
			pq.Array(&agent.RedirectURIs), pq.Array(&agent.AllowedScopes),
			&agent.CreatedAt, &agent.UpdatedAt,
		); err != nil {
			return nil, storage.NewStorageError("ListAgents", storage.ErrorKindConnection, err, "failed to scan agent row")
		}
		if len(serviceReqsJSON) > 0 {
			if err := json.Unmarshal(serviceReqsJSON, &agent.ServiceRequirements); err != nil {
				return nil, storage.NewStorageError("ListAgents", storage.ErrorKindValidation, err, "failed to unmarshal service_requirements from JSON")
			}
		}
		result = append(result, agent.Copy())
	}
	if err := rows.Err(); err != nil {
		return nil, storage.NewStorageError("ListAgents", storage.ErrorKindConnection, err, "error iterating agent rows")
	}

	if result == nil {
		result = []*storage.Agent{}
	}
	return result, nil
}

// GetByClientID retrieves an agent entity by client_id from PostgreSQL.
// Returns StorageError with Kind=NotFound if agent not found.
func (r *AgentRepository) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.Agent, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"GetAgentByClientID",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if clientID.IsZero() {
		return nil, storage.NewStorageError(
			"GetAgentByClientID",
			storage.ErrorKindValidation,
			nil,
			"client_id cannot be empty",
		)
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	var serviceReqsJSON []byte
	agent := &storage.Agent{}
	query := `
		SELECT id, client_id, external_id, display_name, description,
		       governance_url, user_documentation_url, agent_interface_url,
		       service_requirements, redirect_uris, allowed_scopes, created_at, updated_at
		FROM agents
		WHERE client_id = $1
	`

	err := r.adapter.db.QueryRowContext(queryCtx, query, clientID).Scan(
		&agent.ID, &agent.ClientID, &agent.ExternalID,
		&agent.DisplayName, &agent.Description,
		&agent.GovernanceURL, &agent.UserDocumentationURL, &agent.AgentInterfaceURL,
		&serviceReqsJSON,
		pq.Array(&agent.RedirectURIs), pq.Array(&agent.AllowedScopes),
		&agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.NewStorageError(
				"GetAgentByClientID",
				storage.ErrorKindNotFound,
				ports.ErrNotFound,
				"agent not found",
			)
		}
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, storage.NewStorageError(
				"GetAgentByClientID",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}
		return nil, storage.NewStorageError(
			"GetAgentByClientID",
			storage.ErrorKindConnection,
			err,
			"failed to get agent by client_id",
		)
	}

	if len(serviceReqsJSON) > 0 {
		if err := json.Unmarshal(serviceReqsJSON, &agent.ServiceRequirements); err != nil {
			return nil, storage.NewStorageError(
				"GetAgentByClientID",
				storage.ErrorKindValidation,
				err,
				"failed to unmarshal service_requirements from JSON",
			)
		}
	}

	// Return deep copy to prevent external mutation
	return agent.Copy(), nil
}
