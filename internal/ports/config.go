// Package ports defines interfaces for hexagonal architecture boundaries.
// Configuration loading is a driven adapter; domain logic depends on this interface.
package ports

import (
	"context"
	"time"
)

// ConfigPort defines the interface for accessing application configuration.
// This is a hexagonal architecture port - domain logic depends on this interface,
// not concrete implementations.
//
// Implementation: internal/config/loader.go (adapter using Viper/Cobra/godotenv)
type ConfigPort interface {
	// GetConfig returns the fully loaded and validated configuration.
	// ctx allows timeout and cancellation control during loading.
	// Returns error if configuration is invalid or cannot be loaded.
	// Called once at application startup (FR-007).
	GetConfig(ctx context.Context) (*Config, error)

	// GetSources returns metadata about all configuration sources used.
	// Used for audit logging (SR-004) and startup summary (FR-010).
	// This is a pure getter, no I/O, so context not needed.
	GetSources() []ConfigSource

	// Reload reloads configuration from all sources.
	// ctx allows timeout and cancellation control during reload.
	// NOT IMPLEMENTED in initial scope (no hot-reloading).
	// Included for future extensibility.
	Reload(ctx context.Context) error
}

// Config represents the complete application configuration schema.
type Config struct {
	Log              LogConfig              `mapstructure:"log" validate:"required"`
	Server           ServerConfig           `mapstructure:"server" validate:"required"`
	Storage          StorageConfig          `mapstructure:"storage" validate:"required"`
	ThirdPartyOAuth2 ThirdPartyOAuth2Config `mapstructure:"third_party_oauth2"`
	OAuth2AuthServer OAuth2AuthServerConfig `mapstructure:"oauth2_authorization_server"`
	Security         SecurityConfig         `mapstructure:"security"`
}

// ServerConfig contains configuration for both HTTP servers.
type ServerConfig struct {
	EndUser  ServerInstanceConfig `mapstructure:"enduser" validate:"required"`
	Admin    ServerInstanceConfig `mapstructure:"admin" validate:"required"`
	Shutdown ShutdownConfig       `mapstructure:"shutdown" validate:"required"`
}

// ServerInstanceConfig contains configuration for a single HTTP server instance.
type ServerInstanceConfig struct {
	Port           int                  `mapstructure:"port" validate:"required,min=1,max=65535"`
	Bind           string               `mapstructure:"bind" validate:"required"`
	PublicURL      string               `mapstructure:"public_url" validate:"required_if=Port 8000,http_url"`
	Authentication AuthenticationConfig `mapstructure:"authentication"`
}

// AuthenticationConfig holds authentication configuration for a server.
type AuthenticationConfig struct {
	// Preauth holds configuration for pre-authentication (reverse proxy) mode
	Preauth PreauthConfig `mapstructure:"preauth"`

	// Future: JWT configuration can be added here without breaking changes
	// JWT JWTConfig `mapstructure:"jwt"`
}

// PreauthConfig holds configuration for reverse proxy pre-authentication.
type PreauthConfig struct {
	// PrincipalHeaderName is the HTTP header from which principals are extracted
	// This header is set by a trusted reverse proxy after authentication.
	// Example values: "X-Remote-User", "X-Authenticated-User", "Remote-User"
	// Default: "X-Remote-User"
	PrincipalHeaderName string `mapstructure:"principal_header_name" validate:"required,min=1"`
}

// ShutdownConfig contains graceful shutdown settings.
type ShutdownConfig struct {
	Timeout time.Duration `mapstructure:"timeout" validate:"required"`
}

// DefaultServerConfig returns default server configuration values.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		EndUser: ServerInstanceConfig{
			Port:      8000,
			Bind:      "::",                    // Dual-stack (IPv6 with IPv4 fallback)
			PublicURL: "http://localhost:8000", // Default for local development
			Authentication: AuthenticationConfig{
				Preauth: PreauthConfig{
					PrincipalHeaderName: "X-Remote-User",
				},
			},
		},
		Admin: ServerInstanceConfig{
			Port:      14000,
			Bind:      "::",
			PublicURL: "http://localhost:14000", // Default for local development
			Authentication: AuthenticationConfig{
				Preauth: PreauthConfig{
					PrincipalHeaderName: "X-Remote-User",
				},
			},
		},
		Shutdown: ShutdownConfig{
			Timeout: 30 * time.Second,
		},
	}
}

// LogConfig contains logging-related configuration.
type LogConfig struct {
	Level  LogLevel  `mapstructure:"level" validate:"required"`
	Format LogFormat `mapstructure:"format" validate:"required"`
}

// LogLevel is an enumeration of valid log levels.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogFormat is an enumeration of valid log output formats.
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// ConfigSource represents metadata about a configuration source.
type ConfigSource struct {
	Type       SourceType // Type of configuration source
	Path       string     // File path (for file-based sources) or "env" / "cli"
	Precedence int        // Precedence level (higher = higher priority)
	LoadedAt   time.Time  // Timestamp when source was loaded
	Keys       []string   // Configuration keys provided by this source
}

// SourceType enumerates the types of configuration sources.
type SourceType string

const (
	SourceTypeDefault SourceType = "default"  // Built-in defaults (Precedence: 0)
	SourceTypeEnvFile SourceType = "env_file" // .env files (Precedence: 1)
	SourceTypeYAML    SourceType = "yaml"     // YAML file (Precedence: 2)
	SourceTypeCLI     SourceType = "cli"      // Command-line flags (Precedence: 3)
)

