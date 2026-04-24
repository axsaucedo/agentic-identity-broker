package cimd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClientIDMetadataDocumentURL(t *testing.T) {
	t.Run("accepts valid https URL with path", func(t *testing.T) {
		u, err := ParseClientIDMetadataDocumentURL("https://agent.example.com/client")
		require.NoError(t, err)
		assert.Equal(t, "https://agent.example.com/client", u.String())
	})

	t.Run("accepts explicit port 443", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("https://agent.example.com:443/client")
		require.NoError(t, err)
	})

	t.Run("accepts deep path", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("https://example.com/a/b/c/client")
		require.NoError(t, err)
	})

	t.Run("rejects http scheme", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("http://agent.example.com/client")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "https scheme")
	})

	t.Run("rejects empty path", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("https://agent.example.com")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-empty path")
	})

	t.Run("rejects root path only", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("https://agent.example.com/")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-empty path")
	})

	t.Run("rejects fragment", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("https://agent.example.com/client#section")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fragment")
	})

	t.Run("rejects credentials in userinfo", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("https://user:pass@agent.example.com/client")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "credentials")
	})

	t.Run("rejects non-443 port", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("https://agent.example.com:8443/client")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "port")
	})

	t.Run("rejects single dot segment", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("https://agent.example.com/./client")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dot segments")
	})

	t.Run("rejects double dot segment", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("https://agent.example.com/../client")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dot segments")
	})

	t.Run("rejects missing host", func(t *testing.T) {
		_, err := ParseClientIDMetadataDocumentURL("https:///path")
		require.Error(t, err)
	})
}
