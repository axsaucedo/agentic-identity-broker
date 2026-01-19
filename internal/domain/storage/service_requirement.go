package storage

import (
	"fmt"
)

// ServiceRequirement represents a third-party OAuth2 service that an agent requires or can optionally use.
// Each requirement specifies:
// - ServiceID: UUID of the third-party service
// - RequirementType: Whether the service is mandatory or optional
// - RequiredScopes: OAuth2 scopes that must be granted for this service
//
// Domain invariants:
// - ServiceID must be a valid UUID
// - RequirementType must be "mandatory" or "optional"
// - RequiredScopes must contain at least one scope
// - Each scope name must be non-empty
type ServiceRequirement struct {
	ServiceID       string          `json:"service_id" db:"service_id"`
	RequirementType RequirementType `json:"requirement_type" db:"requirement_type"`
	RequiredScopes  []string        `json:"required_scopes" db:"required_scopes"`
}

// Validate checks if the service requirement is valid.
// Returns error if:
// - ServiceID is empty
// - RequirementType is invalid
// - RequiredScopes is empty
// - Any scope is empty
func (sr *ServiceRequirement) Validate() error {
	if sr.ServiceID == "" {
		return fmt.Errorf("service_id is required")
	}

	if err := sr.RequirementType.Validate(); err != nil {
		return fmt.Errorf("requirement_type validation failed: %w", err)
	}

	if len(sr.RequiredScopes) == 0 {
		return fmt.Errorf("required_scopes must contain at least one scope")
	}

	for i, scope := range sr.RequiredScopes {
		if scope == "" {
			return fmt.Errorf("required_scopes[%d] is empty", i)
		}
	}

	return nil
}

// HasScope returns true if the service requirement includes the specified scope.
// Scope comparison is case-sensitive.
func (sr *ServiceRequirement) HasScope(scope string) bool {
	for _, s := range sr.RequiredScopes {
		if s == scope {
			return true
		}
	}
	return false
}

// IsMandatory returns true if this is a mandatory service requirement.
func (sr *ServiceRequirement) IsMandatory() bool {
	return sr.RequirementType.IsMandatory()
}

// IsOptional returns true if this is an optional service requirement.
func (sr *ServiceRequirement) IsOptional() bool {
	return sr.RequirementType.IsOptional()
}
