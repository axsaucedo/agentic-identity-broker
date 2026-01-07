package enduser

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OAuth2TokenHandler handles OAuth2 token endpoint requests (proxy to upstream)
type OAuth2TokenHandler struct {
	UpstreamTokenURL string
	Client           *http.Client
}

// ServeHTTP implements http.Handler for the token endpoint
func (h *OAuth2TokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Verify request method
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate Content-Type (allow charset parameter)
	// OAuth 2.0 token endpoint must accept application/x-www-form-urlencoded per RFC 6749 Section 4.1.3
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		http.Error(w, "invalid Content-Type: expected application/x-www-form-urlencoded", http.StatusBadRequest)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate body is not empty
	if len(body) == 0 {
		http.Error(w, "request body cannot be empty", http.StatusBadRequest)
		return
	}

	// Create upstream request
	upstreamReq, err := http.NewRequest("POST", h.UpstreamTokenURL, strings.NewReader(string(body)))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create upstream request: %v", err), http.StatusInternalServerError)
		return
	}

	// Copy headers from client request to upstream request, filtering hop-by-hop headers
	for key, values := range r.Header {
		// Skip hop-by-hop headers
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			upstreamReq.Header.Add(key, value)
		}
	}

	// Make upstream request
	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}

	upstreamResp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to contact upstream server: %v", err), http.StatusBadGateway)
		return
	}
	defer upstreamResp.Body.Close()

	// Copy response headers from upstream to client, filtering hop-by-hop headers
	for key, values := range upstreamResp.Header {
		// Skip hop-by-hop headers
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy response status code
	w.WriteHeader(upstreamResp.StatusCode)

	// Stream response body from upstream
	if _, err := io.Copy(w, upstreamResp.Body); err != nil {
		fmt.Fprintf(w, "error streaming response: %v", err)
	}
}

// isHopByHopHeader returns true if the header is a hop-by-hop header per RFC 7230
func isHopByHopHeader(headerName string) bool {
	// Normalize to lowercase for comparison
	header := strings.ToLower(headerName)

	hopByHopHeaders := map[string]bool{
		"connection":          true,
		"keep-alive":          true,
		"proxy-authenticate":  true,
		"proxy-authorization": true,
		"te":                  true,
		"trailers":            true,
		"transfer-encoding":   true,
		"upgrade":             true,
	}

	return hopByHopHeaders[header]
}

// NewOAuth2TokenHandler creates a new token handler
func NewOAuth2TokenHandler(upstreamTokenURL string, client *http.Client) *OAuth2TokenHandler {
	return &OAuth2TokenHandler{
		UpstreamTokenURL: upstreamTokenURL,
		Client:           client,
	}
}
