// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	domjwe "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwe"
	domotp2 "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
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

// errSessionExpired is returned by resolveCIMDMetadata when the JWE session token
// cannot be decrypted or has passed its TTL. Callers use errors.Is to distinguish
// this from other validation errors (agent mismatch, principal mismatch) and return
// a machine-readable "session_expired" error code so the frontend can redirect the
// user back through the /oauth2/authorize flow.
var errSessionExpired = errors.New("authorization session expired")

// AgentDetailHandler handles HTTP requests for retrieving detailed agent information.
// Implements User Story 2: Review Agent-Specific Grants (GET /api/consent/agent/:agentId).
// Phase 6 extension: Includes service requirements with user connection status.
type AgentDetailHandler struct {
	consentService  ConsentService
	jweTokenService *domjwe.TokenService
	logger          *slog.Logger
}

// NewAgentDetailHandler creates a new agent detail handler.
func NewAgentDetailHandler(consentService ConsentService, logger *slog.Logger, jweTokenService *domjwe.TokenService) *AgentDetailHandler {
	if jweTokenService == nil {
		panic("AgentDetailHandler requires a non-nil JWE token service")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentDetailHandler{
		consentService:  consentService,
		jweTokenService: jweTokenService,
		logger:          logger,
	}
}

// CIMDMetadataResponse is included in the agent detail response when the authorization
// request originated from a CIMD-based client_id.
type CIMDMetadataResponse struct {
	ClientIDURL     string   `json:"client_id_url"`
	RedirectURI     string   `json:"redirect_uri"`
	VerifiedDomain  string   `json:"verified_domain"`
	RequestedScopes []string `json:"requested_scopes"`
	LogoURI         string   `json:"logo_uri,omitempty"`
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

	cimdMeta, err := h.resolveSessionContext(r, parsedAgentID)
	if err != nil {
		if errors.Is(err, errSessionExpired) {
			h.logger.Warn("authorization session expired", "agent_id", agentID)
			h.writeError(w, http.StatusBadRequest, "session_expired", "authorization session has expired, please restart the authorization flow")
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

// resolveSessionContext decodes the session_token (when present) and returns CIMD display
// metadata extracted from the claims. The session_token itself is passed opaquely to the
// grants endpoint; authorization-resumption state (original_url, PKCE, state) stays server-side.
// CIMD agents require session_token — falling back to query params would reopen the
// metadata-spoofing surface. Non-CIMD/opaque flows do not use session tokens.
func (h *AgentDetailHandler) resolveSessionContext(r *http.Request, agentID id.AgentID) (*CIMDMetadataResponse, error) {
	sessionToken := r.URL.Query().Get("session_token")
	if sessionToken == "" {
		return nil, nil
	}

	var claims domotp2.AuthorizationSessionClaims
	if err := h.jweTokenService.DecryptAndValidate(sessionToken, &claims); err != nil {
		return nil, errSessionExpired
	}
	if claims.AgentID != agentID {
		return nil, errors.New("authorization session does not match requested agent")
	}
	userID, _ := getPrincipalFromContext(r.Context())
	if string(claims.Principal) != userID {
		return nil, errors.New("authorization session does not belong to this user")
	}

	if claims.CIMDMetadata == nil {
		return nil, nil
	}

	u, err := url.Parse(claims.CIMDMetadata.ClientID)
	if err != nil {
		return nil, errors.New("invalid client_id in authorization session")
	}

	orig, err := url.Parse(claims.OriginalURL)
	if err != nil {
		return nil, errors.New("invalid original_url in authorization session")
	}
	q := orig.Query()

	requestedScopes := strings.Fields(q.Get("scope"))
	if len(requestedScopes) == 0 {
		requestedScopes = []string{}
	}

	return &CIMDMetadataResponse{
		ClientIDURL:     claims.CIMDMetadata.ClientID,
		RedirectURI:     q.Get("redirect_uri"),
		VerifiedDomain:  u.Hostname(),
		RequestedScopes: requestedScopes,
		LogoURI:         claims.CIMDMetadata.LogoURI,
	}, nil
}
