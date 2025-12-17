package noop

import "context"

// NoOpEncryption is a pass-through encryption adapter for development and testing.
// WARNING: This adapter does NOT encrypt data. It should only be used in development
// or testing environments where encryption at rest is not required.
// Production deployments MUST use a proper encryption adapter (e.g., AES-256-GCM).
type NoOpEncryption struct{}

// NewNoOpEncryption creates a new no-op encryption adapter.
func NewNoOpEncryption() *NoOpEncryption {
	return &NoOpEncryption{}
}

// Encrypt returns the plaintext unchanged (no encryption).
func (n *NoOpEncryption) Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error) {
	// No-op: return plaintext as-is
	return plaintext, nil
}

// Decrypt returns the ciphertext unchanged (no decryption).
func (n *NoOpEncryption) Decrypt(ctx context.Context, ciphertext []byte, encryptionContext map[string]string) ([]byte, error) {
	// No-op: return ciphertext as-is
	return ciphertext, nil
}
