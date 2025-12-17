// Package ports defines interfaces for hexagonal architecture boundaries.
package ports

import (
	"context"
	"errors"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// Sentinel errors for storage operations
var (
	// ErrNotFound is returned when a requested entity does not exist.
	ErrNotFound = errors.New("entity not found")
)

// StorageLifecycle manages storage backend lifecycle operations.
// Handles initialization, health checking, and cleanup.
// This interface isolates storage lifecycle concerns from CRUD operations.
type StorageLifecycle interface {
	// Initialize performs storage backend initialization and verification.
	// For PostgreSQL: connects to database, verifies schema version
	// For in-memory: initializes empty storage structures
	// Returns error if initialization fails (FR-007).
	// Applications MUST fail startup if Initialize returns an error.
	Initialize(ctx context.Context) error

	// Close gracefully closes storage connections and releases resources.
	// Should be called during application shutdown.
	// Idempotent: safe to call multiple times.
	// Best effort - does not fail startup if Close fails.
	Close(ctx context.Context) error

	// HealthCheck verifies storage backend is operational.
	// Lightweight operation for monitoring (5-20ms typical).
	// Returns nil if healthy, error with details if unhealthy.
	// Used by monitoring and readiness probes.
	HealthCheck(ctx context.Context) error
}

// User represents a user entity in the storage layer.
// Domain entity - does not expose storage implementation details.
type User struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// UserFilter represents optional filtering criteria for ListUsers operations.
// Allows consumers to filter results by common fields.
type UserFilter struct {
	EmailPattern string // SQL LIKE pattern for email filtering (optional)
	Limit        int    // Maximum results to return (0 = no limit)
	Offset       int    // Number of results to skip (pagination)
}

// UserRepository defines storage operations for user entities.
// Follows Repository pattern with focused CRUD operations.
// This interface isolates domain logic from storage implementation.
// Following Interface Segregation Principle: small, focused interfaces.
type UserRepository interface {
	// CreateUser creates a new user entity in storage.
	// Returns error if:
	// - User ID already exists (StorageError with Kind=Conflict)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	CreateUser(ctx context.Context, user *User) error

	// GetUser retrieves a user entity by ID.
	// Returns nil if user not found (checked by caller to distinguish from error).
	// Returns StorageError with Kind=NotFound if user not found.
	// Returns StorageError for connection/timeout issues.
	GetUser(ctx context.Context, id string) (*User, error)

	// UpdateUser updates an existing user entity.
	// Returns error if:
	// - User ID not found (StorageError with Kind=NotFound)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	UpdateUser(ctx context.Context, user *User) error

	// DeleteUser deletes a user entity by ID.
	// Returns error if storage operation fails.
	// It is safe to delete non-existent users (idempotent).
	DeleteUser(ctx context.Context, id string) error

	// ListUsers retrieves multiple user entities matching optional filter.
	// Returns empty slice if no users match filter (not an error).
	// Returns StorageError for connection/timeout issues.
	ListUsers(ctx context.Context, filter *UserFilter) ([]*User, error)
}

// AgentRepository defines storage operations for agent entities.
// Agents represent AI agents registered in the identity broker.
// Following Interface Segregation Principle: focused interface for agent operations.
type AgentRepository interface {
	// Create creates a new agent entity in storage.
	// The agent ID should be generated before calling this method.
	// Returns error if:
	// - Agent ID already exists (StorageError with Kind=Conflict)
	// - Agent client_id already exists (StorageError with Kind=Conflict)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	Create(ctx context.Context, agent *storage.Agent) error

	// Get retrieves an agent entity by ID.
	// Returns StorageError with Kind=NotFound if agent not found.
	// Returns StorageError for connection/timeout issues.
	Get(ctx context.Context, id string) (*storage.Agent, error)

	// Update updates an existing agent entity.
	// Returns error if:
	// - Agent ID not found (StorageError with Kind=NotFound)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	Update(ctx context.Context, agent *storage.Agent) error

	// Delete deletes an agent entity by ID.
	// Associated grants are CASCADE deleted per FR-021.
	// Returns error if storage operation fails.
	// It is safe to delete non-existent agents (idempotent).
	Delete(ctx context.Context, id string) error

	// List retrieves all agent entities.
	// Returns empty slice if no agents exist (not an error).
	// Returns StorageError for connection/timeout issues.
	List(ctx context.Context) ([]*storage.Agent, error)
}

// ThirdpartyOAuth2ServiceRepository defines storage operations for third-party OAuth2 service configurations.
// These services represent external OAuth2 providers (GitHub, Google, etc.) that agents can access.
// Client secrets are encrypted at rest using the EncryptionPort.
type ThirdpartyOAuth2ServiceRepository interface {
	// Create creates a new OAuth2 service configuration.
	// The service ID should be generated before calling this method.
	// Client secret will be encrypted using the configured EncryptionPort.
	// Returns error if:
	// - Service ID already exists (StorageError with Kind=Conflict)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Encryption fails (StorageError with Kind=Internal)
	// - Operation timeout (StorageError with Kind=Timeout)
	Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error

	// Get retrieves an OAuth2 service configuration by ID.
	// Client secret will be decrypted using the configured EncryptionPort.
	// Returns StorageError with Kind=NotFound if service not found.
	// Returns StorageError for connection/timeout/decryption issues.
	Get(ctx context.Context, id string) (*storage.ThirdpartyOAuth2Service, error)

	// Update updates an existing OAuth2 service configuration.
	// Client secret will be encrypted using the configured EncryptionPort.
	// Returns error if:
	// - Service ID not found (StorageError with Kind=NotFound)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Encryption fails (StorageError with Kind=Internal)
	// - Operation timeout (StorageError with Kind=Timeout)
	Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error

	// Delete deletes an OAuth2 service configuration by ID.
	// Returns error if:
	// - Active grants reference this service (StorageError with Kind=Conflict, per FR-022)
	// - Storage operation fails
	// It is safe to delete non-existent services (idempotent).
	Delete(ctx context.Context, id string) error

	// List retrieves all OAuth2 service configurations.
	// Client secrets will be decrypted for each service.
	// Returns empty slice if no services exist (not an error).
	// Returns StorageError for connection/timeout/decryption issues.
	List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error)

	// CountGrantsReferencingService returns the number of active grants that reference this service.
	// Used to enforce FR-022 (block service deletion if grants exist).
	// Returns 0 if no grants reference the service.
	// Returns StorageError for connection/timeout issues.
	CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error)
}

