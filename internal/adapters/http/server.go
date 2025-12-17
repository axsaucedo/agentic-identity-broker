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

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/admin"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
)

// Server implements the ServerPort interface using the chi router framework.
// It manages the lifecycle of a single HTTP server instance.
type Server struct {
	name        string                     // Server identifier ("enduser" or "admin")
	config      ports.ServerInstanceConfig // Server configuration (port, bind address)
	router      *chi.Mux                   // Chi router for request routing
	httpServer  *http.Server               // Underlying HTTP server
	healthState int32                      // Atomic health state (using ports.HealthState as int32)
	startTime   time.Time                  // Time when server started serving requests
	logger      *slog.Logger               // Structured logger
	agentRepo   ports.AgentRepository      // Agent repository (optional)
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

	s.logger.Debug("Routes configured",
		"server", s.name)
}

// setupAdminRoutes registers admin API routes.
func (s *Server) setupAdminRoutes() {
	s.router.Route("/api", func(r chi.Router) {
		// Agent management routes will be registered here
		// This is called from setupRoutes(), after agentRepo is set
		s.registerAgentRoutes(r)
	})
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

// Router returns the underlying chi router.
// This is useful for registering additional routes from outside the server package.
func (s *Server) Router() *chi.Mux {
	return s.router
}
