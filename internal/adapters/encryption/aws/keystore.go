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

// CreateBranchKey creates a branch key in DynamoDB for the provided subject.
// This operation is idempotent: if a branch key with the given subject already exists it is
// returned without creating a new one. This allows callers (including service Update paths
// where a service was originally created with a different encryption backend) to call
// CreateBranchKey unconditionally without failing on pre-existing keys.
func (ks *KeyStore) CreateBranchKey(ctx context.Context, subject encryption.BranchKeySubject) (string, error) {
	if ks == nil || ks.client == nil {
		return "", encryption.NewKEKUnavailableError("KeyStore not initialized", nil)
	}

	if err := subject.Validate(); err != nil {
		return "", encryption.NewKEKUnavailableError(fmt.Sprintf("invalid branch key subject: %v", err), err)
	}

	branchKeyID := branchkey.GenerateBranchKeyId(subject)
	if branchKeyID == "" {
		return "", encryption.NewKEKUnavailableError("failed to generate branch key ID from subject", nil)
	}

	// Idempotency check: if the branch key already exists, return it without creating a duplicate.
	// This handles the migration case where a service was created with a different encryption
	// backend (e.g. raw AES) and then updated after switching to KMS — the branch key for
	// that service ID has not yet been provisioned in DynamoDB, but for services already using
	// KMS the branch key exists and CreateKey would fail.
	_, getErr := ks.client.GetActiveBranchKey(ctx, keystoretypes.GetActiveBranchKeyInput{
		BranchKeyIdentifier: branchKeyID,
	})
	if getErr == nil {
		return branchKeyID, nil
	}
	// getErr is non-nil: either the key does not exist yet (expected migration path) or
	// DynamoDB/KMS is temporarily unavailable. We cannot reliably distinguish the two cases
	// from a KeyStoreException, so we attempt CreateKey regardless. If the infrastructure
	// is unavailable, CreateKey will also fail and return a clear error to the caller.

	branchKey, err := ks.client.CreateKey(ctx, keystoretypes.CreateKeyInput{
		BranchKeyIdentifier: &branchKeyID,
		EncryptionContext:   subject.EncryptionContext(),
	})
	if err != nil {
		return "", encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to create branch key %s in DynamoDB: %v", branchKeyID, err),
			err,
		)
	}
	return branchKey.BranchKeyIdentifier, nil
}
