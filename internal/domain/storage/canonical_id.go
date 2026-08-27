package storage

import "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/canonical"

// ValidateCanonicalID accepts an absent canonical ID and validates a supplied one.
func ValidateCanonicalID(canonicalID *string) error {
	return canonical.Validate(canonicalID)
}
