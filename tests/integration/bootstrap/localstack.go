package bootstrap

import (
	"context"
	"fmt"
	"os"
	"testing"

	keystore "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygenerated"
	keystoretypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygeneratedtypes"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	awsencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/aws"
)

// LocalStackContainer manages LocalStack KMS + DynamoDB for E2E encryption tests
type LocalStackContainer struct {
	Container testcontainers.Container
	Endpoint  string
	KMSKeyID  string
	// Environment variables set for AWS SDK client configuration
	OriginalEndpoint string // Original AWS_ENDPOINT_URL_KMS value
}

// StartLocalStack creates and starts a LocalStack container with KMS + DynamoDB
// Usage: In BeforeEach, `ls := bootstrap.StartLocalStack(ctx, GinkgoT())`
// Usage: In AfterEach, `defer ls.Terminate(ctx)`
func StartLocalStack(ctx context.Context, t *testing.T) *LocalStackContainer {
	// Use default bridge network to avoid network creation issues with testcontainers reaper
	// The bridge network is always available and compatible with all Docker configurations
	req := testcontainers.ContainerRequest{
		Image: "localstack/localstack:4.12.0",
		Env: map[string]string{
			"SERVICES":              "kms,dynamodb",
			"AWS_ACCESS_KEY_ID":     "test",
			"AWS_SECRET_ACCESS_KEY": "test",
			"AWS_DEFAULT_REGION":    "eu-central-1",
			// Disable Ryuk reaper to avoid Docker network issues in constrained environments
			"TESTCONTAINERS_RYUK_DISABLED": "true",
		},
		ExposedPorts: []string{"4566/tcp"},
		WaitingFor:   wait.ForLog("Ready."),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start LocalStack container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "4566")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get mapped port: %v", err)
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())

	// Create KMS key
	kmsClient := kms.NewFromConfig(awsConfigForLocalStack(ctx, endpoint))
	keyOutput, err := kmsClient.CreateKey(ctx, &kms.CreateKeyInput{
		Description: wrap("Test encryption key"),
	})
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to create KMS key: %v", err)
	}

	keyID := *keyOutput.KeyMetadata.KeyId

	// Create DynamoDB table for branch key cache
	// Per AWS Encryption SDK KeyStore requirements:
	// Partition key: "branch-key-id" (S), Sort key: "type" (S)
	dynamoClient := dynamodb.NewFromConfig(awsConfigForLocalStack(ctx, endpoint))
	_, err = dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: wrap("IdentityBrokerEncryptionBranchKeys"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: wrap("branch-key-id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: wrap("type"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: wrap("branch-key-id"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: wrap("type"),
				KeyType:       types.KeyTypeRange,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
		// TTL can be configured separately via UpdateTimeToLive if needed
	})
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to create DynamoDB table: %v", err)
	}

	// Pre-populate branch keys in DynamoDB for test services
	// The hierarchical keyring requires these to exist
	err = preBranchKeysForLocalStack(ctx, kmsClient, dynamoClient, keyID)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to pre-populate branch keys: %v", err)
	}

	// Set environment variables so AWS SDK client code uses LocalStack endpoint
	os.Setenv("AWS_ENDPOINT_URL", endpoint)
	os.Setenv("AWS_ENDPOINT_URL_KMS", endpoint)
	os.Setenv("AWS_ENDPOINT_URL_DYNAMODB", endpoint)
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	os.Setenv("AWS_DEFAULT_REGION", "eu-central-1")

	return &LocalStackContainer{
		Container:        container,
		Endpoint:         endpoint,
		KMSKeyID:         keyID,
		OriginalEndpoint: os.Getenv("AWS_ENDPOINT_URL"),
	}
}

// SetupLocalStackEnvironment configures the AWS SDK to use LocalStack endpoint for the test
// Call this before creating encryption adapters, then call Cleanup() in defer
func (ls *LocalStackContainer) SetupLocalStackEnvironment() {
	os.Setenv("AWS_ENDPOINT_URL", ls.Endpoint)
	os.Setenv("AWS_ENDPOINT_URL_KMS", ls.Endpoint)
	os.Setenv("AWS_ENDPOINT_URL_DYNAMODB", ls.Endpoint)
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	os.Setenv("AWS_DEFAULT_REGION", "eu-central-1")
}

// CleanupLocalStackEnvironment restores original environment variables
func (ls *LocalStackContainer) CleanupLocalStackEnvironment() {
	if ls.OriginalEndpoint != "" {
		os.Setenv("AWS_ENDPOINT_URL", ls.OriginalEndpoint)
	} else {
		os.Unsetenv("AWS_ENDPOINT_URL")
	}
	os.Unsetenv("AWS_ENDPOINT_URL_KMS")
	os.Unsetenv("AWS_ENDPOINT_URL_DYNAMODB")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_DEFAULT_REGION")
}

// Terminate stops and removes the LocalStack container
func (ls *LocalStackContainer) Terminate(ctx context.Context) error {
	if ls.Container == nil {
		return nil
	}
	return ls.Container.Terminate(ctx)
}

// preBranchKeysForLocalStack creates branch keys using the KeyStore for test services.
// Branch keys must exist in DynamoDB before the hierarchical keyring can use them.
// This function uses the AWS Encryption SDK KeyStore to create proper branch key records.
func preBranchKeysForLocalStack(ctx context.Context, kmsClient *kms.Client, dynamoClient *dynamodb.Client, keyID string) error {
	testServices := []string{"oauth2", "github", "google"}

	// Create KeyStore client
	kmsARN := fmt.Sprintf("arn:aws:kms:eu-central-1:000000000000:key/%s", keyID)
	kmsConfig := keystoretypes.KMSConfigurationMemberkmsKeyArn{
		Value: kmsARN,
	}

	keystoreClient, err := keystore.NewClient(keystoretypes.KeyStoreConfig{
		DdbTableName:        "IdentityBrokerEncryptionBranchKeys",
		KmsConfiguration:    &kmsConfig,
		LogicalKeyStoreName: "IdentityBrokerEncryptionVault",
		DdbClient:           dynamoClient,
		KmsClient:           kmsClient,
	})
	if err != nil {
		return fmt.Errorf("failed to create KeyStore client: %w", err)
	}

	// Create branch keys for each test service using centralized ID generation
	for _, service := range testServices {
		branchKeyID := awsencryption.GetBranchKeyID(service)
		encryptionCtx := map[string]string{
			"service_id": service,
		}
		_, err := keystoreClient.CreateKey(ctx, keystoretypes.CreateKeyInput{
			BranchKeyIdentifier: &branchKeyID,
			EncryptionContext:   encryptionCtx,
		})
		if err != nil {
			return fmt.Errorf("failed to create branch key for service %s: %w", service, err)
		}
	}

	return nil
}

// awsConfigForLocalStack creates AWS SDK config pointing to LocalStack endpoint
func awsConfigForLocalStack(ctx context.Context, endpoint string) aws.Config {
	cfg, _ := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-central-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint}, nil
			})),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	return cfg
}

func wrap(s string) *string { return &s }
