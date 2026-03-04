// Package helpers provides test utilities for E2E testing.
package helpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockJWKSServer provides a lightweight mock JWKS endpoint for JWT pre-authentication E2E testing.
// Unlike MockUpstreamOAuth2Server (which simulates a full OAuth2 server), this mock focuses
// solely on serving a JWKS endpoint for JWT signature verification in pre-auth scenarios.
//
// Design:
//   - Reuses GenerateTestRSAKeyPair() and GenerateJWKSFromPublicKey() from jwt_helpers.go
//   - Serves /.well-known/jwks.json with the generated JWKS set
//   - Provides private key for signing test JWTs
//   - Thread-safe request tracking for verification
//
// Usage:
//
//	jwksServer := helpers.NewMockJWKSServer()
//	defer jwksServer.Close()
//	config := fixtures.SignedJWTConfig(jwksServer.JWKSURL())
//	token, _ := jwksServer.SignJWT(claims)
type MockJWKSServer struct {
	Server *httptest.Server

	// RSA key pair for JWT signing/verification
	privateKeyPEM string
	publicKeyPEM  string
	jwksSet       map[string]interface{}

	// Request tracking
	mu          sync.RWMutex
	jwksCalled  bool
	callCount   int
	lastRequest *http.Request
}

// NewMockJWKSServer creates a new mock JWKS server with generated RSA key pair.
// The server starts immediately and listens on a random port.
// Caller must call Close() when done.
func NewMockJWKSServer() *MockJWKSServer {
	m := &MockJWKSServer{}

	// Generate RSA key pair
	privateKeyPEM, publicKeyPEM, err := GenerateTestRSAKeyPair()
	if err != nil {
		panic("MockJWKSServer: failed to generate RSA key pair: " + err.Error())
	}
	m.privateKeyPEM = privateKeyPEM
	m.publicKeyPEM = publicKeyPEM

	// Generate JWKS set from public key
	jwksSet, err := GenerateJWKSFromPublicKey(publicKeyPEM)
	if err != nil {
		panic("MockJWKSServer: failed to generate JWKS set: " + err.Error())
	}
	m.jwksSet = jwksSet

	// Create HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/jwks.json", m.handleJWKS)

	m.Server = httptest.NewServer(mux)
	return m
}

// Close shuts down the mock JWKS server.
func (m *MockJWKSServer) Close() {
	if m.Server != nil {
		m.Server.Close()
	}
}

// URL returns the base URL of the mock JWKS server (e.g., "http://127.0.0.1:PORT").
func (m *MockJWKSServer) URL() string {
	if m.Server != nil {
		return m.Server.URL
	}
	return ""
}

// JWKSURL returns the full URL to the JWKS endpoint.
// This is the value to use for authentication.jwt.jwks_uri in config.
func (m *MockJWKSServer) JWKSURL() string {
	return m.URL() + "/.well-known/jwks.json"
}

// PrivateKeyPEM returns the RSA private key in PEM format for signing test JWTs.
func (m *MockJWKSServer) PrivateKeyPEM() string {
	return m.privateKeyPEM
}

// PublicKeyPEM returns the RSA public key in PEM format.
func (m *MockJWKSServer) PublicKeyPEM() string {
	return m.publicKeyPEM
}

// SignJWT creates and signs a JWT with the server's private key.
// Convenience wrapper around helpers.SignTestJWT using this server's key pair.
func (m *MockJWKSServer) SignJWT(claims map[string]interface{}) (string, error) {
	return SignTestJWT(claims, m.privateKeyPEM)
}

// JWKSCalled returns whether the JWKS endpoint has been called.
func (m *MockJWKSServer) JWKSCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jwksCalled
}

// CallCount returns how many times the JWKS endpoint has been called.
func (m *MockJWKSServer) CallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount
}

// Reset clears request tracking state.
func (m *MockJWKSServer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jwksCalled = false
	m.callCount = 0
	m.lastRequest = nil
}

// handleJWKS serves the JWKS set at /.well-known/jwks.json.
func (m *MockJWKSServer) handleJWKS(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.jwksCalled = true
	m.callCount++
	m.lastRequest = r
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(m.jwksSet)
}
