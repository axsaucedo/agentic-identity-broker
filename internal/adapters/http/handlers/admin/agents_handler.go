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

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/agents"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/go-chi/chi/v5"
)

// AgentsHandler handles HTTP requests for agent CRUD operations.
type AgentsHandler struct {
	agentService    *agents.Service
	providerService *thirdparty.ThirdpartyOAuth2ProviderService
	logger          *slog.Logger
}

// NewAgentsHandler creates a new agents handler.
func NewAgentsHandler(agentService *agents.Service, providerService *thirdparty.ThirdpartyOAuth2ProviderService, logger *slog.Logger) *AgentsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentsHandler{
		agentService:    agentService,
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
	ClientID             *string                     `json:"client_id,omitempty"`
	ExternalID           *string                     `json:"external_id,omitempty"`
	DisplayName          string                      `json:"display_name"`
	Description          string                      `json:"description"`
	GovernanceURL        *string                     `json:"governance_url,omitempty"`
	UserDocumentationURL *string                     `json:"user_documentation_url,omitempty"`
	AgentInterfaceURL    *string                     `json:"agent_interface_url,omitempty"`
	ServiceRequirements  []ServiceRequirementRequest `json:"service_requirements,omitempty"`
	RedirectURIs         []string                    `json:"redirect_uris,omitempty"`
	AllowedScopes        []string                    `json:"allowed_scopes,omitempty"`
	ClientURIs           []string                    `json:"client_uris,omitempty"`
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
	ClientID             *string                      `json:"client_id,omitempty"`
	ExternalID           *string                      `json:"external_id,omitempty"`
	DisplayName          string                       `json:"display_name"`
	Description          string                       `json:"description"`
	GovernanceURL        *string                      `json:"governance_url,omitempty"`
	UserDocumentationURL *string                      `json:"user_documentation_url,omitempty"`
	AgentInterfaceURL    *string                      `json:"agent_interface_url,omitempty"`
	ServiceRequirements  []ServiceRequirementResponse `json:"service_requirements,omitempty"`
	RedirectURIs         []string                     `json:"redirect_uris,omitempty"`
	AllowedScopes        []string                     `json:"allowed_scopes,omitempty"`
	ClientURIs           []string                     `json:"client_uris,omitempty"`
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

	serviceReqs, err := h.convertServiceRequirements(req.ServiceRequirements)
	if err != nil {
		h.logger.Warn("invalid service requirements", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid service requirements", err.Error())
		return
	}

	if err := storage.ValidateClientURIsForWrite(req.ClientURIs); err != nil {
		h.logger.Warn("client_uris validation failed", "error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	now := time.Now().UTC()
	agent := &storage.Agent{
		ClientID:             clientIDFromRequest(req.ClientID),
		ExternalID:           convertExternalID(req.ExternalID),
		DisplayName:          req.DisplayName,
		Description:          req.Description,
		GovernanceURL:        req.GovernanceURL,
		UserDocumentationURL: req.UserDocumentationURL,
		AgentInterfaceURL:    req.AgentInterfaceURL,
		ServiceRequirements:  serviceReqs,
		RedirectURIs:         req.RedirectURIs,
		AllowedScopes:        req.AllowedScopes,
		ClientURIs:           req.ClientURIs,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := h.agentService.Create(ctx, agent); err != nil {
		h.handleDomainError(w, r, "CreateAgent", err)
		return
	}

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

	agent, err := h.agentService.Get(ctx, parsedAgentID)
	if err != nil {
		h.handleDomainError(w, r, "GetAgent", err)
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

	serviceReqs, err := h.convertServiceRequirements(req.ServiceRequirements)
	if err != nil {
		h.logger.Warn("invalid service requirements", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid service requirements", err.Error())
		return
	}

	if err := storage.ValidateClientURIsForWrite(req.ClientURIs); err != nil {
		h.logger.Warn("client_uris validation failed", "error", err)
		h.writeError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	agent := &storage.Agent{
		ClientID:             clientIDFromRequest(req.ClientID),
		ExternalID:           convertExternalID(req.ExternalID),
		DisplayName:          req.DisplayName,
		Description:          req.Description,
		GovernanceURL:        req.GovernanceURL,
		UserDocumentationURL: req.UserDocumentationURL,
		AgentInterfaceURL:    req.AgentInterfaceURL,
		ServiceRequirements:  serviceReqs,
		RedirectURIs:         req.RedirectURIs,
		AllowedScopes:        req.AllowedScopes,
		ClientURIs:           req.ClientURIs,
		UpdatedAt:            time.Now().UTC(),
	}

	if err := h.agentService.Update(ctx, parsedAgentID, agent); err != nil {
		h.handleDomainError(w, r, "UpdateAgent", err)
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

// DeleteAgent handles DELETE /api/agents/:agent-id
func (h *AgentsHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := chi.URLParam(r, "agent-id")

	if agentID == "" {
		h.writeError(w, http.StatusBadRequest, "agent ID is required", "")
		return
	}

	parsedID, parseErr := id.ParseAgentID(agentID)
	if parseErr != nil {
		h.writeError(w, http.StatusBadRequest, "invalid agent ID", parseErr.Error())
		return
	}

	if err := h.agentService.Delete(ctx, parsedID); err != nil {
		h.handleDomainError(w, r, "DeleteAgent", err)
		return
	}

	h.logger.Info("agent deleted", "agent_id", agentID)
	w.WriteHeader(http.StatusNoContent)
}

// ListAgents handles GET /api/agents
func (h *AgentsHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	agents, err := h.agentService.List(ctx)
	if err != nil {
		h.handleDomainError(w, r, "ListAgents", err)
		return
	}

	serviceMap := h.batchLoadServices(ctx, agents)

	responses := make([]AgentResponse, len(agents))
	for i, agent := range agents {
		resp, err := h.toResponseWithServiceMap(agent, serviceMap)
		if err != nil {
			h.logger.Error("failed to convert agent to response", "agent_id", agent.ID, "error", err)
			resp = AgentResponse{
				ID:          agent.ID.String(),
				ClientID:    clientIDToString(agent.ClientID),
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
func (h *AgentsHandler) batchLoadServices(ctx context.Context, agents []*storage.Agent) map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity {
	serviceIDs := make(map[id.ServiceID]bool)
	for _, agent := range agents {
		for _, sr := range agent.ServiceRequirements {
			serviceIDs[sr.ServiceID] = true
		}
	}

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
func (h *AgentsHandler) toResponseWithServiceMap(agent *storage.Agent, serviceMap map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity) (AgentResponse, error) {
	resp := AgentResponse{
		ID:                   agent.ID.String(),
		ClientID:             clientIDToString(agent.ClientID),
		ExternalID:           convertExternalIDToString(agent.ExternalID),
		DisplayName:          agent.DisplayName,
		Description:          agent.Description,
		GovernanceURL:        agent.GovernanceURL,
		UserDocumentationURL: agent.UserDocumentationURL,
		AgentInterfaceURL:    agent.AgentInterfaceURL,
		RedirectURIs:         agent.RedirectURIs,
		AllowedScopes:        agent.AllowedScopes,
		ClientURIs:           agent.ClientURIs,
		CreatedAt:            agent.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            agent.UpdatedAt.Format(time.RFC3339),
	}

	if len(agent.ServiceRequirements) > 0 {
		resp.ServiceRequirements = make([]ServiceRequirementResponse, len(agent.ServiceRequirements))
		for i, sr := range agent.ServiceRequirements {
			respSR := ServiceRequirementResponse{
				ServiceID:       sr.ServiceID.String(),
				RequirementType: sr.RequirementType.String(),
				RequiredScopes:  sr.RequiredScopes,
			}
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
		parsedSvcID, parseErr := id.ParseServiceID(req.ServiceID)
		if parseErr != nil {
			return nil, fmt.Errorf("service_requirements[%d]: invalid service_id: %w", i, parseErr)
		}

		reqType := storage.RequirementType(req.RequirementType)
		if !reqType.Valid() {
			return nil, fmt.Errorf("service_requirements[%d]: invalid requirement_type %q (must be 'mandatory' or 'optional')", i, req.RequirementType)
		}

		result[i] = storage.ServiceRequirement{
			ServiceID:       parsedSvcID,
			RequirementType: reqType,
			RequiredScopes:  req.RequiredScopes,
		}
	}

	return result, nil
}

// handleDomainError converts domain/storage errors to HTTP responses.
func (h *AgentsHandler) handleDomainError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	var storageErr *storage.StorageError
	if errors.As(err, &storageErr) {
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
		return
	}

	// Domain validation errors (fmt.Errorf wrapping validation)
	h.logger.Error("operation failed", "operation", operation, "error", err)
	h.writeError(w, http.StatusBadRequest, "invalid request", err.Error())
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

func clientIDFromRequest(s *string) *id.ClientID {
	if s == nil || *s == "" {
		return nil
	}
	c := id.ClientID(*s)
	return &c
}

func clientIDToString(c *id.ClientID) *string {
	if c == nil {
		return nil
	}
	s := c.String()
	return &s
}
