// Package http provides HTTP server adapters for the identity broker.
package http

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/enduser"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/admin"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
	oauth2sessions "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/oauth2_sessions"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/noop"
	consentservice "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	oauth2service "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2session"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
)

// Server implements the ServerPort interface using the chi router framework.
// It manages the lifecycle of a single HTTP server instance.
type Server struct {
	name                   string                                  // Server identifier ("enduser" or "admin")
	config                 ports.ServerInstanceConfig              // Server configuration (port, bind address)
	router                 *chi.Mux                                // Chi router for request routing
	httpServer             *http.Server                            // Underlying HTTP server
	healthState            int32                                   // Atomic health state (using ports.HealthState as int32)
	startTime              time.Time                               // Time when server started serving requests
	logger                 *slog.Logger                            // Structured logger
	agentRepo              ports.AgentRepository                   // Agent repository (optional)
	serviceRepo            ports.ThirdpartyOAuth2ServiceRepository // OAuth2 service repository (optional)
	grantRepo              ports.UserGrantRepository               // User grant repository (optional)
	sessionRepo            ports.UserSessionRepository             // User session repository (optional)
	oauth2Config           *oauth2service.OAuth2Config             // OAuth2 configuration (optional)
	spaConfig              *SPAConfig                              // SPA configuration (optional)
	thirdPartyOAuth2Config ports.ThirdPartyOAuth2Config            // OAuth2 configuration from application config
}

// SPAConfig contains SPA serving configuration.
type SPAConfig struct {
	StaticFilesPath string
	ServeEnabled    bool
}

// NewServer creates a new HTTP server instance.
// name: Server identifier ("enduser" or "admin")
// config: Server configuration (port, bind address)
// logger: Structured logger for this server instance
func NewServer(name string, config ports.ServerInstanceConfig, logger *slog.Logger) *Server {
	return &Server{
		name:        name,
		config:      config,
		router:      chi.NewRouter(),
		healthState: int32(ports.HealthStateStarting),
		logger:      logger.With("server", name),
	}
}

// Name returns the server identifier.
func (s *Server) Name() string {
	return s.name
}

// SetAgentRepository sets the agent repository for this server.
// This should be called before Listen() to ensure handlers have access to the repository.
func (s *Server) SetAgentRepository(repo ports.AgentRepository) {
	s.agentRepo = repo
}

// SetServiceRepository sets the OAuth2 service repository for this server.
// This should be called before Listen() to ensure handlers have access to the repository.
func (s *Server) SetServiceRepository(repo ports.ThirdpartyOAuth2ServiceRepository) {
	s.serviceRepo = repo
}

// SetGrantRepository sets the user grant repository for this server.
// This should be called before Listen() to ensure handlers have access to the repository.
func (s *Server) SetGrantRepository(repo ports.UserGrantRepository) {
	s.grantRepo = repo
}

// SetSessionRepository sets the user session repository for this server.
// This should be called before Listen() to ensure handlers have access to the repository.
func (s *Server) SetSessionRepository(repo ports.UserSessionRepository) {
	s.sessionRepo = repo
}

// SetOAuth2Config sets the OAuth2 configuration for this server.
// This should be called before Listen() to enable OAuth2 endpoints.
func (s *Server) SetOAuth2Config(config *oauth2service.OAuth2Config) {
	s.oauth2Config = config
}

// SetSPAConfig sets the SPA configuration for this server.
// This should be called before Listen() to enable SPA serving.
func (s *Server) SetSPAConfig(config *SPAConfig) {
	s.spaConfig = config
}

// SetThirdPartyOAuth2Config sets the OAuth2 configuration from the application config.
// This should be called before Listen() to enable OAuth2 session routes with proper configuration.
// Constitution Principle VII (Configuration-Driven Design) compliance.
func (s *Server) SetThirdPartyOAuth2Config(cfg ports.ThirdPartyOAuth2Config) {
	s.thirdPartyOAuth2Config = cfg
}

// Listen binds to the configured address and port and returns a listener.
// This is a fast operation that only binds the socket, does not start serving.
// Implements IPv6 dual-stack support with automatic IPv4 fallback.
func (s *Server) Listen() (net.Listener, error) {
	// Try IPv6 dual-stack first (:: means all interfaces, both IPv4 and IPv6)
	if s.config.Bind == "::" {
		addr := fmt.Sprintf("[::]:%d", s.config.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			// IPv6 not available, fallback to IPv4
			s.logger.Warn("IPv6 bind failed, falling back to IPv4",
				"address", addr,
				"error", err)

			// Fallback to IPv4-only (0.0.0.0)
			addr = fmt.Sprintf("0.0.0.0:%d", s.config.Port)
			listener, err = net.Listen("tcp", addr)
			if err != nil {
				return nil, fmt.Errorf("failed to bind to %s: %w", addr, err)
			}
			s.logger.Info("Server bound to IPv4",
				"address", addr,
				"port", s.config.Port)
			return listener, nil
		}

		s.logger.Info("Server bound to dual-stack (IPv6/IPv4)",
			"address", addr,
			"port", s.config.Port)
		return listener, nil
	}

	// Specific bind address provided
	addr := fmt.Sprintf("%s:%d", s.config.Bind, s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to %s: %w", addr, err)
	}

	s.logger.Info("Server bound",
		"address", addr,
		"port", s.config.Port,
		"bind", s.config.Bind)
	return listener, nil
}

