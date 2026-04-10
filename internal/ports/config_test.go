package ports

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validBaseOAuth2Config returns a minimally valid OAuth2AuthServerConfig for test setup.
func validBaseOAuth2Config() OAuth2AuthServerConfig {
	return OAuth2AuthServerConfig{
		UpstreamIssuerURI:         "https://issuer.example.com",
		UpstreamAuthorizeEndpoint: "https://issuer.example.com/authorize",
		UpstreamTokenEndpoint:     "https://issuer.example.com/token",
	}
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
