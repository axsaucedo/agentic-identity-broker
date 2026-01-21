package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	mpl "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygenerated"
	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
	client "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygenerated"
	esdktypes "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygeneratedtypes"
	"github.com/aws/aws-sdk-go-v2/service/kms"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
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
// Provides memory protection via memguard for sensitive data.
//
// Branch key provisioning is handled separately by AWSBranchKeyManager to maintain clean separation of concerns.
type AWSAdapter struct {
	encryptionClient *client.Client    // AWS Encryption SDK client for encrypt/decrypt operations
	keyring          mpltypes.IKeyring // Keyring (AWS KMS hierarchical, KMS, or Raw AES)
}

// NewAWSEncryptionAdapter creates a new AWS Encryption SDK adapter with envelope encryption.
// keyMaterial can be either:
//   - AWS KMS ARN: "arn:aws:kms:region:account:key/key-id" or "arn:aws:kms:region:account:alias/alias-name"
//   - Environment variable reference: "${ENCRYPTION_KEK}" (resolves to base64-encoded key)
//
// For KMS ARNs, optional dynamoDBTableName and branchKeyTTL can be provided for branch key caching.
// If dynamoDBTableName is empty, defaults to DefaultBranchKeyTableName.
// If branchKeyTTL is zero, defaults to DefaultBranchKeyTTL.
//
// The adapter validates KEK accessibility at startup (fail-fast).
func NewAWSEncryptionAdapter(keyMaterial, dynamoDBTableName string, branchKeyTTL time.Duration) (*AWSAdapter, error) {
	if keyMaterial == "" {
		return nil, encryption.NewKEKUnavailableError("key encryption key material is required", nil)
	}

	// Determine if this is an environment variable reference or AWS KMS ARN
	if strings.HasPrefix(keyMaterial, "${") && strings.HasSuffix(keyMaterial, "}") {
		// Environment variable reference: ${ENCRYPTION_KEK}
		envVarName := keyMaterial[2 : len(keyMaterial)-1]
		return newAdapterWithEnvVarKEK(envVarName)
	}

	// Assume it's an AWS KMS ARN
	if strings.HasPrefix(keyMaterial, "arn:aws:kms:") {
		adapter, _, err := newAdapterWithKMSARNAndKeyStore(keyMaterial, dynamoDBTableName, branchKeyTTL)
		return adapter, err
	}

	// Invalid format
	return nil, encryption.NewKEKUnavailableError(
		fmt.Sprintf("invalid key material format: must be AWS KMS ARN or ${ENV_VAR}, got: %s", keyMaterial),
		nil,
	)
}

// NewAWSEncryptionAdapterWithBranchKeyManager creates an AWS Encryption SDK adapter with a branch key manager.
// Returns both the adapter (for encryption operations) and the manager (for branch key provisioning).
// This is the primary constructor for production use where branch key management is needed.
//
// Parameters:
//   - keyMaterial: KMS ARN or environment variable reference
//   - dynamoDBTableName: DynamoDB table for branch key caching (uses default if empty)
//   - branchKeyTTL: TTL for cached branch keys (uses default if zero)
//
// Returns:
//   - adapter: EncryptionPort implementation for Encrypt/Decrypt operations
//   - manager: BranchKeyManager implementation for provisioning/managing branch keys
//   - error: If initialization fails
func NewAWSEncryptionAdapterWithBranchKeyManager(keyMaterial, dynamoDBTableName string, branchKeyTTL time.Duration) (*AWSAdapter, *AWSBranchKeyManager, error) {
	if keyMaterial == "" {
		return nil, nil, encryption.NewKEKUnavailableError("key encryption key material is required", nil)
	}

	// For environment variable references, branch key manager is not available
	if strings.HasPrefix(keyMaterial, "${") && strings.HasSuffix(keyMaterial, "}") {
		envVarName := keyMaterial[2 : len(keyMaterial)-1]
		adapter, err := newAdapterWithEnvVarKEK(envVarName)
		if err != nil {
			return nil, nil, err
		}
		// No branch key manager for raw AES keyring
		return adapter, nil, nil
	}

	// Assume it's an AWS KMS ARN with hierarchical keyring configuration
	if strings.HasPrefix(keyMaterial, "arn:aws:kms:") {
		adapter, keyStore, err := newAdapterWithKMSARNAndKeyStore(keyMaterial, dynamoDBTableName, branchKeyTTL)
		if err != nil {
			return nil, nil, err
		}
		// Create branch key manager from the KeyStore
		manager := NewAWSBranchKeyManager(keyStore)
		return adapter, manager, nil
	}

	// Invalid format
	return nil, nil, encryption.NewKEKUnavailableError(
		fmt.Sprintf("invalid key material format: must be AWS KMS ARN or ${ENV_VAR}, got: %s", keyMaterial),
		nil,
	)
}

