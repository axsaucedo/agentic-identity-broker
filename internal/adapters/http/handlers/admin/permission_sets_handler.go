// Package admin provides HTTP handlers for administrative APIs.
package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/permissionset"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
)

// PermissionSetsHandler handles HTTP requests for permission set CRUD operations.
type PermissionSetsHandler struct {
	svc    *permissionset.Service
	logger *slog.Logger
}

// NewPermissionSetsHandler creates a new permission sets handler.
func NewPermissionSetsHandler(svc *permissionset.Service, logger *slog.Logger) *PermissionSetsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &PermissionSetsHandler{
		svc:    svc,
		logger: logger,
	}
}

// ServiceScopeRequest represents a service scope in the request.
type ServiceScopeRequest struct {
	ServiceID       string   `json:"service_id"`
	Scopes          []string `json:"scopes"`
	RequirementType string   `json:"requirement_type,omitempty"`
}

// CreatePermissionSetRequest represents the request body for creating/updating a permission set.
type CreatePermissionSetRequest struct {
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	ServiceScopes []ServiceScopeRequest `json:"service_scopes"`
}

// ServiceScopeResponse represents a service scope in the response.
type ServiceScopeResponse struct {
	ServiceID       string   `json:"service_id"`
	Scopes          []string `json:"scopes"`
	RequirementType string   `json:"requirement_type"`
}

