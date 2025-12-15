package unit

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// TestManagerAtomicStartup tests that both servers start successfully.
func TestManagerAtomicStartup(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	enduserServer := NewMockServerPort("enduser")
	adminServer := NewMockServerPort("admin")

	manager := server.NewManager(enduserServer, adminServer, 30*time.Second, logger)

	// Start manager in a goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- manager.Start(ctx)
	}()

	// Wait a bit for servers to start
	time.Sleep(100 * time.Millisecond)

	// Verify both servers are healthy
	if enduserServer.HealthStatus() != ports.HealthStateHealthy {
		t.Errorf("EndUser server health = %v, want %v", enduserServer.HealthStatus(), ports.HealthStateHealthy)
	}
	if adminServer.HealthStatus() != ports.HealthStateHealthy {
		t.Errorf("Admin server health = %v, want %v", adminServer.HealthStatus(), ports.HealthStateHealthy)
	}

	// Cancel context to stop servers
	cancel()

	// Wait for manager to finish
	err := <-errCh
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("Manager.Start() unexpected error = %v", err)
	}
}

// TestManagerAtomicFailure tests that if one server fails to bind, both stop.
func TestManagerAtomicFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	enduserServer := NewMockServerPort("enduser")
	adminServer := NewMockServerPort("admin").WithListenError(errors.New("port already in use"))

	manager := server.NewManager(enduserServer, adminServer, 30*time.Second, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)

	// Should get an error
	if err == nil {
		t.Fatal("Manager.Start() expected error, got nil")
	}

	// Error should mention the admin server bind failure
	if !strings.Contains(err.Error(), "admin") {
		t.Errorf("Manager.Start() error = %q, want error containing 'admin'", err.Error())
	}
}

// TestManagerServeFailure tests that if one server fails during serve, both stop.
func TestManagerServeFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	enduserServer := NewMockServerPort("enduser")
	adminServer := NewMockServerPort("admin").
		WithServeDuration(50 * time.Millisecond).
		WithServeError(errors.New("serve failed"))

	manager := server.NewManager(enduserServer, adminServer, 30*time.Second, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)

	// Should get an error (either from admin server or context cancellation due to atomic failure)
	if err == nil {
		t.Fatal("Manager.Start() expected error, got nil")
	}

	// The error should contain "server failed" indicating at least one server failed
	// Note: Due to timing, we might get either the admin error or context canceled error
	if !strings.Contains(err.Error(), "server failed") && !strings.Contains(err.Error(), "canceled") {
		t.Errorf("Manager.Start() error = %q, want error containing 'server failed' or 'canceled'", err.Error())
	}
}

// TestManagerGracefulShutdown tests graceful shutdown of both servers.
func TestManagerGracefulShutdown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	enduserServer := NewMockServerPort("enduser")
	adminServer := NewMockServerPort("admin")

	manager := server.NewManager(enduserServer, adminServer, 30*time.Second, logger)

	// Start servers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- manager.Start(ctx)
	}()

	// Wait for servers to start
	time.Sleep(100 * time.Millisecond)

	// Initiate shutdown
	shutdownCtx := context.Background()
	shutdownErr := manager.Shutdown(shutdownCtx)

	if shutdownErr != nil {
		t.Errorf("Manager.Shutdown() unexpected error = %v", shutdownErr)
	}

	// Verify both servers transitioned to shutting down state
	if enduserServer.HealthStatus() != ports.HealthStateShuttingDown {
		t.Errorf("EndUser server health after shutdown = %v, want %v",
			enduserServer.HealthStatus(), ports.HealthStateShuttingDown)
	}
	if adminServer.HealthStatus() != ports.HealthStateShuttingDown {
		t.Errorf("Admin server health after shutdown = %v, want %v",
			adminServer.HealthStatus(), ports.HealthStateShuttingDown)
	}

	// Cancel context to fully stop
	cancel()

	// Wait for manager to finish
	<-errCh
}

// TestHealthStateTransitions tests health state changes.
func TestHealthStateTransitions(t *testing.T) {
	mock := NewMockServerPort("test")

	// Initial state should be starting
	if mock.HealthStatus() != ports.HealthStateStarting {
		t.Errorf("Initial health = %v, want %v", mock.HealthStatus(), ports.HealthStateStarting)
	}

	// After Listen, state should still be starting
	_, err := mock.Listen()
	if err != nil {
		t.Fatalf("Listen() unexpected error = %v", err)
	}
	if mock.HealthStatus() != ports.HealthStateStarting {
		t.Errorf("Health after Listen = %v, want %v", mock.HealthStatus(), ports.HealthStateStarting)
	}

	// Start Serve in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveDone := make(chan error, 1)
	go func() {
		listener, _ := net.Listen("tcp", "127.0.0.1:0")
		serveDone <- mock.Serve(ctx, listener)
	}()

	// Wait a bit for Serve to start
	time.Sleep(50 * time.Millisecond)

	// After Serve starts, state should be healthy
	if mock.HealthStatus() != ports.HealthStateHealthy {
		t.Errorf("Health after Serve = %v, want %v", mock.HealthStatus(), ports.HealthStateHealthy)
	}

	// After Shutdown, state should be shutting down
	_ = mock.Shutdown(context.Background(), 1*time.Second)
	if mock.HealthStatus() != ports.HealthStateShuttingDown {
		t.Errorf("Health after Shutdown = %v, want %v", mock.HealthStatus(), ports.HealthStateShuttingDown)
	}

	cancel()
	<-serveDone
}

// TestHealthStateString tests the String() method.
func TestHealthStateString(t *testing.T) {
	tests := []struct {
		state ports.HealthState
		want  string
	}{
		{ports.HealthStateStarting, "starting"},
		{ports.HealthStateHealthy, "healthy"},
		{ports.HealthStateShuttingDown, "shutting_down"},
		{ports.HealthStateUnhealthy, "unhealthy"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("HealthState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHealthStateHTTPStatus tests the HTTPStatus() method.
func TestHealthStateHTTPStatus(t *testing.T) {
	tests := []struct {
		state ports.HealthState
		want  int
	}{
		{ports.HealthStateStarting, 503},
		{ports.HealthStateHealthy, 200},
		{ports.HealthStateShuttingDown, 503},
		{ports.HealthStateUnhealthy, 503},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if got := tt.state.HTTPStatus(); got != tt.want {
				t.Errorf("HealthState.HTTPStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
