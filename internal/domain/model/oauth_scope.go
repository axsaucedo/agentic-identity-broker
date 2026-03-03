package model

import "errors"

// OAuthScope represents a permission scope within a third-party OAuth2 service.
// Scopes define what resources and actions an agent can access.
type OAuthScope struct {
	ScopeValue  string
	Description string
}

// Validate performs validation on the OAuthScope.
func (s *OAuthScope) Validate() error {
	if s.ScopeValue == "" {
		return errors.New("scope_value is required")
	}
	if s.Description == "" {
		return errors.New("description is required")
	}
	return nil
}

// Copy creates a deep copy of the OAuthScope.
func (s *OAuthScope) Copy() OAuthScope {
	return OAuthScope{
		ScopeValue:  s.ScopeValue,
		Description: s.Description,
	}
}
