// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
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
	consentService    ConsentService
	agentRepository   ports.AgentRepository
	sessionRepository ports.UserSessionRepository
	authSessionRepo   ports.AuthorizationSessionRepository
	providerService   *thirdparty.ThirdpartyOAuth2ProviderService
	logger            *slog.Logger
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

// WithAgentRepository sets the agent repository for this handler.
// Used to lookup the full agent entity including service requirements.
func (h *AgentDetailHandler) WithAgentRepository(repo ports.AgentRepository) *AgentDetailHandler {
	h.agentRepository = repo
	return h
}

// WithSessionRepository sets the session repository for this handler.
// Used to lookup user session status with third-party services.
func (h *AgentDetailHandler) WithSessionRepository(repo ports.UserSessionRepository) *AgentDetailHandler {
	h.sessionRepository = repo
	return h
}

// WithProviderService sets the provider service for this handler.
// Used to lookup service metadata including display names and scope descriptions.
func (h *AgentDetailHandler) WithProviderService(svc *thirdparty.ThirdpartyOAuth2ProviderService) *AgentDetailHandler {
	h.providerService = svc
	return h
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
	ClientName          string   `json:"client_name"`
	ClientIDURL         string   `json:"client_id_url"`
	RedirectURI         string   `json:"redirect_uri"`
	VerifiedDomain      string   `json:"verified_domain"`
	IsLocalhostRedirect bool     `json:"is_localhost_redirect"`
	RequestedScopes     []string `json:"requested_scopes"`
	LogoURI             string   `json:"logo_uri,omitempty"`
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

	// Extract principal from context (middleware ensures this is present)
	userID, ok := getPrincipalFromContext(ctx)
	if !ok {
		h.logger.Warn("principal not found in context")
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "principal required")
		return
	}

	// Call consent service to get agent detail (used for basic agent info)
	parsedAgentID, parseErr := id.ParseAgentID(agentID)
	if parseErr != nil {
		h.logger.Warn("invalid agent ID format", "agent_id", agentID, "error", parseErr)
		h.writeError(w, http.StatusBadRequest, "bad request", "invalid agent ID format")
		return
	}

	agentDetail, _, err := h.consentService.GetAgentDetail(ctx, parsedAgentID)
	if err != nil {
		if errors.Is(err, consent.ErrAgentNotFound) {
			h.logger.Warn("agent not found", "agent_id", agentID)
			h.writeError(w, http.StatusNotFound, "not found", "agent not found")
			return
		}

		h.logger.Error("failed to get agent detail",
			"agent_id", agentID,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Load full agent entity to access service requirements
	agent, err := h.getAgent(ctx, parsedAgentID)
	if err != nil {
		h.logger.Error("failed to load agent",
			"agent_id", agentID,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Build service requirements enriched with user session status
	serviceRequirements, err := h.buildServiceRequirementsForUser(ctx, id.Principal(userID), agent)
	if err != nil {
		h.logger.Error("failed to build service requirements",
			"agent_id", agentID,
			"user_id", userID,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Sort services: mandatory first, then optional
	sortServiceRequirements(serviceRequirements)

	cimdMeta, err := h.resolveCIMDMetadata(r, agent, parsedAgentID)
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
			Services:     serviceRequirements,
			CIMDMetadata: cimdMeta,
		},
	}

	h.logger.Info("agent detail retrieved",
		"agent_id", agentID,
		"user_id", userID,
		"services_count", len(serviceRequirements))

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes a JSON response.
func (h *AgentDetailHandler) writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := encodeJSON(w, data); err != nil {
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

// buildServiceRequirementsForUser enriches agent service requirements with user-specific session status.
// Returns a list of ServiceRequirementForUser with connection status for each service.
// Requirements:
// - Query agent's ServiceRequirements array
// - For each requirement, lookup the ThirdPartyOAuth2Service by service_id
// - Check user's session status with that service (from session repository)
// - Include connection status: "connected" if session exists and is valid, "not_connected" otherwise
// - Resolve scope descriptions from service configuration
// - Return enriched ServiceRequirementForUser model
// Error handling:
// - If service not found: Log warning and skip (service may have been removed)
// - If session lookup fails: Treat as "not_connected"
// - Return partial results if some services unavailable (fail-open for fetch)
func (h *AgentDetailHandler) buildServiceRequirementsForUser(ctx context.Context, userID id.Principal, agent *storage.Agent) ([]ServiceRequirementForUser, error) {
	if len(agent.ServiceRequirements) == 0 {
		return []ServiceRequirementForUser{}, nil
	}

	serviceMap := h.batchLoadServices(ctx, agent)

	var results []ServiceRequirementForUser

	for _, req := range agent.ServiceRequirements {
		svc, ok := serviceMap[req.ServiceID.String()]
		if !ok {
			continue
		}

		// Check user's session status with this service
		session, err := h.sessionRepository.FindByPrincipalAndService(ctx, userID, req.ServiceID)
		if err != nil {
			h.logger.Warn("Error checking session status",
				"user_id", userID,
				"service_id", req.ServiceID,
				"error", err)
			// Treat error as no session
			session = nil
		}

		// Determine connection status
		connStatus := "not_connected"
		if session != nil && !session.IsExpired() {
			connStatus = "connected"
		}

		// Build scope list with descriptions
		var scopes []ScopeWithDescription
		for _, scopeName := range req.RequiredScopes {
			// Find scope description from service configuration
			scopeDesc := ""
			for _, svcScope := range svc.Scopes {
				if svcScope.ScopeValue == scopeName {
					scopeDesc = svcScope.Description
					break
				}
			}

			scope := ScopeWithDescription{
				Name:        scopeName,
				Description: scopeDesc,
			}
			scopes = append(scopes, scope)
		}

		result := ServiceRequirementForUser{
			ServiceID:        req.ServiceID.String(),
			ServiceName:      svc.DisplayName,
			RequirementType:  string(req.RequirementType),
			RequiredScopes:   scopes,
			ConnectionStatus: connStatus,
		}
		results = append(results, result)
	}

	return results, nil
}

// batchLoadServices loads all unique services referenced by an agent's service requirements.
// Returns a map of service_id -> service for efficient lookup, avoiding N KMS decryptions.
func (h *AgentDetailHandler) batchLoadServices(ctx context.Context, agent *storage.Agent) map[string]*model.ThirdpartyOAuth2ProviderEntity {
	serviceIDs := make(map[id.ServiceID]bool)
	for _, sr := range agent.ServiceRequirements {
		serviceIDs[sr.ServiceID] = true
	}

	serviceMap := make(map[string]*model.ThirdpartyOAuth2ProviderEntity)
	for serviceID := range serviceIDs {
		svc, err := h.providerService.Get(ctx, serviceID)
		if err != nil {
			h.logger.Warn("Service not found during requirement building",
				"service_id", serviceID,
				"error", err)
			continue
		}
		serviceMap[serviceID.String()] = svc
	}

	return serviceMap
}

// getAgent loads a full agent entity from storage.
// This is used to access service requirements which are not available in AgentDetail DTO.
func (h *AgentDetailHandler) getAgent(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
	// We need an agent repository. For now, we'll use a workaround by checking if we have access to it
	// through the consentService. Since we don't have direct access, we need to add it to the handler.
	// This will be injected via builder or a new method.
	if h.agentRepository == nil {
		return nil, errors.New("agent repository not configured")
	}
	return h.agentRepository.Get(ctx, agentID)
}

// getPrincipalFromContext extracts the principal from the request context.
func getPrincipalFromContext(ctx context.Context) (string, bool) {
	return principal.FromContext(ctx)
}

// sortServiceRequirements sorts service requirements with mandatory services first, then optional.
func sortServiceRequirements(services []ServiceRequirementForUser) {
	// Sort so mandatory services appear first
	for i := range len(services) {
		for j := i + 1; j < len(services); j++ {
			if services[i].RequirementType == "optional" && services[j].RequirementType == "mandatory" {
				services[i], services[j] = services[j], services[i]
			}
		}
	}
}

// buildCIMDMetadata constructs CIMDMetadataResponse from query params and the agent's CIMD snapshot.
// Returns nil if the request does not originate from a CIMD-based authorization.
func buildCIMDMetadata(r *http.Request, agent *storage.Agent) *CIMDMetadataResponse {
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" || !strings.HasPrefix(clientID, "https://") {
		return nil
	}

	if !slices.Contains(agent.ClientURIs, clientID) {
		return nil
	}

	redirectURI := r.URL.Query().Get("redirect_uri")
	// Fail closed: accept redirect_uri only when the CIMD snapshot is populated and matches.
	if redirectURI != "" {
		if len(agent.CIMDRedirectURIs) == 0 || !slices.Contains(agent.CIMDRedirectURIs, redirectURI) {
			redirectURI = ""
		}
	}
	// scope is echoed from the authorization request query param. A future
	// hardening step would bind this to server-side consent session state
	// created during HandleAuthorization, preventing caller-controlled scope display.
	scope := r.URL.Query().Get("scope")

	u, err := url.Parse(clientID)
	if err != nil {
		return nil
	}
	verifiedDomain := u.Hostname()

	var requestedScopes []string
	if scope != "" {
		requestedScopes = strings.Fields(scope)
	}
	if requestedScopes == nil {
		requestedScopes = []string{}
	}

	clientName := agent.DisplayName
	if agent.CIMDClientName != nil && *agent.CIMDClientName != "" {
		clientName = *agent.CIMDClientName
	}

	logoURI := ""
	if agent.CIMDLogoURI != nil {
		logoURI = *agent.CIMDLogoURI
	}

	return &CIMDMetadataResponse{
		ClientName:          clientName,
		ClientIDURL:         clientID,
		RedirectURI:         redirectURI,
		VerifiedDomain:      verifiedDomain,
		IsLocalhostRedirect: isLocalhostURI(redirectURI),
		RequestedScopes:     requestedScopes,
		LogoURI:             logoURI,
	}
}

// resolveCIMDMetadata resolves CIMD metadata for the consent page.
// When session_id is present (FR-028), loads from the server-side AuthorizationSession.
// CIMD agents (agents with URL-format client_uris) require session_id — falling back
// to query params for CIMD agents would reopen the metadata-spoofing surface that
// server-side sessions were designed to close. Non-CIMD/opaque flows may use query params.
func (h *AgentDetailHandler) resolveCIMDMetadata(r *http.Request, agent *storage.Agent, agentID id.AgentID) (*CIMDMetadataResponse, error) {
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
		return buildCIMDMetadata(r, agent), nil
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

	if session.CIMDMetadata == nil {
		return nil, nil
	}

	u, err := url.Parse(session.CIMDMetadata.ClientID)
	if err != nil {
		return nil, errors.New("invalid client_id in authorization session")
	}

	var requestedScopes []string
	if session.Scope != "" {
		requestedScopes = strings.Fields(session.Scope)
	}
	if requestedScopes == nil {
		requestedScopes = []string{}
	}

	clientName := agent.DisplayName
	if session.CIMDMetadata.ClientName != "" {
		clientName = session.CIMDMetadata.ClientName
	}

	return &CIMDMetadataResponse{
		ClientName:          clientName,
		ClientIDURL:         session.CIMDMetadata.ClientID,
		RedirectURI:         session.RedirectURI,
		VerifiedDomain:      u.Hostname(),
		IsLocalhostRedirect: isLocalhostURI(session.RedirectURI),
		RequestedScopes:     requestedScopes,
		LogoURI:             session.CIMDMetadata.LogoURI,
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