// Serve starts serving HTTP requests on the provided listener.
// This is a blocking call that runs until the server is shut down or encounters an error.
// The context can be used to cancel the server operation.
func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	// Register routes and middleware
	s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Set health state to healthy and record start time
	atomic.StoreInt32(&s.healthState, int32(ports.HealthStateHealthy))
	s.startTime = time.Now()

	s.logger.Info("Server started serving requests",
		"name", s.name,
		"address", listener.Addr().String())

	// Start serving (blocking call)
	if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		atomic.StoreInt32(&s.healthState, int32(ports.HealthStateUnhealthy))
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown initiates graceful shutdown of the server.
// Sets health state to shutting down, then calls http.Server.Shutdown with timeout.
func (s *Server) Shutdown(parentCtx context.Context, timeout time.Duration) error {
	s.logger.Info("Initiating graceful shutdown",
		"timeout", timeout)

	// Set health state to shutting down
	atomic.StoreInt32(&s.healthState, int32(ports.HealthStateShuttingDown))

	// Create timeout context for shutdown
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// Shutdown with timeout
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Shutdown error", "error", err)
		return fmt.Errorf("shutdown failed: %w", err)
	}

	s.logger.Info("Server shut down successfully")
	return nil
}

// HealthStatus returns the current health state of the server.
// This is a fast, non-blocking operation using atomic reads.
func (s *Server) HealthStatus() ports.HealthState {
	return ports.HealthState(atomic.LoadInt32(&s.healthState))
}

// setupRoutes configures the router with middleware and routes.
func (s *Server) setupRoutes() {
	// Add middleware (recovery must be first to catch panics in other middleware)
	s.router.Use(RecoveryMiddleware(s.logger))
	s.router.Use(LoggingMiddleware(s.logger))

	// Apply optional principal middleware to all routes
	// This extracts principal if present, but doesn't reject requests without one
	s.router.Use(OptionalPrincipalMiddleware(s.config.Authentication, s.logger))

	// Register public health endpoint (no principal required)
	s.router.Get("/health", s.handleHealth())

	// Register API routes based on server type
	if s.name == "admin" && s.agentRepo != nil {
		s.setupAdminRoutes()
	}

	if s.name == "enduser" {
		s.setupEnduserRoutes()
	}

	s.logger.Debug("Routes configured",
		"server", s.name)
}

// setupAdminRoutes registers admin API routes.
func (s *Server) setupAdminRoutes() {
	s.router.Route("/api", func(r chi.Router) {
		// Agent management routes will be registered here
		// This is called from setupRoutes(), after agentRepo is set
		s.registerAgentRoutes(r)
		s.registerServicesRoutes(r)
	})
}

// setupEnduserRoutes registers enduser API routes (consent management, OAuth2 sessions) and SPA.
func (s *Server) setupEnduserRoutes() {
	// Register API routes with CORS middleware
	s.router.Route("/api", func(r chi.Router) {
		// Apply CORS middleware to all API routes
		r.Use(middleware.CORSMiddleware())

		// Create subrouter for authenticated routes
		// Middleware must be applied before any routes are registered on a chi router
		r.Route("/", func(authRouter chi.Router) {
			// Apply authentication middleware FIRST, before registering any routes
			authRouter.Use(RequirePrincipalMiddleware(s.config.Authentication, s.logger))

			// Register user info endpoint (GET /api/me)
			userInfoHandler := consent.NewUserInfoHandler(s.logger)
			authRouter.Get("/me", http.HandlerFunc(userInfoHandler.GetUserInfo).ServeHTTP)

			// Register OAuth2 sessions routes if sessionRepo and grantRepo are available
			if s.sessionRepo != nil && s.grantRepo != nil {
				s.registerOAuth2SessionsRoutes(authRouter)
			} else {
				s.logger.Warn("OAuth2 sessions routes not registered - missing required repositories",
					"has_session_repo", s.sessionRepo != nil,
					"has_grant_repo", s.grantRepo != nil)
			}

			// Only register consent routes if all required repositories are available
			if s.agentRepo != nil && s.serviceRepo != nil && s.grantRepo != nil {
				// Create consent service
				consentService := consentservice.NewService(s.agentRepo, s.serviceRepo, s.grantRepo)

				// Register consent management routes
				authRouter.Route("/consent", func(consentRouter chi.Router) {
					s.registerConsentRoutes(consentRouter, consentService)
				})
			} else {
				s.logger.Warn("Consent routes not registered - missing required repositories",
					"has_agent_repo", s.agentRepo != nil,
					"has_service_repo", s.serviceRepo != nil,
					"has_grant_repo", s.grantRepo != nil)
			}
		})
	})

	// Register OAuth2 endpoints
	s.registerOAuth2Routes()

	// Register SPA handler if configured
	if s.spaConfig != nil && s.spaConfig.ServeEnabled {
		spaHandler := handlers.NewSPAHandler(s.spaConfig.StaticFilesPath, s.logger)
		// Catch-all route for SPA (must be last, after /api routes)
		s.router.Handle("/*", spaHandler)
		s.logger.Info("SPA handler registered",
			"static_path", s.spaConfig.StaticFilesPath)
	}
}

