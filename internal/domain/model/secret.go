package model

import "errors"

// Secret is a value object representing an OAuth2 client secret in one of two mutually
// exclusive states: plaintext or encrypted. The two-state design enforces at the type
// level that plaintext secrets are never persisted and encrypted bytes are never exposed
// as strings.
//
// State is determined by the invariant: ciphertext == nil means plaintext state, ciphertext != nil means encrypted state.
// A plaintext Secret is created via NewPlaintextSecret and holds the raw secret string.
// An encrypted Secret is created via NewEncryptedSecret and holds opaque ciphertext bytes.
// State transitions produce new Secret instances — Secret is immutable.
type Secret struct {
	plaintext  string
	ciphertext []byte
}

// NewPlaintextSecret creates a Secret in plaintext state.
// The plaintext value can be any string, including empty (validation happens at domain entity level).
func NewPlaintextSecret(plaintext string) Secret {
	return Secret{
		plaintext:  plaintext,
		ciphertext: nil,
	}
}

// NewEncryptedSecret creates a Secret in encrypted state.
// The ciphertext holds opaque encrypted bytes; use GetCiphertext() to retrieve them.
func NewEncryptedSecret(ciphertext []byte) Secret {
	ct := make([]byte, len(ciphertext))
	copy(ct, ciphertext)
	return Secret{
		plaintext:  "",
		ciphertext: ct,
	}
}

// GetPlaintext returns the plaintext value of the secret.
// Returns an error if the secret is in encrypted state or if plaintext is empty.
func (s Secret) GetPlaintext() (string, error) {
	if s.ciphertext != nil {
		return "", errors.New("secret is in encrypted state: ciphertext cannot be converted to plaintext")
	}
	if s.plaintext == "" {
		return "", errors.New("secret plaintext is empty")
	}
	return s.plaintext, nil
}

// GetCiphertext returns the ciphertext bytes of the secret.
// Returns an error if the secret is in plaintext state, or if the ciphertext is empty.
func (s Secret) GetCiphertext() ([]byte, error) {
	if s.ciphertext == nil {
		return nil, errors.New("secret is in plaintext state: plaintext cannot be converted to ciphertext")
	}
	if len(s.ciphertext) == 0 {
		return nil, errors.New("secret ciphertext is empty")
	}
	ct := make([]byte, len(s.ciphertext))
	copy(ct, s.ciphertext)
	return ct, nil
}

// IsEncrypted reports whether the secret is in encrypted state.
// A secret is encrypted if ciphertext is not nil (even if empty).
func (s Secret) IsEncrypted() bool {
	return s.ciphertext != nil
}

// IsPlaintext reports whether the secret is in plaintext state.
// A secret is plaintext if ciphertext is nil.
func (s Secret) IsPlaintext() bool {
	return s.ciphertext == nil
}

// Redacted returns the string "REDACTED" regardless of the secret's state.
// Use this for API responses and log output to avoid exposing the secret value.
func (s Secret) Redacted() string {
	return "REDACTED"
}
