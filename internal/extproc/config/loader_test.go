package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
)

// validConfig returns a minimal valid Config for use in test cases.
func validConfig() *config.Config {
	return &config.Config{
		GRPC: config.GRPCConfig{
			Bind:                 "0.0.0.0",
			Port:                 50051,
			MaxConcurrentStreams: 100,
		},
		OAuth2: config.OAuth2Config{
			TokenEndpoint:       "https://identity-broker.example.com/oauth2/token",
			Issuer:              "https://upstream-oauth2.example.com",
			ClientID:            "extproc-client",
			ClientSecret:        "supersecret",
			ClientAssertionType: "id_token",
			ExchangeTimeout:     5 * time.Second,
			TLS:                 config.TLSConfig{AllowHTTP: false},
		},
		Cache: config.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
		CircuitBreaker: config.CircuitBreakerConfig{
			Enabled:      true,
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}
}

// ---------------------------------------------------------------------------
// Validate: All 10 validation rules (table-driven)
// ---------------------------------------------------------------------------

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*config.Config)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid config passes all rules",
			mutate:  func(_ *config.Config) {},
			wantErr: false,
		},
		// Rule 1: grpc.port
		{
			name:        "rule1: port 0 is invalid",
			mutate:      func(c *config.Config) { c.GRPC.Port = 0 },
			wantErr:     true,
			errContains: "grpc.port",
		},
		{
			name:        "rule1: port 65536 is invalid",
			mutate:      func(c *config.Config) { c.GRPC.Port = 65536 },
			wantErr:     true,
			errContains: "grpc.port",
		},
		{
			name:    "rule1: port 1 is valid",
			mutate:  func(c *config.Config) { c.GRPC.Port = 1 },
			wantErr: false,
		},
		{
			name:    "rule1: port 65535 is valid",
			mutate:  func(c *config.Config) { c.GRPC.Port = 65535 },
			wantErr: false,
		},
		// Rule 2: grpc.bind
		{
			name:        "rule2: empty bind is invalid",
			mutate:      func(c *config.Config) { c.GRPC.Bind = "" },
			wantErr:     true,
			errContains: "grpc.bind",
		},
		{
			name:        "rule2: whitespace-only bind is invalid",
			mutate:      func(c *config.Config) { c.GRPC.Bind = "   " },
			wantErr:     true,
			errContains: "grpc.bind",
		},
		// Rule 3: oauth2.token_endpoint
		{
			name:        "rule3: empty token_endpoint is invalid",
			mutate:      func(c *config.Config) { c.OAuth2.TokenEndpoint = "" },
			wantErr:     true,
			errContains: "oauth2.token_endpoint",
		},
		{
			name:        "rule3: non-URL token_endpoint is invalid",
			mutate:      func(c *config.Config) { c.OAuth2.TokenEndpoint = "not-a-url" },
			wantErr:     true,
			errContains: "oauth2.token_endpoint",
		},
		{
			name:        "rule3: ftp:// scheme is invalid",
			mutate:      func(c *config.Config) { c.OAuth2.TokenEndpoint = "ftp://example.com/token" },
			wantErr:     true,
			errContains: "oauth2.token_endpoint",
		},
		// Rule 4: oauth2.issuer
		{
			name:        "rule4: empty issuer is invalid",
			mutate:      func(c *config.Config) { c.OAuth2.Issuer = "" },
			wantErr:     true,
			errContains: "oauth2.issuer",
		},
		{
			name:        "rule4: non-URL issuer is invalid",
			mutate:      func(c *config.Config) { c.OAuth2.Issuer = "not-a-url" },
			wantErr:     true,
			errContains: "oauth2.issuer",
		},
		// Rule 5: oauth2.client_id
		{
			name:        "rule5: empty client_id is invalid",
			mutate:      func(c *config.Config) { c.OAuth2.ClientID = "" },
			wantErr:     true,
			errContains: "oauth2.client_id",
		},
		// Rule 6: oauth2.client_secret
		{
			name:        "rule6: empty client_secret is invalid",
			mutate:      func(c *config.Config) { c.OAuth2.ClientSecret = "" },
			wantErr:     true,
			errContains: "oauth2.client_secret",
		},
		// Rule 7: cache.default_ttl
		{
			name:        "rule7: zero default_ttl is invalid",
			mutate:      func(c *config.Config) { c.Cache.DefaultTTL = 0 },
			wantErr:     true,
			errContains: "cache.default_ttl",
		},
		{
			name:        "rule7: negative default_ttl is invalid",
			mutate:      func(c *config.Config) { c.Cache.DefaultTTL = -1 * time.Second },
			wantErr:     true,
			errContains: "cache.default_ttl",
		},
		// Rule 8: TLS enforcement
		{
			name: "rule8: http token_endpoint rejected when allow_http is false",
			mutate: func(c *config.Config) {
				c.OAuth2.TokenEndpoint = "http://identity-broker.example.com/oauth2/token"
				c.OAuth2.TLS.AllowHTTP = false
			},
			wantErr:     true,
			errContains: "oauth2.token_endpoint",
		},
		{
			name: "rule8: http issuer rejected when allow_http is false",
			mutate: func(c *config.Config) {
				c.OAuth2.Issuer = "http://upstream-oauth2.example.com"
				c.OAuth2.TLS.AllowHTTP = false
			},
			wantErr:     true,
			errContains: "oauth2.issuer",
		},
		{
			name: "rule8: http endpoints allowed when allow_http is true",
			mutate: func(c *config.Config) {
				c.OAuth2.TokenEndpoint = "http://identity-broker.example.com/oauth2/token"
				c.OAuth2.Issuer = "http://upstream-oauth2.example.com"
				c.OAuth2.TLS.AllowHTTP = true
			},
			wantErr: false,
		},
		// Rule 9: cache.max_ttl
		{
			name:        "rule9: zero max_ttl is invalid",
			mutate:      func(c *config.Config) { c.Cache.MaxTTL = 0 },
			wantErr:     true,
			errContains: "cache.max_ttl",
		},
		// Rule 10: oauth2.exchange_timeout
		{
			name:        "rule10: zero exchange_timeout is invalid",
			mutate:      func(c *config.Config) { c.OAuth2.ExchangeTimeout = 0 },
			wantErr:     true,
			errContains: "oauth2.exchange_timeout",
		},
		{
			name:        "rule10: negative exchange_timeout is invalid",
			mutate:      func(c *config.Config) { c.OAuth2.ExchangeTimeout = -1 * time.Second },
			wantErr:     true,
			errContains: "oauth2.exchange_timeout",
		},
		// Rule 11: log.level enum
		{
			name:        "rule11: invalid log level is rejected",
			mutate:      func(c *config.Config) { c.Log.Level = "verbose" },
			wantErr:     true,
			errContains: "log.level",
		},
		{
			name:    "rule11: valid log levels accepted",
			mutate:  func(c *config.Config) { c.Log.Level = "debug" },
			wantErr: false,
		},
		// Rule 12: log.format enum
		{
			name:        "rule12: invalid log format is rejected",
			mutate:      func(c *config.Config) { c.Log.Format = "logfmt" },
			wantErr:     true,
			errContains: "log.format",
		},
		{
			name:    "rule12: valid log formats accepted",
			mutate:  func(c *config.Config) { c.Log.Format = "json" },
			wantErr: false,
		},
		// Multiple errors collected
		{
			name: "multiple invalid fields returns combined error",
			mutate: func(c *config.Config) {
				c.OAuth2.TokenEndpoint = ""
				c.OAuth2.ClientID = ""
				c.OAuth2.ClientSecret = ""
			},
			wantErr:     true,
			errContains: "configuration validation failed",
		},
		// Rule 14: circuit_breaker.max_failures
		{
			name:        "rule14: zero max_failures is invalid",
			mutate:      func(c *config.Config) { c.CircuitBreaker.MaxFailures = 0 },
			wantErr:     true,
			errContains: "circuit_breaker.max_failures",
		},
		{
			name:        "rule14: negative max_failures is invalid",
			mutate:      func(c *config.Config) { c.CircuitBreaker.MaxFailures = -1 },
			wantErr:     true,
			errContains: "circuit_breaker.max_failures",
		},
		{
			name:    "rule14: max_failures 1 is valid",
			mutate:  func(c *config.Config) { c.CircuitBreaker.MaxFailures = 1 },
			wantErr: false,
		},
		// Rule 15: circuit_breaker.reset_timeout
		{
			name:        "rule15: zero reset_timeout is invalid",
			mutate:      func(c *config.Config) { c.CircuitBreaker.ResetTimeout = 0 },
			wantErr:     true,
			errContains: "circuit_breaker.reset_timeout",
		},
		{
			name:        "rule15: negative reset_timeout is invalid",
			mutate:      func(c *config.Config) { c.CircuitBreaker.ResetTimeout = -1 * time.Second },
			wantErr:     true,
			errContains: "circuit_breaker.reset_timeout",
		},
		// Disabled circuit breaker: rules 14-15 are skipped
		{
			name: "disabled circuit breaker: zero max_failures is valid",
			mutate: func(c *config.Config) {
				c.CircuitBreaker.Enabled = false
				c.CircuitBreaker.MaxFailures = 0
			},
			wantErr: false,
		},
		{
			name: "disabled circuit breaker: zero reset_timeout is valid",
			mutate: func(c *config.Config) {
				c.CircuitBreaker.Enabled = false
				c.CircuitBreaker.ResetTimeout = 0
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.mutate(cfg)

			err := config.Validate(cfg)

			if tt.wantErr {
				require.Error(t, err, "expected validation error")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains,
						"error message should reference the invalid field")
				}
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LoadFromViper: defaults, env vars, ${VAR} expansion
// ---------------------------------------------------------------------------

func TestLoadFromViper_Defaults(t *testing.T) {
	v := viper.New()

	// Set only required fields via Viper directly (simulate env vars)
	v.Set("oauth2.token_endpoint", "https://idp.example.com/oauth2/token")
	v.Set("oauth2.issuer", "https://idp.example.com")
	v.Set("oauth2.client_id", "test-client")
	v.Set("oauth2.client_secret", "test-secret")

	cfg, err := config.LoadFromViper(v)
	require.NoError(t, err)

	// Verify defaults are applied
	assert.Equal(t, "0.0.0.0", cfg.GRPC.Bind)
	assert.Equal(t, 50051, cfg.GRPC.Port)
	assert.Equal(t, 100, cfg.GRPC.MaxConcurrentStreams)
	assert.Equal(t, 5*time.Second, cfg.OAuth2.ExchangeTimeout)
	assert.Equal(t, 5*time.Minute, cfg.Cache.DefaultTTL)
	assert.Equal(t, 1*time.Hour, cfg.Cache.MaxTTL)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "text", cfg.Log.Format)
	assert.False(t, cfg.OAuth2.TLS.InsecureSkipVerify)
	assert.False(t, cfg.OAuth2.TLS.AllowHTTP)
	assert.Equal(t, "", cfg.OAuth2.TLS.CaBundlePath)
	assert.Equal(t, 5, cfg.CircuitBreaker.MaxFailures)
	assert.Equal(t, 30*time.Second, cfg.CircuitBreaker.ResetTimeout)
	assert.True(t, cfg.CircuitBreaker.Enabled, "circuit breaker should be enabled by default")
}

