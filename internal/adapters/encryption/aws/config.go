package aws

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// buildAWSConfig constructs an aws.Config from AWSKMSConfig.
// Supports configurable AWS SDK settings for various deployment scenarios.
//
// Configuration scenarios supported:
//   - Default: Uses AWS SDK default credential chain and region resolution
//   - Region override: Explicitly sets AWS region for all services
//   - Custom endpoints: For LocalStack or custom AWS implementations
//   - Static credentials: Access key ID and secret for testing/CI environments
//   - AWS profile: Uses named profile from ~/.aws/credentials
//   - IAM role assumption: Assumes role using default credentials then switches
//   - SSL disable: For LocalStack only (development only, DANGEROUS in production)
//
// Parameters:
//   - ctx: Context for AWS API calls and credential resolution
//   - cfg: AWSKMSConfig with AWS SDK configuration
//
// Returns:
//   - aws.Config: Configured AWS SDK config ready for service client creation
//   - error: If configuration is invalid (bad URLs, missing credentials, etc.)
func buildAWSConfig(ctx context.Context, cfg *ports.AWSKMSConfig) (aws.Config, error) {
	if cfg == nil {
		return aws.Config{}, encryption.NewKEKUnavailableError("AWSKMSConfig cannot be nil", nil)
	}

	// Validate security settings to prevent insecure configurations
	if err := validateSecuritySettings(cfg); err != nil {
		return aws.Config{}, err
	}

	// Build configuration options slice
	var options []func(*config.LoadOptions) error

	// Region configuration
	if cfg.Region != "" {
		options = append(options, config.WithRegion(cfg.Region))
	}

	// Credentials configuration (priority: static > profile > default chain)
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		// Static credentials for testing/CI
		staticCredentials := credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"", // No session token
		)
		options = append(options, config.WithCredentialsProvider(staticCredentials))
	} else if cfg.Profile != "" {
		// Named AWS profile
		options = append(options, config.WithSharedConfigProfile(cfg.Profile))
	}

	// Custom HTTP client for SSL disable (LocalStack testing)
	if cfg.DisableSSL {
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		options = append(options, config.WithHTTPClient(httpClient))
	}

	// Note: Custom endpoint resolution is applied at the service client level
	// (in createKeyStore) using modern EndpointResolverV2 pattern to avoid deprecated APIs

	// Load base AWS configuration with all options
	awsConfig, err := config.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return aws.Config{}, encryption.NewKEKUnavailableError(
			fmt.Sprintf("failed to load AWS configuration: %v", err),
			err,
		)
	}

	// IAM role assumption (must happen after base config is loaded)
	if cfg.AssumeRoleARN != "" {
		if err := validateRoleARN(cfg.AssumeRoleARN); err != nil {
			return aws.Config{}, err
		}

		// Create STS client for role assumption
		stsClient := sts.NewFromConfig(awsConfig)

		// Create assume role credential provider
		assumeRoleProvider := stscreds.NewAssumeRoleProvider(stsClient, cfg.AssumeRoleARN)

		// Update config with assume role credentials
		awsConfig.Credentials = assumeRoleProvider
	}

	return awsConfig, nil
}

// GetKMSEndpoint returns the configured KMS endpoint URL, or empty string for default.
// This should be passed to the KMS client at creation time using modern endpoint resolution.
func GetKMSEndpoint(cfg *ports.AWSKMSConfig) (string, error) {
	if cfg == nil || cfg.KMSEndpoint == "" {
		return "", nil
	}
	if err := validateEndpointURL(cfg.KMSEndpoint, "kms_endpoint"); err != nil {
		return "", err
	}
	return cfg.KMSEndpoint, nil
}

// GetDynamoDBEndpoint returns the configured DynamoDB endpoint URL, or empty string for default.
// This should be passed to the DynamoDB client at creation time using modern endpoint resolution.
func GetDynamoDBEndpoint(cfg *ports.AWSKMSConfig) (string, error) {
	if cfg == nil || cfg.DynamoDBEndpoint == "" {
		return "", nil
	}
	if err := validateEndpointURL(cfg.DynamoDBEndpoint, "dynamodb_endpoint"); err != nil {
		return "", err
	}
	return cfg.DynamoDBEndpoint, nil
}

// validateEndpointURL validates that a custom endpoint URL is properly formatted.
func validateEndpointURL(endpoint, fieldName string) error {
	if endpoint == "" {
		return nil // Empty endpoint is valid (uses default AWS endpoints)
	}

	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return encryption.NewKEKUnavailableError(
			fmt.Sprintf("invalid %s URL format: %s", fieldName, endpoint),
			err,
		)
	}

	// Check for valid schemes (http or https)
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return encryption.NewKEKUnavailableError(
			fmt.Sprintf("%s must include scheme (http:// or https://): %s", fieldName, endpoint),
			nil,
		)
	}

	if parsedURL.Host == "" {
		return encryption.NewKEKUnavailableError(
			fmt.Sprintf("%s must include host: %s", fieldName, endpoint),
			nil,
		)
	}

	return nil
}

// validateRoleARN validates that an IAM role ARN is properly formatted.
func validateRoleARN(roleARN string) error {
	if roleARN == "" {
		return nil // Empty is valid (no role assumption)
	}

	// Basic ARN format validation
	// Expected format: arn:aws:iam::account-id:role/role-name
	if len(roleARN) < 20 { // Minimum realistic ARN length
		return encryption.NewKEKUnavailableError(
			fmt.Sprintf("assume_role_arn too short to be valid ARN: %s", roleARN),
			nil,
		)
	}

	if len(roleARN) < 7 || roleARN[:7] != "arn:aws" {
		return encryption.NewKEKUnavailableError(
			fmt.Sprintf("assume_role_arn must start with 'arn:aws': %s", roleARN),
			nil,
		)
	}

	// More detailed validation could be added here if needed
	// For now, let AWS SDK validate the ARN during actual assumption

	return nil
}

