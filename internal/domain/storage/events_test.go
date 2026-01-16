package storage

import (
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
)

func TestNewSessionEncrypted(t *testing.T) {
	sessionID := "session-123"
	serviceID := "service-456"

	event := NewSessionEncrypted(sessionID, serviceID)

	if event.EventType() != "session_encrypted" {
		t.Errorf("EventType() = %v, want %v", event.EventType(), "session_encrypted")
	}
	if event.SessionID() != sessionID {
		t.Errorf("SessionID() = %v, want %v", event.SessionID(), sessionID)
	}
	if event.ServiceID != serviceID {
		t.Errorf("ServiceID = %v, want %v", event.ServiceID, serviceID)
	}
	if event.Timestamp().IsZero() {
		t.Error("Timestamp() should not be zero")
	}
	if time.Since(event.Timestamp()) > time.Second {
		t.Error("Timestamp() should be recent")
	}
}

func TestNewSessionDecrypted(t *testing.T) {
	sessionID := "session-123"
	serviceID := "service-456"

	event := NewSessionDecrypted(sessionID, serviceID)

	if event.EventType() != "session_decrypted" {
		t.Errorf("EventType() = %v, want %v", event.EventType(), "session_decrypted")
	}
	if event.SessionID() != sessionID {
		t.Errorf("SessionID() = %v, want %v", event.SessionID(), sessionID)
	}
	if event.ServiceID != serviceID {
		t.Errorf("ServiceID = %v, want %v", event.ServiceID, serviceID)
	}
	if event.Timestamp().IsZero() {
		t.Error("Timestamp() should not be zero")
	}
	if time.Since(event.Timestamp()) > time.Second {
		t.Error("Timestamp() should be recent")
	}
}

func TestNewSessionEncryptionFailed(t *testing.T) {
	sessionID := "session-123"
	serviceID := "service-456"
	errorKind := encryption.ErrorKindKEKUnavailable
	message := "KEK unavailable"

	event := NewSessionEncryptionFailed(sessionID, serviceID, errorKind, message)

	if event.EventType() != "session_encryption_failed" {
		t.Errorf("EventType() = %v, want %v", event.EventType(), "session_encryption_failed")
	}
	if event.SessionID() != sessionID {
		t.Errorf("SessionID() = %v, want %v", event.SessionID(), sessionID)
	}
	if event.ServiceID != serviceID {
		t.Errorf("ServiceID = %v, want %v", event.ServiceID, serviceID)
	}
	if event.ErrorKind != errorKind {
		t.Errorf("ErrorKind = %v, want %v", event.ErrorKind, errorKind)
	}
	if event.Message != message {
		t.Errorf("Message = %v, want %v", event.Message, message)
	}
	if event.Timestamp().IsZero() {
		t.Error("Timestamp() should not be zero")
	}
}

func TestNewSessionDecryptionFailed(t *testing.T) {
	sessionID := "session-123"
	serviceID := "service-456"
	errorKind := encryption.ErrorKindContextMismatch
	message := "context mismatch"

	event := NewSessionDecryptionFailed(sessionID, serviceID, errorKind, message)

	if event.EventType() != "session_decryption_failed" {
		t.Errorf("EventType() = %v, want %v", event.EventType(), "session_decryption_failed")
	}
	if event.SessionID() != sessionID {
		t.Errorf("SessionID() = %v, want %v", event.SessionID(), sessionID)
	}
	if event.ServiceID != serviceID {
		t.Errorf("ServiceID = %v, want %v", event.ServiceID, serviceID)
	}
	if event.ErrorKind != errorKind {
		t.Errorf("ErrorKind = %v, want %v", event.ErrorKind, errorKind)
	}
	if event.Message != message {
		t.Errorf("Message = %v, want %v", event.Message, message)
	}
	if event.Timestamp().IsZero() {
		t.Error("Timestamp() should not be zero")
	}
}

func TestBaseEvent(t *testing.T) {
	eventType := "test_event"
	sessionID := "test-session"
	timestamp := time.Now().UTC()

	base := BaseEvent{
		eventType: eventType,
		timestamp: timestamp,
		sessionID: sessionID,
	}

	if base.EventType() != eventType {
		t.Errorf("EventType() = %v, want %v", base.EventType(), eventType)
	}
	if base.SessionID() != sessionID {
		t.Errorf("SessionID() = %v, want %v", base.SessionID(), sessionID)
	}
	if base.Timestamp() != timestamp {
		t.Errorf("Timestamp() = %v, want %v", base.Timestamp(), timestamp)
	}
}

