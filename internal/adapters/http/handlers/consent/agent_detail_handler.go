// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
)

// ServiceRequirementForUser represents a service requirement enriched with user session status.
// This is used in the consent screen to display which services the agent requires and user's connection status.
type ServiceRequirementForUser struct {
	ServiceID        string                 `json:"serviceId"`
	ServiceName      string                 `json:"serviceName"`
	RequirementType  string                 `json:"requirementType"` // "mandatory" or "optional"
	RequiredScopes   []ScopeWithDescription `json:"requiredScopes"`
	ConnectionStatus string                 `json:"connectionStatus"` // "connected" or "not_connected"
}

// ScopeWithDescription represents an OAuth2 scope with its description.
type ScopeWithDescription struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AgentDetailHandler handles HTTP requests for retrieving detailed agent information.
// Implements User Story 2: Review Agent-Specific Grants (GET /api/consent/agent/:agentId).
// Phase 6 extension: Includes service requirements with user connection status.
type AgentDetailHandler struct {
	consentService  ConsentService
	authSessionRepo ports.AuthorizationSessionRepository
	logger          *slog.Logger
}

// NewAgentDetailHandler creates a new agent detail handler.
func NewAgentDetailHandler(consentService ConsentService, logger *slog.Logger) *AgentDetailHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentDetailHandler{
		consentService: consentService,
		logger:         logger,
	}
}

// WithAuthorizationSessionRepository sets the authorization session repository.
// Required for FR-028: session-based CIMD metadata retrieval.
func (h *AgentDetailHandler) WithAuthorizationSessionRepository(repo ports.AuthorizationSessionRepository) *AgentDetailHandler {
	h.authSessionRepo = repo
	return h
}

// CIMDMetadataResponse is included in the agent detail response when the authorization
// request originated from a CIMD-based client_id.
type CIMDMetadataResponse struct {
	ClientIDURL         string   `json:"client_id_url"`
	RedirectURI         string   `json:"redirect_uri"`
	VerifiedDomain      string   `json:"verified_domain"`
	IsLocalhostRedirect bool     `json:"is_localhost_redirect"`
	RequestedScopes     []string `json:"requested_scopes"`
}

// GetAgentDetailResponse represents the response for GET /api/consent/agent/:agentId.
type GetAgentDetailResponse struct {
	Data AgentDetailData `json:"data"`
}

// AgentDetailData contains the agent detail and associated services.
type AgentDetailData struct {
	Agent        consent.AgentDetail         `json:"agent"`
	Services     []ServiceRequirementForUser `json:"services"`
	CIMDMetadata *CIMDMetadataResponse       `json:"cimd_metadata,omitempty"`
}

