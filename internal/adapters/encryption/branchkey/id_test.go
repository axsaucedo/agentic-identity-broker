package branchkey

import (
	"strings"
	"testing"

	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

func TestGenerateBranchKeyId_ServiceSubject(t *testing.T) {
	tests := []struct {
		name     string
		subject  domainencryption.BranchKeySubject
		expected string
	}{
		{
			name:     "service subject",
			subject:  domainencryption.NewServiceBranchKeySubject(id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440001")),
			expected: "service_550e8400-e29b-41d4-a716-446655440001_branch_key",
		},
		{
			name:    "zero service subject returns error",
			subject: domainencryption.BranchKeySubject{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateBranchKeyId(tt.subject)
			if tt.expected == "" {
				if err == nil {
					t.Fatalf("GenerateBranchKeyId(%q) expected error but got none", tt.subject.Identifier())
				}
				return
			}
			if err != nil {
				t.Fatalf("GenerateBranchKeyId(%q) unexpected error: %v", tt.subject.Identifier(), err)
			}
			if result != tt.expected {
				t.Errorf("GenerateBranchKeyId(%q) = %q, expected %q", tt.subject.Identifier(), result, tt.expected)
			}
		})
	}
}

func TestExtractSubject_ServiceSubject(t *testing.T) {
	tests := []struct {
		name          string
		branchKeyID   string
		expectedID    string
		expectedError bool
		errorContains string
	}{
		{
			name:        "valid service subject",
			branchKeyID: "service_550e8400-e29b-41d4-a716-446655440021_branch_key",
			expectedID:  "550e8400-e29b-41d4-a716-446655440021",
		},
		{
			name:          "empty service ID in key",
			branchKeyID:   "service__branch_key",
			expectedError: true,
			errorContains: "empty service ID",
		},
		{
			name:          "invalid service subject",
			branchKeyID:   "service_not-a-uuid_branch_key",
			expectedError: true,
			errorContains: "invalid service subject",
		},
		{
			name:          "missing prefix",
			branchKeyID:   "550e8400-e29b-41d4-a716-446655440021_branch_key",
			expectedError: true,
			errorContains: "invalid branch key ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, err := ExtractSubject(tt.branchKeyID)

			if tt.expectedError {
				if err == nil {
					t.Errorf("ExtractSubject(%q) expected error but got none", tt.branchKeyID)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ExtractSubject(%q) error %q does not contain %q", tt.branchKeyID, err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractSubject(%q) unexpected error: %v", tt.branchKeyID, err)
				return
			}

			if subject.Kind() != domainencryption.BranchKeySubjectKindService {
				t.Errorf("ExtractSubject(%q) kind = %q, expected %q", tt.branchKeyID, subject.Kind(), domainencryption.BranchKeySubjectKindService)
			}
			if subject.Identifier() != tt.expectedID {
				t.Errorf("ExtractSubject(%q) identifier = %q, expected %q", tt.branchKeyID, subject.Identifier(), tt.expectedID)
			}
		})
	}
}

// TestSymmetricOperations verifies that GenerateBranchKeyId and ExtractSubject are inverse operations.
func TestSymmetricOperations(t *testing.T) {
	testServiceIDs := []struct {
		name string
		uuid string
	}{
		{name: "oauth2", uuid: "550e8400-e29b-41d4-a716-446655440011"},
		{name: "github", uuid: "550e8400-e29b-41d4-a716-446655440012"},
	}

	for _, tt := range testServiceIDs {
		t.Run(tt.name, func(t *testing.T) {
			subject := domainencryption.NewServiceBranchKeySubject(id.MustParseServiceID(tt.uuid))
			branchKeyID, err := GenerateBranchKeyId(subject)
			if err != nil {
				t.Fatalf("GenerateBranchKeyId failed for %q: %v", tt.uuid, err)
			}
			extractedSubject, err := ExtractSubject(branchKeyID)

			if err != nil {
				t.Errorf("ExtractSubject failed for generated ID: %v", err)
				return
			}

			if extractedSubject.Kind() != subject.Kind() {
				t.Errorf("Symmetric operation failed: expected kind %q, got %q", subject.Kind(), extractedSubject.Kind())
			}
			if extractedSubject.Identifier() != subject.Identifier() {
				t.Errorf("Symmetric operation failed: %q -> %q -> %q", tt.uuid, branchKeyID, extractedSubject.Identifier())
			}

			if _, err := ExtractSubject(branchKeyID); err != nil {
				t.Errorf("ExtractSubject failed validation for generated ID %q: %v", branchKeyID, err)
			}
		})
	}
}
