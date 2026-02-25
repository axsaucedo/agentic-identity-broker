package memory

import (
	"errors"
	"fmt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
)

// The memory adapter stores ThirdpartyOAuth2ProviderEntity values directly in a map,
// using deep copies to prevent external mutation. There is no intermediate record struct
// because memory storage requires no serialization — entities are held in process memory.
//
// Invariant: entities stored in the map always have their Secret in encrypted state.
// The domain service (ThirdpartyOAuth2ProviderService) encrypts before storing and
// decrypts after retrieving. The memory adapter is unaware of encryption mechanics.

// providerEntityCopy creates a deep copy of a ThirdpartyOAuth2ProviderEntity for
// internal storage. The entity's Secret must be in encrypted state before storing.
// Returns an error if the entity is nil or the secret is not encrypted.
func providerEntityCopy(entity *model.ThirdpartyOAuth2ProviderEntity) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	if entity == nil {
		return nil, errors.New("entity cannot be nil")
	}

	if !entity.Secret.IsEncrypted() {
		return nil, fmt.Errorf("entity secret must be encrypted before storing in memory adapter")
	}

	return entity.Copy(), nil
}
