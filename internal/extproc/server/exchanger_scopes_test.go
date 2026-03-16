package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClientCredentialsScopes covers the fallback logic of the
// clientCredentialsScopes helper: nil/empty input → ["openid"], otherwise
// returns the configured slice unchanged.
func TestClientCredentialsScopes(t *testing.T) {
	tests := []struct {
		name       string
		configured []string
		want       []string
	}{
		{
			name:       "nil returns default openid",
			configured: nil,
			want:       []string{"openid"},
		},
		{
			name:       "empty slice returns default openid",
			configured: []string{},
			want:       []string{"openid"},
		},
		{
			name:       "single configured scope is returned unchanged",
			configured: []string{"profile"},
			want:       []string{"profile"},
		},
		{
			name:       "multiple configured scopes are returned unchanged",
			configured: []string{"openid", "profile", "email"},
			want:       []string{"openid", "profile", "email"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := clientCredentialsScopes(tc.configured)
			assert.Equal(t, tc.want, got)
		})
	}
}
