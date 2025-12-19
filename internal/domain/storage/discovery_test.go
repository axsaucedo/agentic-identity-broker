package storage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// setupTestServer creates a TLS test server and configures the HTTP client to trust it
func setupTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
	t.Helper()
	server := httptest.NewTLSServer(handler)

	// Configure discovery to trust test server's self-signed cert
	SetHTTPClientForTesting(server.Client())

	cleanup := func() {
		server.Close()
		// Reset to default client
		SetHTTPClientForTesting(&http.Client{Timeout: 10 * time.Second})
	}

	return server, cleanup
}

func TestDiscoverOAuth2Endpoints_Success(t *testing.T) {
	// Create test server with valid OAuth2 metadata
	metadata := map[string]string{
		"token_endpoint":         "https://oauth.example.com/token",
		"authorization_endpoint": "https://oauth.example.com/authorize",
		"jwks_uri":               "https://oauth.example.com/jwks",
	}

	server, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Check request path
		if !strings.HasSuffix(r.URL.Path, "/.well-known/oauth-authorization-server") {
			t.Errorf("unexpected request path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	})
	defer cleanup()

	ctx := context.Background()
	endpoints, err := DiscoverOAuth2Endpoints(ctx, server.URL, nil)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if endpoints.TokenEndpoint != metadata["token_endpoint"] {
		t.Errorf("expected token_endpoint=%s, got=%s", metadata["token_endpoint"], endpoints.TokenEndpoint)
	}

	if endpoints.AuthorizeEndpoint != metadata["authorization_endpoint"] {
		t.Errorf("expected authorization_endpoint=%s, got=%s", metadata["authorization_endpoint"], endpoints.AuthorizeEndpoint)
	}
}

func TestDiscoverOAuth2Endpoints_WithMetadataURLOverride(t *testing.T) {
	metadata := map[string]string{
		"token_endpoint":         "https://oauth.example.com/token",
		"authorization_endpoint": "https://oauth.example.com/authorize",
		"jwks_uri":               "https://oauth.example.com/jwks",
	}

	// Server at custom path
	server, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Should receive request at custom path
		if r.URL.Path != "/custom/metadata" {
			t.Errorf("expected /custom/metadata, got: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	})
	defer cleanup()

	ctx := context.Background()
	customURL := server.URL + "/custom/metadata"
	endpoints, err := DiscoverOAuth2Endpoints(ctx, server.URL, &customURL)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if endpoints.TokenEndpoint != metadata["token_endpoint"] {
		t.Errorf("expected token_endpoint=%s, got=%s", metadata["token_endpoint"], endpoints.TokenEndpoint)
	}
}

func TestDiscoverOAuth2Endpoints_EmptyIssuerURI(t *testing.T) {
	ctx := context.Background()
	_, err := DiscoverOAuth2Endpoints(ctx, "", nil)

	if err == nil {
		t.Fatal("expected error for empty issuer_uri")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_InvalidIssuerURI(t *testing.T) {
	ctx := context.Background()
	_, err := DiscoverOAuth2Endpoints(ctx, "not-a-url", nil)

	if err == nil {
		t.Fatal("expected error for invalid issuer_uri")
	}

	// Error should mention HTTPS requirement since "not-a-url" has no scheme
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_NonHTTPSIssuer(t *testing.T) {
	ctx := context.Background()
	_, err := DiscoverOAuth2Endpoints(ctx, "http://oauth.example.com", nil)

	if err == nil {
		t.Fatal("expected error for non-HTTPS issuer")
	}

	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_InvalidMetadataURL(t *testing.T) {
	ctx := context.Background()
	invalidURL := "not-a-url"
	_, err := DiscoverOAuth2Endpoints(ctx, "https://oauth.example.com", &invalidURL)

	if err == nil {
		t.Fatal("expected error for invalid metadata URL")
	}

	// Error should mention HTTPS requirement since "not-a-url" has no scheme
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_NonHTTPSMetadataURL(t *testing.T) {
	ctx := context.Background()
	httpURL := "http://oauth.example.com/metadata"
	_, err := DiscoverOAuth2Endpoints(ctx, "https://oauth.example.com", &httpURL)

	if err == nil {
		t.Fatal("expected error for non-HTTPS metadata URL")
	}

	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_HTTPError(t *testing.T) {
	server, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer cleanup()

	ctx := context.Background()
	_, err := DiscoverOAuth2Endpoints(ctx, server.URL, nil)

	if err == nil {
		t.Fatal("expected error for HTTP 404")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_InvalidJSON(t *testing.T) {
	server, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	})
	defer cleanup()

	ctx := context.Background()
	_, err := DiscoverOAuth2Endpoints(ctx, server.URL, nil)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_MissingTokenEndpoint(t *testing.T) {
	metadata := map[string]string{
		"authorization_endpoint": "https://oauth.example.com/authorize",
		"jwks_uri":               "https://oauth.example.com/jwks",
	}

	server, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	})
	defer cleanup()

	ctx := context.Background()
	_, err := DiscoverOAuth2Endpoints(ctx, server.URL, nil)

	if err == nil {
		t.Fatal("expected error for missing token_endpoint")
	}

	if !strings.Contains(err.Error(), "token_endpoint") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_MissingAuthorizationEndpoint(t *testing.T) {
	metadata := map[string]string{
		"token_endpoint": "https://oauth.example.com/token",
		"jwks_uri":       "https://oauth.example.com/jwks",
	}

	server, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	})
	defer cleanup()

	ctx := context.Background()
	_, err := DiscoverOAuth2Endpoints(ctx, server.URL, nil)

	if err == nil {
		t.Fatal("expected error for missing authorization_endpoint")
	}

	if !strings.Contains(err.Error(), "authorization_endpoint") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_InvalidTokenEndpointURL(t *testing.T) {
	metadata := map[string]string{
		"token_endpoint":         "http://oauth.example.com/token", // HTTP not HTTPS
		"authorization_endpoint": "https://oauth.example.com/authorize",
		"jwks_uri":               "https://oauth.example.com/jwks",
	}

	server, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	})
	defer cleanup()

	ctx := context.Background()
	_, err := DiscoverOAuth2Endpoints(ctx, server.URL, nil)

	if err == nil {
		t.Fatal("expected error for invalid token_endpoint URL")
	}

	if !strings.Contains(err.Error(), "token_endpoint") || !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_InvalidAuthorizationEndpointURL(t *testing.T) {
	metadata := map[string]string{
		"token_endpoint":         "https://oauth.example.com/token",
		"authorization_endpoint": "http://oauth.example.com/authorize", // HTTP not HTTPS
		"jwks_uri":               "https://oauth.example.com/jwks",
	}

	server, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	})
	defer cleanup()

	ctx := context.Background()
	_, err := DiscoverOAuth2Endpoints(ctx, server.URL, nil)

	if err == nil {
		t.Fatal("expected error for invalid authorization_endpoint URL")
	}

	if !strings.Contains(err.Error(), "authorization_endpoint") || !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_ContextTimeout(t *testing.T) {
	// Server with slow response
	server, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	defer cleanup()

	// Context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := DiscoverOAuth2Endpoints(ctx, server.URL, nil)

	if err == nil {
		t.Fatal("expected error for context timeout")
	}

	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscoverOAuth2Endpoints_ContextCancellation(t *testing.T) {
	// Server with slow response
	server, cleanup := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	_, err := DiscoverOAuth2Endpoints(ctx, server.URL, nil)

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}

	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateEndpointURL_Valid(t *testing.T) {
	err := validateEndpointURL("https://oauth.example.com/token")
	if err != nil {
		t.Errorf("expected no error for valid HTTPS URL, got: %v", err)
	}
}

func TestValidateEndpointURL_Empty(t *testing.T) {
	err := validateEndpointURL("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateEndpointURL_Invalid(t *testing.T) {
	err := validateEndpointURL("not-a-url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestValidateEndpointURL_NonHTTPS(t *testing.T) {
	err := validateEndpointURL("http://oauth.example.com/token")
	if err == nil {
		t.Fatal("expected error for non-HTTPS URL")
	}
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("unexpected error message: %v", err)
	}
}
