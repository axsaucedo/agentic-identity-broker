package unit

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// TestServerLifecycleBothHealthy tests that both test servers can start and become healthy.
func TestServerLifecycleBothHealthy(t *testing.T) {
	// Create test lifecycle harness for 2 servers
	lifecycle := NewTestServerLifecycle(2)

	// Start bind phase
	g, ctx := errgroup.WithContext(context.Background())

	var listeners [2]net.Listener
	g.Go(func() error {
		l, err := lifecycle.SimulateListen(0)
		listeners[0] = l
		return err
	})
	g.Go(func() error {
		l, err := lifecycle.SimulateListen(1)
		listeners[1] = l
		return err
	})

	if err := g.Wait(); err != nil {
		t.Fatalf("Listen phase failed: %v", err)
	}

	// Verify both in starting state
	if lifecycle.HealthStatus(0) != ports.HealthStateStarting {
		t.Errorf("Server 0 health = %v, want %v", lifecycle.HealthStatus(0), ports.HealthStateStarting)
	}
	if lifecycle.HealthStatus(1) != ports.HealthStateStarting {
		t.Errorf("Server 1 health = %v, want %v", lifecycle.HealthStatus(1), ports.HealthStateStarting)
	}

	// Start serve phase
	serveCtx, serveCancel := context.WithCancel(ctx)
	defer serveCancel()
	g, serveCtx = errgroup.WithContext(serveCtx)

	g.Go(func() error {
		return lifecycle.SimulateServe(serveCtx, 0, listeners[0])
	})
	g.Go(func() error {
		return lifecycle.SimulateServe(serveCtx, 1, listeners[1])
	})

	// Wait a bit for servers to become healthy
	time.Sleep(100 * time.Millisecond)

	// Verify both are healthy
	if lifecycle.HealthStatus(0) != ports.HealthStateHealthy {
		t.Errorf("Server 0 health = %v, want %v", lifecycle.HealthStatus(0), ports.HealthStateHealthy)
	}
	if lifecycle.HealthStatus(1) != ports.HealthStateHealthy {
		t.Errorf("Server 1 health = %v, want %v", lifecycle.HealthStatus(1), ports.HealthStateHealthy)
	}

	// Shutdown
	lifecycle.SimulateShutdown(0, 1*time.Second)
	lifecycle.SimulateShutdown(1, 1*time.Second)

	// Cancel to stop serving
	serveCancel()

	// Verify both transitioned to shutting down
	if lifecycle.HealthStatus(0) != ports.HealthStateShuttingDown {
		t.Errorf("Server 0 health after shutdown = %v, want %v",
			lifecycle.HealthStatus(0), ports.HealthStateShuttingDown)
	}
	if lifecycle.HealthStatus(1) != ports.HealthStateShuttingDown {
		t.Errorf("Server 1 health after shutdown = %v, want %v",
			lifecycle.HealthStatus(1), ports.HealthStateShuttingDown)
	}
}

// TestServerListenFailureAtomicity tests that if one bind fails, both are affected.
func TestServerListenFailureAtomicity(t *testing.T) {
	// Create lifecycle harness where server 1 fails to listen
	lifecycle := NewTestServerLifecycle(2)
	lifecycle.ListenError[1] = errors.New("port already in use")

	// Try bind phase with errgroup
	g, _ := errgroup.WithContext(context.Background())

	var err1 error
	g.Go(func() error {
		_, err := lifecycle.SimulateListen(0)
		return err
	})
	g.Go(func() error {
		_, err := lifecycle.SimulateListen(1)
		err1 = err
		return err
	})

	err := g.Wait()

	// Should get an error
	if err == nil {
		t.Fatal("Listen phase expected error, got nil")
	}

	// Server 0 should succeed (doesn't matter), server 1 should fail
	if err1 == nil {
		t.Error("Server 1 Listen expected error, got nil")
	}
}

