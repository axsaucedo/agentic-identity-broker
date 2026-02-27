package thirdparty

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// ThirdpartyOAuth2ProviderService is the consolidated domain service for managing external
// OAuth2 providers. It merges encryption, decryption, and branch key provisioning into a
// single cohesive domain service following the Single Responsibility Principle at the domain
// service level.
//
// Encryption lifecycle:
//   - Create: validates entity, provisions branch key (if manager present),
//     encrypts Secret{plaintext} → Secret{ciphertext}, stores entity
//   - Get/List/Find: retrieves entity with Secret{ciphertext}, decrypts to Secret{plaintext}
//   - Update: if Secret is plaintext (new secret), encrypts before storing;
//     if Secret is encrypted (no change), stores as-is
//
// The repository (ThirdpartyOAuth2ProviderRepository) is unaware of encryption mechanics
// and treats Secret ciphertext as opaque binary data.
type ThirdpartyOAuth2ProviderService struct {
	repo             ports.ThirdpartyOAuth2ProviderRepository
	encryption       ports.EncryptionPort
	branchKeyManager ports.BranchKeyManager // may be nil
	logger           *slog.Logger
}

// NewThirdpartyOAuth2ProviderService creates a new provider service.
// branchKeyManager may be nil if no KMS backend is configured.
func NewThirdpartyOAuth2ProviderService(
	repo ports.ThirdpartyOAuth2ProviderRepository,
	encryption ports.EncryptionPort,
	branchKeyManager ports.BranchKeyManager,
	logger *slog.Logger,
) *ThirdpartyOAuth2ProviderService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ThirdpartyOAuth2ProviderService{
		repo:             repo,
		encryption:       encryption,
		branchKeyManager: branchKeyManager,
		logger:           logger,
	}
}

// Create validates, optionally provisions a branch key, encrypts the secret, and stores the entity.
// entity.Secret must be in plaintext state on entry.
// On success, entity.Secret is in encrypted state.
func (s *ThirdpartyOAuth2ProviderService) Create(
	ctx context.Context,
	entity *model.ThirdpartyOAuth2ProviderEntity,
) error {
	// Generate ID before encryption so context binding matches persisted ID
	if entity.ID == "" {
		entity.ID = uuid.New().String()
	}

	// Provision branch key before creating service (fail-fast on error)
	if s.branchKeyManager != nil {
		s.logger.Info("provisioning branch key for service", "service_id", entity.ID)
		branchKeyID, err := s.branchKeyManager.Create(ctx, entity.ID)
		if err != nil {
			s.logger.Error("failed to provision branch key",
				"service_id", entity.ID,
				"error", err)
			return fmt.Errorf("branch key provisioning failed: %w", err)
		}
		s.logger.Info("branch key provisioned",
			"service_id", entity.ID,
			"branch_key_id", branchKeyID)
	}

	// Extract plaintext secret
	plaintext, err := entity.Secret.GetPlaintext()
	if err != nil {
		return fmt.Errorf("entity secret must be in plaintext state for create: %w", err)
	}

	// Build encryption context with service_id only (ADR 008)
	encContext := map[string]string{"service_id": entity.ID}

	// Encrypt secret
	ciphertext, err := s.encryption.Encrypt(ctx, []byte(plaintext), encContext)
	if err != nil {
		s.logger.Error("encryption_failed",
			"operation", "create_provider",
			"service_id", entity.ID,
			"reason", err)
		return fmt.Errorf("failed to encrypt client secret: %w", err)
	}

	// Transition secret from plaintext to encrypted state
	entity.Secret = model.NewEncryptedSecret(ciphertext)

	s.logger.Info("service_secret_encrypted",
		"operation", "create",
		"service_id", entity.ID)

	if err := s.repo.Create(ctx, entity); err != nil {
		return fmt.Errorf("failed to store provider: %w", err)
	}

	s.logger.Info("provider_created",
		"service_id", entity.ID,
		"display_name", entity.DisplayName)

	return nil
}

// Get retrieves a provider by ID and decrypts its secret.
// Returns entity with Secret in plaintext state.
func (s *ThirdpartyOAuth2ProviderService) Get(
	ctx context.Context,
	serviceID string,
) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	entity, err := s.repo.Get(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Guard: verify ID consistency for data integrity
	if entity.ID != serviceID {
		s.logger.Error("service_id_mismatch",
			"expected_id", serviceID,
			"actual_id", entity.ID)
		return nil, fmt.Errorf("provider ID mismatch: expected %s, got %s", serviceID, entity.ID)
	}

	return s.decryptSecret(ctx, entity)
}