func TestLoadFromViper_EnvVarExpansion(t *testing.T) {
	// Set an environment variable to be expanded
	t.Setenv("TEST_EXTPROC_SECRET", "expanded-secret-value")

	v := viper.New()
	v.Set("oauth2.token_endpoint", "https://idp.example.com/oauth2/token")
	v.Set("oauth2.issuer", "https://idp.example.com")
	v.Set("oauth2.client_id", "test-client")
	v.Set("oauth2.client_secret", "${TEST_EXTPROC_SECRET}")

	cfg, err := config.LoadFromViper(v)
	require.NoError(t, err)

	assert.Equal(t, "expanded-secret-value", cfg.OAuth2.ClientSecret,
		"${VAR} notation should be expanded for client_secret")
}

func TestLoadFromViper_EnvVarExpansion_LogFields(t *testing.T) {
	t.Setenv("TEST_LOG_LEVEL", "debug")
	t.Setenv("TEST_LOG_FORMAT", "json")

	v := viper.New()
	v.Set("oauth2.token_endpoint", "https://idp.example.com/oauth2/token")
	v.Set("oauth2.issuer", "https://idp.example.com")
	v.Set("oauth2.client_id", "test-client")
	v.Set("oauth2.client_secret", "test-secret")
	v.Set("log.level", "${TEST_LOG_LEVEL}")
	v.Set("log.format", "${TEST_LOG_FORMAT}")

	cfg, err := config.LoadFromViper(v)
	require.NoError(t, err)

	assert.Equal(t, "debug", cfg.Log.Level, "${VAR} notation should be expanded for log.level")
	assert.Equal(t, "json", cfg.Log.Format, "${VAR} notation should be expanded for log.format")
}

