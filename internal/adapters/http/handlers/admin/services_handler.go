// Package admin provides HTTP handlers for administrative APIs.
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// ServicesHandler handles HTTP requests for third-party OAuth2 service CRUD operations.
type ServicesHandler struct {
	providerService *thirdparty.ThirdpartyOAuth2ProviderService
	config          *ports.Config
	logger          *slog.Logger
}

// NewServicesHandler creates a new services handler.
func NewServicesHandler(providerService *thirdparty.ThirdpartyOAuth2ProviderService, config *ports.Config, logger *slog.Logger) *ServicesHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &ServicesHandler{
		providerService: providerService,
		config:          config,
		logger:          logger,
	}
}

// ServiceRequest represents the request body for creating/updating a service.
type ServiceRequest struct {
	DisplayName         string                  `json:"display_name"`
	ClientID            string                  `json:"client_id"`
	ClientSecret        string                  `json:"client_secret"`
	OAuth2Flavor        string                  `json:"oauth2_flavor,omitempty"`
	IssuerURI           string                  `json:"issuer_uri"`
	Discovery           DiscoveryConfigRequest  `json:"discovery"`
	Endpoints           *OAuth2EndpointsRequest `json:"endpoints,omitempty"`
	Scopes              []OAuthScopeRequest     `json:"scopes"`
	ProtectedResources  []string                `json:"protected_resources,omitempty"` // RFC 8693 resource URIs
	AuthorizationParams map[string]string       `json:"authorization_params,omitempty"`
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
	ID                  string                  `json:"id"`
	DisplayName         string                  `json:"display_name"`
	ClientID            string                  `json:"client_id"`
	ClientSecret        string                  `json:"client_secret"` // Always "REDACTED"
	OAuth2Flavor        string                  `json:"oauth2_flavor"`
	IssuerURI           string                  `json:"issuer_uri"`
	Discovery           DiscoveryConfigResponse `json:"discovery"`
	Endpoints           OAuth2EndpointsResponse `json:"endpoints"`
	Scopes              []OAuthScopeResponse    `json:"scopes"`
	ProtectedResources  []string                `json:"protected_resources,omitempty"` // RFC 8693 resource URIs
	AuthorizationParams map[string]string       `json:"authorization_params,omitempty"`
	CreatedAt           string                  `json:"created_at"`
	UpdatedAt           string                  `json:"updated_at"`
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

	skipHTTPSValidation := false
	if h.config != nil {
		skipHTTPSValidation = h.config.Security.SkipThirdpartyHTTPSValidation
	}

	// Parse and validate oauth2_flavor; default to "standard" when omitted,
	// or auto-detect "github" from token endpoint URL.
	flavor := model.OAuth2Flavor(req.OAuth2Flavor)
	if flavor == "" {
		var tokenEndpoint string
		if req.Endpoints != nil {
			tokenEndpoint = req.Endpoints.TokenEndpoint
		}
		flavor = model.InferOAuth2Flavor(tokenEndpoint)
	}

	// Build entity from request
	now := time.Now().UTC()
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:                  id.NewServiceID(),
		DisplayName:         req.DisplayName,
		ClientID:            id.ClientID(req.ClientID),
		Secret:              model.NewPlaintextSecret(req.ClientSecret),
		Flavor:              flavor,
		IssuerURI:           req.IssuerURI,
		Discovery:           model.DiscoveryConfig{EnableDiscovery: req.Discovery.EnableDiscovery, MetadataURL: req.Discovery.MetadataURL},
		ProtectedResources:  req.ProtectedResources,
		AuthorizationParams: req.AuthorizationParams,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// Convert scopes
	entity.Scopes = make([]model.OAuthScope, len(req.Scopes))
	for i, scope := range req.Scopes {
		entity.Scopes[i] = model.OAuthScope{ScopeValue: scope.ScopeValue, Description: scope.Description}
	}

	// For Google flavor, ValidateForCreate will call enrichForGoogleFlavor internally
	// (parsing the credential once and populating ClientID, Endpoints, IssuerURI).
	// For standard flavor, set endpoints from discovery or the request.
	if flavor != model.OAuth2FlavorGoogle {
		// If discovery is enabled, attempt to discover endpoints
		if req.Discovery.EnableDiscovery {
			endpoints, err := storage.DiscoverOAuth2Endpoints(ctx, req.IssuerURI, req.Discovery.MetadataURL, skipHTTPSValidation)
			if err != nil {
				h.logger.Warn("OAuth2 endpoint discovery failed, falling back to manual endpoints",
					"issuer_uri", req.IssuerURI,
					"error", err)
				// Fall back to manually provided endpoints
				if req.Endpoints != nil {
					entity.Endpoints = model.OAuth2Endpoints{
						TokenEndpoint:     req.Endpoints.TokenEndpoint,
						AuthorizeEndpoint: req.Endpoints.AuthorizeEndpoint,
					}
				}
			} else {
				entity.Endpoints = model.OAuth2Endpoints{
					TokenEndpoint:     endpoints.TokenEndpoint,
					AuthorizeEndpoint: endpoints.AuthorizeEndpoint,
				}
			}
		} else if req.Endpoints != nil {
			entity.Endpoints = model.OAuth2Endpoints{
				TokenEndpoint:     req.Endpoints.TokenEndpoint,
				AuthorizeEndpoint: req.Endpoints.AuthorizeEndpoint,
			}
		}
	}

	// Validate entity (for Google flavor, this also enriches ClientID, Endpoints, IssuerURI).
	if err := entity.ValidateForCreate(skipHTTPSValidation); err != nil {
		h.logger.Warn("service validation failed",
			"client_id", entity.ClientID,
			"error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Defense-in-depth: normalize and validate protected_resources here so the duplicate
	// check below operates on canonical URIs. The domain service repeats this
	// authoritatively before persistence.
	entity.NormalizeProtectedResources()
	if err := entity.ValidateProtectedResources(); err != nil {
		h.logger.Warn("protected_resources validation failed",
			"service_id", entity.ID,
			"error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	if !h.checkProtectedResourceConflicts(ctx, w, r, entity, id.ServiceID{}, "CreateService") {
		return
	}

	// Create service (branch key provisioning and encryption handled by domain service)
	if err := h.providerService.Create(ctx, entity); err != nil {
		h.handleStorageError(w, r, "CreateService", err)
		return
	}

	// Return created service with redacted secret
	resp := h.toResponse(entity.RedactedCopy())
	h.writeJSON(w, http.StatusCreated, resp)
}

// GetService handles GET /api/services/{service-id}
func (h *ServicesHandler) GetService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serviceID := chi.URLParam(r, "service-id")

	if serviceID == "" {
		h.writeError(w, http.StatusBadRequest, "service ID is required", "")
		return
	}

	parsedSvcID, parseErr := id.ParseServiceID(serviceID)
	if parseErr != nil {
		h.writeError(w, http.StatusBadRequest, "invalid service ID", parseErr.Error())
		return
	}

	entity, err := h.providerService.Get(ctx, parsedSvcID)
	if err != nil {
		h.handleStorageError(w, r, "GetService", err)
		return
	}

	// Return service with redacted secret
	resp := h.toResponse(entity.RedactedCopy())
	h.writeJSON(w, http.StatusOK, resp)
}

// UpdateService handles PUT /api/services/{service-id}
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

	if req.ClientSecret == "" {
		h.writeError(w, http.StatusBadRequest, "validation failed", "client_secret is required")
		return
	}

	skipHTTPSValidation := false
	if h.config != nil {
		skipHTTPSValidation = h.config.Security.SkipThirdpartyHTTPSValidation
	}

	parsedSvcID, parseErr := id.ParseServiceID(serviceID)
	if parseErr != nil {
		h.writeError(w, http.StatusBadRequest, "invalid service ID", parseErr.Error())
		return
	}

	// Parse and validate oauth2_flavor; default to "standard" when omitted,
	// or auto-detect "github" from token endpoint URL.
	flavor := model.OAuth2Flavor(req.OAuth2Flavor)
	if flavor == "" {
		var tokenEndpoint string
		if req.Endpoints != nil {
			tokenEndpoint = req.Endpoints.TokenEndpoint
		}
		flavor = model.InferOAuth2Flavor(tokenEndpoint)
	}

	// created_at is not set here; repo.Update() populates it from storage (no KMS decrypt needed).
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:                  parsedSvcID,
		DisplayName:         req.DisplayName,
		ClientID:            id.ClientID(req.ClientID),
		Secret:              model.NewPlaintextSecret(req.ClientSecret),
		Flavor:              flavor,
		IssuerURI:           req.IssuerURI,
		Discovery:           model.DiscoveryConfig{EnableDiscovery: req.Discovery.EnableDiscovery, MetadataURL: req.Discovery.MetadataURL},
		ProtectedResources:  req.ProtectedResources,
		AuthorizationParams: req.AuthorizationParams,
		UpdatedAt:           time.Now().UTC(),
	}

	// Convert scopes
	entity.Scopes = make([]model.OAuthScope, len(req.Scopes))
	for i, scope := range req.Scopes {
		entity.Scopes[i] = model.OAuthScope{ScopeValue: scope.ScopeValue, Description: scope.Description}
	}

	// For Google flavor, ValidateForUpdate will call enrichForGoogleFlavor internally
	// (parsing the credential once and populating ClientID, Endpoints, IssuerURI).
	// For standard flavor, set endpoints from discovery or the request.
	if flavor != model.OAuth2FlavorGoogle {
		// If discovery is enabled, attempt to discover endpoints
		if req.Discovery.EnableDiscovery {
			endpoints, err := storage.DiscoverOAuth2Endpoints(ctx, req.IssuerURI, req.Discovery.MetadataURL, skipHTTPSValidation)
			if err != nil {
				h.logger.Warn("OAuth2 endpoint discovery failed, falling back to manual endpoints",
					"issuer_uri", req.IssuerURI,
					"error", err)
				if req.Endpoints != nil {
					entity.Endpoints = model.OAuth2Endpoints{
						TokenEndpoint:     req.Endpoints.TokenEndpoint,
						AuthorizeEndpoint: req.Endpoints.AuthorizeEndpoint,
					}
				}
			} else {
				entity.Endpoints = model.OAuth2Endpoints{
					TokenEndpoint:     endpoints.TokenEndpoint,
					AuthorizeEndpoint: endpoints.AuthorizeEndpoint,
				}
			}
		} else if req.Endpoints != nil {
			entity.Endpoints = model.OAuth2Endpoints{
				TokenEndpoint:     req.Endpoints.TokenEndpoint,
				AuthorizeEndpoint: req.Endpoints.AuthorizeEndpoint,
			}
		}
	}

	// Validate entity (for Google flavor, this also enriches ClientID, Endpoints, IssuerURI).
	if err := entity.ValidateForUpdate(skipHTTPSValidation); err != nil {
		h.logger.Warn("service validation failed",
			"client_id", entity.ClientID,
			"error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Defense-in-depth: normalize and validate protected_resources here so the duplicate
	// check below operates on canonical URIs. The domain service repeats this
	// authoritatively before persistence.
	entity.NormalizeProtectedResources()
	if err := entity.ValidateProtectedResources(); err != nil {
		h.logger.Warn("protected_resources validation failed",
			"service_id", entity.ID,
			"error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	if !h.checkProtectedResourceConflicts(ctx, w, r, entity, entity.ID, "UpdateService") {
		return
	}

	// Update in domain service (handles encryption if secret changed)
	if err := h.providerService.Update(ctx, entity); err != nil {
		h.handleStorageError(w, r, "UpdateService", err)
		return
	}

	h.logger.Info("OAuth2 service updated",
		"service_id", entity.ID,
		"client_id", entity.ClientID)

	// Return updated service with redacted secret
	resp := h.toResponse(entity.RedactedCopy())
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *ServicesHandler) checkProtectedResourceConflicts(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	entity *model.ThirdpartyOAuth2ProviderEntity,
	excludeServiceID id.ServiceID,
	operation string,
) bool {
	for _, resource := range entity.ProtectedResources {
		existing, err := h.providerService.FindByProtectedResource(ctx, resource)
		switch {
		case err == nil && existing == nil:
			continue
		case err == nil && existing.ID == excludeServiceID:
			continue
		case err == nil:
			h.logger.Warn("duplicate protected resource",
				"resource", resource,
				"service_id", existing.ID)
			h.writeError(w, http.StatusConflict, "conflict", "protected resource URI already configured for another service")
			return false
		case tokenexchange.IsResourceNotConfigured(err):
			continue
		case tokenexchange.IsResourceAmbiguous(err):
			h.logger.Warn("protected resource lookup is ambiguous",
				"resource", resource,
				"error", err)
			h.writeError(w, http.StatusConflict, "conflict", "protected resource URI already configured for another service")
			return false
		default:
			var storageErr *storage.StorageError
			if errors.As(err, &storageErr) {
				h.handleStorageError(w, r, operation+"DuplicateCheck", err)
				return false
			}
			h.logger.Error("protected resource duplicate check failed",
				"resource", resource,
				"error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error", "")
			return false
		}
	}

	return true
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
	parsedSvcID, parseErr := id.ParseServiceID(serviceID)
	if parseErr != nil {
		h.writeError(w, http.StatusBadRequest, "invalid service ID", parseErr.Error())
		return
	}

	if err := h.providerService.Delete(ctx, parsedSvcID); err != nil {
		// Special handling for conflict errors (grants exist); use errors.As for wrapped errors
		var conflictErr *storage.StorageError
		if errors.As(err, &conflictErr) && conflictErr.Kind == storage.ErrorKindConflict {
			message := conflictErr.Message
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

	entities, err := h.providerService.List(ctx)
	if err != nil {
		h.handleStorageError(w, r, "ListServices", err)
		return
	}

	// Convert to response format with redacted secrets
	responses := make([]ServiceResponse, len(entities))
	for i, entity := range entities {
		responses[i] = h.toResponse(entity.RedactedCopy())
	}

	h.writeJSON(w, http.StatusOK, responses)
}

// toResponse converts a ThirdpartyOAuth2ProviderEntity to ServiceResponse.
func (h *ServicesHandler) toResponse(entity *model.ThirdpartyOAuth2ProviderEntity) ServiceResponse {
	scopes := make([]OAuthScopeResponse, len(entity.Scopes))
	for i, scope := range entity.Scopes {
		scopes[i] = OAuthScopeResponse{
			ScopeValue:  scope.ScopeValue,
			Description: scope.Description,
		}
	}

	// Normalize flavor: if empty, use default.
	flavor := entity.Flavor
	if flavor == "" {
		flavor = model.DefaultOAuth2Flavor
	}

	return ServiceResponse{
		ID:           entity.ID.String(),
		DisplayName:  entity.DisplayName,
		ClientID:     entity.ClientID.String(),
		ClientSecret: entity.Secret.Redacted(),
		OAuth2Flavor: flavor.String(),
		IssuerURI:    entity.IssuerURI,
		Discovery: DiscoveryConfigResponse{
			EnableDiscovery: entity.Discovery.EnableDiscovery,
			MetadataURL:     entity.Discovery.MetadataURL,
		},
		Endpoints: OAuth2EndpointsResponse{
			TokenEndpoint:     entity.Endpoints.TokenEndpoint,
			AuthorizeEndpoint: entity.Endpoints.AuthorizeEndpoint,
		},
		Scopes:              scopes,
		ProtectedResources:  entity.ProtectedResources,
		AuthorizationParams: entity.AuthorizationParams,
		CreatedAt:           entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           entity.UpdatedAt.Format(time.RFC3339),
	}
}

// handleStorageError converts storage errors to HTTP responses.
func (h *ServicesHandler) handleStorageError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	var storageErr *storage.StorageError
	if !errors.As(err, &storageErr) {
		h.logger.Error("unexpected error type", "operation", operation, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	switch storageErr.Kind {
	case storage.ErrorKindNotFound:
		h.writeError(w, http.StatusNotFound, "service not found", storageErr.Message)
	case storage.ErrorKindConflict:
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
func (h *ServicesHandler) writeJSON(w http.ResponseWriter, statusCode int, data any) {
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
