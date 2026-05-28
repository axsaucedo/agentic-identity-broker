package enduser

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
)

// testEncryptor is a minimal encryption stub for testing.
type testEncryptor struct{}

func (e *testEncryptor) Encrypt(_ context.Context, plaintext []byte, _ map[string]string) ([]byte, error) {
	result := make([]byte, 0, len(plaintext)+4)
	result = append(result, []byte("ENC:")...)
	result = append(result, plaintext...)
	return result, nil
}

func (e *testEncryptor) Decrypt(_ context.Context, ciphertext []byte, _ map[string]string) ([]byte, error) {
	if len(ciphertext) < 4 || string(ciphertext[:4]) != "ENC:" {
		return nil, assert.AnError
	}
	return ciphertext[4:], nil
}

func newTestJWKSHandler(t *testing.T) *JWKSHandler {
	t.Helper()
	repo := memory.NewSigningKeyStore()
	enc := &testEncryptor{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := oauth2server.NewSigningKeyService(repo, enc, nil, logger)

	// Generate a signing key so the JWKS is non-empty
	_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
	require.NoError(t, err)

	return NewJWKSHandler(svc, logger)
}

func TestJWKSHandler_ServeJWKS(t *testing.T) {
	t.Run("returns JWK Set with active public keys", func(t *testing.T) {
		handler := newTestJWKSHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/oauth2/jwks.json", nil)
		rec := httptest.NewRecorder()

		handler.ServeJWKS(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var body map[string]interface{}
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)

		keys, ok := body["keys"].([]interface{})
		require.True(t, ok, "response should contain 'keys' array")
		require.Len(t, keys, 1, "should have exactly one key")

		key := keys[0].(map[string]interface{})
		assert.NotEmpty(t, key["kid"], "key should have a kid")
		assert.Equal(t, "sig", key["use"], "key usage should be 'sig'")
	})

	t.Run("sets Cache-Control header", func(t *testing.T) {
		handler := newTestJWKSHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/oauth2/jwks.json", nil)
		rec := httptest.NewRecorder()

		handler.ServeJWKS(rec, req)

		assert.Equal(t, "public, max-age=300", rec.Header().Get("Cache-Control"))
	})

	t.Run("no private key material in response", func(t *testing.T) {
		handler := newTestJWKSHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/oauth2/jwks.json", nil)
		rec := httptest.NewRecorder()

		handler.ServeJWKS(rec, req)

		// Parse the response body
		var body map[string]interface{}
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)

		keys, ok := body["keys"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, keys)

		// Verify no private key exponent "d" field in any key
		for _, rawKey := range keys {
			key := rawKey.(map[string]interface{})
			_, hasD := key["d"]
			assert.False(t, hasD, "JWKS response must not contain private key exponent 'd'")
		}

		// Also verify via raw body that no "d" JSON key (private exponent) is present
		rawBody := rec.Body.String()
		assert.False(t, strings.Contains(rawBody, `"d":`), "raw response must not contain '\"d\":' private key field")
	})
}
