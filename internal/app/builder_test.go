package app

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/testutil"
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
		Encryption: ports.EncryptionConfig{
			Memory: &ports.MemoryConfig{
				RawKey: testutil.TestKEKBase64,
			},
		},
		OAuth2AuthServer: ports.OAuth2AuthServerConfig{
			Mode: "proxy",
			Proxy: ports.ProxyModeConfig{
				UpstreamIssuerURI:         "https://auth.example.com",
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				UpstreamTokenEndpoint:     "https://auth.example.com/token",
			},
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

	// Try to build with config + storage + logger but no encryption configured.
	// This specifically exercises the encryption guard at builder.go:130-132 which
	// fires after the nil-dependency checks. The test ensures that reordering or
	// removing that guard would cause a failure here, not silently pass.
	t.Run("missing encryption config", func(t *testing.T) {
		storageAdapter, err := storage.NewAdapter(&ports.StorageConfig{
			Backend: "memory",
			Timeouts: ports.StorageTimeouts{
				Read:  5 * time.Second,
				Write: 5 * time.Second,
			},
		})
		if err != nil {
			t.Fatalf("failed to create storage adapter: %v", err)
		}

		_, err = NewBuilder().
			WithConfig(&ports.Config{
				OAuth2AuthServer: ports.OAuth2AuthServerConfig{
					Mode: "proxy",
					Proxy: ports.ProxyModeConfig{
						UpstreamIssuerURI:         "https://auth.example.com",
						UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
						UpstreamTokenEndpoint:     "https://auth.example.com/token",
					},
				},
			}).
			WithStorage(storageAdapter).
			WithLogger(logger).
			Build()

		if err == nil {
			t.Fatal("expected error when building without encryption configuration")
		}
		if !strings.Contains(err.Error(), "encryption configuration required") {
			t.Errorf("expected error to contain %q, got: %v", "encryption configuration required", err)
		}
	})
}

// TestBuilderTokenExchangeExpectedAudience verifies that the builder correctly reads
// ExpectedAudience from config and passes it to the JWT validator, including the
// default fallback when no value is set.
//
// The test spins up a minimal httptest server that mimics the discovery and JWKS
// endpoints of a real upstream OAuth2 server, so the builder can complete
// JWKS-discovery and JWT-validator wiring without any external dependencies.
func TestBuilderTokenExchangeExpectedAudience(t *testing.T) {
	jweKey := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))

	// Spin up a minimal mock upstream that serves OAuth2 discovery + empty JWKS.
	// We need discovery so the builder can find the jwks_uri, and we need
	// the JWKS endpoint so NewJWKSAdapter succeeds. The key set can be empty
	// because we are only testing wiring, not actual JWT verification here.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/oauth-authorization-server":
			baseURL := "http://" + r.Host
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer":                 baseURL,
				"authorization_endpoint": baseURL + "/oauth/authorize",
				"token_endpoint":         baseURL + "/oauth/token",
				"jwks_uri":               baseURL + "/.well-known/jwks.json",
			})
		case "/.well-known/jwks.json":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"keys": []interface{}{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	// buildAppWithTokenExchange constructs a complete App with token exchange enabled,
	// using the provided config modifier to adjust individual fields.
	buildAppWithTokenExchange := func(t *testing.T, modifyConfig func(*ports.Config)) (*App, error) {
		t.Helper()
		storageAdapter, err := storage.NewAdapter(&ports.StorageConfig{
			Backend: "memory",
			Timeouts: ports.StorageTimeouts{
				Read:  5 * time.Second,
				Write: 5 * time.Second,
			},
		})
		if err != nil {
			t.Fatalf("failed to create storage adapter: %v", err)
		}

		cfg := &ports.Config{
			Log: ports.LogConfig{Level: ports.LogLevelInfo, Format: ports.LogFormatText},
			Server: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port:      8000,
					Bind:      "::1",
					PublicURL: "http://localhost:8000",
					Authentication: ports.AuthenticationConfig{
						Preauth: ports.PreauthConfig{PrincipalHeaderName: "X-Remote-User"},
					},
				},
				Admin: ports.ServerInstanceConfig{
					Port:      14000,
					Bind:      "::1",
					PublicURL: "http://localhost:14000",
					Authentication: ports.AuthenticationConfig{
						Preauth: ports.PreauthConfig{PrincipalHeaderName: "X-Remote-User"},
					},
				},
				Shutdown: ports.ShutdownConfig{Timeout: 5 * time.Second},
			},
			Storage: ports.StorageConfig{
				Backend:  "memory",
				Timeouts: ports.StorageTimeouts{Read: 5 * time.Second, Write: 5 * time.Second},
			},
			ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{JWESigningKey: jweKey},
			Encryption:       ports.EncryptionConfig{Memory: &ports.MemoryConfig{RawKey: testutil.TestKEKBase64}},
			OAuth2AuthServer: ports.OAuth2AuthServerConfig{
				Mode: "proxy",
				Proxy: ports.ProxyModeConfig{
					UpstreamIssuerURI:         upstream.URL,
					UpstreamAuthorizeEndpoint: upstream.URL + "/oauth/authorize",
					UpstreamTokenEndpoint:     upstream.URL + "/oauth/token",
					UpstreamTimeoutSeconds:    5,
				},
			},
			TokenExchange: ports.TokenExchangeConfig{
				ClaimExtraction: ports.ClaimExtractionConfig{
					PrincipalExpression: "subject_token.sub",
					AgentIDExpression:   "subject_token.azp",
				},
				Authorization: ports.AuthorizationConfig{
					Type: "cel",
					CEL: ports.CELAuthorizationConfig{
						Expression:        "true",
						EvaluationTimeout: 100 * time.Millisecond,
					},
				},
			},
			Security: ports.SecurityConfig{SkipThirdpartyHTTPSValidation: true},
		}
		modifyConfig(cfg)

		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

		return NewBuilder().
			WithConfig(cfg).
			WithStorage(storageAdapter).
			WithLogger(logger).
			Build()
	}

	t.Run("empty ExpectedAudience falls back to default and builds successfully", func(t *testing.T) {
		app, err := buildAppWithTokenExchange(t, func(cfg *ports.Config) {
			cfg.TokenExchange.ExpectedAudience = "" // use default "token-exchange-broker"
		})
		if err != nil {
			t.Fatalf("Build() with empty ExpectedAudience failed: %v", err)
		}
		if app.TokenExchangeService == nil {
			t.Error("expected TokenExchangeService to be set when token exchange is configured")
		}
	})

	t.Run("custom ExpectedAudience is accepted and service is wired", func(t *testing.T) {
		app, err := buildAppWithTokenExchange(t, func(cfg *ports.Config) {
			cfg.TokenExchange.ExpectedAudience = "my-gateway"
		})
		if err != nil {
			t.Fatalf("Build() with custom ExpectedAudience failed: %v", err)
		}
		if app.TokenExchangeService == nil {
			t.Error("expected TokenExchangeService to be set when token exchange is configured")
		}
	})
}

