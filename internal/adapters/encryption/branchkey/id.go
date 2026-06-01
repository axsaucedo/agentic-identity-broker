package branchkey

import (
	"fmt"
	"strings"

	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

const (
	servicePrefix    = "service_"
	signingKeyPrefix = "key_"
	suffix           = "_branch_key"

	ServiceIDFormat    = servicePrefix + "%s" + suffix
	SigningKeyIDFormat = signingKeyPrefix + "%s" + suffix
)

func generateServiceBranchKeyID(serviceID string) string {
	return fmt.Sprintf(ServiceIDFormat, serviceID)
}

func generateSigningKeyBranchKeyID(signingKeyID string) string {
	return fmt.Sprintf(SigningKeyIDFormat, signingKeyID)
}

// GenerateBranchKeyId generates a deterministic branch key ID from a typed branch key subject.
func GenerateBranchKeyId(subject domainencryption.BranchKeySubject) string {
	switch subject.Kind() {
	case domainencryption.BranchKeySubjectKindService:
		serviceID, _ := subject.ServiceID()
		return generateServiceBranchKeyID(serviceID.String())
	case domainencryption.BranchKeySubjectKindSigningKey:
		keyID, _ := subject.KeyID()
		return generateSigningKeyBranchKeyID(keyID.String())
	default:
		return ""
	}
}

func extractServiceID(branchKeyID string) (string, error) {
	return extractIdentifier(branchKeyID, servicePrefix, ServiceIDFormat, "service ID")
}

func extractSigningKeyID(branchKeyID string) (string, error) {
	return extractIdentifier(branchKeyID, signingKeyPrefix, SigningKeyIDFormat, "signing key ID")
}

// ExtractSubject parses a branch key ID back into its typed branch key subject.
func ExtractSubject(branchKeyID string) (domainencryption.BranchKeySubject, error) {
	switch {
	case strings.HasPrefix(branchKeyID, servicePrefix):
		serviceID, err := extractServiceID(branchKeyID)
		if err != nil {
			return domainencryption.BranchKeySubject{}, err
		}
		parsed, err := id.ParseServiceID(serviceID)
		if err != nil {
			return domainencryption.BranchKeySubject{}, fmt.Errorf("invalid service subject in branch key ID %q: %w", branchKeyID, err)
		}
		return domainencryption.NewServiceBranchKeySubject(parsed), nil
	case strings.HasPrefix(branchKeyID, signingKeyPrefix):
		signingKeyID, err := extractSigningKeyID(branchKeyID)
		if err != nil {
			return domainencryption.BranchKeySubject{}, err
		}
		return domainencryption.NewSigningKeyBranchKeySubject(id.NewKeyID(signingKeyID)), nil
	default:
		return domainencryption.BranchKeySubject{}, fmt.Errorf(
			"invalid branch key ID format: %q (expected format: %s or %s)",
			branchKeyID,
			ServiceIDFormat,
			SigningKeyIDFormat,
		)
	}
}

func extractIdentifier(branchKeyID, prefix, format, identifierName string) (string, error) {
	if len(branchKeyID) > len(prefix)+len(suffix) &&
		strings.HasPrefix(branchKeyID, prefix) &&
		strings.HasSuffix(branchKeyID, suffix) {
		identifier := branchKeyID[len(prefix) : len(branchKeyID)-len(suffix)]
		if identifier == "" {
			return "", fmt.Errorf("invalid branch key ID: empty %s in %q", identifierName, branchKeyID)
		}
		return identifier, nil
	}

	if branchKeyID == fmt.Sprintf(format, "") {
		return "", fmt.Errorf("invalid branch key ID: empty %s in %q", identifierName, branchKeyID)
	}

	return "", fmt.Errorf("invalid branch key ID format: %q (expected format: %s)", branchKeyID, format)
}
