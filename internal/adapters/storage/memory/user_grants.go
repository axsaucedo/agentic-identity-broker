package memory

import (
	"context"
	"sync"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/google/uuid"
)

// UserGrantRepository provides in-memory storage for UserGrant entities.
// Thread-safe implementation using sync.RWMutex.
// Implements upsert semantics: one grant per (principal, agent_id) pair.
type UserGrantRepository struct {
	mu                  sync.RWMutex
	grants              map[string]*storage.UserGrant // ID -> Grant
	byPrincipalAndAgent map[string]string             // "principal:agent_id" -> ID
	grantIDsByAgent     map[string][]string           // agent_id -> []grant_id (for cascade delete)
}

// NewUserGrantRepository creates a new in-memory user grant repository.
func NewUserGrantRepository() *UserGrantRepository {
	return &UserGrantRepository{
		grants:              make(map[string]*storage.UserGrant),
		byPrincipalAndAgent: make(map[string]string),
		grantIDsByAgent:     make(map[string][]string),
	}
}

// Create creates a new user grant or updates existing grant for same principal+agent (upsert semantics).
// Validates the grant before creating/updating.
// Returns deep copy of the created/updated grant.
func (r *UserGrantRepository) Create(ctx context.Context, grant *storage.UserGrant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate ID if not provided
	if grant.ID == "" {
		grant.ID = uuid.New().String()
	}

	// Validate before storing
	if err := grant.ValidateForCreate(); err != nil {
		return storage.NewStorageError(
			"CreateUserGrant",
			storage.ErrorKindValidation,
			err,
			"user grant validation failed",
		)
	}

	// Check if grant already exists for this principal+agent pair (upsert semantics)
	key := principalAgentKey(grant.Principal, grant.AgentID)
	if existingID, exists := r.byPrincipalAndAgent[key]; exists {
		// Update existing grant
		existingGrant := r.grants[existingID]
		existingGrant.ValidUntil = grant.ValidUntil
		existingGrant.DelegatedOAuth2Tokens = grant.DelegatedOAuth2Tokens
		existingGrant.UpdatedAt = grant.UpdatedAt

		// Copy back the existing ID to the provided grant
		grant.ID = existingID
	} else {
		// Store new grant
		r.grants[grant.ID] = grant.Copy()
		r.byPrincipalAndAgent[key] = grant.ID

		// Update agent index for cascade delete
		r.grantIDsByAgent[grant.AgentID] = append(r.grantIDsByAgent[grant.AgentID], grant.ID)
	}

	return nil
}

// Get retrieves a user grant by ID.
// Returns StorageError with Kind=NotFound if grant not found.
// Returns deep copy to prevent external mutation.
func (r *UserGrantRepository) Get(ctx context.Context, id string) (*storage.UserGrant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	grant, exists := r.grants[id]
	if !exists {
		return nil, storage.NewStorageError(
			"GetUserGrant",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"user grant not found",
		)
	}

	return grant.Copy(), nil
}

// Update updates an existing user grant.
// Validates the grant before updating.
// Returns StorageError with Kind=NotFound if grant not found.
func (r *UserGrantRepository) Update(ctx context.Context, grant *storage.UserGrant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if grant exists
	_, exists := r.grants[grant.ID]
	if !exists {
		return storage.NewStorageError(
			"UpdateUserGrant",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"user grant not found",
		)
	}

	// Validate before updating
	if err := grant.Validate(); err != nil {
		return storage.NewStorageError(
			"UpdateUserGrant",
			storage.ErrorKindValidation,
			err,
			"user grant validation failed",
		)
	}

	// Store deep copy
	r.grants[grant.ID] = grant.Copy()

	return nil
}

// Delete deletes a user grant by ID.
// Idempotent: returns nil if grant doesn't exist.
func (r *UserGrantRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get grant to clean up indexes
	if grant, exists := r.grants[id]; exists {
		// Remove from principal+agent index
		key := principalAgentKey(grant.Principal, grant.AgentID)
		delete(r.byPrincipalAndAgent, key)

		// Remove from agent index
		r.removeGrantFromAgentIndex(grant.AgentID, id)

		// Remove grant
		delete(r.grants, id)
	}

	return nil
}

