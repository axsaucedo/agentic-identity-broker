// Package upstream implements adapters for communication with upstream services.
package upstream

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"time"
)

// NewSecureUpstreamClient creates a new HTTP client configured for secure communication
// with upstream OAuth2 servers. It enforces TLS certificate validation and modern TLS versions.
//
// Security features:
// - Minimum TLS 1.2 (no legacy SSL/TLS)
// - System certificate pool for validation (respects system-installed CA certificates)
// - No InsecureSkipVerify (certificate validation is mandatory)
// - Timeout configured for connection and request limits
// - Modern cipher suites selected by Go (excludes weak ciphers)
//
// Parameters:
// - timeout: Maximum time for an HTTP request (includes connection, request, response)
//
// Returns:
// - *http.Client configured for secure upstream communication
// - error if certificate pool cannot be loaded
func NewSecureUpstreamClient(timeout time.Duration) (*http.Client, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	// Load system certificate pool for TLS validation
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to load system certificate pool: %w", err)
	}

	// Configure TLS with security-first defaults
	tlsConfig := &tls.Config{
		// Enforce minimum TLS version - reject SSLv3, TLS 1.0, 1.1
		MinVersion: tls.VersionTLS12,

		// Use system certificate pool for validation
		RootCAs: certPool,

		// Certificate validation is mandatory - no InsecureSkipVerify
		// This ensures we reject:
		// - Self-signed certificates
		// - Expired certificates
		// - Certificates with wrong hostname
		// - Certificates from untrusted CAs
		InsecureSkipVerify: false,
	}

	// Configure HTTP transport with timeouts and TLS
	transport := &http.Transport{
		// TLS configuration
		TLSClientConfig: tlsConfig,

		// Connection timeouts
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // Connection establishment timeout
			KeepAlive: 30 * time.Second,
		}).DialContext,

		// HTTP/1.1 keepalive
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,

		// Request limits
		MaxConnsPerHost:       10,
		DisableCompression:    false, // Allow compression for efficiency
		DisableKeepAlives:     false, // Keep connections alive
		ForceAttemptHTTP2:     true,  // Enable HTTP/2 when available
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Create HTTP client with configured transport
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout, // Total request timeout
	}

	return client, nil
}
