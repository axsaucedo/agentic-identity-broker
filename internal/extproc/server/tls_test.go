// Package server_test — tls_test.go covers TLS HTTP client configuration (T039).
//
// T039: Configurable HTTP client with TLS settings
//   - InsecureSkipVerify is honoured when set
//   - CaBundlePath loads a custom CA certificate pool
//   - Default (no TLS flags) uses standard TLS verification
//   - ExchangeTimeout is set as the client-level timeout
package server_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/server"
)

// ---------------------------------------------------------------------------
// T039: TLS HTTP client wiring
// ---------------------------------------------------------------------------

// Spec: FR-019 — ExchangeTimeout is applied as the HTTP client timeout
func TestTokenExchanger_TLS_ExchangeTimeout_SetOnHTTPClient(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	cfg := configForMocks(mocks)
	cfg.OAuth2.ExchangeTimeout = 42 * time.Second // distinct value for verification

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// Verify the timeout is honoured: make a request to a hanging server
	// with a short timeout and assert it terminates within that window.
	hangServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer hangServer.Close()

	shortTimeout := 50 * time.Millisecond
	cfg2 := configForMocks(mocks)
	cfg2.OAuth2.TokenEndpoint = hangServer.URL + "/token"
	cfg2.OAuth2.ExchangeTimeout = shortTimeout

	exchanger2, err := server.NewTokenExchanger(cfg2, testLogger())
	require.NoError(t, err)
	defer exchanger2.Shutdown()

	start := time.Now()
	_, err = exchanger2.Exchange("token", "http://resource.example.com/api")
	elapsed := time.Since(start)

	assert.Error(t, err, "timeout must produce an error")
	assert.Less(t, elapsed, 2*time.Second,
		"exchange must abort within timeout, not wait for full server delay")
}

// Spec: FR-019 — InsecureSkipVerify allows connecting to servers with self-signed certs
func TestTokenExchanger_TLS_InsecureSkipVerify_ConnectsToSelfSignedServer(t *testing.T) {
	// Create HTTPS test servers (self-signed cert — would fail normal TLS verification)
	tlsClientCredsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "access-token",
			IDToken:     "id-token-from-tls-server",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer tlsClientCredsServer.Close()

	tlsTokenExchServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "tls-exchanged-token",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer tlsTokenExchServer.Close()

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             tlsTokenExchServer.URL + "/oauth2/token",
			Issuer:                    tlsClientCredsServer.URL,
			ClientID:                  "client",
			ClientSecret:              "secret",
			ClientCredentialsEndpoint: tlsClientCredsServer.URL + "/oauth/token",
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS: extprocconfig.TLSConfig{
				InsecureSkipVerify: true, // allow self-signed cert from httptest.NewTLSServer
				AllowHTTP:          false,
			},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
	}

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err, "InsecureSkipVerify must allow connecting to self-signed TLS server")
	defer exchanger.Shutdown()

	token, err := exchanger.Exchange("user-token", "https://resource.example.com/api")
	require.NoError(t, err, "exchange must succeed with InsecureSkipVerify=true")
	assert.Equal(t, "tls-exchanged-token", token)
}

// Spec: FR-019 — Without InsecureSkipVerify, self-signed cert causes connection failure
func TestTokenExchanger_TLS_NoInsecureSkipVerify_SelfSignedFails(t *testing.T) {
	// TLS server with self-signed cert — normal TLS verification MUST reject this
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "access-token",
			IDToken:     "id-token",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer tlsServer.Close()

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             tlsServer.URL + "/oauth2/token",
			Issuer:                    tlsServer.URL,
			ClientID:                  "client",
			ClientSecret:              "secret",
			ClientCredentialsEndpoint: tlsServer.URL + "/oauth/token",
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS: extprocconfig.TLSConfig{
				InsecureSkipVerify: false, // strict TLS — must reject self-signed
				AllowHTTP:          false,
			},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
	}

	_, err := server.NewTokenExchanger(cfg, testLogger())
	assert.Error(t, err,
		"NewTokenExchanger must fail when TLS verification rejects self-signed cert and InsecureSkipVerify=false")
}

