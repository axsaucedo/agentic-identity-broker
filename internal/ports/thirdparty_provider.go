// Package ports defines interfaces for hexagonal architecture boundaries.
package ports

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
)

// ThirdpartyOAuth2ProviderRepository defines storage operations for third-party OAuth2
// provider entities. All methods operate on ThirdpartyOAuth2ProviderEntity with Secret
// in encrypted state.
//
// IMPORTANT CONTRACT (Encryption Invariant):
//   - Input (Create/Update): Entity must have Secret in encrypted state (Secret.IsEncrypted() == true)
//   - Output (Get/List/Find): Entity has Secret in encrypted state; the domain service decrypts it
//   - The repository is unaware of encryption mechanics; it treats Secret as opaque ciphertext
//
// All third-party OAuth2 provider storage operations use ThirdpartyOAuth2ProviderEntity.

// ProtectedResourceMutationResult is the atomic post-mutation resource state.
// Version is the strong ETag value for the provider after the operation. Changed
// is false only for successful idempotent or no-op mutations.
type ProtectedResourceMutationResult struct {
	Resource           string
	ProtectedResources []string
	Version            int64
	Changed            bool
}
type ThirdpartyOAuth2ProviderRepository interface {
	// Create stores a new provider entity.
	// Entity.Secret must be in encrypted state before calling.
	// Returns error if:
	//   - Entity ID already exists (StorageError with Kind=Conflict)
	//   - Secret is not in encrypted state (StorageError with Kind=Validation)
	//   - Storage connection fails (StorageError with Kind=Connection)
	Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error

	// Get retrieves a provider entity by ID.
	// Returns entity with Secret in encrypted state.
	// Returns StorageError with Kind=NotFound if provider not found.
	Get(ctx context.Context, id id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error)

	// Update updates an existing provider entity. expectedVersion is required only
	// when entity.ProtectedResources replaces the complete resource set.
	// Entity.Secret must be in encrypted state before calling. On success, entity
	// contains its stored CreatedAt and new Version values.
	// Returns StorageError with Kind=NotFound if provider not found and Conflict for
	// a stale expectedVersion.
	Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity, expectedVersion *int64) error

	// Delete removes a provider entity by ID.
	// Idempotent: returns nil if provider doesn't exist.
	Delete(ctx context.Context, id id.ServiceID) error

	// List retrieves all provider entities.
	// Returns entities with Secret in encrypted state.
	// Returns empty slice if no providers exist (not an error).
	List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error)

	// CountGrantsReferencingService returns the number of active grants referencing this provider.
	// Used to enforce FR-022 (block deletion if grants exist).
	// Returns 0 if no grants reference the provider.
	CountGrantsReferencingService(ctx context.Context, serviceID id.ServiceID) (int, error)

	// FindByProtectedResource retrieves a provider by matching resource URI against
	// protected_resources field. The resourceURI must be normalized (trailing slashes removed).
	// Returns entity with Secret in encrypted state.
	// Returns InvalidTargetError if no provider matches or if multiple providers match.
	FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error)

	// AddProtectedResource atomically claims resourceURI for serviceID. A successful
	// replay owned by the same service returns Changed=false.
	// Returns StorageError with Kind=Conflict if another service owns the URI, or
	// Kind=NotFound if serviceID does not exist.
	AddProtectedResource(ctx context.Context, serviceID id.ServiceID, resourceURI string) (ProtectedResourceMutationResult, error)

	// RemoveProtectedResource atomically releases resourceURI from serviceID.
	// Returns StorageError with Kind=NotFound if the service or resource is absent.
	RemoveProtectedResource(ctx context.Context, serviceID id.ServiceID, resourceURI string) (ProtectedResourceMutationResult, error)

	// RenameProtectedResource atomically replaces fromURI with toURI on serviceID.
	// Returns StorageError with Kind=Conflict if toURI is owned and Kind=NotFound
	// if the service or fromURI is absent.
	RenameProtectedResource(ctx context.Context, serviceID id.ServiceID, fromURI, toURI string) (ProtectedResourceMutationResult, error)

	// ListProtectedResources returns the normalized resource URIs and current strong
	// ETag version for serviceID.
	// Returns StorageError with Kind=NotFound if serviceID does not exist.
	ListProtectedResources(ctx context.Context, serviceID id.ServiceID) ([]string, int64, error)
}
