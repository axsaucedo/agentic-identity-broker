package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"time"

	mpl "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygenerated"
	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
	client "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygenerated"
	esdktypes "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygeneratedtypes"
	"github.com/aws/aws-sdk-go-v2/service/kms"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

const (
	// DefaultKEKValidationTimeout is the timeout for KMS key accessibility checks
	DefaultKEKValidationTimeout = 30 * time.Second
	// DefaultBranchKeyTTL is the TTL for branch keys in the DynamoDB cache
	DefaultBranchKeyTTL = 1 * time.Hour
	// DefaultBranchKeyTableName is the default DynamoDB table for caching branch keys
	DefaultBranchKeyTableName = "IdentityBrokerEncryptionBranchKeys"
)

// AWSAdapter implements the EncryptionPort interface using AWS Encryption SDK.
// Focuses solely on runtime encryption/decryption operations.
// Supports envelope encryption with AWS KMS hierarchical keyring (production) and environment variable KEK injection (development).
// The hierarchical keyring uses DynamoDB for caching branch keys, reducing KMS API calls and improving performance.
//
// Branch key provisioning is handled separately by AWSBranchKeyManager to maintain clean separation of concerns.
type AWSAdapter struct {
	encryptionClient *client.Client    // AWS Encryption SDK client for encrypt/decrypt operations
	keyring          mpltypes.IKeyring // Keyring (AWS KMS hierarchical, KMS, or Raw AES)
}

// NewAWSEncryption creates an AWS Encryption SDK adapter with automatic fan-out to supported scenarios:
// Scenario A (development): Base64-encoded AES-256 key (raw material)
// Scenario B (production): Hierarchical keyring with AWS KMS and DynamoDB branch key caching (KMS ARN)
//
// keyMaterial can be either:
//   - AWS KMS ARN: "arn:aws:kms:region:account:key/key-id" or "arn:aws:kms:region:account:alias/alias-name"
//   - Base64-encoded 32-byte AES-256 key: "aBcDeFgHiJkLmNoPqRsTuVwXyZ1234567890AB=="
//
// Environment variable interpolation is handled by the config loader before this function is called.
// Configuration example with environment variable:
//   - YAML: encryption.key: "${ENCRYPTION_KEK}"
//   - Config loader expands ${ENCRYPTION_KEK} → reads ENCRYPTION_KEK environment variable
//   - This function receives: the actual base64 key value (not the ${...} reference)
//
// Parameters:
//   - keyMaterial: AWS KMS ARN or base64-encoded 32-byte AES-256 key
//   - dynamoDBTableName: DynamoDB table for branch key caching (uses default if empty, ignored for base64 scenario)
//   - branchKeyTTL: TTL for cached branch keys (uses default if zero, ignored for base64 scenario)
//
// Returns:
//   - adapter: EncryptionPort implementation for Encrypt/Decrypt operations
//   - manager: BranchKeyManager implementation for provisioning/managing branch keys (nil for base64 scenario)
//   - error: If initialization fails
func NewAWSEncryption(keyMaterial, dynamoDBTableName string, branchKeyTTL time.Duration) (*AWSAdapter, *AWSBranchKeyManager, error) {
	if keyMaterial == "" {
		return nil, nil, encryption.NewKEKUnavailableError("key encryption key material is required", nil)
	}

	// Scenario B: AWS KMS ARN for production (hierarchical keyring with DynamoDB caching)
	if strings.HasPrefix(keyMaterial, "arn:aws:kms:") {
		awsConfig := &ports.AWSKMSConfig{
			KeyARN: keyMaterial,
		}
		adapter, keyStore, err := newAdapterWithKMSARNAndKeyStore(keyMaterial, dynamoDBTableName, branchKeyTTL, awsConfig)
		if err != nil {
			return nil, nil, err
		}
		manager := NewAWSBranchKeyManager(keyStore)
		return adapter, manager, nil
	}

	// Scenario A: Base64-encoded AES-256 key for development (raw AES keyring)
	// Note: Environment variable interpolation (${VAR_NAME}) is handled by the config loader
	// This function receives the actual base64-encoded key value
	adapter, err := newAdapterWithBase64KEK(keyMaterial)
	if err != nil {
		return nil, nil, err
	}
	// No branch key manager for raw AES keyring
	return adapter, nil, nil
}

