package urivalidation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidRedirectURI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid https", "https://client.example.com/callback", true},
		{"valid https with port", "https://client.example.com:8080/cb", true},
		{"valid https with query", "https://client.example.com/cb?foo=bar", true},
		{"http localhost allowed", "http://localhost/callback", true},
		{"http 127.0.0.1 allowed", "http://127.0.0.1/callback", true},
		{"http ::1 allowed", "http://[::1]/callback", true},
		{"http non-local rejected", "http://client.example.com/callback", false},
		{"http non-local with port rejected", "http://client.example.com:8080/cb", false},
		{"empty string", "", false},
		{"no scheme", "client.example.com/callback", false},
		{"non-http scheme", "ftp://client.example.com/callback", false},
		{"no host", "https:///callback", false},
		{"fragment present", "https://client.example.com/callback#section", false},
		{"empty fragment", "https://client.example.com/callback#", false},
		{"space in path", "https://client.example.com/call back", false},
		{"relative path", "/callback", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidRedirectURI(tt.input))
		})
	}
}

func TestMatchesRedirectURI(t *testing.T) {
	cases := []struct {
		name       string
		registered string
		incoming   string
		want       bool
	}{
		// Loopback — port ignored
		{"loopback: portless reg, port in request", "http://localhost/cb", "http://localhost:52341/cb", true},
		{"loopback: explicit port reg, different port", "http://localhost:3000/cb", "http://localhost:9999/cb", true},
		{"loopback: explicit port reg, portless request", "http://localhost:3000/cb", "http://localhost/cb", true},
		{"loopback: both portless", "http://localhost/cb", "http://localhost/cb", true},
		{"loopback: 127.0.0.1 different ports", "http://127.0.0.1:8080/cb", "http://127.0.0.1:51234/cb", true},
		{"loopback: [::1] different ports", "http://[::1]:8080/cb", "http://[::1]:51234/cb", true},
		// Loopback — other components must still match
		{"loopback: path mismatch", "http://localhost/cb", "http://localhost:3000/other", false},
		{"loopback: scheme mismatch", "http://localhost/cb", "https://localhost/cb", false},
		{"loopback: host mismatch (localhost vs 127.0.0.1)", "http://localhost/cb", "http://127.0.0.1/cb", false},
		// Non-loopback — exact match required
		{"non-loopback: exact match", "https://app.example.com/cb", "https://app.example.com/cb", true},
		{"non-loopback: port differs", "https://app.example.com/cb", "https://app.example.com:9999/cb", false},
		{"non-loopback: explicit port differs", "https://app.example.com:8443/cb", "https://app.example.com:9000/cb", false},
		{"non-loopback: portless reg, port in request", "https://app.example.com/cb", "https://app.example.com:443/cb", false},
		// Defense-in-depth — reject malformed values even if legacy data bypassed write validation
		{"invalid: raw fragment rejected", "https://app.example.com/cb#frag", "https://app.example.com/cb#frag", false},
		{"invalid: raw whitespace rejected", "https://app.example.com/call back", "https://app.example.com/call back", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MatchesRedirectURI(tc.registered, tc.incoming)
			if got != tc.want {
				t.Errorf("MatchesRedirectURI(%q, %q) = %v, want %v", tc.registered, tc.incoming, got, tc.want)
			}
		})
	}
}
