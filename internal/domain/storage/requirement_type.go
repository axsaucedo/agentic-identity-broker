package storage

import (
	"fmt"
)

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
