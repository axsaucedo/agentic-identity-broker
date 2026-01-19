package bootstrap

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	localStackTestNetwork = "localstack-test"
)

// LocalStackContainer manages LocalStack KMS + DynamoDB for E2E encryption tests
type LocalStackContainer struct {
	Container testcontainers.Container
	Network   testcontainers.Network
	Endpoint  string
	KMSKeyID  string
	// Environment variables set for AWS SDK client configuration
	OriginalEndpoint string // Original AWS_ENDPOINT_URL_KMS value
}

// StartLocalStack creates and starts a LocalStack container with KMS + DynamoDB
// Usage: In BeforeEach, `ls := bootstrap.StartLocalStack(ctx, GinkgoT())`
// Usage: In AfterEach, `defer ls.Terminate(ctx)`
func StartLocalStack(ctx context.Context, t *testing.T) *LocalStackContainer {
	// Create a named network for LocalStack to support both Docker and Podman
	// This avoids issues with the default bridge network in some environments
	net, err := network.New(ctx, network.WithLabels(map[string]string{"name": localStackTestNetwork}))
	if err != nil {
		t.Fatalf("failed to create network for LocalStack: %v", err)
	}

	req := testcontainers.ContainerRequest{
		Image: "localstack/localstack:latest",
		Networks: []string{net.Name},
		Env: map[string]string{
			"SERVICES":              "kms,dynamodb",
			"AWS_ACCESS_KEY_ID":     "test",
			"AWS_SECRET_ACCESS_KEY": "test",
			"AWS_DEFAULT_REGION":    "eu-central-1",
		},
		ExposedPorts: []string{"4566/tcp"},
		WaitingFor:   wait.ForLog("Ready."),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		net.Remove(ctx)
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
	// Partition key: "partition_key" (S), Sort key: "sort_key" (S)
	dynamoClient := dynamodb.NewFromConfig(awsConfigForLocalStack(ctx, endpoint))
	_, err = dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: wrap("IdentityBrokerEncryptionBranchKeys"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: wrap("partition_key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: wrap("sort_key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: wrap("partition_key"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: wrap("sort_key"),
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

	// Set environment variables so AWS SDK client code uses LocalStack endpoint
	os.Setenv("AWS_ENDPOINT_URL", endpoint)
	os.Setenv("AWS_ENDPOINT_URL_KMS", endpoint)
	os.Setenv("AWS_ENDPOINT_URL_DYNAMODB", endpoint)
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	os.Setenv("AWS_DEFAULT_REGION", "eu-central-1")

	return &LocalStackContainer{
		Container:        container,
		Network:          net,
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

// Terminate stops and removes the LocalStack container and its network
func (ls *LocalStackContainer) Terminate(ctx context.Context) error {
	// First terminate the container
	if err := ls.Container.Terminate(ctx); err != nil {
		return err
	}
	// Then remove the network
	if ls.Network != nil {
		return ls.Network.Remove(ctx)
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
