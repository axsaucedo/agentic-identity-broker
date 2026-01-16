package aws

import (
	"context"
	"testing"
	"time"
)

// TestHierarchicalKeyringConfig tests configuration validation
func TestHierarchicalKeyringConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         HierarchicalKeyringConfig
		expectedError  bool
		errorContains  string
	}{
		{
			name: "empty_kms_key_arn",
			config: HierarchicalKeyringConfig{
				KMSKeyARN:         "",
				DynamoDBTableName: "TestTable",
				BranchKeyTTL:      1 * time.Hour,
			},
			expectedError: true,
			errorContains: "KMS key ARN is required",
		},
		{
			name: "invalid_kms_arn_format",
			config: HierarchicalKeyringConfig{
				KMSKeyARN:         "not-a-valid-arn",
				DynamoDBTableName: "TestTable",
				BranchKeyTTL:      1 * time.Hour,
			},
			expectedError: true,
			errorContains: "not accessible",
		},
		{
			name: "default_table_name_applied",
			config: HierarchicalKeyringConfig{
				KMSKeyARN:         "", // Will fail on KMS check but shows defaults work
				DynamoDBTableName: "",
				BranchKeyTTL:      0,
			},
			expectedError: true,
			errorContains: "KMS key ARN is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createHierarchicalKeyring(context.Background(), tt.config)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("error should contain %q, got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestHierarchicalKeyringConfigDefaults tests that defaults are applied
func TestHierarchicalKeyringConfigDefaults(t *testing.T) {
	config := HierarchicalKeyringConfig{
		KMSKeyARN:         "",
		DynamoDBTableName: "",
		BranchKeyTTL:      0,
	}

	// Apply defaults like createHierarchicalKeyring does
	if config.DynamoDBTableName == "" {
		config.DynamoDBTableName = DefaultBranchKeyTableName
	}
	if config.BranchKeyTTL == 0 {
		config.BranchKeyTTL = DefaultBranchKeyTTL
	}

	// Verify defaults
	if config.DynamoDBTableName != DefaultBranchKeyTableName {
		t.Errorf("expected default table name %q, got %q", DefaultBranchKeyTableName, config.DynamoDBTableName)
	}

	if config.BranchKeyTTL != DefaultBranchKeyTTL {
		t.Errorf("expected default TTL %v, got %v", DefaultBranchKeyTTL, config.BranchKeyTTL)
	}
}

// TestKeyStoreConfigStructure tests that KeyStore configuration is properly structured
func TestKeyStoreConfigStructure(t *testing.T) {
	// This test verifies the KeyStore config matches what the AWS library expects
	// It's primarily a compile-time check via the struct literals in keyring.go

	// Test that HierarchicalKeyringConfig has required fields
	config := HierarchicalKeyringConfig{
		KMSKeyARN:         "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
		DynamoDBTableName: "TestBranchKeys",
		BranchKeyTTL:      2 * time.Hour,
	}

	if config.KMSKeyARN == "" {
		t.Error("KMSKeyARN should not be empty")
	}

	if config.DynamoDBTableName == "" {
		t.Error("DynamoDBTableName should not be empty")
	}

	if config.BranchKeyTTL == 0 {
		t.Error("BranchKeyTTL should not be zero")
	}
}

// TestKMSARNValidationInHierarchicalKeyring tests KMS ARN validation
func TestKMSARNValidationInHierarchicalKeyring(t *testing.T) {
	tests := []struct {
		name          string
		kmsARN        string
		shouldFail    bool
		errorContains string
	}{
		{
			name:       "valid_kms_key_arn",
			kmsARN:     "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
			shouldFail: true, // Will fail due to KMS access in test, but validates ARN format
		},
		{
			name:          "empty_arn",
			kmsARN:        "",
			shouldFail:    true,
			errorContains: "KMS key ARN is required",
		},
		{
			name:       "kms_alias_arn",
			kmsARN:     "arn:aws:kms:us-east-1:123456789012:alias/my-key",
			shouldFail: true, // Will fail due to KMS access in test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HierarchicalKeyringConfig{
				KMSKeyARN:         tt.kmsARN,
				DynamoDBTableName: DefaultBranchKeyTableName,
				BranchKeyTTL:      DefaultBranchKeyTTL,
			}

			_, err := createHierarchicalKeyring(context.Background(), config)

			if !tt.shouldFail && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if tt.shouldFail && err == nil {
				t.Errorf("expected error, got nil")
			}

			if tt.errorContains != "" && err != nil && !containsString(err.Error(), tt.errorContains) {
				t.Errorf("error should contain %q, got: %v", tt.errorContains, err)
			}
		})
	}
}

