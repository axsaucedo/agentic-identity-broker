package thirdparty

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// ThirdpartyOAuth2ProviderService is the consolidated domain service for managing external
// OAuth2 providers. It merges encryption, decryption, and branch key provisioning into a
// single cohesive domain service following the Single Responsibility Principle at the domain
// service level.
//
// Encryption lifecycle:
//   - Create: validates entity, provisions the branch key,
//     encrypts Secret{plaintext} → Secret{ciphertext}, stores entity
//   - Get/List/Find: retrieves entity with Secret{ciphertext}, decrypts to Secret{plaintext}
//   - Update: validates entity (requires plaintext Secret), provisions the branch key
//     idempotently, encrypts Secret{plaintext} → Secret{ciphertext}, stores entity.
//     Encrypted state is rejected to ensure re-encryption always runs (e.g. during
//     key rotation or after switching encryption backends).
//
// The repository (ThirdpartyOAuth2ProviderRepository) is unaware of encryption mechanics
// and treats Secret ciphertext as opaque binary data.
type ThirdpartyOAuth2ProviderService struct {
	repo                ports.ThirdpartyOAuth2ProviderRepository
	encryption          ports.EncryptionPort
	branchKeyManager    ports.BranchKeyManager
	permissionSetRepo   ports.PermissionSetRepository // may be nil
	skipHTTPSValidation bool
	logger              *slog.Logger
}

// NewThirdpartyOAuth2ProviderService creates a new provider service.
// branchKeyManager must not be nil; inject the noop BranchKeyManager when no KMS store is configured.
// skipHTTPSValidation allows HTTP issuer/metadata URLs in development or test environments;
// set from config.Security.SkipThirdpartyHTTPSValidation.
func NewThirdpartyOAuth2ProviderService(
	repo ports.ThirdpartyOAuth2ProviderRepository,
	encryption ports.EncryptionPort,
	branchKeyManager ports.BranchKeyManager,
	permissionSetRepo ports.PermissionSetRepository,
	skipHTTPSValidation bool,
	logger *slog.Logger,
) *ThirdpartyOAuth2ProviderService {
	if logger == nil {
		logger = slog.Default()
	}
	if repo == nil {
		panic("thirdparty.NewThirdpartyOAuth2ProviderService: repo must not be nil")
	}
	if encryption == nil {
		panic("thirdparty.NewThirdpartyOAuth2ProviderService: encryption must not be nil")
	}
	if branchKeyManager == nil {
		panic("thirdparty.NewThirdpartyOAuth2ProviderService: branchKeyManager must not be nil")
	}
	return &ThirdpartyOAuth2ProviderService{
		repo:                repo,
		encryption:          encryption,
		branchKeyManager:    branchKeyManager,
		permissionSetRepo:   permissionSetRepo,
		skipHTTPSValidation: skipHTTPSValidation,
		logger:              logger,
	}
}