func TestLoadFromViper_EnvVarOverridesDefault(t *testing.T) {
	// Set EXTPROC_GRPC_PORT env var — t.Setenv auto-restores after test
	t.Setenv("EXTPROC_GRPC_PORT", "9090")

	v := viper.New()
	v.SetEnvPrefix("EXTPROC")
	v.SetEnvKeyReplacer(replaceDotsWithUnderscores())
	v.AutomaticEnv()

	v.Set("oauth2.token_endpoint", "https://idp.example.com/oauth2/token")
	v.Set("oauth2.issuer", "https://idp.example.com")
	v.Set("oauth2.client_id", "test-client")
	v.Set("oauth2.client_secret", "test-secret")

	cfg, err := config.LoadFromViper(v)
	require.NoError(t, err)

	assert.Equal(t, 9090, cfg.GRPC.Port,
		"EXTPROC_GRPC_PORT env var should override the default port")
}

func TestLoadFromViper_MissingRequiredField_ReturnsError(t *testing.T) {
	v := viper.New()
	// Do not set token_endpoint — should fail validation

	_, err := config.LoadFromViper(v)
	require.Error(t, err, "missing required field should return error")
	assert.Contains(t, err.Error(), "oauth2.token_endpoint",
		"error should mention the missing field")
}

func TestLoadFromViper_InvalidConfigFile_ReturnsError(t *testing.T) {
	v := viper.New()
	v.Set("config_path", "/nonexistent/path/config.yaml")

	_, err := config.LoadFromViper(v)
	require.Error(t, err, "nonexistent config file should return error")
}

