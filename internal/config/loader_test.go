package config

import (
	"context"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationDefaults(t *testing.T) {
	// Set valid JWESigningKey for all tests

	t.Run("missing authentication config fails validation", func(t *testing.T) {
		t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY", generateBase64EncodedString(32))
		// Do NOT set principal_header_name or JWT — no authentication method configured
		loader := NewLoader()
		_, err := loader.GetConfig(context.Background())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "authentication")
	})

	t.Run("custom principal header name from environment variable", func(t *testing.T) {
		// Set environment variable
		t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Authenticated-User")
		t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Admin-User")

		loader := NewLoader()
		cfg, err := loader.GetConfig(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "X-Authenticated-User", cfg.Server.EndUser.Authentication.Preauth.PrincipalHeaderName)
		assert.Equal(t, "X-Admin-User", cfg.Server.Admin.Authentication.Preauth.PrincipalHeaderName)
	})

	t.Run("authentication configuration is not nil", func(t *testing.T) {
		t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
		t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
		loader := NewLoader()
		cfg, err := loader.GetConfig(context.Background())

		require.NoError(t, err)
		assert.NotNil(t, cfg.Server.EndUser.Authentication)
		assert.NotNil(t, cfg.Server.Admin.Authentication)
		assert.NotNil(t, cfg.Server.EndUser.Authentication.Preauth)
		assert.NotNil(t, cfg.Server.Admin.Authentication.Preauth)
	})
}

func TestDefaultServerConfig(t *testing.T) {
	defaults := ports.DefaultServerConfig()

	t.Run("enduser server has no default principal header", func(t *testing.T) {
		assert.Empty(t, defaults.EndUser.Authentication.Preauth.PrincipalHeaderName)
	})

	t.Run("admin server has no default principal header", func(t *testing.T) {
		assert.Empty(t, defaults.Admin.Authentication.Preauth.PrincipalHeaderName)
	})
}

func TestConfigurationPrecedence(t *testing.T) {
	t.Run("environment variable sets principal header", func(t *testing.T) {
		t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Custom-Header")
		t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Admin-Header")

		loader := NewLoader()
		cfg, err := loader.GetConfig(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "X-Custom-Header", cfg.Server.EndUser.Authentication.Preauth.PrincipalHeaderName)
		assert.Equal(t, "X-Admin-Header", cfg.Server.Admin.Authentication.Preauth.PrincipalHeaderName)
	})

	t.Run("admin and enduser can have different headers", func(t *testing.T) {
		t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-User-Header")
		t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Admin-Header")

		loader := NewLoader()
		cfg, err := loader.GetConfig(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "X-User-Header", cfg.Server.EndUser.Authentication.Preauth.PrincipalHeaderName)
		assert.Equal(t, "X-Admin-Header", cfg.Server.Admin.Authentication.Preauth.PrincipalHeaderName)
	})
}

func TestConfigurationSources(t *testing.T) {
	// Set valid JWESigningKey, KeyEncryptionKey, and required principal headers
	t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", generateBase64EncodedString(32))
	t.Setenv("IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY", generateBase64EncodedString(32))
	t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
	t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")

	loader := NewLoader()
	_, err := loader.GetConfig(context.Background())
	require.NoError(t, err)

	sources := loader.GetSources()
	assert.Greater(t, len(sources), 0)

	// Verify defaults source is present
	hasDefaults := false
	for _, source := range sources {
		if source.Type == ports.SourceTypeDefault {
			hasDefaults = true
			// Verify authentication defaults are recorded
			// principal_header_name should NOT have a default (security: must be explicit)
			hasAuthEnduser := false
			hasAuthAdmin := false
			for _, key := range source.Keys {
				if key == "server.enduser.authentication.preauth.principal_header_name" {
					hasAuthEnduser = true
				}
				if key == "server.admin.authentication.preauth.principal_header_name" {
					hasAuthAdmin = true
				}
			}
			assert.False(t, hasAuthEnduser, "enduser principal_header_name should not have a default")
			assert.False(t, hasAuthAdmin, "admin principal_header_name should not have a default")
		}
	}
	assert.True(t, hasDefaults, "defaults source should be present")
}

