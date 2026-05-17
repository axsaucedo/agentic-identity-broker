package config

import (
	"slices"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/servermode"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
)

// TestOAuth2AuthServerConfig_Validate tests OAuth2AuthServerConfig.Validate() method
func TestOAuth2AuthServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name         string
		config       *ports.OAuth2AuthServerConfig
		wantErr      bool
		errField     string
		wantDefaults bool
	}{
		{
			name: "valid config with all required fields",
			config: &ports.OAuth2AuthServerConfig{
				Mode: "proxy",
				Proxy: ports.ProxyModeConfig{
					UpstreamIssuerURI:         "https://auth.example.com",
					UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
					UpstreamTokenEndpoint:     "https://auth.example.com/token",
				},
			},
			wantErr:      false,
			wantDefaults: true, // Defaults should be set
		},
		{
			name: "valid config with all fields including optional",
			config: &ports.OAuth2AuthServerConfig{
				Mode: "proxy",
				Proxy: ports.ProxyModeConfig{
					UpstreamIssuerURI:         "https://auth.example.com",
					UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
					UpstreamTokenEndpoint:     "https://auth.example.com/token",
					UpstreamTimeoutSeconds:    60,
				},
				SupportedResponseTypes: []string{"code"},
				SupportedGrantTypes:    []string{"authorization_code"},
			},
			wantErr:      false,
			wantDefaults: false, // No defaults should be set
		},
		{
			name: "missing upstream_issuer_uri",
			config: &ports.OAuth2AuthServerConfig{
				Mode: "proxy",
				Proxy: ports.ProxyModeConfig{
					UpstreamIssuerURI:         "", // missing
					UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
					UpstreamTokenEndpoint:     "https://auth.example.com/token",
				},
			},
			wantErr:  true,
			errField: "oauth2_authorization_server.proxy.upstream_issuer_uri",
		},
		{
			name: "missing upstream_authorize_endpoint",
			config: &ports.OAuth2AuthServerConfig{
				Mode: "proxy",
				Proxy: ports.ProxyModeConfig{
					UpstreamIssuerURI:         "https://auth.example.com",
					UpstreamAuthorizeEndpoint: "", // missing
					UpstreamTokenEndpoint:     "https://auth.example.com/token",
				},
			},
			wantErr:  true,
			errField: "oauth2_authorization_server.proxy.upstream_authorize_endpoint",
		},
		{
			name: "missing upstream_token_endpoint",
			config: &ports.OAuth2AuthServerConfig{
				Mode: "proxy",
				Proxy: ports.ProxyModeConfig{
					UpstreamIssuerURI:         "https://auth.example.com",
					UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
					UpstreamTokenEndpoint:     "", // missing
				},
			},
			wantErr:  true,
			errField: "oauth2_authorization_server.proxy.upstream_token_endpoint",
		},
		{
			name: "empty slices should be replaced with defaults",
			config: &ports.OAuth2AuthServerConfig{
				Mode: "proxy",
				Proxy: ports.ProxyModeConfig{
					UpstreamIssuerURI:         "https://auth.example.com",
					UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
					UpstreamTokenEndpoint:     "https://auth.example.com/token",
				},
				SupportedResponseTypes: []string{}, // empty, should be set to default
				SupportedGrantTypes:    []string{}, // empty, should be set to default
			},
			wantErr:      false,
			wantDefaults: true,
		},
		{
			name: "zero timeout should be set to default",
			config: &ports.OAuth2AuthServerConfig{
				Mode: "proxy",
				Proxy: ports.ProxyModeConfig{
					UpstreamIssuerURI:         "https://auth.example.com",
					UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
					UpstreamTokenEndpoint:     "https://auth.example.com/token",
					UpstreamTimeoutSeconds:    0, // zero, should be set to default 30
				},
			},
			wantErr:      false,
			wantDefaults: true,
		},
		{
			name: "empty mode with proxy fields returns validation error",
			config: &ports.OAuth2AuthServerConfig{
				Proxy: ports.ProxyModeConfig{
					UpstreamIssuerURI:         "https://auth.example.com",
					UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
					UpstreamTokenEndpoint:     "https://auth.example.com/token",
				},
				Mode: "",
			},
			wantErr:  true,
			errField: "oauth2_authorization_server.mode is required (use 'proxy', 'local', or 'hybrid')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				// Check that the error field is correct
				if configErr, ok := err.(*config.ConfigError); ok {
					if configErr.Field != tt.errField {
						t.Errorf("Validate() error field = %q, want %q", configErr.Field, tt.errField)
					}
				} else if errWithField, ok := err.(interface{ Field() string }); ok {
					if errWithField.Field() != tt.errField {
						t.Errorf("Validate() error field = %q, want %q", errWithField.Field(), tt.errField)
					}
				} else {
					t.Errorf("Validate() error type = %T, want *config.ConfigError", err)
				}
			}

			// Check defaults if applicable
			if !tt.wantErr && tt.wantDefaults {
				if len(tt.config.SupportedResponseTypes) == 0 {
					t.Error("SupportedResponseTypes should have default values")
				}
				if len(tt.config.SupportedGrantTypes) == 0 {
					t.Error("SupportedGrantTypes should have default values")
				}
				if tt.config.Proxy.UpstreamTimeoutSeconds == 0 {
					t.Error("UpstreamTimeoutSeconds should have default value")
				}
				if tt.config.Mode == "" {
					t.Error("Mode should have default value")
				}
			}

			// Check that defaults contain expected values
			if !tt.wantErr && tt.wantDefaults {
				assert.True(t, slices.Contains(tt.config.SupportedResponseTypes, "code"), "SupportedResponseTypes should contain 'code'")
				assert.True(t, slices.Contains(tt.config.SupportedGrantTypes, "authorization_code"), "SupportedGrantTypes should contain 'authorization_code'")
				assert.Equal(t, 30, tt.config.Proxy.UpstreamTimeoutSeconds, "UpstreamTimeoutSeconds should be 30")
				assert.Equal(t, servermode.Proxy, tt.config.Mode, "Mode should be 'proxy'")
			}
		})
	}
}

