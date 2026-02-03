package config

import (
	"slices"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"
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
				UpstreamIssuerURI:         "https://auth.example.com",
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				UpstreamTokenEndpoint:     "https://auth.example.com/token",
			},
			wantErr:      false,
			wantDefaults: true, // Defaults should be set
		},
		{
			name: "valid config with all fields including optional",
			config: &ports.OAuth2AuthServerConfig{
				UpstreamIssuerURI:         "https://auth.example.com",
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				UpstreamTokenEndpoint:     "https://auth.example.com/token",
				SupportedResponseTypes:    []string{"code"},
				SupportedGrantTypes:       []string{"authorization_code"},
				UpstreamTimeoutSeconds:    60,
				Mode:                      "proxy",
			},
			wantErr:      false,
			wantDefaults: false, // No defaults should be set
		},
		{
			name: "missing upstream_issuer_uri",
			config: &ports.OAuth2AuthServerConfig{
				UpstreamIssuerURI:         "", // missing
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				UpstreamTokenEndpoint:     "https://auth.example.com/token",
			},
			wantErr:  true,
			errField: "oauth2_authorization_server.upstream_issuer_uri",
		},
		{
			name: "missing upstream_authorize_endpoint",
			config: &ports.OAuth2AuthServerConfig{
				UpstreamIssuerURI:         "https://auth.example.com",
				UpstreamAuthorizeEndpoint: "", // missing
				UpstreamTokenEndpoint:     "https://auth.example.com/token",
			},
			wantErr:  true,
			errField: "oauth2_authorization_server.upstream_authorize_endpoint",
		},
		{
			name: "missing upstream_token_endpoint",
			config: &ports.OAuth2AuthServerConfig{
				UpstreamIssuerURI:         "https://auth.example.com",
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				UpstreamTokenEndpoint:     "", // missing
			},
			wantErr:  true,
			errField: "oauth2_authorization_server.upstream_token_endpoint",
		},
		{
			name: "empty slices should be replaced with defaults",
			config: &ports.OAuth2AuthServerConfig{
				UpstreamIssuerURI:         "https://auth.example.com",
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				UpstreamTokenEndpoint:     "https://auth.example.com/token",
				SupportedResponseTypes:    []string{}, // empty, should be set to default
				SupportedGrantTypes:       []string{}, // empty, should be set to default
			},
			wantErr:      false,
			wantDefaults: true,
		},
		{
			name: "zero timeout should be set to default",
			config: &ports.OAuth2AuthServerConfig{
				UpstreamIssuerURI:         "https://auth.example.com",
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				UpstreamTokenEndpoint:     "https://auth.example.com/token",
				UpstreamTimeoutSeconds:    0, // zero, should be set to default 30
			},
			wantErr:      false,
			wantDefaults: true,
		},
		{
			name: "empty mode should be set to default",
			config: &ports.OAuth2AuthServerConfig{
				UpstreamIssuerURI:         "https://auth.example.com",
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				UpstreamTokenEndpoint:     "https://auth.example.com/token",
				Mode:                      "", // empty, should be set to default "proxy"
			},
			wantErr:      false,
			wantDefaults: true,
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
				if tt.config.UpstreamTimeoutSeconds == 0 {
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
				assert.Equal(t, 30, tt.config.UpstreamTimeoutSeconds, "UpstreamTimeoutSeconds should be 30")
				assert.Equal(t, "proxy", tt.config.Mode, "Mode should be 'proxy'")
			}
		})
	}
}