// ---------------------------------------------------------------------------
// Validate: additional edge cases
// ---------------------------------------------------------------------------

func TestValidate_AllRequiredFieldsMissing(t *testing.T) {
	cfg := &config.Config{}
	err := config.Validate(cfg)
	require.Error(t, err)
	// Error should mention token_endpoint specifically (used by E2E test)
	assert.Contains(t, err.Error(), "token_endpoint")
}

func TestValidate_OptionalClientCredentialsEndpointNotRequired(t *testing.T) {
	cfg := validConfig()
	cfg.OAuth2.ClientCredentialsEndpoint = "" // optional field
	err := config.Validate(cfg)
	assert.NoError(t, err, "client_credentials_endpoint is optional")
}

// ---------------------------------------------------------------------------
// LoadWithCommand: CLI flag precedence tests
// ---------------------------------------------------------------------------

// newTestCommand creates a minimal Cobra command registered with the same flags
// as the extproc-token-exchange binary via RegisterFlags, ensuring tests always
// stay in sync with the production flag set.
func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	config.RegisterFlags(cmd)
	return cmd
}

func TestLoadWithCommand_CLIFlagOverridesEnvVar(t *testing.T) {
	// Set an env var that would normally take effect.
	t.Setenv("EXTPROC_GRPC_PORT", "9090")
	t.Setenv("EXTPROC_OAUTH2_TOKEN_ENDPOINT", "https://env.example.com/token")
	t.Setenv("EXTPROC_OAUTH2_ISSUER", "https://env.example.com")
	t.Setenv("EXTPROC_OAUTH2_CLIENT_ID", "env-client")
	t.Setenv("EXTPROC_OAUTH2_CLIENT_SECRET", "env-secret")

	cmd := newTestCommand()
	// Simulate the user passing --grpc.port on the CLI.
	require.NoError(t, cmd.Flags().Set("grpc.port", "7777"))

	cfg, err := config.LoadWithCommand(cmd)
	require.NoError(t, err)

	// CLI flag wins over env var.
	assert.Equal(t, 7777, cfg.GRPC.Port,
		"CLI --grpc.port should override EXTPROC_GRPC_PORT env var")
	// Other env vars are unaffected.
	assert.Equal(t, "https://env.example.com/token", cfg.OAuth2.TokenEndpoint)
}

