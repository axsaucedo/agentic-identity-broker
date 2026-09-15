package bootstrap

import (
	"context"
	"errors"
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
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/branchkey"
	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

const localStackCompatibleHealthEndpoint = "/_localstack/health"

// AWSEmulatorContainer manages a LocalStack-compatible AWS emulator for encryption integration tests.
type AWSEmulatorContainer struct {
	Container testcontainers.Container
	Endpoint  string
	KMSKeyID  string
	// Environment variables set for AWS SDK client configuration.
	OriginalEndpoint    string // Original AWS_ENDPOINT_URL value
	originalEnvironment map[string]environmentVariable
}

type environmentVariable struct {
	value   string
	present bool
}

// StartAWSEmulatorForSuite starts a LocalStack-compatible AWS emulator intended to be shared across a test suite.
// Returns an error instead of calling t.Fatal so that TestMain can handle startup failures gracefully.
func StartAWSEmulatorForSuite(ctx context.Context) (*AWSEmulatorContainer, error) {
	return startEmulator(ctx)
}

// StartAWSEmulator creates and starts a LocalStack-compatible AWS emulator with KMS + DynamoDB.
func StartAWSEmulator(ctx context.Context, t *testing.T) *AWSEmulatorContainer {
	t.Helper()
	container, err := startEmulator(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return container
}

func startEmulator(ctx context.Context) (*AWSEmulatorContainer, error) {
	// Use default bridge network to avoid network creation issues with testcontainers reaper.
	// The bridge network is always available and compatible with all Docker configurations.
	req := testcontainers.ContainerRequest{
		Image: "floci/floci:1.5.23",
		Env: map[string]string{
			"FLOCI_DEFAULT_REGION": "eu-central-1",
			"FLOCI_STORAGE_MODE":   "memory",
		},
		ExposedPorts: []string{"4566/tcp"},
		WaitingFor:   wait.ForHTTP(localStackCompatibleHealthEndpoint).WithPort("4566/tcp"),
	}

	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		if err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"); err != nil {
			return nil, fmt.Errorf("failed to disable testcontainers Ryuk: %w", err)
		}
	}

	// testcontainers panics (via sync.Once) when no Docker host is reachable.
	// Narrow recovery to this call only to avoid masking panics from later steps.
	var startErr error
	container, err := func() (testcontainers.Container, error) {
		defer func() {
			if r := recover(); r != nil {
				startErr = fmt.Errorf("docker runtime unavailable — ensure Colima or Docker Desktop is running: %v", r)
			}
		}()
		return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	}()
	if startErr != nil {
		return nil, startErr
	}
	if err != nil {
		return nil, fmt.Errorf("failed to start AWS emulator container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, cleanupStartFailure(ctx, container, "failed to get container host", err)
	}

	port, err := container.MappedPort(ctx, "4566")
	if err != nil {
		return nil, cleanupStartFailure(ctx, container, "failed to get mapped port", err)
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())

	kmsConfig, err := awsConfigForEmulator(ctx, endpoint)
	if err != nil {
		return nil, cleanupStartFailure(ctx, container, "failed to load AWS config for KMS", err)
	}
	kmsClient := kms.NewFromConfig(kmsConfig)
	keyOutput, err := kmsClient.CreateKey(ctx, &kms.CreateKeyInput{
		Description: wrap("Test encryption key"),
	})
	if err != nil {
		return nil, cleanupStartFailure(ctx, container, "failed to create KMS key", err)
	}

	keyID := *keyOutput.KeyMetadata.KeyId

	// Create DynamoDB table for branch key cache.
	// Per AWS Encryption SDK KeyStore requirements (matches CDK stack schema):
	// Partition key: "branch-key-id" (S), Sort key: "type" (S).
	dynamoConfig, err := awsConfigForEmulator(ctx, endpoint)
	if err != nil {
		return nil, cleanupStartFailure(ctx, container, "failed to load AWS config for DynamoDB", err)
	}
	dynamoClient := dynamodb.NewFromConfig(dynamoConfig)
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
	})
	if err != nil {
		return nil, cleanupStartFailure(ctx, container, "failed to create DynamoDB table", err)
	}

	err = preBranchKeysForEmulator(ctx, kmsClient, dynamoClient, keyID)
	if err != nil {
		return nil, cleanupStartFailure(ctx, container, "failed to pre-populate branch keys", err)
	}

	emulator := &AWSEmulatorContainer{
		Container: container,
		Endpoint:  endpoint,
		KMSKeyID:  keyID,
	}
	if err := emulator.SetupEnvironment(); err != nil {
		return nil, cleanupStartFailure(ctx, container, "failed to configure AWS emulator environment", err)
	}
	return emulator, nil
}

func cleanupStartFailure(ctx context.Context, container testcontainers.Container, operation string, err error) error {
	operationErr := fmt.Errorf("%s: %w", operation, err)
	if terminateErr := container.Terminate(ctx); terminateErr != nil {
		return errors.Join(operationErr, fmt.Errorf("terminate AWS emulator container: %w", terminateErr))
	}
	return operationErr
}