// UserGrantRepository defines storage operations for user grant entities.
// Grants represent users delegating permissions to agents for third-party services.
// One grant per user-agent pair (upsert semantics).
type UserGrantRepository interface {
	// Create creates a new user grant or updates existing grant for same principal+agent (upsert).
	// The grant ID should be generated before calling this method.
	// Returns error if:
	// - Agent ID doesn't exist (StorageError with Kind=NotFound)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	Create(ctx context.Context, grant *storage.UserGrant) error

	// Get retrieves a user grant by ID.
	// Returns StorageError with Kind=NotFound if grant not found.
	// Returns StorageError for connection/timeout issues.
	Get(ctx context.Context, id string) (*storage.UserGrant, error)

	// Update updates an existing user grant.
	// Returns error if:
	// - Grant ID not found (StorageError with Kind=NotFound)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	Update(ctx context.Context, grant *storage.UserGrant) error

	// Delete deletes a user grant by ID.
	// Returns error if storage operation fails.
	// It is safe to delete non-existent grants (idempotent).
	Delete(ctx context.Context, id string) error

	// ListByPrincipalAndAgent retrieves all active grants for a principal and specific agent.
	// Filters expired grants (valid_until < NOW()).
	// Returns empty slice if no active grants exist (not an error).
	// Returns StorageError for connection/timeout issues.
	ListByPrincipalAndAgent(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error)

	// FindByPrincipalAndAgent retrieves the grant for a principal and agent (if exists).
	// Returns nil if no grant exists (checked by caller to distinguish from error).
	// Returns StorageError with Kind=NotFound if grant not found.
	// Returns StorageError for connection/timeout issues.
	FindByPrincipalAndAgent(ctx context.Context, principal string, agentID string) (*storage.UserGrant, error)

	// DeleteByAgent deletes all grants associated with an agent.
	// Used during cascade deletion when agent is deleted (FR-021).
	// Returns error if storage operation fails.
	// It is safe to call with non-existent agent (idempotent).
	DeleteByAgent(ctx context.Context, agentID string) error
}
