package app

import (
	"context"
	"encoding/base64"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// TestBuilderMinimalConfiguration validates that the builder can construct an application
// with in-memory storage and the minimum required configuration.
//
// This test documents the minimal setup needed to start the application:
// - Configuration with logging and server settings
// - In-memory storage adapter
// - Logger instance
//
// The test proves that the builder pattern works for dependency injection
// and that the application can be wired up with in-memory storage for testing.
func TestBuilderMinimalConfiguration(t *testing.T) {
	// Generate a valid JWE signing key for testing (32 bytes = 256 bits)
	jweKey := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))

	// Create minimal configuration
	cfg := &ports.Config{
		Log: ports.LogConfig{
			Level:  ports.LogLevelInfo,
			Format: ports.LogFormatText,
		},
		Server: ports.ServerConfig{
			EndUser: ports.ServerInstanceConfig{
				Port:      8000,
				Bind:      "::1",
				PublicURL: "http://localhost:8000",
				Authentication: ports.AuthenticationConfig{
					Preauth: ports.PreauthConfig{
						PrincipalHeaderName: "X-Remote-User",
					},
				},
			},
			Admin: ports.ServerInstanceConfig{
				Port:      14000,
				Bind:      "::1",
				PublicURL: "http://localhost:14000",
				Authentication: ports.AuthenticationConfig{
					Preauth: ports.PreauthConfig{
						PrincipalHeaderName: "X-Remote-User",
					},
				},
			},
			Shutdown: ports.ShutdownConfig{
				Timeout: 30 * time.Second,
			},
		},
		Storage: ports.StorageConfig{
			Backend: "memory",
			Timeouts: ports.StorageTimeouts{
				Read:  5 * time.Second,
				Write: 5 * time.Second,
			},
		},
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{
			JWESigningKey: jweKey,
		},
	}

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create in-memory storage
	adapter, err := storage.NewAdapter(&ports.StorageConfig{
		Backend: "memory",
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Second,
			Write: 5 * time.Second,
		},
	})
	if err != nil {
		t.Fatalf("failed to create storage adapter: %v", err)
	}

	// Build application with minimal configuration
	app, err := NewBuilder().
		WithConfig(cfg).
		WithStorage(adapter).
		WithLogger(logger).
		Build()

	if err != nil {
		t.Fatalf("failed to build application: %v", err)
	}

	// Verify application is wired correctly
	if app == nil {
		t.Fatal("expected non-nil application")
	}

	// Verify required fields are set
	if app.Config == nil {
		t.Error("expected Config to be set")
	}

	if app.Storage == nil {
		t.Error("expected Storage to be set")
	}

	if app.Logger == nil {
		t.Error("expected Logger to be set")
	}

	// Verify handlers are created
	if app.AdminHandlers == nil {
		t.Error("expected AdminHandlers to be set")
	}

	if app.EnduserHandlers == nil {
		t.Error("expected EnduserHandlers to be set")
	}

	// Verify basic handler structure in minimal config
	// (Optional services like OAuth2 and OAuth2Sessions are not configured)
	if app.EnduserHandlers.UserInfo == nil {
		t.Error("expected UserInfo handler to be created")
	}
}

// TestBuilderMissingRequiredDependency validates that the builder rejects invalid configurations.
func TestBuilderMissingRequiredDependency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Try to build without configuration
	_, err := NewBuilder().
		WithStorage(nil).
		WithLogger(logger).
		Build()

	if err == nil {
		t.Error("expected error when building without configuration")
	}

	// Try to build without storage
	cfg := &ports.Config{}
	_, err = NewBuilder().
		WithConfig(cfg).
		WithLogger(logger).
		Build()

	if err == nil {
		t.Error("expected error when building without storage")
	}

	// Try to build without logger
	_, err = NewBuilder().
		WithConfig(cfg).
		WithStorage(nil).
		Build()

	if err == nil {
		t.Error("expected error when building without logger")
	}
}

// TestBuilderWithCustomEncryption validates that the builder accepts
// injected encryption implementations for production use.
func TestBuilderWithCustomEncryption(t *testing.T) {
	// Generate a valid JWE signing key for testing
	jweKey := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))

	// Minimal config setup
	cfg := &ports.Config{
		Log: ports.LogConfig{
			Level:  ports.LogLevelInfo,
			Format: ports.LogFormatText,
		},
		Server: ports.ServerConfig{
			EndUser: ports.ServerInstanceConfig{
				Port: 8000, Bind: "::1", PublicURL: "http://localhost:8000",
				Authentication: ports.AuthenticationConfig{
					Preauth: ports.PreauthConfig{PrincipalHeaderName: "X-Remote-User"},
				},
			},
			Admin: ports.ServerInstanceConfig{
				Port: 14000, Bind: "::1", PublicURL: "http://localhost:14000",
				Authentication: ports.AuthenticationConfig{
					Preauth: ports.PreauthConfig{PrincipalHeaderName: "X-Remote-User"},
				},
			},
			Shutdown: ports.ShutdownConfig{Timeout: 30 * time.Second},
		},
		Storage: ports.StorageConfig{
			Backend: "memory",
			Timeouts: ports.StorageTimeouts{
				Read:  5 * time.Second,
				Write: 5 * time.Second,
			},
		},
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{
			JWESigningKey: jweKey,
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	adapter, err := storage.NewAdapter(&ports.StorageConfig{
		Backend: "memory",
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Second,
			Write: 5 * time.Second,
		},
	})
	if err != nil {
		t.Fatalf("failed to create storage adapter: %v", err)
	}

	// Mock encryption implementation
	mockEncryptor := &mockEncryption{}

	// Build with custom encryption
	app, err := NewBuilder().
		WithConfig(cfg).
		WithStorage(adapter).
		WithLogger(logger).
		WithEncryption(mockEncryptor).
		Build()

	if err != nil {
		t.Fatalf("failed to build with custom encryption: %v", err)
	}

	if app == nil {
		t.Fatal("expected non-nil application")
	}

	// Verify builder accepted the encryption implementation
	// (OAuth2SessionService would use it if third-party OAuth2 was configured)
}

// mockEncryption implements ports.EncryptionPort for testing
type mockEncryption struct{}

func (m *mockEncryption) Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error) {
	return plaintext, nil
}

func (m *mockEncryption) Decrypt(ctx context.Context, ciphertext []byte, encryptionContext map[string]string) ([]byte, error) {
	return ciphertext, nil
}
