package memory

import (
	"context"
	"sync"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/google/uuid"
)

// ThirdpartyServiceRepository provides in-memory storage for third-party OAuth2 service entities.
// Thread-safe implementation using sync.RWMutex.
type ThirdpartyServiceRepository struct {
	mu       sync.RWMutex
	services map[string]*storage.ThirdpartyOAuth2Service // ID -> Service
}

// NewThirdpartyServiceRepository creates a new in-memory third-party service repository.
func NewThirdpartyServiceRepository() *ThirdpartyServiceRepository {
	return &ThirdpartyServiceRepository{
		services: make(map[string]*storage.ThirdpartyOAuth2Service),
	}
}

// Create creates a new OAuth2 service configuration in storage.
// Generates a UUID for the service if ID is empty.
// Returns StorageError with Kind=Conflict if service ID already exists.
func (r *ThirdpartyServiceRepository) Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate ID if not provided
	if service.ID == "" {
		service.ID = uuid.New().String()
	}

	// Check for duplicate ID
	if _, exists := r.services[service.ID]; exists {
		return storage.NewStorageError(
			"CreateThirdpartyOAuth2Service",
			storage.ErrorKindConflict,
			nil,
			"service with this ID already exists",
		)
	}

	// Validate before storing
	if err := service.ValidateForCreate(); err != nil {
		return storage.NewStorageError(
			"CreateThirdpartyOAuth2Service",
			storage.ErrorKindValidation,
			err,
			"service validation failed",
		)
	}

	// Store deep copy to prevent external mutation
	r.services[service.ID] = service.Copy()

	return nil
}

// Get retrieves an OAuth2 service configuration by ID.
// Returns StorageError with Kind=NotFound if service not found.
func (r *ThirdpartyServiceRepository) Get(ctx context.Context, id string) (*storage.ThirdpartyOAuth2Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[id]
	if !exists {
		return nil, storage.NewStorageError(
			"GetThirdpartyOAuth2Service",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"service not found",
		)
	}

	// Return deep copy to prevent external mutation
	return service.Copy(), nil
}

// Update updates an existing OAuth2 service configuration.
// Returns StorageError with Kind=NotFound if service ID not found.
func (r *ThirdpartyServiceRepository) Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if service exists
	if _, exists := r.services[service.ID]; !exists {
		return storage.NewStorageError(
			"UpdateThirdpartyOAuth2Service",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"service not found",
		)
	}

	// Validate before updating
	if err := service.Validate(); err != nil {
		return storage.NewStorageError(
			"UpdateThirdpartyOAuth2Service",
			storage.ErrorKindValidation,
			err,
			"service validation failed",
		)
	}

	// Store deep copy
	r.services[service.ID] = service.Copy()

	return nil
}

// Delete deletes an OAuth2 service configuration by ID.
// Idempotent: returns nil if service doesn't exist.
func (r *ThirdpartyServiceRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// For in-memory implementation, we don't track grants
	// So we can simply delete the service
	delete(r.services, id)

	return nil
}

// List retrieves all OAuth2 service configurations.
// Returns empty slice if no services exist (not an error).
func (r *ThirdpartyServiceRepository) List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*storage.ThirdpartyOAuth2Service, 0, len(r.services))
	for _, service := range r.services {
		result = append(result, service.Copy())
	}

	return result, nil
}

// CountGrantsReferencingService returns the number of active grants that reference this service.
// In-memory implementation always returns 0 as we don't track grants in this simple implementation.
func (r *ThirdpartyServiceRepository) CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error) {
	// In-memory implementation does not track grants
	return 0, nil
}
