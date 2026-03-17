package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validGoogleServiceAccountJSON is a realistic but non-functional service account JSON
// that passes all structural validation checks.
const validGoogleServiceAccountJSON = `{
  "type": "service_account",
  "project_id": "my-project-123",
  "private_key_id": "key-id-1234",
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA2a2rwplBQLf0kTmkGp5RJFBpJOFBBhfJmLO0YjCGSLCuoP7\noc5RfakePrivateKeyDataForTestingPurposesOnlyNotRealCryptographicKey\n-----END RSA PRIVATE KEY-----\n",
  "client_email": "my-service@my-project-123.iam.gserviceaccount.com",
  "client_id": "123456789012345678901",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/my-service%40my-project-123.iam.gserviceaccount.com"
}`

func TestParseGoogleServiceAccountKey(t *testing.T) {
	t.Run("valid service account JSON returns populated struct", func(t *testing.T) {
		gsk, err := ParseGoogleServiceAccountKey(validGoogleServiceAccountJSON)
		require.NoError(t, err)
		require.NotNil(t, gsk)
		assert.Equal(t, "service_account", gsk.Type)
		assert.Equal(t, "my-service@my-project-123.iam.gserviceaccount.com", gsk.ClientEmail)
		assert.NotEmpty(t, gsk.PrivateKey)
		assert.Equal(t, "https://oauth2.googleapis.com/token", gsk.TokenURI)
		assert.Equal(t, "123456789012345678901", gsk.ClientID)
	})

	t.Run("wrong type field returns error mentioning service_account", func(t *testing.T) {
		wrongType := strings.ReplaceAll(validGoogleServiceAccountJSON,
			`"type": "service_account"`,
			`"type": "authorized_user"`)
		_, err := ParseGoogleServiceAccountKey(wrongType)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service_account")
	})

	t.Run("invalid JSON returns parse error", func(t *testing.T) {
		_, err := ParseGoogleServiceAccountKey(`{not valid json}`)
		require.Error(t, err)
	})

	t.Run("missing private_key returns error naming the field", func(t *testing.T) {
		noPrivateKey := strings.ReplaceAll(validGoogleServiceAccountJSON,
			`"private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA2a2rwplBQLf0kTmkGp5RJFBpJOFBBhfJmLO0YjCGSLCuoP7\noc5RfakePrivateKeyDataForTestingPurposesOnlyNotRealCryptographicKey\n-----END RSA PRIVATE KEY-----\n"`,
			`"private_key": ""`)
		_, err := ParseGoogleServiceAccountKey(noPrivateKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "private_key")
	})

	t.Run("missing client_email returns error naming the field", func(t *testing.T) {
		noEmail := strings.ReplaceAll(validGoogleServiceAccountJSON,
			`"client_email": "my-service@my-project-123.iam.gserviceaccount.com"`,
			`"client_email": ""`)
		_, err := ParseGoogleServiceAccountKey(noEmail)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client_email")
	})

	t.Run("missing token_uri returns error naming the field", func(t *testing.T) {
		noTokenURI := strings.ReplaceAll(validGoogleServiceAccountJSON,
			`"token_uri": "https://oauth2.googleapis.com/token"`,
			`"token_uri": ""`)
		_, err := ParseGoogleServiceAccountKey(noTokenURI)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token_uri")
	})

	t.Run("missing client_id returns error naming the field", func(t *testing.T) {
		noClientID := strings.ReplaceAll(validGoogleServiceAccountJSON,
			`"client_id": "123456789012345678901"`,
			`"client_id": ""`)
		_, err := ParseGoogleServiceAccountKey(noClientID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client_id")
	})

	t.Run("empty private_key returns non-empty validation error", func(t *testing.T) {
		// Explicitly set empty private_key (different from missing private_key check)
		emptyKey := strings.ReplaceAll(validGoogleServiceAccountJSON,
			`"private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA2a2rwplBQLf0kTmkGp5RJFBpJOFBBhfJmLO0YjCGSLCuoP7\noc5RfakePrivateKeyDataForTestingPurposesOnlyNotRealCryptographicKey\n-----END RSA PRIVATE KEY-----\n"`,
			`"private_key": ""`)
		_, err := ParseGoogleServiceAccountKey(emptyKey)
		require.Error(t, err)
		assert.NotEmpty(t, err.Error())
		// Verify the error message does NOT contain any private key material
		assert.NotContains(t, err.Error(), "BEGIN RSA PRIVATE KEY")
		assert.NotContains(t, err.Error(), "PRIVATE KEY")
	})

	t.Run("JSON exceeding 32 KB returns size limit error", func(t *testing.T) {
		// Build a JSON that exceeds 32 KB
		largePadding := strings.Repeat("x", 33*1024)
		largeJSON := strings.ReplaceAll(validGoogleServiceAccountJSON,
			`"project_id": "my-project-123"`,
			`"project_id": "`+largePadding+`"`)
		_, err := ParseGoogleServiceAccountKey(largeJSON)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "32 KB")
	})

	t.Run("issuer_uri host matching token_uri host is valid", func(t *testing.T) {
		// This tests validateIssuerURIAgainstTokenURI via the entity validation path.
		// token_uri host in the fixture is "oauth2.googleapis.com"
		// issuerURI with same host should pass.
		err := validateIssuerURIAgainstTokenURI(
			"https://oauth2.googleapis.com",
			"https://oauth2.googleapis.com/token",
		)
		require.NoError(t, err)
	})

	t.Run("issuer_uri host differing from token_uri host returns validation error", func(t *testing.T) {
		err := validateIssuerURIAgainstTokenURI(
			"https://accounts.google.com",
			"https://oauth2.googleapis.com/token",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "accounts.google.com")
		assert.Contains(t, err.Error(), "oauth2.googleapis.com")
	})
}
