package encryption

import (
	"errors"
	"fmt"
)

// ErrorKind represents the type of encryption error that occurred.
// These kinds help classify errors for monitoring and handling decisions.
type ErrorKind string

const (
	// ErrorKindEncryptionFailed indicates that encryption operation failed.
	// This could be due to key unavailability, invalid plaintext, or internal crypto errors.
	ErrorKindEncryptionFailed ErrorKind = "encryption_failed"

	// ErrorKindDecryptionFailed indicates that decryption operation failed.
	// This could be due to corrupted ciphertext, key unavailability, or internal crypto errors.
	ErrorKindDecryptionFailed ErrorKind = "decryption_failed"

	// ErrorKindContextMismatch indicates that the encryption context provided
	// for decryption does not match the context used during encryption.
	// This prevents cross-service or cross-session token usage.
	ErrorKindContextMismatch ErrorKind = "context_mismatch"

	// ErrorKindIntegrityViolation indicates that ciphertext integrity check failed.
	// This suggests tampering, corruption, or using wrong decryption context.
	ErrorKindIntegrityViolation ErrorKind = "integrity_violation"

	// ErrorKindKEKUnavailable indicates that the Key Encryption Key (KEK) is unavailable.
	// This could be due to AWS KMS issues, network problems, or permission errors.
	ErrorKindKEKUnavailable ErrorKind = "kek_unavailable"
)

// EncryptionError represents an error that occurred during encryption or decryption operations.
// Error messages are sanitized to prevent leaking sensitive data like key material or token contents.
type EncryptionError struct {
	// Kind categorizes the type of error for handling and monitoring
	Kind ErrorKind

	// Message contains a sanitized error description safe for logging and user display.
	// SECURITY: Never includes key material, plaintext tokens, or other sensitive data.
	Message string

	// Wrapped contains the underlying error if available.
	// May contain implementation-specific details for debugging.
	Wrapped error
}

// Error implements the error interface.
// Returns the sanitized error message.
func (e *EncryptionError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.Kind)
}

// Unwrap implements the error unwrapping interface.
// Allows errors.Is() and errors.As() to work with wrapped errors.
func (e *EncryptionError) Unwrap() error {
	return e.Wrapped
}

// Is implements the error identity checking interface.
// Allows errors.Is(err, &EncryptionError{Kind: ErrorKindEncryptionFailed}) to work.
func (e *EncryptionError) Is(target error) bool {
	var encErr *EncryptionError
	if errors.As(target, &encErr) {
		return e.Kind == encErr.Kind
	}
	return false
}

// NewEncryptionError creates a new EncryptionError with the specified kind and message.
// The message should be sanitized and safe for logging.
func NewEncryptionError(kind ErrorKind, message string, wrapped error) *EncryptionError {
	return &EncryptionError{
		Kind:    kind,
		Message: message,
		Wrapped: wrapped,
	}
}

// NewEncryptionFailedError creates an EncryptionError for encryption failures.
func NewEncryptionFailedError(message string, wrapped error) *EncryptionError {
	return NewEncryptionError(ErrorKindEncryptionFailed, message, wrapped)
}

// NewDecryptionFailedError creates an EncryptionError for decryption failures.
func NewDecryptionFailedError(message string, wrapped error) *EncryptionError {
	return NewEncryptionError(ErrorKindDecryptionFailed, message, wrapped)
}

// NewContextMismatchError creates an EncryptionError for context mismatches.
func NewContextMismatchError(message string, wrapped error) *EncryptionError {
	return NewEncryptionError(ErrorKindContextMismatch, message, wrapped)
}

// NewIntegrityViolationError creates an EncryptionError for integrity violations.
func NewIntegrityViolationError(message string, wrapped error) *EncryptionError {
	return NewEncryptionError(ErrorKindIntegrityViolation, message, wrapped)
}

// NewKEKUnavailableError creates an EncryptionError for KEK unavailability.
func NewKEKUnavailableError(message string, wrapped error) *EncryptionError {
	return NewEncryptionError(ErrorKindKEKUnavailable, message, wrapped)
}

// FormatSanitizedError creates a sanitized error message that's safe for logging.
// Removes any potentially sensitive data from the message.
func FormatSanitizedError(operation string, context map[string]string) string {
	// Only include non-sensitive context keys in error messages
	safeContext := make(map[string]string)
	for key, value := range context {
		switch key {
		case "service_id", "session_id", "principal":
			// These are identifiers, not secret data
			safeContext[key] = value
		default:
			// Redact unknown context keys to be safe
			safeContext[key] = "[REDACTED]"
		}
	}

	if len(safeContext) > 0 {
		return fmt.Sprintf("%s with context: %v", operation, safeContext)
	}
	return operation
}
