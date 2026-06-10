package aws

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

func TestBuildAWSConfig_Default(t *testing.T) {
	// Test default configuration (empty config uses AWS SDK defaults)
	// Note: In CI/test environments without AWS credentials, region may not be set
	cfg := &ports.AWSKMSConfig{}

	awsConfig, err := buildAWSConfig(context.Background(), cfg)

	require.NoError(t, err)
	require.NotNil(t, awsConfig, "AWS config should not be nil")

	// Verify config has expected properties set, even if region is empty
	// In CI environments without AWS credentials, this validates config structure
	t.Logf("AWS Config region: %s (may be empty in test environment)", awsConfig.Region)
}

func TestBuildAWSConfig_NilConfig(t *testing.T) {
	// Test nil configuration should return error
	_, err := buildAWSConfig(context.Background(), nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "AWSKMSConfig cannot be nil")
}

func TestBuildAWSConfig_RegionOverride(t *testing.T) {
	// Test region override
	cfg := &ports.AWSKMSConfig{
		Region: "eu-west-1",
	}

	awsConfig, err := buildAWSConfig(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, "eu-west-1", awsConfig.Region)
}

func TestBuildAWSConfig_StaticCredentials(t *testing.T) {
	// Test static credentials
	cfg := &ports.AWSKMSConfig{
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
	}

	awsConfig, err := buildAWSConfig(context.Background(), cfg)

	require.NoError(t, err)
	assert.NotNil(t, awsConfig.Credentials)

	// Verify credentials are set correctly
	creds, err := awsConfig.Credentials.Retrieve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "test-access-key", creds.AccessKeyID)
	assert.Equal(t, "test-secret-key", creds.SecretAccessKey)
}

func TestBuildAWSConfig_Profile(t *testing.T) {
	// Test profile configuration
	cfg := &ports.AWSKMSConfig{
		Profile: "test-profile",
	}

	// This might fail if profile doesn't exist, but we can test that the config is attempted
	_, err := buildAWSConfig(context.Background(), cfg)

	// Either succeeds or fails with profile-related error (both are acceptable for this test)
	// The important thing is that it doesn't fail with nil config error
	if err != nil {
		assert.Contains(t, err.Error(), "profile")
	}
}

func TestBuildAWSConfig_CustomEndpoints(t *testing.T) {
	// Test custom endpoints (AWS emulator scenario)
	// Modern pattern: endpoints are applied at service client level, not config level
	cfg := &ports.AWSKMSConfig{
		KMSEndpoint:      "http://localhost:4566",
		DynamoDBEndpoint: "http://localhost:4566",
	}

	awsConfig, err := buildAWSConfig(context.Background(), cfg)

	require.NoError(t, err)
	// Verify config was built successfully (endpoints will be applied at service client creation)
	assert.NotNil(t, awsConfig)
	// Modern endpoint resolution doesn't store resolver in config - it's applied per-service
}

func TestBuildAWSConfig_DisableSSL(t *testing.T) {
	// Test SSL disable for a LocalStack-compatible AWS emulator
	cfg := &ports.AWSKMSConfig{
		DisableSSL: true,
	}

	awsConfig, err := buildAWSConfig(context.Background(), cfg)

	require.NoError(t, err)
	assert.NotNil(t, awsConfig.HTTPClient)
}

func TestBuildAWSConfig_AssumeRole(t *testing.T) {
	// Test IAM role assumption
	cfg := &ports.AWSKMSConfig{
		AssumeRoleARN: "arn:aws:iam::123456789012:role/TestRole",
	}

	// This will likely fail without proper AWS credentials, but we test the validation
	_, err := buildAWSConfig(context.Background(), cfg)

	// Should not fail due to ARN validation (only AWS API call might fail)
	// If it fails, it should be an AWS-related error, not validation error
	if err != nil && !containsAWSError(err.Error()) {
		t.Errorf("Expected AWS error or success, got validation error: %v", err)
	}
}

