package unit

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// MockServerPort implements ports.ServerPort for testing.
type MockServerPort struct {
	name          string
	listenErr     error
	serveErr      error
	serveDuration time.Duration
	shutdownErr   error
	healthState   int32 // atomic access
	listener      net.Listener
	listenerMu    sync.Mutex // Protects listener access
}

// NewMockServerPort creates a new mock server with configurable behavior.
func NewMockServerPort(name string) *MockServerPort {
	return &MockServerPort{
		name:        name,
		healthState: int32(ports.HealthStateStarting),
	}
}

// WithListenError configures the mock to return an error from Listen().
func (m *MockServerPort) WithListenError(err error) *MockServerPort {
	m.listenErr = err
	return m
}

// WithServeError configures the mock to return an error from Serve().
func (m *MockServerPort) WithServeError(err error) *MockServerPort {
	m.serveErr = err
	return m
}

// WithServeDuration configures how long Serve() should block before returning.
func (m *MockServerPort) WithServeDuration(d time.Duration) *MockServerPort {
	m.serveDuration = d
	return m
}

// WithShutdownError configures the mock to return an error from Shutdown().
func (m *MockServerPort) WithShutdownError(err error) *MockServerPort {
	m.shutdownErr = err
	return m
}

// Name returns the server identifier.
func (m *MockServerPort) Name() string {
	return m.name
}

// Listen simulates binding to a port.
// Returns configured error or a mock listener.
func (m *MockServerPort) Listen() (net.Listener, error) {
	if m.listenErr != nil {
		return nil, m.listenErr
	}

	// Create a mock listener on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("mock listen failed: %w", err)
	}

	m.listenerMu.Lock()
	m.listener = listener
	m.listenerMu.Unlock()
	return listener, nil
}

// Serve simulates serving HTTP requests.
// Blocks for configured duration, then returns configured error.
func (m *MockServerPort) Serve(ctx context.Context, listener net.Listener) error {
	// Transition to healthy state
	atomic.StoreInt32(&m.healthState, int32(ports.HealthStateHealthy))

	// If we have a serve error configured, fail after the duration
	if m.serveErr != nil {
		if m.serveDuration > 0 {
			select {
			case <-time.After(m.serveDuration):
				// Duration elapsed, return error
				atomic.StoreInt32(&m.healthState, int32(ports.HealthStateUnhealthy))
				return m.serveErr
			case <-ctx.Done():
				// Context cancelled before error could happen
				return ctx.Err()
			}
		} else {
			// No duration set, return error immediately
			atomic.StoreInt32(&m.healthState, int32(ports.HealthStateUnhealthy))
			return m.serveErr
		}
	}

	// No error configured, wait for context cancellation
	if m.serveDuration > 0 {
		select {
		case <-time.After(m.serveDuration):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Wait indefinitely for context cancellation
	<-ctx.Done()
	return ctx.Err()
}

// Shutdown simulates graceful shutdown.
func (m *MockServerPort) Shutdown(parentCtx context.Context, timeout time.Duration) error {
	// Transition to shutting down state
	atomic.StoreInt32(&m.healthState, int32(ports.HealthStateShuttingDown))

	// Close the listener if it exists
	m.listenerMu.Lock()
	if m.listener != nil {
		m.listener.Close()
	}
	m.listenerMu.Unlock()

	// Simulate shutdown delay
	time.Sleep(10 * time.Millisecond)

	// Return configured error
	return m.shutdownErr
}

// HealthStatus returns the current health state.
func (m *MockServerPort) HealthStatus() ports.HealthState {
	return ports.HealthState(atomic.LoadInt32(&m.healthState))
}