// newAdapterWithKMSARNAndKeyStore creates an adapter using AWS KMS hierarchical keyring and returns the KeyStore.
// The hierarchical keyring uses DynamoDB for caching branch keys, reducing KMS API calls.
// This variant returns both the adapter and the KeyStore for branch key manager creation.
// awsCfg provides AWS SDK configuration including region, credentials, and endpoints.
func newAdapterWithKMSARNAndKeyStore(kmsARN, dynamoDBTableName string, branchKeyTTL time.Duration, awsCfg *ports.AWSKMSConfig) (*AWSAdapter, *KeyStore, error) {
	ctx := context.Background()

	// Create KeyStore with configured or default values
	dynamoDBTable := dynamoDBTableName
	if dynamoDBTable == "" {
		dynamoDBTable = DefaultBranchKeyTableName
	}

	ttl := branchKeyTTL
	if ttl == 0 {
		ttl = DefaultBranchKeyTTL
	}

	keyStoreCfg := KeyStoreConfig{
		KMSKeyARN:         kmsARN,
		DynamoDBTableName: dynamoDBTable,
		BranchKeyTTL:      ttl,
	}

	keyStore, err := createKeyStore(ctx, keyStoreCfg, awsCfg, "IdentityBrokerEncryptionVault")
	if err != nil {
		return nil, nil, err
	}

	// Create branch key supplier
	supplier := &BranchKeyIdSupplier{}

	// Create hierarchical keyring using KeyStore and supplier
	keyring, err := createHierarchicalKeyring(ctx, keyStore, supplier)
	if err != nil {
		return nil, nil, err
	}

	// Create Encryption SDK client with commitment policy for key commitment
	policy := mpltypes.ESDKCommitmentPolicyRequireEncryptRequireDecrypt
	encryptionClient, err := client.NewClient(esdktypes.AwsEncryptionSdkConfig{
		CommitmentPolicy: &policy,
	})
	if err != nil {
		return nil, nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create encryption SDK client: %v", err),
			err,
		)
	}

	adapter := &AWSAdapter{
		encryptionClient: encryptionClient,
		keyring:          keyring,
	}

	return adapter, keyStore, nil
}

// newAdapterWithEnvVarKEK creates an adapter using environment variable for key material.
// newAdapterWithBase64KEK creates an adapter with a base64-encoded AES-256 key.
// The keyMaterial should be a base64-encoded 32-byte key (already interpolated from environment variables by config loader).
func newAdapterWithBase64KEK(keyMaterial string) (*AWSAdapter, error) {
	if keyMaterial == "" {
		return nil, encryption.NewKEKUnavailableError(
			"key_encryption_key is empty; must be AWS KMS ARN or base64-encoded 32-byte key",
			nil,
		)
	}

	// Decode base64-encoded KEK
	kekBytes, err := base64.StdEncoding.DecodeString(keyMaterial)
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to decode key_encryption_key: key must be base64-encoded 32-byte AES-256 key, got: %v", err),
			err,
		)
	}

	// Validate KEK length (must be 32 bytes for AES-256)
	if len(kekBytes) != 32 {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("KEK must be 32 bytes (256-bit), got %d bytes", len(kekBytes)),
			nil,
		)
	}

	// Create Material Providers client
	matProvider, err := mpl.NewClient(mpltypes.MaterialProvidersConfig{})
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create Material Providers client: %v", err),
			err,
		)
	}

	// Create Raw AES keyring for local key material
	keyringInput := mpltypes.CreateRawAesKeyringInput{
		KeyNamespace: "my-application",
		KeyName:      "encryption-key",
		WrappingKey:  kekBytes,
		WrappingAlg:  mpltypes.AesWrappingAlgAlgAes256GcmIv12Tag16,
	}
	keyring, err := matProvider.CreateRawAesKeyring(context.Background(), keyringInput)
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create raw AES keyring: %v", err),
			err,
		)
	}

	// Create Encryption SDK client with commitment policy for key commitment
	policy := mpltypes.ESDKCommitmentPolicyRequireEncryptRequireDecrypt
	encryptionClient, err := client.NewClient(esdktypes.AwsEncryptionSdkConfig{
		CommitmentPolicy: &policy,
	})
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create encryption SDK client: %v", err),
			err,
		)
	}

	return &AWSAdapter{
		encryptionClient: encryptionClient,
		keyring:          keyring,
	}, nil
}

