// Package config implements the configuration loading adapter.
package config

import "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"

// Config represents the complete application configuration schema.
// Uses mapstructure tags for Viper unmarshaling and validate tags for validation.
type Config struct {
	Log LogConfig `mapstructure:"log" validate:"required"`
	// Future: Server, Database, Auth, etc.
}

// LogConfig contains logging-related configuration.
type LogConfig struct {
	Level  config.LogLevel  `mapstructure:"level" validate:"required"`
	Format config.LogFormat `mapstructure:"format" validate:"required"`
}

// DefaultConfig returns the default configuration values.
var DefaultConfig = Config{
	Log: LogConfig{
		Level:  config.LogLevelInfo,
		Format: config.LogFormatText,
	},
}
