package storage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThirdpartyOAuth2Service_Validate(t *testing.T) {
	metadataURL := "https://github.com/.well-known/oauth-authorization-server"

	tests := []struct {
		name    string
		service *ThirdpartyOAuth2Service
		wantErr string
	}{
		{
			name: "valid service with discovery enabled",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Discovery: DiscoveryConfig{
					EnableDiscovery: true,
					MetadataURL:     &metadataURL,
				},
				Endpoints: OAuth2Endpoints{
					TokenEndpoint:     "https://github.com/login/oauth/access_token",
					AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
				},
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "",
		},
		{
			name: "valid service with discovery disabled",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Discovery: DiscoveryConfig{
					EnableDiscovery: false,
				},
				Endpoints: OAuth2Endpoints{
					TokenEndpoint:     "https://github.com/login/oauth/access_token",
					AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
				},
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "",
		},
		{
			name: "missing ID",
			service: &ThirdpartyOAuth2Service{
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "service ID cannot be empty",
		},
		{
			name: "missing display_name",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "display_name is required",
		},
		{
			name: "display_name too long",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  strings.Repeat("a", 256),
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "display_name exceeds 255 characters",
		},
		{
			name: "missing client_id",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "client_id is required",
		},
		{
			name: "missing client_secret",
			service: &ThirdpartyOAuth2Service{
				ID:          "650e8400-e29b-41d4-a716-446655440001",
				DisplayName: "GitHub",
				ClientID:    "Iv1.abcd1234",
				IssuerURI:   "https://github.com",
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "client_secret is required",
		},
		{
			name: "missing issuer_uri",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "issuer_uri is required",
		},
		{
			name: "issuer_uri not HTTPS",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "http://github.com",
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "issuer_uri must be a valid HTTPS URL",
		},
		{
			name: "discovery disabled but missing token_endpoint",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Discovery: DiscoveryConfig{
					EnableDiscovery: false,
				},
				Endpoints: OAuth2Endpoints{
					AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
				},
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "token_endpoint is required when discovery is disabled",
		},
		{
			name: "discovery disabled but missing authorize_endpoint",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Discovery: DiscoveryConfig{
					EnableDiscovery: false,
				},
				Endpoints: OAuth2Endpoints{
					TokenEndpoint: "https://github.com/login/oauth/access_token",
				},
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "authorize_endpoint is required when discovery is disabled",
		},
		{
			name: "invalid metadata_url",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Discovery: DiscoveryConfig{
					EnableDiscovery: true,
					MetadataURL:     stringPtr("http://insecure.com"),
				},
				Scopes: []OAuthScope{
					{ScopeValue: "repo", Description: "Full repository access"},
				},
			},
			wantErr: "metadata_url must be a valid HTTPS URL",
		},
		{
			name: "no scopes",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Discovery: DiscoveryConfig{
					EnableDiscovery: true,
				},
				Scopes: []OAuthScope{},
			},
			wantErr: "at least one scope is required",
		},
		{
			name: "scope missing scope_value",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Discovery: DiscoveryConfig{
					EnableDiscovery: true,
				},
				Scopes: []OAuthScope{
					{Description: "Full repository access"},
				},
			},
			wantErr: "scope 0: scope_value is required",
		},
		{
			name: "scope missing description",
			service: &ThirdpartyOAuth2Service{
				ID:           "650e8400-e29b-41d4-a716-446655440001",
				DisplayName:  "GitHub",
				ClientID:     "Iv1.abcd1234",
				ClientSecret: "secret123",
				IssuerURI:    "https://github.com",
				Discovery: DiscoveryConfig{
					EnableDiscovery: true,
				},
				Scopes: []OAuthScope{
					{ScopeValue: "repo"},
				},
			},
			wantErr: "scope 0: description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestThirdpartyOAuth2Service_RedactedCopy(t *testing.T) {
	service := &ThirdpartyOAuth2Service{
		ID:           "650e8400-e29b-41d4-a716-446655440001",
		DisplayName:  "GitHub",
		ClientID:     "Iv1.abcd1234",
		ClientSecret: "secret123",
		IssuerURI:    "https://github.com",
		Scopes: []OAuthScope{
			{ScopeValue: "repo", Description: "Full repository access"},
		},
	}

	redacted := service.RedactedCopy()

	assert.Equal(t, service.ID, redacted.ID)
	assert.Equal(t, service.ClientID, redacted.ClientID)
	assert.Equal(t, "REDACTED", redacted.ClientSecret)
	assert.Equal(t, "secret123", service.ClientSecret) // Original unchanged
}

func TestThirdpartyOAuth2Service_Copy(t *testing.T) {
	metadataURL := "https://github.com/.well-known/oauth-authorization-server"

	original := &ThirdpartyOAuth2Service{
		ID:           "650e8400-e29b-41d4-a716-446655440001",
		DisplayName:  "GitHub",
		ClientID:     "Iv1.abcd1234",
		ClientSecret: "secret123",
		IssuerURI:    "https://github.com",
		Discovery: DiscoveryConfig{
			EnableDiscovery: true,
			MetadataURL:     &metadataURL,
		},
		Endpoints: OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/login/oauth/access_token",
			AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
		},
		Scopes: []OAuthScope{
			{ScopeValue: "repo", Description: "Full repository access"},
			{ScopeValue: "user:email", Description: "Access user email"},
		},
	}

	copy := original.Copy()

	// Verify equality
	assert.Equal(t, original.ID, copy.ID)
	assert.Equal(t, original.ClientSecret, copy.ClientSecret)
	assert.Equal(t, len(original.Scopes), len(copy.Scopes))

	// Verify deep copy of scopes
	copy.Scopes[0].Description = "Modified"
	assert.Equal(t, "Full repository access", original.Scopes[0].Description)
	assert.Equal(t, "Modified", copy.Scopes[0].Description)
}

