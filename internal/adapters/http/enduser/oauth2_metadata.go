package enduser

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// OAuth2MetadataHandler handles OAuth2 metadata endpoint requests (.well-known/oauth-authorization-server)
type OAuth2MetadataHandler struct {
	Service ports.OAuth2Service
}

// ServeHTTP implements http.Handler for the metadata endpoint
func (h *OAuth2MetadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Service == nil {
		http.Error(w, "OAuth2 authorization server not configured", http.StatusServiceUnavailable)
		return
	}

	// Only allow GET requests
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate metadata from service
	metadata, err := h.Service.GenerateMetadata(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	// Metadata can be cached since server configuration changes rarely
	w.Header().Set("Cache-Control", "max-age=3600")

	// Encode and send metadata response
	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		_, _ = fmt.Fprintf(w, "error encoding metadata: %v", err)
		return
	}
}
