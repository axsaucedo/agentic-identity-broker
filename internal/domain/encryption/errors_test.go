package encryption

import (
	"errors"
	"fmt"
	"testing"
)

func TestEncryptionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *EncryptionError
		expected string
	}{
		{
			name: "with message",
			err: &EncryptionError{
				Kind:    ErrorKindEncryptionFailed,
				Message: "custom message",
			},
			expected: "custom message",
		},
		{
			name: "without message",
			err: &EncryptionError{
				Kind: ErrorKindDecryptionFailed,
			},
			expected: "decryption_failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("EncryptionError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEncryptionError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := &EncryptionError{
		Kind:    ErrorKindEncryptionFailed,
		Message: "wrapped error",
		Wrapped: originalErr,
	}

	if unwrapped := err.Unwrap(); unwrapped != originalErr {
		t.Errorf("EncryptionError.Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Test that errors.Is works
	if !errors.Is(err, originalErr) {
		t.Errorf("errors.Is(err, originalErr) = false, want true")
	}
}

func TestEncryptionError_Is(t *testing.T) {
	tests := []struct {
		name     string
		err      *EncryptionError
		target   error
		expected bool
	}{
		{
			name: "same kind",
			err: &EncryptionError{
				Kind:    ErrorKindEncryptionFailed,
				Message: "test error",
			},
			target: &EncryptionError{
				Kind: ErrorKindEncryptionFailed,
			},
			expected: true,
		},
		{
			name: "different kind",
			err: &EncryptionError{
				Kind:    ErrorKindEncryptionFailed,
				Message: "test error",
			},
			target: &EncryptionError{
				Kind: ErrorKindDecryptionFailed,
			},
			expected: false,
		},
		{
			name: "non-encryption error",
			err: &EncryptionError{
				Kind:    ErrorKindEncryptionFailed,
				Message: "test error",
			},
			target:   errors.New("other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Is(tt.target); got != tt.expected {
				t.Errorf("EncryptionError.Is() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewEncryptionError(t *testing.T) {
	originalErr := errors.New("original error")
	err := NewEncryptionError(ErrorKindContextMismatch, "test message", originalErr)

	if err.Kind != ErrorKindContextMismatch {
		t.Errorf("Kind = %v, want %v", err.Kind, ErrorKindContextMismatch)
	}
	if err.Message != "test message" {
		t.Errorf("Message = %v, want %v", err.Message, "test message")
	}
	if err.Wrapped != originalErr {
		t.Errorf("Wrapped = %v, want %v", err.Wrapped, originalErr)
	}
}

func TestNewEncryptionFailedError(t *testing.T) {
	originalErr := errors.New("crypto error")
	err := NewEncryptionFailedError("encryption failed", originalErr)

	if err.Kind != ErrorKindEncryptionFailed {
		t.Errorf("Kind = %v, want %v", err.Kind, ErrorKindEncryptionFailed)
	}
	if err.Message != "encryption failed" {
		t.Errorf("Message = %v, want %v", err.Message, "encryption failed")
	}
	if err.Wrapped != originalErr {
		t.Errorf("Wrapped = %v, want %v", err.Wrapped, originalErr)
	}
}

func TestNewDecryptionFailedError(t *testing.T) {
	originalErr := errors.New("crypto error")
	err := NewDecryptionFailedError("decryption failed", originalErr)

	if err.Kind != ErrorKindDecryptionFailed {
		t.Errorf("Kind = %v, want %v", err.Kind, ErrorKindDecryptionFailed)
	}
	if err.Message != "decryption failed" {
		t.Errorf("Message = %v, want %v", err.Message, "decryption failed")
	}
	if err.Wrapped != originalErr {
		t.Errorf("Wrapped = %v, want %v", err.Wrapped, originalErr)
	}
}

func TestNewContextMismatchError(t *testing.T) {
	originalErr := errors.New("context error")
	err := NewContextMismatchError("context mismatch", originalErr)

	if err.Kind != ErrorKindContextMismatch {
		t.Errorf("Kind = %v, want %v", err.Kind, ErrorKindContextMismatch)
	}
	if err.Message != "context mismatch" {
		t.Errorf("Message = %v, want %v", err.Message, "context mismatch")
	}
	if err.Wrapped != originalErr {
		t.Errorf("Wrapped = %v, want %v", err.Wrapped, originalErr)
	}
}

func TestNewIntegrityViolationError(t *testing.T) {
	originalErr := errors.New("integrity error")
	err := NewIntegrityViolationError("integrity violated", originalErr)

	if err.Kind != ErrorKindIntegrityViolation {
		t.Errorf("Kind = %v, want %v", err.Kind, ErrorKindIntegrityViolation)
	}
	if err.Message != "integrity violated" {
		t.Errorf("Message = %v, want %v", err.Message, "integrity violated")
	}
	if err.Wrapped != originalErr {
		t.Errorf("Wrapped = %v, want %v", err.Wrapped, originalErr)
	}
}

func TestNewKEKUnavailableError(t *testing.T) {
	originalErr := errors.New("kms error")
	err := NewKEKUnavailableError("KEK unavailable", originalErr)

	if err.Kind != ErrorKindKEKUnavailable {
		t.Errorf("Kind = %v, want %v", err.Kind, ErrorKindKEKUnavailable)
	}
	if err.Message != "KEK unavailable" {
		t.Errorf("Message = %v, want %v", err.Message, "KEK unavailable")
	}
	if err.Wrapped != originalErr {
		t.Errorf("Wrapped = %v, want %v", err.Wrapped, originalErr)
	}
}

func TestFormatSanitizedError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		context   map[string]string
		expected  string
	}{
		{
			name:      "empty context",
			operation: "encryption failed",
			context:   map[string]string{},
			expected:  "encryption failed",
		},
		{
			name:      "safe context keys",
			operation: "decryption failed",
			context: map[string]string{
				"service_id": "service-123",
				"session_id": "session-456",
				"principal":  "user@example.com",
			},
			expected: "decryption failed with context: map[principal:user@example.com service_id:service-123 session_id:session-456]",
		},
		{
			name:      "mixed safe and unsafe context keys",
			operation: "encryption failed",
			context: map[string]string{
				"service_id": "service-123",
				"secret_key": "super-secret",
				"token":      "oauth-token",
			},
			expected: "encryption failed with context: map[secret_key:[REDACTED] service_id:service-123 token:[REDACTED]]",
		},
		{
			name:      "all unsafe context keys",
			operation: "decryption failed",
			context: map[string]string{
				"secret_key": "super-secret",
				"password":   "secret-password",
			},
			expected: "decryption failed with context: map[password:[REDACTED] secret_key:[REDACTED]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSanitizedError(tt.operation, tt.context)
			if got != tt.expected {
				t.Errorf("FormatSanitizedError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorsAs(t *testing.T) {
	originalErr := errors.New("original")
	err := NewEncryptionFailedError("test", originalErr)

	var encErr *EncryptionError
	if !errors.As(err, &encErr) {
		t.Error("errors.As() failed to extract EncryptionError")
	}

	if encErr.Kind != ErrorKindEncryptionFailed {
		t.Errorf("Extracted error kind = %v, want %v", encErr.Kind, ErrorKindEncryptionFailed)
	}
}

// TestChainedErrors verifies that error chaining works correctly
func TestChainedErrors(t *testing.T) {
	rootErr := errors.New("root cause")
	wrappedErr := fmt.Errorf("wrapped: %w", rootErr)
	encErr := NewEncryptionFailedError("encryption failed", wrappedErr)

	// Test that we can find the root error
	if !errors.Is(encErr, rootErr) {
		t.Error("errors.Is() failed to find root error through chain")
	}

	// Test that we can find the wrapped error
	if !errors.Is(encErr, wrappedErr) {
		t.Error("errors.Is() failed to find wrapped error")
	}
}
