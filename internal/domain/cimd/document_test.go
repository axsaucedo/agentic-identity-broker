package cimd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validDoc(overrides map[string]any) []byte {
	doc := map[string]any{
		"client_id":     "https://agent.example.com/client",
		"client_name":   "Test Agent",
		"redirect_uris": []string{"https://agent.example.com/callback"},
	}
	for k, v := range overrides {
		if v == nil {
			delete(doc, k)
		} else {
			doc[k] = v
		}
	}
	data, _ := json.Marshal(doc)
	return data
}

const fetchURL = "https://agent.example.com/client"

func TestParseDocument(t *testing.T) {
	t.Run("accepts valid document", func(t *testing.T) {
		doc, err := ParseDocument(validDoc(nil), fetchURL, nil)
		require.NoError(t, err)
		assert.Equal(t, fetchURL, doc.ClientID)
		assert.Equal(t, "Test Agent", doc.ClientName)
		assert.Equal(t, []string{"https://agent.example.com/callback"}, doc.RedirectURIs)
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		_, err := ParseDocument([]byte("not json"), fetchURL, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "malformed")
	})

	t.Run("rejects client_id mismatch", func(t *testing.T) {
		data := validDoc(map[string]any{"client_id": "https://other.example.com/client"})
		_, err := ParseDocument(data, fetchURL, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client_id mismatch")
	})

	t.Run("rejects empty redirect_uris", func(t *testing.T) {
		data := validDoc(map[string]any{"redirect_uris": []string{}})
		_, err := ParseDocument(data, fetchURL, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "redirect_uris")
	})

	t.Run("rejects missing redirect_uris", func(t *testing.T) {
		data := validDoc(map[string]any{"redirect_uris": nil})
		_, err := ParseDocument(data, fetchURL, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "redirect_uris")
	})

	t.Run("rejects client_secret_post auth method", func(t *testing.T) {
		data := validDoc(map[string]any{"token_endpoint_auth_method": "client_secret_post"})
		_, err := ParseDocument(data, fetchURL, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client_secret_post")
	})

	t.Run("rejects client_secret_basic auth method", func(t *testing.T) {
		data := validDoc(map[string]any{"token_endpoint_auth_method": "client_secret_basic"})
		_, err := ParseDocument(data, fetchURL, nil)
		require.Error(t, err)
	})

	t.Run("rejects client_secret_jwt auth method", func(t *testing.T) {
		data := validDoc(map[string]any{"token_endpoint_auth_method": "client_secret_jwt"})
		_, err := ParseDocument(data, fetchURL, nil)
		require.Error(t, err)
	})

	t.Run("accepts none auth method", func(t *testing.T) {
		data := validDoc(map[string]any{"token_endpoint_auth_method": "none"})
		_, err := ParseDocument(data, fetchURL, nil)
		require.NoError(t, err)
	})

	t.Run("accepts private_key_jwt auth method", func(t *testing.T) {
		data := validDoc(map[string]any{"token_endpoint_auth_method": "private_key_jwt"})
		_, err := ParseDocument(data, fetchURL, nil)
		require.NoError(t, err)
	})

	t.Run("rejects blocked client_name keyword", func(t *testing.T) {
		data := validDoc(map[string]any{"client_name": "Admin"})
		_, err := ParseDocument(data, fetchURL, []string{"Admin"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "blocked term")
	})

	t.Run("keyword check is case-insensitive", func(t *testing.T) {
		data := validDoc(map[string]any{"client_name": "ADMIN"})
		_, err := ParseDocument(data, fetchURL, []string{"admin"})
		require.Error(t, err)
	})

	t.Run("partial name not blocked by exact match", func(t *testing.T) {
		data := validDoc(map[string]any{"client_name": "Admin Panel"})
		_, err := ParseDocument(data, fetchURL, []string{"admin"})
		require.NoError(t, err, "partial substring must not trigger exact match blocklist")
	})

	t.Run("allows client_name not in blocklist", func(t *testing.T) {
		data := validDoc(map[string]any{"client_name": "Helpful Assistant"})
		_, err := ParseDocument(data, fetchURL, []string{"admin"})
		require.NoError(t, err)
	})

	t.Run("rejects cross-origin redirect URI", func(t *testing.T) {
		data := validDoc(map[string]any{
			"redirect_uris": []string{"https://other.example.com/callback"},
		})
		_, err := ParseDocument(data, fetchURL, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "same-origin")
	})

	t.Run("allows localhost redirect URI (exception)", func(t *testing.T) {
		data := validDoc(map[string]any{
			"redirect_uris": []string{
				"https://agent.example.com/callback",
				"http://localhost/callback",
			},
		})
		_, err := ParseDocument(data, fetchURL, nil)
		require.NoError(t, err)
	})

	t.Run("allows 127.0.0.1 redirect URI (exception)", func(t *testing.T) {
		data := validDoc(map[string]any{
			"redirect_uris": []string{"http://127.0.0.1/callback"},
		})
		_, err := ParseDocument(data, fetchURL, nil)
		require.NoError(t, err)
	})

	t.Run("accepts same-origin logo_uri", func(t *testing.T) {
		data := validDoc(map[string]any{"logo_uri": "https://agent.example.com/logo.png"})
		_, err := ParseDocument(data, fetchURL, nil)
		require.NoError(t, err)
	})

	t.Run("rejects logo_uri on different host", func(t *testing.T) {
		data := validDoc(map[string]any{"logo_uri": "https://cdn.example.com/logo.png"})
		_, err := ParseDocument(data, fetchURL, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "logo_uri")
	})

	t.Run("rejects logo_uri on same host but different port", func(t *testing.T) {
		data := validDoc(map[string]any{"logo_uri": "https://agent.example.com:8443/logo.png"})
		_, err := ParseDocument(data, fetchURL, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "same-origin")
	})

	t.Run("rejects http logo_uri", func(t *testing.T) {
		data := validDoc(map[string]any{"logo_uri": "http://agent.example.com/logo.png"})
		_, err := ParseDocument(data, fetchURL, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "HTTPS")
	})

	t.Run("accepts same-origin logo_uri when client_id has explicit port", func(t *testing.T) {
		fetchURLPort := "https://agent.example.com:9000/client"
		data, _ := json.Marshal(map[string]any{
			"client_id":     fetchURLPort,
			"redirect_uris": []string{"https://agent.example.com:9000/callback"},
			"logo_uri":      "https://agent.example.com:9000/logo.png",
		})
		_, err := ParseDocument(data, fetchURLPort, nil)
		require.NoError(t, err)
	})

	t.Run("rejects logo_uri with different port than client_id explicit port", func(t *testing.T) {
		fetchURLPort := "https://agent.example.com:9000/client"
		data, _ := json.Marshal(map[string]any{
			"client_id":     fetchURLPort,
			"redirect_uris": []string{"https://agent.example.com:9000/callback"},
			"logo_uri":      "https://agent.example.com:8443/logo.png",
		})
		_, err := ParseDocument(data, fetchURLPort, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "same-origin")
	})
}
