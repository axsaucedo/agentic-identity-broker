// Package config implements the configuration loading adapter.
package config

import (
	"net/url"

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

	// Validate third-party OAuth2 configuration
	if err := validateThirdPartyOAuth2Config(&cfg.ThirdPartyOAuth2); err != nil {
		return err
	}

	return nil
}

// validateServerConfig validates the server configuration for all instances.
// Ensures required authentication settings are properly configured.
func validateServerConfig(sc *ports.ServerConfig) error {
	// Validate EndUser server
	if err := validateServerInstance(&sc.EndUser, "server.enduser"); err != nil {
		return err
	}

	// Validate Admin server
	if err := validateServerInstance(&sc.Admin, "server.admin"); err != nil {
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

// validateServerInstanceConfig validates a server instance configuration.
func validateServerInstance(sic *ports.ServerInstanceConfig, prefix string) error {
	// Validate port
	if sic.Port < 1 || sic.Port > 65535 {
		return formatValidationError(
			prefix+".port",
			string(rune(sic.Port)),
			"port number between 1 and 65535",
			nil,
		)
	}

	// Validate bind address
	if sic.Bind == "" {
		return formatValidationError(
			prefix+".bind",
			"",
			"non-empty bind address",
			nil,
		)
	}

	// Validate public URL (required for enduser server for OAuth2 callbacks)
	if prefix == "server.enduser" && sic.PublicURL == "" {
		return formatValidationError(
			prefix+".public_url",
			"",
			"non-empty public URL (required for OAuth2 callbacks)",
			nil,
		)
	}

	// Validate public URL format if provided
	if sic.PublicURL != "" {
		if !isValidURL(sic.PublicURL) {
			return formatValidationError(
				prefix+".public_url",
				sic.PublicURL,
				"valid HTTP/HTTPS URL",
				nil,
			)
		}
	}

	// Validate authentication
	if err := validateServerInstanceAuth(sic, prefix); err != nil {
		return err
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

// validateThirdPartyOAuth2Config validates the third-party OAuth2 configuration.
// OAuth2 is optional, so we only validate if configuration is provided.
// If JWESigningKey is empty, OAuth2 routes will not be registered (server logs warning).
func validateThirdPartyOAuth2Config(cfg *ports.ThirdPartyOAuth2Config) error {
	// If no OAuth2 configuration is provided, skip validation
	if cfg.JWESigningKey == "" && cfg.StateTokenTTL == 0 && cfg.PKCEVerifierLength == 0 {
		return nil
	}

	// If JWESigningKey is provided, validate it's in proper format
	if cfg.JWESigningKey != "" {
		// JWESigningKey should be base64-encoded (min 44 chars for 32 bytes)
		if len(cfg.JWESigningKey) < 44 {
			return formatValidationError(
				"third_party_oauth2.jwe_signing_key",
				"",
				"base64-encoded key with minimum 44 characters (32 bytes)",
				nil,
			)
		}
	}

	// Validate StateTokenTTL if provided (should be positive)
	if cfg.StateTokenTTL > 0 {
		// Validate it's within the maximum of 15 minutes per SR-008
		if cfg.StateTokenTTL > 15*60*1e9 { // 15 minutes in nanoseconds
			return formatValidationError(
				"third_party_oauth2.state_token_ttl",
				cfg.StateTokenTTL.String(),
				"positive duration (max 15 minutes per SR-008)",
				nil,
			)
		}
	}

	// Validate PKCEVerifierLength if provided (should be within RFC 7636 limits)
	if cfg.PKCEVerifierLength > 0 {
		if cfg.PKCEVerifierLength < 32 || cfg.PKCEVerifierLength > 128 {
			return formatValidationError(
				"third_party_oauth2.pkce_verifier_length",
				string(rune(cfg.PKCEVerifierLength)),
				"32-128 bytes per RFC 7636",
				nil,
			)
		}
	}

	return nil
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

// isValidURL checks if a string is a valid HTTP or HTTPS URL.
func isValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	// Parse URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Check scheme is http or https
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	// Check host is present
	if u.Host == "" {
		return false
	}

	return true
}
