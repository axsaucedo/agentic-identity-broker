// Package config implements the configuration loading adapter.
package config

import "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"

// Config represents the complete application configuration schema.
// Uses mapstructure tags for Viper unmarshaling and validate tags for validation.
type Config struct {
	Log LogConfig `mapstructure:"log" validate:"required"`
	SPA SPAConfig `mapstructure:"spa"`
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
}
