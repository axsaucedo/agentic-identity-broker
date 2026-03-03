package memory

import (
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderEntityCopy_Success(t *testing.T) {
	ciphertext := []byte("encrypted-bytes")
	original := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "svc-1",
		DisplayName: "Test Provider",
		ClientID:    "client-abc",
		Secret:      model.NewEncryptedSecret(ciphertext),
		IssuerURI:   "https://issuer.example.com",
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		ProtectedResources: []string{"https://api.example.com"},
	}

	copied, err := providerEntityCopy(original)
	require.NoError(t, err)
	require.NotNil(t, copied)

	assert.Equal(t, original.ID, copied.ID)
	assert.Equal(t, original.DisplayName, copied.DisplayName)
	assert.True(t, copied.Secret.IsEncrypted())

	ct, err := copied.Secret.GetCiphertext()
	require.NoError(t, err)
	assert.Equal(t, ciphertext, ct)
}

func TestProviderEntityCopy_NilEntityFails(t *testing.T) {
	_, err := providerEntityCopy(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestProviderEntityCopy_PlaintextSecretFails(t *testing.T) {
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "svc-1",
		DisplayName: "Provider",
		ClientID:    "client-1",
		Secret:      model.NewPlaintextSecret("plain-secret"),
		IssuerURI:   "https://issuer.example.com",
	}

	_, err := providerEntityCopy(entity)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "encrypted")
}

func TestProviderEntityCopy_IsolatesFromOriginal(t *testing.T) {
	original := &model.ThirdpartyOAuth2ProviderEntity{
		ID:                 "svc-1",
		DisplayName:        "Original",
		ClientID:           "client-1",
		Secret:             model.NewEncryptedSecret([]byte("ct")),
		IssuerURI:          "https://issuer.example.com",
		ProtectedResources: []string{"https://resource.example.com"},
		Scopes:             []model.OAuthScope{{ScopeValue: "read", Description: "Read"}},
	}

	copied, err := providerEntityCopy(original)
	require.NoError(t, err)

	// Mutate original — copy must remain unchanged
	original.DisplayName = "Mutated"
	original.ProtectedResources[0] = "https://mutated.example.com"
	original.Scopes[0].ScopeValue = "mutated"

	assert.Equal(t, "Original", copied.DisplayName)
	assert.Equal(t, "https://resource.example.com", copied.ProtectedResources[0])
	assert.Equal(t, "read", copied.Scopes[0].ScopeValue)
}
