package thirdparty

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/google/uuid"
)

// ServiceManager handles third-party OAuth2 service management with transparent encryption.
// It follows Domain Service Encryption pattern where encryption logic resides in the
// domain service layer, and the repository operates on opaque encrypted bytes.
//
// Encryption lifecycle:
// - Create: encrypts ClientSecret → ClientSecretCiphertext before repository storage
// - Get: retrieves encrypted bytes from repository → decrypts to ClientSecret
// - Update: encrypts ClientSecret if changed → stores updated ciphertext
//
// The repository (ThirdpartyOAuth2ServiceRepository) is unaware of encryption mechanics
// and treats ClientSecretCiphertext as opaque binary data.
type ServiceManager struct {
	serviceRepo ports.ThirdpartyOAuth2ServiceRepository
	encryption  ports.EncryptionPort
	logger      *slog.Logger
}

// NewServiceManager creates a new third-party service manager.
func NewServiceManager(
	serviceRepo ports.ThirdpartyOAuth2ServiceRepository,
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

	// Set encrypted bytes and clear plaintext
	service.ClientSecretCiphertext = encryptedSecret
	service.ClientSecret = ""

	// Log successful encryption for audit trail
	sm.logger.Info(
		"service_secret_encrypted",
		"operation", "create",
		"service_id", service.ID,
	)

	// Store pre-encrypted service
	if err := sm.serviceRepo.Create(ctx, service); err != nil {
		return fmt.Errorf("failed to store service: %w", err)
	}

	sm.logger.Info(
		"service_created",
		"service_id", service.ID,
		"display_name", service.DisplayName,
	)

	return nil
}

// Get retrieves service and decrypts client secret.
func (sm *ServiceManager) Get(ctx context.Context, serviceID string) (*storage.ThirdpartyOAuth2Service, error) {
	// Repository returns service with encrypted bytes
	service, err := sm.serviceRepo.Get(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Guard: verify ID consistency for data integrity
	if service.ID != serviceID {
		sm.logger.Error(
			"service_id_mismatch",
			"expected_id", serviceID,
			"actual_id", service.ID,
		)
		return nil, fmt.Errorf("service ID mismatch: expected %s, got %s", serviceID, service.ID)
	}

	// Build encryption context with service_id only
	// Must match the context used during encryption
	encContext := map[string]string{
		"service_id": serviceID,
	}

	// Decrypt client secret
	decryptedSecret, err := sm.encryption.Decrypt(
		ctx,
		service.ClientSecretCiphertext,
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

	// Set decrypted plaintext
	service.ClientSecret = string(decryptedSecret)

	// Log successful decryption for audit trail
	sm.logger.Info(
		"service_secret_decrypted",
		"operation", "get",
		"service_id", serviceID,
	)

	return service, nil
}

// Update encrypts and stores updated service.
func (sm *ServiceManager) Update(
	ctx context.Context,
	service *storage.ThirdpartyOAuth2Service,
) error {
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

		service.ClientSecretCiphertext = encryptedSecret
		service.ClientSecret = ""

		// Log successful encryption for audit trail
		sm.logger.Info(
			"service_secret_encrypted",
			"operation", "update",
			"service_id", service.ID,
		)
	}

	return sm.serviceRepo.Update(ctx, service)
}

// List retrieves all services and decrypts their client secrets.
func (sm *ServiceManager) List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error) {
	services, err := sm.serviceRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Decrypt each service's client secret
	// Fail-fast on first decryption error for operational visibility
	for _, service := range services {
		// Build encryption context with service_id only
		encContext := map[string]string{
			"service_id": service.ID,
		}

		decryptedSecret, err := sm.encryption.Decrypt(
			ctx,
			service.ClientSecretCiphertext,
			encContext,
		)
		if err != nil {
			sm.logger.Error(
				"decryption_failed",
				"operation", "list_services",
				"service_id", service.ID,
				"reason", err,
			)
			// Fail-fast: return error immediately for operational visibility
			// This ensures decryption failures (security issues, corruption) are not silently ignored
			return nil, fmt.Errorf("failed to decrypt service %s: %w", service.ID, err)
		}

		service.ClientSecret = string(decryptedSecret)
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
	// Repository returns service with encrypted bytes
	service, err := sm.serviceRepo.FindByProtectedResource(ctx, resourceURI)
	if err != nil {
		return nil, err
	}

	// Build encryption context with service_id only
	// Must match the context used during encryption
	encContext := map[string]string{
		"service_id": service.ID,
	}

	// Decrypt client secret
	decryptedSecret, err := sm.encryption.Decrypt(
		ctx,
		service.ClientSecretCiphertext,
		encContext,
	)
	if err != nil {
		sm.logger.Error(
			"decryption_failed",
			"operation", "find_by_protected_resource",
			"service_id", service.ID,
			"resource_uri", resourceURI,
			"reason", err,
		)
		return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
	}

	// Set decrypted plaintext
	service.ClientSecret = string(decryptedSecret)

	// Log successful decryption for audit trail
	sm.logger.Info(
		"service_secret_decrypted",
		"operation", "find_by_protected_resource",
		"service_id", service.ID,
		"resource_uri", resourceURI,
	)

	return service, nil
}