// Create validates, optionally provisions a branch key, encrypts the secret, and stores the entity.
// entity.Secret must be in plaintext state on entry.
// On success, entity.Secret is in encrypted state.
func (s *ThirdpartyOAuth2ProviderService) Create(
	ctx context.Context,
	entity *model.ThirdpartyOAuth2ProviderEntity,
) error {
	// Validate before any side effects: catches invalid entities before ID generation
	// and branch key provisioning (branch keys cannot be rolled back once provisioned).
	if err := entity.ValidateForCreate(s.skipHTTPSValidation); err != nil {
		return fmt.Errorf("provider validation failed: %w", err)
	}

	// Generate ID before encryption so context binding matches persisted ID
	if entity.ID.IsZero() {
		entity.ID = id.NewServiceID()
	}

	serviceSubject := domainencryption.NewServiceBranchKeySubject(entity.ID)

	// Provision branch key before creating service (fail-fast on error)
	s.logger.Info("provisioning branch key for service", "service_id", entity.ID)
	branchKeyID, err := s.branchKeyManager.Create(ctx, serviceSubject)
	if err != nil {
		s.logger.Error("failed to provision branch key",
			"service_id", entity.ID,
			"error", err)
		return fmt.Errorf("branch key provisioning failed: %w", err)
	}
	s.logger.Info("branch key provisioned",
		"service_id", entity.ID,
		"branch_key_id", branchKeyID)

	// Extract plaintext secret
	plaintext, err := entity.Secret.GetPlaintext()
	if err != nil {
		return fmt.Errorf("entity secret must be in plaintext state for create: %w", err)
	}

	// Encrypt secret
	ciphertext, err := s.encryption.Encrypt(ctx, []byte(plaintext), serviceSubject.EncryptionContext())
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
// Returns entity with Secret in plaintext state when decryption succeeds.
// If decryption fails (e.g. after switching encryption backends), the entity
// is returned with its Secret still in encrypted state and a nil error.
// This enables admin workflows (list, view, update) to continue operating
// even when the encryption backend has changed and old ciphertexts cannot
// be decrypted. Callers that require the plaintext secret (e.g. OAuth2
// session flows) should check entity.Secret.IsPlaintext() before use.
func (s *ThirdpartyOAuth2ProviderService) Get(
	ctx context.Context,
	serviceID id.ServiceID,
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

	dec, decErr := s.decryptSecret(ctx, entity)
	if decErr != nil {
		s.logger.Warn("secret_decryption_failed_returning_encrypted",
			"operation", "get",
			"service_id", entity.ID,
			"reason", decErr)
		return entity, nil
	}
	return dec, nil
}

// Update validates, encrypts the secret, and stores the updated entity.
// entity.Secret must be in plaintext state — the HTTP contract requires callers to
// always supply the secret. Passing encrypted state is rejected by ValidateForUpdate.
// On success, entity.Secret is in encrypted state.
//
// Update provisions the branch key before encrypting. This handles the migration case
// where a service was originally created with a different encryption backend (e.g. raw
// AES in-memory) that has no branch key entry in the current KMS key store.
// branchKeyManager.Create is idempotent: it is safe to call on services whose branch
// key already exists.
func (s *ThirdpartyOAuth2ProviderService) Update(
	ctx context.Context,
	entity *model.ThirdpartyOAuth2ProviderEntity,
) error {
	// Validate before any side effects to prevent wasted KMS calls on invalid input.
	if err := entity.ValidateForUpdate(s.skipHTTPSValidation); err != nil {
		return fmt.Errorf("provider validation failed: %w", err)
	}

	serviceSubject := domainencryption.NewServiceBranchKeySubject(entity.ID)

	// Provision branch key before encrypting (idempotent — safe for already-provisioned services).
	// Required when updating a service that was created with a different encryption backend and
	// therefore has no branch key in the current KMS key store.
	s.logger.Info("ensuring branch key exists for service update", "service_id", entity.ID)
	branchKeyID, err := s.branchKeyManager.Create(ctx, serviceSubject)
	if err != nil {
		s.logger.Error("failed to ensure branch key for update",
			"service_id", entity.ID,
			"error", err)
		return fmt.Errorf("branch key provisioning failed: %w", err)
	}
	s.logger.Info("branch key ready for update",
		"service_id", entity.ID,
		"branch_key_id", branchKeyID)

	plaintext, err := entity.Secret.GetPlaintext()
	if err != nil {
		return fmt.Errorf("failed to read plaintext secret for update: %w", err)
	}

	ciphertext, err := s.encryption.Encrypt(ctx, []byte(plaintext), serviceSubject.EncryptionContext())
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

	return s.repo.Update(ctx, entity)
}

// List retrieves all providers and decrypts their secrets.
// If decryption fails for individual entities (e.g. after switching encryption
// backends), those entities are returned with their Secret still in encrypted
// state. A warning is logged for each decryption failure. This enables admin
// workflows to continue operating even when old ciphertexts cannot be decrypted.
func (s *ThirdpartyOAuth2ProviderService) List(
	ctx context.Context,
) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	entities, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.ThirdpartyOAuth2ProviderEntity, 0, len(entities))
	for _, entity := range entities {
		dec, err := s.decryptSecret(ctx, entity)
		if err != nil {
			s.logger.Warn("secret_decryption_failed_returning_encrypted",
				"operation", "list",
				"service_id", entity.ID,
				"reason", err)
			result = append(result, entity)
			continue
		}
		result = append(result, dec)
	}

	return result, nil
}

// Delete removes a provider by ID.
// Returns a conflict error if permission sets reference this service.
func (s *ThirdpartyOAuth2ProviderService) Delete(
	ctx context.Context,
	serviceID id.ServiceID,
) error {
	// Check if any permission sets reference this service (deletion protection)
	if s.permissionSetRepo != nil {
		count, err := s.permissionSetRepo.CountPermissionSetsForService(ctx, serviceID)
		if err != nil {
			return fmt.Errorf("failed to check permission set references: %w", err)
		}

		if count > 0 {
			return storage.NewStorageError(
				"DeleteService",
				storage.ErrorKindConflict,
				nil,
				fmt.Sprintf("cannot delete service: %d permission set(s) reference it", count),
			)
		}
	}

	if err := s.repo.Delete(ctx, serviceID); err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	s.logger.Info("provider_deleted", "service_id", serviceID)
	return nil
}

