// Package storage defines domain types for the storage layer.
// All types are technology-agnostic and represent pure domain concepts.
// Implementation-specific details (SQL, database drivers) are isolated in adapters.
package storage

// StorageBackend represents the active storage backend at runtime.
// Enumeration of supported storage backends.
type StorageBackend string

const (
	// BackendMemory: In-memory storage using Go maps
	// Use case: Development, testing, zero-configuration
	// Characteristics: Instant initialization, no persistence, thread-safe with sync.RWMutex
	BackendMemory StorageBackend = "memory"

	// BackendPostgres: PostgreSQL database
	// Use case: Production workloads
	// Characteristics: Durable, connection pooling, schema versioning, TLS support
	BackendPostgres StorageBackend = "postgres"
)