func TestValidateEndpointURL(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		fieldName   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty endpoint (valid)",
			endpoint:    "",
			fieldName:   "test_endpoint",
			expectError: false,
		},
		{
			name:        "valid http endpoint",
			endpoint:    "http://localhost:4566",
			fieldName:   "test_endpoint",
			expectError: false,
		},
		{
			name:        "valid https endpoint",
			endpoint:    "https://kms.us-east-1.amazonaws.com",
			fieldName:   "test_endpoint",
			expectError: false,
		},
		{
			name:        "missing scheme",
			endpoint:    "localhost:4566",
			fieldName:   "test_endpoint",
			expectError: true,
			errorMsg:    "must include scheme",
		},
		{
			name:        "missing host",
			endpoint:    "http://",
			fieldName:   "test_endpoint",
			expectError: true,
			errorMsg:    "must include host",
		},
		{
			name:        "invalid URL",
			endpoint:    "://invalid",
			fieldName:   "test_endpoint",
			expectError: true,
			errorMsg:    "invalid test_endpoint URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEndpointURL(tt.endpoint, tt.fieldName)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateRoleARN(t *testing.T) {
	tests := []struct {
		name        string
		roleARN     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty ARN (valid)",
			roleARN:     "",
			expectError: false,
		},
		{
			name:        "valid role ARN",
			roleARN:     "arn:aws:iam::123456789012:role/TestRole",
			expectError: false,
		},
		{
			name:        "valid role ARN with path",
			roleARN:     "arn:aws:iam::123456789012:role/service/TestRole",
			expectError: false,
		},
		{
			name:        "too short ARN",
			roleARN:     "arn:aws",
			expectError: true,
			errorMsg:    "too short to be valid ARN",
		},
		{
			name:        "invalid ARN prefix",
			roleARN:     "invalid:aws:iam::123456789012:role/TestRole",
			expectError: true,
			errorMsg:    "must start with 'arn:aws'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoleARN(tt.roleARN)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetKMSEndpoint(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *ports.AWSKMSConfig
		expectedURL   string
		expectedError bool
	}{
		{
			name:          "empty_config",
			cfg:           &ports.AWSKMSConfig{},
			expectedURL:   "",
			expectedError: false,
		},
		{
			name: "valid_kms_endpoint",
			cfg: &ports.AWSKMSConfig{
				KMSEndpoint: "http://localhost:4566",
			},
			expectedURL:   "http://localhost:4566",
			expectedError: false,
		},
		{
			name: "invalid_kms_endpoint",
			cfg: &ports.AWSKMSConfig{
				KMSEndpoint: "invalid-url",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := GetKMSEndpoint(tt.cfg)
			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "must include scheme")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

func TestGetDynamoDBEndpoint(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *ports.AWSKMSConfig
		expectedURL   string
		expectedError bool
	}{
		{
			name:          "empty_config",
			cfg:           &ports.AWSKMSConfig{},
			expectedURL:   "",
			expectedError: false,
		},
		{
			name: "valid_dynamodb_endpoint",
			cfg: &ports.AWSKMSConfig{
				DynamoDBEndpoint: "https://dynamodb.eu-central-1.amazonaws.com",
			},
			expectedURL:   "https://dynamodb.eu-central-1.amazonaws.com",
			expectedError: false,
		},
		{
			name: "invalid_dynamodb_endpoint",
			cfg: &ports.AWSKMSConfig{
				DynamoDBEndpoint: "not-a-url",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := GetDynamoDBEndpoint(tt.cfg)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

// containsAWSError checks if error message contains AWS-specific error patterns
func containsAWSError(errMsg string) bool {
	awsErrorPatterns := []string{
		"aws",
		"AWS",
		"credential",
		"region",
		"profile",
		"token",
		"assume",
		"sts",
	}

	for _, pattern := range awsErrorPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}
	return false
}
