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

	// Validate server configuration
	if err := validateServerConfig(&cfg.Server); err != nil {
		return err
	}

	// Validate storage configuration
	if err := validateStorageConfig(&cfg.Storage); err != nil {
		return err
	}

	return nil
}

// validateServerConfig validates the server configuration for all instances.
// Ensures required authentication settings are properly configured.
func validateServerConfig(sc *ports.ServerConfig) error {
	// Validate EndUser server authentication
	if err := validateServerInstanceAuth(&sc.EndUser, "server.enduser"); err != nil {
		return err
	}

	// Validate Admin server authentication
	if err := validateServerInstanceAuth(&sc.Admin, "server.admin"); err != nil {
		return err
	}

	return nil
}

// validateServerInstanceAuth validates authentication configuration for a server instance.
func validateServerInstanceAuth(sic *ports.ServerInstanceConfig, prefix string) error {
	// Validate preauth principal header name
	if sic.Authentication.Preauth.PrincipalHeaderName == "" {
		return formatValidationError(
			prefix+".authentication.preauth.principal_header_name",
			"",
			"non-empty HTTP header name",
			nil,
		)
	}

	return nil
}

// validateStorageConfig validates the storage configuration.
// Ensures backend is valid and backend-specific parameters are present.
func validateStorageConfig(sc *ports.StorageConfig) error {
	// Validate backend value
	if sc.Backend != "memory" && sc.Backend != "postgres" {
		return formatValidationError("storage.backend", sc.Backend, "memory or postgres", nil)
	}

	// Validate storage timeouts
	if err := validateStorageTimeouts(&sc.Timeouts); err != nil {
		return err
	}

	// Validate backend-specific parameters
	if sc.Backend == "postgres" {
		if err := validatePostgresConfig(&sc.Postgres); err != nil {
			return err
		}
	}

	return nil
}

// validateStorageTimeouts validates storage timeout configuration.
func validateStorageTimeouts(st *ports.StorageTimeouts) error {
	if st.Read <= 0 {
		return formatValidationError("storage.timeouts.read", st.Read.String(), "positive duration", nil)
	}
	if st.Write <= 0 {
		return formatValidationError("storage.timeouts.write", st.Write.String(), "positive duration", nil)
	}
	return nil
}

// validatePostgresConfig validates PostgreSQL-specific configuration.
func validatePostgresConfig(pc *ports.PostgresConfig) error {
	if pc.ConnectionURL == "" {
		return formatValidationError("storage.postgres.connection_url", "", "non-empty PostgreSQL connection URL", nil)
	}

	// Validate connection URL format
	if !isValidPostgresURL(pc.ConnectionURL) {
		return formatValidationError("storage.postgres.connection_url", pc.ConnectionURL, "valid postgresql:// URL", nil)
	}

	return nil
}

// isValidPostgresURL checks if a string is a valid PostgreSQL URL format.
func isValidPostgresURL(s string) bool {
	if s == "" {
		return false
	}
	return (len(s) > 13 && (s[:13] == "postgresql://" || s[:11] == "postgres://"))
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
