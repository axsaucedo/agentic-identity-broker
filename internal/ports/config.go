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
	Log     LogConfig     `mapstructure:"log" validate:"required"`
	Server  ServerConfig  `mapstructure:"server" validate:"required"`
	Storage StorageConfig `mapstructure:"storage" validate:"required"`
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
			Port: 8000,
			Bind: "::", // Dual-stack (IPv6 with IPv4 fallback)
			Authentication: AuthenticationConfig{
				Preauth: PreauthConfig{
					PrincipalHeaderName: "X-Remote-User",
				},
			},
		},
		Admin: ServerInstanceConfig{
			Port: 14000,
			Bind: "::",
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
