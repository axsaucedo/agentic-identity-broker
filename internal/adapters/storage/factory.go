// Package storage implements storage adapters for different backends.
// Adapters implement the ports.StorageLifecycle and ports.UserRepository interfaces.
package storage

import (
	"fmt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/postgres"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Adapter composes storage functionality.
// Adapters implement both StorageLifecycle and repository interfaces (UserRepository, etc.)
// This struct is returned by NewAdapter factory function.
type Adapter struct {
	lifecycle   ports.StorageLifecycle
	users       ports.UserRepository
	agents      ports.AgentRepository
	services    ports.ThirdpartyOAuth2ServiceRepository
	userGrants  ports.UserGrantRepository
}

// NewAdapter creates a storage adapter based on configuration.
// Returns appropriate adapter implementation (memory or PostgreSQL).
// Factory pattern allows backend selection without domain logic knowing details.
//
// Returns error if:
// - Backend type is unsupported
// - Backend-specific initialization fails
func NewAdapter(config *ports.StorageConfig) (*Adapter, error) {
	backend := storage.StorageBackend(config.Backend)

	switch backend {
	case storage.BackendMemory:
		return newMemoryAdapter(config)

	case storage.BackendPostgres:
		return newPostgresAdapter(config)

	default:
		return nil, fmt.Errorf("unsupported storage backend: %s", config.Backend)
	}
}

// newMemoryAdapter creates an in-memory storage adapter.
func newMemoryAdapter(config *ports.StorageConfig) (*Adapter, error) {
	memAdapter := memory.NewAdapter()
	return &Adapter{
		lifecycle:   memAdapter,
		users:       memAdapter,
		agents:      memory.NewAgentRepository(),
		services:    memory.NewThirdpartyServiceRepository(),
		userGrants:  memory.NewUserGrantRepository(),
	}, nil
}

// newPostgresAdapter creates a PostgreSQL storage adapter.
func newPostgresAdapter(config *ports.StorageConfig) (*Adapter, error) {
	pgAdapter, err := postgres.NewAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL adapter: %w", err)
	}
	return &Adapter{
		lifecycle:   pgAdapter,
		users:       pgAdapter,
		agents:      postgres.NewAgentRepository(pgAdapter),
		services:    postgres.NewThirdpartyServiceRepository(pgAdapter, nil), // TODO: Initialize proper EncryptionPort
		userGrants:  postgres.NewUserGrantRepository(pgAdapter),
	}, nil
}

// Lifecycle returns the StorageLifecycle interface implementation.
// Used for Initialize, Close, and HealthCheck operations.
func (a *Adapter) Lifecycle() ports.StorageLifecycle {
	return a.lifecycle
}

// Users returns the UserRepository interface implementation.
// Used for CRUD operations on user entities.
func (a *Adapter) Users() ports.UserRepository {
	return a.users
}

// Agents returns the AgentRepository interface implementation.
// Used for agent entity CRUD operations.
func (a *Adapter) Agents() ports.AgentRepository {
	return a.agents
}

// Services returns the ThirdpartyOAuth2ServiceRepository interface implementation.
// Used for OAuth2 service configuration CRUD operations.
func (a *Adapter) Services() ports.ThirdpartyOAuth2ServiceRepository {
	return a.services
}

// UserGrants returns the UserGrantRepository interface implementation.
// Used for user grant CRUD operations.
func (a *Adapter) UserGrants() ports.UserGrantRepository {
	return a.userGrants
}