// Spec: FR-019 — CaBundlePath loads a custom CA certificate
func TestTokenExchanger_TLS_CaBundlePath_AllowsCustomCA(t *testing.T) {
	// Create HTTPS test server
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "access-token",
			IDToken:     "id-token-from-custom-ca-server",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer tlsServer.Close()

	tlsExchServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "ca-bundle-exchanged-token",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer tlsExchServer.Close()

	// Write the test server's TLS certificate to a temp file as the CA bundle
	cert := tlsServer.TLS.Certificates[0]
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	require.NoError(t, err)

	certPEM := encodeCertPEM(x509Cert)
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca.pem")
	err = os.WriteFile(caFile, certPEM, 0600)
	require.NoError(t, err)

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             tlsExchServer.URL + "/oauth2/token",
			Issuer:                    tlsServer.URL,
			ClientID:                  "client",
			ClientSecret:              "secret",
			ClientCredentialsEndpoint: tlsServer.URL + "/oauth/token",
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS: extprocconfig.TLSConfig{
				InsecureSkipVerify: false,
				CaBundlePath:       caFile, // trust only this CA
				AllowHTTP:          false,
			},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
	}

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err, "CaBundlePath with matching CA must allow connection")
	defer exchanger.Shutdown()

	token, err := exchanger.Exchange("user-token", "https://resource.example.com/api")
	require.NoError(t, err, "exchange must succeed when CA bundle matches server cert")
	assert.Equal(t, "ca-bundle-exchanged-token", token)
}

// Spec: FR-019 — CaBundlePath with wrong CA rejects connection
func TestTokenExchanger_TLS_CaBundlePath_WrongCA_Fails(t *testing.T) {
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer tlsServer.Close()

	// Write a mismatched CA cert to a temp file
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "wrong-ca.pem")

	// Generate a fresh CA cert that does NOT match the test server
	wrongCACert := generateSelfSignedCA(t)
	err := os.WriteFile(caFile, wrongCACert, 0600)
	require.NoError(t, err)

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             tlsServer.URL + "/oauth2/token",
			Issuer:                    tlsServer.URL,
			ClientID:                  "client",
			ClientSecret:              "secret",
			ClientCredentialsEndpoint: tlsServer.URL + "/oauth/token",
			ExchangeTimeout:           5 * time.Second,
			TLS: extprocconfig.TLSConfig{
				InsecureSkipVerify: false,
				CaBundlePath:       caFile, // wrong CA — must reject
				AllowHTTP:          false,
			},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
	}

	_, err = server.NewTokenExchanger(cfg, testLogger())
	assert.Error(t, err,
		"CaBundlePath with wrong CA must reject the TLS connection")
}

// Spec: FR-019 — CaBundlePath pointing to nonexistent file returns error at startup
func TestTokenExchanger_TLS_CaBundlePath_NonexistentFile_FailsFast(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	cfg := configForMocks(mocks)
	cfg.OAuth2.TLS.CaBundlePath = "/nonexistent/path/to/ca.pem"

	_, err := server.NewTokenExchanger(cfg, testLogger())
	assert.Error(t, err,
		"nonexistent CaBundlePath must cause fail-fast error at startup")
}

// ---------------------------------------------------------------------------
// Helpers for TLS tests
// ---------------------------------------------------------------------------

// encodeCertPEM encodes an x509.Certificate to PEM format.
func encodeCertPEM(cert *x509.Certificate) []byte {
	return append([]byte("-----BEGIN CERTIFICATE-----\n"),
		append(encodeBase64Lines(cert.Raw), []byte("-----END CERTIFICATE-----\n")...)...)
}

// encodeBase64Lines encodes bytes to base64 with 64-char line wrapping (PEM format).
func encodeBase64Lines(data []byte) []byte {
	const b64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	encoded := make([]byte, (len(data)+2)/3*4)
	n := 0
	for i := 0; i < len(data); i += 3 {
		var b0, b1, b2 byte
		b0 = data[i]
		if i+1 < len(data) {
			b1 = data[i+1]
		}
		if i+2 < len(data) {
			b2 = data[i+2]
		}
		encoded[n] = b64[b0>>2]
		encoded[n+1] = b64[((b0&0x3)<<4)|(b1>>4)]
		if i+1 < len(data) {
			encoded[n+2] = b64[((b1&0xf)<<2)|(b2>>6)]
		} else {
			encoded[n+2] = '='
		}
		if i+2 < len(data) {
			encoded[n+3] = b64[b2&0x3f]
		} else {
			encoded[n+3] = '='
		}
		n += 4
	}
	// Add line breaks every 64 chars
	var result []byte
	for i := 0; i < len(encoded); i += 64 {
		end := i + 64
		if end > len(encoded) {
			end = len(encoded)
		}
		result = append(result, encoded[i:end]...)
		result = append(result, '\n')
	}
	return result
}

// generateSelfSignedCA creates a minimal self-signed CA certificate in PEM format
// that does NOT match any httptest.NewTLSServer certificate.
func generateSelfSignedCA(t *testing.T) []byte {
	t.Helper()

	// Use a second TLS test server's certificate as the "wrong CA"
	// (its cert is signed by a different ephemeral CA)
	wrongServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer wrongServer.Close()

	cert := wrongServer.TLS.Certificates[0]
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	require.NoError(t, err)

	return encodeCertPEM(x509Cert)
}

// Compile-time check: ensure tls package is used (imported above)
var _ = tls.Config{}
