package enduser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryHandler_ServeDiscovery(t *testing.T) {
	t.Run("returns correct issuer and endpoints", func(t *testing.T) {
		issuer := "https://broker.example.com"
		handler := NewDiscoveryHandler(issuer)

		req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
		rec := httptest.NewRecorder()

		handler.ServeDiscovery(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var body map[string]interface{}
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, issuer, body["issuer"])
		assert.Equal(t, issuer+"/oauth2/authorize", body["authorization_endpoint"])
		assert.Equal(t, issuer+"/oauth2/token", body["token_endpoint"])
		assert.Equal(t, issuer+"/oauth2/jwks.json", body["jwks_uri"])

		responseTypes, ok := body["response_types_supported"].([]interface{})
		require.True(t, ok)
		assert.Contains(t, responseTypes, "code")

		grantTypes, ok := body["grant_types_supported"].([]interface{})
		require.True(t, ok)
		assert.Contains(t, grantTypes, "authorization_code")
		assert.Contains(t, grantTypes, "client_credentials")

		challengeMethods, ok := body["code_challenge_methods_supported"].([]interface{})
		require.True(t, ok)
		assert.Contains(t, challengeMethods, "S256")
	})

	t.Run("sets Cache-Control header", func(t *testing.T) {
		handler := NewDiscoveryHandler("https://broker.example.com")

		req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
		rec := httptest.NewRecorder()

		handler.ServeDiscovery(rec, req)

		assert.Equal(t, "public, max-age=3600", rec.Header().Get("Cache-Control"))
	})
}
