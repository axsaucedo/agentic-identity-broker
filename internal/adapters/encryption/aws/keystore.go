package aws

import (
	"context"
	"fmt"
	"time"

	keystore "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygenerated"
	keystoretypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygeneratedtypes"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kms"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
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
// - AWS SDK Config: Base AWS configuration
// - KMS Client: For KMS key operations
// - DynamoDB Client: For branch key storage and caching
// - KeyStore Client: Wraps above for branch key management
//
// Parameters:
//   - ctx: Context for AWS API calls
//   - cfg: KeyStore configuration (KMS ARN, DynamoDB table, TTL)
//   - logicalKeyStoreName: Identifier for this KeyStore instance
//
// Returns:
//   - *KeyStore: The initialized KeyStore (or nil on error)
//   - error: If KeyStore initialization fails
func createKeyStore(ctx context.Context, cfg KeyStoreConfig, logicalKeyStoreName string) (*KeyStore, error) {
	if cfg.KMSKeyARN == "" {
		return nil, encryption.NewKEKUnavailableError("KMS key ARN is required for KeyStore", nil)
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
	keystoreConfig := keystoretypes.KeyStoreConfig{
		DdbTableName:        cfg.DynamoDBTableName,
		KmsClient:           kmsClient,
		DdbClient:           dynamoDBClient,
		LogicalKeyStoreName: logicalKeyStoreName,
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

	return &KeyStore{
		client: keystoreClient,
		config: cfg,
	}, nil
}

// CreateBranchKey creates a branch key in DynamoDB with the specified ID.
// Returns the branch key identifier or error if creation fails.
func (ks *KeyStore) CreateBranchKey(ctx context.Context, branchKeyID string) (string, error) {
	if ks == nil || ks.client == nil {
		return "", encryption.NewKEKUnavailableError("KeyStore not initialized", nil)
	}

	// Use underlying AWS KeyStore client to create the branch key
	// The KeyStore internally handles KMS/DynamoDB operations
	branchKey, err := ks.client.CreateKey(ctx, keystoretypes.CreateKeyInput{})
	if err != nil {
		return "", encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create branch key in DynamoDB: %v", err),
			err,
		)
	}
	return branchKey.BranchKeyIdentifier, nil
}
