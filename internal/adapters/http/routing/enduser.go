// Package routing provides HTTP route configuration for different server types.
package routing

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/app"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwtauth"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// EnduserRouteConfig provides optional configuration for enduser route setup.
type EnduserRouteConfig struct {
	// AuthenticationConfig for principal extraction
	Authentication ports.AuthenticationConfig

	// JWTAuthenticator is an optional JWT authenticator for JWT-based pre-authentication.
	// When nil, only plain-header pre-auth is used (backward-compatible).
	// When set, JWT is used for authentication; absent JWT header is rejected with 401 (fail-closed).
	JWTAuthenticator jwtauth.JWTAuthenticator

	// Logger for middleware
	Logger *slog.Logger

	// CORS configuration for API routes
	CORS ports.CORSConfig

	// CSRFStore provides CSRF token storage for state-mutating consent endpoints.
	// When nil, CSRF protection is not applied.
	CSRFStore *middleware.CSRFStore

	// Telemetry contains observability configuration. When Telemetry.Enabled and
	// Telemetry.Traces.Enabled are both true, otelchi HTTP tracing middleware is registered.
	Telemetry ports.TelemetryConfig
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
//	DELETE /api/consent/agent/{agent-id}/grants       - Revoke grant (FR-014)
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
	// Register OTel HTTP tracing middleware when enabled (ADR-011, T031).
	// Propagators are passed explicitly so the middleware always uses the globally
	// registered propagator and correctly attaches to any configured inbound trace
	// context instead of unconditionally creating new root traces.
	if cfg.Telemetry.Enabled && cfg.Telemetry.Traces.Enabled {
		r.Use(otelchi.Middleware("enduser",
			otelchi.WithChiRoutes(r),
			otelchi.WithRequestMethodInSpanName(true),
			otelchi.WithPropagators(otel.GetTextMapPropagator()),
		))
	}

	// Register API routes with CORS middleware
	r.Route("/api", func(r chi.Router) {
		// Apply CORS middleware (no-op if AllowedOrigins empty)
		r.Use(middleware.CORSMiddleware(cfg.CORS))

		// Create subrouter for authenticated routes
		r.Route("/", func(authRouter chi.Router) {
			// Apply authentication middleware FIRST, before registering any routes
			authRouter.Use(middleware.RequirePrincipalMiddleware(cfg.Authentication, cfg.JWTAuthenticator, cfg.Logger))

			// Register user info endpoint (GET /api/me)
			authRouter.Get("/me", http.HandlerFunc(h.UserInfo.GetUserInfo).ServeHTTP)

			// Register OAuth2 sessions routes if handler is available
			if h.OAuth2Sessions != nil {
				h.OAuth2Sessions.RegisterRoutes(authRouter)
			}

			// Register consent routes if handlers are available
			if h.Agents != nil && h.AgentDetail != nil && h.AgentGrants != nil && h.Grants != nil {
				authRouter.Route("/consent", func(consentRouter chi.Router) {
					if cfg.CSRFStore != nil {
						consentRouter.Use(middleware.CSRFProtection(cfg.CSRFStore))
					}

					// Agents list endpoint
					consentRouter.Get("/agents", h.Agents.GetAgentDelegations)

					// Consent session decode endpoint
					consentRouter.Get("/session", h.AgentDetail.GetConsentSession)

					// Agent-specific routes
					consentRouter.Route("/agent/{agent-id}", func(r chi.Router) {
						r.Get("/", h.AgentDetail.GetAgentDetail)
						r.Get("/grants", h.AgentGrants.GetAgentGrants)
						r.Post("/grants", h.Grants.CreateGrant)
						if h.RevokeGrant != nil {
							r.Delete("/grants", h.RevokeGrant.RevokeGrant)
						}
					})
				})
			}
		})
	})

	// Register OAuth2 authorization server endpoints (optional, public routes).
	// Single authorize handler serves both proxy and issue_token mode.
	// In issue_token mode, the handler's CodeIssuer strategy issues local codes.
	if h.OAuth2Authorize != nil {
		r.With(
			middleware.OAuth2AuditMiddleware(cfg.Logger),
			middleware.RequirePrincipalMiddleware(cfg.Authentication, cfg.JWTAuthenticator, cfg.Logger),
		).Get("/oauth2/authorize", h.OAuth2Authorize.ServeHTTP)
	}

	// Single token handler serves both proxy and issue_token mode.
	// In issue_token mode, the handler's TokenMinting strategy mints local tokens.
	if h.OAuth2Token != nil {
		r.Post("/oauth2/token", h.OAuth2Token.ServeHTTP)
	}

	// RFC 8414 discovery endpoint — single handler serves both modes.
	// The OAuth2Service.GenerateMetadata() includes JWKS URI in issue_token mode.
	if h.OAuth2Metadata != nil {
		r.Get("/.well-known/oauth-authorization-server", h.OAuth2Metadata.ServeHTTP)
	}

	// JWKS endpoint (issue_token mode only — serves signing key public material)
	if h.JWKS != nil {
		r.Get("/oauth2/jwks.json", h.JWKS.ServeJWKS)
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
