package ports

import "context"

// EncryptionPort defines the interface for encrypting and decrypting sensitive data.
// This port allows the domain to remain independent of specific encryption implementations.
// Production implementations should use AES-256-GCM or similar authenticated encryption.
type EncryptionPort interface {
	// Encrypt encrypts plaintext data with optional encryption context.
	// The encryption context provides additional authenticated data (AAD) that is
	// cryptographically bound to the ciphertext but not encrypted.
	// Returns the encrypted ciphertext or an error if encryption fails.
	Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error)

	// Decrypt decrypts ciphertext data with optional encryption context.
	// The encryption context must match the context used during encryption.
	// Returns the decrypted plaintext or an error if decryption fails or context doesn't match.
	Decrypt(ctx context.Context, ciphertext []byte, encryptionContext map[string]string) ([]byte, error)
}
