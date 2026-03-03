package model

import "fmt"

// RequirementType defines whether a service is mandatory or optional for an agent.
// Mandatory services block authorization if requirements aren't met.
// Optional services are displayed but don't block authorization.
type RequirementType string

const (
	// RequirementTypeMandatory indicates the service is required.
	// Authorization fails if user lacks active session with required scopes.
	RequirementTypeMandatory RequirementType = "mandatory"

	// RequirementTypeOptional indicates the service is optional.
	// Authorization proceeds even if user lacks session or scopes.
	RequirementTypeOptional RequirementType = "optional"
)

// Valid returns true if the requirement type is valid.
func (rt RequirementType) Valid() bool {
	return rt == RequirementTypeMandatory || rt == RequirementTypeOptional
}

// Validate returns an error if the requirement type is invalid.
func (rt RequirementType) Validate() error {
	if !rt.Valid() {
		return fmt.Errorf("invalid requirement type: %q (must be 'mandatory' or 'optional')", rt)
	}
	return nil
}

// String returns the string representation of the requirement type.
func (rt RequirementType) String() string {
	return string(rt)
}

// IsMandatory returns true if the requirement type is mandatory.
func (rt RequirementType) IsMandatory() bool {
	return rt == RequirementTypeMandatory
}

// IsOptional returns true if the requirement type is optional.
func (rt RequirementType) IsOptional() bool {
	return rt == RequirementTypeOptional
}

// ServiceRequirement represents a third-party OAuth2 service that an agent requires or can optionally use.
// Each requirement specifies:
// - ServiceID: UUID of the third-party service
// - RequirementType: Whether the service is mandatory or optional
// - RequiredScopes: OAuth2 scopes that must be granted for this service
//
// Domain invariants:
// - ServiceID must be non-empty
// - RequirementType must be "mandatory" or "optional"
// - RequiredScopes must contain at least one scope
// - Each scope name must be non-empty
type ServiceRequirement struct {
	ServiceID       string
	RequirementType RequirementType
	RequiredScopes  []string
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

// Copy creates a deep copy of the ServiceRequirement.
func (sr *ServiceRequirement) Copy() ServiceRequirement {
	result := ServiceRequirement{
		ServiceID:       sr.ServiceID,
		RequirementType: sr.RequirementType,
	}
	if sr.RequiredScopes != nil {
		result.RequiredScopes = make([]string, len(sr.RequiredScopes))
		copy(result.RequiredScopes, sr.RequiredScopes)
	}
	return result
}
