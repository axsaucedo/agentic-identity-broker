package enduser

import (
	"encoding/json"
	"net/http"
)

// DiscoveryHandler serves the OAuth2 Authorization Server Metadata (RFC 8414).
type DiscoveryHandler struct {
	issuerURI string
}

// NewDiscoveryHandler creates a new DiscoveryHandler.
func NewDiscoveryHandler(issuerURI string) *DiscoveryHandler {
	return &DiscoveryHandler{issuerURI: issuerURI}
}

// discoveryResponse represents the OAuth2 Authorization Server Metadata document.
type discoveryResponse struct {
	Issuer                        string   `json:"issuer"`
	AuthorizationEndpoint         string   `json:"authorization_endpoint"`
	TokenEndpoint                 string   `json:"token_endpoint"`
	JWKSURI                       string   `json:"jwks_uri"`
	ResponseTypesSupported        []string `json:"response_types_supported"`
	GrantTypesSupported           []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
}

// ServeDiscovery returns the OAuth2 server metadata as JSON.
// GET /.well-known/oauth-authorization-server
func (h *DiscoveryHandler) ServeDiscovery(w http.ResponseWriter, r *http.Request) {
	resp := discoveryResponse{
		Issuer:                        h.issuerURI,
		AuthorizationEndpoint:         h.issuerURI + "/oauth2/authorize",
		TokenEndpoint:                 h.issuerURI + "/oauth2/token",
		JWKSURI:                       h.issuerURI + "/oauth2/jwks.json",
		ResponseTypesSupported:        []string{"code"},
		GrantTypesSupported:           []string{"authorization_code", "client_credentials"},
		CodeChallengeMethodsSupported: []string{"S256"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
