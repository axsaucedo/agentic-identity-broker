package urivalidation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCIMDClientURIRegistration(t *testing.T) {
	t.Run("accepts a whole path segment wildcard", func(t *testing.T) {
		err := ValidateCIMDClientURIRegistration("https://chatgpt.com/oauth/codex/*/client.json")
		require.NoError(t, err)
	})

	t.Run("accepts a wildcard as the first path segment", func(t *testing.T) {
		err := ValidateCIMDClientURIRegistration("https://chatgpt.com/*/client.json")
		require.NoError(t, err)
	})

	t.Run("accepts a literal client URI", func(t *testing.T) {
		err := ValidateCIMDClientURIRegistration("https://chatgpt.com/oauth/codex/client.json")
		require.NoError(t, err)
	})

	for _, tc := range []struct {
		name string
		uri  string
	}{
		{"rejects partial wildcard segment", "https://chatgpt.com/oauth/codex-*/client.json"},
		{"rejects wildcard in host", "https://*.chatgpt.com/oauth/codex/client.json"},
		{"rejects wildcard in query", "https://chatgpt.com/oauth/codex/client.json?installation=*"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Error(t, ValidateCIMDClientURIRegistration(tc.uri))
		})
	}
}

func TestMatchesCIMDClientURI(t *testing.T) {
	pattern := "https://chatgpt.com/oauth/codex/*/client.json"

	for _, tc := range []struct {
		name      string
		candidate string
		want      bool
	}{
		{"matches one installation segment", "https://chatgpt.com/oauth/codex/dIwd44EtAHp-/client.json", true},
		{"rejects extra path segment", "https://chatgpt.com/oauth/codex/dIwd44EtAHp-/nested/client.json", false},
		{"rejects different literal segment", "https://chatgpt.com/oauth/other/dIwd44EtAHp-/client.json", false},
		{"rejects different host", "https://other.example/oauth/codex/dIwd44EtAHp-/client.json", false},
		{"rejects differently cased host", "https://ChatGPT.com/oauth/codex/dIwd44EtAHp-/client.json", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, MatchesCIMDClientURI(pattern, tc.candidate))
		})
	}
}
