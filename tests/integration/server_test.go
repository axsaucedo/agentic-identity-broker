package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	httpAdapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// TestServerStartup tests basic dual-server startup and health endpoints
func TestServerStartup(t *testing.T) {
	// Create logger for tests
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create server configurations with test ports
	enduserConfig := ports.ServerInstanceConfig{
		Port: 18000, // Test port to avoid conflicts
		Bind: "::",
	}
	adminConfig := ports.ServerInstanceConfig{
		Port: 18001, // Test port to avoid conflicts
		Bind: "::",
	}

	// Create server instances
	enduserServer := httpAdapter.NewServer("enduser", enduserConfig, logger)
	adminServer := httpAdapter.NewServer("admin", adminConfig, logger)

	// Create manager
	mgr := server.NewManager(enduserServer, adminServer, 5*time.Second, logger)

	// Start servers in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- mgr.Start(ctx)
	}()

	// Wait for servers to start
	time.Sleep(500 * time.Millisecond)

	// Test enduser server health endpoint
	t.Run("EndUserHealth", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", enduserConfig.Port))
		if err != nil {
			t.Fatalf("Failed to reach enduser health endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var health httpAdapter.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
			t.Fatalf("Failed to decode health response: %v", err)
		}

		if health.Status != "healthy" {
			t.Errorf("Expected status 'healthy', got '%s'", health.Status)
		}
		if health.Server != "enduser" {
			t.Errorf("Expected server 'enduser', got '%s'", health.Server)
		}
	})

	// Test admin server health endpoint
	t.Run("AdminHealth", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", adminConfig.Port))
		if err != nil {
			t.Fatalf("Failed to reach admin health endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var health httpAdapter.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
			t.Fatalf("Failed to decode health response: %v", err)
		}

		if health.Status != "healthy" {
			t.Errorf("Expected status 'healthy', got '%s'", health.Status)
		}
		if health.Server != "admin" {
			t.Errorf("Expected server 'admin', got '%s'", health.Server)
		}
	})

	// Shutdown servers gracefully
	if err := mgr.Shutdown(context.Background()); err != nil {
		t.Logf("Shutdown error (may be expected): %v", err)
	}

	// Cancel context to ensure goroutines exit
	cancel()

	// Wait for servers to stop
	select {
	case err := <-errChan:
		if err != nil && err != context.Canceled && err != http.ErrServerClosed {
			t.Errorf("Unexpected server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server did not stop within timeout")
	}
}

// TestPortConnectivity tests IPv4 and IPv6 connectivity
func TestPortConnectivity(t *testing.T) {
	// Create logger for tests
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Create server configuration with test port
	config := ports.ServerInstanceConfig{
		Port: 18002,
		Bind: "::", // Dual-stack
	}

	// Create and start server
	srv := httpAdapter.NewServer("test", config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		listener, err := srv.Listen()
		if err != nil {
			errChan <- err
			return
		}
		errChan <- srv.Serve(ctx, listener)
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	// Test IPv4 connectivity
	t.Run("IPv4", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", config.Port))
		if err != nil {
			t.Fatalf("Failed to connect via IPv4: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	// Test IPv6 connectivity
	t.Run("IPv6", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://[::1]:%d/health", config.Port))
		if err != nil {
			t.Skipf("IPv6 not available: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	// Test localhost DNS resolution
	t.Run("Localhost", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", config.Port))
		if err != nil {
			t.Fatalf("Failed to connect via localhost: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	// Shutdown server gracefully
	if err := srv.Shutdown(context.Background(), 1*time.Second); err != nil {
		t.Logf("Shutdown error (may be expected): %v", err)
	}

	// Cancel context to ensure goroutine exits
	cancel()

	select {
	case <-errChan:
		// Server stopped
	case <-time.After(2 * time.Second):
		t.Error("Server did not stop within timeout")
	}
}

// TestAtomicStartupFailure tests that if one server fails to bind, both servers fail
func TestAtomicStartupFailure(t *testing.T) {
	// Create logger for tests
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Use same port for both servers - this should cause atomic failure
	conflictPort := 18003
	enduserConfig := ports.ServerInstanceConfig{
		Port: conflictPort,
		Bind: "::",
	}
	adminConfig := ports.ServerInstanceConfig{
		Port: conflictPort, // Same port - will conflict!
		Bind: "::",
	}

	// Create server instances
	enduserServer := httpAdapter.NewServer("enduser", enduserConfig, logger)
	adminServer := httpAdapter.NewServer("admin", adminConfig, logger)

	// Create manager
	mgr := server.NewManager(enduserServer, adminServer, 5*time.Second, logger)

	// Start servers - should fail atomically
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := mgr.Start(ctx)
	if err == nil {
		t.Error("Expected startup to fail due to port conflict, but it succeeded")
	}

	// Verify error message is descriptive
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}
