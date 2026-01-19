package bootstrap

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// LocalStackContainer manages LocalStack KMS + DynamoDB for E2E encryption tests
type LocalStackContainer struct {
	Container testcontainers.Container
	Endpoint  string
	KMSKeyID  string
}

// StartLocalStack creates and starts a LocalStack container with KMS + DynamoDB
// Usage: In BeforeEach, `ls := bootstrap.StartLocalStack(ctx, GinkgoT())`
// Usage: In AfterEach, `defer ls.Terminate(ctx)`
func StartLocalStack(ctx context.Context, t *testing.T) *LocalStackContainer {
	req := testcontainers.ContainerRequest{
		Image: "localstack/localstack:latest",
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
	dynamoClient := dynamodb.NewFromConfig(awsConfigForLocalStack(ctx, endpoint))
	_, err = dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: wrap("IdentityBrokerEncryptionBranchKeys"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: wrap("branchKeyId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: wrap("sort_key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: wrap("branchKeyId"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: wrap("sort_key"),
				KeyType:       types.KeyTypeRange,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
		TimeToLiveSpecification: &types.TimeToLiveSpecification{
			Enabled:       wrap(true),
			AttributeName: wrap("ttl"),
		},
	})
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to create DynamoDB table: %v", err)
	}

	return &LocalStackContainer{
		Container: container,
		Endpoint:  endpoint,
		KMSKeyID:  keyID,
	}
}

// Terminate stops and removes the LocalStack container
func (ls *LocalStackContainer) Terminate(ctx context.Context) error {
	return ls.Container.Terminate(ctx)
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
func wrapBool(b bool) *bool  { return &b }
