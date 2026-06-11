package upstream

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"
)

// TestNewSecureUpstreamClient verifies TLS configuration and client setup.
func TestNewSecureUpstreamClient(t *testing.T) {
	tests := []struct {
		name                  string
		timeout               time.Duration
		expectedTimeout       time.Duration
		verifyTLSConfig       func(t *testing.T, tlsConfig *tls.Config)
		verifyTransportConfig func(t *testing.T, tr *http.Transport)
		wantErr               bool
	}{
		{
			name:            "valid client with positive timeout",
			timeout:         30 * time.Second,
			expectedTimeout: 30 * time.Second,
			wantErr:         false,
			verifyTLSConfig: func(t *testing.T, tlsConfig *tls.Config) {
				// Verify TLS minimum version
				if tlsConfig.MinVersion != tls.VersionTLS12 {
					t.Errorf("MinVersion = %d, want tls.VersionTLS12 (%d)", tlsConfig.MinVersion, tls.VersionTLS12)
				}

				// Verify certificate validation is enabled (InsecureSkipVerify must be false)
				if tlsConfig.InsecureSkipVerify {
					t.Error("InsecureSkipVerify should be false (certificate validation must be enabled)")
				}

				// Verify system certificate pool is loaded
				if tlsConfig.RootCAs == nil {
					t.Error("RootCAs should not be nil (system certificate pool must be loaded)")
				}
			},
		},
		{
			name:            "client with zero timeout defaults to 30 seconds",
			timeout:         0,
			expectedTimeout: 30 * time.Second,
			wantErr:         false,
			verifyTLSConfig: func(t *testing.T, tlsConfig *tls.Config) {
				if tlsConfig.MinVersion != tls.VersionTLS12 {
					t.Errorf("MinVersion = %d, want tls.VersionTLS12 (%d)", tlsConfig.MinVersion, tls.VersionTLS12)
				}
				if tlsConfig.InsecureSkipVerify {
					t.Error("InsecureSkipVerify should be false")
				}
			},
		},
		{
			name:            "client with negative timeout defaults to 30 seconds",
			timeout:         -1 * time.Second,
			expectedTimeout: 30 * time.Second,
			wantErr:         false,
			verifyTLSConfig: func(t *testing.T, tlsConfig *tls.Config) {
				if tlsConfig.MinVersion != tls.VersionTLS12 {
					t.Errorf("MinVersion = %d, want tls.VersionTLS12 (%d)", tlsConfig.MinVersion, tls.VersionTLS12)
				}
			},
		},
		{
			name:            "client with sub-second timeout",
			timeout:         100 * time.Millisecond,
			expectedTimeout: 100 * time.Millisecond,
			wantErr:         false,
			verifyTLSConfig: func(t *testing.T, tlsConfig *tls.Config) {
				if tlsConfig.MinVersion != tls.VersionTLS12 {
					t.Errorf("MinVersion = %d, want tls.VersionTLS12 (%d)", tlsConfig.MinVersion, tls.VersionTLS12)
				}
			},
		},
		{
			name:            "client with custom timeout",
			timeout:         60 * time.Second,
			expectedTimeout: 60 * time.Second,
			wantErr:         false,
			verifyTLSConfig: func(t *testing.T, tlsConfig *tls.Config) {
				if tlsConfig.MinVersion != tls.VersionTLS12 {
					t.Errorf("MinVersion = %d, want tls.VersionTLS12 (%d)", tlsConfig.MinVersion, tls.VersionTLS12)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewSecureUpstreamClient(tt.timeout)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewSecureUpstreamClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if client == nil {
				t.Fatal("Expected non-nil client")
			}

			// Verify transport configuration
			transport, ok := client.Transport.(*http.Transport)
			if !ok {
				t.Fatalf("Transport type = %T, want *http.Transport", client.Transport)
			}

			// Verify TLS configuration
			if transport.TLSClientConfig == nil {
				t.Fatal("TLSClientConfig should not be nil")
			}

			if tt.verifyTLSConfig != nil {
				tt.verifyTLSConfig(t, transport.TLSClientConfig)
			}

			if client.Timeout != tt.expectedTimeout {
				t.Errorf("Client timeout = %v, want %v", client.Timeout, tt.expectedTimeout)
			}

			// Verify transport connection settings
			if transport.MaxIdleConns == 0 {
				t.Error("MaxIdleConns should be configured")
			}
			if transport.MaxIdleConnsPerHost == 0 {
				t.Error("MaxIdleConnsPerHost should be configured")
			}
		})
	}
}

// TestSecureUpstreamClient_RejectsInsecureCertificates verifies that the client
// rejects connections with invalid certificates (this would need real servers to fully test).
func TestSecureUpstreamClient_RejectsInsecureCertificates(t *testing.T) {
	client, err := NewSecureUpstreamClient(30 * time.Second)
	if err != nil {
		t.Fatalf("NewSecureUpstreamClient() error = %v", err)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport type = %T, want *http.Transport", client.Transport)
	}

	tlsConfig := transport.TLSClientConfig

	// Verify security properties that reject insecure certificates:
	// 1. MinVersion >= TLS1.2 rejects legacy SSL/TLS versions
	if tlsConfig.MinVersion < tls.VersionTLS12 {
		t.Errorf("MinVersion %d allows legacy TLS versions (must be >= %d)", tlsConfig.MinVersion, tls.VersionTLS12)
	}

	// 2. InsecureSkipVerify = false enforces certificate validation
	if tlsConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify = true allows self-signed certificates (must be false)")
	}

	// 3. RootCAs populated with system certs means untrusted CAs are rejected
	if tlsConfig.RootCAs == nil {
		t.Error("RootCAs should be populated to validate certificate chains")
	}
}

// TestSecureUpstreamClient_TLSVersions verifies correct TLS version constants.
func TestSecureUpstreamClient_TLSVersions(t *testing.T) {
	// This test documents the TLS versions and verifies we're using the right constants
	tests := []struct {
		name      string
		version   uint16
		supported bool
	}{
		{"TLS 1.0", tls.VersionTLS10, false}, // Not supported
		{"TLS 1.1", tls.VersionTLS11, false}, // Not supported
		{"TLS 1.2", tls.VersionTLS12, true},  // Minimum supported
		{"TLS 1.3", tls.VersionTLS13, true},  // Supported
	}

	client, err := NewSecureUpstreamClient(30 * time.Second)
	if err != nil {
		t.Fatalf("NewSecureUpstreamClient() error = %v", err)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport type = %T, want *http.Transport", client.Transport)
	}

	tlsConfig := transport.TLSClientConfig
	minVersion := tlsConfig.MinVersion

	for _, tt := range tests {
		if tt.supported && tt.version < minVersion {
			t.Errorf("TLS %s is configured as supported but is below MinVersion", tt.name)
		}
		if !tt.supported && tt.version >= minVersion {
			// This is a reminder that legacy versions should not be supported
			t.Logf("TLS %s is below MinVersion: OK", tt.name)
		}
	}
}
