package tokenexchange

import (
	"errors"
	"net/url"
	"strings"
)

// ResourceURI represents a normalized resource URI for RFC 8693 token exchange.
// URIs are normalized by removing trailing slashes for consistent matching.
type ResourceURI struct {
	value string
}

// NewResourceURI creates a new normalized ResourceURI from a string.
// Validates that the input is a valid absolute URL and normalizes by removing trailing slashes.
// Returns error if the URI is invalid.
func NewResourceURI(uri string) (*ResourceURI, error) {
	if uri == "" {
		return nil, errors.New("resource URI cannot be empty")
	}

	// Parse to validate it's a valid URL
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, errors.New("resource URI is not a valid URL")
	}

	// Ensure absolute URL (must have scheme and host)
	if parsed.Scheme == "" {
		return nil, errors.New("resource URI must include a scheme (e.g., https://)")
	}
	if parsed.Host == "" {
		return nil, errors.New("resource URI must include a host")
	}

	// Normalize: remove trailing slash
	normalized := strings.TrimSuffix(uri, "/")

	return &ResourceURI{
		value: normalized,
	}, nil
}

// Value returns the normalized URI string.
func (r *ResourceURI) Value() string {
	if r == nil {
		return ""
	}
	return r.value
}

// String returns the normalized URI string (implements Stringer interface).
func (r *ResourceURI) String() string {
	return r.Value()
}

// Equal compares two ResourceURIs for equality.
func (r *ResourceURI) Equal(other *ResourceURI) bool {
	if r == nil && other == nil {
		return true
	}
	if r == nil || other == nil {
		return false
	}
	return r.value == other.value
}

// Normalize returns the normalized (trailing slash removed) resource URI string.
// This function is idempotent - normalizing a normalized URI yields the same result.
// Example: "https://api.example.com/" → "https://api.example.com"
// Note: This function only removes trailing slashes; it does NOT validate the URI format.
// For validation, use NewResourceURI instead.
// This is useful for normalizing URIs before calling FindByProtectedResource:
//
//	normalizedURI := Normalize(requestURI)
//	service, err := repo.FindByProtectedResource(ctx, normalizedURI)
func Normalize(uri string) string {
	return strings.TrimSuffix(uri, "/")
}