// FindByProtectedResource retrieves a provider by resource URI and decrypts its secret.
// If decryption fails (e.g. after switching encryption backends), the entity is
// returned with its Secret still in encrypted state. This ensures that callers
// performing existence/ID checks (such as duplicate resource URI detection in
// PUT/POST handlers) continue to work even when old ciphertexts cannot be decrypted.
func (s *ThirdpartyOAuth2ProviderService) FindByProtectedResource(
	ctx context.Context,
	resourceURI string,
) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	entity, err := s.repo.FindByProtectedResource(ctx, resourceURI)
	if err != nil {
		return nil, err
	}

	dec, decErr := s.decryptSecret(ctx, entity)
	if decErr != nil {
		s.logger.Warn("secret_decryption_failed_returning_encrypted",
			"operation", "find_by_protected_resource",
			"service_id", entity.ID,
			"reason", decErr)
		return entity, nil
	}
	return dec, nil
}

// CountGrantsReferencingService returns the number of grants referencing this provider.
func (s *ThirdpartyOAuth2ProviderService) CountGrantsReferencingService(
	ctx context.Context,
	serviceID id.ServiceID,
) (int, error) {
	return s.repo.CountGrantsReferencingService(ctx, serviceID)
}

// ValidateServiceRequirements validates that all service_ids in the requirements exist and
// that all required_scopes are valid for the referenced service. This enforces referential
// integrity between agents and their declared service requirements.
//
// Uses the repository directly to avoid unnecessary secret decryption — only Scopes and
// DisplayName are needed for validation.
func (s *ThirdpartyOAuth2ProviderService) ValidateServiceRequirements(
	ctx context.Context,
	serviceReqs []storage.ServiceRequirement,
) error {
	if len(serviceReqs) == 0 {
		return nil
	}

	for i, sr := range serviceReqs {
		entity, err := s.repo.Get(ctx, sr.ServiceID)
		if err != nil {
			var storageErr *storage.StorageError
			if errors.As(err, &storageErr) && storageErr.Kind == storage.ErrorKindNotFound {
				s.logger.Warn("service not found for requirement",
					"service_id", sr.ServiceID, "index", i)
				return storage.NewStorageError(
					"ValidateServiceRequirements",
					storage.ErrorKindValidation,
					nil,
					"service_id "+sr.ServiceID.String()+" not found (index "+strconv.Itoa(i)+")",
				)
			}
			return err
		}

		serviceScopes := make(map[string]bool)
		for _, scope := range entity.Scopes {
			serviceScopes[scope.ScopeValue] = true
		}

		for _, requiredScope := range sr.RequiredScopes {
			if !serviceScopes[requiredScope] {
				s.logger.Warn("invalid scope for service",
					"service_id", sr.ServiceID,
					"service_name", entity.DisplayName,
					"scope", requiredScope,
					"index", i)
				return storage.NewStorageError(
					"ValidateServiceRequirements",
					storage.ErrorKindValidation,
					nil,
					"scope "+requiredScope+" not found in service "+entity.DisplayName+" (index "+strconv.Itoa(i)+")",
				)
			}
		}
	}

	return nil
}

// decryptSecret decrypts the entity's Secret field and returns a copy with the
// plaintext secret. It does not log on failure — callers are responsible for
// logging at the appropriate severity level for their use case.
func (s *ThirdpartyOAuth2ProviderService) decryptSecret(
	ctx context.Context,
	entity *model.ThirdpartyOAuth2ProviderEntity,
) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	ciphertext, err := entity.Secret.GetCiphertext()
	if err != nil {
		return nil, fmt.Errorf("entity has no encrypted secret: %w", err)
	}

	serviceSubject := domainencryption.NewServiceBranchKeySubject(entity.ID)

	plaintext, err := s.encryption.Decrypt(ctx, ciphertext, serviceSubject.EncryptionContext())
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
	}

	// Work on a copy to avoid mutating the entity returned by the repository
	result := entity.Copy()
	result.Secret = model.NewPlaintextSecret(string(plaintext))

	s.logger.Info("service_secret_decrypted",
		"service_id", entity.ID)

	return result, nil
}
