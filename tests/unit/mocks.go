package unit

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"

	httpAdapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// SimpleMockListener is a minimal mock net.Listener for testing.
// It simulates listening on a socket without actually binding a real port.
type SimpleMockListener struct {
	addr      net.Addr
	isClosed  bool
	closeChan chan struct{}
	mu        sync.Mutex
}

// NewSimpleMockListener creates a mock listener.
func NewSimpleMockListener(port int) *SimpleMockListener {
	return &SimpleMockListener{
		addr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
		},
		closeChan: make(chan struct{}),
	}
}

// Accept simulates accepting connections (blocks until closed).
func (l *SimpleMockListener) Accept() (net.Conn, error) {
	l.mu.Lock()
	if l.isClosed {
		l.mu.Unlock()
		return nil, fmt.Errorf("mock listener closed")
	}
	l.mu.Unlock()
	// In real use, Accept() would block and wait for connections
	// For testing, we block until the listener is closed
	<-l.closeChan
	return nil, fmt.Errorf("listener closed")
}

// Close closes the listener.
func (l *SimpleMockListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.isClosed {
		l.isClosed = true
		close(l.closeChan)
	}
	return nil
}

// Addr returns the listener's address.
func (l *SimpleMockListener) Addr() net.Addr {
	return l.addr
}

// TestServerLifecycle provides a test harness for HTTP server startup/shutdown coordination.
// It simulates dual servers that need to coordinate binding and serving phases.
type TestServerLifecycle struct {
	ServerCount   int
	ListenError   []error // Error per server
	ServeError    []error // Error per server
	ServeDuration []time.Duration

	healthState []int32 // Atomic access per server
	listeners   []net.Listener
	listenerMu  sync.Mutex
}

// NewTestServerLifecycle creates a test harness for N servers.
func NewTestServerLifecycle(serverCount int) *TestServerLifecycle {
	h := &TestServerLifecycle{
		ServerCount:   serverCount,
		ListenError:   make([]error, serverCount),
		ServeError:    make([]error, serverCount),
		ServeDuration: make([]time.Duration, serverCount),
		healthState:   make([]int32, serverCount),
		listeners:     make([]net.Listener, serverCount),
	}
	for i := 0; i < serverCount; i++ {
		h.healthState[i] = int32(ports.HealthStateStarting)
	}
	return h
}

// SimulateListen simulates the listen phase for server i.
func (h *TestServerLifecycle) SimulateListen(serverIdx int) (net.Listener, error) {
	if h.ListenError[serverIdx] != nil {
		return nil, h.ListenError[serverIdx]
	}

	listener := NewSimpleMockListener(8000 + serverIdx)
	h.listenerMu.Lock()
	h.listeners[serverIdx] = listener
	h.listenerMu.Unlock()
	return listener, nil
}

// SimulateServe simulates the serve phase for server i.
func (h *TestServerLifecycle) SimulateServe(ctx context.Context, serverIdx int, listener net.Listener) error {
	// Transition to healthy
	atomic.StoreInt32(&h.healthState[serverIdx], int32(ports.HealthStateHealthy))

	if h.ServeError[serverIdx] != nil {
		if h.ServeDuration[serverIdx] > 0 {
			select {
			case <-time.After(h.ServeDuration[serverIdx]):
				atomic.StoreInt32(&h.healthState[serverIdx], int32(ports.HealthStateUnhealthy))
				return h.ServeError[serverIdx]
			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			atomic.StoreInt32(&h.healthState[serverIdx], int32(ports.HealthStateUnhealthy))
			return h.ServeError[serverIdx]
		}
	}

	if h.ServeDuration[serverIdx] > 0 {
		select {
		case <-time.After(h.ServeDuration[serverIdx]):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	<-ctx.Done()
	return ctx.Err()
}

// SimulateShutdown simulates graceful shutdown for server i.
func (h *TestServerLifecycle) SimulateShutdown(serverIdx int, timeout time.Duration) error {
	atomic.StoreInt32(&h.healthState[serverIdx], int32(ports.HealthStateShuttingDown))
	h.listenerMu.Lock()
	if h.listeners[serverIdx] != nil {
		h.listeners[serverIdx].Close()
	}
	h.listenerMu.Unlock()
	time.Sleep(10 * time.Millisecond)
	return nil
}

// HealthStatus returns the current health state for server i.
func (h *TestServerLifecycle) HealthStatus(serverIdx int) ports.HealthState {
	return ports.HealthState(atomic.LoadInt32(&h.healthState[serverIdx]))
}

// NewTestServer creates an HTTP server with a simple no-op route setup for testing.
// This is a convenience function for tests that just need to verify server lifecycle
// without registering actual routes.
func NewTestServer(config httpAdapter.ServerConfig, logger *slog.Logger) *httpAdapter.Server {
	// Simple route setup function that does nothing
	routeSetup := func(r chi.Router) {
		// No-op - health endpoint is already registered by Server.setupRoutes()
	}

	return httpAdapter.NewServer(config, routeSetup, logger)
}
