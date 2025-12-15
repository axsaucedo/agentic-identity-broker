// Package ports defines interfaces for hexagonal architecture boundaries.
package ports

import (
	"context"
	"net"
	"net/http"
	"time"
)

// ServerPort defines the interface for HTTP server operations.
// This is a hexagonal architecture port - domain logic depends on this interface,
// not concrete implementations.
//
// Implementations: internal/adapters/http/server.go
type ServerPort interface {
	// Name returns the server identifier ("enduser" or "admin").
	Name() string

	// Listen binds to the configured address and port and returns a listener.
	// This is a fast operation that only binds the socket, does not start serving.
	// Returns error if binding fails (e.g., port already in use, permission denied).
	Listen() (net.Listener, error)

	// Serve starts serving HTTP requests on the provided listener.
	// This is a blocking call that runs until the server is shut down or encounters an error.
	// The context can be used to cancel the server operation.
	Serve(ctx context.Context, listener net.Listener) error

	// Shutdown initiates graceful shutdown of the server.
	// Stops accepting new connections immediately.
	// Waits for in-flight requests to complete or timeout to expire.
	// parentCtx is used for forced termination if the parent context is cancelled.
	// timeout specifies the maximum time to wait for graceful shutdown.
	Shutdown(parentCtx context.Context, timeout time.Duration) error

	// HealthStatus returns the current health state of the server.
	// This is a fast, non-blocking operation using atomic reads.
	HealthStatus() HealthState
}

// HealthState represents the operational state of a server.
// States are ordered from startup to shutdown.
type HealthState int32

const (
	// HealthStateStarting indicates the server is initializing.
	// Server is binding to port and setting up routes, not yet ready.
	HealthStateStarting HealthState = iota

	// HealthStateHealthy indicates the server is operational.
	// Server is accepting and processing requests normally.
	HealthStateHealthy

	// HealthStateShuttingDown indicates graceful shutdown is in progress.
	// Server has stopped accepting new connections and is completing in-flight requests.
	HealthStateShuttingDown

	// HealthStateUnhealthy indicates the server encountered an error.
	// Server is not accepting requests and requires restart.
	HealthStateUnhealthy
)

// String returns the string representation of the health state.
func (h HealthState) String() string {
	switch h {
	case HealthStateStarting:
		return "starting"
	case HealthStateHealthy:
		return "healthy"
	case HealthStateShuttingDown:
		return "shutting_down"
	case HealthStateUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// HTTPStatus returns the HTTP status code for this health state.
func (h HealthState) HTTPStatus() int {
	switch h {
	case HealthStateHealthy:
		return http.StatusOK // 200
	case HealthStateStarting, HealthStateShuttingDown, HealthStateUnhealthy:
		return http.StatusServiceUnavailable // 503
	default:
		return http.StatusServiceUnavailable
	}
}
