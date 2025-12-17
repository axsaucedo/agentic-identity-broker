// Package admin provides HTTP handlers for administrative APIs.
package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AgentsHandler handles HTTP requests for agent CRUD operations.
type AgentsHandler struct {
	repo   ports.AgentRepository
	logger *slog.Logger
}

// NewAgentsHandler creates a new agents handler.
func NewAgentsHandler(repo ports.AgentRepository, logger *slog.Logger) *AgentsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentsHandler{
		repo:   repo,
		logger: logger,
	}
}

// AgentRequest represents the request body for creating/updating an agent.
type AgentRequest struct {
	ClientID             string  `json:"client_id"`
	ExternalID           *string `json:"external_id,omitempty"`
	DisplayName          string  `json:"display_name"`
	Description          string  `json:"description"`
	GovernanceURL        *string `json:"governance_url,omitempty"`
	UserDocumentationURL *string `json:"user_documentation_url,omitempty"`
	AgentInterfaceURL    *string `json:"agent_interface_url,omitempty"`
}

// AgentResponse represents the response body for agent operations.
type AgentResponse struct {
	ID                   string  `json:"id"`
	ClientID             string  `json:"client_id"`
	ExternalID           *string `json:"external_id,omitempty"`
	DisplayName          string  `json:"display_name"`
	Description          string  `json:"description"`
	GovernanceURL        *string `json:"governance_url,omitempty"`
	UserDocumentationURL *string `json:"user_documentation_url,omitempty"`
	AgentInterfaceURL    *string `json:"agent_interface_url,omitempty"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
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

	// Create agent entity
	now := time.Now().UTC()
	agent := &storage.Agent{
		ID:                   uuid.New().String(),
		ClientID:             req.ClientID,
		ExternalID:           req.ExternalID,
		DisplayName:          req.DisplayName,
		Description:          req.Description,
		GovernanceURL:        req.GovernanceURL,
		UserDocumentationURL: req.UserDocumentationURL,
		AgentInterfaceURL:    req.AgentInterfaceURL,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Create in repository
	if err := h.repo.Create(ctx, agent); err != nil {
		h.handleStorageError(w, r, "CreateAgent", err)
		return
	}

	h.logger.Info("agent created", "agent_id", agent.ID, "client_id", agent.ClientID)

	// Return created agent
	resp := h.toResponse(agent)
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

	agent, err := h.repo.Get(ctx, agentID)
	if err != nil {
		h.handleStorageError(w, r, "GetAgent", err)
		return
	}

	resp := h.toResponse(agent)
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

	var req AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		h.writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Get existing agent to preserve created_at
	existing, err := h.repo.Get(ctx, agentID)
	if err != nil {
		h.handleStorageError(w, r, "UpdateAgent", err)
		return
	}

	// Update agent entity
	agent := &storage.Agent{
		ID:                   agentID,
		ClientID:             req.ClientID,
		ExternalID:           req.ExternalID,
		DisplayName:          req.DisplayName,
		Description:          req.Description,
		GovernanceURL:        req.GovernanceURL,
		UserDocumentationURL: req.UserDocumentationURL,
		AgentInterfaceURL:    req.AgentInterfaceURL,
		CreatedAt:            existing.CreatedAt,
		UpdatedAt:            time.Now().UTC(),
	}

	// Update in repository
	if err := h.repo.Update(ctx, agent); err != nil {
		h.handleStorageError(w, r, "UpdateAgent", err)
		return
	}

	h.logger.Info("agent updated", "agent_id", agent.ID, "client_id", agent.ClientID)

	// Return updated agent
	resp := h.toResponse(agent)
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
	if err := h.repo.Delete(ctx, agentID); err != nil {
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

	// Convert to response format
	responses := make([]AgentResponse, len(agents))
	for i, agent := range agents {
		responses[i] = h.toResponse(agent)
	}

	h.writeJSON(w, http.StatusOK, responses)
}

// toResponse converts an Agent entity to AgentResponse.
func (h *AgentsHandler) toResponse(agent *storage.Agent) AgentResponse {
	return AgentResponse{
		ID:                   agent.ID,
		ClientID:             agent.ClientID,
		ExternalID:           agent.ExternalID,
		DisplayName:          agent.DisplayName,
		Description:          agent.Description,
		GovernanceURL:        agent.GovernanceURL,
		UserDocumentationURL: agent.UserDocumentationURL,
		AgentInterfaceURL:    agent.AgentInterfaceURL,
		CreatedAt:            agent.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            agent.UpdatedAt.Format(time.RFC3339),
	}
}

// handleStorageError converts storage errors to HTTP responses.
func (h *AgentsHandler) handleStorageError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	storageErr, ok := err.(*storage.StorageError)
	if !ok {
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
