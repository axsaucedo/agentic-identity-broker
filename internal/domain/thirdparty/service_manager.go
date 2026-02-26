package thirdparty

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/google/uuid"
)

// ServiceManager handles third-party OAuth2 service management with transparent encryption.
// It follows Domain Service Encryption pattern where encryption logic resides in the
// domain service layer, and the repository operates on opaque encrypted bytes.
//
// Encryption lifecycle:
// - Create: encrypts ClientSecret → Secret (encrypted state) before repository storage
// - Get: retrieves encrypted Secret from repository → decrypts to ClientSecret
// - Update: encrypts ClientSecret if changed → stores updated encrypted Secret
//
// The repository (ThirdpartyOAuth2ProviderRepository) is unaware of encryption mechanics
// and treats Secret ciphertext as opaque binary data.
type ServiceManager struct {
	serviceRepo ports.ThirdpartyOAuth2ProviderRepository
	encryption  ports.EncryptionPort
	logger      *slog.Logger
}

// NewServiceManager creates a new third-party service manager.
func NewServiceManager(
	serviceRepo ports.ThirdpartyOAuth2ProviderRepository,
	encryption ports.EncryptionPort,
	logger *slog.Logger,
) *ServiceManager {
	return &ServiceManager{
		serviceRepo: serviceRepo,
		encryption:  encryption,
		logger:      logger,
	}
}

// Create encrypts client secret and stores the service.
func (sm *ServiceManager) Create(
	ctx context.Context,
	service *storage.ThirdpartyOAuth2Service,
) error {
	if service.ClientSecret == "" {
		return fmt.Errorf("client secret required")
	}

	// Generate service ID if not provided
	// Must happen before encryption so the encryption context matches the persisted ID
	if service.ID == "" {
		service.ID = uuid.New().String()
	}

	// Build encryption context with service_id only
	encContext := map[string]string{
		"service_id": service.ID,
	}

	// Encrypt client secret
	encryptedSecret, err := sm.encryption.Encrypt(
		ctx,
		[]byte(service.ClientSecret),
		encContext,
	)
	if err != nil {
		sm.logger.Error(
			"encryption_failed",
			"operation", "create_service",
			"service_id", service.ID,
			"reason", err,
		)
		return fmt.Errorf("failed to encrypt client secret: %w", err)
	}

	// Convert to entity with encrypted Secret
	entity := storageServiceToEntity(service)
	entity.Secret = model.NewEncryptedSecret(encryptedSecret)

	// Log successful encryption for audit trail
	sm.logger.Info(
		"service_secret_encrypted",
		"operation", "create",
		"service_id", service.ID,
	)

	// Store pre-encrypted entity
	if err := sm.serviceRepo.Create(ctx, entity); err != nil {
		return fmt.Errorf("failed to store service: %w", err)
	}

	// Clear plaintext secret now that it is stored
	service.ClientSecret = ""
	service.ClientSecretCiphertext = encryptedSecret

	sm.logger.Info(
		"service_created",
		"service_id", service.ID,
		"display_name", service.DisplayName,
	)

	return nil
}

