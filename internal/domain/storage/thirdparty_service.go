package storage

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// ThirdpartyOAuth2Service represents an external OAuth2 provider that agents can access
// on behalf of users (e.g., GitHub, Google, Databricks).
type ThirdpartyOAuth2Service struct {
	ID           string          `json:"id" db:"id"`
	DisplayName  string          `json:"display_name" db:"display_name"`
	ClientID     string          `json:"client_id" db:"client_id"`
	ClientSecret string          `json:"-" db:"client_secret_encrypted"` // Never serialized to JSON
	IssuerURI    string          `json:"issuer_uri" db:"issuer_uri"`
	Discovery    DiscoveryConfig `json:"discovery" db:"-"`
	Endpoints    OAuth2Endpoints `json:"endpoints" db:"-"`
	Scopes       []OAuthScope    `json:"scopes" db:"scopes"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// DiscoveryConfig holds OAuth2 endpoint discovery configuration.
type DiscoveryConfig struct {
	EnableDiscovery bool    `json:"enable_discovery" db:"enable_discovery"`
	MetadataURL     *string `json:"metadata_url,omitempty" db:"metadata_url"`
}

// OAuth2Endpoints holds the OAuth2 endpoints for a service.
type OAuth2Endpoints struct {
	TokenEndpoint     string `json:"token_endpoint" db:"token_endpoint"`
	AuthorizeEndpoint string `json:"authorize_endpoint" db:"authorize_endpoint"`
}

// isAllowedScheme checks if a URL uses an allowed scheme.
// HTTPS is always allowed. HTTP is allowed for localhost addresses in development
// or when HTTPS validation is skipped (dev/test mode).
func isAllowedScheme(urlStr string, skipHTTPSValidation bool) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// HTTPS is always allowed
	if parsed.Scheme == "https" {
		return true
	}

	// HTTP is allowed if HTTPS validation is explicitly skipped (dev/test only)
	if skipHTTPSValidation && parsed.Scheme == "http" {
		return true
	}

	// HTTP is allowed for localhost/127.0.0.1 (development)
	if parsed.Scheme == "http" {
		hostname := parsed.Hostname()
		return hostname == "localhost" || hostname == "127.0.0.1"
	}

	return false
}

// ValidateWith performs validation on the ThirdpartyOAuth2Service entity with optional HTTPS validation skip.
// skipHTTPSValidation allows HTTP URLs in dev/test mode (WARNING: insecure, dev-only).
func (s *ThirdpartyOAuth2Service) ValidateWith(skipHTTPSValidation bool) error {
	// Required fields
	if s.ID == "" {
		return errors.New("service ID cannot be empty")
	}
	if s.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if len(s.DisplayName) > 255 {
		return fmt.Errorf("display_name exceeds 255 characters (got %d)", len(s.DisplayName))
	}
	if s.ClientID == "" {
		return errors.New("client_id is required")
	}
	if s.ClientSecret == "" {
		return errors.New("client_secret is required")
	}
	if s.IssuerURI == "" {
		return errors.New("issuer_uri is required")
	}

	// Validate issuer_uri is HTTPS (or HTTP for localhost/dev mode)
	if !isAllowedScheme(s.IssuerURI, skipHTTPSValidation) {
		return errors.New("issuer_uri must be a valid HTTPS URL (HTTP allowed only for localhost in dev mode)")
	}

	// If discovery disabled, endpoints are required
	if !s.Discovery.EnableDiscovery {
		if s.Endpoints.TokenEndpoint == "" {
			return errors.New("token_endpoint is required when discovery is disabled")
		}
		if s.Endpoints.AuthorizeEndpoint == "" {
			return errors.New("authorize_endpoint is required when discovery is disabled")
		}
	}

	// Validate metadata_url if provided (HTTPS or HTTP for localhost/dev mode)
	if s.Discovery.MetadataURL != nil {
		if !isAllowedScheme(*s.Discovery.MetadataURL, skipHTTPSValidation) {
			return errors.New("metadata_url must be a valid HTTPS URL (HTTP allowed only for localhost in dev mode)")
		}
	}

	// At least one scope required
	if len(s.Scopes) == 0 {
		return errors.New("at least one scope is required")
	}

	// Validate each scope
	for i, scope := range s.Scopes {
		if err := scope.Validate(); err != nil {
			return fmt.Errorf("scope %d: %w", i, err)
		}
	}

	return nil
}

// Validate performs validation on the ThirdpartyOAuth2Service entity (uses default strict HTTPS validation).
func (s *ThirdpartyOAuth2Service) Validate() error {
	return s.ValidateWith(false)
}

// ValidateForCreateWith validates a service before creation with optional HTTPS validation skip.
// skipHTTPSValidation allows HTTP URLs in dev/test mode (WARNING: insecure, dev-only).
func (s *ThirdpartyOAuth2Service) ValidateForCreateWith(skipHTTPSValidation bool) error {
	if s.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if len(s.DisplayName) > 255 {
		return fmt.Errorf("display_name exceeds 255 characters (got %d)", len(s.DisplayName))
	}
	if s.ClientID == "" {
		return errors.New("client_id is required")
	}
	if s.ClientSecret == "" {
		return errors.New("client_secret is required")
	}
	if s.IssuerURI == "" {
		return errors.New("issuer_uri is required")
	}

	// Validate issuer_uri is HTTPS (or HTTP for localhost/dev mode)
	if !isAllowedScheme(s.IssuerURI, skipHTTPSValidation) {
		return errors.New("issuer_uri must be a valid HTTPS URL (HTTP allowed only for localhost in dev mode)")
	}

	if !s.Discovery.EnableDiscovery {
		if s.Endpoints.TokenEndpoint == "" {
			return errors.New("token_endpoint is required when discovery is disabled")
		}
		if s.Endpoints.AuthorizeEndpoint == "" {
			return errors.New("authorize_endpoint is required when discovery is disabled")
		}
	}

	// Validate metadata_url if provided (HTTPS or HTTP for localhost/dev mode)
	if s.Discovery.MetadataURL != nil {
		if !isAllowedScheme(*s.Discovery.MetadataURL, skipHTTPSValidation) {
			return errors.New("metadata_url must be a valid HTTPS URL (HTTP allowed only for localhost in dev mode)")
		}
	}

	if len(s.Scopes) == 0 {
		return errors.New("at least one scope is required")
	}

	for i, scope := range s.Scopes {
		if err := scope.Validate(); err != nil {
			return fmt.Errorf("scope %d: %w", i, err)
		}
	}

	return nil
}

// RedactedCopy creates a copy of the service with client_secret redacted.
// This is used for API responses to comply with SR-003.
func (s *ThirdpartyOAuth2Service) RedactedCopy() *ThirdpartyOAuth2Service {
	if s == nil {
		return nil
	}

	copy := s.Copy()
	copy.ClientSecret = "REDACTED"
	return copy
}

// Copy creates a deep copy of the ThirdpartyOAuth2Service.
func (s *ThirdpartyOAuth2Service) Copy() *ThirdpartyOAuth2Service {
	if s == nil {
		return nil
	}

	copy := &ThirdpartyOAuth2Service{
		ID:           s.ID,
		DisplayName:  s.DisplayName,
		ClientID:     s.ClientID,
		ClientSecret: s.ClientSecret,
		IssuerURI:    s.IssuerURI,
		Discovery: DiscoveryConfig{
			EnableDiscovery: s.Discovery.EnableDiscovery,
		},
		Endpoints: OAuth2Endpoints{
			TokenEndpoint:     s.Endpoints.TokenEndpoint,
			AuthorizeEndpoint: s.Endpoints.AuthorizeEndpoint,
		},
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}

	if s.Discovery.MetadataURL != nil {
		metadataURL := *s.Discovery.MetadataURL
		copy.Discovery.MetadataURL = &metadataURL
	}

	// Deep copy scopes
	copy.Scopes = make([]OAuthScope, len(s.Scopes))
	for i, scope := range s.Scopes {
		copy.Scopes[i] = scope.Copy()
	}

	return copy
}

// Scan implements sql.Scanner for JSONB deserialization of scopes.
func (s *OAuthScopeArray) Scan(value interface{}) error {
	if value == nil {
		*s = []OAuthScope{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan OAuthScopeArray: expected []byte")
	}

	var scopes []OAuthScope
	if err := json.Unmarshal(bytes, &scopes); err != nil {
		return fmt.Errorf("failed to unmarshal OAuthScopeArray: %w", err)
	}

	*s = scopes
	return nil
}

// Value implements driver.Valuer for JSONB serialization of scopes.
func (s OAuthScopeArray) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]OAuthScope{})
	}
	return json.Marshal(s)
}

// OAuthScopeArray is a custom type for handling JSONB arrays in PostgreSQL.
type OAuthScopeArray []OAuthScope