func TestAWSKMSConfigurationEnvironmentVariables(t *testing.T) {
	t.Run("AWS KMS configuration from environment variables", func(t *testing.T) {
		// Set basic required env vars
		t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
		t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN", "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_TABLE_NAME", "MyBranchKeys")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_REGION", "us-east-1")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_ASSUME_ROLE_ARN", "arn:aws:iam::123456789012:role/EncryptionRole-prod")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_ENDPOINT", "http://localhost:4566")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_ENDPOINT", "http://localhost:8000")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_PROFILE", "production")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DISABLE_SSL", "false")

		loader := NewLoader()
		cfg, err := loader.GetConfig(context.Background())

		require.NoError(t, err)
		require.NotNil(t, cfg.Encryption.AWSKMS)
		assert.Equal(t, "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012", cfg.Encryption.AWSKMS.KeyARN)
		assert.Equal(t, "MyBranchKeys", cfg.Encryption.AWSKMS.DynamoDBTableName)
		assert.Equal(t, "us-east-1", cfg.Encryption.AWSKMS.Region)
		assert.Equal(t, "arn:aws:iam::123456789012:role/EncryptionRole-prod", cfg.Encryption.AWSKMS.AssumeRoleARN)
		assert.Equal(t, "http://localhost:4566", cfg.Encryption.AWSKMS.KMSEndpoint)
		assert.Equal(t, "http://localhost:8000", cfg.Encryption.AWSKMS.DynamoDBEndpoint)
		assert.Equal(t, "production", cfg.Encryption.AWSKMS.Profile)
		assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", cfg.Encryption.AWSKMS.AccessKeyID)
		assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", cfg.Encryption.AWSKMS.SecretAccessKey)
		assert.Equal(t, false, cfg.Encryption.AWSKMS.DisableSSL)
	})

	t.Run("DynamoDB timeout configuration from environment variables", func(t *testing.T) {
		t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", generateBase64EncodedString(32))
		t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
		t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN", "arn:aws:kms:us-east-1:123456789012:key/12345678")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_READ_TIMEOUT", "10s")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_WRITE_TIMEOUT", "15s")
		t.Setenv("IDENTITY_BROKER_ENCRYPTION_AWS_KMS_BRANCH_KEY_TTL", "30m")

		loader := NewLoader()
		cfg, err := loader.GetConfig(context.Background())

		require.NoError(t, err)
		require.NotNil(t, cfg.Encryption.AWSKMS)
		assert.Equal(t, "10s", cfg.Encryption.AWSKMS.DynamoDBReadTimeout)
		assert.Equal(t, "15s", cfg.Encryption.AWSKMS.DynamoDBWriteTimeout)
		assert.Equal(t, "30m", cfg.Encryption.AWSKMS.BranchKeyTTL)
	})
}

// setMinimalConfigEnv configures the minimum set of environment variables required to
// pass config validation, matching the pattern used throughout this test file.
// It uses t.Setenv so all variables are automatically restored after the test.
func setMinimalConfigEnv(t *testing.T) {
t.Helper()
t.Setenv("IDENTITY_BROKER_JWE_SIGNING_KEY", generateBase64EncodedString(32))
t.Setenv("IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY", generateBase64EncodedString(32))
t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Remote-User")
}

func TestTokenExchangeExpectedAudienceConfiguration(t *testing.T) {
t.Run("default expected_audience is token-exchange-broker when not configured", func(t *testing.T) {
setMinimalConfigEnv(t)
// No IDENTITY_BROKER_TOKEN_EXCHANGE_EXPECTED_AUDIENCE set

loader := NewLoader()
cfg, err := loader.GetConfig(context.Background())

require.NoError(t, err)
assert.Equal(t, "token-exchange-broker", cfg.TokenExchange.ExpectedAudience,
"expected_audience should default to token-exchange-broker via Viper SetDefault")
})

t.Run("expected_audience is overridden by environment variable", func(t *testing.T) {
setMinimalConfigEnv(t)
t.Setenv("IDENTITY_BROKER_TOKEN_EXCHANGE_EXPECTED_AUDIENCE", "my-custom-gateway-audience")

loader := NewLoader()
cfg, err := loader.GetConfig(context.Background())

require.NoError(t, err)
assert.Equal(t, "my-custom-gateway-audience", cfg.TokenExchange.ExpectedAudience,
"expected_audience should be overridden by IDENTITY_BROKER_TOKEN_EXCHANGE_EXPECTED_AUDIENCE env var")
})
}
