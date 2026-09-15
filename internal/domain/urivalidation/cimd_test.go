package urivalidation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCIMDClientURL(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cases := []string{
			"https://example.com/client",
			"https://example.com:443/client",
			"https://example.com/a/b/c/client",
			"https://example.com/client?key=val",
		}
		for _, uri := range cases {
			require.NoError(t, ValidateCIMDClientURL(uri), "expected valid: %s", uri)
		}
	})

	tests := []struct {
		name    string
		uri     string
		wantErr string
	}{
		{name: "http scheme", uri: "http://example.com/client", wantErr: "must use https scheme"},
		{name: "non-443 port", uri: "https://example.com:8080/client", wantErr: "port must be 443 or absent"},
		{name: "non-443 loopback port", uri: "https://localhost:8080/client", wantErr: "port"},
		{name: "port zero", uri: "https://example.com:0/client", wantErr: "port is not a valid port number"},
		{name: "empty path", uri: "https://example.com", wantErr: "non-empty path"},
		{name: "root path", uri: "https://example.com/", wantErr: "non-empty path"},
		{name: "fragment", uri: "https://example.com/client#frag", wantErr: "fragment"},
		{name: "dot segment", uri: "https://example.com/./client", wantErr: "dot segments"},
		{name: "dot-dot segment", uri: "https://example.com/../client", wantErr: "dot segments"},
		{name: "credentials", uri: "https://user:pass@example.com/client", wantErr: "credentials"},
		{name: "wildcard", uri: "https://example.com/*/client", wantErr: "wildcard"},
		{name: "encoded slash", uri: "https://example.com/tenant%2Fchild/client", wantErr: "encoded path separator"},
		{name: "encoded backslash", uri: "https://example.com/tenant%5Cchild/client", wantErr: "encoded path separator"},
		{name: "literal backslash", uri: "https://example.com/tenant\\child/client", wantErr: "path separator"},
		{name: "trailing colon empty port", uri: "https://example.com:/client", wantErr: "malformed authority"},
		{name: "empty hostname with port", uri: "https://:443/client", wantErr: ""},
		{name: "missing host", uri: "https:///path", wantErr: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCIMDClientURL(tc.uri)
			require.Error(t, err)
			if tc.wantErr != "" {
				assert.Contains(t, err.Error(), tc.wantErr)
			}
		})
	}
}
