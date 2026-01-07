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

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
)

// ServerConfig contains the configuration for a Server instance.
type ServerConfig struct {
	Port           int
	Bind           string
	PublicURL      string
	Authentication ports.AuthenticationConfig
}

// Server implements HTTP server lifecycle management using the chi router framework.
// It is generic and not aware of server type ("admin" vs "enduser").
// Route registration is handled via a provided RouteSetupFunc closure.
//
// Responsibilities:
// - Bind to port (Listen)
// - Serve HTTP requests (Serve)
// - Graceful shutdown (Shutdown)
// - Health status reporting (HealthStatus)
// - Common middleware setup (recovery, logging, optional principal)
// - Public health endpoint (/health)
type Server struct {
	config      ServerConfig       // Server configuration (port, bind address, auth)
	router      *chi.Mux           // Chi router for request routing
	routeSetup  func(r chi.Router) // Route registration function (provided by caller)
	httpServer  *http.Server       // Underlying HTTP server
	healthState int32              // Atomic health state
	startTime   time.Time          // Time when server started serving requests
	logger      *slog.Logger       // Structured logger
}

// NewServer creates a new HTTP server instance.
//
// Parameters:
//   - config: Server configuration (port, bind address, authentication)
//   - routeSetup: Function that registers routes on the provided chi.Router
//   - logger: Structured logger for this server instance
//
// The routeSetup function is called during Serve() after middleware setup,
// allowing the caller to register application-specific routes without the
// server knowing about them.
func NewServer(config ServerConfig, routeSetup func(r chi.Router), logger *slog.Logger) *Server {
	return &Server{
		config:      config,
		router:      chi.NewRouter(),
		routeSetup:  routeSetup,
		healthState: int32(HealthStateStarting),
		logger:      logger,
	}
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
	// Setup routes and middleware
	s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Set health state to healthy and record start time
	atomic.StoreInt32(&s.healthState, int32(HealthStateHealthy))
	s.startTime = time.Now()

	s.logger.Info("Server started serving requests",
		"address", listener.Addr().String())

	// Start serving (blocking call)
	if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		atomic.StoreInt32(&s.healthState, int32(HealthStateUnhealthy))
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
	atomic.StoreInt32(&s.healthState, int32(HealthStateShuttingDown))

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
	s.router.Use(middleware.OptionalPrincipalMiddleware(s.config.Authentication, s.logger))

	// Register public health endpoint (no principal required)
	s.router.Get("/health", s.handleHealth())

	// Call the provided route setup function to register application routes
	s.routeSetup(s.router)

	s.logger.Debug("Routes configured")
}

// Router returns the underlying chi router.
// This is useful for registering additional routes from outside the server package.
func (s *Server) Router() *chi.Mux {
	return s.router
}
