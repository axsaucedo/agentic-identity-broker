package oauth2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitScope(t *testing.T) {
	t.Run("empty scope", func(t *testing.T) {
		assert.Empty(t, SplitScope(""))
	})

	t.Run("literal space separated scope", func(t *testing.T) {
		assert.Equal(t, []string{"read", "", "write"}, SplitScope("read  write"))
	})
}

func TestIsScopeAllowed(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		scope   string
		want    bool
	}{
		{"allowed scope", []string{"read", "write"}, "read", true},
		{"unlisted scope", []string{"read", "write"}, "admin", false},
		{"empty allow list is unrestricted", nil, "admin", true},
		{"explicitly empty allow list is unrestricted", []string{}, "admin", true},
		{"reserved refresh scope", []string{"read"}, "offline_access", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsScopeAllowed(tc.allowed, tc.scope))
		})
	}
}
