// Package ports defines interfaces for hexagonal architecture boundaries.
package ports

import (
	"context"
	"time"
)

// StorageLifecycle manages storage backend lifecycle operations.
// Handles initialization, health checking, and cleanup.
// This interface isolates storage lifecycle concerns from CRUD operations.
type StorageLifecycle interface {
	// Initialize performs storage backend initialization and verification.
	// For PostgreSQL: connects to database, verifies schema version
	// For in-memory: initializes empty storage structures
	// Returns error if initialization fails (FR-007).
	// Applications MUST fail startup if Initialize returns an error.
	Initialize(ctx context.Context) error

	// Close gracefully closes storage connections and releases resources.
	// Should be called during application shutdown.
	// Idempotent: safe to call multiple times.
	// Best effort - does not fail startup if Close fails.
	Close(ctx context.Context) error

	// HealthCheck verifies storage backend is operational.
	// Lightweight operation for monitoring (5-20ms typical).
	// Returns nil if healthy, error with details if unhealthy.
	// Used by monitoring and readiness probes.
	HealthCheck(ctx context.Context) error
}

// User represents a user entity in the storage layer.
// Domain entity - does not expose storage implementation details.
type User struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// UserFilter represents optional filtering criteria for ListUsers operations.
// Allows consumers to filter results by common fields.
type UserFilter struct {
	EmailPattern string // SQL LIKE pattern for email filtering (optional)
	Limit        int    // Maximum results to return (0 = no limit)
	Offset       int    // Number of results to skip (pagination)
}

// UserRepository defines storage operations for user entities.
// Follows Repository pattern with focused CRUD operations.
// This interface isolates domain logic from storage implementation.
// Following Interface Segregation Principle: small, focused interfaces.
type UserRepository interface {
	// CreateUser creates a new user entity in storage.
	// Returns error if:
	// - User ID already exists (StorageError with Kind=Conflict)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	CreateUser(ctx context.Context, user *User) error

	// GetUser retrieves a user entity by ID.
	// Returns nil if user not found (checked by caller to distinguish from error).
	// Returns StorageError with Kind=NotFound if user not found.
	// Returns StorageError for connection/timeout issues.
	GetUser(ctx context.Context, id string) (*User, error)

	// UpdateUser updates an existing user entity.
	// Returns error if:
	// - User ID not found (StorageError with Kind=NotFound)
	// - Storage connection fails (StorageError with Kind=Connection)
	// - Operation timeout (StorageError with Kind=Timeout)
	UpdateUser(ctx context.Context, user *User) error

	// DeleteUser deletes a user entity by ID.
	// Returns error if storage operation fails.
	// It is safe to delete non-existent users (idempotent).
	DeleteUser(ctx context.Context, id string) error

	// ListUsers retrieves multiple user entities matching optional filter.
	// Returns empty slice if no users match filter (not an error).
	// Returns StorageError for connection/timeout issues.
	ListUsers(ctx context.Context, filter *UserFilter) ([]*User, error)
}
