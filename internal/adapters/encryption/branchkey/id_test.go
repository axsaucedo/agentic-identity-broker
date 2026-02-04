package branchkey

import (
	"strings"
	"testing"
)

func TestGenerateBranchKeyId(t *testing.T) {
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
			name:      "single character service",
			serviceID: "a",
			expected:  "service_a_branch_key",
		},
		{
			name:      "service with numbers",
			serviceID: "api123",
			expected:  "service_api123_branch_key",
		},
		{
			name:      "service with dashes",
			serviceID: "user-service",
			expected:  "service_user-service_branch_key",
		},
		{
			name:      "empty service ID",
			serviceID: "",
			expected:  "service__branch_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateBranchKeyId(tt.serviceID)
			if result != tt.expected {
				t.Errorf("GenerateBranchKeyId(%q) = %q, expected %q", tt.serviceID, result, tt.expected)
			}
		})
	}
}

func TestExtractServiceID(t *testing.T) {
	tests := []struct {
		name          string
		branchKeyID   string
		expectedID    string
		expectedError bool
		errorContains string
	}{
		{
			name:        "valid oauth2 service",
			branchKeyID: "service_oauth2_branch_key",
			expectedID:  "oauth2",
		},
		{
			name:        "valid github service",
			branchKeyID: "service_github_branch_key",
			expectedID:  "github",
		},
		{
			name:        "valid single character service",
			branchKeyID: "service_a_branch_key",
			expectedID:  "a",
		},
		{
			name:        "valid service with numbers",
			branchKeyID: "service_api123_branch_key",
			expectedID:  "api123",
		},
		{
			name:        "valid service with dashes",
			branchKeyID: "service_user-service_branch_key",
			expectedID:  "user-service",
		},
		{
			name:          "empty service ID in key",
			branchKeyID:   "service__branch_key",
			expectedError: true,
			errorContains: "empty service ID",
		},
		{
			name:          "missing prefix",
			branchKeyID:   "oauth2_branch_key",
			expectedError: true,
			errorContains: "invalid branch key ID format",
		},
		{
			name:          "missing suffix",
			branchKeyID:   "service_oauth2",
			expectedError: true,
			errorContains: "invalid branch key ID format",
		},
		{
			name:          "wrong prefix",
			branchKeyID:   "svc_oauth2_branch_key",
			expectedError: true,
			errorContains: "invalid branch key ID format",
		},
		{
			name:          "wrong suffix",
			branchKeyID:   "service_oauth2_key",
			expectedError: true,
			errorContains: "invalid branch key ID format",
		},
		{
			name:          "empty string",
			branchKeyID:   "",
			expectedError: true,
			errorContains: "invalid branch key ID format",
		},
		{
			name:          "too short",
			branchKeyID:   "service_branch_key",
			expectedError: true,
			errorContains: "invalid branch key ID format",
		},
		{
			name:        "malformed with extra underscores",
			branchKeyID: "service__oauth2__branch_key",
			expectedID:  "_oauth2_",
		},
		{
			name:          "case sensitive prefix",
			branchKeyID:   "Service_oauth2_branch_key",
			expectedError: true,
			errorContains: "invalid branch key ID format",
		},
		{
			name:          "case sensitive suffix",
			branchKeyID:   "service_oauth2_Branch_Key",
			expectedError: true,
			errorContains: "invalid branch key ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractServiceID(tt.branchKeyID)

			if tt.expectedError {
				if err == nil {
					t.Errorf("ExtractServiceID(%q) expected error but got none", tt.branchKeyID)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ExtractServiceID(%q) error %q does not contain %q", tt.branchKeyID, err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractServiceID(%q) unexpected error: %v", tt.branchKeyID, err)
				return
			}

			if result != tt.expectedID {
				t.Errorf("ExtractServiceID(%q) = %q, expected %q", tt.branchKeyID, result, tt.expectedID)
			}
		})
	}
}

func TestValidateBranchKeyID(t *testing.T) {
	tests := []struct {
		name        string
		branchKeyID string
		expected    bool
	}{
		{
			name:        "valid oauth2 service",
			branchKeyID: "service_oauth2_branch_key",
			expected:    true,
		},
		{
			name:        "valid github service",
			branchKeyID: "service_github_branch_key",
			expected:    true,
		},
		{
			name:        "valid single character service",
			branchKeyID: "service_a_branch_key",
			expected:    true,
		},
		{
			name:        "empty service ID in key",
			branchKeyID: "service__branch_key",
			expected:    false,
		},
		{
			name:        "missing prefix",
			branchKeyID: "oauth2_branch_key",
			expected:    false,
		},
		{
			name:        "missing suffix",
			branchKeyID: "service_oauth2",
			expected:    false,
		},
		{
			name:        "empty string",
			branchKeyID: "",
			expected:    false,
		},
		{
			name:        "wrong format",
			branchKeyID: "invalid_format",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateBranchKeyID(tt.branchKeyID)
			if result != tt.expected {
				t.Errorf("ValidateBranchKeyID(%q) = %v, expected %v", tt.branchKeyID, result, tt.expected)
			}
		})
	}
}

// TestSymmetricOperations verifies that GenerateBranchKeyId and ExtractServiceID are inverse operations
func TestSymmetricOperations(t *testing.T) {
	testServiceIDs := []string{
		"oauth2",
		"github",
		"api123",
		"user-service",
		"a",
		"service_with_underscores",
		"service-with-dashes",
	}

	for _, serviceID := range testServiceIDs {
		t.Run(serviceID, func(t *testing.T) {
			// Generate -> Extract should return original service ID
			branchKeyID := GenerateBranchKeyId(serviceID)
			extractedID, err := ExtractServiceID(branchKeyID)

			if err != nil {
				t.Errorf("ExtractServiceID failed for generated ID: %v", err)
				return
			}

			if extractedID != serviceID {
				t.Errorf("Symmetric operation failed: %q -> %q -> %q", serviceID, branchKeyID, extractedID)
			}

			// Validate the generated ID
			if !ValidateBranchKeyID(branchKeyID) {
				t.Errorf("Generated branch key ID failed validation: %q", branchKeyID)
			}
		})
	}
}
