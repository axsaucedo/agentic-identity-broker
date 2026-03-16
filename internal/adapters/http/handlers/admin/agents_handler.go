// Package admin provides HTTP handlers for administrative APIs.
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
)

// AgentsHandler handles HTTP requests for agent CRUD operations.
type AgentsHandler struct {
	repo            ports.AgentRepository
	providerService *thirdparty.ThirdpartyOAuth2ProviderService
	logger          *slog.Logger
}

// NewAgentsHandler creates a new agents handler.
func NewAgentsHandler(repo ports.AgentRepository, providerService *thirdparty.ThirdpartyOAuth2ProviderService, logger *slog.Logger) *AgentsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentsHandler{
		repo:            repo,
		providerService: providerService,
		logger:          logger,
	}
}

// ServiceRequirementRequest represents a service requirement in the request.
type ServiceRequirementRequest struct {
	ServiceID       string   `json:"service_id"`
	RequirementType string   `json:"requirement_type"`
	RequiredScopes  []string `json:"required_scopes"`
}

// AgentRequest represents the request body for creating/updating an agent.
type AgentRequest struct {
	ClientID             string                      `json:"client_id"`
	ExternalID           *string                     `json:"external_id,omitempty"`
	DisplayName          string                      `json:"display_name"`
	Description          string                      `json:"description"`
	GovernanceURL        *string                     `json:"governance_url,omitempty"`
	UserDocumentationURL *string                     `json:"user_documentation_url,omitempty"`
	AgentInterfaceURL    *string                     `json:"agent_interface_url,omitempty"`
	ServiceRequirements  []ServiceRequirementRequest `json:"service_requirements,omitempty"`
}

// ServiceRequirementResponse represents a service requirement in the response.
type ServiceRequirementResponse struct {
	ServiceID       string   `json:"service_id"`
	ServiceName     string   `json:"service_name,omitempty"`
	RequirementType string   `json:"requirement_type"`
	RequiredScopes  []string `json:"required_scopes"`
}

