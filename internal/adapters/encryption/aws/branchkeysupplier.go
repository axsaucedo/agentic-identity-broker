package aws

import (
	"fmt"

	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
)

type branchKeySupplier struct {
	branchKeyA string
	branchKeyB string
}

func (b *branchKeySupplier) GetBranchKeyId(input mpltypes.GetBranchKeyIdInput) (*mpltypes.GetBranchKeyIdOutput, error) {
	// We MUST use the encryption context to determine
	// the Branch Key ID.
	ec := input.EncryptionContext
	if value, exists := ec["tenant"]; !exists || value == "" {
		return nil, fmt.Errorf("EncryptionContext invalid, does not contain expected tenant key value pair.")
	}
	branchKeyIdentifier := ec["tenant"]
	if branchKeyIdentifier == "TenantA" {
		return &mpltypes.GetBranchKeyIdOutput{BranchKeyId: b.branchKeyA}, nil
	} else if branchKeyIdentifier == "TenantB" {
		return &mpltypes.GetBranchKeyIdOutput{BranchKeyId: b.branchKeyB}, nil
	} else {
		return &mpltypes.GetBranchKeyIdOutput{}, fmt.Errorf("unknown branch key identifier")
	}
}
