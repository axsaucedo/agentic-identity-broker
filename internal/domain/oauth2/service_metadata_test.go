package oauth2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateMetadata_ProxyMode(t *testing.T) {
	svc := NewService(nil, nil, &OAuth2Config{
		PublicURL:              "https://broker.example.com",
		SupportedResponseTypes: []string{"code"},
		SupportedGrantTypes:    []string{"authorization_code", "refresh_token"},
	})

	metadata, err := svc.GenerateMetadata(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "https://broker.example.com", metadata.Issuer)
	assert.Equal(t, "https://broker.example.com/oauth2/authorize", metadata.AuthorizationEndpoint)
	assert.Equal(t, "https://broker.example.com/oauth2/token", metadata.TokenEndpoint)
	assert.Empty(t, metadata.JWKSURI, "JWKS URI should be empty in proxy mode")
	assert.Empty(t, metadata.CodeChallengeMethodsSupported, "code challenge methods should be empty in proxy mode")
}

func TestGenerateMetadata_LocalMode(t *testing.T) {
	svc := NewService(nil, nil, &OAuth2Config{
		Mode:                   "local",
		PublicURL:              "https://broker.example.com",
		SupportedResponseTypes: []string{"code"},
		SupportedGrantTypes:    []string{"authorization_code", "client_credentials"},
	})

	metadata, err := svc.GenerateMetadata(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "https://broker.example.com", metadata.Issuer)
	assert.Equal(t, "https://broker.example.com/oauth2/authorize", metadata.AuthorizationEndpoint)
	assert.Equal(t, "https://broker.example.com/oauth2/token", metadata.TokenEndpoint)
	assert.Equal(t, "https://broker.example.com/oauth2/jwks.json", metadata.JWKSURI)
	assert.Equal(t, []string{"S256"}, metadata.CodeChallengeMethodsSupported)
	assert.Equal(t, []string{"client_secret_post"}, metadata.TokenEndpointAuthMethodsSupported)
}

func TestGenerateMetadata_LocalModeWithCIMD(t *testing.T) {
	svc := NewService(nil, nil, &OAuth2Config{
		Mode:                   "local",
		PublicURL:              "https://broker.example.com",
		SupportedResponseTypes: []string{"code"},
		SupportedGrantTypes:    []string{"authorization_code", "client_credentials"},
		CIMDEnabled:            true,
	})

	metadata, err := svc.GenerateMetadata(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"none", "client_secret_post"}, metadata.TokenEndpointAuthMethodsSupported)
	cimdEnabled := metadata.ClientIDMetadataDocumentSupported
	require.NotNil(t, cimdEnabled)
	assert.True(t, *cimdEnabled)
}
