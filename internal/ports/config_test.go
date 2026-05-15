package ports

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validBaseOAuth2Config returns a minimally valid OAuth2AuthServerConfig for test setup.
func validBaseOAuth2Config() OAuth2AuthServerConfig {
	return OAuth2AuthServerConfig{
		Mode: "proxy",
		Proxy: ProxyModeConfig{
			UpstreamIssuerURI:         "https://issuer.example.com",
			UpstreamAuthorizeEndpoint: "https://issuer.example.com/authorize",
			UpstreamTokenEndpoint:     "https://issuer.example.com/token",
		},
	}
}

func validLocalOAuth2Config() OAuth2AuthServerConfig {
	return OAuth2AuthServerConfig{
		Mode: "local",
		Local: LocalModeConfig{
			TokenTTL: time.Hour,
		},
	}
}

// TestOAuth2AuthServerConfig_Validate_OppositeModeRejection verifies that proxy mode rejects
// local.* settings and local mode rejects proxy.* settings.
func TestOAuth2AuthServerConfig_Validate_OppositeModeRejection(t *testing.T) {
	t.Run("proxy mode with local.token_ttl set returns error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.Local.TokenTTL = time.Hour
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("proxy mode with local.token_claims_expression set returns error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.Local.TokenClaimsExpression = `{"sub": claims.sub}`
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("proxy mode with no local settings — no error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("local mode with proxy.upstream_issuer_uri set returns error", func(t *testing.T) {
		cfg := validLocalOAuth2Config()
		cfg.Proxy.UpstreamIssuerURI = "https://issuer.example.com"
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("local mode with proxy.upstream_authorize_endpoint set returns error", func(t *testing.T) {
		cfg := validLocalOAuth2Config()
		cfg.Proxy.UpstreamAuthorizeEndpoint = "https://issuer.example.com/authorize"
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("local mode with proxy.upstream_token_endpoint set returns error", func(t *testing.T) {
		cfg := validLocalOAuth2Config()
		cfg.Proxy.UpstreamTokenEndpoint = "https://issuer.example.com/token"
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("local mode with proxy.upstream_timeout_seconds set returns error", func(t *testing.T) {
		cfg := validLocalOAuth2Config()
		cfg.Proxy.UpstreamTimeoutSeconds = 30
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("local mode with no proxy settings — no error", func(t *testing.T) {
		cfg := validLocalOAuth2Config()
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}

// T032: Unit tests for OAuth2AuthServerConfig.Validate() — multi-agent client fields.
func TestOAuth2AuthServerConfig_Validate_MultiAgentClient(t *testing.T) {
	t.Run("Enabled=true missing AgentIDParamName returns startup error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.MultiAgentClient = MultiAgentClientConfig{
			Enabled:          true,
			AgentIDParamName: "",
			AgentIDClaimName: "x_agent_id",
		}
		err := cfg.Validate()
		require.Error(t, err)
		type fieldError interface{ Field() string }
		fe, ok := err.(fieldError)
		require.True(t, ok, "error must implement Field() method")
		assert.Contains(t, fe.Field(), "agent_id_param_name")
		assert.Contains(t, fe.Field(), "required")
	})

	t.Run("Enabled=true missing AgentIDClaimName returns startup error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.MultiAgentClient = MultiAgentClientConfig{
			Enabled:          true,
			AgentIDParamName: "x_agent_id",
			AgentIDClaimName: "",
		}
		err := cfg.Validate()
		require.Error(t, err)
		type fieldError interface{ Field() string }
		fe, ok := err.(fieldError)
		require.True(t, ok, "error must implement Field() method")
		assert.Contains(t, fe.Field(), "agent_id_claim_name")
		assert.Contains(t, fe.Field(), "required")
	})

	t.Run("Enabled=false with empty param and claim names — no error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.MultiAgentClient = MultiAgentClientConfig{
			Enabled:          false,
			AgentIDParamName: "",
			AgentIDClaimName: "",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("Enabled=true with both fields set — no error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.MultiAgentClient = MultiAgentClientConfig{
			Enabled:          true,
			AgentIDParamName: "x_agent_id",
			AgentIDClaimName: "x_agent_id",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}

// T033: Proxy mode config validation.
func TestOAuth2AuthServerConfig_Validate_ProxyMode(t *testing.T) {
	t.Run("proxy mode with required proxy section — no error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("proxy mode missing upstream_issuer_uri — error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.Proxy.UpstreamIssuerURI = ""
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "upstream_issuer_uri")
	})

	t.Run("proxy mode missing upstream_authorize_endpoint — error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.Proxy.UpstreamAuthorizeEndpoint = ""
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("proxy mode missing upstream_token_endpoint — error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.Proxy.UpstreamTokenEndpoint = ""
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("proxy mode with local section set — error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.Local.TokenTTL = time.Hour
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "local")
	})
}

// T034: Local mode config validation.
func TestOAuth2AuthServerConfig_Validate_LocalMode(t *testing.T) {
	t.Run("local mode with local section — no error", func(t *testing.T) {
		cfg := validLocalOAuth2Config()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("local mode sets default token_ttl when empty", func(t *testing.T) {
		cfg := OAuth2AuthServerConfig{Mode: "local"}
		err := cfg.Validate()
		require.NoError(t, err)
		assert.Equal(t, time.Hour, cfg.Local.TokenTTL)
	})

	t.Run("local mode with proxy section set — error", func(t *testing.T) {
		cfg := validLocalOAuth2Config()
		cfg.Proxy.UpstreamIssuerURI = "https://issuer.example.com"
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "proxy")
	})
}

// T035: Hybrid mode config validation.
func TestOAuth2AuthServerConfig_Validate_HybridMode(t *testing.T) {
	validHybridConfig := func() OAuth2AuthServerConfig {
		return OAuth2AuthServerConfig{
			Mode: "hybrid",
			Proxy: ProxyModeConfig{
				UpstreamIssuerURI:         "https://issuer.example.com",
				UpstreamAuthorizeEndpoint: "https://issuer.example.com/authorize",
				UpstreamTokenEndpoint:     "https://issuer.example.com/token",
			},
			Local: LocalModeConfig{
				TokenTTL: time.Hour,
			},
		}
	}

	t.Run("hybrid mode with both sections — no error", func(t *testing.T) {
		cfg := validHybridConfig()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("hybrid mode missing proxy.upstream_issuer_uri — error", func(t *testing.T) {
		cfg := validHybridConfig()
		cfg.Proxy.UpstreamIssuerURI = ""
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "upstream_issuer_uri")
	})

	t.Run("hybrid mode missing proxy.upstream_authorize_endpoint — error", func(t *testing.T) {
		cfg := validHybridConfig()
		cfg.Proxy.UpstreamAuthorizeEndpoint = ""
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("hybrid mode missing proxy.upstream_token_endpoint — error", func(t *testing.T) {
		cfg := validHybridConfig()
		cfg.Proxy.UpstreamTokenEndpoint = ""
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("hybrid mode sets default token_ttl when empty", func(t *testing.T) {
		cfg := validHybridConfig()
		cfg.Local.TokenTTL = 0
		err := cfg.Validate()
		require.NoError(t, err)
		assert.Equal(t, time.Hour, cfg.Local.TokenTTL)
	})
}

// T036: issue_token deprecation error.
func TestOAuth2AuthServerConfig_Validate_IssueTokenDeprecation(t *testing.T) {
	t.Run("issue_token mode returns deprecation error", func(t *testing.T) {
		cfg := OAuth2AuthServerConfig{Mode: "issue_token"}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issue_token")
		assert.Contains(t, err.Error(), "local")
	})
}

// T036a: empty mode defaults to proxy when other proxy fields are provided.
func TestOAuth2AuthServerConfig_Validate_EmptyModeDefaultsToProxy(t *testing.T) {
	t.Run("empty mode with proxy fields defaults to proxy — no error", func(t *testing.T) {
		cfg := OAuth2AuthServerConfig{
			Mode: "",
			Proxy: ProxyModeConfig{
				UpstreamIssuerURI:         "https://issuer.example.com",
				UpstreamAuthorizeEndpoint: "https://issuer.example.com/authorize",
				UpstreamTokenEndpoint:     "https://issuer.example.com/token",
			},
		}
		err := cfg.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "proxy", cfg.Mode)
	})
}

// T052-T053: CIMD gating by mode.
func TestOAuth2AuthServerConfig_Validate_CIMDGating(t *testing.T) {
	t.Run("CIMD enabled in proxy mode — error", func(t *testing.T) {
		cfg := validBaseOAuth2Config()
		cfg.CIMD = CIMDConfig{Enabled: true}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cimd")
	})

	t.Run("CIMD enabled in local mode — no error", func(t *testing.T) {
		cfg := validLocalOAuth2Config()
		cfg.CIMD = CIMDConfig{
			Enabled: true,
			Cache:   CIMDCacheConfig{MinTTL: time.Minute, MaxTTL: time.Hour},
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("CIMD enabled in hybrid mode — no error", func(t *testing.T) {
		cfg := OAuth2AuthServerConfig{
			Mode: "hybrid",
			Proxy: ProxyModeConfig{
				UpstreamIssuerURI:         "https://issuer.example.com",
				UpstreamAuthorizeEndpoint: "https://issuer.example.com/authorize",
				UpstreamTokenEndpoint:     "https://issuer.example.com/token",
			},
			Local: LocalModeConfig{TokenTTL: time.Hour},
			CIMD: CIMDConfig{
				Enabled: true,
				Cache:   CIMDCacheConfig{MinTTL: time.Minute, MaxTTL: time.Hour},
			},
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
