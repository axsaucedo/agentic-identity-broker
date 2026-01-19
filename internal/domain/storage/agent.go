package storage

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Agent represents an AI agent registered in the identity broker.
// An agent can request delegated permissions from users to access third-party services.
type Agent struct {
	ID                   string               `json:"id" db:"id"`
	ClientID             string               `json:"client_id" db:"client_id"`
	ExternalID           *string              `json:"external_id,omitempty" db:"external_id"`
	DisplayName          string               `json:"display_name" db:"display_name"`
	Description          string               `json:"description" db:"description"`
	GovernanceURL        *string              `json:"governance_url,omitempty" db:"governance_url"`
	UserDocumentationURL *string              `json:"user_documentation_url,omitempty" db:"user_documentation_url"`
	AgentInterfaceURL    *string              `json:"agent_interface_url,omitempty" db:"agent_interface_url"`
	ServiceRequirements  []ServiceRequirement `json:"service_requirements,omitempty" db:"service_requirements"`
	CreatedAt            time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at" db:"updated_at"`
}

// Validate performs validation on the Agent entity.
// Returns an error if any validation rules are violated.
func (a *Agent) Validate() error {
	// Required fields
	if a.ID == "" {
		return errors.New("agent ID cannot be empty")
	}
	if a.ClientID == "" {
		return errors.New("client_id is required")
	}
	if a.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if len(a.DisplayName) > 255 {
		return errors.New("display_name exceeds 255 characters")
	}
	if a.Description == "" {
		return errors.New("description is required")
	}
	if len(a.Description) > 1000 {
		return errors.New("description exceeds 1000 characters")
	}

	// URL validation (SR-011: prevent injection attacks)
	if a.GovernanceURL != nil && !isValidURL(*a.GovernanceURL) {
		return errors.New("governance_url is not a valid HTTP/HTTPS URL")
	}
	if a.UserDocumentationURL != nil && !isValidURL(*a.UserDocumentationURL) {
		return errors.New("user_documentation_url is not a valid HTTP/HTTPS URL")
	}
	if a.AgentInterfaceURL != nil && !isValidURL(*a.AgentInterfaceURL) {
		return errors.New("agent_interface_url is not a valid HTTP/HTTPS URL")
	}

	// Service requirements validation
	if err := a.ValidateServiceRequirements(); err != nil {
		return fmt.Errorf("service_requirements validation failed: %w", err)
	}

	return nil
}

// isValidURL validates that a string is a valid HTTP or HTTPS URL.
func isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// Copy creates a deep copy of the Agent to prevent external mutation.
func (a *Agent) Copy() *Agent {
	if a == nil {
		return nil
	}

	copy := &Agent{
		ID:          a.ID,
		ClientID:    a.ClientID,
		DisplayName: a.DisplayName,
		Description: a.Description,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}

	if a.ExternalID != nil {
		externalID := *a.ExternalID
		copy.ExternalID = &externalID
	}
	if a.GovernanceURL != nil {
		governanceURL := *a.GovernanceURL
		copy.GovernanceURL = &governanceURL
	}
	if a.UserDocumentationURL != nil {
		userDocURL := *a.UserDocumentationURL
		copy.UserDocumentationURL = &userDocURL
	}
	if a.AgentInterfaceURL != nil {
		agentURL := *a.AgentInterfaceURL
		copy.AgentInterfaceURL = &agentURL
	}

	// Deep copy service requirements
	if a.ServiceRequirements != nil {
		copy.ServiceRequirements = make([]ServiceRequirement, len(a.ServiceRequirements))
		for i, sr := range a.ServiceRequirements {
			copy.ServiceRequirements[i] = ServiceRequirement{
				ServiceID:       sr.ServiceID,
				RequirementType: sr.RequirementType,
				RequiredScopes:  append([]string(nil), sr.RequiredScopes...),
			}
		}
	}

	return copy
}

// ValidateForCreate validates an agent before creation.
// ID will be generated, so it may be empty.
func (a *Agent) ValidateForCreate() error {
	if a.ClientID == "" {
		return errors.New("client_id is required")
	}
	if a.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if len(a.DisplayName) > 255 {
		return fmt.Errorf("display_name exceeds 255 characters (got %d)", len(a.DisplayName))
	}
	if a.Description == "" {
		return errors.New("description is required")
	}
	if len(a.Description) > 1000 {
		return fmt.Errorf("description exceeds 1000 characters (got %d)", len(a.Description))
	}

	// URL validation
	if a.GovernanceURL != nil && !isValidURL(*a.GovernanceURL) {
		return errors.New("governance_url is not a valid HTTP/HTTPS URL")
	}
	if a.UserDocumentationURL != nil && !isValidURL(*a.UserDocumentationURL) {
		return errors.New("user_documentation_url is not a valid HTTP/HTTPS URL")
	}
	if a.AgentInterfaceURL != nil && !isValidURL(*a.AgentInterfaceURL) {
		return errors.New("agent_interface_url is not a valid HTTP/HTTPS URL")
	}

	// Service requirements validation
	if err := a.ValidateServiceRequirements(); err != nil {
		return fmt.Errorf("service_requirements validation failed: %w", err)
	}

	return nil
}

// ValidateServiceRequirements validates the service requirements array.
// Returns error if:
// - Any service requirement is invalid
// - Duplicate service_id exists in the array
func (a *Agent) ValidateServiceRequirements() error {
	if len(a.ServiceRequirements) == 0 {
		return nil // Empty array is valid (no requirements)
	}

	// Validate each service requirement
	for i, sr := range a.ServiceRequirements {
		if err := sr.Validate(); err != nil {
			return fmt.Errorf("service_requirements[%d] invalid: %w", i, err)
		}
	}

	// Check for duplicate service_id (domain invariant)
	seen := make(map[string]int)
	for i, sr := range a.ServiceRequirements {
		if prevIdx, exists := seen[sr.ServiceID]; exists {
			return fmt.Errorf("duplicate service_id %q found at indices %d and %d", sr.ServiceID, prevIdx, i)
		}
		seen[sr.ServiceID] = i
	}

	return nil
}