func TestEventInterface(t *testing.T) {
	// Test that all event types implement the Event interface
	var events []Event = []Event{
		NewSessionEncrypted("session-1", "service-1"),
		NewSessionDecrypted("session-2", "service-2"),
		NewSessionEncryptionFailed("session-3", "service-3", encryption.ErrorKindEncryptionFailed, "encryption failed"),
		NewSessionDecryptionFailed("session-4", "service-4", encryption.ErrorKindDecryptionFailed, "decryption failed"),
	}

	for i, event := range events {
		if event.EventType() == "" {
			t.Errorf("Event %d: EventType() should not be empty", i)
		}
		if event.Timestamp().IsZero() {
			t.Errorf("Event %d: Timestamp() should not be zero", i)
		}
		if event.SessionID() == "" {
			t.Errorf("Event %d: SessionID() should not be empty", i)
		}
	}
}

func TestNoOpEventPublisher(t *testing.T) {
	publisher := NewNoOpEventPublisher()

	// Test that Publish doesn't return an error and handles any event type
	events := []Event{
		NewSessionEncrypted("session-1", "service-1"),
		NewSessionDecrypted("session-2", "service-2"),
		NewSessionEncryptionFailed("session-3", "service-3", encryption.ErrorKindEncryptionFailed, "test"),
		NewSessionDecryptionFailed("session-4", "service-4", encryption.ErrorKindDecryptionFailed, "test"),
	}

	for i, event := range events {
		if err := publisher.Publish(event); err != nil {
			t.Errorf("Event %d: Publish() returned error: %v", i, err)
		}
	}
}

func TestEventPublisherInterface(t *testing.T) {
	// Test that NoOpEventPublisher implements EventPublisher interface
	var publisher EventPublisher = NewNoOpEventPublisher()
	event := NewSessionEncrypted("session-1", "service-1")

	if err := publisher.Publish(event); err != nil {
		t.Errorf("Publish() returned error: %v", err)
	}
}

// TestEventImmutability verifies that events are immutable after creation
func TestEventImmutability(t *testing.T) {
	event := NewSessionEncrypted("session-1", "service-1")
	originalTimestamp := event.Timestamp()
	originalEventType := event.EventType()
	originalSessionID := event.SessionID()
	originalServiceID := event.ServiceID

	// Wait a bit to ensure time has passed
	time.Sleep(time.Millisecond)

	// Verify that getters still return the same values
	if event.Timestamp() != originalTimestamp {
		t.Error("Event timestamp changed - events should be immutable")
	}
	if event.EventType() != originalEventType {
		t.Error("Event type changed - events should be immutable")
	}
	if event.SessionID() != originalSessionID {
		t.Error("Event session ID changed - events should be immutable")
	}
	if event.ServiceID != originalServiceID {
		t.Error("Event service ID changed - events should be immutable")
	}
}

// TestErrorKindIntegration verifies encryption error kinds work with events
func TestErrorKindIntegration(t *testing.T) {
	testCases := []struct {
		name      string
		errorKind encryption.ErrorKind
	}{
		{"encryption_failed", encryption.ErrorKindEncryptionFailed},
		{"decryption_failed", encryption.ErrorKindDecryptionFailed},
		{"context_mismatch", encryption.ErrorKindContextMismatch},
		{"integrity_violation", encryption.ErrorKindIntegrityViolation},
		{"kek_unavailable", encryption.ErrorKindKEKUnavailable},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encEvent := NewSessionEncryptionFailed("session-1", "service-1", tc.errorKind, "test message")
			if encEvent.ErrorKind != tc.errorKind {
				t.Errorf("SessionEncryptionFailed.ErrorKind = %v, want %v", encEvent.ErrorKind, tc.errorKind)
			}

			decEvent := NewSessionDecryptionFailed("session-2", "service-2", tc.errorKind, "test message")
			if decEvent.ErrorKind != tc.errorKind {
				t.Errorf("SessionDecryptionFailed.ErrorKind = %v, want %v", decEvent.ErrorKind, tc.errorKind)
			}
		})
	}
}