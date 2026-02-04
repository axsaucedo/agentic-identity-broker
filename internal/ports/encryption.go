package ports

import "context"

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
// Branch keys are service-specific KEK cache entries in DynamoDB used by the hierarchical keyring.
// Each service gets a deterministic branch key: service_{service_id}_branch_key
// This follows the Repository pattern for infrastructure provisioning (not data persistence).
type BranchKeyRepository interface {
	// Create creates a branch key for a service in the key store.
	// The branch key ID follows the pattern: service_{service_id}_branch_key
	// Returns the generated branch key ID or error if provisioning fails.
	// ATOMIC: Should fail immediately if KMS/DynamoDB operations fail - no partial state.
	Create(ctx context.Context, serviceID string) (string, error)
}

// BranchKeyIdProvider defines the interface for generating and parsing branch key IDs.
// This abstraction eliminates duplicate branch key ID logic across different encryption adapters.
// Both AWS and Memory implementations use the same deterministic ID format: service_{service_id}_branch_key
type BranchKeyIdProvider interface {
	// GenerateBranchKeyId generates a deterministic branch key ID from a service ID.
	// Format: service_{service_id}_branch_key
	// Example: service_oauth2_branch_key, service_github_branch_key
	GenerateBranchKeyId(serviceID string) string

	// ExtractServiceIdFromBranchKey extracts the service ID from a branch key ID.
	// This is the inverse operation of GenerateBranchKeyId.
	// Format: service_{service_id}_branch_key -> service_id
	// Example: service_oauth2_branch_key -> oauth2
	// Returns empty string if parsing fails.
	ExtractServiceIdFromBranchKey(branchKeyID string) string
}

// BranchKeyManager is an alias for BranchKeyRepository, consolidating branch key lifecycle management.
// This provides a unified interface for branch key provisioning during service creation.
//
// Architecture:
// - Create: Provisions a new branch key for a service in DynamoDB
// - Get: Retrieves an existing branch key (future use)
// - ID resolution at runtime happens via BranchKeyIdSupplier (not this interface)
//
// This separation enables:
// - Clean EncryptionPort that handles only Encrypt/Decrypt operations
// - Simple, focused branch key provisioning interface
// - Easy testing and mocking of provisioning logic separately from encryption
type BranchKeyManager = BranchKeyRepository
