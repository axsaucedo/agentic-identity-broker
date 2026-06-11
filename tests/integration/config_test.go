package integration

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/spf13/cobra"
)

// testJWESigningKey is a valid test JWE key (32 bytes base64 encoded)
var testJWESigningKey = base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))

// testEncryptionKey is a valid test encryption key (32 bytes base64 encoded)
var testEncryptionKey = base64.StdEncoding.EncodeToString([]byte("abcdef0123456789abcdef0123456789"))

// TestConfigurationPrecedence tests that configuration sources are applied in correct precedence order.
// Order: CLI flags > Environment variables > YAML > Defaults
func TestConfigurationPrecedence(t *testing.T) {
	tests := []struct {
		name          string
		yamlContent   string
		envVars       map[string]string
		cliFlags      map[string]interface{}
		expectedPort  int
		expectedBind  string
		expectedLevel string
	}{
		{
			name: "defaults only",
			yamlContent: fmt.Sprintf(`
storage:
  backend: memory
  timeouts:
    read: 5s
    write: 10s
encryption:
  key: %s
`, testEncryptionKey),
			expectedPort:  8000,
			expectedBind:  "::",
			expectedLevel: "info",
		},
		{
			name: "yaml overrides defaults",
			yamlContent: fmt.Sprintf(`
log:
  level: debug
  format: text
server:
  enduser:
    port: 9000
    bind: "127.0.0.1"
    public_url: http://localhost:9000
  admin:
    port: 14000
    bind: "::"
    public_url: http://localhost:14000
  shutdown:
    timeout: 30s
storage:
  backend: memory
  timeouts:
    read: 5s
    write: 10s
encryption:
  key: %s
`, testEncryptionKey),
			expectedPort:  9000,
			expectedBind:  "127.0.0.1",
			expectedLevel: "debug",
		},
		{
			name: "env overrides yaml",
			yamlContent: fmt.Sprintf(`
log:
  level: debug
  format: text
server:
  enduser:
    port: 9000
    bind: "127.0.0.1"
    public_url: http://localhost:9000
  admin:
    port: 14000
    bind: "::"
    public_url: http://localhost:14000
  shutdown:
    timeout: 30s
storage:
  backend: memory
  timeouts:
    read: 5s
    write: 10s
encryption:
  key: %s
`, testEncryptionKey),
			envVars: map[string]string{
				"IDENTITY_BROKER_SERVER_ENDUSER_PORT": "9500",
				"IDENTITY_BROKER_LOG_LEVEL":           "warn",
			},
			expectedPort:  9500,
			expectedBind:  "127.0.0.1", // From YAML
			expectedLevel: "warn",
		},
		{
			name: "cli overrides all",
			yamlContent: fmt.Sprintf(`
log:
  level: debug
  format: text
server:
  enduser:
    port: 9000
    bind: "127.0.0.1"
    public_url: http://localhost:9000
  admin:
    port: 14000
    bind: "::"
    public_url: http://localhost:14000
  shutdown:
    timeout: 30s
storage:
  backend: memory
  timeouts:
    read: 5s
    write: 10s
encryption:
  key: %s
`, testEncryptionKey),
			envVars: map[string]string{
				"IDENTITY_BROKER_SERVER_ENDUSER_PORT": "9500",
				"IDENTITY_BROKER_LOG_LEVEL":           "warn",
			},
			cliFlags: map[string]interface{}{
				"server.enduser.port": 10000,
				"log-level":           "error",
			},
			expectedPort:  10000,
			expectedBind:  "127.0.0.1", // From YAML
			expectedLevel: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test files
			tmpDir := t.TempDir()

			// Clear any existing environment variables to ensure clean test state
			// This prevents inherited env vars from interfering with test
			t.Setenv("IDENTITY_BROKER_LOG_LEVEL", "")
			t.Setenv("IDENTITY_BROKER_LOG_FORMAT", "")
			t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_PORT", "")
			t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_BIND", "")
			t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_PORT", "")
			t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_BIND", "")
			t.Setenv("IDENTITY_BROKER_SERVER_SHUTDOWN_TIMEOUT", "")

			// Set mandatory JWESigningKey, encryption key, principal headers, and oauth2 mode for all tests
			t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", testJWESigningKey)
			t.Setenv("IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY", testEncryptionKey)
			t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
			t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
			t.Setenv("IDENTITY_BROKER_OAUTH2_AUTH_SERVER_MODE", "local")

			// Create YAML config file if content is provided
			var configPath string
			if tt.yamlContent != "" {
				configPath = filepath.Join(tmpDir, "config.yaml")
				if err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			}

			// Set environment variables (after clearing)
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// If we have a config file, set it in environment
			if configPath != "" {
				t.Setenv("IDENTITY_BROKER_CONFIG_PATH", configPath)
			}

			// Create loader
			loader := config.NewLoader()

			// Create cobra command if CLI flags are provided
			if tt.cliFlags != nil {
				cmd := &cobra.Command{
					Use: "test",
				}

				// Add flags
				cmd.Flags().Int("server.enduser.port", 0, "")
				cmd.Flags().String("server.enduser.bind", "", "")
				cmd.Flags().String("log-level", "", "")

				// Set flag values (need to parse flags to mark them as changed)
				for key, value := range tt.cliFlags {
					var err error
					switch key {
					case "server.enduser.port":
						err = cmd.Flags().Set(key, fmt.Sprintf("%d", value.(int)))
					case "log-level":
						err = cmd.Flags().Set(key, value.(string))
					}
					if err != nil {
						t.Fatalf("Failed to set flag %s: %v", key, err)
					}
				}

				loader.SetCommand(cmd)
			}

			// Load configuration
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := loader.GetConfig(ctx)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Verify precedence
			if cfg.Server.EndUser.Port != tt.expectedPort {
				t.Errorf("Expected port %d, got %d", tt.expectedPort, cfg.Server.EndUser.Port)
			}

			if cfg.Server.EndUser.Bind != tt.expectedBind {
				t.Errorf("Expected bind %q, got %q", tt.expectedBind, cfg.Server.EndUser.Bind)
			}

			if string(cfg.Log.Level) != tt.expectedLevel {
				t.Errorf("Expected log level %q, got %q", tt.expectedLevel, cfg.Log.Level)
			}
		})
	}
}

