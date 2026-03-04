// Package config implements the configuration loading adapter.
package config

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	domconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwtauth"
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

	// Validate server configuration (pass security config for HTTPS validation)
	if err := validateServerConfig(&cfg.Server, &cfg.Security); err != nil {
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

	// Validate OAuth2 Authorization Server configuration (if provided)
	if err := validateOAuth2AuthServerConfig(&cfg.OAuth2AuthServer); err != nil {
		return err
	}

	// Validate encryption configuration
	if err := validateEncryptionConfig(&cfg.Encryption); err != nil {
		return err
	}

	return nil
}

// validateServerConfig validates the server configuration for all instances.
// Ensures required authentication settings are properly configured.
func validateServerConfig(sc *ports.ServerConfig, security *ports.SecurityConfig) error {
	// Validate EndUser server
	if err := validateServerInstance(&sc.EndUser, "server.enduser", security); err != nil {
		return err
	}

	// Validate Admin server
	if err := validateServerInstance(&sc.Admin, "server.admin", security); err != nil {
		return err
	}

	return nil
}

// validateServerInstanceAuth validates authentication configuration for a server instance.
func validateServerInstanceAuth(sic *ports.ServerInstanceConfig, prefix string, security *ports.SecurityConfig) error {
	// Validate preauth principal header name
	if sic.Authentication.Preauth.PrincipalHeaderName == "" {
		return formatValidationError(
			prefix+".authentication.preauth.principal_header_name",
			"",
			"non-empty HTTP header name",
			nil,
		)
	}

	// Validate JWT configuration if present
	if sic.Authentication.JWT != nil {
		if err := validateJWTConfig(sic.Authentication.JWT, prefix+".authentication.jwt", security); err != nil {
			return err
		}
	}

	return nil
}

