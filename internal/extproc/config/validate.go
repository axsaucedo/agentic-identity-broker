// Package config — validate.go provides startup validation for ExtProc configuration.
// Validation rules from specs/015-extproc-token-exchange/contracts/configuration.md
// are enforced here. Validation runs after loading and before the service starts.
package config

import (
	"fmt"
	"net/url"
	"strings"
)

// Validate checks the Config for required fields and constraint violations.
// Returns a descriptive error if any rules are violated (all errors collected).
// Implements fail-fast startup validation per FR-016 and US3 Scenario 2.
//
// Validation rules (per contracts/configuration.md):
//  1. grpc.port must be 1-65535
//  2. grpc.bind must not be empty
//  3. oauth2.token_endpoint must be a valid URL starting with http:// or https://
//  4. oauth2.issuer must be a valid URL starting with http:// or https://
//  5. oauth2.client_id must not be empty
//  6. oauth2.client_secret must not be empty after env var expansion
//  7. cache.default_ttl must be a positive duration
//  8. token_endpoint and issuer must use https:// unless oauth2.tls.allow_http is true
//  9. cache.max_ttl must be a positive duration
//  10. oauth2.exchange_timeout must be a positive duration
//  11. log.level must be one of debug, info, warn, error
//  12. log.format must be one of text, json
//  13. oauth2.client_assertion_type must be one of id_token, access_token
//  14. circuit_breaker.max_failures must be >= 1
//  15. circuit_breaker.reset_timeout must be a positive duration
func Validate(cfg *Config) error {
	var errs []string

	// Rule 1: grpc.port must be 1-65535
	if cfg.GRPC.Port < 1 || cfg.GRPC.Port > 65535 {
		errs = append(errs, fmt.Sprintf("grpc.port must be between 1 and 65535, got %d", cfg.GRPC.Port))
	}

	// Rule 2: grpc.bind must not be empty
	if strings.TrimSpace(cfg.GRPC.Bind) == "" {
		errs = append(errs, "grpc.bind must not be empty")
	}

	// Rule 3: oauth2.token_endpoint must be a valid URL with http/https scheme
	if err := validateURL("oauth2.token_endpoint", cfg.OAuth2.TokenEndpoint, true); err != nil {
		errs = append(errs, err.Error())
	}

	// Rule 4: oauth2.issuer must be a valid URL with http or https scheme
	if err := validateURL("oauth2.issuer", cfg.OAuth2.Issuer, true); err != nil {
		errs = append(errs, err.Error())
	}

	// Rule 5: oauth2.client_id must not be empty
	if strings.TrimSpace(cfg.OAuth2.ClientID) == "" {
		errs = append(errs, "oauth2.client_id must not be empty")
	}

	// Rule 6: oauth2.client_secret must not be empty
	if strings.TrimSpace(cfg.OAuth2.ClientSecret) == "" {
		errs = append(errs, "oauth2.client_secret must not be empty")
	}

	// Rule 7: cache.default_ttl must be positive
	if cfg.Cache.DefaultTTL <= 0 {
		errs = append(errs, "cache.default_ttl must be a positive duration")
	}

	// Rule 8: TLS enforcement — token_endpoint and issuer must use https:// unless allow_http is set
	if !cfg.OAuth2.TLS.AllowHTTP {
		if cfg.OAuth2.TokenEndpoint != "" && strings.HasPrefix(cfg.OAuth2.TokenEndpoint, "http://") {
			errs = append(errs, "oauth2.token_endpoint must use https:// scheme (set oauth2.tls.allow_http: true to disable — DEV ONLY)")
		}
		if cfg.OAuth2.Issuer != "" && strings.HasPrefix(cfg.OAuth2.Issuer, "http://") {
			errs = append(errs, "oauth2.issuer must use https:// scheme (set oauth2.tls.allow_http: true to disable — DEV ONLY)")
		}
	}

	// Rule 9: cache.max_ttl must be positive
	if cfg.Cache.MaxTTL <= 0 {
		errs = append(errs, "cache.max_ttl must be a positive duration")
	}

	// Rule 10: oauth2.exchange_timeout must be positive
	if cfg.OAuth2.ExchangeTimeout <= 0 {
		errs = append(errs, "oauth2.exchange_timeout must be a positive duration")
	}

	// Rule 11: log.level must be a valid level if non-empty
	switch cfg.Log.Level {
	case "debug", "info", "warn", "error", "":
		// valid
	default:
		errs = append(errs, fmt.Sprintf("log.level must be one of debug, info, warn, error; got %q", cfg.Log.Level))
	}

	// Rule 12: log.format must be a valid format if non-empty
	switch cfg.Log.Format {
	case "text", "json", "":
		// valid
	default:
		errs = append(errs, fmt.Sprintf("log.format must be one of text, json; got %q", cfg.Log.Format))
	}

	// Rule 13: oauth2.client_assertion_type must be either "id_token" or "access_token"
	switch cfg.OAuth2.ClientAssertionType {
	case "id_token", "access_token":
		// valid
	default:
		errs = append(errs, fmt.Sprintf("oauth2.client_assertion_type must be one of id_token, access_token; got %q", cfg.OAuth2.ClientAssertionType))
	}

	// Rule 14: circuit_breaker.max_failures must be positive
	if cfg.CircuitBreaker.MaxFailures < 1 {
		errs = append(errs, fmt.Sprintf("circuit_breaker.max_failures must be >= 1, got %d", cfg.CircuitBreaker.MaxFailures))
	}

	// Rule 15: circuit_breaker.reset_timeout must be positive
	if cfg.CircuitBreaker.ResetTimeout <= 0 {
		errs = append(errs, "circuit_breaker.reset_timeout must be a positive duration")
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// validateURL checks that s is a non-empty valid URL. When requireHTTPScheme is true,
// the URL must have an http or https scheme. The host must be non-empty.
func validateURL(field, s string, requireHTTPScheme bool) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s must not be empty", field)
	}
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return fmt.Errorf("%s must be a valid URL: %w", field, err)
	}
	if requireHTTPScheme && u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%s must be a valid URL starting with http:// or https://", field)
	}
	if u.Host == "" {
		return fmt.Errorf("%s must be a valid URL with a host", field)
	}
	return nil
}
