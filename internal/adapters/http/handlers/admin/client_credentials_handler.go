package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// ClientCredentialsHandler handles admin API requests for broker client credentials.
type ClientCredentialsHandler struct {
	credentialRepo ports.BrokerClientCredentialRepository
	agentRepo      ports.AgentRepository
	clientAuth     *oauth2server.ClientAuthService
	logger         *slog.Logger
}

// NewClientCredentialsHandler creates a new ClientCredentialsHandler.
func NewClientCredentialsHandler(
	credentialRepo ports.BrokerClientCredentialRepository,
	agentRepo ports.AgentRepository,
	logger *slog.Logger,
) *ClientCredentialsHandler {
	return &ClientCredentialsHandler{
		credentialRepo: credentialRepo,
		agentRepo:      agentRepo,
		clientAuth:     oauth2server.NewClientAuthService(credentialRepo, agentRepo, logger),
		logger:         logger,
	}
}

// credentialResponse is the JSON response for credential operations.
type credentialResponse struct {
	BrokerClientID string  `json:"broker_client_id"`
	ClientSecret   string  `json:"client_secret,omitempty"` // Only present on generate/rotate
	CreatedAt      string  `json:"created_at"`
	RotatedAt      *string `json:"rotated_at,omitempty"`
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

	if isRotation {
		if err := h.credentialRepo.Delete(r.Context(), agentID); err != nil {
			h.logger.Error("failed to delete existing credentials during rotation", "agent_id", agentID, "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error", "")
			return
		}
	}

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
	}

	if err := h.credentialRepo.Create(r.Context(), credential); err != nil {
		h.logger.Error("failed to store credentials", "agent_id", agentID, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	resp := credentialResponse{
		BrokerClientID: credential.BrokerClientID.String(),
		ClientSecret:   plaintextSecret,
		CreatedAt:      credential.CreatedAt.Format(time.RFC3339),
	}
	if credential.RotatedAt != nil {
		rotatedStr := credential.RotatedAt.Format(time.RFC3339)
		resp.RotatedAt = &rotatedStr
	}

	if isRotation {
		h.logger.Info("CredentialRotated",
			"event", "CredentialRotated",
			"agent_id", agentID,
			"broker_client_id", credential.BrokerClientID,
		)
	} else {
		h.logger.Info("CredentialGenerated",
			"event", "CredentialGenerated",
			"agent_id", agentID,
			"broker_client_id", credential.BrokerClientID,
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

	resp := credentialResponse{
		BrokerClientID: cred.BrokerClientID.String(),
		CreatedAt:      cred.CreatedAt.Format(time.RFC3339),
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
