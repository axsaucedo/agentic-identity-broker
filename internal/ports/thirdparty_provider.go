// Package ports defines interfaces for hexagonal architecture boundaries.
package ports

import (
	"context"

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
	Get(ctx context.Context, id string) (*model.ThirdpartyOAuth2ProviderEntity, error)

	// Update updates an existing provider entity.
	// Entity.Secret must be in encrypted state before calling.
	// Returns StorageError with Kind=NotFound if provider not found.
	Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error

	// Delete removes a provider entity by ID.
	// Idempotent: returns nil if provider doesn't exist.
	Delete(ctx context.Context, id string) error

	// List retrieves all provider entities.
	// Returns entities with Secret in encrypted state.
	// Returns empty slice if no providers exist (not an error).
	List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error)

	// CountGrantsReferencingService returns the number of active grants referencing this provider.
	// Used to enforce FR-022 (block deletion if grants exist).
	// Returns 0 if no grants reference the provider.
	CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error)

	// FindByProtectedResource retrieves a provider by matching resource URI against
	// protected_resources field. The resourceURI must be normalized (trailing slashes removed).
	// Returns entity with Secret in encrypted state.
	// Returns InvalidTargetError if no provider matches or if multiple providers match.
	FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error)
}
