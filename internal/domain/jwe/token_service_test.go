package jwe

import (
	"crypto/rand"
	"testing"

	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/stretchr/testify/require"
)

type testClaims struct {
	Subject string `json:"sub"`
	Value   int    `json:"val"`
}

type expirableClaims struct {
	Subject string `json:"sub"`
	Expired bool   `json:"expired"`
}

func (c *expirableClaims) IsExpired() bool { return c.Expired }

func newTestService(t *testing.T) *TokenService {
	t.Helper()
	var raw [32]byte
	_, err := rand.Read(raw[:])
	require.NoError(t, err)
	key, err := jwk.Import[jwk.Key](raw[:])
	require.NoError(t, err)
	return New(key)
}

func TestTokenService_RoundTrip(t *testing.T) {
	svc := newTestService(t)
	original := testClaims{Subject: "user@example.com", Value: 42}

	token, err := svc.Encrypt(original)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	var decoded testClaims
	err = svc.Decrypt(token, &decoded)
	require.NoError(t, err)
	require.Equal(t, original, decoded)
}

func TestTokenService_TamperedToken(t *testing.T) {
	svc := newTestService(t)
	token, err := svc.Encrypt(testClaims{Subject: "user@example.com"})
	require.NoError(t, err)

	tampered := token[:len(token)-4] + "XXXX"
	var decoded testClaims
	err = svc.Decrypt(tampered, &decoded)
	require.Error(t, err)
}

func TestTokenService_EmptyToken(t *testing.T) {
	svc := newTestService(t)
	var decoded testClaims
	err := svc.Decrypt("", &decoded)
	require.Error(t, err)
}

func TestTokenService_DecryptAndValidate(t *testing.T) {
	svc := newTestService(t)

	t.Run("valid non-expired token succeeds", func(t *testing.T) {
		token, err := svc.Encrypt(expirableClaims{Subject: "user@example.com", Expired: false})
		require.NoError(t, err)
		var decoded expirableClaims
		err = svc.DecryptAndValidate(token, &decoded)
		require.NoError(t, err)
		require.Equal(t, "user@example.com", decoded.Subject)
	})

	t.Run("expired token returns error", func(t *testing.T) {
		token, err := svc.Encrypt(expirableClaims{Subject: "user@example.com", Expired: true})
		require.NoError(t, err)
		var decoded expirableClaims
		err = svc.DecryptAndValidate(token, &decoded)
		require.Error(t, err)
		require.Contains(t, err.Error(), "expired")
	})

	t.Run("tampered token returns error", func(t *testing.T) {
		token, err := svc.Encrypt(expirableClaims{Subject: "user@example.com"})
		require.NoError(t, err)
		tampered := token[:len(token)-4] + "XXXX"
		var decoded expirableClaims
		err = svc.DecryptAndValidate(tampered, &decoded)
		require.Error(t, err)
	})
}
