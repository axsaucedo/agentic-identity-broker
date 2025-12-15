// Package config implements the configuration loading adapter.
package config

import (
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Validate validates the complete configuration structure.
// Returns a ConfigError with detailed information on validation failures.
func Validate(cfg *ports.Config) error {
	// Validate log level
	if err := validateLogLevel(cfg.Log.Level); err != nil {
		return formatValidationError("log.level", string(cfg.Log.Level), "debug, info, warn, or error", err)
	}

	// Validate log format
	if err := validateLogFormat(cfg.Log.Format); err != nil {
		return formatValidationError("log.format", string(cfg.Log.Format), "text or json", err)
	}

	return nil
}

// validateLogLevel validates the log level value.
// Fast, zero-allocation validation using domain type.
func validateLogLevel(level ports.LogLevel) error {
	domainLevel := config.LogLevel(level)
	return domainLevel.Validate()
}

// validateLogFormat validates the log format value.
// Fast, zero-allocation validation using domain type.
func validateLogFormat(format ports.LogFormat) error {
	domainFormat := config.LogFormat(format)
	return domainFormat.Validate()
}

// formatValidationError converts validation errors to ConfigError with context.
func formatValidationError(field string, value string, expected string, err error) error {
	return &config.ConfigError{
		Field:    field,
		Value:    value,
		Expected: expected,
		Source:   "", // Will be filled in by caller if known
		Err:      err,
	}
}
