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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
)

const (
	// DefaultKEKValidationTimeout is the timeout for KMS key accessibility checks
	DefaultKEKValidationTimeout = 30 * time.Second
)

// AWSAdapter implements the EncryptionPort interface using AWS Encryption SDK.
// Supports envelope encryption with AWS KMS (production) and environment variable KEK injection (development).
// Provides memory protection via memguard for sensitive data.
type AWSAdapter struct {
	encryptionClient *client.Client                                                    // AWS Encryption SDK client for encrypt/decrypt operations
	matProvider      *mpl.Client                                                       // Material Providers client for keyring operations
	keyring          mpltypes.IKeyring                                                 // Keyring (AWS KMS or Raw AES)
	kekType          string                                                            // "kms" or "env_var"
	kmsKeyID         string                                                            // KMS key ARN/alias (for kms mode only)
	kekMaterial      []byte                                                            // Memory-protected KEK (for env_var mode only)
}

// NewAWSEncryptionAdapter creates a new AWS Encryption SDK adapter with envelope encryption.
// keyMaterial can be either:
//   - AWS KMS ARN: "arn:aws:kms:region:account:key/key-id" or "arn:aws:kms:region:account:alias/alias-name"
//   - Environment variable reference: "${ENCRYPTION_KEK}" (resolves to base64-encoded key)
//
// The adapter validates KEK accessibility at startup (fail-fast).
func NewAWSEncryptionAdapter(keyMaterial string) (*AWSAdapter, error) {
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
		return newAdapterWithKMSARN(keyMaterial)
	}

	// Invalid format
	return nil, encryption.NewKEKUnavailableError(
		fmt.Sprintf("invalid key material format: must be AWS KMS ARN or ${ENV_VAR}, got: %s", keyMaterial),
		nil,
	)
}

// newAdapterWithKMSARN creates an adapter using AWS KMS for key management.
func newAdapterWithKMSARN(kmsARN string) (*AWSAdapter, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to load AWS config: %v", err),
			err,
		)
	}

	// Create KMS client
	kmsClient := kms.NewFromConfig(cfg)

	// Verify KMS key is accessible
	ctx, cancel := context.WithTimeout(context.Background(), DefaultKEKValidationTimeout)
	defer cancel()

	if err := verifyKMSKeyAccessible(ctx, kmsClient, kmsARN); err != nil {
		return nil, err
	}

	// Create Material Providers client
	matProvider, err := mpl.NewClient(mpltypes.MaterialProvidersConfig{})
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create Material Providers client: %v", err),
			err,
		)
	}

	// Create AWS KMS keyring
	keyringInput := mpltypes.CreateAwsKmsKeyringInput{
		KmsClient: kmsClient,
		KmsKeyId:  kmsARN,
	}
	keyring, err := matProvider.CreateAwsKmsKeyring(context.Background(), keyringInput)
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create AWS KMS keyring: %v", err),
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
		matProvider:      matProvider,
		keyring:          keyring,
		kekType:          "kms",
		kmsKeyID:         kmsARN,
	}, nil
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
		matProvider:      matProvider,
		keyring:          keyring,
		kekType:          "env_var",
		kekMaterial:      kekBytes,
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

// GetKEKType returns the type of Key Encryption Key in use ("kms" or "env_var").
// Useful for diagnostics and logging.
func (a *AWSAdapter) GetKEKType() string {
	if a == nil {
		return ""
	}
	return a.kekType
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
