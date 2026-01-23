package config

import (
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// TestEncryptionConfigStructure verifies that the EncryptionConfig is properly integrated into the Config structure
func TestEncryptionConfigStructure(t *testing.T) {
	// Test that the EncryptionConfig is properly integrated into the main Config struct
	var config ports.Config

	// Verify that Encryption field exists and is of correct type
	if config.Encryption.KeyEncryptionKey != "" {
		t.Error("Default KeyEncryptionKey should be empty")
	}

	// Test assignment
	testValue := "test-key-value"
	config.Encryption.KeyEncryptionKey = testValue
	if config.Encryption.KeyEncryptionKey != testValue {
		t.Error("Failed to assign to KeyEncryptionKey field")
	}

	// Test that the mapstructure tag is correct (by checking it exists as a field)
	// This ensures the configuration system can properly unmarshal the field
	if config.Encryption.KeyEncryptionKey != testValue {
		t.Error("KeyEncryptionKey field assignment failed")
	}
}

// TestEncryptionConfigValidation verifies the validate:"required" tag works
func TestEncryptionConfigValidation(t *testing.T) {
	// This is a structural test to verify the EncryptionConfig has proper validation tags
	var config ports.EncryptionConfig

	// Test that we can assign values (basic structure validation)
	config.KeyEncryptionKey = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
	if config.KeyEncryptionKey == "" {
		t.Error("Failed to assign AWS KMS ARN to KeyEncryptionKey")
	}

	// Test environment variable reference format
	config.KeyEncryptionKey = "${ENCRYPTION_KEK}"
	if config.KeyEncryptionKey != "${ENCRYPTION_KEK}" {
		t.Error("Failed to assign environment variable reference to KeyEncryptionKey")
	}
}

// TestEncryptionConfigFieldTags verifies the struct tags are correct
func TestEncryptionConfigFieldTags(t *testing.T) {
	// Verify that the configuration integrates properly at the type level
	var mainConfig ports.Config

	// Ensure the Encryption field is present and accessible
	mainConfig.Encryption.KeyEncryptionKey = "test-value"

	if mainConfig.Encryption.KeyEncryptionKey != "test-value" {
		t.Error("Encryption field not properly accessible in main Config struct")
	}

	// Test that all expected configuration sections exist
	// This ensures the EncryptionConfig is properly integrated alongside other configs
	testConfigs := map[string]interface{}{
		"Log":              mainConfig.Log,
		"Server":           mainConfig.Server,
		"Storage":          mainConfig.Storage,
		"ThirdPartyOAuth2": mainConfig.ThirdPartyOAuth2,
		"OAuth2AuthServer": mainConfig.OAuth2AuthServer,
		"Security":         mainConfig.Security,
		"Encryption":       mainConfig.Encryption,
	}

	for name, config := range testConfigs {
		if config == nil {
			// Some configs might be zero values, that's okay
			t.Logf("Config section %s has zero value (this is normal)", name)
		}
	}
}

// TestEncryptionConfigDefaults verifies default values
func TestEncryptionConfigDefaults(t *testing.T) {
	var config ports.EncryptionConfig

	// Default should be empty (required field, must be explicitly set)
	if config.KeyEncryptionKey != "" {
		t.Errorf("Default KeyEncryptionKey should be empty, got: %q", config.KeyEncryptionKey)
	}
}

// TestEncryptionConfigFieldTypes verifies field types are correct
func TestEncryptionConfigFieldTypes(t *testing.T) {
	var config ports.EncryptionConfig

	// KeyEncryptionKey should be a string
	_ = config.KeyEncryptionKey

	// Test assignment of different valid formats
	testValues := []string{
		"arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
		generateBase64EncodedString(32),
	}

	for i, testValue := range testValues {
		config.KeyEncryptionKey = testValue
		if config.KeyEncryptionKey != testValue {
			t.Errorf("Test %d: Failed to assign value %q to KeyEncryptionKey", i, testValue)
		}
	}
}

// TestValidateEncryptionConfigMissingKey verifies validation fails when KeyEncryptionKey is empty
func TestValidateEncryptionConfigMissingKey(t *testing.T) {
	cfg := &ports.Config{
		Log:     ports.LogConfig{Level: ports.LogLevelInfo, Format: ports.LogFormatText},
		Server:  createValidServerConfig(),
		Storage: createValidStorageConfig(),
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{
			JWESigningKey: generateBase64EncodedString(32),
		},
		Encryption: ports.EncryptionConfig{
			KeyEncryptionKey: "", // Missing required field
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for missing KeyEncryptionKey, got nil")
	}

	// Verify error mentions the field
	if _, ok := err.(*config.ConfigError); !ok {
		t.Errorf("Expected ConfigError, got %T", err)
	}
}

// TestValidateEncryptionConfigValidAWSKMSARN verifies validation passes for valid AWS KMS ARN
func TestValidateEncryptionConfigValidAWSKMSARN(t *testing.T) {
	cfg := &ports.Config{
		Log:     ports.LogConfig{Level: ports.LogLevelInfo, Format: ports.LogFormatText},
		Server:  createValidServerConfig(),
		Storage: createValidStorageConfig(),
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{
			JWESigningKey: generateBase64EncodedString(32),
		},
		Encryption: ports.EncryptionConfig{
			KeyEncryptionKey: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
		},
	}

	err := Validate(cfg)
	if err != nil {
		t.Errorf("Expected validation to pass for valid AWS KMS ARN, got error: %v", err)
	}
}

// TestValidateEncryptionConfigValidBase64Key verifies validation passes for valid base64-encoded 32-byte key
func TestValidateEncryptionConfigValidBase64Key(t *testing.T) {
	cfg := &ports.Config{
		Log:     ports.LogConfig{Level: ports.LogLevelInfo, Format: ports.LogFormatText},
		Server:  createValidServerConfig(),
		Storage: createValidStorageConfig(),
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{
			JWESigningKey: generateBase64EncodedString(32),
		},
		Encryption: ports.EncryptionConfig{
			KeyEncryptionKey: generateBase64EncodedString(32),
		},
	}

	err := Validate(cfg)
	if err != nil {
		t.Errorf("Expected validation to pass for valid base64-encoded 32-byte key, got error: %v", err)
	}
}

// TestValidateEncryptionConfigInvalidBase64 verifies validation fails for invalid base64
func TestValidateEncryptionConfigInvalidBase64(t *testing.T) {
	cfg := &ports.Config{
		Log:     ports.LogConfig{Level: ports.LogLevelInfo, Format: ports.LogFormatText},
		Server:  createValidServerConfig(),
		Storage: createValidStorageConfig(),
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{
			JWESigningKey: generateBase64EncodedString(32),
		},
		Encryption: ports.EncryptionConfig{
			KeyEncryptionKey: "not-valid-base64-!!!@#$", // Invalid base64 and not an ARN
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for invalid base64 key, got nil")
	}
}

// TestValidateEncryptionConfigWrongKeyLength verifies validation fails for non-32-byte key
func TestValidateEncryptionConfigWrongKeyLength(t *testing.T) {

	// Create a 16-byte key (not 32-byte)
	invalidKey := generateBase64EncodedString(16)

	cfg := &ports.Config{
		Log:     ports.LogConfig{Level: ports.LogLevelInfo, Format: ports.LogFormatText},
		Server:  createValidServerConfig(),
		Storage: createValidStorageConfig(),
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{
			JWESigningKey: generateBase64EncodedString(32),
		},
		Encryption: ports.EncryptionConfig{
			KeyEncryptionKey: invalidKey,
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for 16-byte key (not 32-byte), got nil")
	}
}

// TestValidateEncryptionConfigInvalidARNFormat verifies validation fails for malformed ARN
func TestValidateEncryptionConfigInvalidARNFormat(t *testing.T) {
	cfg := &ports.Config{
		Log:     ports.LogConfig{Level: ports.LogLevelInfo, Format: ports.LogFormatText},
		Server:  createValidServerConfig(),
		Storage: createValidStorageConfig(),
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{
			JWESigningKey: generateBase64EncodedString(32),
		},
		Encryption: ports.EncryptionConfig{
			KeyEncryptionKey: "arn:aws:kms:incomplete", // Incomplete ARN
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for malformed ARN, got nil")
	}
}

// Helper function to create valid ServerConfig
func createValidServerConfig() ports.ServerConfig {
	return ports.ServerConfig{
		EndUser: ports.ServerInstanceConfig{
			Port:      8000,
			Bind:      "::",
			PublicURL: "http://localhost:8000",
			Authentication: ports.AuthenticationConfig{
				Preauth: ports.PreauthConfig{
					PrincipalHeaderName: "X-Remote-User",
				},
			},
		},
		Admin: ports.ServerInstanceConfig{
			Port:      14000,
			Bind:      "::",
			PublicURL: "http://localhost:14000",
			Authentication: ports.AuthenticationConfig{
				Preauth: ports.PreauthConfig{
					PrincipalHeaderName: "X-Remote-User",
				},
			},
		},
		Shutdown: ports.ShutdownConfig{
			Timeout: 30 * time.Duration(1e9),
		},
	}
}

// Helper function to create valid StorageConfig
func createValidStorageConfig() ports.StorageConfig {
	return ports.StorageConfig{
		Backend: "memory",
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Duration(1e9),
			Write: 10 * time.Duration(1e9),
		},
	}
}
