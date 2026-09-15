package urivalidation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCIMDClientURIPathEscapeValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		uri  string
	}{
		{"pattern registration rejects encoded slash", "https://chatgpt.com/oauth/*/tenant%2Fchild/client.json"},
		{"pattern registration rejects lower-case encoded slash", "https://chatgpt.com/oauth/*/tenant%2fchild/client.json"},
		{"pattern registration rejects encoded backslash", "https://chatgpt.com/oauth/*/tenant%5Cchild/client.json"},
		{"pattern registration rejects lower-case encoded backslash", "https://chatgpt.com/oauth/*/tenant%5cchild/client.json"},
		{"pattern registration rejects literal backslash", "https://chatgpt.com/oauth/*/tenant\\child/client.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Error(t, ValidateCIMDClientURIRegistration(tc.uri))
		})
	}

	pattern := "https://chatgpt.com/oauth/codex/*/client.json"
	for _, tc := range []struct {
		name      string
		candidate string
	}{
		{"candidate rejects encoded slash", "https://chatgpt.com/oauth/codex/tenant%2Fchild/client.json"},
		{"candidate rejects encoded backslash", "https://chatgpt.com/oauth/codex/tenant%5Cchild/client.json"},
		{"candidate rejects literal backslash", "https://chatgpt.com/oauth/codex/tenant\\child/client.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, MatchesCIMDClientURI(pattern, tc.candidate))
		})
	}
}
