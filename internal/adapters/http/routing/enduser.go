// Package routing provides HTTP route configuration for different server types.
package routing

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/app"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// EnduserRouteConfig provides optional configuration for enduser route setup.
type EnduserRouteConfig struct {
	// AuthenticationConfig for principal extraction
	Authentication ports.AuthenticationConfig

	// Logger for middleware
	Logger *slog.Logger
}

// SetupEnduserRoutes registers all end-user API routes and optional SPA serving.
//
// Route structure:
//
//	API Routes:
//	GET    /api/me                                    - Get current user info
//
//	Consent Routes (authenticated):
//	GET    /api/consent/agents                        - List agents with delegations
//	GET    /api/consent/agent/{agent-id}              - Get agent details
//	GET    /api/consent/agent/{agent-id}/grants       - Get agent grants
//	POST   /api/consent/agent/{agent-id}/grants       - Create grant
//
//	OAuth2 Session Routes (authenticated, optional):
//	GET    /api/third-party/sessions                  - List sessions
//	GET    /api/third-party/{serviceId}/oauth2/authorize - Initiate auth
//	GET    /api/third-party/{serviceId}/oauth2/callback  - Handle callback
//	GET    /api/third-party/{serviceId}/session       - Get session details
//	DELETE /api/third-party/{serviceId}/session       - Terminate session
//
//	OAuth2 Authorization Server Routes (optional):
//	GET    /oauth2/authorize                          - Authorization endpoint
//	POST   /oauth2/token                              - Token endpoint
//	GET    /.well-known/oauth-authorization-server    - Metadata endpoint
//
//	SPA Serving (optional):
//	GET    /*                                         - Serve static SPA files
func SetupEnduserRoutes(r chi.Router, h *app.EnduserHandlers, cfg EnduserRouteConfig) {
	// Register API routes with CORS middleware
	r.Route("/api", func(r chi.Router) {
		// Apply CORS middleware to all API routes
		r.Use(middleware.CORSMiddleware())

		// Create subrouter for authenticated routes
		r.Route("/", func(authRouter chi.Router) {
			// Apply authentication middleware FIRST, before registering any routes
			authRouter.Use(middleware.RequirePrincipalMiddleware(cfg.Authentication, cfg.Logger))

			// Register user info endpoint (GET /api/me)
			authRouter.Get("/me", http.HandlerFunc(h.UserInfo.GetUserInfo).ServeHTTP)

			// Register OAuth2 sessions routes if handler is available
			if h.OAuth2Sessions != nil {
				h.OAuth2Sessions.RegisterRoutes(authRouter)
			}

			// Register consent routes if handlers are available
			if h.Agents != nil && h.AgentDetail != nil && h.AgentGrants != nil && h.Grants != nil {
				authRouter.Route("/consent", func(consentRouter chi.Router) {
					// Agents list endpoint
					consentRouter.Get("/agents", h.Agents.GetAgentDelegations)

					// Agent-specific routes
					consentRouter.Route("/agent/{agent-id}", func(r chi.Router) {
						r.Get("/", h.AgentDetail.GetAgentDetail)
						r.Get("/grants", h.AgentGrants.GetAgentGrants)
						r.Post("/grants", h.Grants.CreateGrant)
					})
				})
			}
		})
	})

	// Register OAuth2 authorization server endpoints (optional, public routes)
	if h.OAuth2Authorize != nil && h.OAuth2Token != nil && h.OAuth2Metadata != nil {
		// T025: Authorization endpoint with audit middleware and principal requirement
		// GET /oauth2/authorize
		r.With(
			middleware.OAuth2AuditMiddleware(cfg.Logger),
			middleware.RequirePrincipalMiddleware(cfg.Authentication, cfg.Logger),
		).Get("/oauth2/authorize", h.OAuth2Authorize.ServeHTTP)

		// T034: Token endpoint (no authentication required, proxies to upstream)
		// POST /oauth2/token
		cfg.Logger.Info("Registering POST /oauth2/token endpoint")
		r.Post("/oauth2/token", h.OAuth2Token.ServeHTTP)

		// T041: Metadata endpoint (public, RFC 8414 compliant)
		// GET /.well-known/oauth-authorization-server
		r.Get("/.well-known/oauth-authorization-server", h.OAuth2Metadata.ServeHTTP)
	}

	// Register SPA handler if configured (must be last, after /api routes)
	if h.SPA != nil {
		// Redirect root and /consent to /consent/ for better UX
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/consent/", http.StatusMovedPermanently)
		})
		r.Get("/consent", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/consent/", http.StatusMovedPermanently)
		})
		r.Handle("/consent/*", h.SPA)
	}
}
