// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	domotp2 "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
)

// GrantsHandler handles HTTP requests for user grants management.
// Implements FR-011 through FR-014 (grant CRUD operations).
type GrantsHandler struct {
	consentService        ConsentService
	sessionTokenValidator SessionTokenValidator
	logger                *slog.Logger
}

// NewGrantsHandler creates a new grants handler.
func NewGrantsHandler(consentService ConsentService, logger *slog.Logger, sessionTokenValidator SessionTokenValidator) *GrantsHandler {
	if sessionTokenValidator == nil {
		panic("GrantsHandler requires a non-nil SessionTokenValidator")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &GrantsHandler{
		consentService:        consentService,
		sessionTokenValidator: sessionTokenValidator,
		logger:                logger,
	}
}

// DelegatedTokenRequest represents a delegated token in the request.
type DelegatedTokenRequest struct {
	ThirdpartyOAuth2ServiceID string   `json:"thirdparty_oauth2_service_id"`
	Scopes                    []string `json:"scopes"`
}

// GrantRequest represents the request body for creating/updating a grant.
type GrantRequest struct {
	ValidUntil            *time.Time              `json:"valid_until,omitempty"`
	DelegatedOAuth2Tokens []DelegatedTokenRequest `json:"delegated_oauth2_tokens"`
}

// GrantResponse represents a user grant in the response.
type GrantResponse struct {
	ID                    string                  `json:"id"`
	Principal             string                  `json:"principal"`
	AgentID               string                  `json:"agent_id"`
	ValidUntil            *time.Time              `json:"valid_until,omitempty"`
	DelegatedOAuth2Tokens []DelegatedTokenRequest `json:"delegated_oauth2_tokens"`
	CreatedAt             string                  `json:"created_at"`
	UpdatedAt             string                  `json:"updated_at"`
}

// CreateGrant handles POST /api/consent/agent/:agent-id/grants
// Creates or updates a grant (upsert semantics). Empty delegated_oauth2_tokens
// is valid and creates a grant with no service delegations (e.g. optional-only agents).
//
// Response codes:
// - 201 Created: Grant created/updated
// - 400 Bad Request: Invalid request body or validation error
// - 401 Unauthorized: No principal in context
// - 404 Not Found: Agent doesn't exist
// - 500 Internal Server Error: Service error
func (h *GrantsHandler) CreateGrant(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agent-id")

	// Extract principal from context
	principalValue, ok := principal.FromContext(r.Context())
	if !ok || principalValue == "" {
		h.logger.Warn("principal not found in context")
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	parsedAgentID, err := id.ParseAgentID(agentID)
	if err != nil {
		h.logger.Warn("invalid agent ID format", "agent_id", agentID)
		h.writeError(w, http.StatusBadRequest, "bad request", "agent ID must be a valid UUID")
		return
	}

	// Parse request
	var req GrantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body",
			"error", err,
			"principal", principalValue,
			"agent_id", agentID)
		h.writeError(w, http.StatusBadRequest, "invalid request", "request body must be valid JSON")
		return
	}

	// FR-029: Validate JWE session_token when present.
	// FR-005: redirect_uri fallback is not permitted — session_token is the only
	// accepted state transport for all agent modes (local, proxy, CIMD).
	sessionToken := r.URL.Query().Get("session_token")
	var sessionRedirectURI string
	if sessionToken != "" {
		claims, err := h.sessionTokenValidator.ValidateAuthorizationSessionToken(sessionToken, parsedAgentID, id.Principal(principalValue))
		if err != nil {
			switch {
			case errors.Is(err, domotp2.ErrSessionExpired):
				h.logger.Info("authorization session token expired", "agent_id", agentID, "principal", principalValue)
				h.writeError(w, http.StatusBadRequest, "session_expired", "authorization session has expired, please restart the authorization flow")
			case errors.Is(err, domotp2.ErrSessionAgentMismatch):
				h.logger.Warn("authorization session agent mismatch",
					"expected_agent", parsedAgentID,
					"principal", principalValue)
				h.writeError(w, http.StatusBadRequest, "bad request", "authorization session does not match requested agent")
			case errors.Is(err, domotp2.ErrSessionPrincipalMismatch):
				h.logger.Warn("authorization session principal mismatch", "principal", principalValue)
				h.writeError(w, http.StatusForbidden, "forbidden", "authorization session does not belong to this user")
			default:
				h.logger.Error("authorization session token invalid", "agent_id", agentID, "principal", principalValue, "error", err)
				h.writeError(w, http.StatusBadRequest, "invalid_token", "authorization session token is invalid")
			}
			return
		}
		sessionRedirectURI = claims.OriginalURL
		if sessionRedirectURI == "" {
			h.logger.Error("authorization session token contains empty original_url",
				"agent_id", agentID, "principal", principalValue)
			h.writeError(w, http.StatusInternalServerError, "internal server error", "")
			return
		}
	}

	// Validate valid_until is in future
	if req.ValidUntil != nil && req.ValidUntil.Before(time.Now()) {
		h.logger.Warn("valid_until is in the past",
			"valid_until", req.ValidUntil,
			"principal", principalValue,
			"agent_id", agentID)
		h.writeError(w, http.StatusBadRequest, "invalid request", "valid_until must be in the future")
		return
	}

	// Convert request to domain tokens
	tokens := make([]storage.DelegatedToken, len(req.DelegatedOAuth2Tokens))
	for i, token := range req.DelegatedOAuth2Tokens {
		parsedServiceID, err := id.ParseServiceID(token.ThirdpartyOAuth2ServiceID)
		if err != nil {
			h.logger.Warn("invalid service ID format in grant request",
				"service_id", token.ThirdpartyOAuth2ServiceID,
				"principal", principalValue,
				"agent_id", agentID)
			h.writeError(w, http.StatusBadRequest, "invalid request", fmt.Sprintf("service ID %q must be a valid UUID", token.ThirdpartyOAuth2ServiceID))
			return
		}
		tokens[i] = storage.DelegatedToken{
			ThirdpartyOAuth2ServiceID: parsedServiceID,
			Scopes:                    token.Scopes,
		}
	}

	// Create grant request
	grantReq := &consent.GrantRequest{
		Principal:             id.Principal(principalValue),
		AgentID:               parsedAgentID,
		ValidUntil:            req.ValidUntil,
		DelegatedOAuth2Tokens: tokens,
	}

	handleGrantError := func(err error) {
		// Check for specific error types
		if errors.Is(err, consent.ErrAgentNotFound) {
			h.logger.Warn("agent not found",
				"agent_id", agentID,
				"principal", principalValue)
			h.writeError(w, http.StatusNotFound, "agent not found", "")
			return
		}

		if errors.Is(err, consent.ErrInvalidScopes) {
			h.logger.Warn("invalid scopes requested",
				"agent_id", agentID,
				"principal", principalValue,
				"error", err)
			h.writeError(w, http.StatusBadRequest, "invalid scopes", err.Error())
			return
		}

		if errors.Is(err, consent.ErrServiceNotFound) {
			h.logger.Warn("service not found",
				"agent_id", agentID,
				"principal", principalValue,
				"error", err)
			h.writeError(w, http.StatusBadRequest, "service not found", err.Error())
			return
		}

		if errors.Is(err, consent.ErrGrantValidation) {
			h.logger.Warn("grant validation failed",
				"agent_id", agentID,
				"principal", principalValue,
				"error", err)
			h.writeError(w, http.StatusBadRequest, "invalid request", err.Error())
			return
		}

		h.logger.Error("failed to grant consent",
			"agent_id", agentID,
			"principal", principalValue,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
	}

	grant, err := h.consentService.GrantConsent(r.Context(), grantReq)
	if err != nil {
		handleGrantError(err)
		return
	}

	if grant == nil {
		h.logger.Error("grant consent returned nil grant without error", "agent_id", agentID)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Audit logging
	h.logger.Info("grant created",
		"principal", principalValue,
		"agent_id", agentID,
		"grant_id", grant.ID)

	resp := map[string]interface{}{"data": h.toGrantResponse(grant)}
	if sessionRedirectURI != "" {
		resp["redirect_url"] = sessionRedirectURI
	}
	h.writeJSON(w, http.StatusCreated, resp)
}

// toGrantResponse converts a UserGrant to GrantResponse.
func (h *GrantsHandler) toGrantResponse(grant *storage.UserGrant) GrantResponse {
	tokens := make([]DelegatedTokenRequest, len(grant.DelegatedOAuth2Tokens))
	for i, token := range grant.DelegatedOAuth2Tokens {
		tokens[i] = DelegatedTokenRequest{
			ThirdpartyOAuth2ServiceID: token.ThirdpartyOAuth2ServiceID.String(),
			Scopes:                    token.Scopes,
		}
	}

	return GrantResponse{
		ID:                    grant.ID.String(),
		Principal:             grant.Principal.String(),
		AgentID:               grant.AgentID.String(),
		ValidUntil:            grant.ValidUntil,
		DelegatedOAuth2Tokens: tokens,
		CreatedAt:             grant.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             grant.UpdatedAt.Format(time.RFC3339),
	}
}

// writeJSON writes a JSON response.
func (h *GrantsHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	if err := writeBufferedJSON(w, statusCode, data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response.
func (h *GrantsHandler) writeError(w http.ResponseWriter, statusCode int, error string, message string) {
	resp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}
