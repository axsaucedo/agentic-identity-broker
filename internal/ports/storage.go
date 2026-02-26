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

// HealthChecker verifies storage backend health.
type HealthChecker interface {
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

	// GetByClientID retrieves an agent entity by client_id.
	// Returns StorageError with Kind=NotFound if agent not found.
	// Returns StorageError for connection/timeout issues.
	GetByClientID(ctx context.Context, clientID string) (*storage.Agent, error)
}

// ThirdpartyOAuth2ServiceRepository defines storage operations for third-party OAuth2 service configurations.
// These services represent external OAuth2 providers (GitHub, Google, etc.) that agents can access.
//
// IMPORTANT CONTRACT (Domain Layer Encryption):
// - Input (Create/Update): Service with ClientSecretCiphertext (encrypted by caller)
// - Output (Get/List): Service with ClientSecretCiphertext (encrypted bytes, opaque)
// - Caller (ThirdpartyServiceManager) is responsible for encrypt/decrypt lifecycle
//
// The repository treats ClientSecretCiphertext as opaque binary data and MUST NOT
// attempt to decrypt, re-encrypt, or make assumptions about its content. This maintains
// hexagonal architecture purity where the storage adapter is unaware of encryption mechanics.
type ThirdpartyOAuth2ServiceRepository interface {
	// Create creates a new OAuth2 service configuration.
	// The service ID should be generated before calling this method.
	// Expects ClientSecretCiphertext to be pre-encrypted by ThirdpartyServiceManager.
	// Returns error if:
	// - Service ID already exists (StorageError with Kind=Conflict)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error

	// Get retrieves an OAuth2 service configuration by ID.
	// Returns ClientSecretCiphertext as opaque encrypted bytes.
	// Returns StorageError with Kind=NotFound if service not found.
	// Returns StorageError for connection/timeout issues.
	Get(ctx context.Context, id string) (*storage.ThirdpartyOAuth2Service, error)

	// Update updates an existing OAuth2 service configuration.
	// Expects ClientSecretCiphertext to be pre-encrypted by ThirdpartyServiceManager.
	// Returns error if:
	// - Service ID not found (StorageError with Kind=NotFound)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error

	// Delete deletes an OAuth2 service configuration by ID.
	// Returns error if:
	// - Active grants reference this service (StorageError with Kind=Conflict, per FR-022)
	// - Storage operation fails
	// It is safe to delete non-existent services (idempotent).
	Delete(ctx context.Context, id string) error

	// List retrieves all OAuth2 service configurations.
	// Returns ClientSecretCiphertext as opaque encrypted bytes.
	// Returns empty slice if no services exist (not an error).
	// Returns StorageError for connection/timeout issues.
	List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error)

	// CountGrantsReferencingService returns the number of active grants that reference this service.
	// Used to enforce FR-022 (block service deletion if grants exist).
	// Returns 0 if no grants reference the service.
	// Returns StorageError for connection/timeout issues.
	CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error)

	// FindByProtectedResource retrieves an OAuth2 service configuration by matching resource URI
	// against protected_resources field. Used for resource-based service discovery in token exchange.
	// The resourceURI parameter MUST be normalized before calling using tokenexchange.Normalize()
	// to remove trailing slashes for consistent matching.
	// Example: normalizedURI := tokenexchange.Normalize(requestURI)
	// Returns the service whose protected_resources contains the resourceURI (case-sensitive match).
	// Returns ClientSecretCiphertext as opaque encrypted bytes.
	// Returns error if:
	// - No service configured with matching protected_resources (InvalidTargetError: "No service configured for the requested resource")
	// - Multiple services match the same resource (misconfiguration) (InvalidTargetError: "Multiple services configured for the same resource")
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	FindByProtectedResource(ctx context.Context, resourceURI string) (*storage.ThirdpartyOAuth2Service, error)
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

	// ListByPrincipal retrieves all active grants for a principal across all agents.
	// Filters expired grants (valid_until < NOW()).
	// Returns empty slice if no active grants exist (not an error).
	// Returns StorageError for connection/timeout issues.
	ListByPrincipal(ctx context.Context, principal string) ([]storage.UserGrant, error)

	// CountAgentsByServiceID counts how many agents have delegated OAuth2 tokens for a given service.
	// This is used to show dependent agent count when terminating a session.
	// Returns the count of distinct agents with delegated_oauth2_tokens JSONB entries for the service.
	CountAgentsByServiceID(ctx context.Context, serviceID string) (int, error)

	// ListByServiceID retrieves all agent IDs that have delegated OAuth2 tokens for a given service.
	// This is used to show the actual dependent agents when terminating a session.
	// Returns the list of distinct agent IDs with delegated_oauth2_tokens JSONB entries for the service.
	// Returns empty slice if no agents have delegated tokens for the service.
	// Returns StorageError for connection/timeout issues.
	ListByServiceID(ctx context.Context, serviceID string) ([]string, error)
}

// UserSessionRepository stores and retrieves user OAuth2 sessions.
//
// IMPORTANT: This repository operates on encrypted tokens. The contract is:
// - Input (Create): UserSession with tokens ALREADY ENCRYPTED by caller (OAuth2SessionService)
// - Output (Get): UserSession with tokens STILL ENCRYPTED (opaque binary data)
// - Caller (OAuth2SessionService) is responsible for decrypt/encrypt lifecycle
//
// This design maintains hexagonal architecture purity: the storage adapter is unaware
// of encryption concerns and focuses solely on persistence.
//
// Encryption Context (EncryptionContext field):
// - Stored as JSONB for audit/debugging purposes
// - Contains {"service_id": "<oauth2-service-id>"}
// - Used by OAuth2SessionService.Decrypt* methods as Additional Authenticated Data (AAD)
//
// One session per (principal, service_id) pair.
type UserSessionRepository interface {
	// Create creates a new user session.
	// Uses upsert semantics: if session exists for (principal, service_id), replaces tokens.
	// Returns error if:
	// - Service ID doesn't exist (StorageError with Kind=NotFound via FK constraint)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	Create(ctx context.Context, session *storage.UserSession) error

	// Get retrieves a session by ID.
	// Returns StorageError with Kind=NotFound if session not found.
	Get(ctx context.Context, id string) (*storage.UserSession, error)

	// FindByPrincipalAndService retrieves the session for a principal and service.
	// Returns nil if no session exists (not an error).
	FindByPrincipalAndService(ctx context.Context, principal, serviceID string) (*storage.UserSession, error)

	// ListByPrincipal retrieves all sessions for a principal.
	// Returns empty slice if no sessions exist (not an error).
	ListByPrincipal(ctx context.Context, principal string) ([]*storage.UserSession, error)

	// Delete deletes a session by ID.
	// Returns error if storage operation fails.
	// Idempotent: safe to delete non-existent session.
	Delete(ctx context.Context, id string) error

	// DeleteByPrincipalAndService deletes the session for a principal and service.
	// Returns error if storage operation fails.
	// Idempotent: safe to delete non-existent session.
	DeleteByPrincipalAndService(ctx context.Context, principal, serviceID string) error

	// CountByService counts sessions referencing a service.
	// Used to enforce deletion protection (cannot delete service with active sessions).
	CountByService(ctx context.Context, serviceID string) (int, error)
}
