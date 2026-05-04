package jwe

import (
	"crypto/rand"
	"testing"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

type testClaims struct {
	Subject string `json:"sub"`
	Value   int    `json:"val"`
}

func newTestService(t *testing.T) *TokenService {
	t.Helper()
	var raw [32]byte
	_, err := rand.Read(raw[:])
	require.NoError(t, err)
	key, err := jwk.Import(raw[:])
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
