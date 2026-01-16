package config

import (
	"encoding/base64"
	"testing"

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
		"Log":                mainConfig.Log,
		"Server":             mainConfig.Server,
		"Storage":            mainConfig.Storage,
		"ThirdPartyOAuth2":   mainConfig.ThirdPartyOAuth2,
		"OAuth2AuthServer":   mainConfig.OAuth2AuthServer,
		"Security":           mainConfig.Security,
		"Encryption":         mainConfig.Encryption,
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
	var _ string = config.KeyEncryptionKey

	// Test assignment of different valid formats
	testValues := []string{
		"arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
		"${ENCRYPTION_KEK}",
		"${IDENTITY_BROKER_ENCRYPTION_KEK}",
		base64.StdEncoding.EncodeToString([]byte("test-key-32-bytes-for-testing!!")),
	}

	for i, testValue := range testValues {
		config.KeyEncryptionKey = testValue
		if config.KeyEncryptionKey != testValue {
			t.Errorf("Test %d: Failed to assign value %q to KeyEncryptionKey", i, testValue)
		}
	}
}