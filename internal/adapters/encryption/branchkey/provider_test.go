package branchkey

import "testing"

func TestDefaultProvider_GenerateBranchKeyId(t *testing.T) {
	provider := NewDefaultProvider()

	tests := []struct {
		name      string
		serviceID string
		expected  string
	}{
		{
			name:      "oauth2 service",
			serviceID: "oauth2",
			expected:  "service_oauth2_branch_key",
		},
		{
			name:      "github service",
			serviceID: "github",
			expected:  "service_github_branch_key",
		},
		{
			name:      "empty service ID",
			serviceID: "",
			expected:  "service__branch_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.GenerateBranchKeyId(tt.serviceID)
			if result != tt.expected {
				t.Errorf("GenerateBranchKeyId(%q) = %q, expected %q", tt.serviceID, result, tt.expected)
			}
		})
	}
}

func TestDefaultProvider_ExtractServiceIdFromBranchKey(t *testing.T) {
	provider := NewDefaultProvider()

	tests := []struct {
		name        string
		branchKeyID string
		expected    string
	}{
		{
			name:        "valid oauth2 service",
			branchKeyID: "service_oauth2_branch_key",
			expected:    "oauth2",
		},
		{
			name:        "valid github service",
			branchKeyID: "service_github_branch_key",
			expected:    "github",
		},
		{
			name:        "invalid format returns empty string",
			branchKeyID: "invalid_format",
			expected:    "",
		},
		{
			name:        "missing prefix returns empty string",
			branchKeyID: "oauth2_branch_key",
			expected:    "",
		},
		{
			name:        "missing suffix returns empty string",
			branchKeyID: "service_oauth2",
			expected:    "",
		},
		{
			name:        "empty service ID returns empty string",
			branchKeyID: "service__branch_key",
			expected:    "",
		},
		{
			name:        "empty string returns empty string",
			branchKeyID: "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.ExtractServiceIdFromBranchKey(tt.branchKeyID)
			if result != tt.expected {
				t.Errorf("ExtractServiceIdFromBranchKey(%q) = %q, expected %q", tt.branchKeyID, result, tt.expected)
			}
		})
	}
}

// TestDefaultProvider_SymmetricOperations verifies that the provider's methods are inverse operations
func TestDefaultProvider_SymmetricOperations(t *testing.T) {
	provider := NewDefaultProvider()

	testServiceIDs := []string{
		"oauth2",
		"github",
		"api123",
		"user-service",
		"a",
	}

	for _, serviceID := range testServiceIDs {
		t.Run(serviceID, func(t *testing.T) {
			// Generate -> Extract should return original service ID
			branchKeyID := provider.GenerateBranchKeyId(serviceID)
			extractedID := provider.ExtractServiceIdFromBranchKey(branchKeyID)

			if extractedID != serviceID {
				t.Errorf("Symmetric operation failed: %q -> %q -> %q", serviceID, branchKeyID, extractedID)
			}
		})
	}
}

// TestNewDefaultProvider ensures the constructor creates a valid instance
func TestNewDefaultProvider(t *testing.T) {
	provider := NewDefaultProvider()
	if provider == nil {
		t.Error("NewDefaultProvider() returned nil")
	}

	// Test that the provider works correctly
	result := provider.GenerateBranchKeyId("test")
	expected := "service_test_branch_key"
	if result != expected {
		t.Errorf("New provider GenerateBranchKeyId(\"test\") = %q, expected %q", result, expected)
	}
}