// registerAgentRoutes registers agent CRUD routes.
func (s *Server) registerAgentRoutes(r chi.Router) {
	if s.agentRepo == nil {
		return
	}

	s.logger.Debug("Registering agent routes")

	// Create agents handler
	agentsHandler := admin.NewAgentsHandler(s.agentRepo, s.logger)

	// Register agent routes
	r.Route("/agents", func(r chi.Router) {
		r.Post("/", agentsHandler.CreateAgent)             // POST /api/agents
		r.Get("/", agentsHandler.ListAgents)               // GET /api/agents
		r.Get("/{agent-id}", agentsHandler.GetAgent)       // GET /api/agents/:agent-id
		r.Put("/{agent-id}", agentsHandler.UpdateAgent)    // PUT /api/agents/:agent-id
		r.Delete("/{agent-id}", agentsHandler.DeleteAgent) // DELETE /api/agents/:agent-id
	})
}

// registerServicesRoutes registers third-party OAuth2 service CRUD routes.
func (s *Server) registerServicesRoutes(r chi.Router) {
	if s.serviceRepo == nil {
		return
	}

	s.logger.Debug("Registering services routes")

	// Create services handler
	servicesHandler := admin.NewServicesHandler(s.serviceRepo, s.logger)

	// Register services routes
	r.Route("/services", func(r chi.Router) {
		r.Post("/", servicesHandler.CreateService)               // POST /api/services
		r.Get("/", servicesHandler.ListServices)                 // GET /api/services
		r.Get("/{service-id}", servicesHandler.GetService)       // GET /api/services/:service-id
		r.Put("/{service-id}", servicesHandler.UpdateService)    // PUT /api/services/:service-id
		r.Delete("/{service-id}", servicesHandler.DeleteService) // DELETE /api/services/:service-id
	})
}

// registerOAuth2SessionsRoutes registers OAuth2 sessions routes.
// Note: Authentication middleware must be applied at the parent router level before calling this.
func (s *Server) registerOAuth2SessionsRoutes(r chi.Router) {
	s.logger.Debug("Registering OAuth2 sessions routes")

	if s.sessionRepo == nil || s.grantRepo == nil {
		s.logger.Warn("OAuth2 sessions routes require both session and grant repositories")
		return
	}

	if s.serviceRepo == nil {
		s.logger.Warn("OAuth2 sessions routes require service repository")
		return
	}

	// Get JWE signing key from application configuration
	if s.thirdPartyOAuth2Config.JWESigningKey == "" {
		s.logger.Warn("JWE signing key not configured - OAuth2 sessions routes not registered")
		return
	}

	var jweKey jwk.Key
	// Decode base64 JWE signing key
	keyBytes, err := base64.StdEncoding.DecodeString(s.thirdPartyOAuth2Config.JWESigningKey)
	if err != nil {
		s.logger.Error("failed to decode JWE signing key", "error", err)
		return
	}

	// Import key as JWK
	jweKey, err = jwk.Import(keyBytes)
	if err != nil {
		s.logger.Error("failed to import JWE signing key", "error", err)
		return
	}

	// Create OAuth2SessionService with repositories and JWE key
	// Use PublicURL from server configuration for OAuth2 callbacks (Constitution Principle VII compliance)
	// In production, this would be configured via environment variable (IDENTITY_BROKER_SERVER_ENDUSER_PUBLIC_URL)
	callbackBaseURL := s.config.PublicURL

	// Build service configuration from application config (Constitution Principle VII compliance)
	// This ensures the service uses configured values instead of hardcoded defaults
	cfg := oauth2session.NewConfigFromPorts(s.thirdPartyOAuth2Config, callbackBaseURL)

	sessionService := oauth2session.NewOAuth2SessionService(
		s.serviceRepo,
		s.sessionRepo,
		s.grantRepo,
		s.agentRepo,
		noop.NewNoOpEncryption(), // Use no-op encryption for development (tokens stored in plaintext)
		jweKey,
		cfg,
		s.logger,
	)

	// Create handler with service
	sessionsHandler := oauth2sessions.NewHandler(sessionService)

	// Register routes
	sessionsHandler.RegisterRoutes(r)
}

