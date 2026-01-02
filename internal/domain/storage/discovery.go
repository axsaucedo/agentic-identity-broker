package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

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
//
// Returns:
//   - OAuth2Endpoints with token_endpoint and authorization_endpoint
//   - Error if discovery fails, network timeout, invalid JSON, or missing required fields
func DiscoverOAuth2Endpoints(ctx context.Context, issuerURI string, metadataURL *string) (*OAuth2Endpoints, error) {
	// Validate issuer URI
	if issuerURI == "" {
		return nil, errors.New("issuer_uri cannot be empty")
	}

	// Validate issuer_uri is HTTPS (or HTTP for localhost)
	if !isAllowedScheme(issuerURI) {
		return nil, errors.New("issuer_uri must be a valid HTTPS URL (HTTP allowed only for localhost)")
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

	// Validate discovery URL (HTTPS or HTTP for localhost)
	if !isAllowedScheme(discoveryURL) {
		return nil, errors.New("discovery URL must be a valid HTTPS URL (HTTP allowed only for localhost)")
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
	defer resp.Body.Close()

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

	// Validate endpoint URLs are valid HTTPS URLs
	if err := validateEndpointURL(metadata.TokenEndpoint); err != nil {
		return nil, fmt.Errorf("invalid token_endpoint: %w", err)
	}

	if err := validateEndpointURL(metadata.AuthorizationEndpoint); err != nil {
		return nil, fmt.Errorf("invalid authorization_endpoint: %w", err)
	}

	// Return discovered endpoints
	return &OAuth2Endpoints{
		TokenEndpoint:     metadata.TokenEndpoint,
		AuthorizeEndpoint: metadata.AuthorizationEndpoint,
	}, nil
}

// validateEndpointURL validates that a URL is a valid HTTPS URL (or HTTP for localhost).
func validateEndpointURL(urlStr string) error {
	if urlStr == "" {
		return errors.New("URL cannot be empty")
	}

	// Use isAllowedScheme to allow HTTP for localhost
	if !isAllowedScheme(urlStr) {
		return errors.New("URL must be a valid HTTPS URL (HTTP allowed only for localhost)")
	}

	return nil
}
