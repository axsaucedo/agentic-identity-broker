package storage

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// UserGrant represents a user delegating specific permissions to an agent
// for one or more third-party OAuth2 services.
type UserGrant struct {
	ID                     string            `json:"id" db:"id"`
	Principal              string            `json:"principal" db:"principal"`
	AgentID                string            `json:"agent_id" db:"agent_id"`
	ValidUntil             *time.Time        `json:"valid_until,omitempty" db:"valid_until"`
	DelegatedOAuth2Tokens  []DelegatedToken  `json:"delegated_oauth2_tokens" db:"delegated_oauth2_tokens"`
	CreatedAt              time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time         `json:"updated_at" db:"updated_at"`
}

// DelegatedToken represents delegation of specific scopes to a third-party service.
type DelegatedToken struct {
	ThirdpartyOAuth2ServiceID string   `json:"thirdparty_oauth2_service_id"`
	Scopes                    []string `json:"scopes"`
}

// Validate performs validation on the UserGrant entity.
func (g *UserGrant) Validate() error {
	// Required fields
	if g.ID == "" {
		return errors.New("grant ID cannot be empty")
	}
	if g.Principal == "" {
		return errors.New("principal is required")
	}
	if g.AgentID == "" {
		return errors.New("agent_id is required")
	}

	// valid_until must be in future if provided
	if g.ValidUntil != nil && g.ValidUntil.Before(time.Now()) {
		return errors.New("valid_until must be in the future")
	}

	// At least one delegation required
	if len(g.DelegatedOAuth2Tokens) == 0 {
		return errors.New("at least one delegated service is required")
	}

	// Validate each delegation
	for i, token := range g.DelegatedOAuth2Tokens {
		if token.ThirdpartyOAuth2ServiceID == "" {
			return fmt.Errorf("delegation %d: thirdparty_oauth2_service_id is required", i)
		}
		if len(token.Scopes) == 0 {
			return fmt.Errorf("delegation %d: at least one scope is required", i)
		}

		// Validate no duplicate scopes
		scopeSet := make(map[string]bool)
		for _, scope := range token.Scopes {
			if scopeSet[scope] {
				return fmt.Errorf("delegation %d: duplicate scope '%s'", i, scope)
			}
			scopeSet[scope] = true
		}
	}

	return nil
}

// ValidateForCreate validates a grant before creation.
func (g *UserGrant) ValidateForCreate() error {
	if g.Principal == "" {
		return errors.New("principal is required")
	}
	if g.AgentID == "" {
		return errors.New("agent_id is required")
	}

	if g.ValidUntil != nil && g.ValidUntil.Before(time.Now()) {
		return errors.New("valid_until must be in the future")
	}

	if len(g.DelegatedOAuth2Tokens) == 0 {
		return errors.New("at least one delegated service is required")
	}

	for i, token := range g.DelegatedOAuth2Tokens {
		if token.ThirdpartyOAuth2ServiceID == "" {
			return fmt.Errorf("delegation %d: thirdparty_oauth2_service_id is required", i)
		}
		if len(token.Scopes) == 0 {
			return fmt.Errorf("delegation %d: at least one scope is required", i)
		}

		scopeSet := make(map[string]bool)
		for _, scope := range token.Scopes {
			if scopeSet[scope] {
				return fmt.Errorf("delegation %d: duplicate scope '%s'", i, scope)
			}
			scopeSet[scope] = true
		}
	}

	return nil
}

// IsActive returns true if the grant is currently active (not expired).
func (g *UserGrant) IsActive() bool {
	// Indefinite grant (valid_until = null)
	if g.ValidUntil == nil {
		return true
	}
	// Active if not yet expired
	return g.ValidUntil.After(time.Now())
}

// Copy creates a deep copy of the UserGrant.
func (g *UserGrant) Copy() *UserGrant {
	if g == nil {
		return nil
	}

	copy := &UserGrant{
		ID:        g.ID,
		Principal: g.Principal,
		AgentID:   g.AgentID,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}

	if g.ValidUntil != nil {
		validUntil := *g.ValidUntil
		copy.ValidUntil = &validUntil
	}

	// Deep copy delegated tokens
	copy.DelegatedOAuth2Tokens = make([]DelegatedToken, len(g.DelegatedOAuth2Tokens))
	for i, token := range g.DelegatedOAuth2Tokens {
		copy.DelegatedOAuth2Tokens[i] = DelegatedToken{
			ThirdpartyOAuth2ServiceID: token.ThirdpartyOAuth2ServiceID,
			Scopes:                    append([]string{}, token.Scopes...),
		}
	}

	return copy
}

// Scan implements sql.Scanner for JSONB deserialization of delegated tokens.
func (d *DelegatedTokenArray) Scan(value interface{}) error {
	if value == nil {
		*d = []DelegatedToken{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan DelegatedTokenArray: expected []byte")
	}

	var tokens []DelegatedToken
	if err := json.Unmarshal(bytes, &tokens); err != nil {
		return fmt.Errorf("failed to unmarshal DelegatedTokenArray: %w", err)
	}

	*d = tokens
	return nil
}

// Value implements driver.Valuer for JSONB serialization of delegated tokens.
func (d DelegatedTokenArray) Value() (driver.Value, error) {
	if d == nil {
		return json.Marshal([]DelegatedToken{})
	}
	return json.Marshal(d)
}

// DelegatedTokenArray is a custom type for handling JSONB arrays in PostgreSQL.
type DelegatedTokenArray []DelegatedToken
