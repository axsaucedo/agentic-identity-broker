package model

import (
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThirdpartyOAuth2ProviderEntity_Validate(t *testing.T) {
	metadataURL := "https://github.com/.well-known/oauth-authorization-server"

	tests := []struct {
		name    string
		entity  *ThirdpartyOAuth2ProviderEntity
		wantErr string
	}{
		{
			name: "valid entity with plaintext secret",
			entity: &ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("Iv1.abcd1234"),
				Secret:      NewPlaintextSecret("super-secret"),
				IssuerURI:   "https://github.com",
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
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			wantErr: "",
		},
		{
			name: "valid entity with encrypted secret",
			entity: &ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("Iv1.abcd1234"),
				Secret:      NewEncryptedSecret([]byte{1, 2, 3, 4}),
				IssuerURI:   "https://github.com",
				Discovery:   DiscoveryConfig{EnableDiscovery: true},
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
			entity: &ThirdpartyOAuth2ProviderEntity{
				DisplayName: "GitHub",
				ClientID:    id.ClientID("Iv1.abcd1234"),
				Secret:      NewPlaintextSecret("secret"),
				IssuerURI:   "https://github.com",
			},
			wantErr: "provider ID cannot be empty",
		},
		{
			name: "missing display_name",
			entity: &ThirdpartyOAuth2ProviderEntity{
				ID:        id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				ClientID:  id.ClientID("Iv1.abcd1234"),
				Secret:    NewPlaintextSecret("secret"),
				IssuerURI: "https://github.com",
			},
			wantErr: "display_name is required",
		},
		{
			name: "display_name too long",
			entity: &ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				DisplayName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				ClientID:    id.ClientID("Iv1.abcd1234"),
				Secret:      NewPlaintextSecret("secret"),
				IssuerURI:   "https://github.com",
			},
			wantErr: "display_name exceeds 255 characters",
		},
		{
			name: "missing client_id",
			entity: &ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				DisplayName: "GitHub",
				Secret:      NewPlaintextSecret("secret"),
				IssuerURI:   "https://github.com",
			},
			wantErr: "client_id is required",
		},
		{
			name: "secret in plaintext state with empty value",
			entity: &ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("Iv1.abcd1234"),
				Secret:      NewPlaintextSecret(""),
				IssuerURI:   "https://github.com",
			},
			wantErr: "secret plaintext is empty",
		},
		{
			// A zero-value Secret (var s Secret / Secret{}) is the "uninitialized" state
			// described in the Secret type comment. IsPlaintext() returns true for it, so
			// callers who write "if s.IsPlaintext() { use(s.GetPlaintext()) }" would reach
			// GetPlaintext() — which then errors. Validate() must catch this so that an
			// entity that was never given a real secret is always rejected.
			name: "zero-value Secret (uninitialized) is rejected by Validate",
			entity: &ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("Iv1.abcd1234"),
				Secret:      Secret{}, // var s Secret — not constructed via NewPlaintextSecret
				IssuerURI:   "https://github.com",
			},
			wantErr: "secret plaintext is empty",
		},
		{
			name: "secret in encrypted state with empty ciphertext",
			entity: &ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("Iv1.abcd1234"),
				Secret:      NewEncryptedSecret([]byte{}),
				IssuerURI:   "https://github.com",
			},
			wantErr: "secret ciphertext is empty",
		},
		{
			name: "missing issuer_uri",
			entity: &ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("Iv1.abcd1234"),
				Secret:      NewPlaintextSecret("secret"),
			},
			wantErr: "issuer_uri is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entity.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
