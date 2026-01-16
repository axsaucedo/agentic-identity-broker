package aws

import (
	"context"
	"fmt"
	"time"

	mpl "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygenerated"
	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
	keystore "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygenerated"
	keystoretypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygeneratedtypes"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kms"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
)

// HierarchicalKeyringConfig contains configuration for hierarchical keyring setup
type HierarchicalKeyringConfig struct {
	KMSKeyARN        string        // AWS KMS key ARN for root key
	DynamoDBTableName string        // DynamoDB table for caching branch keys
	BranchKeyTTL     time.Duration // TTL for cached branch keys
}

// createHierarchicalKeyring creates an AWS KMS hierarchical keyring with DynamoDB caching.
// The hierarchical keyring uses a branch key supplier that caches branch keys in DynamoDB,
// reducing the number of KMS API calls and improving performance.
//
// Architecture (per AWS Encryption SDK example):
// 1. AWS SDK Client (for signing/validation) + AWS Config
// 2. KMS Client (for KMS API calls)
// 3. DynamoDB Client (for branch key caching)
// 4. KeyStore Client (wraps KMS + DynamoDB for branch key management)
// 5. BranchKeySupplier (manages GenerateDataKey with DynamoDB caching)
// 6. AWS KMS Hierarchical Keyring (uses KeyStore + BranchKeySupplier)
// 7. AwsEncryptionSdk Client (uses keyring for Encrypt/Decrypt)
func createHierarchicalKeyring(ctx context.Context, cfg HierarchicalKeyringConfig) (mpltypes.IKeyring, error) {
	if cfg.KMSKeyARN == "" {
		return nil, encryption.NewKEKUnavailableError("KMS key ARN is required for hierarchical keyring", nil)
	}

	if cfg.DynamoDBTableName == "" {
		cfg.DynamoDBTableName = DefaultBranchKeyTableName
	}

	if cfg.BranchKeyTTL == 0 {
		cfg.BranchKeyTTL = DefaultBranchKeyTTL
	}

	// Load AWS SDK configuration
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to load AWS config: %v", err),
			err,
		)
	}

	// Create KMS client
	kmsClient := kms.NewFromConfig(awsCfg)

	// Create DynamoDB client
	dynamoDBClient := dynamodb.NewFromConfig(awsCfg)

	// Verify KMS key is accessible
	if err := verifyKMSKeyAccessible(ctx, kmsClient, cfg.KMSKeyARN); err != nil {
		return nil, err
	}

	// Create KeyStore client (wraps KMS + DynamoDB for branch key management)
	// KeyStore handles the DynamoDB cache and KMS key wrapping operations
	keystoreConfig := keystoretypes.KeyStoreConfig{
		DdbTableName: cfg.DynamoDBTableName,
		KmsClient:    kmsClient,
		DdbClient:    dynamoDBClient,
		KmsConfiguration: &keystoretypes.KMSConfigurationMemberkmsKeyArn{
			Value: cfg.KMSKeyARN,
		},
	}

	keystoreClient, err := keystore.NewClient(keystoreConfig)
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create KeyStore client: %v", err),
			err,
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

	// Create hierarchical keyring using the KeyStore
	// The hierarchical keyring will cache branch keys in DynamoDB via the KeyStore
	// This reduces KMS API calls significantly for high-throughput scenarios
	hierarchicalKeyringInput := mpltypes.CreateAwsKmsHierarchicalKeyringInput{
		KeyStore:   keystoreClient,
		TtlSeconds: int64(cfg.BranchKeyTTL.Seconds()),
	}

	keyring, err := matProvider.CreateAwsKmsHierarchicalKeyring(ctx, hierarchicalKeyringInput)
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create AWS KMS hierarchical keyring: %v", err),
			err,
		)
	}

	return keyring, nil
}
