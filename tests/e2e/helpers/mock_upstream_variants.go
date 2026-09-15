package helpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// MockUpstreamWithBrokenJWKS serves valid OAuth2 metadata (with jwks_uri) but returns
// HTTP 500 for the actual JWKS endpoint. This simulates an upstream that is discoverable
// but whose key material is unavailable at runtime.
type MockUpstreamWithBrokenJWKS struct {
	Server *httptest.Server
}

// NewMockUpstreamWithBrokenJWKS creates and starts a mock server.
func NewMockUpstreamWithBrokenJWKS() *MockUpstreamWithBrokenJWKS {
	m := &MockUpstreamWithBrokenJWKS{}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-authorization-server", m.handleMetadata)
	mux.HandleFunc("/.well-known/openid-configuration", m.handleMetadata)
	mux.HandleFunc("/.well-known/jwks.json", m.handleBrokenJWKS)
	m.Server = httptest.NewServer(mux)
	return m
}

func (m *MockUpstreamWithBrokenJWKS) URL() string { return m.Server.URL }
func (m *MockUpstreamWithBrokenJWKS) Close()      { m.Server.Close() }

func (m *MockUpstreamWithBrokenJWKS) handleMetadata(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	metadata := map[string]interface{}{
		"issuer":                   m.Server.URL,
		"authorization_endpoint":   m.Server.URL + "/oauth/authorize",
		"token_endpoint":           m.Server.URL + "/oauth/token",
		"jwks_uri":                 m.Server.URL + "/.well-known/jwks.json",
		"response_types_supported": []string{"code"},
		"grant_types_supported":    []string{"authorization_code"},
	}
	_ = json.NewEncoder(w).Encode(metadata)
}

func (m *MockUpstreamWithBrokenJWKS) handleBrokenJWKS(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

// MockUpstreamWithMalformedJWKS serves valid OAuth2 metadata but returns HTTP 200
// with invalid JSON for the JWKS endpoint.
type MockUpstreamWithMalformedJWKS struct {
	Server *httptest.Server
}

// NewMockUpstreamWithMalformedJWKS creates and starts a mock server.
func NewMockUpstreamWithMalformedJWKS() *MockUpstreamWithMalformedJWKS {
	m := &MockUpstreamWithMalformedJWKS{}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-authorization-server", m.handleMetadata)
	mux.HandleFunc("/.well-known/openid-configuration", m.handleMetadata)
	mux.HandleFunc("/.well-known/jwks.json", m.handleMalformedJWKS)
	m.Server = httptest.NewServer(mux)
	return m
}

func (m *MockUpstreamWithMalformedJWKS) URL() string { return m.Server.URL }
func (m *MockUpstreamWithMalformedJWKS) Close()      { m.Server.Close() }

func (m *MockUpstreamWithMalformedJWKS) handleMetadata(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	metadata := map[string]interface{}{
		"issuer":                   m.Server.URL,
		"authorization_endpoint":   m.Server.URL + "/oauth/authorize",
		"token_endpoint":           m.Server.URL + "/oauth/token",
		"jwks_uri":                 m.Server.URL + "/.well-known/jwks.json",
		"response_types_supported": []string{"code"},
		"grant_types_supported":    []string{"authorization_code"},
	}
	_ = json.NewEncoder(w).Encode(metadata)
}

func (m *MockUpstreamWithMalformedJWKS) handleMalformedJWKS(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("not valid json {"))
}

// MockUpstreamWithKid serves valid OAuth2 metadata and a JWKS endpoint containing
// a single RSA public key with a configurable kid. Used for testing kid conflicts.
type MockUpstreamWithKid struct {
	Server *httptest.Server
	kid    string
}

// NewMockUpstreamWithKid creates and starts a mock server serving a key with the given kid.
func NewMockUpstreamWithKid(kid string) *MockUpstreamWithKid {
	m := &MockUpstreamWithKid{kid: kid}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-authorization-server", m.handleMetadata)
	mux.HandleFunc("/.well-known/openid-configuration", m.handleMetadata)
	mux.HandleFunc("/.well-known/jwks.json", m.handleJWKS)
	m.Server = httptest.NewServer(mux)
	return m
}

func (m *MockUpstreamWithKid) URL() string { return m.Server.URL }
func (m *MockUpstreamWithKid) Close()      { m.Server.Close() }

func (m *MockUpstreamWithKid) handleMetadata(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	metadata := map[string]interface{}{
		"issuer":                   m.Server.URL,
		"authorization_endpoint":   m.Server.URL + "/oauth/authorize",
		"token_endpoint":           m.Server.URL + "/oauth/token",
		"jwks_uri":                 m.Server.URL + "/.well-known/jwks.json",
		"response_types_supported": []string{"code"},
		"grant_types_supported":    []string{"authorization_code"},
	}
	_ = json.NewEncoder(w).Encode(metadata)
}

func (m *MockUpstreamWithKid) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	_, publicKeyPEM, _ := GenerateTestRSAKeyPair()
	jwksSet, _ := GenerateJWKSFromPublicKey(publicKeyPEM)

	// Override the kid in the generated key set
	keys := jwksSet["keys"].([]map[string]interface{})
	keys[0]["kid"] = m.kid

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jwksSet)
}
