// Package services provides domain services that orchestrate between different ports.
package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// AuthProvider defines the interface for OAuth2 service management.
type AuthProvider interface {
	Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) (*storage.ThirdpartyOAuth2Service, error)
	Get(ctx context.Context, clientID string) (*storage.ThirdpartyOAuth2Service, error)
	Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error
	Delete(ctx context.Context, clientID string) error
	List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error)
	FindByProtectedResource(ctx context.Context, resourceURI string) (*storage.ThirdpartyOAuth2Service, error)
}

// ThirdpartyOAuth2ServiceProvider orchestrates third-party OAuth2 service management with branch key provisioning.
// This domain service encapsulates the business logic: "when creating an OAuth2 service, provision its branch key".
// It follows Domain Service Encryption pattern by delegating encryption to ThirdpartyServiceManager
// and focusing on branch key orchestration.
type ThirdpartyOAuth2ServiceProvider struct {
	serviceManager   *thirdparty.ServiceManager
	branchKeyManager ports.BranchKeyManager
	logger           *slog.Logger
}

// NewAuthProvider creates a new ThirdpartyOAuth2ServiceProvider domain service.
// branchKeyManager may be nil if no encryption backend is configured.
// serviceManager handles encryption/decryption following Pattern B.
func NewAuthProvider(
	serviceManager *thirdparty.ServiceManager,
	branchKeyManager ports.BranchKeyManager,
	logger *slog.Logger,
) *ThirdpartyOAuth2ServiceProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &ThirdpartyOAuth2ServiceProvider{
		serviceManager:   serviceManager,
		branchKeyManager: branchKeyManager,
		logger:           logger,
	}
}

// Create creates a new OAuth2 service and provisions its branch key atomically.
// This orchestrates the business logic: provision branch key before creating service (fail-fast on error).
// Delegates to ThirdpartyServiceManager for encryption.
// Returns the created service or error if branch key provisioning or service creation fails.
func (ap *ThirdpartyOAuth2ServiceProvider) Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) (*storage.ThirdpartyOAuth2Service, error) {
	// Provision branch key before creating service (fail-fast for encryption setup)
	if ap.branchKeyManager != nil {
		ap.logger.Info("provisioning branch key for service", "service_id", service.ID)
		branchKeyID, err := ap.branchKeyManager.Create(ctx, service.ID)
		if err != nil {
			ap.logger.Error("failed to provision branch key", "service_id", service.ID, "error", err)
			return nil, fmt.Errorf("branch key provisioning failed: %w", err)
		}
		ap.logger.Info("branch key provisioned", "service_id", service.ID, "branch_key_id", branchKeyID)
	}

	// Create service using ServiceManager (handles encryption transparently)
	if err := ap.serviceManager.Create(ctx, service); err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	ap.logger.Info("OAuth2 service created",
		"service_id", service.ID,
		"client_id", service.ClientID,
		"issuer_uri", service.IssuerURI)

	return service, nil
}

// Get retrieves a service by ID, delegating to ServiceManager (handles decryption).
func (ap *ThirdpartyOAuth2ServiceProvider) Get(ctx context.Context, clientID string) (*storage.ThirdpartyOAuth2Service, error) {
	return ap.serviceManager.Get(ctx, clientID)
}

// Update updates an existing service, delegating to ServiceManager (handles encryption).
func (ap *ThirdpartyOAuth2ServiceProvider) Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	return ap.serviceManager.Update(ctx, service)
}

// Delete deletes a service by ID, delegating to ServiceManager.
func (ap *ThirdpartyOAuth2ServiceProvider) Delete(ctx context.Context, clientID string) error {
	return ap.serviceManager.Delete(ctx, clientID)
}

// List retrieves all services, delegating to ServiceManager (handles decryption).
func (ap *ThirdpartyOAuth2ServiceProvider) List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error) {
	return ap.serviceManager.List(ctx)
}

// FindByProtectedResource finds a service by protected resource URI.
// Delegates to ServiceManager which handles decryption transparently.
func (ap *ThirdpartyOAuth2ServiceProvider) FindByProtectedResource(ctx context.Context, resourceURI string) (*storage.ThirdpartyOAuth2Service, error) {
	return ap.serviceManager.FindByProtectedResource(ctx, resourceURI)
}
