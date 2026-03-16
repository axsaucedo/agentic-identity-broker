package branchkey

import (
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

const (
	testUUID1 = "550e8400-e29b-41d4-a716-446655440001"
	testUUID2 = "550e8400-e29b-41d4-a716-446655440002"
	testUUID3 = "550e8400-e29b-41d4-a716-446655440003"
	testUUID4 = "550e8400-e29b-41d4-a716-446655440004"
	testUUID5 = "550e8400-e29b-41d4-a716-446655440005"
	testUUID6 = "550e8400-e29b-41d4-a716-446655440006"
)

func TestDefaultProvider_GenerateBranchKeyId(t *testing.T) {
	provider := NewDefaultProvider()

	tests := []struct {
		name      string
		serviceID id.ServiceID
		expected  string
	}{
		{
			name:      "first service",
			serviceID: id.MustParseServiceID(testUUID1),
			expected:  "service_" + testUUID1 + "_branch_key",
		},
		{
			name:      "second service",
			serviceID: id.MustParseServiceID(testUUID2),
			expected:  "service_" + testUUID2 + "_branch_key",
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
		expected    id.ServiceID
	}{
		{
			name:        "valid first service",
			branchKeyID: "service_" + testUUID1 + "_branch_key",
			expected:    id.MustParseServiceID(testUUID1),
		},
		{
			name:        "valid second service",
			branchKeyID: "service_" + testUUID2 + "_branch_key",
			expected:    id.MustParseServiceID(testUUID2),
		},
		{
			name:        "invalid format returns zero",
			branchKeyID: "invalid_format",
			expected:    id.ServiceID{},
		},
		{
			name:        "missing prefix returns zero",
			branchKeyID: testUUID1 + "_branch_key",
			expected:    id.ServiceID{},
		},
		{
			name:        "missing suffix returns zero",
			branchKeyID: "service_" + testUUID1,
			expected:    id.ServiceID{},
		},
		{
			name:        "non-uuid service ID returns zero",
			branchKeyID: "service_not-a-uuid_branch_key",
			expected:    id.ServiceID{},
		},
		{
			name:        "empty service ID returns zero",
			branchKeyID: "service__branch_key",
			expected:    id.ServiceID{},
		},
		{
			name:        "empty string returns zero",
			branchKeyID: "",
			expected:    id.ServiceID{},
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

	testServiceIDs := []id.ServiceID{
		id.MustParseServiceID(testUUID1),
		id.MustParseServiceID(testUUID2),
		id.MustParseServiceID(testUUID3),
		id.MustParseServiceID(testUUID4),
		id.MustParseServiceID(testUUID5),
	}

	for _, serviceID := range testServiceIDs {
		t.Run(serviceID.String(), func(t *testing.T) {
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
	testID := id.MustParseServiceID(testUUID6)
	result := provider.GenerateBranchKeyId(testID)
	expected := "service_" + testUUID6 + "_branch_key"
	if result != expected {
		t.Errorf("New provider GenerateBranchKeyId(%q) = %q, expected %q", testID, result, expected)
	}
}
