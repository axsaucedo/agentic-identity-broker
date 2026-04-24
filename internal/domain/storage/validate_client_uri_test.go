package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateClientURIs_MultipleRejected(t *testing.T) {
	err := validateClientURIs([]string{
		"https://agent.example.com/client-a",
		"https://agent.example.com/client-b",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at most one CIMD")
}

func TestValidateClientURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		wantErr string
	}{
		{name: "valid https no port", uri: "https://example.com/client", wantErr: ""},
		{name: "valid https port 443", uri: "https://example.com:443/client", wantErr: ""},
		{name: "http scheme rejected", uri: "http://example.com/client", wantErr: "must use https scheme"},
		{name: "non-443 port rejected", uri: "https://example.com:8080/client", wantErr: "port must be 443 or absent"},
		{name: "malformed port rejected", uri: "https://example.com:bad/client", wantErr: "invalid URL"},
		{name: "port zero rejected", uri: "https://example.com:0/client", wantErr: "port is not a valid port number"},
		{name: "empty path rejected", uri: "https://example.com", wantErr: "must have a non-empty path"},
		{name: "root path rejected", uri: "https://example.com/", wantErr: "must have a non-empty path"},
		{name: "fragment rejected", uri: "https://example.com/client#frag", wantErr: "must not contain a fragment"},
		{name: "dot segment rejected", uri: "https://example.com/./client", wantErr: "dot segments"},
		{name: "dot-dot segment rejected", uri: "https://example.com/../client", wantErr: "dot segments"},
		{name: "credentials rejected", uri: "https://user:pass@example.com/client", wantErr: "must not contain credentials"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateClientURI(tc.uri)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			}
		})
	}
}
