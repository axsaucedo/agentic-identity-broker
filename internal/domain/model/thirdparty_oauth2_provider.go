package model

import (
	"errors"
	"time"
)

// ThirdpartyOAuth2ProviderEntity is the domain entity for an external OAuth2 provider that
// agents can access on behalf of users (e.g., GitHub, Google, Databricks).
//
// The Secret field uses the Secret value object with two mutually exclusive states:
// - Plaintext state: used when creating/updating a provider (before encryption)
// - Encrypted state: used when the entity was loaded from storage (after decryption by a domain service)
//
// Encryption and decryption happens exclusively in domain services, not in this entity.
type ThirdpartyOAuth2ProviderEntity struct {
	ID                  string
	DisplayName         string
	ClientID            string
	Secret              Secret
	IssuerURI           string
	Discovery           DiscoveryConfig
	Endpoints           OAuth2Endpoints
	Scopes              []OAuthScope
	ProtectedResources  []string
	ServiceRequirements []ServiceRequirement
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Validate performs basic validation on the entity.
// It checks that required fields are present and the secret is in a valid state.
func (e *ThirdpartyOAuth2ProviderEntity) Validate() error {
	if e.ID == "" {
		return errors.New("provider ID cannot be empty")
	}
	if e.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if len(e.DisplayName) > 255 {
		return errors.New("display_name exceeds 255 characters")
	}
	if e.ClientID == "" {
		return errors.New("client_id is required")
	}

	// Secret must be in a valid state (either plaintext or encrypted)
	if e.Secret.IsPlaintext() {
		if _, err := e.Secret.GetPlaintext(); err != nil {
			return err
		}
	} else if e.Secret.IsEncrypted() {
		if _, err := e.Secret.GetCiphertext(); err != nil {
			return err
		}
	}

	if e.IssuerURI == "" {
		return errors.New("issuer_uri is required")
	}

	return nil
}

// RedactedCopy returns a deep copy of the entity with Secret replaced by a plaintext
// "REDACTED" value. Use for API responses and logs to comply with SR-003.
func (e *ThirdpartyOAuth2ProviderEntity) RedactedCopy() *ThirdpartyOAuth2ProviderEntity {
	if e == nil {
		return nil
	}
	c := e.Copy()
	c.Secret = NewPlaintextSecret("REDACTED")
	return c
}

// Copy creates a deep copy of the entity to prevent external mutation.
// The Secret is copied by value (immutable), with internal slices independently copied.
func (e *ThirdpartyOAuth2ProviderEntity) Copy() *ThirdpartyOAuth2ProviderEntity {
	if e == nil {
		return nil
	}

	result := &ThirdpartyOAuth2ProviderEntity{
		ID:          e.ID,
		DisplayName: e.DisplayName,
		ClientID:    e.ClientID,
		Secret:      e.Secret, // Value type; internal ciphertext slice independently copied by Secret
		IssuerURI:   e.IssuerURI,
		Discovery: DiscoveryConfig{
			EnableDiscovery: e.Discovery.EnableDiscovery,
		},
		Endpoints: OAuth2Endpoints{
			TokenEndpoint:     e.Endpoints.TokenEndpoint,
			AuthorizeEndpoint: e.Endpoints.AuthorizeEndpoint,
		},
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	// Deep copy optional MetadataURL pointer
	if e.Discovery.MetadataURL != nil {
		metadataURL := *e.Discovery.MetadataURL
		result.Discovery.MetadataURL = &metadataURL
	}

	// Deep copy slices
	if e.Scopes != nil {
		result.Scopes = make([]OAuthScope, len(e.Scopes))
		for i, scope := range e.Scopes {
			result.Scopes[i] = scope.Copy()
		}
	}

	if e.ProtectedResources != nil {
		result.ProtectedResources = make([]string, len(e.ProtectedResources))
		copy(result.ProtectedResources, e.ProtectedResources)
	}

	if e.ServiceRequirements != nil {
		result.ServiceRequirements = make([]ServiceRequirement, len(e.ServiceRequirements))
		for i, sr := range e.ServiceRequirements {
			result.ServiceRequirements[i] = sr.Copy()
		}
	}

	return result
}