// SetupEnvironment configures the AWS SDK to use the emulator endpoint for the test.
func (c *AWSEmulatorContainer) SetupEnvironment() error {
	c.captureOriginalEnvironment()

	if err := os.Setenv("AWS_ENDPOINT_URL", c.Endpoint); err != nil {
		return fmt.Errorf("set AWS_ENDPOINT_URL: %w", err)
	}
	if err := os.Setenv("AWS_ENDPOINT_URL_KMS", c.Endpoint); err != nil {
		return fmt.Errorf("set AWS_ENDPOINT_URL_KMS: %w", err)
	}
	if err := os.Setenv("AWS_ENDPOINT_URL_DYNAMODB", c.Endpoint); err != nil {
		return fmt.Errorf("set AWS_ENDPOINT_URL_DYNAMODB: %w", err)
	}
	if err := os.Setenv("AWS_ACCESS_KEY_ID", "test"); err != nil {
		return fmt.Errorf("set AWS_ACCESS_KEY_ID: %w", err)
	}
	if err := os.Setenv("AWS_SECRET_ACCESS_KEY", "test"); err != nil {
		return fmt.Errorf("set AWS_SECRET_ACCESS_KEY: %w", err)
	}
	if err := os.Setenv("AWS_DEFAULT_REGION", "eu-central-1"); err != nil {
		return fmt.Errorf("set AWS_DEFAULT_REGION: %w", err)
	}

	return nil
}

// CleanupEnvironment restores original environment variables.
func (c *AWSEmulatorContainer) CleanupEnvironment() error {
	c.captureOriginalEnvironment()

	var cleanupErr error
	for _, key := range []string{
		"AWS_ENDPOINT_URL",
		"AWS_ENDPOINT_URL_KMS",
		"AWS_ENDPOINT_URL_DYNAMODB",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_DEFAULT_REGION",
	} {
		if err := c.restoreEnvironmentVariable(key); err != nil {
			cleanupErr = errors.Join(cleanupErr, err)
		}
	}

	return cleanupErr
}

func (c *AWSEmulatorContainer) captureOriginalEnvironment() {
	if c.originalEnvironment != nil {
		return
	}

	c.originalEnvironment = map[string]environmentVariable{}
	for _, key := range []string{
		"AWS_ENDPOINT_URL",
		"AWS_ENDPOINT_URL_KMS",
		"AWS_ENDPOINT_URL_DYNAMODB",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_DEFAULT_REGION",
	} {
		value, present := os.LookupEnv(key)
		c.originalEnvironment[key] = environmentVariable{value: value, present: present}
	}
	c.OriginalEndpoint = c.originalEnvironment["AWS_ENDPOINT_URL"].value
}

func (c *AWSEmulatorContainer) restoreEnvironmentVariable(key string) error {
	original := c.originalEnvironment[key]
	if original.present {
		if err := os.Setenv(key, original.value); err != nil {
			return fmt.Errorf("restore %s: %w", key, err)
		}
		return nil
	}
	if err := os.Unsetenv(key); err != nil {
		return fmt.Errorf("unset %s: %w", key, err)
	}
	return nil
}

// Terminate stops and removes the emulator container.
func (c *AWSEmulatorContainer) Terminate(ctx context.Context) error {
	if c.Container == nil {
		return nil
	}
	if err := c.Container.Terminate(ctx); err != nil {
		return err
	}
	c.Container = nil
	return nil
}

// preBranchKeysForEmulator creates branch keys using the KeyStore for test services.
func preBranchKeysForEmulator(ctx context.Context, kmsClient *kms.Client, dynamoClient *dynamodb.Client, keyID string) error {
	testServiceIDs := []id.ServiceID{
		id.MustParseServiceID("01234567-89ab-cdef-0123-456789abcdef"),
		id.MustParseServiceID("12345678-9abc-def0-1234-56789abcdef0"),
		id.MustParseServiceID("23456789-abcd-ef01-2345-6789abcdef01"),
	}

	kmsARN := fmt.Sprintf("arn:aws:kms:eu-central-1:000000000000:key/%s", keyID)
	kmsConfig := keystoretypes.KMSConfigurationMemberkmsKeyArn{Value: kmsARN}

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

	branchKeyIDProvider := awsencryption.NewBranchKeyIdSupplier(branchkey.NewDefaultProvider())
	for _, serviceID := range testServiceIDs {
		serviceSubject := domainencryption.NewServiceBranchKeySubject(serviceID)
		branchKeyID, err := branchKeyIDProvider.GenerateBranchKeyId(serviceSubject)
		if err != nil {
			return fmt.Errorf("failed to derive branch key ID for service %s: %w", serviceID, err)
		}
		encryptionCtx := serviceSubject.EncryptionContext()
		_, err = keystoreClient.CreateKey(ctx, keystoretypes.CreateKeyInput{
			BranchKeyIdentifier: &branchKeyID,
			EncryptionContext:   encryptionCtx,
		})
		if err != nil {
			return fmt.Errorf("failed to create branch key for service %s: %w", serviceID, err)
		}
	}

	return nil
}

// awsConfigForEmulator creates AWS SDK config pointing to the emulator endpoint.
func awsConfigForEmulator(ctx context.Context, endpoint string) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-central-1"),
		config.WithBaseEndpoint(endpoint),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
}

func wrap(s string) *string { return &s }

// NewKMSClientForEmulator creates a KMS client configured for the emulator endpoint.
func NewKMSClientForEmulator(ctx context.Context, endpoint string) (*kms.Client, error) {
	cfg, err := awsConfigForEmulator(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	return kms.NewFromConfig(cfg), nil
}