// TestConfigurationFromExamples tests that example configuration files load correctly.
func TestConfigurationFromExamples(t *testing.T) {
	tests := []struct {
		name         string
		configFile   string
		expectedPort int
		expectedBind string
	}{
		{
			name:         "development config",
			configFile:   "../../examples/config/config.development.yaml",
			expectedPort: 3000,
			expectedBind: "127.0.0.1",
		},
		{
			name:         "staging config",
			configFile:   "../../examples/config/config.staging.yaml",
			expectedPort: 8000,
			expectedBind: "::",
		},
		{
			name:         "production config",
			configFile:   "../../examples/config/config.production.yaml",
			expectedPort: 8000,
			expectedBind: "::",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get absolute path to config file
			configPath, err := filepath.Abs(tt.configFile)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			// Set config path in environment
			t.Setenv("IDENTITY_BROKER_CONFIG_PATH", configPath)

			// Set mandatory JWESigningKey only — example configs now carry their own mode
			t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", testJWESigningKey)

			// Set encryption backend environment variables based on config type
			if tt.name == "production config" {
				// Production config expects AWS KMS backend and proxy-mode upstream config
				t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN", "arn:aws:kms:us-east-1:123456789012:key/test-key-id")
				t.Setenv("IDENTITY_BROKER_OAUTH2_AUTH_SERVER_PROXY_UPSTREAM_ISSUER_URI", "https://idp.example.com")
				t.Setenv("IDENTITY_BROKER_OAUTH2_AUTH_SERVER_PROXY_UPSTREAM_AUTHORIZE_ENDPOINT", "https://idp.example.com/oauth2/authorize")
				t.Setenv("IDENTITY_BROKER_OAUTH2_AUTH_SERVER_PROXY_UPSTREAM_TOKEN_ENDPOINT", "https://idp.example.com/oauth2/token")
			} else {
				// Development and staging configs use Memory backend and local mode
				t.Setenv("IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY", testEncryptionKey)
			}

			// Create loader
			loader := config.NewLoader()

			// Load configuration
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := loader.GetConfig(ctx)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Verify configuration
			if cfg.Server.EndUser.Port != tt.expectedPort {
				t.Errorf("Expected port %d, got %d", tt.expectedPort, cfg.Server.EndUser.Port)
			}

			if cfg.Server.EndUser.Bind != tt.expectedBind {
				t.Errorf("Expected bind %q, got %q", tt.expectedBind, cfg.Server.EndUser.Bind)
			}

			if tt.name == "production config" && cfg.OAuth2AuthServer.Proxy.UpstreamTimeout != 30*time.Second {
				t.Errorf("Expected proxy upstream timeout %v, got %v", 30*time.Second, cfg.OAuth2AuthServer.Proxy.UpstreamTimeout)
			}

			// Verify sources are tracked
			sources := loader.GetSources()
			if len(sources) == 0 {
				t.Error("Expected configuration sources to be tracked")
			}

			// Verify YAML source is present
			hasYAML := false
			for _, src := range sources {
				if src.Type == ports.SourceTypeYAML {
					hasYAML = true
					break
				}
			}
			if !hasYAML {
				t.Error("Expected YAML source to be tracked")
			}
		})
	}
}

func TestConfigurationParsesProxyUpstreamTimeoutFromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(`oauth2_authorization_server:
  mode: proxy
  proxy:
    upstream_issuer_uri: https://idp.example.com
    upstream_authorize_endpoint: https://idp.example.com/oauth2/authorize
    upstream_token_endpoint: https://idp.example.com/oauth2/token
    upstream_timeout: 500ms
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	t.Setenv("IDENTITY_BROKER_CONFIG_PATH", configPath)
	t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", testJWESigningKey)
	t.Setenv("IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY", testEncryptionKey)
	t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
	t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")

	loader := config.NewLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := loader.GetConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.OAuth2AuthServer.Proxy.UpstreamTimeout != 500*time.Millisecond {
		t.Fatalf("Expected proxy upstream timeout %v, got %v", 500*time.Millisecond, cfg.OAuth2AuthServer.Proxy.UpstreamTimeout)
	}
}
