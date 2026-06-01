package aws

import (
	"fmt"

	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/branchkey"
	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
)

// BranchKeyIdSupplier implements IBranchKeyIdSupplier for the AWS hierarchical keyring.
// It resolves branch key IDs based on the typed branch-key subject embedded in the
// encryption context (for example service_id for third-party services or kid for signing keys).
type BranchKeyIdSupplier struct{}

// GetBranchKeyId returns a deterministic branch key identifier based on the encryption context.
// Called by the AWS Encryption SDK hierarchical keyring during encryption/decryption.
func (d *BranchKeyIdSupplier) GetBranchKeyId(input mpltypes.GetBranchKeyIdInput) (*mpltypes.GetBranchKeyIdOutput, error) {
	subject, err := branchkey.SubjectFromEncryptionContext(input.EncryptionContext)
	if err != nil {
		return nil, fmt.Errorf("invalid branch key subject in encryption context: %w", err)
	}

	branchKeyID := branchkey.GenerateBranchKeyId(subject)
	if branchKeyID == "" {
		return nil, fmt.Errorf("failed to derive branch key ID from branch key subject")
	}

	return &mpltypes.GetBranchKeyIdOutput{BranchKeyId: branchKeyID}, nil
}

func (d *BranchKeyIdSupplier) GenerateBranchKeyId(subject domainencryption.BranchKeySubject) string {
	return branchkey.GenerateBranchKeyId(subject)
}

func (d *BranchKeyIdSupplier) ExtractSubjectFromBranchKey(branchKeyID string) (domainencryption.BranchKeySubject, error) {
	return branchkey.ExtractSubject(branchKeyID)
}
