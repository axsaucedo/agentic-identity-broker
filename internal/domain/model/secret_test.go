package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlaintextSecret(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{name: "normal secret", plaintext: "mysecret"},
		{name: "long secret", plaintext: strings.Repeat("x", 1024)},
		{name: "special characters", plaintext: "s3cr3t!@#$%^&*()-_=+"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewPlaintextSecret(tt.plaintext)
			assert.True(t, s.IsPlaintext())
			assert.False(t, s.IsEncrypted())

			got, err := s.GetPlaintext()
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, got)
		})
	}
}

func TestNewEncryptedSecret(t *testing.T) {
	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{name: "normal ciphertext", ciphertext: []byte{0x01, 0x02, 0x03, 0xAB, 0xCD, 0xEF}},
		{name: "long ciphertext", ciphertext: make([]byte, 256)},
		{name: "single byte", ciphertext: []byte{0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewEncryptedSecret(tt.ciphertext)
			assert.True(t, s.IsEncrypted())
			assert.False(t, s.IsPlaintext())

			got, err := s.GetCiphertext()
			require.NoError(t, err)
			assert.Equal(t, tt.ciphertext, got)
		})
	}
}

func TestSecret_GetPlaintext_Errors(t *testing.T) {
	tests := []struct {
		name      string
		secret    Secret
		wantErrIn string
	}{
		{
			name:      "encrypted state returns error",
			secret:    NewEncryptedSecret([]byte{0x01, 0x02}),
			wantErrIn: "encrypted state",
		},
		{
			name:      "empty plaintext returns error",
			secret:    NewPlaintextSecret(""),
			wantErrIn: "plaintext is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.secret.GetPlaintext()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrIn)
			assert.Empty(t, got)
		})
	}
}

func TestSecret_GetCiphertext_Errors(t *testing.T) {
	tests := []struct {
		name      string
		secret    Secret
		wantErrIn string
	}{
		{
			name:      "plaintext state returns error",
			secret:    NewPlaintextSecret("mysecret"),
			wantErrIn: "plaintext state",
		},
		{
			name:      "empty ciphertext returns error",
			secret:    NewEncryptedSecret([]byte{}),
			wantErrIn: "ciphertext is empty",
		},
		{
			name:      "nil ciphertext returns error",
			secret:    NewEncryptedSecret(nil),
			wantErrIn: "ciphertext is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.secret.GetCiphertext()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrIn)
			assert.Nil(t, got)
		})
	}
}

func TestSecret_Redacted(t *testing.T) {
	t.Run("plaintext state", func(t *testing.T) {
		s := NewPlaintextSecret("mysecret")
		assert.Equal(t, "REDACTED", s.Redacted())
	})

	t.Run("encrypted state", func(t *testing.T) {
		s := NewEncryptedSecret([]byte{0x01, 0x02})
		assert.Equal(t, "REDACTED", s.Redacted())
	})

	t.Run("zero value", func(t *testing.T) {
		var s Secret
		assert.Equal(t, "REDACTED", s.Redacted())
	})
}

func TestSecret_ImmutabilityOnCreate(t *testing.T) {
	t.Run("NewEncryptedSecret does not share backing array", func(t *testing.T) {
		original := []byte{0x01, 0x02, 0x03}
		s := NewEncryptedSecret(original)

		// Mutate the original slice
		original[0] = 0xFF

		got, err := s.GetCiphertext()
		require.NoError(t, err)
		assert.Equal(t, byte(0x01), got[0], "Secret should not reflect mutation of original slice")
	})

	t.Run("GetCiphertext returns independent copy", func(t *testing.T) {
		s := NewEncryptedSecret([]byte{0x01, 0x02, 0x03})

		ct1, err := s.GetCiphertext()
		require.NoError(t, err)
		ct1[0] = 0xFF

		ct2, err := s.GetCiphertext()
		require.NoError(t, err)
		assert.Equal(t, byte(0x01), ct2[0], "Subsequent GetCiphertext call should return independent copy")
	})
}

func TestSecret_SecurityNoLeakage(t *testing.T) {
	t.Run("error message from GetPlaintext on encrypted secret does not leak ciphertext", func(t *testing.T) {
		sensitiveBytes := []byte("this-should-never-appear-in-error")
		s := NewEncryptedSecret(sensitiveBytes)

		_, err := s.GetPlaintext()
		require.Error(t, err)
		assert.NotContains(t, err.Error(), string(sensitiveBytes))
	})

	t.Run("error message from GetCiphertext on plaintext secret does not leak plaintext", func(t *testing.T) {
		sensitive := "super-secret-password-123"
		s := NewPlaintextSecret(sensitive)

		_, err := s.GetCiphertext()
		require.Error(t, err)
		assert.NotContains(t, err.Error(), sensitive)
	})

	t.Run("Redacted never exposes plaintext value", func(t *testing.T) {
		sensitive := "super-secret-password-123"
		s := NewPlaintextSecret(sensitive)
		assert.NotContains(t, s.Redacted(), sensitive)
	})

	t.Run("Redacted never exposes ciphertext value", func(t *testing.T) {
		s := NewEncryptedSecret([]byte{0xDE, 0xAD, 0xBE, 0xEF})
		// Redacted() must return fixed string, not hex-encoded or base64-encoded ciphertext
		assert.Equal(t, "REDACTED", s.Redacted())
	})
}

func TestSecret_ZeroValue(t *testing.T) {
	var s Secret

	t.Run("zero value is plaintext state", func(t *testing.T) {
		assert.False(t, s.IsEncrypted())
		assert.True(t, s.IsPlaintext())
	})

	t.Run("zero value GetPlaintext returns error for empty", func(t *testing.T) {
		_, err := s.GetPlaintext()
		require.Error(t, err)
	})
}