// TestOAuth2AuthServerConfig_LocalMode tests local mode validation (T003).
func TestOAuth2AuthServerConfig_LocalMode(t *testing.T) {
	t.Run("local mode succeeds without upstream fields", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Mode: "local",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("local mode defaults token_ttl to 1h", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Mode: "local",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
		assert.Equal(t, time.Hour, cfg.Local.TokenTTL, "TokenTTL should default to 1 hour")
	})

	t.Run("local mode preserves custom token_ttl", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Mode:  "local",
			Local: ports.LocalModeConfig{TokenTTL: 30 * time.Minute},
		}
		err := cfg.Validate()
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Minute, cfg.Local.TokenTTL, "custom TokenTTL should be preserved")
	})

	t.Run("local mode does not require upstream fields", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Mode: "local",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("local mode sets default response types and grant types", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Mode: "local",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
		assert.Contains(t, cfg.SupportedResponseTypes, "code")
		assert.Contains(t, cfg.SupportedGrantTypes, "authorization_code")
		assert.Contains(t, cfg.SupportedGrantTypes, "client_credentials")
	})

	t.Run("multi_agent_client enabled in local mode is rejected", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Mode:             "local",
			MultiAgentClient: ports.MultiAgentClientConfig{Enabled: true},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "multi_agent_client is not supported in local mode")
	})

	t.Run("proxy mode unchanged - still requires upstream fields", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Mode: "proxy",
			// Missing upstream fields
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "upstream_issuer_uri")
	})

	t.Run("cimd enabled in proxy mode rejected", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Mode: "proxy",
			Proxy: ports.ProxyModeConfig{
				UpstreamIssuerURI:         "https://auth.example.com",
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				UpstreamTokenEndpoint:     "https://auth.example.com/token",
			},
			CIMD: ports.CIMDConfig{Enabled: true},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cimd.enabled requires mode 'local'")
	})

	t.Run("empty mode with proxy fields returns validation error", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Proxy: ports.ProxyModeConfig{
				UpstreamIssuerURI:         "https://auth.example.com",
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				UpstreamTokenEndpoint:     "https://auth.example.com/token",
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mode is required")
	})

	t.Run("invalid mode rejected", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Mode: "invalid_mode",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mode")
	})
}

// TestOAuth2AuthServerConfig_PartialConfigFails is a regression test ensuring that a
// partially populated OAuth2AuthServerConfig (non-zero but incomplete) is rejected rather
// than silently skipped by the unconfigured-block guard.
func TestOAuth2AuthServerConfig_PartialConfigFails(t *testing.T) {
	t.Run("only token_ttl set fails validation", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Local: ports.LocalModeConfig{TokenTTL: 30 * time.Minute},
		}
		err := cfg.Validate()
		assert.Error(t, err, "partial config with only token_ttl must fail, not be silently skipped")
	})

	t.Run("only token_claims_expression set fails validation", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Local: ports.LocalModeConfig{TokenClaimsExpression: `{"sub": subject_token.sub}`},
		}
		err := cfg.Validate()
		assert.Error(t, err, "partial config with only token_claims_expression must fail")
	})

	t.Run("only upstream_timeout_seconds set fails validation", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			Proxy: ports.ProxyModeConfig{UpstreamTimeoutSeconds: 60},
		}
		err := cfg.Validate()
		assert.Error(t, err, "partial config with only upstream_timeout_seconds must fail")
	})

	t.Run("only multi_agent_client.agent_id_param_name set fails validation", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			MultiAgentClient: ports.MultiAgentClientConfig{
				AgentIDParamName: "x_agent_id",
			},
		}
		err := cfg.Validate()
		assert.Error(t, err, "partial config with only multi_agent_client.agent_id_param_name must fail")
	})

	t.Run("only supported_response_types set fails validation", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			SupportedResponseTypes: []string{"code"},
		}
		err := cfg.Validate()
		assert.Error(t, err, "partial config with only supported_response_types must fail")
	})

	// Regression: len()==0 collapsed nil and []string{} so an explicitly empty list
	// was indistinguishable from "field not set" and the block was silently skipped.
	t.Run("explicit empty supported_response_types fails validation", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			SupportedResponseTypes: []string{},
		}
		err := cfg.Validate()
		assert.Error(t, err, "explicitly empty supported_response_types must not be skipped as unconfigured")
	})

	t.Run("explicit empty supported_grant_types fails validation", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{
			SupportedGrantTypes: []string{},
		}
		err := cfg.Validate()
		assert.Error(t, err, "explicitly empty supported_grant_types must not be skipped as unconfigured")
	})

	t.Run("zero value config fails validation — oauth2_authorization_server is mandatory", func(t *testing.T) {
		cfg := &ports.OAuth2AuthServerConfig{}
		err := cfg.Validate()
		assert.Error(t, err, "zero-value OAuth2AuthServerConfig must fail — mode is required")
		assert.Contains(t, err.Error(), "mode")
	})
}
