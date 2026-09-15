package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
)

// isAllowedScheme checks if a URL uses an allowed scheme.
// HTTPS is always allowed. HTTP is allowed for localhost addresses in development
// or when HTTPS validation is skipped (dev/test mode).
func isAllowedScheme(urlStr string, skipHTTPSValidation bool) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	if parsed.Scheme == "https" {
		return true
	}

	if skipHTTPSValidation && parsed.Scheme == "http" {
		return true
	}

	if parsed.Scheme == "http" {
		hostname := parsed.Hostname()
		return hostname == "localhost" || hostname == "127.0.0.1"
	}

	return false
}

// defaultHTTPClient is the HTTP client used for discovery requests.
// Can be overridden for testing via SetHTTPClientForTesting.
var defaultHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

// SetHTTPClientForTesting allows tests to provide a custom HTTP client.
// This is primarily used to configure TLS trust for test servers.
func SetHTTPClientForTesting(client *http.Client) {
	defaultHTTPClient = client
}

// DiscoverOAuth2Endpoints performs OAuth2 endpoint discovery per RFC 8414.
// It attempts to retrieve the OAuth 2.0 Authorization Server Metadata from the issuer's
// well-known endpoint. If metadataURL is provided, it is used instead of constructing
// the well-known path.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout
//   - issuerURI: The OAuth2 issuer URI (must be HTTPS)
//   - metadataURL: Optional override URL for metadata endpoint
//   - skipHTTPSValidation: If true, allows HTTP URLs in dev/test mode
//
// Returns:
//   - OAuth2Endpoints with token_endpoint and authorization_endpoint
//   - Error if discovery fails, network timeout, invalid JSON, or missing required fields
func DiscoverOAuth2Endpoints(ctx context.Context, issuerURI string, metadataURL *string, skipHTTPSValidation bool) (*model.OAuth2Endpoints, error) {
	// Validate issuer URI
	if issuerURI == "" {
		return nil, errors.New("issuer_uri cannot be empty")
	}

	// Validate issuer_uri is HTTPS (or HTTP for localhost/dev mode)
	if !isAllowedScheme(issuerURI, skipHTTPSValidation) {
		return nil, errors.New("issuer_uri must be a valid HTTPS URL (HTTP allowed only for localhost in dev mode)")
	}

	// Determine discovery URL
	discoveryURL := ""
	if metadataURL != nil && *metadataURL != "" {
		// Use explicit metadata URL override
		discoveryURL = *metadataURL
	} else {
		// Construct well-known path per RFC 8414 Section 3
		// https://example.com/.well-known/oauth-authorization-server
		discoveryURL = fmt.Sprintf("%s/.well-known/oauth-authorization-server", issuerURI)
	}

	// Validate discovery URL (HTTPS or HTTP for localhost/dev mode)
	if !isAllowedScheme(discoveryURL, skipHTTPSValidation) {
		return nil, errors.New("discovery URL must be a valid HTTPS URL (HTTP allowed only for localhost in dev mode)")
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	// Set Accept header for JSON
	req.Header.Set("Accept", "application/json")

	// Execute request using default client (or test client if configured)
	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		// Check for context cancellation or timeout
		if ctx.Err() != nil {
			return nil, fmt.Errorf("discovery request cancelled: %w", ctx.Err())
		}
		return nil, fmt.Errorf("discovery request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery request returned HTTP %d", resp.StatusCode)
	}

	// Parse JSON response
	var metadata struct {
		TokenEndpoint         string `json:"token_endpoint"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		JWKsURI               string `json:"jwks_uri"`
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to parse discovery response: %w", err)
	}

	// Validate required fields are present
	if metadata.TokenEndpoint == "" {
		return nil, errors.New("discovery response missing required field: token_endpoint")
	}

	if metadata.AuthorizationEndpoint == "" {
		return nil, errors.New("discovery response missing required field: authorization_endpoint")
	}

	// Validate endpoint URLs are valid HTTPS URLs (or HTTP in dev mode)
	if err := validateEndpointURL(metadata.TokenEndpoint, skipHTTPSValidation); err != nil {
		return nil, fmt.Errorf("invalid token_endpoint: %w", err)
	}

	if err := validateEndpointURL(metadata.AuthorizationEndpoint, skipHTTPSValidation); err != nil {
		return nil, fmt.Errorf("invalid authorization_endpoint: %w", err)
	}

	// Validate jwks_uri when present (optional per RFC 8414, but must be a valid URL if set)
	if metadata.JWKsURI != "" {
		if err := validateEndpointURL(metadata.JWKsURI, skipHTTPSValidation); err != nil {
			return nil, fmt.Errorf("invalid jwks_uri: %w", err)
		}
	}

	// Return discovered endpoints
	return &model.OAuth2Endpoints{
		TokenEndpoint:     metadata.TokenEndpoint,
		AuthorizeEndpoint: metadata.AuthorizationEndpoint,
		JWKsURI:           metadata.JWKsURI,
	}, nil
}

// validateEndpointURL validates that a URL is a valid HTTPS URL (or HTTP for localhost/dev mode).
func validateEndpointURL(urlStr string, skipHTTPSValidation bool) error {
	if urlStr == "" {
		return errors.New("URL cannot be empty")
	}

	// Use isAllowedScheme with configurable validation
	if !isAllowedScheme(urlStr, skipHTTPSValidation) {
		return errors.New("URL must be a valid HTTPS URL (HTTP allowed only for localhost in dev mode)")
	}

	return nil
}