// validateSecuritySettings performs runtime security validation to prevent dangerous configurations.
// This prevents insecure settings that could compromise encryption in production environments.
func validateSecuritySettings(cfg *ports.AWSKMSConfig) error {
	if cfg == nil {
		return nil
	}

	// SSL VERIFICATION MUST BE ENABLED IN PRODUCTION
	if cfg.DisableSSL {
		// Check for production indicators
		isProduction := detectProductionEnvironment()

		if isProduction {
			return encryption.NewKEKUnavailableError(
				"CRITICAL SECURITY ERROR: disable_ssl is set to true but production environment detected. "+
					"SSL verification MUST be enabled in production to prevent man-in-the-middle attacks. "+
					"disable_ssl should ONLY be used for LocalStack testing. "+
					"To enable production mode, unset disable_ssl or set it to false.",
				nil,
			)
		}

		// Even in non-production, log a strong warning
		fmt.Fprintf(os.Stderr, "WARNING: SSL verification is DISABLED (disable_ssl: true). "+
			"This is DANGEROUS in production and should only be used for LocalStack testing.\n")
	}

	// CUSTOM ENDPOINTS REQUIRE CAUTION
	if cfg.KMSEndpoint != "" || cfg.DynamoDBEndpoint != "" {
		if detectProductionEnvironment() && (cfg.KMSEndpoint == "" || !isSecureEndpoint(cfg.KMSEndpoint)) {
			return encryption.NewKEKUnavailableError(
				"Production environment detected with incomplete endpoint configuration. "+
					"When using custom endpoints in production, all endpoints must use HTTPS and be explicitly configured.",
				nil,
			)
		}
	}

	// STATIC CREDENTIALS ARE DANGEROUS IN PRODUCTION
	if cfg.AccessKeyID != "" || cfg.SecretAccessKey != "" {
		if detectProductionEnvironment() {
			return encryption.NewKEKUnavailableError(
				"CRITICAL SECURITY ERROR: Static credentials detected in production environment. "+
					"Production deployments MUST use IAM roles (assume_role_arn) or AWS profiles instead of static credentials. "+
					"Static credentials should ONLY be used for testing/development. "+
					"For production, configure assume_role_arn with appropriate IAM permissions.",
				nil,
			)
		}
	}

	return nil
}

// detectProductionEnvironment attempts to detect if the application is running in a production environment
// based on multiple indicators. Uses conservative defaults - when in doubt, assumes production.
func detectProductionEnvironment() bool {
	// Check common production environment indicators
	env := os.Getenv("ENVIRONMENT")
	if env != "" {
		env = strings.ToLower(strings.TrimSpace(env))
		if env == "production" || env == "prod" || env == "prd" {
			return true
		}
	}

	// Check if running with explicit production indicators
	nodeEnv := os.Getenv("NODE_ENV")
	if nodeEnv != "" {
		nodeEnv = strings.ToLower(strings.TrimSpace(nodeEnv))
		if nodeEnv == "production" || nodeEnv == "prod" {
			return true
		}
	}

	// Check for Identity Broker specific environment variable
	brokerEnv := os.Getenv("IDENTITY_BROKER_ENV")
	if brokerEnv != "" {
		brokerEnv = strings.ToLower(strings.TrimSpace(brokerEnv))
		if brokerEnv == "production" || brokerEnv == "prod" || brokerEnv == "prd" {
			return true
		}
	}

	// Check for Kubernetes/container production indicators
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		// Running in Kubernetes - likely production unless explicitly opted out
		if os.Getenv("ALLOW_INSECURE_CONFIG") != "true" {
			return true
		}
	}

	// Check for container environment (Docker, systemd-nspawn, etc.)
	// If running in container without explicit development marker, assume production
	if isRunningInContainer() {
		if os.Getenv("DEVELOPMENT_MODE") != "true" && os.Getenv("ALLOW_INSECURE_CONFIG") != "true" {
			return true
		}
	}

	// Detect test environment indicators
	if isTestEnvironment() {
		return false
	}

	// Default to non-production (development) for safety
	// Only production if explicitly marked or detected
	return false
}

// isRunningInContainer detects if the application is running inside a container
func isRunningInContainer() bool {
	// Docker container detection
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check for cgroup-based container detection (works for Docker, Podman, containerd, etc.)
	cgroupContent, err := os.ReadFile("/proc/self/cgroup")
	if err == nil {
		cgroupStr := string(cgroupContent)
		if strings.Contains(cgroupStr, "/docker") ||
			strings.Contains(cgroupStr, "/lxc") ||
			strings.Contains(cgroupStr, "/podman") ||
			strings.Contains(cgroupStr, "kubelet") {
			return true
		}
	}

	return false
}

// isTestEnvironment detects if running under test
func isTestEnvironment() bool {
	// Go test sets testing.Testing() or GOTEST environment variable
	if os.Getenv("GOTEST") != "" {
		return true
	}

	// Check for explicit test mode markers
	if os.Getenv("DEVELOPMENT_MODE") == "true" {
		return true
	}

	if os.Getenv("ALLOW_INSECURE_CONFIG") == "true" {
		return true
	}

	return false
}

// isSecureEndpoint checks if an endpoint URL uses HTTPS
func isSecureEndpoint(endpoint string) bool {
	return strings.HasPrefix(endpoint, "https://")
}