// GetAgentDetail handles GET /api/consent/agent/:agentId
// Returns detailed information about an agent and its required services with user session status.
// Only returns services that are configured as requirements for the agent (not all system services).
//
// Response codes:
// - 200 OK: Returns agent details and services with connection status
// - 400 Bad Request: Invalid agent ID format
// - 401 Unauthorized: Principal not found in context
// - 404 Not Found: Agent doesn't exist
// - 500 Internal Server Error: Service error
func (h *AgentDetailHandler) GetAgentDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := chi.URLParam(r, "agent-id")

	if agentID == "" {
		h.logger.Warn("agent ID is missing in request")
		h.writeError(w, http.StatusBadRequest, "bad request", "agent ID is required")
		return
	}

	userID, ok := getPrincipalFromContext(ctx)
	if !ok {
		h.logger.Warn("principal not found in context")
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "principal required")
		return
	}

	parsedAgentID, parseErr := id.ParseAgentID(agentID)
	if parseErr != nil {
		h.logger.Warn("invalid agent ID format", "agent_id", agentID, "error", parseErr)
		h.writeError(w, http.StatusBadRequest, "bad request", "invalid agent ID format")
		return
	}

	agent, serviceRequirements, err := h.consentService.GetAgentWithServiceRequirements(ctx, id.Principal(userID), parsedAgentID)
	if err != nil {
		if errors.Is(err, consent.ErrAgentNotFound) {
			h.logger.Warn("agent not found", "agent_id", agentID)
			h.writeError(w, http.StatusNotFound, "not found", "agent not found")
			return
		}
		h.logger.Error("failed to get agent", "agent_id", agentID, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	agentDetail := &consent.AgentDetail{
		AgentID:              agent.ID,
		DisplayName:          agent.DisplayName,
		Description:          agent.Description,
		GovernanceURL:        agent.GovernanceURL,
		UserDocumentationURL: agent.UserDocumentationURL,
		AgentInterfaceURL:    agent.AgentInterfaceURL,
	}

	services := toServiceRequirementForUser(serviceRequirements)
	sortServiceRequirements(services)

	cimdMeta, err := h.resolveCIMDMetadata(r, parsedAgentID)
	if err != nil {
		var storErr *storage.StorageError
		if errors.As(err, &storErr) {
			h.logger.Error("authorization session repository error", "agent_id", agentID, "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error", "")
			return
		}
		h.logger.Warn("authorization session error", "agent_id", agentID, "error", err)
		h.writeError(w, http.StatusBadRequest, "bad request", err.Error())
		return
	}

	response := GetAgentDetailResponse{
		Data: AgentDetailData{
			Agent:        *agentDetail,
			Services:     services,
			CIMDMetadata: cimdMeta,
		},
	}

	h.logger.Info("agent detail retrieved",
		"agent_id", agentID,
		"user_id", userID,
		"services_count", len(services))

	h.writeJSON(w, http.StatusOK, response)
}

func toServiceRequirementForUser(reqs []consent.ServiceRequirementStatus) []ServiceRequirementForUser {
	result := make([]ServiceRequirementForUser, len(reqs))
	for i, req := range reqs {
		scopes := make([]ScopeWithDescription, len(req.RequiredScopes))
		for j, s := range req.RequiredScopes {
			scopes[j] = ScopeWithDescription{Name: s.Name, Description: s.Description}
		}
		connStatus := "not_connected"
		if req.IsConnected {
			connStatus = "connected"
		}
		result[i] = ServiceRequirementForUser{
			ServiceID:        req.ServiceID.String(),
			ServiceName:      req.DisplayName,
			RequirementType:  string(req.RequirementType),
			RequiredScopes:   scopes,
			ConnectionStatus: connStatus,
		}
	}
	return result
}

// writeJSON writes a JSON response.
func (h *AgentDetailHandler) writeJSON(w http.ResponseWriter, statusCode int, data any) {
	if err := writeBufferedJSON(w, statusCode, data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response.
func (h *AgentDetailHandler) writeError(w http.ResponseWriter, statusCode int, error string, message string) {
	resp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}

// getPrincipalFromContext extracts the principal from the request context.
func getPrincipalFromContext(ctx context.Context) (string, bool) {
	return principal.FromContext(ctx)
}

// sortServiceRequirements sorts service requirements with mandatory services first, then optional.
func sortServiceRequirements(services []ServiceRequirementForUser) {
	for i := range len(services) {
		for j := i + 1; j < len(services); j++ {
			if services[i].RequirementType == "optional" && services[j].RequirementType == "mandatory" {
				services[i], services[j] = services[j], services[i]
			}
		}
	}
}

// resolveCIMDMetadata resolves CIMD metadata for the consent page.
// When session_id is present (FR-028), loads from the server-side AuthorizationSession.
// CIMD agents (agents with URL-format client_uris) require session_id — falling back
// to query params for CIMD agents would reopen the metadata-spoofing surface that
// server-side sessions were designed to close. Non-CIMD/opaque flows may use query params.
func (h *AgentDetailHandler) resolveCIMDMetadata(r *http.Request, agentID id.AgentID) (*CIMDMetadataResponse, error) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		// Require session_id only when the request is a CIMD authorization flow,
		// identified by a URL-format client_id in the query params. Opaque flows
		// (no client_id, or non-URL client_id) do not use sessions and may use
		// the query-param path. Keying off agent.ClientURIs alone would break opaque
		// flows for agents that also have CIMD client_uris configured.
		clientID := r.URL.Query().Get("client_id")
		if clientID != "" && strings.HasPrefix(clientID, "https://") {
			return nil, errors.New("session_id is required for CIMD agent authorization")
		}
		return nil, nil
	}

	if h.authSessionRepo == nil {
		return nil, errors.New("session_id provided but authorization session repository not configured")
	}

	session, err := h.authSessionRepo.GetBySessionID(r.Context(), sessionID)
	if err != nil {
		var storErr *storage.StorageError
		if errors.As(err, &storErr) && storErr.Kind != storage.ErrorKindNotFound {
			return nil, fmt.Errorf("authorization session lookup failed: %w", err)
		}
		return nil, errors.New("authorization session not found or expired")
	}
	if session.IsExpired() {
		return nil, errors.New("authorization session has expired")
	}
	if session.IsConsumed() {
		return nil, errors.New("authorization session has already been used")
	}
	if session.AgentID != agentID {
		return nil, errors.New("authorization session does not match requested agent")
	}
	userID, _ := getPrincipalFromContext(r.Context())
	if string(session.Principal) != userID {
		return nil, errors.New("authorization session does not belong to this user")
	}

	if session.CIMDMetadata == nil {
		return nil, nil
	}

	u, err := url.Parse(session.CIMDMetadata.ClientID)
	if err != nil {
		return nil, errors.New("invalid client_id in authorization session")
	}

	requestedScopes := strings.Fields(session.Scope)
	if requestedScopes == nil {
		requestedScopes = []string{}
	}

	return &CIMDMetadataResponse{
		ClientIDURL:         session.CIMDMetadata.ClientID,
		RedirectURI:         session.RedirectURI,
		VerifiedDomain:      u.Hostname(),
		IsLocalhostRedirect: isLocalhostURI(session.RedirectURI),
		RequestedScopes:     requestedScopes,
	}, nil
}

// isLocalhostURI returns true if the URI's host is localhost or 127.0.0.1.
func isLocalhostURI(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	h := u.Hostname()
	return h == "localhost" || h == "127.0.0.1" || h == "::1"
}
