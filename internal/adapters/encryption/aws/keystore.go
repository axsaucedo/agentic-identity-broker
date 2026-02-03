package aws

import (
	"context"
	"fmt"
	"time"

	keystore "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygenerated"
	keystoretypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygeneratedtypes"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kms"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/branchkey"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// KeyStoreConfig contains configuration for AWS KMS KeyStore setup.
// The KeyStore manages branch key creation and caching in DynamoDB.
type KeyStoreConfig struct {
	KMSKeyARN         string        // AWS KMS key ARN for root key
	DynamoDBTableName string        // DynamoDB table for caching branch keys
	BranchKeyTTL      time.Duration // TTL for cached branch keys
}

// KeyStore wraps the AWS KeyStore client and manages its lifecycle.
// The KeyStore is responsible for creating and managing branch keys in DynamoDB,
// which are then used by the hierarchical keyring for envelope encryption.
type KeyStore struct {
	client *keystore.Client
	config KeyStoreConfig
}

// createKeyStore creates an AWS KMS KeyStore client with DynamoDB caching.
// The KeyStore handles branch key creation and caching, reducing KMS API calls.
//
// Architecture:
// - AWS SDK Config: Configurable AWS configuration (region, credentials, endpoints)
// - KMS Client: For KMS key operations
// - DynamoDB Client: For branch key storage and caching
// - KeyStore Client: Wraps above for branch key management
//
// Parameters:
//   - ctx: Context for AWS API calls
//   - ksCfg: KeyStore configuration (KMS ARN, DynamoDB table, TTL)
//   - awsCfg: AWS SDK configuration (region, credentials, endpoints, etc.)
//   - logicalKeyStoreName: Identifier for this KeyStore instance
//
// Returns:
//   - *KeyStore: The initialized KeyStore (or nil on error)
//   - error: If KeyStore initialization fails
func createKeyStore(ctx context.Context, ksCfg KeyStoreConfig, awsCfg *ports.AWSKMSConfig, logicalKeyStoreName string) (*KeyStore, error) {
	if ksCfg.KMSKeyARN == "" {
		return nil, encryption.NewKEKUnavailableError("KMS key ARN is required for KeyStore", nil)
	}

	if ksCfg.DynamoDBTableName == "" {
		ksCfg.DynamoDBTableName = DefaultBranchKeyTableName
	}

	if ksCfg.BranchKeyTTL == 0 {
		ksCfg.BranchKeyTTL = DefaultBranchKeyTTL
	}

	// Build AWS SDK configuration from provided settings
	sdkCfg, err := buildAWSConfig(ctx, awsCfg)
	if err != nil {
		return nil, err
	}

	// Create KMS client with configured AWS settings
	// Apply custom endpoint if configured (modern EndpointResolverV2 pattern)
	kmsEndpoint, err := GetKMSEndpoint(awsCfg)
	if err != nil {
		return nil, err
	}
	var kmsOpts []func(*kms.Options)
	if kmsEndpoint != "" {
		kmsOpts = append(kmsOpts, func(o *kms.Options) {
			o.BaseEndpoint = &kmsEndpoint
		})
	}
	kmsClient := kms.NewFromConfig(sdkCfg, kmsOpts...)

	// Create DynamoDB client with configured AWS settings
	// Override region if DynamoDBRegion is specifically configured
	dynamoDBConfig := sdkCfg
	if awsCfg != nil && awsCfg.DynamoDBRegion != "" && awsCfg.DynamoDBRegion != sdkCfg.Region {
		dynamoDBConfig = sdkCfg.Copy()
		dynamoDBConfig.Region = awsCfg.DynamoDBRegion
	}
	// Apply custom endpoint if configured (modern EndpointResolverV2 pattern)
	dynamoDBEndpoint, err := GetDynamoDBEndpoint(awsCfg)
	if err != nil {
		return nil, err
	}
	var dynamoDBOpts []func(*dynamodb.Options)
	if dynamoDBEndpoint != "" {
		dynamoDBOpts = append(dynamoDBOpts, func(o *dynamodb.Options) {
			o.BaseEndpoint = &dynamoDBEndpoint
		})
	}
	dynamoDBClient := dynamodb.NewFromConfig(dynamoDBConfig, dynamoDBOpts...)

	// Verify KMS key is accessible
	if err := verifyKMSKeyAccessible(ctx, kmsClient, ksCfg.KMSKeyARN); err != nil {
		return nil, err
	}

	// Create KeyStore client (wraps KMS + DynamoDB for branch key management)
	keystoreConfig := keystoretypes.KeyStoreConfig{
		DdbTableName:        ksCfg.DynamoDBTableName,
		KmsClient:           kmsClient,
		DdbClient:           dynamoDBClient,
		LogicalKeyStoreName: logicalKeyStoreName,
		KmsConfiguration: &keystoretypes.KMSConfigurationMemberkmsKeyArn{
			Value: ksCfg.KMSKeyARN,
		},
	}

	keystoreClient, err := keystore.NewClient(keystoreConfig)
	if err != nil {
		return nil, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create KeyStore client: %v", err),
			err,
		)
	}

	return &KeyStore{
		client: keystoreClient,
		config: ksCfg,
	}, nil
}

// CreateBranchKey creates a branch key in DynamoDB with the specified ID and encryption context.
// Returns the branch key identifier or error if creation fails.
// Note: AWS Encryption SDK KeyStore requires encryption context when using custom branch key identifiers.
func (ks *KeyStore) CreateBranchKey(ctx context.Context, branchKeyID string) (string, error) {
	if ks == nil || ks.client == nil {
		return "", encryption.NewKEKUnavailableError("KeyStore not initialized", nil)
	}

	if branchKeyID == "" {
		return "", encryption.NewKEKUnavailableError("branchKeyID cannot be empty", nil)
	}

	// AWS Encryption SDK KeyStore requires encryption context when using custom branch key identifiers
	// Extract service_id from branch key ID using the centralized parser
	serviceID, err := branchkey.ExtractServiceID(branchKeyID)
	if err != nil {
		return "", encryption.NewKEKUnavailableError(
			fmt.Sprintf("invalid branch key ID format: %s", branchKeyID),
			err,
		)
	}

	encryptionCtx := map[string]string{"service_id": serviceID}

	// Use underlying AWS KeyStore client to create the branch key with the specified ID
	// The BranchKeyIdentifier must be set to ensure deterministic ID matching with BranchKeyIdSupplier
	branchKey, err := ks.client.CreateKey(ctx, keystoretypes.CreateKeyInput{
		BranchKeyIdentifier: &branchKeyID,
		EncryptionContext:   encryptionCtx,
	})
	if err != nil {
		return "", encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create branch key %s in DynamoDB: %v", branchKeyID, err),
			err,
		)
	}
	return branchKey.BranchKeyIdentifier, nil
}