// StorageConfig contains configuration for the storage layer.
// Specifies which backend (memory or postgres) to use and its parameters.
type StorageConfig struct {
	Backend  string          `mapstructure:"backend" validate:"required,oneof=memory postgres"`
	Postgres PostgresConfig  `mapstructure:"postgres"`
	Timeouts StorageTimeouts `mapstructure:"timeouts" validate:"required"`
}

// PostgresConfig contains PostgreSQL-specific connection parameters.
// Only used when StorageConfig.Backend is "postgres".
type PostgresConfig struct {
	ConnectionURL string `mapstructure:"connection_url" validate:"required_if=Backend postgres"`
}

// StorageTimeouts defines timeout durations for storage operations.
// Applied to all backends to prevent indefinite hangs.
type StorageTimeouts struct {
	Read  time.Duration `mapstructure:"read" validate:"required"`
	Write time.Duration `mapstructure:"write" validate:"required"`
}

// ThirdPartyOAuth2Config contains configuration for OAuth2 session management with third-party services.
// This configuration is optional - if not provided, OAuth2 sessions routes will not be registered.
// When provided, enables users to authenticate with external OAuth2 providers
// (GitHub, Google, Microsoft, etc.) through the identity broker.
type ThirdPartyOAuth2Config struct {
	// JWESigningKey is the base64-encoded key for signing JWE state tokens (REQUIRED, 32 bytes).
	// This key MUST be kept secret. It protects OAuth2 state tokens during authorization flows.
	// REQUIRED - application will fail to start if not provided.
	// Generate with: openssl rand -base64 32
	// Store in environment variable: IDENTITY_BROKER_JWE_SIGNING_KEY
	JWESigningKey string `mapstructure:"jwe_signing_key"`

	// StateTokenTTL is the time-to-live for OAuth2 state tokens.
	// Maximum allowed: 15 minutes per security requirements (SR-008).
	// Shorter TTL reduces exposure window for state token leakage.
	// Default: 10 minutes
	StateTokenTTL time.Duration `mapstructure:"state_token_ttl"`

	// PKCEVerifierLength is the length of PKCE code verifier in bytes.
	// Must be 32-128 bytes per RFC 7636.
	// Default: 32 bytes (256 bits of entropy)
	PKCEVerifierLength int `mapstructure:"pkce_verifier_length"`
}

// OAuth2AuthServerConfig represents configuration for OAuth2 authorization server functionality.
type OAuth2AuthServerConfig struct {
	UpstreamIssuerURI         string   `mapstructure:"upstream_issuer_uri"`
	UpstreamAuthorizeEndpoint string   `mapstructure:"upstream_authorize_endpoint"`
	UpstreamTokenEndpoint     string   `mapstructure:"upstream_token_endpoint"`
	SupportedResponseTypes    []string `mapstructure:"supported_response_types"`
	SupportedGrantTypes       []string `mapstructure:"supported_grant_types"`
	UpstreamTimeoutSeconds    int      `mapstructure:"upstream_timeout_seconds"`
	Mode                      string   `mapstructure:"mode"`
}

// Validate validates the OAuth2AuthServerConfig structure.
// Sets defaults for empty fields and returns an error for missing required fields.
func (c *OAuth2AuthServerConfig) Validate() error {
	// Check required fields
	if c.UpstreamIssuerURI == "" {
		return c.newValidationError("oauth2_authorization_server.upstream_issuer_uri")
	}

	if c.UpstreamAuthorizeEndpoint == "" {
		return c.newValidationError("oauth2_authorization_server.upstream_authorize_endpoint")
	}

	if c.UpstreamTokenEndpoint == "" {
		return c.newValidationError("oauth2_authorization_server.upstream_token_endpoint")
	}

	// Set defaults for optional fields
	if len(c.SupportedResponseTypes) == 0 {
		c.SupportedResponseTypes = []string{"code"}
	}

	if len(c.SupportedGrantTypes) == 0 {
		c.SupportedGrantTypes = []string{"authorization_code"}
	}

	if c.UpstreamTimeoutSeconds == 0 {
		c.UpstreamTimeoutSeconds = 30
	}

	if c.Mode == "" {
		c.Mode = "proxy"
	}

	return nil
}

// newValidationError creates a validation error for the given field.
// This is a helper to create errors compatible with domain/config.ConfigError.
func (c *OAuth2AuthServerConfig) newValidationError(field string) error {
	// Return a generic error that has a Field() method for test compatibility
	// The actual ConfigError type is created in the config adapter layer
	return &oauth2ValidationError{field: field}
}

// oauth2ValidationError is a simple internal error type to avoid circular imports
type oauth2ValidationError struct {
	field string
}

// Error implements the error interface
func (e *oauth2ValidationError) Error() string {
	return "config validation error"
}

// Field returns the error field for test compatibility
func (e *oauth2ValidationError) Field() string {
	return e.field
}

// SecurityConfig contains security-related configuration.
type SecurityConfig struct {
	// SkipThirdpartyHTTPSValidation skips HTTPS certificate validation for third-party OAuth2 services.
	// WARNING: This is ONLY for development/test environments!
	// Allows HTTP connections and invalid HTTPS certificates.
	// NEVER enable this in production.
	SkipThirdpartyHTTPSValidation bool `mapstructure:"skip_thirdparty_https_validation"`
}
