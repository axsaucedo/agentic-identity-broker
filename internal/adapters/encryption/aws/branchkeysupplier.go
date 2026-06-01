package aws

import (
	"fmt"

	mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"

	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// BranchKeyIdSupplier implements IBranchKeyIdSupplier for the AWS hierarchical keyring.
// It resolves branch key IDs based on the typed branch-key subject embedded in the
// encryption context (for example service_id for third-party services or kid for signing keys).
type BranchKeyIdSupplier struct {
	provider ports.BranchKeyIdProvider
}

func NewBranchKeyIdSupplier(provider ports.BranchKeyIdProvider) *BranchKeyIdSupplier {
	return &BranchKeyIdSupplier{provider: provider}
}

// GetBranchKeyId returns a deterministic branch key identifier based on the encryption context.
// Called by the AWS Encryption SDK hierarchical keyring during encryption/decryption.
func (d *BranchKeyIdSupplier) GetBranchKeyId(input mpltypes.GetBranchKeyIdInput) (*mpltypes.GetBranchKeyIdOutput, error) {
	if d.provider == nil {
		return nil, fmt.Errorf("branch key ID provider not configured")
	}

	subject, err := domainencryption.BranchKeySubjectFromEncryptionContext(input.EncryptionContext)
	if err != nil {
		return nil, fmt.Errorf("invalid branch key subject in encryption context: %w", err)
	}

	branchKeyID := d.provider.GenerateBranchKeyId(subject)
	if branchKeyID == "" {
		return nil, fmt.Errorf("failed to derive branch key ID from branch key subject")
	}

	return &mpltypes.GetBranchKeyIdOutput{BranchKeyId: branchKeyID}, nil
}

func (d *BranchKeyIdSupplier) GenerateBranchKeyId(subject domainencryption.BranchKeySubject) string {
	if d.provider == nil {
		return ""
	}
	return d.provider.GenerateBranchKeyId(subject)
}

func (d *BranchKeyIdSupplier) ExtractSubjectFromBranchKey(branchKeyID string) (domainencryption.BranchKeySubject, error) {
	if d.provider == nil {
		return domainencryption.BranchKeySubject{}, fmt.Errorf("branch key ID provider not configured")
	}
	return d.provider.ExtractSubjectFromBranchKey(branchKeyID)
}
