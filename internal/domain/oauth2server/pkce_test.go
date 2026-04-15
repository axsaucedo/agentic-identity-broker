package oauth2server

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyPKCE(t *testing.T) {
	t.Run("valid verifier matches challenge", func(t *testing.T) {
		verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		hash := sha256.Sum256([]byte(verifier))
		challenge := base64.RawURLEncoding.EncodeToString(hash[:])

		err := verifyPKCE(challenge, verifier)
		assert.NoError(t, err)
	})

	t.Run("wrong verifier fails", func(t *testing.T) {
		verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		hash := sha256.Sum256([]byte(verifier))
		challenge := base64.RawURLEncoding.EncodeToString(hash[:])

		wrongVerifier := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" // 43 chars, wrong value
		err := verifyPKCE(challenge, wrongVerifier)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match")
	})

	t.Run("verifier too short rejected", func(t *testing.T) {
		err := verifyPKCE("some-challenge", "tooshort")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "out of range")
	})

	t.Run("verifier too long rejected", func(t *testing.T) {
		err := verifyPKCE("some-challenge", strings.Repeat("A", 129))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "out of range")
	})

	t.Run("verifier at minimum length accepted", func(t *testing.T) {
		verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk" // exactly 43 chars
		hash := sha256.Sum256([]byte(verifier))
		challenge := base64.RawURLEncoding.EncodeToString(hash[:])

		err := verifyPKCE(challenge, verifier)
		assert.NoError(t, err)
	})

	t.Run("verifier at maximum length accepted", func(t *testing.T) {
		verifier := strings.Repeat("A", 128) // 128 valid unreserved chars per RFC 7636 §4.1
		hash := sha256.Sum256([]byte(verifier))
		challenge := base64.RawURLEncoding.EncodeToString(hash[:])

		err := verifyPKCE(challenge, verifier)
		assert.NoError(t, err)
	})

	t.Run("verifier with invalid characters rejected", func(t *testing.T) {
		// '@' is not in [A-Za-z0-9-._~]
		verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEj@k"
		err := verifyPKCE("any-challenge", verifier)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
	})
}
