// Package server contains domain logic for dual-server coordination.
package server

import (
	"fmt"
	"net"
	"regexp"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// ValidateServerConfig validates server configuration according to business rules.
// This performs manual validation as per constitution requirements (no external validator library).
func ValidateServerConfig(cfg *ports.ServerConfig) error {
	// Validate end-user server configuration
	if err := validateServerInstanceConfig("enduser", &cfg.EndUser); err != nil {
		return err
	}

	// Validate admin server configuration
	if err := validateServerInstanceConfig("admin", &cfg.Admin); err != nil {
		return err
	}

	// Validate shutdown configuration
	if err := validateShutdownConfig(&cfg.Shutdown); err != nil {
		return err
	}

	// Cross-field validation: ports must be different
	if cfg.EndUser.Port == cfg.Admin.Port {
		return fmt.Errorf("enduser and admin ports must be different (both are %d)", cfg.EndUser.Port)
	}

	return nil
}

// validateServerInstanceConfig validates a single server instance configuration.
func validateServerInstanceConfig(name string, cfg *ports.ServerInstanceConfig) error {
	// Validate port range
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("%s server port must be between 1 and 65535, got %d", name, cfg.Port)
	}

	// Validate bind address
	if err := validateBindAddress(cfg.Bind); err != nil {
		return fmt.Errorf("%s server bind address invalid: %w", name, err)
	}

	return nil
}

// validateShutdownConfig validates shutdown timeout configuration.
func validateShutdownConfig(cfg *ports.ShutdownConfig) error {
	// Validate timeout is positive
	if cfg.Timeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive, got %v", cfg.Timeout)
	}

	// Validate timeout is reasonable (1s to 5m as per spec)
	const minTimeout = 1e9  // 1 second in nanoseconds
	const maxTimeout = 3e11 // 5 minutes in nanoseconds

	if cfg.Timeout < minTimeout {
		return fmt.Errorf("shutdown timeout must be at least 1 second, got %v", cfg.Timeout)
	}

	if cfg.Timeout > maxTimeout {
		return fmt.Errorf("shutdown timeout must be at most 5 minutes, got %v", cfg.Timeout)
	}

	return nil
}

// validateBindAddress validates a bind address (IPv4, IPv6, or hostname).
// Accepts empty string (will use default), valid IP addresses, or valid hostnames.
// This performs syntax validation only, no DNS lookups.
func validateBindAddress(addr string) error {
	// Empty string is valid (uses default)
	if addr == "" {
		return nil
	}

	// Check for valid IPv4 address
	if ip := net.ParseIP(addr); ip != nil {
		if ip.To4() != nil {
			// Valid IPv4
			return nil
		}
		if ip.To16() != nil {
			// Valid IPv6
			return nil
		}
	}

	// Check for valid hostname
	// RFC 1123 hostname: alphanumeric, hyphens, dots, max 255 chars
	// Each label max 63 chars, must start/end with alphanumeric
	if len(addr) > 255 {
		return fmt.Errorf("hostname too long (max 255 characters): %s", addr)
	}

	// Hostname regex: labels separated by dots, each label 1-63 chars
	hostnameRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
	if !hostnameRegex.MatchString(addr) {
		return fmt.Errorf("not a valid IPv4, IPv6, or hostname: %s", addr)
	}

	return nil
}
