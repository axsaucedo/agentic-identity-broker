package branchkey

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

func TestGenerateBranchKeyId_SigningKeySubject(t *testing.T) {
	subject := domainencryption.NewSigningKeyBranchKeySubject(id.NewKeyID("kid-123"))
	assert.Equal(t, "key_kid-123_branch_key", GenerateBranchKeyId(subject))
}

func TestExtractSubject_SigningKeySubject(t *testing.T) {
	tests := []struct {
		name          string
		branchKeyID   string
		expectedID    string
		expectedError string
	}{
		{
			name:        "valid signing key",
			branchKeyID: "key_kid-123_branch_key",
			expectedID:  "kid-123",
		},
		{
			name:          "empty signing key id",
			branchKeyID:   "key__branch_key",
			expectedError: "empty signing key ID",
		},
		{
			name:          "wrong prefix",
			branchKeyID:   "service_kid-123_branch_key",
			expectedError: "invalid service subject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, err := ExtractSubject(tt.branchKeyID)
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, domainencryption.BranchKeySubjectKindSigningKey, subject.Kind())
			assert.Equal(t, tt.expectedID, subject.Identifier())
		})
	}
}

func TestExtractSubject_SigningKeyValidation(t *testing.T) {
	_, err := ExtractSubject("key_kid-123_branch_key")
	require.NoError(t, err)

	_, err = ExtractSubject("key__branch_key")
	require.Error(t, err)
}