// Get retrieves service and decrypts client secret.
func (sm *ServiceManager) Get(ctx context.Context, serviceID string) (*storage.ThirdpartyOAuth2Service, error) {
	// Repository returns entity with encrypted Secret
	entity, err := sm.serviceRepo.Get(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Guard: verify ID consistency for data integrity
	if entity.ID != serviceID {
		sm.logger.Error(
			"service_id_mismatch",
			"expected_id", serviceID,
			"actual_id", entity.ID,
		)
		return nil, fmt.Errorf("service ID mismatch: expected %s, got %s", serviceID, entity.ID)
	}

	// Build encryption context with service_id only
	// Must match the context used during encryption
	encContext := map[string]string{
		"service_id": serviceID,
	}

	// Extract ciphertext from entity Secret
	ciphertext, err := entity.Secret.GetCiphertext()
	if err != nil {
		return nil, fmt.Errorf("entity has no encrypted secret: %w", err)
	}

	// Decrypt client secret
	decryptedSecret, err := sm.encryption.Decrypt(
		ctx,
		ciphertext,
		encContext,
	)
	if err != nil {
		sm.logger.Error(
			"decryption_failed",
			"operation", "get_service",
			"service_id", serviceID,
			"reason", err,
		)
		return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
	}

	// Convert to storage DTO and set decrypted plaintext
	svc := entityToStorageService(entity)
	svc.ClientSecret = string(decryptedSecret)

	// Log successful decryption for audit trail
	sm.logger.Info(
		"service_secret_decrypted",
		"operation", "get",
		"service_id", serviceID,
	)

	return svc, nil
}

// Update encrypts and stores updated service.
func (sm *ServiceManager) Update(
	ctx context.Context,
	service *storage.ThirdpartyOAuth2Service,
) error {
	entity := storageServiceToEntity(service)

	// Encrypt if secret changed
	if service.ClientSecret != "" {
		// Build encryption context with service_id only
		encContext := map[string]string{
			"service_id": service.ID,
		}

		encryptedSecret, err := sm.encryption.Encrypt(
			ctx,
			[]byte(service.ClientSecret),
			encContext,
		)
		if err != nil {
			return fmt.Errorf("failed to encrypt client secret: %w", err)
		}

		entity.Secret = model.NewEncryptedSecret(encryptedSecret)

		// Log successful encryption for audit trail
		sm.logger.Info(
			"service_secret_encrypted",
			"operation", "update",
			"service_id", service.ID,
		)
	}

	return sm.serviceRepo.Update(ctx, entity)
}

// List retrieves all services and decrypts their client secrets.
func (sm *ServiceManager) List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error) {
	entities, err := sm.serviceRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Decrypt each service's client secret
	// Fail-fast on first decryption error for operational visibility
	services := make([]*storage.ThirdpartyOAuth2Service, 0, len(entities))
	for _, entity := range entities {
		// Build encryption context with service_id only
		encContext := map[string]string{
			"service_id": entity.ID,
		}

		ciphertext, err := entity.Secret.GetCiphertext()
		if err != nil {
			sm.logger.Error(
				"decryption_failed",
				"operation", "list_services",
				"service_id", entity.ID,
				"reason", "entity has no encrypted secret",
			)
			return nil, fmt.Errorf("entity %s has no encrypted secret: %w", entity.ID, err)
		}

		decryptedSecret, err := sm.encryption.Decrypt(
			ctx,
			ciphertext,
			encContext,
		)
		if err != nil {
			sm.logger.Error(
				"decryption_failed",
				"operation", "list_services",
				"service_id", entity.ID,
				"reason", err,
			)
			// Fail-fast: return error immediately for operational visibility
			// This ensures decryption failures (security issues, corruption) are not silently ignored
			return nil, fmt.Errorf("failed to decrypt service %s: %w", entity.ID, err)
		}

		svc := entityToStorageService(entity)
		svc.ClientSecret = string(decryptedSecret)
		services = append(services, svc)
	}

	return services, nil
}

// Delete removes a service from storage.
func (sm *ServiceManager) Delete(ctx context.Context, serviceID string) error {
	if err := sm.serviceRepo.Delete(ctx, serviceID); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	sm.logger.Info(
		"service_deleted",
		"service_id", serviceID,
	)

	return nil
}

// FindByProtectedResource retrieves service by protected resource URI and decrypts client secret.
func (sm *ServiceManager) FindByProtectedResource(ctx context.Context, resourceURI string) (*storage.ThirdpartyOAuth2Service, error) {
	// Repository returns entity with encrypted Secret
	entity, err := sm.serviceRepo.FindByProtectedResource(ctx, resourceURI)
	if err != nil {
		return nil, err
	}

	// Build encryption context with service_id only
	// Must match the context used during encryption
	encContext := map[string]string{
		"service_id": entity.ID,
	}

	// Extract ciphertext from entity Secret
	ciphertext, err := entity.Secret.GetCiphertext()
	if err != nil {
		return nil, fmt.Errorf("entity has no encrypted secret: %w", err)
	}

	// Decrypt client secret
	decryptedSecret, err := sm.encryption.Decrypt(
		ctx,
		ciphertext,
		encContext,
	)
	if err != nil {
		sm.logger.Error(
			"decryption_failed",
			"operation", "find_by_protected_resource",
			"service_id", entity.ID,
			"resource_uri", resourceURI,
			"reason", err,
		)
		return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
	}

	// Convert to storage DTO and set decrypted plaintext
	svc := entityToStorageService(entity)
	svc.ClientSecret = string(decryptedSecret)

	// Log successful decryption for audit trail
	sm.logger.Info(
		"service_secret_decrypted",
		"operation", "find_by_protected_resource",
		"service_id", entity.ID,
		"resource_uri", resourceURI,
	)

	return svc, nil
}

