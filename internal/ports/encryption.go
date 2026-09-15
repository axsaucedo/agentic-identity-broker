package ports

import (
	"context"

	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
)

// EncryptionPort defines the interface for encrypting and decrypting sensitive data.
// This port allows the domain to remain independent of specific encryption implementations.
// Production implementations perform envelope encryption (DEK + KEK wrapping) with context binding
// for service isolation. Implementations should use AES-256-GCM or AESGCMSIV for authenticated encryption.
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

// BranchKeyRepository defines operations for provisioning and managing branch keys.
// Branch keys are subject-specific KEK cache entries in DynamoDB used by the hierarchical keyring.
// Service subjects keep the production branch key naming scheme `service_{service_id}_branch_key`.
// Other subject kinds (for example signing keys) use their own deterministic namespaces.
// This follows the Repository pattern for infrastructure provisioning (not data persistence).
type BranchKeyRepository interface {
	// Create creates a branch key for a subject in the key store.
	// Returns the generated branch key ID or error if provisioning fails.
	// ATOMIC: Should fail immediately if KMS/DynamoDB operations fail - no partial state.
	Create(ctx context.Context, subject domainencryption.BranchKeySubject) (string, error)
}

// BranchKeyIdProvider defines the interface for generating and parsing branch key IDs.
// This abstraction eliminates duplicate branch key ID logic across different encryption adapters.
type BranchKeyIdProvider interface {
	// GenerateBranchKeyId generates a deterministic branch key ID from a branch key subject.
	GenerateBranchKeyId(subject domainencryption.BranchKeySubject) (string, error)

	// ExtractSubjectFromBranchKey parses a branch key ID back into its subject.
	ExtractSubjectFromBranchKey(branchKeyID string) (domainencryption.BranchKeySubject, error)
}

// BranchKeyManager is an alias for BranchKeyRepository, consolidating branch key lifecycle management.
// This provides a unified interface for branch key provisioning during subject creation.
//
// Architecture:
// - Create: Provisions a branch key for a typed subject in DynamoDB
// - Runtime ID resolution happens via BranchKeyIdSupplier (not this interface)
//
// This separation enables:
// - Clean EncryptionPort that handles only Encrypt/Decrypt operations
// - Simple, focused branch key provisioning interface
// - Easy testing and mocking of provisioning logic separately from encryption
type BranchKeyManager = BranchKeyRepository
