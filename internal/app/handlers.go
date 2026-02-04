// Package app provides application-layer wiring and orchestration.
package app

import (
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/enduser"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/admin"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/oauth2_sessions"
)

// AdminHandlers groups all handler instances needed by the admin server.
// This struct is passed to routing functions to avoid scattering handler creation.
type AdminHandlers struct {
	// Agents handler for admin API - manages agent CRUD operations
	Agents *admin.AgentsHandler

	// Services handler for admin API - manages OAuth2 service CRUD operations
	Services *admin.ServicesHandler
}

// EnduserHandlers groups all handler instances needed by the enduser server.
// This struct is passed to routing functions to avoid scattering handler creation.
type EnduserHandlers struct {
	// UserInfo handler for GET /api/me - returns current user information
	UserInfo *consent.UserInfoHandler

	// Consent handlers for /api/consent routes
	Agents      *consent.AgentsHandler
	AgentDetail *consent.AgentDetailHandler
	AgentGrants *consent.AgentGrantsHandler
	Grants      *consent.GrantsHandler

	// OAuth2 sessions handler for /api/third-party routes
	OAuth2Sessions *oauth2_sessions.Handler

	// OAuth2 authorization server handlers
	OAuth2Authorize *enduser.OAuth2AuthorizeHandler
	OAuth2Token     *enduser.OAuth2TokenHandler
	OAuth2Metadata  *enduser.OAuth2MetadataHandler

	// SPA handler for serving static files (must be last)
	SPA *handlers.SPAHandler
}
