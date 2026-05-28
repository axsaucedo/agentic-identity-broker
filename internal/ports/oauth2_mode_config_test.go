package ports_test

import (
	"testing"

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