// PermissionSetResponse represents the response body for permission set operations.
type PermissionSetResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	ServiceScopes []ServiceScopeResponse `json:"service_scopes"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// PermissionSetListResponse represents a list of permission sets.
type PermissionSetListResponse struct {
	Items []PermissionSetResponse `json:"items"`
}

// Create handles POST /api/permission-sets
func (h *PermissionSetsHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreatePermissionSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Validate request
	if err := h.validateCreateRequest(&req); err != nil {
		h.logger.Warn("validation failed", "error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Convert request service scopes to domain model
	serviceScopes, err := h.convertServiceScopes(req.ServiceScopes)
	if err != nil {
		h.logger.Warn("invalid service scopes", "error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Create permission set entity
	now := time.Now().UTC()
	ps := &storage.PermissionSet{
		ID:            id.NewPermissionSetID(),
		Name:          req.Name,
		Description:   req.Description,
		ServiceScopes: serviceScopes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Create in service
	if err := h.svc.Create(ctx, ps); err != nil {
		h.handleStorageError(w, r, "Create", err)
		return
	}

	h.logger.Info("permission set created", "permission_set_id", ps.ID, "name", ps.Name)

	resp := h.toResponse(ps)
	w.Header().Set("Location", fmt.Sprintf("/api/permission-sets/%s", ps.ID))
	h.writeJSON(w, http.StatusCreated, resp)
}

// Get handles GET /api/permission-sets/:id
func (h *PermissionSetsHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	psIDStr := chi.URLParam(r, "id")

	if psIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "permission set ID is required", "")
		return
	}

	psID, err := id.ParsePermissionSetID(psIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid permission set ID", err.Error())
		return
	}

	ps, err := h.svc.Get(ctx, psID)
	if err != nil {
		h.handleStorageError(w, r, "Get", err)
		return
	}

	resp := h.toResponse(ps)
	h.writeJSON(w, http.StatusOK, resp)
}

// List handles GET /api/permission-sets
func (h *PermissionSetsHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse optional service_id filter
	serviceIDStr := r.URL.Query().Get("service_id")
	var serviceID id.ServiceID

	if serviceIDStr != "" {
		parsedServiceID, err := id.ParseServiceID(serviceIDStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid service_id", err.Error())
			return
		}
		serviceID = parsedServiceID
	}

	permissionSets, err := h.svc.List(ctx, serviceID)
	if err != nil {
		h.handleStorageError(w, r, "List", err)
		return
	}

	// Convert to response format
	items := make([]PermissionSetResponse, 0)
	if permissionSets != nil {
		items = make([]PermissionSetResponse, len(permissionSets))
		for i, ps := range permissionSets {
			items[i] = h.toResponse(ps)
		}
	}

	resp := PermissionSetListResponse{Items: items}
	h.writeJSON(w, http.StatusOK, resp)
}

// Update handles PUT /api/permission-sets/:id
func (h *PermissionSetsHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	psIDStr := chi.URLParam(r, "id")

	if psIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "permission set ID is required", "")
		return
	}

	psID, err := id.ParsePermissionSetID(psIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid permission set ID", err.Error())
		return
	}

	var req CreatePermissionSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Validate request
	if err := h.validateCreateRequest(&req); err != nil {
		h.logger.Warn("validation failed", "error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Convert request service scopes to domain model
	serviceScopes, err := h.convertServiceScopes(req.ServiceScopes)
	if err != nil {
		h.logger.Warn("invalid service scopes", "error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Get existing permission set to preserve created_at
	existing, err := h.svc.Get(ctx, psID)
	if err != nil {
		h.handleStorageError(w, r, "Get", err)
		return
	}

	// Update permission set entity
	ps := &storage.PermissionSet{
		ID:            psID,
		Name:          req.Name,
		Description:   req.Description,
		ServiceScopes: serviceScopes,
		CreatedAt:     existing.CreatedAt,
		UpdatedAt:     time.Now().UTC(),
	}

	// Update in service
	if err := h.svc.Update(ctx, ps); err != nil {
		h.handleStorageError(w, r, "Update", err)
		return
	}

	h.logger.Info("permission set updated", "permission_set_id", ps.ID, "name", ps.Name)

	// Return updated permission set
	resp := h.toResponse(ps)
	h.writeJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /api/permission-sets/:id
func (h *PermissionSetsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	psIDStr := chi.URLParam(r, "id")

	if psIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "permission set ID is required", "")
		return
	}

	psID, err := id.ParsePermissionSetID(psIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid permission set ID", err.Error())
		return
	}

	// Delete from service. Both storage implementations are idempotent (missing ID
	// is a no-op), so NotFound never occurs in practice — treat it as success.
	if err := h.svc.Delete(ctx, psID); err != nil {
		var storageErr *storage.StorageError
		if errors.As(err, &storageErr) && storageErr.Kind == storage.ErrorKindNotFound {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.handleStorageError(w, r, "Delete", err)
		return
	}

	h.logger.Info("permission set deleted", "permission_set_id", psIDStr)

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// validateCreateRequest validates the create permission set request.
func (h *PermissionSetsHandler) validateCreateRequest(req *CreatePermissionSetRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if len(req.Name) > 255 {
		return errors.New("name must not exceed 255 characters")
	}
	if req.Description == "" {
		return errors.New("description is required")
	}
	if len(req.ServiceScopes) == 0 {
		return errors.New("service_scopes must contain at least one entry")
	}

	// Validate each service scope
	for i, ss := range req.ServiceScopes {
		if ss.ServiceID == "" {
			return fmt.Errorf("service_scope[%d]: service_id is required", i)
		}
		if len(ss.Scopes) == 0 {
			return fmt.Errorf("service_scope[%d]: at least one scope is required", i)
		}
		if ss.RequirementType != "" {
			rt := storage.RequirementType(ss.RequirementType)
			if !rt.Valid() {
				return fmt.Errorf("service_scope[%d]: invalid requirement_type %q", i, ss.RequirementType)
			}
		}
	}

	return nil
}

// convertServiceScopes converts request DTOs to domain models.
func (h *PermissionSetsHandler) convertServiceScopes(reqScopes []ServiceScopeRequest) ([]storage.ServiceScope, error) {
	if len(reqScopes) == 0 {
		return nil, errors.New("service_scopes must contain at least one entry")
	}

	result := make([]storage.ServiceScope, len(reqScopes))
	for i, req := range reqScopes {
		parsedServiceID, err := id.ParseServiceID(req.ServiceID)
		if err != nil {
			return nil, fmt.Errorf("service_scope[%d]: invalid service_id", i)
		}

		rt := storage.RequirementTypeOptional
		if req.RequirementType != "" {
			rt = storage.RequirementType(req.RequirementType)
		}

		result[i] = storage.ServiceScope{
			ServiceID:       parsedServiceID,
			Scopes:          req.Scopes,
			RequirementType: rt,
		}
	}

	return result, nil
}

// toResponse converts a PermissionSet domain entity to PermissionSetResponse.
func (h *PermissionSetsHandler) toResponse(ps *storage.PermissionSet) PermissionSetResponse {
	resp := PermissionSetResponse{
		ID:          ps.ID.String(),
		Name:        ps.Name,
		Description: ps.Description,
		CreatedAt:   ps.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   ps.UpdatedAt.Format(time.RFC3339),
	}

	resp.ServiceScopes = make([]ServiceScopeResponse, len(ps.ServiceScopes))
	for i, ss := range ps.ServiceScopes {
		resp.ServiceScopes[i] = ServiceScopeResponse{
			ServiceID:       ss.ServiceID.String(),
			Scopes:          ss.Scopes,
			RequirementType: string(ss.RequirementType),
		}
	}

	return resp
}

// handleStorageError converts storage errors to HTTP responses.
func (h *PermissionSetsHandler) handleStorageError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	var storageErr *storage.StorageError
	if !errors.As(err, &storageErr) {
		h.logger.Error("unexpected error type", "operation", operation, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	switch storageErr.Kind {
	case storage.ErrorKindNotFound:
		h.writeError(w, http.StatusNotFound, "permission set not found", storageErr.Message)
	case storage.ErrorKindConflict:
		h.writeError(w, http.StatusConflict, "conflict", storageErr.Message)
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
func (h *PermissionSetsHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response.
func (h *PermissionSetsHandler) writeError(w http.ResponseWriter, statusCode int, error string, message string) {
	resp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}
