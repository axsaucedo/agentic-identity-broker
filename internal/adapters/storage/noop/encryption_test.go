package noop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoOpEncryption_Encrypt(t *testing.T) {
	enc := NewNoOpEncryption()
	ctx := context.Background()

	plaintext := []byte("test-secret-value")
	encryptionContext := map[string]string{
		"service_id": "test-service",
	}

	ciphertext, err := enc.Encrypt(ctx, plaintext, encryptionContext)
	require.NoError(t, err)

	// NoOp encryption returns plaintext unchanged
	assert.Equal(t, plaintext, ciphertext)
	assert.Equal(t, "test-secret-value", string(ciphertext))
}

func TestNoOpEncryption_Decrypt(t *testing.T) {
	enc := NewNoOpEncryption()
	ctx := context.Background()

	ciphertext := []byte("test-secret-value")
	encryptionContext := map[string]string{
		"service_id": "test-service",
	}

	plaintext, err := enc.Decrypt(ctx, ciphertext, encryptionContext)
	require.NoError(t, err)

	// NoOp decryption returns ciphertext unchanged
	assert.Equal(t, ciphertext, plaintext)
	assert.Equal(t, "test-secret-value", string(plaintext))
}

func TestNoOpEncryption_RoundTrip(t *testing.T) {
	enc := NewNoOpEncryption()
	ctx := context.Background()

	original := []byte("my-client-secret")
	encryptionContext := map[string]string{
		"service_id": "github",
		"key_id":     "v1",
	}

	// Encrypt
	encrypted, err := enc.Encrypt(ctx, original, encryptionContext)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := enc.Decrypt(ctx, encrypted, encryptionContext)
	require.NoError(t, err)

	// Should match original (because no-op)
	assert.Equal(t, original, decrypted)
}

func TestNoOpEncryption_NilContext(t *testing.T) {
	enc := NewNoOpEncryption()
	ctx := context.Background()

	plaintext := []byte("test-secret")

	// Should work with nil encryption context
	encrypted, err := enc.Encrypt(ctx, plaintext, nil)
	require.NoError(t, err)
	assert.Equal(t, plaintext, encrypted)

	decrypted, err := enc.Decrypt(ctx, encrypted, nil)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