// validateServerInstanceConfig validates a server instance configuration.
func validateServerInstance(sic *ports.ServerInstanceConfig, prefix string, security *ports.SecurityConfig) error {
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
	if err := validateServerInstanceAuth(sic, prefix, security); err != nil {
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
	domainLevel := domconfig.LogLevel(level)
	return domainLevel.Validate()
}

// validateLogFormat validates the log format value.
// Fast, zero-allocation validation using domain type.
func validateLogFormat(format ports.LogFormat) error {
	domainFormat := domconfig.LogFormat(format)
	return domainFormat.Validate()
}

// validateThirdPartyOAuth2Config validates the third-party OAuth2 configuration.
func validateThirdPartyOAuth2Config(cfg *ports.ThirdPartyOAuth2Config) error {

	// JWESigningKey is mandatory - must be provided
	if cfg.JWESigningKey == "" {
		return formatValidationError(
			"third_party_oauth2.jwe_signing_key",
			"",
			"base64-encoded key (32 bytes)",
			nil,
		)
	}

	// Validate base64 encoding
	keyBytes, err := base64.StdEncoding.DecodeString(cfg.JWESigningKey)
	if err != nil {
		// Don't log full key value for security
		truncated := cfg.JWESigningKey
		if len(truncated) > 10 {
			truncated = truncated[:10] + "..."
		}
		return formatValidationError(
			"third_party_oauth2.jwe_signing_key",
			truncated,
			"valid base64-encoded string",
			err,
		)
	}

	// Validate key length (must be exactly 32 bytes for A256GCMKW)
	if len(keyBytes) != 32 {
		return formatValidationError(
			"third_party_oauth2.jwe_signing_key",
			fmt.Sprintf("%d bytes", len(keyBytes)),
			"exactly 32 bytes when decoded",
			nil,
		)
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

// validateOAuth2AuthServerConfig validates the OAuth2 Authorization Server configuration.
// OAuth2AuthServer is optional, so we only validate if configuration is provided.
func validateOAuth2AuthServerConfig(cfg *ports.OAuth2AuthServerConfig) error {
	// If no OAuth2 Authorization Server configuration is provided, skip validation
	if cfg.UpstreamAuthorizeEndpoint == "" && cfg.UpstreamTokenEndpoint == "" && cfg.UpstreamIssuerURI == "" {
		return nil
	}

	// If any OAuth2 config is provided, use the config struct's Validate method
	if err := cfg.Validate(); err != nil {
		// Convert validation error to ConfigError format for consistency
		// The Validate() method returns oauth2ValidationError, we'll wrap it
		return formatValidationError(
			"oauth2_authorization_server",
			"",
			"valid OAuth2 authorization server configuration",
			err,
		)
	}

	return nil
}

// formatValidationError converts validation errors to ConfigError with context.
func formatValidationError(field string, value string, expected string, err error) error {
	return &domconfig.ConfigError{
		Field:    field,
		Value:    value,
		Expected: expected,
		Source:   "", // Will be filled in by caller if known
		Err:      err,
	}
}

// validateEncryptionConfig validates the new backend-explicit encryption configuration.
// Ensures exactly one backend (AWS KMS or Memory) is configured.
func validateEncryptionConfig(cfg *ports.EncryptionConfig) error {
	// Count configured backends
	backendCount := 0
	if cfg.AWSKMS != nil {
		backendCount++
	}
	if cfg.Memory != nil {
		backendCount++
	}

	// Ensure exactly one backend is configured
	if backendCount == 0 {
		return formatValidationError(
			"encryption",
			"neither backend configured",
			"exactly one backend (aws_kms or memory) must be configured",
			nil,
		)
	}
	if backendCount > 1 {
		return formatValidationError(
			"encryption",
			"multiple backends configured",
			"exactly one backend (aws_kms or memory) must be configured, not both",
			nil,
		)
	}

	// Validate backend-specific configuration
	if cfg.AWSKMS != nil {
		if err := validateAWSKMSConfig(cfg.AWSKMS); err != nil {
			return err
		}
	}
	if cfg.Memory != nil {
		if err := validateMemoryConfig(cfg.Memory); err != nil {
			return err
		}
	}

	return nil
}

// validateAWSKMSConfig validates AWS KMS backend configuration.
func validateAWSKMSConfig(cfg *ports.AWSKMSConfig) error {
	// KeyARN is required
	if cfg.KeyARN == "" {
		return formatValidationError(
			"encryption.aws_kms.key_arn",
			"",
			"AWS KMS ARN (arn:aws:kms:region:account:key/key-id)",
			nil,
		)
	}

	// Validate KeyARN format
	if !strings.HasPrefix(cfg.KeyARN, "arn:aws:kms:") {
		return formatValidationError(
			"encryption.aws_kms.key_arn",
			maskSensitiveValue(cfg.KeyARN),
			"valid AWS KMS ARN format (arn:aws:kms:region:account:key/key-id)",
			nil,
		)
	}

	// Basic ARN structure validation
	parts := strings.Split(cfg.KeyARN, ":")
	if len(parts) < 6 {
		return formatValidationError(
			"encryption.aws_kms.key_arn",
			maskSensitiveValue(cfg.KeyARN),
			"valid AWS KMS ARN with at least 6 colon-separated parts",
			nil,
		)
	}

	return nil
}

// validateMemoryConfig validates Memory backend configuration.
func validateMemoryConfig(cfg *ports.MemoryConfig) error {
	// RawKey is required
	if cfg.RawKey == "" {
		return formatValidationError(
			"encryption.memory.raw_key",
			"",
			"base64-encoded 32-byte AES-256 key",
			nil,
		)
	}

	// Validate base64 encoding
	keyBytes, err := base64.StdEncoding.DecodeString(cfg.RawKey)
	if err != nil {
		return formatValidationError(
			"encryption.memory.raw_key",
			maskSensitiveValue(cfg.RawKey),
			"valid base64-encoded string",
			err,
		)
	}

	// Validate key length (must be exactly 32 bytes for AES-256)
	if len(keyBytes) != 32 {
		return formatValidationError(
			"encryption.memory.raw_key",
			fmt.Sprintf("%d bytes", len(keyBytes)),
			"exactly 32 bytes when decoded (AES-256)",
			nil,
		)
	}

	return nil
}

// validateJWTConfig validates JWT pre-authentication configuration.
// Applies defaults for HeaderName, Verification, and PrincipalExpression when not set.
// Enforces mutual exclusivity between verification: none and jwks_uri.
// Enforces HTTPS for JWKS URI per SR-004.
func validateJWTConfig(jwt *ports.JWTConfig, prefix string, security *ports.SecurityConfig) error {
	// Apply defaults for fields that have default values
	if jwt.HeaderName == "" {
		jwt.HeaderName = "Authorization"
	}
	if jwt.Verification == "" {
		jwt.Verification = "jwks"
	}
	if jwt.ClaimExtraction.PrincipalExpression == "" {
		jwt.ClaimExtraction.PrincipalExpression = "claims.sub"
	}

	// Validate verification mode
	if jwt.Verification != "jwks" && jwt.Verification != "none" {
		return formatValidationError(
			prefix+".verification",
			jwt.Verification,
			"'jwks' or 'none'",
			nil,
		)
	}

	// Mutual exclusivity: verification: none + jwks_uri → startup error (FR-003a)
	if jwt.Verification == "none" && jwt.JWKSURI != "" {
		return formatValidationError(
			prefix,
			"verification: none with jwks_uri: "+jwt.JWKSURI,
			"verification 'none' and jwks_uri are mutually exclusive",
			nil,
		)
	}

	// Required jwks_uri when verification is jwks
	if jwt.Verification == "jwks" && jwt.JWKSURI == "" {
		return formatValidationError(
			prefix+".jwks_uri",
			"",
			"non-empty JWKS URI (required when verification is 'jwks')",
			nil,
		)
	}

	// HTTPS enforcement for JWKS URI (SR-004)
	if jwt.JWKSURI != "" {
		u, err := url.Parse(jwt.JWKSURI)
		if err != nil {
			return formatValidationError(
				prefix+".jwks_uri",
				jwt.JWKSURI,
				"valid URL",
				err,
			)
		}
		skipHTTPS := security != nil && security.SkipThirdpartyHTTPSValidation
		if u.Scheme == "http" && !skipHTTPS {
			return formatValidationError(
				prefix+".jwks_uri",
				jwt.JWKSURI,
				"HTTPS URL (set security.skip_thirdparty_https_validation to allow HTTP in development)",
				nil,
			)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return formatValidationError(
				prefix+".jwks_uri",
				jwt.JWKSURI,
				"HTTP or HTTPS URL",
				nil,
			)
		}
	}

	// Validate CEL claim extraction expressions by attempting compilation.
	// This catches syntax/type errors at startup rather than at request time.
	_, err := jwtauth.NewCELEvaluator(jwtauth.CELEvaluatorConfig{
		PrincipalExpression:   jwt.ClaimExtraction.PrincipalExpression,
		DisplayNameExpression: jwt.ClaimExtraction.DisplayNameExpression,
		EmailExpression:       jwt.ClaimExtraction.EmailExpression,
		PictureURLExpression:  jwt.ClaimExtraction.PictureURLExpression,
	}, nil)
	if err != nil {
		return formatValidationError(
			prefix+".claim_extraction",
			err.Error(),
			"valid CEL expressions for principal_expression, display_name_expression, email_expression, picture_url_expression",
			err,
		)
	}

	return nil
}

// maskSensitiveValue masks sensitive values for display in error messages.
// Shows only the first 10 and last 5 characters for readability while maintaining some specificity.
func maskSensitiveValue(value string) string {
	if len(value) <= 15 {
		return "***" // Too short to safely display any part
	}
	return value[:10] + "..." + value[len(value)-5:]
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
