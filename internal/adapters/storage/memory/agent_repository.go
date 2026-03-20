package memory

import (
	"context"
	"sync"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// AgentRepository provides in-memory storage for Agent entities.
// Thread-safe implementation using sync.RWMutex.
type AgentRepository struct {
	mu         sync.RWMutex
	agents     map[id.AgentID]*storage.Agent // ID -> Agent
	byClientID map[id.ClientID]id.AgentID    // ClientID -> ID
}

// NewAgentRepository creates a new in-memory agent repository.
func NewAgentRepository() *AgentRepository {
	return &AgentRepository{
		agents:     make(map[id.AgentID]*storage.Agent),
		byClientID: make(map[id.ClientID]id.AgentID),
	}
}

// Create creates a new agent entity in storage.
// Generates a UUID for the agent if ID is empty.
// Returns StorageError with Kind=Conflict if agent ID or client_id already exists.
func (r *AgentRepository) Create(ctx context.Context, agent *storage.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate ID if not provided
	if agent.ID.IsZero() {
		agent.ID = id.NewAgentID()
	}

	// Check for duplicate ID
	if _, exists := r.agents[agent.ID]; exists {
		return storage.NewStorageError(
			"CreateAgent",
			storage.ErrorKindConflict,
			nil,
			"agent with this ID already exists",
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

	// Store deep copy to prevent external mutation
	r.agents[agent.ID] = agent.Copy()
	r.byClientID[agent.ClientID] = agent.ID

	return nil
}

// Get retrieves an agent entity by ID.
// Returns StorageError with Kind=NotFound if agent not found.
func (r *AgentRepository) Get(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, storage.NewStorageError(
			"GetAgent",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"agent not found",
		)
	}

	// Return deep copy to prevent external mutation
	return agent.Copy(), nil
}

// Update updates an existing agent entity.
// Returns StorageError with Kind=NotFound if agent ID not found.
func (r *AgentRepository) Update(ctx context.Context, agent *storage.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if agent exists
	existing, exists := r.agents[agent.ID]
	if !exists {
		return storage.NewStorageError(
			"UpdateAgent",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"agent not found",
		)
	}

	// If client_id changed, check for conflicts
	if existing.ClientID != agent.ClientID {
		if otherID, exists := r.byClientID[agent.ClientID]; exists && otherID != agent.ID {
			return storage.NewStorageError(
				"UpdateAgent",
				storage.ErrorKindConflict,
				nil,
				"agent with this client_id already exists",
			)
		}
		// Update client_id index
		delete(r.byClientID, existing.ClientID)
		r.byClientID[agent.ClientID] = agent.ID
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

	// Store deep copy
	r.agents[agent.ID] = agent.Copy()

	return nil
}

// Delete deletes an agent entity by ID.
// Idempotent: returns nil if agent doesn't exist.
func (r *AgentRepository) Delete(ctx context.Context, agentID id.AgentID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get agent to clean up indexes
	if agent, exists := r.agents[agentID]; exists {
		delete(r.byClientID, agent.ClientID)
		delete(r.agents, agentID)
	}

	return nil
}

// List retrieves all agent entities.
// Returns empty slice if no agents exist (not an error).
func (r *AgentRepository) List(ctx context.Context) ([]*storage.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*storage.Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		result = append(result, agent.Copy())
	}

	return result, nil
}

// GetByClientID retrieves an agent entity by client_id.
// Returns StorageError with Kind=NotFound if agent not found.
func (r *AgentRepository) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Use existing byClientID index
	agentID, exists := r.byClientID[clientID]
	if !exists {
		return nil, storage.NewStorageError(
			"GetAgentByClientID",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"agent not found",
		)
	}

	agent := r.agents[agentID]
	// Return deep copy to prevent external mutation
	return agent.Copy(), nil
}
