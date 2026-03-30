package oauth2server

import (
	"crypto/sha256"
	"encoding/base64"
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
		verifier := "correct-verifier"
		hash := sha256.Sum256([]byte(verifier))
		challenge := base64.RawURLEncoding.EncodeToString(hash[:])

		err := verifyPKCE(challenge, "wrong-verifier")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match")
	})

	t.Run("empty verifier fails", func(t *testing.T) {
		err := verifyPKCE("some-challenge", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code_verifier is required")
	})
}