// AgentResponse represents the response body for agent operations.
type AgentResponse struct {
	ID                   string                       `json:"id"`
	ClientID             string                       `json:"client_id"`
	ExternalID           *string                      `json:"external_id,omitempty"`
	DisplayName          string                       `json:"display_name"`
	Description          string                       `json:"description"`
	GovernanceURL        *string                      `json:"governance_url,omitempty"`
	UserDocumentationURL *string                      `json:"user_documentation_url,omitempty"`
	AgentInterfaceURL    *string                      `json:"agent_interface_url,omitempty"`
	ServiceRequirements  []ServiceRequirementResponse `json:"service_requirements,omitempty"`
	CreatedAt            string                       `json:"created_at"`
	UpdatedAt            string                       `json:"updated_at"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// CreateAgent handles POST /api/agents
func (h *AgentsHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Convert request service requirements to domain model
	serviceReqs, err := h.convertServiceRequirements(req.ServiceRequirements)
	if err != nil {
		h.logger.Warn("invalid service requirements", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid service requirements", err.Error())
		return
	}

	// Validate service requirements (referential integrity)
	if err := h.providerService.ValidateServiceRequirements(ctx, serviceReqs); err != nil {
		h.logger.Warn("service requirements validation failed", "error", err)
		h.writeError(w, http.StatusBadRequest, "service requirements validation failed", err.Error())
		return
	}

	// Create agent entity
	now := time.Now().UTC()
	agent := &storage.Agent{
		ID:                   id.NewAgentID(),
		ClientID:             id.ClientID(req.ClientID),
		ExternalID:           convertExternalID(req.ExternalID),
		DisplayName:          req.DisplayName,
		Description:          req.Description,
		GovernanceURL:        req.GovernanceURL,
		UserDocumentationURL: req.UserDocumentationURL,
		AgentInterfaceURL:    req.AgentInterfaceURL,
		ServiceRequirements:  serviceReqs,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Create in repository
	if err := h.repo.Create(ctx, agent); err != nil {
		h.handleStorageError(w, r, "CreateAgent", err)
		return
	}

	h.logger.Info("agent created", "agent_id", agent.ID, "client_id", agent.ClientID, "service_requirements_count", len(agent.ServiceRequirements))

	// Return created agent
	serviceMap := h.batchLoadServices(ctx, []*storage.Agent{agent})
	resp, err := h.toResponseWithServiceMap(agent, serviceMap)
	if err != nil {
		h.logger.Error("failed to convert agent to response", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}
	h.writeJSON(w, http.StatusCreated, resp)
}

// GetAgent handles GET /api/agents/:agent-id
func (h *AgentsHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := chi.URLParam(r, "agent-id")

	if agentID == "" {
		h.writeError(w, http.StatusBadRequest, "agent ID is required", "")
		return
	}

	parsedAgentID, parseErr := id.ParseAgentID(agentID)
	if parseErr != nil {
		h.writeError(w, http.StatusBadRequest, "invalid agent ID", parseErr.Error())
		return
	}

	agent, err := h.repo.Get(ctx, parsedAgentID)
	if err != nil {
		h.handleStorageError(w, r, "GetAgent", err)
		return
	}

	serviceMap := h.batchLoadServices(ctx, []*storage.Agent{agent})
	resp, err := h.toResponseWithServiceMap(agent, serviceMap)
	if err != nil {
		h.logger.Error("failed to convert agent to response", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}
	h.writeJSON(w, http.StatusOK, resp)
}

// UpdateAgent handles PUT /api/agents/:agent-id
func (h *AgentsHandler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := chi.URLParam(r, "agent-id")

	if agentID == "" {
		h.writeError(w, http.StatusBadRequest, "agent ID is required", "")
		return
	}

	parsedAgentID, parseErr := id.ParseAgentID(agentID)
	if parseErr != nil {
		h.writeError(w, http.StatusBadRequest, "invalid agent ID", parseErr.Error())
		return
	}

	var req AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Get existing agent to preserve created_at
	existing, err := h.repo.Get(ctx, parsedAgentID)
	if err != nil {
		h.handleStorageError(w, r, "UpdateAgent", err)
		return
	}

	// Convert request service requirements to domain model
	serviceReqs, err := h.convertServiceRequirements(req.ServiceRequirements)
	if err != nil {
		h.logger.Warn("invalid service requirements", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid service requirements", err.Error())
		return
	}

	// Validate service requirements (referential integrity)
	if err := h.providerService.ValidateServiceRequirements(ctx, serviceReqs); err != nil {
		h.logger.Warn("service requirements validation failed", "error", err)
		h.writeError(w, http.StatusBadRequest, "service requirements validation failed", err.Error())
		return
	}

	// Update agent entity
	agent := &storage.Agent{
		ID:                   parsedAgentID,
		ClientID:             id.ClientID(req.ClientID),
		ExternalID:           convertExternalID(req.ExternalID),
		DisplayName:          req.DisplayName,
		Description:          req.Description,
		GovernanceURL:        req.GovernanceURL,
		UserDocumentationURL: req.UserDocumentationURL,
		AgentInterfaceURL:    req.AgentInterfaceURL,
		ServiceRequirements:  serviceReqs,
		CreatedAt:            existing.CreatedAt,
		UpdatedAt:            time.Now().UTC(),
	}

	// Update in repository
	if err := h.repo.Update(ctx, agent); err != nil {
		h.handleStorageError(w, r, "UpdateAgent", err)
		return
	}

	h.logger.Info("agent updated", "agent_id", agent.ID, "client_id", agent.ClientID, "service_requirements_count", len(agent.ServiceRequirements))

	// Return updated agent
	serviceMap := h.batchLoadServices(ctx, []*storage.Agent{agent})
	resp, err := h.toResponseWithServiceMap(agent, serviceMap)
	if err != nil {
		h.logger.Error("failed to convert agent to response", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}
	h.writeJSON(w, http.StatusOK, resp)
}

// DeleteAgent handles DELETE /api/agents/:agent-id
func (h *AgentsHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := chi.URLParam(r, "agent-id")

	if agentID == "" {
		h.writeError(w, http.StatusBadRequest, "agent ID is required", "")
		return
	}

	// Delete from repository
	parsedID, parseErr := id.ParseAgentID(agentID)
	if parseErr != nil {
		h.writeError(w, http.StatusBadRequest, "invalid agent ID", parseErr.Error())
		return
	}

	if err := h.repo.Delete(ctx, parsedID); err != nil {
		h.handleStorageError(w, r, "DeleteAgent", err)
		return
	}

	h.logger.Info("agent deleted", "agent_id", agentID)

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// ListAgents handles GET /api/agents
func (h *AgentsHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	agents, err := h.repo.List(ctx)
	if err != nil {
		h.handleStorageError(w, r, "ListAgents", err)
		return
	}

	// Batch lookup all unique service IDs to avoid N+1 queries
	serviceMap := h.batchLoadServices(ctx, agents)

	// Convert to response format
	responses := make([]AgentResponse, len(agents))
	for i, agent := range agents {
		resp, err := h.toResponseWithServiceMap(agent, serviceMap)
		if err != nil {
			h.logger.Error("failed to convert agent to response", "agent_id", agent.ID, "error", err)
			// Continue with partial response
			resp = AgentResponse{
				ID:          agent.ID.String(),
				ClientID:    agent.ClientID.String(),
				DisplayName: agent.DisplayName,
				Description: agent.Description,
				CreatedAt:   agent.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   agent.UpdatedAt.Format(time.RFC3339),
			}
		}
		responses[i] = resp
	}

	h.writeJSON(w, http.StatusOK, responses)
}

// batchLoadServices loads all unique services referenced by agents in a single batch.
// Returns a map of service_id -> service for efficient lookup.
func (h *AgentsHandler) batchLoadServices(ctx context.Context, agents []*storage.Agent) map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity {
	// Collect all unique service IDs
	serviceIDs := make(map[id.ServiceID]bool)
	for _, agent := range agents {
		for _, sr := range agent.ServiceRequirements {
			serviceIDs[sr.ServiceID] = true
		}
	}

	// Load all services
	serviceMap := make(map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity)
	for serviceID := range serviceIDs {
		service, err := h.providerService.Get(ctx, serviceID)
		if err != nil {
			h.logger.Warn("failed to load service for batch", "service_id", serviceID, "error", err)
			continue
		}
		serviceMap[serviceID] = service
	}

	return serviceMap
}

// toResponseWithServiceMap converts an Agent entity to AgentResponse using a pre-loaded service map.
// This avoids N+1 queries when converting multiple agents.
func (h *AgentsHandler) toResponseWithServiceMap(agent *storage.Agent, serviceMap map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity) (AgentResponse, error) {
	resp := AgentResponse{
		ID:                   agent.ID.String(),
		ClientID:             agent.ClientID.String(),
		ExternalID:           convertExternalIDToString(agent.ExternalID),
		DisplayName:          agent.DisplayName,
		Description:          agent.Description,
		GovernanceURL:        agent.GovernanceURL,
		UserDocumentationURL: agent.UserDocumentationURL,
		AgentInterfaceURL:    agent.AgentInterfaceURL,
		CreatedAt:            agent.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            agent.UpdatedAt.Format(time.RFC3339),
	}

	// Convert service requirements and resolve service names from map
	if len(agent.ServiceRequirements) > 0 {
		resp.ServiceRequirements = make([]ServiceRequirementResponse, len(agent.ServiceRequirements))
		for i, sr := range agent.ServiceRequirements {
			respSR := ServiceRequirementResponse{
				ServiceID:       sr.ServiceID.String(),
				RequirementType: sr.RequirementType.String(),
				RequiredScopes:  sr.RequiredScopes,
			}

			// Resolve service name from map (best effort)
			if service, ok := serviceMap[sr.ServiceID]; ok {
				respSR.ServiceName = service.DisplayName
			}

			resp.ServiceRequirements[i] = respSR
		}
	}

	return resp, nil
}

// convertServiceRequirements converts request DTOs to domain models.
func (h *AgentsHandler) convertServiceRequirements(reqSRs []ServiceRequirementRequest) ([]storage.ServiceRequirement, error) {
	if len(reqSRs) == 0 {
		return nil, nil
	}

	result := make([]storage.ServiceRequirement, len(reqSRs))
	for i, req := range reqSRs {
		// Validate and convert requirement type
		reqType := storage.RequirementType(req.RequirementType)
		if err := reqType.Validate(); err != nil {
			return nil, err
		}

		parsedSvcID, err := id.ParseServiceID(req.ServiceID)
		if err != nil {
			return nil, fmt.Errorf("invalid service_id %q: %w", req.ServiceID, err)
		}

		result[i] = storage.ServiceRequirement{
			ServiceID:       parsedSvcID,
			RequirementType: reqType,
			RequiredScopes:  req.RequiredScopes,
		}
	}

	return result, nil
}

// handleStorageError converts storage errors to HTTP responses.
func (h *AgentsHandler) handleStorageError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	var storageErr *storage.StorageError
	if !errors.As(err, &storageErr) {
		h.logger.Error("unexpected error type", "operation", operation, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	switch storageErr.Kind {
	case storage.ErrorKindNotFound:
		h.writeError(w, http.StatusNotFound, "agent not found", storageErr.Message)
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
func (h *AgentsHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response.
func (h *AgentsHandler) writeError(w http.ResponseWriter, statusCode int, error string, message string) {
	resp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}

// convertExternalID converts *string to *id.ExternalID.
func convertExternalID(s *string) *id.ExternalID {
	if s == nil {
		return nil
	}
	eid := id.ExternalID(*s)
	return &eid
}

// convertExternalIDToString converts *id.ExternalID to *string.
func convertExternalIDToString(eid *id.ExternalID) *string {
	if eid == nil {
		return nil
	}
	s := string(*eid)
	return &s
}
