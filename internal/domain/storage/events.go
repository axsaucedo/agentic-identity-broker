package storage

import (
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
)

// Event represents a domain event that occurred during system operations.
// Events are immutable and should not contain sensitive data like plaintext tokens or keys.
type Event interface {
	// EventType returns the type identifier for this event
	EventType() string
	// Timestamp returns when the event occurred
	Timestamp() time.Time
	// SessionID returns the session ID associated with this event (if applicable)
	SessionID() string
}

// BaseEvent provides common fields for all events.
type BaseEvent struct {
	eventType string
	timestamp time.Time
	sessionID string
}

// EventType implements Event interface
func (e BaseEvent) EventType() string {
	return e.eventType
}

// Timestamp implements Event interface
func (e BaseEvent) Timestamp() time.Time {
	return e.timestamp
}

// SessionID implements Event interface
func (e BaseEvent) SessionID() string {
	return e.sessionID
}

// SessionEncrypted represents a successful session encryption event.
// Emitted when OAuth2 tokens are successfully encrypted for storage.
type SessionEncrypted struct {
	BaseEvent
	ServiceID string `json:"service_id"`
}

// NewSessionEncrypted creates a new SessionEncrypted event.
func NewSessionEncrypted(sessionID, serviceID string) *SessionEncrypted {
	return &SessionEncrypted{
		BaseEvent: BaseEvent{
			eventType: "session_encrypted",
			timestamp: time.Now().UTC(),
			sessionID: sessionID,
		},
		ServiceID: serviceID,
	}
}

// SessionDecrypted represents a successful session decryption event.
// Emitted when encrypted OAuth2 tokens are successfully decrypted for use.
type SessionDecrypted struct {
	BaseEvent
	ServiceID string `json:"service_id"`
}

// NewSessionDecrypted creates a new SessionDecrypted event.
func NewSessionDecrypted(sessionID, serviceID string) *SessionDecrypted {
	return &SessionDecrypted{
		BaseEvent: BaseEvent{
			eventType: "session_decrypted",
			timestamp: time.Now().UTC(),
			sessionID: sessionID,
		},
		ServiceID: serviceID,
	}
}

// SessionEncryptionFailed represents a failed session encryption event.
// Emitted when OAuth2 token encryption fails for any reason.
type SessionEncryptionFailed struct {
	BaseEvent
	ServiceID string               `json:"service_id"`
	ErrorKind encryption.ErrorKind `json:"error_kind"`
	Message   string               `json:"message"` // Sanitized message safe for logging
}

// NewSessionEncryptionFailed creates a new SessionEncryptionFailed event.
func NewSessionEncryptionFailed(sessionID, serviceID string, errorKind encryption.ErrorKind, sanitizedMessage string) *SessionEncryptionFailed {
	return &SessionEncryptionFailed{
		BaseEvent: BaseEvent{
			eventType: "session_encryption_failed",
			timestamp: time.Now().UTC(),
			sessionID: sessionID,
		},
		ServiceID: serviceID,
		ErrorKind: errorKind,
		Message:   sanitizedMessage,
	}
}

// SessionDecryptionFailed represents a failed session decryption event.
// Emitted when encrypted OAuth2 token decryption fails for any reason.
type SessionDecryptionFailed struct {
	BaseEvent
	ServiceID string               `json:"service_id"`
	ErrorKind encryption.ErrorKind `json:"error_kind"`
	Message   string               `json:"message"` // Sanitized message safe for logging
}

// NewSessionDecryptionFailed creates a new SessionDecryptionFailed event.
func NewSessionDecryptionFailed(sessionID, serviceID string, errorKind encryption.ErrorKind, sanitizedMessage string) *SessionDecryptionFailed {
	return &SessionDecryptionFailed{
		BaseEvent: BaseEvent{
			eventType: "session_decryption_failed",
			timestamp: time.Now().UTC(),
			sessionID: sessionID,
		},
		ServiceID: serviceID,
		ErrorKind: errorKind,
		Message:   sanitizedMessage,
	}
}

// EventPublisher defines the interface for publishing domain events.
// This allows the domain to emit events without depending on specific event bus implementations.
type EventPublisher interface {
	// Publish emits a domain event for processing by event handlers.
	// Events should be processed asynchronously and failures should not affect the main operation.
	Publish(event Event) error
}

// NoOpEventPublisher is a default implementation that discards all events.
// Used for testing or when event publishing is disabled.
type NoOpEventPublisher struct{}

// Publish implements EventPublisher interface by doing nothing.
func (p *NoOpEventPublisher) Publish(event Event) error {
	// Silently discard the event
	return nil
}

// NewNoOpEventPublisher creates a new NoOpEventPublisher instance.
func NewNoOpEventPublisher() *NoOpEventPublisher {
	return &NoOpEventPublisher{}
}
