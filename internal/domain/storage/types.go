package storage

import (
	"fmt"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

// User represents a user entity stored in the storage layer.
// This is a pure domain entity - implementation-agnostic.
// Database column mappings are in adapters, not here.
type User struct {
	ID        id.UserID `db:"id"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Validate performs validation on User entity.
// Returns error if entity violates domain constraints.
func (u *User) Validate() error {
	if u.ID.IsZero() {
		return fmt.Errorf("user ID cannot be empty")
	}
	if u.Email == "" {
		return fmt.Errorf("user email cannot be empty")
	}
	// Basic email format check (not comprehensive validation)
	if len(u.Email) < 3 || len(u.Email) > 254 {
		return fmt.Errorf("user email must be between 3 and 254 characters")
	}
	if u.CreatedAt.IsZero() {
		return fmt.Errorf("user created_at cannot be zero time")
	}
	if u.UpdatedAt.IsZero() {
		return fmt.Errorf("user updated_at cannot be zero time")
	}
	return nil
}