func TestOAuthScope_Validate(t *testing.T) {
	tests := []struct {
		name    string
		scope   OAuthScope
		wantErr string
	}{
		{
			name: "valid scope",
			scope: OAuthScope{
				ScopeValue:  "repo",
				Description: "Full repository access",
			},
			wantErr: "",
		},
		{
			name: "missing scope_value",
			scope: OAuthScope{
				Description: "Full repository access",
			},
			wantErr: "scope_value is required",
		},
		{
			name: "missing description",
			scope: OAuthScope{
				ScopeValue: "repo",
			},
			wantErr: "description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scope.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestThirdpartyOAuth2Service_ValidateProtectedResources(t *testing.T) {
	tests := []struct {
		name               string
		protectedResources []string
		wantErr            string
	}{
		{
			name:               "empty protected_resources",
			protectedResources: []string{},
			wantErr:            "",
		},
		{
			name:               "single valid HTTPS URI",
			protectedResources: []string{"https://api.github.com"},
			wantErr:            "",
		},
		{
			name:               "multiple valid URIs",
			protectedResources: []string{"https://api.github.com", "https://www.googleapis.com", "https://graph.microsoft.com"},
			wantErr:            "",
		},
		{
			name:               "URI with path",
			protectedResources: []string{"https://api.example.com/v1"},
			wantErr:            "",
		},
		{
			name:               "URI with trailing slash",
			protectedResources: []string{"https://api.github.com/"},
			wantErr:            "",
		},
		{
			name:               "local HTTP URI",
			protectedResources: []string{"http://localhost:8080"},
			wantErr:            "",
		},
		{
			name:               "invalid URI - no scheme",
			protectedResources: []string{"api.github.com"},
			wantErr:            "protected_resources[0]",
		},
		{
			name:               "invalid URI - not a URL",
			protectedResources: []string{"not-a-valid-uri"},
			wantErr:            "protected_resources[0]",
		},
		{
			name:               "invalid URI - scheme only",
			protectedResources: []string{"https://"},
			wantErr:            "protected_resources[0]",
		},
		{
			name:               "mixed valid and invalid URIs",
			protectedResources: []string{"https://api.github.com", "invalid-uri"},
			wantErr:            "protected_resources[1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &ThirdpartyOAuth2Service{
				ProtectedResources: tt.protectedResources,
			}

			err := service.ValidateProtectedResources()
			if tt.wantErr != "" {
				require.Error(t, err, "expected error but got nil")
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err, "expected no error but got: %v", err)
			}
		})
	}
}
