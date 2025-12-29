// Package config implements the configuration loading adapter.
package config

import (
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"
)

// Config represents the complete application configuration schema.
// Uses mapstructure tags for Viper unmarshaling and validate tags for validation.
type Config struct {
	Log               LogConfig                `mapstructure:"log" validate:"required"`
	SPA               SPAConfig                `mapstructure:"spa"`
	ThirdPartyOAuth2  ThirdPartyOAuth2Config   `mapstructure:"third_party_oauth2"`
	// Future: Server, Database, Auth, etc.
}

// LogConfig contains logging-related configuration.
type LogConfig struct {
	Level  config.LogLevel  `mapstructure:"level" validate:"required"`
	Format config.LogFormat `mapstructure:"format" validate:"required"`
}

// SPAConfig contains Single Page Application serving configuration.
type SPAConfig struct {
	StaticFilesPath string `mapstructure:"static_files_path"` // Path to dist/consent/ directory
	ServeEnabled    bool   `mapstructure:"serve_enabled"`     // Whether to serve SPA
}

// ThirdPartyOAuth2Config contains configuration for OAuth2 session management with third-party services.
// This configuration is required when users authenticate with external OAuth2 providers
// (GitHub, Google, Microsoft, etc.) through the identity broker.
type ThirdPartyOAuth2Config struct {
	// JWESigningKey is the base64-encoded key for signing JWE state tokens (32+ bytes).
	// This key MUST be kept secret. It protects OAuth2 state tokens during authorization flows.
	// Generate with: openssl rand -base64 32
	// Store in environment variable: IDENTITY_BROKER_JWE_SIGNING_KEY
	JWESigningKey string `mapstructure:"jwe_signing_key" validate:"required,base64,min=44"`

	// StateTokenTTL is the time-to-live for OAuth2 state tokens.
	// Maximum allowed: 15 minutes per security requirements (SR-008).
	// Shorter TTL reduces exposure window for state token leakage.
	// Default: 10 minutes
	StateTokenTTL time.Duration `mapstructure:"state_token_ttl" validate:"required,max=15m"`

	// PKCEVerifierLength is the length of PKCE code verifier in bytes.
	// Must be 32-128 bytes per RFC 7636.
	// Default: 32 bytes (256 bits of entropy)
	PKCEVerifierLength int `mapstructure:"pkce_verifier_length" validate:"required,min=32,max=128"`
}

// DefaultConfig returns the default configuration values.
var DefaultConfig = Config{
	Log: LogConfig{
		Level:  config.LogLevelInfo,
		Format: config.LogFormatText,
	},
	SPA: SPAConfig{
		StaticFilesPath: "./dist/consent",
		ServeEnabled:    false,
	},
	ThirdPartyOAuth2: ThirdPartyOAuth2Config{
		StateTokenTTL:      10 * time.Minute,
		PKCEVerifierLength: 32,
		// JWESigningKey has no default - must be set by user
	},
}
