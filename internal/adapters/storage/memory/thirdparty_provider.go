package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// InMemoryThirdpartyOAuth2ProviderRepository implements ports.ThirdpartyOAuth2ProviderRepository
// with thread-safe in-memory storage using sync.RWMutex.
//
// Invariant: entities stored in the map always have Secret in encrypted state.
// The domain service encrypts before storing and decrypts after retrieving.
type InMemoryThirdpartyOAuth2ProviderRepository struct {
	mu        sync.RWMutex
	providers map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity
}

// NewInMemoryThirdpartyOAuth2ProviderRepository creates a new in-memory provider repository.
func NewInMemoryThirdpartyOAuth2ProviderRepository() *InMemoryThirdpartyOAuth2ProviderRepository {
	return &InMemoryThirdpartyOAuth2ProviderRepository{
		providers: make(map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity),
	}
}

// Create stores a new provider entity in memory.
// Entity.ID must be set by the caller before storing (the domain service generates it before encrypting).
// Entity.Secret must be in encrypted state.
func (r *InMemoryThirdpartyOAuth2ProviderRepository) Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entity.ID.IsZero() {
		return storage.NewStorageError("CreateThirdpartyOAuth2Provider",
			storage.ErrorKindValidation, nil,
			"provider ID cannot be empty: caller must set ID before storing")
	}

	if _, exists := r.providers[entity.ID]; exists {
		return storage.NewStorageError("CreateThirdpartyOAuth2Provider", storage.ErrorKindConflict, nil, "provider with this ID already exists")
	}

	stored, err := providerEntityCopy(entity)
	if err != nil {
		return storage.NewStorageError("CreateThirdpartyOAuth2Provider", storage.ErrorKindValidation, err, "failed to store provider")
	}

	r.providers[entity.ID] = stored
	return nil
}

// Get retrieves a provider entity by ID.
// Returns entity with Secret in encrypted state.
func (r *InMemoryThirdpartyOAuth2ProviderRepository) Get(ctx context.Context, serviceID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, exists := r.providers[serviceID]
	if !exists {
		return nil, storage.NewStorageError("GetThirdpartyOAuth2Provider", storage.ErrorKindNotFound, ports.ErrNotFound, "provider not found")
	}

	return entity.Copy(), nil
}

// Update updates an existing provider entity in memory.
// Entity.Secret must be in encrypted state.
// On success, entity.CreatedAt is set to the value from the stored entity.
func (r *InMemoryThirdpartyOAuth2ProviderRepository) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.providers[entity.ID]
	if !exists {
		return storage.NewStorageError("UpdateThirdpartyOAuth2Provider", storage.ErrorKindNotFound, ports.ErrNotFound, "provider not found")
	}

	// Preserve immutable created_at from storage — callers must not set it.
	entity.CreatedAt = existing.CreatedAt

	stored, err := providerEntityCopy(entity)
	if err != nil {
		return storage.NewStorageError("UpdateThirdpartyOAuth2Provider", storage.ErrorKindValidation, err, "failed to store provider")
	}

	r.providers[entity.ID] = stored
	return nil
}

// Delete removes a provider entity by ID.
// Idempotent: returns nil if provider doesn't exist.
func (r *InMemoryThirdpartyOAuth2ProviderRepository) Delete(ctx context.Context, serviceID id.ServiceID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.providers, serviceID)
	return nil
}

// List retrieves all provider entities.
// Returns entities with Secret in encrypted state.
func (r *InMemoryThirdpartyOAuth2ProviderRepository) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.ThirdpartyOAuth2ProviderEntity, 0, len(r.providers))
	for _, entity := range r.providers {
		result = append(result, entity.Copy())
	}

	return result, nil
}

// CountGrantsReferencingService always returns 0.
// The in-memory adapter does not track grants.
func (r *InMemoryThirdpartyOAuth2ProviderRepository) CountGrantsReferencingService(_ context.Context, _ id.ServiceID) (int, error) {
	return 0, nil
}

// FindByProtectedResource retrieves a provider by matching resource URI
// against protected_resources (case-sensitive).
func (r *InMemoryThirdpartyOAuth2ProviderRepository) FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	select {
	case <-ctx.Done():
		return nil, storage.NewStorageError("FindByProtectedResource", storage.ErrorKindTimeout, ctx.Err(), "context cancelled or timeout")
	default:
	}

	if resourceURI == "" {
		return nil, tokenexchange.NewInvalidTargetError("resource parameter cannot be empty")
	}

	var matchingEntity *model.ThirdpartyOAuth2ProviderEntity
	matchCount := 0

	for _, entity := range r.providers {
		if slices.Contains(entity.ProtectedResources, resourceURI) {
			matchingEntity = entity
			matchCount++
		}
	}

	if matchCount == 0 {
		return nil, tokenexchange.NewInvalidTargetError("no service configured for the requested resource")
	}
	if matchCount > 1 {
		return nil, tokenexchange.NewInvalidTargetError("multiple services configured for the same resource")
	}

	return matchingEntity.Copy(), nil
}
