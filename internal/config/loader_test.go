package config_test

import (
	"context"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationDefaults(t *testing.T) {
	t.Run("default principal header name is set", func(t *testing.T) {
		loader := config.NewLoader()
		cfg, err := loader.GetConfig(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "X-Remote-User", cfg.Server.EndUser.Authentication.Preauth.PrincipalHeaderName)
		assert.Equal(t, "X-Remote-User", cfg.Server.Admin.Authentication.Preauth.PrincipalHeaderName)
	})

	t.Run("custom principal header name from environment variable", func(t *testing.T) {
		// Set environment variable
		t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Authenticated-User")
		t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Admin-User")

		loader := config.NewLoader()
		cfg, err := loader.GetConfig(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "X-Authenticated-User", cfg.Server.EndUser.Authentication.Preauth.PrincipalHeaderName)
		assert.Equal(t, "X-Admin-User", cfg.Server.Admin.Authentication.Preauth.PrincipalHeaderName)
	})

	t.Run("authentication configuration is not nil", func(t *testing.T) {
		loader := config.NewLoader()
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

	t.Run("enduser server has default principal header", func(t *testing.T) {
		assert.Equal(t, "X-Remote-User", defaults.EndUser.Authentication.Preauth.PrincipalHeaderName)
	})

	t.Run("admin server has default principal header", func(t *testing.T) {
		assert.Equal(t, "X-Remote-User", defaults.Admin.Authentication.Preauth.PrincipalHeaderName)
	})

	t.Run("authentication configuration structure is complete", func(t *testing.T) {
		assert.NotNil(t, defaults.EndUser.Authentication)
		assert.NotNil(t, defaults.EndUser.Authentication.Preauth)
		assert.NotEmpty(t, defaults.EndUser.Authentication.Preauth.PrincipalHeaderName)

		assert.NotNil(t, defaults.Admin.Authentication)
		assert.NotNil(t, defaults.Admin.Authentication.Preauth)
		assert.NotEmpty(t, defaults.Admin.Authentication.Preauth.PrincipalHeaderName)
	})
}

func TestConfigurationPrecedence(t *testing.T) {
	t.Run("environment variable overrides default", func(t *testing.T) {
		// Set environment variable to override default
		t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Custom-Header")

		loader := config.NewLoader()
		cfg, err := loader.GetConfig(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "X-Custom-Header", cfg.Server.EndUser.Authentication.Preauth.PrincipalHeaderName)
	})

	t.Run("admin and enduser can have different headers", func(t *testing.T) {
		t.Setenv("IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-User-Header")
		t.Setenv("IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME", "X-Admin-Header")

		loader := config.NewLoader()
		cfg, err := loader.GetConfig(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "X-User-Header", cfg.Server.EndUser.Authentication.Preauth.PrincipalHeaderName)
		assert.Equal(t, "X-Admin-Header", cfg.Server.Admin.Authentication.Preauth.PrincipalHeaderName)
	})
}

func TestConfigurationSources(t *testing.T) {
	loader := config.NewLoader()
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
			assert.True(t, hasAuthEnduser, "enduser authentication config should be in default keys")
			assert.True(t, hasAuthAdmin, "admin authentication config should be in default keys")
		}
	}
	assert.True(t, hasDefaults, "defaults source should be present")
}