func TestLoadWithCommand_UnsetFlagsDoNotShadowEnvVars(t *testing.T) {
	// Set all required fields via env vars.
	t.Setenv("EXTPROC_OAUTH2_TOKEN_ENDPOINT", "https://env.example.com/token")
	t.Setenv("EXTPROC_OAUTH2_ISSUER", "https://env.example.com")
	t.Setenv("EXTPROC_OAUTH2_CLIENT_ID", "env-client")
	t.Setenv("EXTPROC_OAUTH2_CLIENT_SECRET", "env-secret")
	t.Setenv("EXTPROC_GRPC_PORT", "8888")

	// Create a command but do NOT set any flags explicitly.
	cmd := newTestCommand()

	cfg, err := config.LoadWithCommand(cmd)
	require.NoError(t, err)

	// Env var should still win over the Cobra default (0).
	assert.Equal(t, 8888, cfg.GRPC.Port,
		"unset CLI flag must not shadow EXTPROC_GRPC_PORT env var")
}

func TestLoadWithCommand_CLIFlagOverridesDuration(t *testing.T) {
	t.Setenv("EXTPROC_OAUTH2_TOKEN_ENDPOINT", "https://env.example.com/token")
	t.Setenv("EXTPROC_OAUTH2_ISSUER", "https://env.example.com")
	t.Setenv("EXTPROC_OAUTH2_CLIENT_ID", "env-client")
	t.Setenv("EXTPROC_OAUTH2_CLIENT_SECRET", "env-secret")

	cmd := newTestCommand()
	require.NoError(t, cmd.Flags().Set("cache.default_ttl", "30m"))
	require.NoError(t, cmd.Flags().Set("cache.max_ttl", "2h"))

	cfg, err := config.LoadWithCommand(cmd)
	require.NoError(t, err)

	assert.Equal(t, 30*time.Minute, cfg.Cache.DefaultTTL,
		"CLI --cache.default_ttl should override the default")
	assert.Equal(t, 2*time.Hour, cfg.Cache.MaxTTL,
		"CLI --cache.max_ttl should override the default")
}

func TestLoadWithCommand_LogLevelAndFormatFromCLI(t *testing.T) {
	t.Setenv("EXTPROC_OAUTH2_TOKEN_ENDPOINT", "https://env.example.com/token")
	t.Setenv("EXTPROC_OAUTH2_ISSUER", "https://env.example.com")
	t.Setenv("EXTPROC_OAUTH2_CLIENT_ID", "env-client")
	t.Setenv("EXTPROC_OAUTH2_CLIENT_SECRET", "env-secret")

	cmd := newTestCommand()
	require.NoError(t, cmd.Flags().Set("log.level", "debug"))
	require.NoError(t, cmd.Flags().Set("log.format", "json"))

	cfg, err := config.LoadWithCommand(cmd)
	require.NoError(t, err)

	assert.Equal(t, "debug", cfg.Log.Level, "CLI --log.level should take effect")
	assert.Equal(t, "json", cfg.Log.Format, "CLI --log.format should take effect")
}

// replaceDotsWithUnderscores returns a string replacer for Viper key mapping.
// Used in tests that set up their own Viper instance with env prefix.
func replaceDotsWithUnderscores() *strings.Replacer {
	return strings.NewReplacer(".", "_")
}