// registerConsentRoutes registers consent management routes.
// Note: Authentication middleware must be applied at the parent router level before calling this.
func (s *Server) registerConsentRoutes(r chi.Router, consentService *consentservice.Service) {
	s.logger.Debug("Registering consent routes")

	// Create handlers
	agentsHandler := consent.NewAgentsHandler(consentService, s.logger)
	agentDetailHandler := consent.NewAgentDetailHandler(consentService, s.logger)
	agentGrantsHandler := consent.NewAgentGrantsHandler(consentService, s.logger)
	grantsHandler := consent.NewGrantsHandler(consentService, s.logger)

	// Agent delegations list endpoint (User Story 1: View Active Delegations)
	// GET /api/consent/agents - Returns all agents with active delegations for the current user
	r.Get("/agents", agentsHandler.GetAgentDelegations)

	// Register routes for specific agent operations
	// Both User Story 1 and User Story 2 use the same base path with different handlers
	r.Route("/agent/{agent-id}", func(r chi.Router) {
		// Agent detail endpoint (T057: GET /api/consent/agent/:agent-id)
		// Returns agent metadata and all available third-party services
		r.Get("/", agentDetailHandler.GetAgentDetail)

		// Agent-specific grants endpoints
		// GET /api/consent/agent/:agent-id/grants (T062) - Returns all grants the authenticated user has granted to this agent
		r.Get("/grants", agentGrantsHandler.GetAgentGrants)

		// POST /api/consent/agent/:agent-id/grants - Creates a new grant for the authenticated user
		r.Post("/grants", grantsHandler.CreateGrant)
	})
}

// registerOAuth2Routes registers OAuth2 endpoints.
func (s *Server) registerOAuth2Routes() {
	// Only register OAuth2 routes if configuration is available
	if s.oauth2Config == nil {
		s.logger.Debug("OAuth2 routes not registered - missing oauth2 config")
		return
	}

	// Only register if all required repositories are available
	if s.agentRepo == nil || s.serviceRepo == nil || s.grantRepo == nil {
		s.logger.Warn("OAuth2 routes not registered - missing required repositories",
			"has_agent_repo", s.agentRepo != nil,
			"has_service_repo", s.serviceRepo != nil,
			"has_grant_repo", s.grantRepo != nil)
		return
	}

	// Create OAuth2Config with PublicURL from server configuration
	oauth2Config := &oauth2service.OAuth2Config{
		UpstreamAuthorizeEndpoint: s.oauth2Config.UpstreamAuthorizeEndpoint,
		UpstreamTokenEndpoint:     s.oauth2Config.UpstreamTokenEndpoint,
		PublicURL:                 s.config.PublicURL,
		SupportedResponseTypes:    s.oauth2Config.SupportedResponseTypes,
		SupportedGrantTypes:       s.oauth2Config.SupportedGrantTypes,
	}

	// Create OAuth2 service
	oauth2Svc := oauth2service.NewService(s.agentRepo, s.grantRepo, oauth2Config)

	// T025: Authorization endpoint with audit middleware and principal requirement
	// GET /oauth2/authorize
	authorizeHandler := &enduser.OAuth2AuthorizeHandler{
		Service: oauth2Svc,
	}
	s.router.With(
		middleware.OAuth2AuditMiddleware(s.logger),
		RequirePrincipalMiddleware(s.config.Authentication, s.logger),
	).Get("/oauth2/authorize", authorizeHandler.ServeHTTP)

	// T034: Token endpoint (no authentication required, proxies to upstream)
	// POST /oauth2/token
	tokenHandler := &enduser.OAuth2TokenHandler{
		UpstreamTokenURL: s.oauth2Config.UpstreamTokenEndpoint,
	}
	s.router.Post("/oauth2/token", tokenHandler.ServeHTTP)

	// T041: Metadata endpoint (public, RFC 8414 compliant)
	// GET /.well-known/oauth-authorization-server
	metadataHandler := &enduser.OAuth2MetadataHandler{
		Service: oauth2Svc,
	}
	s.router.Get("/.well-known/oauth-authorization-server", metadataHandler.ServeHTTP)

	s.logger.Debug("OAuth2 routes registered")
}

// Router returns the underlying chi router.
// This is useful for registering additional routes from outside the server package.
func (s *Server) Router() *chi.Mux {
	return s.router
}