// Encrypt encrypts plaintext using envelope encryption with DEK + KEK wrapping.
// Uses AWS Encryption SDK with AESGCMSIV authenticated encryption.
// Context binding is enforced at both DEK and KEK layers.
func (a *AWSAdapter) Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error) {
	if a == nil || a.encryptionClient == nil {
		slog.Error("encryption_failed",
			"operation", "encrypt",
			"error_kind", encryption.ErrorKindKEKUnavailable,
			"reason", "adapter_not_initialized",
		)
		return nil, encryption.NewKEKUnavailableError("encryption adapter not properly initialized", nil)
	}

	// Extract service_id from context for logging (sanitized)
	serviceID := encryptionContext["service_id"]

	// Encrypt using AWS Encryption SDK
	// The SDK handles:
	// - Fresh DEK generation per call
	// - DEK encryption with plaintext using AESGCMSIV
	// - DEK wrapping with KEK using encryptionContext for AAD
	// - Serialization of envelope (wrapped DEK + ciphertext + auth tag)
	encryptInput := esdktypes.EncryptInput{
		Plaintext:         plaintext,
		EncryptionContext: encryptionContext,
		Keyring:           a.keyring,
	}

	result, err := a.encryptionClient.Encrypt(ctx, encryptInput)

	if err != nil {
		// Map AWS SDK errors to domain error types
		if strings.Contains(err.Error(), "context") {
			slog.Error("encryption_failed",
				"operation", "encrypt",
				"service_id", serviceID,
				"error_kind", encryption.ErrorKindContextMismatch,
			)
			return nil, encryption.NewContextMismatchError(
				fmt.Sprintf("context verification failed during encryption: %v", err),
				err,
			)
		}
		if strings.Contains(err.Error(), "integrity") || strings.Contains(err.Error(), "authentication") {
			slog.Error("encryption_failed",
				"operation", "encrypt",
				"service_id", serviceID,
				"error_kind", encryption.ErrorKindIntegrityViolation,
			)
			return nil, encryption.NewIntegrityViolationError(
				fmt.Sprintf("integrity verification failed during encryption: %v", err),
				err,
			)
		}

		slog.Error("encryption_failed",
			"operation", "encrypt",
			"service_id", serviceID,
			"error_kind", encryption.ErrorKindEncryptionFailed,
		)
		return nil, encryption.NewEncryptionFailedError(
			fmt.Sprintf("encryption failed: %v", err),
			err,
		)
	}

	if result == nil {
		slog.Error("encryption_failed",
			"operation", "encrypt",
			"service_id", serviceID,
			"error_kind", encryption.ErrorKindEncryptionFailed,
			"reason", "nil_result",
		)
		return nil, encryption.NewEncryptionFailedError("encryption returned nil result", nil)
	}

	// Ensure ciphertext is not empty
	if len(result.Ciphertext) == 0 {
		slog.Error("encryption_failed",
			"operation", "encrypt",
			"service_id", serviceID,
			"error_kind", encryption.ErrorKindEncryptionFailed,
			"reason", "empty_ciphertext",
		)
		return nil, encryption.NewEncryptionFailedError("encryption produced empty ciphertext", nil)
	}

	return result.Ciphertext, nil
}

