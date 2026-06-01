package branchkey

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
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
		name     string
		subject  domainencryption.BranchKeySubject
		expected string
	}{
		{
			name:     "service subject",
			subject:  domainencryption.NewServiceBranchKeySubject(id.MustParseServiceID(testUUID1)),
			expected: "service_" + testUUID1 + "_branch_key",
		},
		{
			name:     "signing key subject",
			subject:  domainencryption.NewSigningKeyBranchKeySubject(id.NewKeyID("kid-123")),
			expected: "key_kid-123_branch_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branchKeyID, err := provider.GenerateBranchKeyId(tt.subject)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, branchKeyID)
		})
	}
}

func TestDefaultProvider_ExtractSubjectFromBranchKey(t *testing.T) {
	provider := NewDefaultProvider()

	tests := []struct {
		name        string
		branchKeyID string
		assertFn    func(t *testing.T, subject domainencryption.BranchKeySubject)
		wantErr     string
	}{
		{
			name:        "service branch key",
			branchKeyID: "service_" + testUUID1 + "_branch_key",
			assertFn: func(t *testing.T, subject domainencryption.BranchKeySubject) {
				t.Helper()
				assert.Equal(t, domainencryption.BranchKeySubjectKindService, subject.Kind())
				serviceID, ok := subject.ServiceID()
				require.True(t, ok)
				assert.Equal(t, id.MustParseServiceID(testUUID1), serviceID)
			},
		},
		{
			name:        "signing key branch key",
			branchKeyID: "key_kid-123_branch_key",
			assertFn: func(t *testing.T, subject domainencryption.BranchKeySubject) {
				t.Helper()
				assert.Equal(t, domainencryption.BranchKeySubjectKindSigningKey, subject.Kind())
				signingKeyID, ok := subject.KeyID()
				require.True(t, ok)
				assert.Equal(t, id.NewKeyID("kid-123"), signingKeyID)
			},
		},
		{
			name:        "invalid format",
			branchKeyID: "invalid_format",
			wantErr:     "invalid branch key ID format",
		},
		{
			name:        "non uuid service subject",
			branchKeyID: "service_not-a-uuid_branch_key",
			wantErr:     "invalid service subject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, err := provider.ExtractSubjectFromBranchKey(tt.branchKeyID)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			tt.assertFn(t, subject)
		})
	}
}

func TestDefaultProvider_SymmetricOperations(t *testing.T) {
	provider := NewDefaultProvider()

	testSubjects := []domainencryption.BranchKeySubject{
		domainencryption.NewServiceBranchKeySubject(id.MustParseServiceID(testUUID1)),
		domainencryption.NewServiceBranchKeySubject(id.MustParseServiceID(testUUID2)),
		domainencryption.NewServiceBranchKeySubject(id.MustParseServiceID(testUUID3)),
		domainencryption.NewSigningKeyBranchKeySubject(id.NewKeyID("kid-1")),
		domainencryption.NewSigningKeyBranchKeySubject(id.NewKeyID("kid-2")),
	}

	for _, subject := range testSubjects {
		t.Run(subject.Identifier(), func(t *testing.T) {
			branchKeyID, err := provider.GenerateBranchKeyId(subject)
			require.NoError(t, err)
			extracted, err := provider.ExtractSubjectFromBranchKey(branchKeyID)
			require.NoError(t, err)
			assert.Equal(t, subject.Kind(), extracted.Kind())
			assert.Equal(t, subject.Identifier(), extracted.Identifier())
		})
	}
}

func TestNewDefaultProvider(t *testing.T) {
	provider := NewDefaultProvider()
	if provider == nil {
		t.Fatal("NewDefaultProvider() returned nil")
	}

	serviceSubject := domainencryption.NewServiceBranchKeySubject(id.MustParseServiceID(testUUID6))
	branchKeyID, err := provider.GenerateBranchKeyId(serviceSubject)
	require.NoError(t, err)
	assert.Equal(t, "service_"+testUUID6+"_branch_key", branchKeyID)
}
