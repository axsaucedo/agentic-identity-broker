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
//
// The byClientID index is a 1:many map to correctly support multi-agent sharing
// (multiple agents sharing the same upstream OAuth2 client_id).
type AgentRepository struct {
	mu         sync.RWMutex
	agents     map[id.AgentID]*storage.Agent // ID -> Agent
	byClientID map[id.ClientID][]id.AgentID  // ClientID -> []ID (1:many for multi-agent support)
}

// NewAgentRepository creates a new in-memory agent repository.
func NewAgentRepository() *AgentRepository {
	return &AgentRepository{
		agents:     make(map[id.AgentID]*storage.Agent),
		byClientID: make(map[id.ClientID][]id.AgentID),
	}
}

// Create creates a new agent entity in storage.
// Generates a UUID for the agent if ID is empty.
// Returns StorageError with Kind=Conflict if agent ID already exists.
// Multiple agents may share the same client_id (multi-agent mode support).
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
	r.byClientID[agent.ClientID] = append(r.byClientID[agent.ClientID], agent.ID)

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

	// If client_id changed, update the secondary index.
	// No conflict check here: uniqueness enforcement is the app handler's responsibility.
	// Multiple agents may share a client_id in multi-agent mode.
	if existing.ClientID != agent.ClientID {
		r.removeFromClientIDIndex(existing.ClientID, agent.ID)
		r.byClientID[agent.ClientID] = append(r.byClientID[agent.ClientID], agent.ID)
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
		r.removeFromClientIDIndex(agent.ClientID, agentID)
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
// When multiple agents share the same client_id (multi-agent mode), returns the first registered one.
// Returns StorageError with Kind=NotFound if no agent with that client_id exists.
func (r *AgentRepository) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, exists := r.byClientID[clientID]
	if !exists || len(ids) == 0 {
		return nil, storage.NewStorageError(
			"GetAgentByClientID",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"agent not found",
		)
	}

	agent := r.agents[ids[0]]
	// Return deep copy to prevent external mutation
	return agent.Copy(), nil
}

// removeFromClientIDIndex removes a specific agentID from the byClientID slice for clientID.
// Deletes the map entry if the slice becomes empty. Must be called with r.mu held (write lock).
func (r *AgentRepository) removeFromClientIDIndex(clientID id.ClientID, agentID id.AgentID) {
	ids := r.byClientID[clientID]
	for i, v := range ids {
		if v == agentID {
			ids = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	if len(ids) == 0 {
		delete(r.byClientID, clientID)
	} else {
		r.byClientID[clientID] = ids
	}
}
