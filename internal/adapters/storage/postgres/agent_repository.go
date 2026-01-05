// Package postgres implements PostgreSQL storage adapters.
package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

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
	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `
		INSERT INTO agents (
			id, client_id, external_id, display_name, description,
			governance_url, user_documentation_url, agent_interface_url,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.adapter.db.ExecContext(
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
func (r *AgentRepository) Get(ctx context.Context, id string) (*storage.Agent, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"GetAgent",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if id == "" {
		return nil, storage.NewStorageError(
			"GetAgent",
			storage.ErrorKindValidation,
			nil,
			"agent ID cannot be empty",
		)
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	agent := &storage.Agent{}
	query := `
		SELECT id, client_id, external_id, display_name, description,
		       governance_url, user_documentation_url, agent_interface_url,
		       created_at, updated_at
		FROM agents
		WHERE id = $1
	`

	err := r.adapter.db.GetContext(queryCtx, agent, query, id)
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

	query := `
		UPDATE agents
		SET client_id = $2,
		    external_id = $3,
		    display_name = $4,
		    description = $5,
		    governance_url = $6,
		    user_documentation_url = $7,
		    agent_interface_url = $8,
		    updated_at = $9
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
func (r *AgentRepository) Delete(ctx context.Context, id string) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"DeleteAgent",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if id == "" {
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

	_, err := r.adapter.db.ExecContext(execCtx, query, id)
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
		       created_at, updated_at
		FROM agents
		ORDER BY created_at DESC
	`

	var agents []*storage.Agent
	err := r.adapter.db.SelectContext(queryCtx, &agents, query)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, storage.NewStorageError(
				"ListAgents",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}
		return nil, storage.NewStorageError(
			"ListAgents",
			storage.ErrorKindConnection,
			err,
			"failed to list agents",
		)
	}

	// Return deep copies to prevent external mutation
	result := make([]*storage.Agent, len(agents))
	for i, agent := range agents {
		result[i] = agent.Copy()
	}

	return result, nil
}

// GetByClientID retrieves an agent entity by client_id from PostgreSQL.
// Returns StorageError with Kind=NotFound if agent not found.
func (r *AgentRepository) GetByClientID(ctx context.Context, clientID string) (*storage.Agent, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"GetAgentByClientID",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if clientID == "" {
		return nil, storage.NewStorageError(
			"GetAgentByClientID",
			storage.ErrorKindValidation,
			nil,
			"client_id cannot be empty",
		)
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	agent := &storage.Agent{}
	query := `
		SELECT id, client_id, external_id, display_name, description,
		       governance_url, user_documentation_url, agent_interface_url,
		       created_at, updated_at
		FROM agents
		WHERE client_id = $1
	`

	err := r.adapter.db.GetContext(queryCtx, agent, query, clientID)
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

	// Return deep copy to prevent external mutation
	return agent.Copy(), nil
}