// Decrypt decrypts ciphertext using envelope encryption with context verification.
// The AWS Encryption SDK automatically verifies:
// - Context binding at DEK decryption layer
// - Authentication tag verification
// - DEK unwrapping with KEK using same context
func (a *AWSAdapter) Decrypt(ctx context.Context, ciphertext []byte, encryptionContext map[string]string) ([]byte, error) {
	if a == nil || a.encryptionClient == nil {
		slog.Error("decryption_failed",
			"operation", "decrypt",
			"error_kind", encryption.ErrorKindKEKUnavailable,
			"reason", "adapter_not_initialized",
		)
		return nil, encryption.NewKEKUnavailableError("encryption adapter not properly initialized", nil)
	}

	// Extract service_id from context for logging (sanitized)
	serviceID := encryptionContext["service_id"]

	// Validate ciphertext is not empty
	if len(ciphertext) == 0 {
		slog.Error("decryption_failed",
			"operation", "decrypt",
			"service_id", serviceID,
			"error_kind", encryption.ErrorKindDecryptionFailed,
			"reason", "empty_ciphertext",
		)
		return nil, encryption.NewDecryptionFailedError("ciphertext cannot be empty", nil)
	}

	// Decrypt using AWS Encryption SDK
	// The SDK handles:
	// - Envelope parsing (extract wrapped DEK, ciphertext, auth tag)
	// - DEK unwrapping with KEK (verifying encryptionContext as AAD)
	// - Ciphertext decryption with DEK (verifying auth tag)
	// - Context verification at both layers (fail-closed on mismatch)
	decryptInput := esdktypes.DecryptInput{
		Ciphertext:        ciphertext,
		EncryptionContext: encryptionContext,
		Keyring:           a.keyring,
	}

	result, err := a.encryptionClient.Decrypt(ctx, decryptInput)
	if err != nil {
		// Map AWS SDK errors to domain error types
		if strings.Contains(err.Error(), "context") {
			slog.Error("decryption_failed",
				"operation", "decrypt",
				"service_id", serviceID,
				"error_kind", encryption.ErrorKindContextMismatch,
			)
			return nil, encryption.NewContextMismatchError(
				fmt.Sprintf("context verification failed during decryption: %v", err),
				err,
			)
		}
		if strings.Contains(err.Error(), "integrity") || strings.Contains(err.Error(), "authentication") {
			slog.Error("decryption_failed",
				"operation", "decrypt",
				"service_id", serviceID,
				"error_kind", encryption.ErrorKindIntegrityViolation,
			)
			return nil, encryption.NewIntegrityViolationError(
				fmt.Sprintf("integrity verification failed during decryption: %v", err),
				err,
			)
		}

		slog.Error("decryption_failed",
			"operation", "decrypt",
			"service_id", serviceID,
			"error_kind", encryption.ErrorKindDecryptionFailed,
		)
		return nil, encryption.NewDecryptionFailedError(
			fmt.Sprintf("decryption failed: %v", err),
			err,
		)
	}

	if result == nil {
		slog.Error("decryption_failed",
			"operation", "decrypt",
			"service_id", serviceID,
			"error_kind", encryption.ErrorKindDecryptionFailed,
			"reason", "nil_result",
		)
		return nil, encryption.NewDecryptionFailedError("decryption returned nil result", nil)
	}

	// Ensure plaintext is not empty
	if len(result.Plaintext) == 0 {
		slog.Error("decryption_failed",
			"operation", "decrypt",
			"service_id", serviceID,
			"error_kind", encryption.ErrorKindDecryptionFailed,
			"reason", "empty_plaintext",
		)
		return nil, encryption.NewDecryptionFailedError("decryption produced empty plaintext", nil)
	}

	return result.Plaintext, nil
}

