// Package services provides domain services that orchestrate between different ports.
package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// AuthProvider orchestrates third-party OAuth2 service management with branch key provisioning.
// This domain service encapsulates the business logic: "when creating an OAuth2 service, provision its branch key".
// It follows the established pattern from ConsentService, maintaining clean hexagonal boundaries.
type AuthProvider struct {
	serviceRepository ports.ThirdpartyOAuth2ServiceRepository
	branchKeyManager  ports.BranchKeyManager
	logger            *slog.Logger
}

// NewAuthProvider creates a new AuthProvider domain service.
// branchKeyManager may be nil if no encryption backend is configured.
func NewAuthProvider(
	serviceRepository ports.ThirdpartyOAuth2ServiceRepository,
	branchKeyManager ports.BranchKeyManager,
	logger *slog.Logger,
) *AuthProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthProvider{
		serviceRepository: serviceRepository,
		branchKeyManager:  branchKeyManager,
		logger:            logger,
	}
}

// Create creates a new OAuth2 service and provisions its branch key atomically.
// This orchestrates the business logic: provision branch key before creating service (fail-fast on error).
// Returns the created service or error if branch key provisioning or service creation fails.
func (ap *AuthProvider) Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) (*storage.ThirdpartyOAuth2Service, error) {
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

	// Create service in repository
	if err := ap.serviceRepository.Create(ctx, service); err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	ap.logger.Info("OAuth2 service created",
		"service_id", service.ID,
		"client_id", service.ClientID,
		"issuer_uri", service.IssuerURI)

	return service, nil
}

// Get retrieves a service by ID, delegating to the repository.
func (ap *AuthProvider) Get(ctx context.Context, clientID string) (*storage.ThirdpartyOAuth2Service, error) {
	return ap.serviceRepository.Get(ctx, clientID)
}

// Update updates an existing service, delegating to the repository.
func (ap *AuthProvider) Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	return ap.serviceRepository.Update(ctx, service)
}

// Delete deletes a service by ID, delegating to the repository.
func (ap *AuthProvider) Delete(ctx context.Context, clientID string) error {
	return ap.serviceRepository.Delete(ctx, clientID)
}

// List retrieves all services, delegating to the repository.
func (ap *AuthProvider) List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error) {
	return ap.serviceRepository.List(ctx)
}