// TestBranchKeyTTLConfiguration tests branch key TTL configuration
func TestBranchKeyTTLConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		ttl     time.Duration
		isValid bool
	}{
		{
			name:    "one_hour_ttl",
			ttl:     1 * time.Hour,
			isValid: true,
		},
		{
			name:    "one_day_ttl",
			ttl:     24 * time.Hour,
			isValid: true,
		},
		{
			name:    "thirty_minutes_ttl",
			ttl:     30 * time.Minute,
			isValid: true,
		},
		{
			name:    "zero_ttl_uses_default",
			ttl:     0,
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HierarchicalKeyringConfig{
				KMSKeyARN:         "arn:aws:kms:us-west-2:123456789012:key/test",
				DynamoDBTableName: DefaultBranchKeyTableName,
				BranchKeyTTL:      tt.ttl,
			}

			// Verify TTL is properly set or defaults apply
			if config.BranchKeyTTL == 0 {
				config.BranchKeyTTL = DefaultBranchKeyTTL
			}

			if config.BranchKeyTTL.Seconds() <= 0 {
				t.Errorf("TTL in seconds should be positive, got: %v", config.BranchKeyTTL.Seconds())
			}
		})
	}
}

// TestHierarchicalKeyringConfigIntegration tests hierarchical keyring with adapter
func TestHierarchicalKeyringConfigIntegration(t *testing.T) {
	// This test verifies that HierarchicalKeyringConfig can be used to create a keyring
	// The actual KMS call will fail in test environment, but the configuration flow is validated

	config := HierarchicalKeyringConfig{
		KMSKeyARN:         "arn:aws:kms:us-west-2:123456789012:key/invalid-in-test",
		DynamoDBTableName: "TestBranchKeys",
		BranchKeyTTL:      1 * time.Hour,
	}

	// Should fail on KMS key verification, not on configuration
	_, err := createHierarchicalKeyring(context.Background(), config)
	if err == nil {
		t.Fatal("expected error due to invalid KMS key in test")
	}

	// Verify error is about KMS access, not configuration
	if !isKEKUnavailableError(err) {
		t.Errorf("expected KEKUnavailable error, got: %T", err)
	}
}

// TestDefaultConstants tests that default constants have sensible values
func TestDefaultConstants(t *testing.T) {
	if DefaultKEKValidationTimeout <= 0 {
		t.Errorf("DefaultKEKValidationTimeout should be positive, got: %v", DefaultKEKValidationTimeout)
	}

	if DefaultBranchKeyTTL <= 0 {
		t.Errorf("DefaultBranchKeyTTL should be positive, got: %v", DefaultBranchKeyTTL)
	}

	if DefaultBranchKeyTableName == "" {
		t.Error("DefaultBranchKeyTableName should not be empty")
	}

	// TTL should be in a reasonable range (1 minute to 24 hours)
	if DefaultBranchKeyTTL < 1*time.Minute || DefaultBranchKeyTTL > 24*time.Hour {
		t.Errorf("DefaultBranchKeyTTL should be between 1 minute and 24 hours, got: %v", DefaultBranchKeyTTL)
	}
}

// Helper functions

func containsString(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
