package ports_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// TestHybridOAuth2Config_GrantTypes_ReturnsUnion verifies that GrantTypes() returns the
// deduplicated union of proxy and local grant types, not just the proxy side.
func TestHybridOAuth2Config_GrantTypes_ReturnsUnion(t *testing.T) {
	cfg := &ports.HybridOAuth2Config{
		Proxy: ports.ProxyOAuth2Config{
			SupportedGrantTypes: []string{"authorization_code"},
		},
		Local: ports.LocalOAuth2Config{
			SupportedGrantTypes: []string{"authorization_code", "client_credentials"},
		},
	}

	got := cfg.GrantTypes()

	assert.ElementsMatch(t, []string{"authorization_code", "client_credentials"}, got,
		"GrantTypes must return union of proxy and local grant types")
}

// TestHybridOAuth2Config_ResponseTypes_ReturnsUnion verifies that ResponseTypes() returns
// the deduplicated union of proxy and local response types.
func TestHybridOAuth2Config_ResponseTypes_ReturnsUnion(t *testing.T) {
	cfg := &ports.HybridOAuth2Config{
		Proxy: ports.ProxyOAuth2Config{
			SupportedResponseTypes: []string{"code"},
		},
		Local: ports.LocalOAuth2Config{
			SupportedResponseTypes: []string{"code", "token"},
		},
	}

	got := cfg.ResponseTypes()

	assert.ElementsMatch(t, []string{"code", "token"}, got,
		"ResponseTypes must return union of proxy and local response types")
}

// TestHybridOAuth2Config_GrantTypes_IdenticalSlices verifies that when both sides are
// identical (the common case today), the result is not duplicated.
func TestHybridOAuth2Config_GrantTypes_IdenticalSlices(t *testing.T) {
	cfg := &ports.HybridOAuth2Config{
		Proxy: ports.ProxyOAuth2Config{
			SupportedGrantTypes: []string{"authorization_code", "client_credentials"},
		},
		Local: ports.LocalOAuth2Config{
			SupportedGrantTypes: []string{"authorization_code", "client_credentials"},
		},
	}

	got := cfg.GrantTypes()

	assert.ElementsMatch(t, []string{"authorization_code", "client_credentials"}, got)
	assert.Len(t, got, 2, "no duplicates when both sides are identical")
}

func TestProxyOAuth2Config_JWKSMaxRefresh_GuardsAgainstMinInversion(t *testing.T) {
	tests := []struct {
		name string
		cfg  ports.ProxyOAuth2Config
		want time.Duration
	}{
		{
			name: "zero max clamps to min when min exceeds default",
			cfg: ports.ProxyOAuth2Config{
				UpstreamJWKSMinRefresh: 2 * time.Hour,
			},
			want: 2 * time.Hour,
		},
		{
			name: "configured max below min clamps to min",
			cfg: ports.ProxyOAuth2Config{
				UpstreamJWKSMinRefresh: 2 * time.Hour,
				UpstreamJWKSMaxRefresh: 30 * time.Minute,
			},
			want: 2 * time.Hour,
		},
		{
			name: "defaults remain unchanged",
			cfg:  ports.ProxyOAuth2Config{},
			want: time.Hour,
		},
		{
			name: "configured max above min is preserved",
			cfg: ports.ProxyOAuth2Config{
				UpstreamJWKSMinRefresh: 15 * time.Minute,
				UpstreamJWKSMaxRefresh: 45 * time.Minute,
			},
			want: 45 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cfg.JWKSMaxRefresh())
		})
	}
}