// ListByPrincipalAndAgent retrieves all grants for a principal and specific agent.
// Includes expired grants (filtering happens in ConsentService).
// Returns empty slice if no grants exist (not an error).
// Returns deep copies to prevent external mutation.
func (r *UserGrantRepository) ListByPrincipalAndAgent(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := principalAgentKey(principal, agentID)
	grantID, exists := r.byPrincipalAndAgent[key]
	if !exists {
		return []*storage.UserGrant{}, nil
	}

	grant, exists := r.grants[grantID]
	if !exists {
		return []*storage.UserGrant{}, nil
	}

	return []*storage.UserGrant{grant.Copy()}, nil
}

// FindByPrincipalAndAgent retrieves the grant for a principal and agent pair.
// Returns nil if no grant exists (checked by caller to distinguish from error).
// Returns StorageError with Kind=NotFound if grant not found.
// Returns deep copy to prevent external mutation.
func (r *UserGrantRepository) FindByPrincipalAndAgent(ctx context.Context, principal string, agentID string) (*storage.UserGrant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := principalAgentKey(principal, agentID)
	grantID, exists := r.byPrincipalAndAgent[key]
	if !exists {
		return nil, storage.NewStorageError(
			"FindUserGrant",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"user grant not found",
		)
	}

	grant, exists := r.grants[grantID]
	if !exists {
		return nil, storage.NewStorageError(
			"FindUserGrant",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"user grant not found",
		)
	}

	return grant.Copy(), nil
}

// DeleteByAgent deletes all grants associated with an agent.
// Used during cascade deletion when agent is deleted (FR-021).
// Idempotent: returns nil if agent has no grants.
func (r *UserGrantRepository) DeleteByAgent(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get all grant IDs for this agent
	grantIDs, exists := r.grantIDsByAgent[agentID]
	if !exists || len(grantIDs) == 0 {
		return nil
	}

	// Delete all grants for this agent
	for _, grantID := range grantIDs {
		if grant, exists := r.grants[grantID]; exists {
			// Remove from principal+agent index
			key := principalAgentKey(grant.Principal, grant.AgentID)
			delete(r.byPrincipalAndAgent, key)

			// Remove grant
			delete(r.grants, grantID)
		}
	}

	// Remove agent index
	delete(r.grantIDsByAgent, agentID)

	return nil
}

// principalAgentKey creates a composite key for indexing by principal and agent.
func principalAgentKey(principal string, agentID string) string {
	return principal + ":" + agentID
}

// ListByPrincipal retrieves all active grants for a principal across all agents.
// Filters expired grants (valid_until < NOW()).
// Returns empty slice if no active grants exist (not an error).
// Returns deep copies to prevent external mutation.
func (r *UserGrantRepository) ListByPrincipal(ctx context.Context, principal string) ([]storage.UserGrant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var activeGrants []storage.UserGrant

	// Iterate through all grants and filter by principal
	for _, grant := range r.grants {
		if grant.Principal == principal && grant.IsActive() {
			activeGrants = append(activeGrants, *grant.Copy())
		}
	}

	return activeGrants, nil
}

// CountAgentsByServiceID counts how many agents have delegated OAuth2 tokens for a given service.
// This is used to show dependent agent count when terminating a session.
// Returns the count of distinct agents with delegated_oauth2_tokens JSONB entries for the service.
func (r *UserGrantRepository) CountAgentsByServiceID(ctx context.Context, serviceID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Use a set to track unique agent IDs that have delegated tokens for this service
	uniqueAgents := make(map[string]bool)

	for _, grant := range r.grants {
		// Check if this grant has delegated tokens for the service
		for _, token := range grant.DelegatedOAuth2Tokens {
			if token.ThirdpartyOAuth2ServiceID == serviceID {
				uniqueAgents[grant.AgentID] = true
				break // Only count each agent once
			}
		}
	}

	return len(uniqueAgents), nil
}

// removeGrantFromAgentIndex removes a grant ID from the agent's grant list.
func (r *UserGrantRepository) removeGrantFromAgentIndex(agentID string, grantID string) {
	grantIDs, exists := r.grantIDsByAgent[agentID]
	if !exists {
		return
	}

	// Find and remove grant ID
	for i, id := range grantIDs {
		if id == grantID {
			// Remove by swapping with last element and truncating
			grantIDs[i] = grantIDs[len(grantIDs)-1]
			r.grantIDsByAgent[agentID] = grantIDs[:len(grantIDs)-1]
			break
		}
	}

	// Clean up empty agent index
	if len(r.grantIDsByAgent[agentID]) == 0 {
		delete(r.grantIDsByAgent, agentID)
	}
}