// storageServiceToEntity converts a storage DTO to a domain entity.
// Used when passing to the repository for Create/Update operations.
// The Secret state is set by the caller after encryption.
func storageServiceToEntity(svc *storage.ThirdpartyOAuth2Service) *model.ThirdpartyOAuth2ProviderEntity {
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          svc.ID,
		DisplayName: svc.DisplayName,
		ClientID:    svc.ClientID,
		IssuerURI:   svc.IssuerURI,
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: svc.Discovery.EnableDiscovery,
			MetadataURL:     svc.Discovery.MetadataURL,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     svc.Endpoints.TokenEndpoint,
			AuthorizeEndpoint: svc.Endpoints.AuthorizeEndpoint,
		},
		Scopes:    storageScopes(svc.Scopes),
		CreatedAt: svc.CreatedAt,
		UpdatedAt: svc.UpdatedAt,
	}

	if len(svc.ProtectedResources) > 0 {
		entity.ProtectedResources = make([]string, len(svc.ProtectedResources))
		copy(entity.ProtectedResources, svc.ProtectedResources)
	}

	// Set Secret state based on what the storage DTO carries
	if svc.ClientSecret != "" {
		entity.Secret = model.NewPlaintextSecret(svc.ClientSecret)
	} else if len(svc.ClientSecretCiphertext) > 0 {
		entity.Secret = model.NewEncryptedSecret(svc.ClientSecretCiphertext)
	}

	return entity
}

// entityToStorageService converts a domain entity back to a storage DTO.
// Used when returning data to callers that work with the old storage type.
func entityToStorageService(entity *model.ThirdpartyOAuth2ProviderEntity) *storage.ThirdpartyOAuth2Service {
	svc := &storage.ThirdpartyOAuth2Service{
		ID:          entity.ID,
		DisplayName: entity.DisplayName,
		ClientID:    entity.ClientID,
		IssuerURI:   entity.IssuerURI,
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: entity.Discovery.EnableDiscovery,
			MetadataURL:     entity.Discovery.MetadataURL,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     entity.Endpoints.TokenEndpoint,
			AuthorizeEndpoint: entity.Endpoints.AuthorizeEndpoint,
		},
		Scopes:    modelScopes(entity.Scopes),
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}

	if len(entity.ProtectedResources) > 0 {
		svc.ProtectedResources = make([]string, len(entity.ProtectedResources))
		copy(svc.ProtectedResources, entity.ProtectedResources)
	}

	if entity.Secret.IsEncrypted() {
		ct, _ := entity.Secret.GetCiphertext()
		svc.ClientSecretCiphertext = ct
	}

	return svc
}

// storageScopes converts []storage.OAuthScope to []model.OAuthScope.
func storageScopes(in []storage.OAuthScope) []model.OAuthScope {
	out := make([]model.OAuthScope, len(in))
	for i, s := range in {
		out[i] = model.OAuthScope{ScopeValue: s.ScopeValue, Description: s.Description}
	}
	return out
}

// modelScopes converts []model.OAuthScope to []storage.OAuthScope.
func modelScopes(in []model.OAuthScope) []storage.OAuthScope {
	out := make([]storage.OAuthScope, len(in))
	for i, s := range in {
		out[i] = storage.OAuthScope{ScopeValue: s.ScopeValue, Description: s.Description}
	}
	return out
}