// Update encrypts the secret if changed (plaintext) and stores the updated entity.
// If entity.Secret is in plaintext state, it will be encrypted before storing.
// If entity.Secret is in encrypted state (no change), it is stored as-is.
func (s *ThirdpartyOAuth2ProviderService) Update(
	ctx context.Context,
	entity *model.ThirdpartyOAuth2ProviderEntity,
) error {
	if entity.Secret.IsPlaintext() {
		plaintext, err := entity.Secret.GetPlaintext()
		if err != nil {
			return fmt.Errorf("failed to read plaintext secret for update: %w", err)
		}

		encContext := map[string]string{"service_id": entity.ID}
		ciphertext, err := s.encryption.Encrypt(ctx, []byte(plaintext), encContext)
		if err != nil {
			s.logger.Error("encryption_failed",
				"operation", "update_provider",
				"service_id", entity.ID,
				"reason", err)
			return fmt.Errorf("failed to encrypt client secret: %w", err)
		}

		entity.Secret = model.NewEncryptedSecret(ciphertext)
		s.logger.Info("service_secret_encrypted",
			"operation", "update",
			"service_id", entity.ID)
	}

	return s.repo.Update(ctx, entity)
}

// List retrieves all providers and decrypts their secrets.
// Fails fast on first decryption error for operational visibility.
func (s *ThirdpartyOAuth2ProviderService) List(
	ctx context.Context,
) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	entities, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	decrypted := make([]*model.ThirdpartyOAuth2ProviderEntity, 0, len(entities))
	for _, entity := range entities {
		dec, err := s.decryptSecret(ctx, entity)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt provider %s: %w", entity.ID, err)
		}
		decrypted = append(decrypted, dec)
	}

	return decrypted, nil
}

// Delete removes a provider from storage.
func (s *ThirdpartyOAuth2ProviderService) Delete(
	ctx context.Context,
	serviceID string,
) error {
	if err := s.repo.Delete(ctx, serviceID); err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	s.logger.Info("provider_deleted", "service_id", serviceID)
	return nil
}

// FindByProtectedResource retrieves a provider by resource URI and decrypts its secret.
func (s *ThirdpartyOAuth2ProviderService) FindByProtectedResource(
	ctx context.Context,
	resourceURI string,
) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	entity, err := s.repo.FindByProtectedResource(ctx, resourceURI)
	if err != nil {
		return nil, err
	}

	dec, err := s.decryptSecret(ctx, entity)
	if err != nil {
		s.logger.Error("decryption_failed",
			"operation", "find_by_protected_resource",
			"service_id", entity.ID,
			"resource_uri", resourceURI,
			"reason", err)
		return nil, err
	}

	s.logger.Info("service_secret_decrypted",
		"operation", "find_by_protected_resource",
		"service_id", entity.ID,
		"resource_uri", resourceURI)

	return dec, nil
}

// CountGrantsReferencingService returns the number of grants referencing this provider.
func (s *ThirdpartyOAuth2ProviderService) CountGrantsReferencingService(
	ctx context.Context,
	serviceID string,
) (int, error) {
	return s.repo.CountGrantsReferencingService(ctx, serviceID)
}

// decryptSecret decrypts the entity's Secret field in place (returns a copy with plaintext).
func (s *ThirdpartyOAuth2ProviderService) decryptSecret(
	ctx context.Context,
	entity *model.ThirdpartyOAuth2ProviderEntity,
) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	ciphertext, err := entity.Secret.GetCiphertext()
	if err != nil {
		return nil, fmt.Errorf("entity has no encrypted secret: %w", err)
	}

	encContext := map[string]string{"service_id": entity.ID}

	plaintext, err := s.encryption.Decrypt(ctx, ciphertext, encContext)
	if err != nil {
		s.logger.Error("decryption_failed",
			"operation", "decrypt_secret",
			"service_id", entity.ID,
			"reason", err)
		return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
	}

	// Work on a copy to avoid mutating the entity returned by the repository
	result := entity.Copy()
	result.Secret = model.NewPlaintextSecret(string(plaintext))

	s.logger.Info("service_secret_decrypted",
		"operation", "get",
		"service_id", entity.ID)

	return result, nil
}
