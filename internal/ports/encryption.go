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

	// Get retrieves a branch key ID for the given service.
	// Returns the branch key ID if found, error if not found or query fails.
	// Returns error if branch key does not exist.
	Get(ctx context.Context, serviceID string) (string, error)
}

// BranchKeyManager defines unified operations for branch key lifecycle management.
// Consolidates both branch key ID resolution (for runtime encryption context binding)
// and branch key provisioning (for infrastructure setup during service creation).
//
// This separation of concerns enables:
// - Clean EncryptionPort that handles only Encrypt/Decrypt operations
// - Centralized branch key operations without adapter overload
// - Easy testing and mocking of provisioning logic separately from encryption
type BranchKeyManager interface {
	BranchKeyRepository

	// ResolveBranchKeyID returns a deterministic branch key identifier based on service_id.
	// Used by the hierarchical keyring during encryption/decryption to resolve the correct branch key.
	// This is a pure function - same service_id always returns the same branch key ID.
	//
	// Parameters:
	//   - serviceID: Service identifier from encryption context
	//
	// Returns:
	//   - Branch key ID in format: service_{service_id}_branch_key
	//   - error: If service_id is empty or invalid
	ResolveBranchKeyID(serviceID string) (string, error)
}
