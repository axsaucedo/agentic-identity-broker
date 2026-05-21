// Package routing provides HTTP route configuration for different server types.
package routing

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	gorillacsrf "github.com/gorilla/csrf"
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

	// CSRFKey is the 32-byte HMAC signing key for stateless CSRF protection.
	// When nil, CSRF protection is not applied.
	CSRFKey []byte

	// CSRFSecure controls the Secure flag on the CSRF cookie.
	// Set to false in tests (httptest has no TLS).
	CSRFSecure bool

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
//	GET    /api/consent/agents/{agent-id}             - Get agent details
//	GET    /api/consent/agents/{agent-id}/consent-info - Extended consent info with permission sets
//	GET    /api/consent/agents/{agent-id}/grants      - Get agent grants
//	POST   /api/consent/agents/{agent-id}/grants      - Create grant
//	DELETE /api/consent/agents/{agent-id}/grants      - Revoke grant (FR-014)
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
			if h.Agents != nil && h.AgentDetail != nil && h.Grants != nil {
				authRouter.Route("/consent", func(consentRouter chi.Router) {
					if cfg.CSRFKey != nil {
						if !cfg.CSRFSecure {
							consentRouter.Use(csrfPlaintextMiddleware)
						}
						consentRouter.Use(gorillacsrf.Protect(cfg.CSRFKey,
							gorillacsrf.RequestHeader("X-CSRF-Token"),
							gorillacsrf.CookieName("_csrf"),
							gorillacsrf.HttpOnly(true),
							gorillacsrf.Secure(cfg.CSRFSecure),
							gorillacsrf.SameSite(gorillacsrf.SameSiteStrictMode),
							gorillacsrf.Path("/"),
						))
						consentRouter.Use(csrfTokenCookie(cfg.CSRFSecure))
					}

					// Agents list endpoint
					consentRouter.Get("/agents", h.Agents.GetAgentDelegations)

					// Agent-specific routes: /api/consent/agents/{agent-id}/...
					consentRouter.Route("/agents/{agent-id}", func(r chi.Router) {
						r.Get("/", h.AgentDetail.GetAgentDetail)
						r.Get("/grants", h.Grants.GetGrant)
						r.Post("/grants", h.Grants.CreateGrant)
						r.Delete("/grants", h.Grants.RevokeGrant)
						if h.AgentInfo != nil {
							r.Get("/consent-info", h.AgentInfo.GetAgentConsentInfo)
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

// csrfTokenCookie is a middleware that exposes the gorilla/csrf masked token in a
// csrfPlaintextMiddleware marks requests as plaintext HTTP so gorilla/csrf
// skips Referer-based origin checks (only applicable for non-TLS environments).
func csrfPlaintextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, gorillacsrf.PlaintextHTTPRequest(r))
	})
}

// csrfTokenCookie is a middleware that exposes the masked CSRF token in a
// JS-readable cookie named "csrf_token". The frontend reads this cookie and sends
// its value back in the X-CSRF-Token header on mutating requests.
// Must run after gorilla/csrf middleware (which populates the token in context).
func csrfTokenCookie(secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := gorillacsrf.Token(r)
			if token != "" {
				http.SetCookie(w, &http.Cookie{
					Name:     "csrf_token",
					Value:    token,
					Path:     "/",
					HttpOnly: false,
					Secure:   secure,
					SameSite: http.SameSiteStrictMode,
				})
			}
			next.ServeHTTP(w, r)
		})
	}
}
