// Package http provides HTTP server adapters for the identity broker.
package http

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/admin"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
	consentservice "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
)

// Server implements the ServerPort interface using the chi router framework.
// It manages the lifecycle of a single HTTP server instance.
type Server struct {
	name           string                                  // Server identifier ("enduser" or "admin")
	config         ports.ServerInstanceConfig              // Server configuration (port, bind address)
	router         *chi.Mux                                // Chi router for request routing
	httpServer     *http.Server                            // Underlying HTTP server
	healthState    int32                                   // Atomic health state (using ports.HealthState as int32)
	startTime      time.Time                               // Time when server started serving requests
	logger         *slog.Logger                            // Structured logger
	agentRepo      ports.AgentRepository                   // Agent repository (optional)
	serviceRepo    ports.ThirdpartyOAuth2ServiceRepository // OAuth2 service repository (optional)
	grantRepo      ports.UserGrantRepository               // User grant repository (optional)
	spaConfig      *SPAConfig                              // SPA configuration (optional)
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

// SetSPAConfig sets the SPA configuration for this server.
// This should be called before Listen() to enable SPA serving.
func (s *Server) SetSPAConfig(config *SPAConfig) {
	s.spaConfig = config
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

// setupEnduserRoutes registers enduser API routes (consent management) and SPA.
func (s *Server) setupEnduserRoutes() {
	// Register API routes with CORS middleware
	s.router.Route("/api", func(r chi.Router) {
		// Apply CORS middleware to all API routes
		r.Use(middleware.CORSMiddleware())

		// Register user info endpoint (GET /api/me)
		// This endpoint requires authentication but no repositories
		userInfoHandler := consent.NewUserInfoHandler(s.logger)
		r.Get("/me", RequirePrincipalMiddleware(s.config.Authentication, s.logger)(
			http.HandlerFunc(userInfoHandler.GetUserInfo),
		).ServeHTTP)

		// Only register consent routes if all required repositories are available
		if s.agentRepo != nil && s.serviceRepo != nil && s.grantRepo != nil {
			// Create consent service
			consentService := consentservice.NewService(s.agentRepo, s.serviceRepo, s.grantRepo)

			// Register consent management routes
			r.Route("/consent", func(consentRouter chi.Router) {
				s.registerConsentRoutes(consentRouter, consentService)
			})
		} else {
			s.logger.Warn("Consent routes not registered - missing required repositories",
				"has_agent_repo", s.agentRepo != nil,
				"has_service_repo", s.serviceRepo != nil,
				"has_grant_repo", s.grantRepo != nil)
		}
	})

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
		r.Post("/", servicesHandler.CreateService)             // POST /api/services
		r.Get("/", servicesHandler.ListServices)               // GET /api/services
		r.Get("/{service-id}", servicesHandler.GetService)     // GET /api/services/:service-id
		r.Put("/{service-id}", servicesHandler.UpdateService)  // PUT /api/services/:service-id
		r.Delete("/{service-id}", servicesHandler.DeleteService) // DELETE /api/services/:service-id
	})
}

// registerConsentRoutes registers consent management routes.
func (s *Server) registerConsentRoutes(r chi.Router, consentService *consentservice.Service) {
	s.logger.Debug("Registering consent routes")

	// Apply authentication middleware to all consent routes
	r.Use(RequirePrincipalMiddleware(s.config.Authentication, s.logger))

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

// Router returns the underlying chi router.
// This is useful for registering additional routes from outside the server package.
func (s *Server) Router() *chi.Mux {
	return s.router
}