// T056: Builder produces the correct strategy set for each OAuth2 server mode.
// Proxy → JWKS handler nil; Local/Hybrid → JWKS handler non-nil (signing keys served).
func TestBuilder_ModeStrategyWiring(t *testing.T) {
	baseConfig := func(jweKey string) *ports.Config {
		return &ports.Config{
			Log: ports.LogConfig{Level: ports.LogLevelInfo, Format: ports.LogFormatText},
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
				Shutdown: ports.ShutdownConfig{Timeout: 5 * time.Second},
			},
			Storage: ports.StorageConfig{
				Backend:  "memory",
				Timeouts: ports.StorageTimeouts{Read: 5 * time.Second, Write: 5 * time.Second},
			},
			ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{JWESigningKey: jweKey},
			Encryption:       ports.EncryptionConfig{Memory: &ports.MemoryConfig{RawKey: testutil.TestKEKBase64}},
		}
	}

	newStorage := func(t *testing.T) *storage.Adapter {
		t.Helper()
		a, err := storage.NewAdapter(&ports.StorageConfig{
			Backend:  "memory",
			Timeouts: ports.StorageTimeouts{Read: 5 * time.Second, Write: 5 * time.Second},
		})
		if err != nil {
			t.Fatalf("storage.NewAdapter: %v", err)
		}
		return a
	}

	jweKey := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("proxy mode — JWKS handler is nil", func(t *testing.T) {
		cfg := baseConfig(jweKey)
		cfg.OAuth2AuthServer = ports.OAuth2AuthServerConfig{
			Mode: "proxy",
			Proxy: ports.ProxyModeConfig{
				UpstreamIssuerURI:         "https://issuer.example.com",
				UpstreamAuthorizeEndpoint: "https://issuer.example.com/authorize",
				UpstreamTokenEndpoint:     "https://issuer.example.com/token",
			},
		}
		app, err := NewBuilder().WithConfig(cfg).WithStorage(newStorage(t)).WithLogger(logger).Build()
		if err != nil {
			t.Fatalf("Build() in proxy mode failed: %v", err)
		}
		if app.EnduserHandlers.JWKS != nil {
			t.Error("proxy mode must not wire a JWKS handler")
		}
	})

	t.Run("local mode — JWKS handler is non-nil", func(t *testing.T) {
		cfg := baseConfig(jweKey)
		cfg.OAuth2AuthServer = ports.OAuth2AuthServerConfig{
			Mode:  "local",
			Local: ports.LocalModeConfig{TokenTTL: time.Hour},
		}
		app, err := NewBuilder().WithConfig(cfg).WithStorage(newStorage(t)).WithLogger(logger).Build()
		if err != nil {
			t.Fatalf("Build() in local mode failed: %v", err)
		}
		if app.EnduserHandlers.JWKS == nil {
			t.Error("local mode must wire a JWKS handler")
		}
	})

	t.Run("local mode with token-exchange defaults — Build() succeeds without upstream discovery", func(t *testing.T) {
		// Mirrors the chart's packaged defaults: local mode + non-empty token exchange CEL expressions
		// but no proxy config. The builder must not attempt OAuth2 endpoint discovery in this case.
		cfg := baseConfig(jweKey)
		cfg.OAuth2AuthServer = ports.OAuth2AuthServerConfig{
			Mode:  "local",
			Local: ports.LocalModeConfig{TokenTTL: time.Hour},
		}
		cfg.TokenExchange = ports.TokenExchangeConfig{
			ClaimExtraction: ports.ClaimExtractionConfig{
				PrincipalExpression: "subject_token.sub",
				AgentIDExpression:   "resolveAgentIdByClientId(subject_token.azp)",
			},
			Authorization: ports.AuthorizationConfig{
				Type: "cel",
				CEL:  ports.CELAuthorizationConfig{Expression: "true"},
			},
		}
		_, err := NewBuilder().WithConfig(cfg).WithStorage(newStorage(t)).WithLogger(logger).Build()
		if err != nil {
			t.Fatalf("Build() in local mode with token-exchange defaults failed: %v", err)
		}
	})

	t.Run("hybrid mode — JWKS handler is non-nil", func(t *testing.T) {
		cfg := baseConfig(jweKey)
		cfg.OAuth2AuthServer = ports.OAuth2AuthServerConfig{
			Mode: "hybrid",
			Proxy: ports.ProxyModeConfig{
				UpstreamIssuerURI:         "https://issuer.example.com",
				UpstreamAuthorizeEndpoint: "https://issuer.example.com/authorize",
				UpstreamTokenEndpoint:     "https://issuer.example.com/token",
			},
			Local: ports.LocalModeConfig{TokenTTL: time.Hour},
		}
		app, err := NewBuilder().WithConfig(cfg).WithStorage(newStorage(t)).WithLogger(logger).Build()
		if err != nil {
			t.Fatalf("Build() in hybrid mode failed: %v", err)
		}
		if app.EnduserHandlers.JWKS == nil {
			t.Error("hybrid mode must wire a JWKS handler")
		}
	})
}