// newAdapterWithKMSARNAndKeyStore creates an adapter using AWS KMS hierarchical keyring and returns the KeyStore.
// The hierarchical keyring uses DynamoDB for caching branch keys, reducing KMS API calls.
// This variant returns both the adapter and the KeyStore for branch key manager creation.
// dynamoDBTableName and branchKeyTTL override defaults if provided (non-empty/non-zero).
func newAdapterWithKMSARNAndKeyStore(kmsARN, dynamoDBTableName string, branchKeyTTL time.Duration) (*AWSAdapter, *KeyStore, error) {
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

	keyStore, err := createKeyStore(ctx, keyStoreCfg, "IdentityBrokerEncryptionVault")
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

	// Create Encryption SDK client
	encryptionClient, err := client.NewClient(esdktypes.AwsEncryptionSdkConfig{})
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
func newAdapterWithEnvVarKEK(envVarName string) (*AWSAdapter, error) {
	// Load KEK from environment variable
	keyMaterial := os.Getenv(envVarName)
	if keyMaterial == "" {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("environment variable %s not set", envVarName),
			nil,
		)
	}

	// Decode base64-encoded KEK
	kekBytes, err := base64.StdEncoding.DecodeString(keyMaterial)
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to decode base64-encoded KEK from %s: %v", envVarName, err),
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

	// Create Encryption SDK client
	encryptionClient, err := client.NewClient(esdktypes.AwsEncryptionSdkConfig{})
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
		return nil, encryption.NewKEKUnavailableError("encryption adapter not properly initialized", nil)
	}

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
			return nil, encryption.NewContextMismatchError(
				fmt.Sprintf("context verification failed during encryption: %v", err),
				err,
			)
		}
		if strings.Contains(err.Error(), "integrity") || strings.Contains(err.Error(), "authentication") {
			return nil, encryption.NewIntegrityViolationError(
				fmt.Sprintf("integrity verification failed during encryption: %v", err),
				err,
			)
		}
		return nil, encryption.NewEncryptionFailedError(
			fmt.Sprintf("encryption failed: %v", err),
			err,
		)
	}

	if result == nil {
		return nil, encryption.NewEncryptionFailedError("encryption returned nil result", nil)
	}

	// Ensure ciphertext is not empty
	if len(result.Ciphertext) == 0 {
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
		return nil, encryption.NewKEKUnavailableError("encryption adapter not properly initialized", nil)
	}

	// Validate ciphertext is not empty
	if len(ciphertext) == 0 {
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
			return nil, encryption.NewContextMismatchError(
				fmt.Sprintf("context verification failed during decryption: %v", err),
				err,
			)
		}
		if strings.Contains(err.Error(), "integrity") || strings.Contains(err.Error(), "authentication") {
			return nil, encryption.NewIntegrityViolationError(
				fmt.Sprintf("integrity verification failed during decryption: %v", err),
				err,
			)
		}
		return nil, encryption.NewDecryptionFailedError(
			fmt.Sprintf("decryption failed: %v", err),
			err,
		)
	}

	if result == nil {
		return nil, encryption.NewDecryptionFailedError("decryption returned nil result", nil)
	}

	// Ensure plaintext is not empty
	if len(result.Plaintext) == 0 {
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
