// Package admin provides HTTP handlers for administrative APIs.
package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/services"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ServicesHandler handles HTTP requests for third-party OAuth2 service CRUD operations.
type ServicesHandler struct {
	authProvider services.AuthProvider
	config       *ports.Config
	logger       *slog.Logger
}

// NewServicesHandler creates a new services handler.
func NewServicesHandler(authProvider services.AuthProvider, config *ports.Config, logger *slog.Logger) *ServicesHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &ServicesHandler{
		authProvider: authProvider,
		config:       config,
		logger:       logger,
	}
}

// ServiceRequest represents the request body for creating/updating a service.
type ServiceRequest struct {
	DisplayName        string                  `json:"display_name"`
	ClientID           string                  `json:"client_id"`
	ClientSecret       string                  `json:"client_secret"`
	IssuerURI          string                  `json:"issuer_uri"`
	Discovery          DiscoveryConfigRequest  `json:"discovery"`
	Endpoints          *OAuth2EndpointsRequest `json:"endpoints,omitempty"`
	Scopes             []OAuthScopeRequest     `json:"scopes"`
	ProtectedResources []string                `json:"protected_resources,omitempty"` // RFC 8693 resource URIs
}

// DiscoveryConfigRequest represents the discovery configuration in requests.
type DiscoveryConfigRequest struct {
	EnableDiscovery bool    `json:"enable_discovery"`
	MetadataURL     *string `json:"metadata_url,omitempty"`
}

// OAuth2EndpointsRequest represents OAuth2 endpoints in requests.
type OAuth2EndpointsRequest struct {
	TokenEndpoint     string `json:"token_endpoint"`
	AuthorizeEndpoint string `json:"authorize_endpoint"`
}

// OAuthScopeRequest represents a scope in requests.
type OAuthScopeRequest struct {
	ScopeValue  string `json:"scope_value"`
	Description string `json:"description"`
}

// ServiceResponse represents the response body for service operations.
// Client secret is always redacted in responses per SR-003.
type ServiceResponse struct {
	ID                 string                  `json:"id"`
	DisplayName        string                  `json:"display_name"`
	ClientID           string                  `json:"client_id"`
	ClientSecret       string                  `json:"client_secret"` // Always "REDACTED"
	IssuerURI          string                  `json:"issuer_uri"`
	Discovery          DiscoveryConfigResponse `json:"discovery"`
	Endpoints          OAuth2EndpointsResponse `json:"endpoints"`
	Scopes             []OAuthScopeResponse    `json:"scopes"`
	ProtectedResources []string                `json:"protected_resources,omitempty"` // RFC 8693 resource URIs
	CreatedAt          string                  `json:"created_at"`
	UpdatedAt          string                  `json:"updated_at"`
}

// DiscoveryConfigResponse represents the discovery configuration in responses.
type DiscoveryConfigResponse struct {
	EnableDiscovery bool    `json:"enable_discovery"`
	MetadataURL     *string `json:"metadata_url,omitempty"`
}

// OAuth2EndpointsResponse represents OAuth2 endpoints in responses.
type OAuth2EndpointsResponse struct {
	TokenEndpoint     string `json:"token_endpoint"`
	AuthorizeEndpoint string `json:"authorize_endpoint"`
}

// OAuthScopeResponse represents a scope in responses.
type OAuthScopeResponse struct {
	ScopeValue  string `json:"scope_value"`
	Description string `json:"description"`
}