// TestBuilder_MissingOAuth2AuthServerConfig verifies that Build() fails when the
// oauth2_authorization_server block is absent. Mode is mandatory — no default exists.
func TestBuilder_MissingOAuth2AuthServerConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	adapter, err := storage.NewAdapter(&ports.StorageConfig{
		Backend:  "memory",
		Timeouts: ports.StorageTimeouts{Read: 5 * time.Second, Write: 5 * time.Second},
	})
	if err != nil {
		t.Fatalf("failed to create storage adapter: %v", err)
	}

	jweKey := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	cfg := &ports.Config{
		Log: ports.LogConfig{Level: ports.LogLevelInfo, Format: ports.LogFormatText},
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
			Shutdown: ports.ShutdownConfig{Timeout: 5 * time.Second},
		},
		Storage: ports.StorageConfig{
			Backend:  "memory",
			Timeouts: ports.StorageTimeouts{Read: 5 * time.Second, Write: 5 * time.Second},
		},
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{JWESigningKey: jweKey},
		Encryption:       ports.EncryptionConfig{Memory: &ports.MemoryConfig{RawKey: testutil.TestKEKBase64}},
		// OAuth2AuthServer intentionally absent — must cause startup failure
	}

	_, err = NewBuilder().WithConfig(cfg).WithStorage(adapter).WithLogger(logger).Build()
	if err == nil {
		t.Fatal("Build() with absent OAuth2AuthServerConfig must fail")
	}
	if !strings.Contains(err.Error(), "mode") {
		t.Errorf("expected error to mention 'mode', got: %v", err)
	}
}

func TestModeStrategyFor_PanicsOnUnknownMode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected modeStrategyFor to panic on unknown mode, but it did not")
		}
	}()
	modeStrategyFor("bogus")
}

func TestResolveUpstreamTimeout(t *testing.T) {
	tests := []struct {
		configured int
		want       time.Duration
	}{
		{0, 30 * time.Second},  // local mode sentinel → application default
		{5, 5 * time.Second},   // explicit proxy config
		{60, 60 * time.Second}, // custom timeout
	}
	for _, tt := range tests {
		got := resolveUpstreamTimeout(tt.configured)
		if got != tt.want {
			t.Errorf("resolveUpstreamTimeout(%d) = %v, want %v", tt.configured, got, tt.want)
		}
	}
}