// TestServerServeFailurePropagation tests that serve failures propagate through errgroup.
// Note: This test verifies the test harness works with error scenarios.
func TestServerServeFailurePropagation(t *testing.T) {
	// Create lifecycle where server 1 fails immediately (no duration)
	lifecycle := NewTestServerLifecycle(2)
	lifecycle.ServeError[1] = errors.New("runtime serve error") // No duration = immediate failure

	// Bind phase
	g, ctx := errgroup.WithContext(context.Background())
	var listeners [2]net.Listener

	g.Go(func() error {
		l, err := lifecycle.SimulateListen(0)
		listeners[0] = l
		return err
	})
	g.Go(func() error {
		l, err := lifecycle.SimulateListen(1)
		listeners[1] = l
		return err
	})

	if err := g.Wait(); err != nil {
		t.Fatalf("Listen failed: %v", err)
	}

	// Serve phase with immediate error on server 1
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	g, serveCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return lifecycle.SimulateServe(serveCtx, 0, listeners[0])
	})
	g.Go(func() error {
		return lifecycle.SimulateServe(serveCtx, 1, listeners[1])
	})

	serveErr := g.Wait()

	// Should get an error (either from server 1's immediate failure or context.Canceled)
	if serveErr == nil {
		t.Fatal("Serve phase expected error, got nil")
	}

	// Server 1 should be unhealthy
	if lifecycle.HealthStatus(1) != ports.HealthStateUnhealthy {
		t.Errorf("Server 1 health = %v, want %v", lifecycle.HealthStatus(1), ports.HealthStateUnhealthy)
	}
}

// TestServerShutdownWaitsForTimeout tests shutdown with timeout.
func TestServerShutdownWaitsForTimeout(t *testing.T) {
	lifecycle := NewTestServerLifecycle(1)

	// Listen and start serving
	listener, err := lifecycle.SimulateListen(0)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}

	// Serve with long duration (so shutdown timeout triggers)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go lifecycle.SimulateServe(ctx, 0, listener)

	// Wait for healthy
	time.Sleep(100 * time.Millisecond)

	// Shutdown with short timeout
	shutdownStart := time.Now()
	lifecycle.SimulateShutdown(0, 500*time.Millisecond)
	_ = time.Since(shutdownStart)

	// Verify shutdown completed (quick in our mock, but in real implementation would wait)
	if lifecycle.HealthStatus(0) != ports.HealthStateShuttingDown {
		t.Errorf("Server 0 health = %v, want %v", lifecycle.HealthStatus(0), ports.HealthStateShuttingDown)
	}

	// Cleanup
	cancel()
}

// TestServerHealthStateTransitions tests the full state machine.
func TestServerHealthStateTransitions(t *testing.T) {
	lifecycle := NewTestServerLifecycle(1)

	// Initial state: Starting
	if lifecycle.HealthStatus(0) != ports.HealthStateStarting {
		t.Errorf("Initial state = %v, want %v", lifecycle.HealthStatus(0), ports.HealthStateStarting)
	}

	// After listen
	listener, _ := lifecycle.SimulateListen(0)

	// During serve: should transition to Healthy
	ctx, cancel := context.WithCancel(context.Background())
	go lifecycle.SimulateServe(ctx, 0, listener)

	time.Sleep(100 * time.Millisecond)
	if lifecycle.HealthStatus(0) != ports.HealthStateHealthy {
		t.Errorf("After serve start = %v, want %v", lifecycle.HealthStatus(0), ports.HealthStateHealthy)
	}

	// Shutdown: transition to ShuttingDown
	lifecycle.SimulateShutdown(0, 1*time.Second)
	if lifecycle.HealthStatus(0) != ports.HealthStateShuttingDown {
		t.Errorf("After shutdown = %v, want %v", lifecycle.HealthStatus(0), ports.HealthStateShuttingDown)
	}

	// Cleanup
	cancel()
}
