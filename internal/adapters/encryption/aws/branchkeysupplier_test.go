package aws

import (
	"testing"

	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBranchKeyIdSupplier_GetBranchKeyId(t *testing.T) {
	supplier := &BranchKeyIdSupplier{}

	tests := []struct {
		name              string
		encryptionContext map[string]string
		wantID            string
		wantErr           string
	}{
		{
			name:              "service subject keeps existing naming",
			encryptionContext: map[string]string{"service_id": "550e8400-e29b-41d4-a716-446655440000"},
			wantID:            "service_550e8400-e29b-41d4-a716-446655440000_branch_key",
		},
		{
			name:              "signing key subject uses kid namespace",
			encryptionContext: map[string]string{"kid": "kid-123"},
			wantID:            "key_kid-123_branch_key",
		},
		{
			name:              "missing subject",
			encryptionContext: map[string]string{},
			wantErr:           "missing branch key subject",
		},
		{
			name:              "ambiguous subject",
			encryptionContext: map[string]string{"service_id": "550e8400-e29b-41d4-a716-446655440000", "kid": "kid-123"},
			wantErr:           "exactly one branch key subject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := supplier.GetBranchKeyId(mpltypes.GetBranchKeyIdInput{EncryptionContext: tt.encryptionContext})
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)
			assert.Equal(t, tt.wantID, output.BranchKeyId)
		})
	}
}
