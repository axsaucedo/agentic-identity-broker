package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// ClientCredentialsHandler handles admin API requests for broker client credentials.
type ClientCredentialsHandler struct {
	credentialRepo ports.ClientCredentialRepository
	agentRepo      ports.AgentRepository
	clientAuth     ports.CredentialGenerator
	logger         *slog.Logger
}

// NewClientCredentialsHandler creates a new ClientCredentialsHandler.
func NewClientCredentialsHandler(
	credentialRepo ports.ClientCredentialRepository,
	agentRepo ports.AgentRepository,
	clientAuth ports.CredentialGenerator,
	logger *slog.Logger,
) *ClientCredentialsHandler {
	return &ClientCredentialsHandler{
		credentialRepo: credentialRepo,
		agentRepo:      agentRepo,
		clientAuth:     clientAuth,
		logger:         logger,
	}
}

// credentialGenerateResponse is the JSON response for POST (generate/rotate).
// Matches ClientCredentialResponse schema in OpenAPI.
type credentialGenerateResponse struct {
	ClientID              string  `json:"client_id"`
	ClientSecret          string  `json:"client_secret"`
	CreatedAt             string  `json:"created_at"`
	PreviousInvalidatedAt *string `json:"previous_invalidated_at,omitempty"`
}

// credentialMetadataResponse is the JSON response for GET (read-only metadata).
// Matches ClientCredentialMetadata schema in OpenAPI.
type credentialMetadataResponse struct {
	ClientID  string  `json:"client_id"`
	CreatedAt string  `json:"created_at"`
	RotatedAt *string `json:"rotated_at,omitempty"`
}

// Generate creates or rotates broker client credentials for an agent.
// POST /api/agents/{agent-id}/client-credentials
func (h *ClientCredentialsHandler) Generate(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agent-id")
	if agentIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "agent ID is required", "")
		return
	}

	agentID, err := id.ParseAgentID(agentIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid agent ID", err.Error())
		return
	}

	// Verify agent exists
	_, err = h.agentRepo.Get(r.Context(), agentID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "agent not found", "")
		return
	}

	// Check if credentials already exist (rotation)
	existing, _ := h.credentialRepo.GetByAgentID(r.Context(), agentID)
	isRotation := existing != nil

	credential, plaintextSecret, err := h.clientAuth.GenerateCredentials(agentID)
	if err != nil {
		h.logger.Error("failed to generate credentials", "agent_id", agentID, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	now := time.Now()
	credential.CreatedAt = now
	if isRotation {
		credential.RotatedAt = &now
		if err := h.credentialRepo.Rotate(r.Context(), agentID, credential); err != nil {
			h.logger.Error("failed to rotate credentials", "agent_id", agentID, "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error", "")
			return
		}
	} else {
		if err := h.credentialRepo.Create(r.Context(), credential); err != nil {
			h.logger.Error("failed to store credentials", "agent_id", agentID, "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error", "")
			return
		}
	}

	resp := credentialGenerateResponse{
		ClientID:     credential.AgentID.String(),
		ClientSecret: plaintextSecret,
		CreatedAt:    credential.CreatedAt.Format(time.RFC3339),
	}
	if credential.RotatedAt != nil {
		rotatedStr := credential.RotatedAt.Format(time.RFC3339)
		resp.PreviousInvalidatedAt = &rotatedStr
	}

	if isRotation {
		h.logger.Info("CredentialRotated",
			"event", "CredentialRotated",
			"agent_id", agentID,
			"client_id", credential.AgentID,
		)
	} else {
		h.logger.Info("CredentialGenerated",
			"event", "CredentialGenerated",
			"agent_id", agentID,
			"client_id", credential.AgentID,
		)
	}

	statusCode := http.StatusCreated
	if isRotation {
		statusCode = http.StatusOK
	}
	h.writeJSON(w, statusCode, resp)
}

// Get retrieves credential metadata (without secret) for an agent.
// GET /api/agents/{agent-id}/client-credentials
func (h *ClientCredentialsHandler) Get(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agent-id")
	if agentIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "agent ID is required", "")
		return
	}

	agentID, err := id.ParseAgentID(agentIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid agent ID", err.Error())
		return
	}

	cred, err := h.credentialRepo.GetByAgentID(r.Context(), agentID)
	if err != nil {
		if isNotFoundErr(err) {
			h.writeError(w, http.StatusNotFound, "credentials not found", "")
			return
		}
		h.logger.Error("failed to get credentials", "agent_id", agentID, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	h.logger.Debug("CredentialRetrieved",
		"event", "CredentialRetrieved",
		"agent_id", agentID,
	)

	resp := credentialMetadataResponse{
		ClientID:  cred.AgentID.String(),
		CreatedAt: cred.CreatedAt.Format(time.RFC3339),
	}
	if cred.RotatedAt != nil {
		rotatedStr := cred.RotatedAt.Format(time.RFC3339)
		resp.RotatedAt = &rotatedStr
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// Revoke deletes broker client credentials for an agent.
// DELETE /api/agents/{agent-id}/client-credentials
func (h *ClientCredentialsHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agent-id")
	if agentIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "agent ID is required", "")
		return
	}

	agentID, err := id.ParseAgentID(agentIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid agent ID", err.Error())
		return
	}

	err = h.credentialRepo.Delete(r.Context(), agentID)
	if err != nil {
		if isNotFoundErr(err) {
			h.writeError(w, http.StatusNotFound, "credentials not found", "")
			return
		}
		h.logger.Error("failed to revoke credentials", "agent_id", agentID, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	h.logger.Info("CredentialRevoked",
		"event", "CredentialRevoked",
		"agent_id", agentID,
	)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ClientCredentialsHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *ClientCredentialsHandler) writeError(w http.ResponseWriter, statusCode int, errMsg string, message string) {
	resp := ErrorResponse{
		Error:   errMsg,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}

func isNotFoundErr(err error) bool {
	if storageErr, ok := err.(*storage.StorageError); ok {
		return storageErr.Kind == storage.ErrorKindNotFound
	}
	return false
}