// CreateService handles POST /api/third-party/oauth2/clients
func (h *ServicesHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Create service entity
	now := time.Now().UTC()
	service := &storage.ThirdpartyOAuth2Service{
		ID:           uuid.New().String(),
		DisplayName:  req.DisplayName,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		IssuerURI:    req.IssuerURI,
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: req.Discovery.EnableDiscovery,
			MetadataURL:     req.Discovery.MetadataURL,
		},
		Scopes:             make([]storage.OAuthScope, len(req.Scopes)),
		ProtectedResources: req.ProtectedResources,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Convert scopes
	for i, scope := range req.Scopes {
		service.Scopes[i] = storage.OAuthScope{
			ScopeValue:  scope.ScopeValue,
			Description: scope.Description,
		}
	}

	// If discovery is enabled, attempt to discover endpoints
	skipHTTPSValidation := false
	if h.config != nil {
		skipHTTPSValidation = h.config.Security.SkipThirdpartyHTTPSValidation
	}

	if req.Discovery.EnableDiscovery {
		endpoints, err := storage.DiscoverOAuth2Endpoints(ctx, req.IssuerURI, req.Discovery.MetadataURL, skipHTTPSValidation)
		if err != nil {
			h.logger.Warn("OAuth2 endpoint discovery failed, falling back to manual endpoints",
				"issuer_uri", req.IssuerURI,
				"error", err)

			// Fall back to manually provided endpoints if discovery fails
			if req.Endpoints != nil {
				service.Endpoints = storage.OAuth2Endpoints{
					TokenEndpoint:     req.Endpoints.TokenEndpoint,
					AuthorizeEndpoint: req.Endpoints.AuthorizeEndpoint,
				}
			}
		} else {
			// Use discovered endpoints
			service.Endpoints = *endpoints
		}
	} else {
		// Use manually provided endpoints
		if req.Endpoints != nil {
			service.Endpoints = storage.OAuth2Endpoints{
				TokenEndpoint:     req.Endpoints.TokenEndpoint,
				AuthorizeEndpoint: req.Endpoints.AuthorizeEndpoint,
			}
		}
	}

	// Validate service with configuration-based HTTPS validation skipping
	if err := service.ValidateForCreateWith(skipHTTPSValidation); err != nil {
		h.logger.Warn("service validation failed",
			"client_id", service.ClientID,
			"error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Validate protected_resources (RFC 8693 resource URIs) - return 400 for invalid format (T023)
	if err := service.ValidateProtectedResources(); err != nil {
		h.logger.Warn("protected_resources validation failed",
			"service_id", service.ID,
			"error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Check for duplicate resource URIs across existing services (T022)
	if len(service.ProtectedResources) > 0 {
		for _, resource := range service.ProtectedResources {
			existing, err := h.repo.FindByProtectedResource(ctx, resource)
			if err == nil && existing != nil {
				// Conflict: another service already has this resource URI
				h.logger.Warn("duplicate protected resource",
					"resource", resource,
					"service_id", existing.ID)
				h.writeError(w, http.StatusConflict, "conflict", "protected resource URI already configured for another service")
				return
			}
			// Ignore NotFound errors, those are expected when no existing service has the resource
		}
	}

	// Create service using domain service (handles branch key provisioning and creation atomically)
	createdService, err := h.authProvider.Create(ctx, service)
	if err != nil {
		h.handleStorageError(w, r, "CreateService", err)
		return
	}

	// Return created service with redacted secret
	resp := h.toResponse(createdService.RedactedCopy())
	h.writeJSON(w, http.StatusCreated, resp)
}

// GetService handles GET /api/third-party/oauth2/clients/:client-id
func (h *ServicesHandler) GetService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serviceID := chi.URLParam(r, "service-id")

	if serviceID == "" {
		h.writeError(w, http.StatusBadRequest, "service ID is required", "")
		return
	}

	service, err := h.authProvider.Get(ctx, serviceID)
	if err != nil {
		h.handleStorageError(w, r, "GetService", err)
		return
	}

	// Return service with redacted secret
	resp := h.toResponse(service.RedactedCopy())
	h.writeJSON(w, http.StatusOK, resp)
}

// UpdateService handles PUT /api/services/:service-id
func (h *ServicesHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serviceID := chi.URLParam(r, "service-id")

	if serviceID == "" {
		h.writeError(w, http.StatusBadRequest, "service ID is required", "")
		return
	}

	var req ServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Get existing service to preserve created_at
	existing, err := h.authProvider.Get(ctx, serviceID)
	if err != nil {
		h.handleStorageError(w, r, "UpdateService", err)
		return
	}

	// Update service entity
	service := &storage.ThirdpartyOAuth2Service{
		ID:           serviceID,
		DisplayName:  req.DisplayName,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		IssuerURI:    req.IssuerURI,
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: req.Discovery.EnableDiscovery,
			MetadataURL:     req.Discovery.MetadataURL,
		},
		Scopes:             make([]storage.OAuthScope, len(req.Scopes)),
		ProtectedResources: req.ProtectedResources,
		CreatedAt:          existing.CreatedAt,
		UpdatedAt:          time.Now().UTC(),
	}

	// Convert scopes
	for i, scope := range req.Scopes {
		service.Scopes[i] = storage.OAuthScope{
			ScopeValue:  scope.ScopeValue,
			Description: scope.Description,
		}
	}

	// If discovery is enabled, attempt to discover endpoints
	skipHTTPSValidation := false
	if h.config != nil {
		skipHTTPSValidation = h.config.Security.SkipThirdpartyHTTPSValidation
	}

	if req.Discovery.EnableDiscovery {
		endpoints, err := storage.DiscoverOAuth2Endpoints(ctx, req.IssuerURI, req.Discovery.MetadataURL, skipHTTPSValidation)
		if err != nil {
			h.logger.Warn("OAuth2 endpoint discovery failed, falling back to manual endpoints",
				"issuer_uri", req.IssuerURI,
				"error", err)

			// Fall back to manually provided endpoints if discovery fails
			if req.Endpoints != nil {
				service.Endpoints = storage.OAuth2Endpoints{
					TokenEndpoint:     req.Endpoints.TokenEndpoint,
					AuthorizeEndpoint: req.Endpoints.AuthorizeEndpoint,
				}
			}
		} else {
			// Use discovered endpoints
			service.Endpoints = *endpoints
		}
	} else {
		// Use manually provided endpoints
		if req.Endpoints != nil {
			service.Endpoints = storage.OAuth2Endpoints{
				TokenEndpoint:     req.Endpoints.TokenEndpoint,
				AuthorizeEndpoint: req.Endpoints.AuthorizeEndpoint,
			}
		}
	}

	// Validate service with configuration-based HTTPS validation skipping
	if err := service.ValidateWith(skipHTTPSValidation); err != nil {
		h.logger.Warn("service validation failed",
			"client_id", service.ClientID,
			"error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Validate protected_resources (RFC 8693 resource URIs) - return 400 for invalid format (T023)
	if err := service.ValidateProtectedResources(); err != nil {
		h.logger.Warn("protected_resources validation failed",
			"service_id", service.ID,
			"error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Check for duplicate resource URIs across other services (T022)
	// Only check resources that are different from existing service's resources
	if len(service.ProtectedResources) > 0 {
		for _, resource := range service.ProtectedResources {
			existing, err := h.repo.FindByProtectedResource(ctx, resource)
			if err == nil && existing != nil && existing.ID != service.ID {
				// Conflict: another service already has this resource URI
				h.logger.Warn("duplicate protected resource",
					"resource", resource,
					"service_id", existing.ID)
				h.writeError(w, http.StatusConflict, "conflict", "protected resource URI already configured for another service")
				return
			}
			// Ignore NotFound errors, those are expected when no existing service has the resource
		}
	}

	// Update in repository (skip validation since we already did it above)
	if err := h.authProvider.Update(ctx, service); err != nil {
		h.handleStorageError(w, r, "UpdateService", err)
		return
	}

	h.logger.Info("OAuth2 service updated",
		"service_id", service.ID,
		"client_id", service.ClientID)

	// Return updated service with redacted secret
	resp := h.toResponse(service.RedactedCopy())
	h.writeJSON(w, http.StatusOK, resp)
}

// DeleteService handles DELETE /api/services/:service-id
func (h *ServicesHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serviceID := chi.URLParam(r, "service-id")

	if serviceID == "" {
		h.writeError(w, http.StatusBadRequest, "service ID is required", "")
		return
	}

	// Delete via domain service
	if err := h.authProvider.Delete(ctx, serviceID); err != nil {
		// Special handling for conflict errors (grants exist)
		if storageErr, ok := err.(*storage.StorageError); ok && storageErr.Kind == storage.ErrorKindConflict {
			// Extract grant count from error message if possible
			message := storageErr.Message
			h.logger.Warn("service deletion blocked due to existing grants",
				"service_id", serviceID,
				"error", message)
			h.writeError(w, http.StatusConflict, "conflict", message)
			return
		}

		h.handleStorageError(w, r, "DeleteService", err)
		return
	}

	h.logger.Info("OAuth2 service deleted", "service_id", serviceID)

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// ListServices handles GET /api/third-party/oauth2/clients
func (h *ServicesHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	services, err := h.authProvider.List(ctx)
	if err != nil {
		h.handleStorageError(w, r, "ListServices", err)
		return
	}

	// Convert to response format with redacted secrets
	responses := make([]ServiceResponse, len(services))
	for i, service := range services {
		responses[i] = h.toResponse(service.RedactedCopy())
	}

	h.writeJSON(w, http.StatusOK, responses)
}

// toResponse converts a ThirdpartyOAuth2Service entity to ServiceResponse.
// Assumes the service has already been redacted.
func (h *ServicesHandler) toResponse(service *storage.ThirdpartyOAuth2Service) ServiceResponse {
	scopes := make([]OAuthScopeResponse, len(service.Scopes))
	for i, scope := range service.Scopes {
		scopes[i] = OAuthScopeResponse{
			ScopeValue:  scope.ScopeValue,
			Description: scope.Description,
		}
	}

	return ServiceResponse{
		ID:           service.ID,
		DisplayName:  service.DisplayName,
		ClientID:     service.ClientID,
		ClientSecret: service.ClientSecret, // Should be "REDACTED"
		IssuerURI:    service.IssuerURI,
		Discovery: DiscoveryConfigResponse{
			EnableDiscovery: service.Discovery.EnableDiscovery,
			MetadataURL:     service.Discovery.MetadataURL,
		},
		Endpoints: OAuth2EndpointsResponse{
			TokenEndpoint:     service.Endpoints.TokenEndpoint,
			AuthorizeEndpoint: service.Endpoints.AuthorizeEndpoint,
		},
		Scopes:             scopes,
		ProtectedResources: service.ProtectedResources,
		CreatedAt:          service.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          service.UpdatedAt.Format(time.RFC3339),
	}
}

// handleStorageError converts storage errors to HTTP responses.
func (h *ServicesHandler) handleStorageError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	storageErr, ok := err.(*storage.StorageError)
	if !ok {
		h.logger.Error("unexpected error type", "operation", operation, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	switch storageErr.Kind {
	case storage.ErrorKindNotFound:
		h.writeError(w, http.StatusNotFound, "service not found", storageErr.Message)
	case storage.ErrorKindConflict:
		// Extract error details for user-friendly message
		message := storageErr.Message
		if strings.Contains(message, "grants reference it") {
			h.writeError(w, http.StatusConflict, "conflict", message)
		} else {
			h.writeError(w, http.StatusConflict, "conflict", storageErr.Message)
		}
	case storage.ErrorKindValidation:
		h.writeError(w, http.StatusBadRequest, "validation failed", storageErr.Message)
	case storage.ErrorKindTimeout:
		h.writeError(w, http.StatusGatewayTimeout, "operation timed out", storageErr.Message)
	default:
		h.logger.Error("storage operation failed", "operation", operation, "kind", storageErr.Kind, "error", storageErr)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
	}
}

// writeJSON writes a JSON response.
func (h *ServicesHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response.
func (h *ServicesHandler) writeError(w http.ResponseWriter, statusCode int, error string, message string) {
	resp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}
