package helpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

// RotatingJWKSUpstream serves stable RFC 8414 metadata while allowing tests to
// swap the published JWKS kid at runtime without changing the server URL.
type RotatingJWKSUpstream struct {
	Server       *httptest.Server
	publicKeyPEM string

	mu  sync.RWMutex
	kid string
}

// NewRotatingJWKSUpstream creates a mock upstream whose JWKS response can be
// updated in-place via SetKID.
func NewRotatingJWKSUpstream(initialKid string) *RotatingJWKSUpstream {
	_, publicKeyPEM, _ := sharedKeyPair()

	m := &RotatingJWKSUpstream{
		publicKeyPEM: publicKeyPEM,
		kid:          initialKid,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-authorization-server", m.handleMetadata)
	mux.HandleFunc("/.well-known/openid-configuration", m.handleMetadata)
	mux.HandleFunc("/.well-known/jwks.json", m.handleJWKS)
	m.Server = httptest.NewServer(mux)
	return m
}

func (m *RotatingJWKSUpstream) URL() string { return m.Server.URL }
func (m *RotatingJWKSUpstream) Close()      { m.Server.Close() }

// SetKID changes the kid served by the upstream JWKS endpoint for subsequent
// refreshes.
func (m *RotatingJWKSUpstream) SetKID(kid string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kid = kid
}

func (m *RotatingJWKSUpstream) currentKID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.kid
}

func (m *RotatingJWKSUpstream) handleMetadata(w http.ResponseWriter, _ *http.Request) {
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

func (m *RotatingJWKSUpstream) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	jwksSet, _ := GenerateJWKSFromPublicKey(m.publicKeyPEM)
	keys := jwksSet["keys"].([]map[string]interface{})
	keys[0]["kid"] = m.currentKID()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=1")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jwksSet)
}