// verifyKMSKeyAccessible checks that the KMS key is accessible.
func verifyKMSKeyAccessible(ctx context.Context, client *kms.Client, keyArn string) error {
	_, err := client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: &keyArn,
	})
	if err != nil {
		return encryption.NewKEKUnavailableError(
			fmt.Sprintf("KMS key not accessible: %v", err),
			err,
		)
	}
	return nil
}

// NewEncryptionAdapter creates an encryption adapter based on the configuration backend.
// This factory function implements the backend-explicit configuration design.
// Returns adapter, branch key manager, and error if configuration is invalid or adapter creation fails.
func NewEncryptionAdapter(config *ports.EncryptionConfig) (*AWSAdapter, ports.BranchKeyManager, error) {
	if config == nil {
		return nil, nil, encryption.NewKEKUnavailableError("encryption configuration is required", nil)
	}

	// Count configured backends to ensure exactly one is set
	backendCount := 0
	if config.AWSKMS != nil {
		backendCount++
	}
	if config.Memory != nil {
		backendCount++
	}

	// Validate exactly one backend is configured
	if backendCount == 0 {
		return nil, nil, encryption.NewKEKUnavailableError(
			"exactly one encryption backend must be configured (aws_kms or memory)",
			nil,
		)
	}
	if backendCount > 1 {
		return nil, nil, encryption.NewKEKUnavailableError(
			"exactly one encryption backend must be configured, not both aws_kms and memory",
			nil,
		)
	}

	// Create adapter based on configured backend
	if config.AWSKMS != nil {
		return createAWSKMSAdapter(config.AWSKMS)
	}

	if config.Memory != nil {
		return createMemoryAdapter(config.Memory)
	}

	// Should never reach here due to validation above
	return nil, nil, encryption.NewKEKUnavailableError("no valid backend configuration found", nil)
}

// createAWSKMSAdapter creates an adapter for AWS KMS backend
func createAWSKMSAdapter(config *ports.AWSKMSConfig) (*AWSAdapter, ports.BranchKeyManager, error) {
	// Apply defaults for optional fields
	dynamoDBTableName := config.DynamoDBTableName
	if dynamoDBTableName == "" {
		dynamoDBTableName = "IdentityBrokerEncryptionBranchKeys"
	}

	// Parse optional duration fields with defaults
	var branchKeyTTL time.Duration
	branchKeyTTLStr := config.BranchKeyTTL
	if branchKeyTTLStr == "" {
		branchKeyTTLStr = "1h" // Default
	}

	if branchKeyTTLStr != "" {
		ttl, err := time.ParseDuration(branchKeyTTLStr)
		if err != nil {
			return nil, nil, encryption.NewKEKUnavailableError(
				fmt.Sprintf("invalid branch_key_ttl duration: %v", err),
				err,
			)
		}
		branchKeyTTL = ttl
	}

	// Create adapter using new configuration-aware function
	adapter, keyStore, err := newAdapterWithKMSARNAndKeyStore(
		config.KeyARN,
		dynamoDBTableName,
		branchKeyTTL,
		config, // Pass full AWS configuration
	)
	if err != nil {
		return nil, nil, err
	}

	// Create branch key manager from the KeyStore
	var manager ports.BranchKeyManager
	if keyStore != nil {
		manager = NewAWSBranchKeyManager(keyStore)
	}

	return adapter, manager, nil
}

// createMemoryAdapter creates an adapter for Memory backend
func createMemoryAdapter(config *ports.MemoryConfig) (*AWSAdapter, ports.BranchKeyManager, error) {
	// Create adapter using existing function with base64 key
	adapter, manager, err := NewAWSEncryption(config.RawKey, "", 0)
	if err != nil {
		return nil, nil, err
	}

	// Handle interface nil pointer issue: if manager is a nil pointer, return nil interface
	if manager == nil {
		return adapter, nil, nil
	}

	return adapter, manager, nil
}
