package principal

import "fmt"

// MissingPrincipalError is returned when a principal is required but missing or empty.
type MissingPrincipalError struct {
	// HeaderName is the HTTP header name where the principal should have been found
	HeaderName string
}

// Error implements the error interface for MissingPrincipalError.
func (e *MissingPrincipalError) Error() string {
	return "missing or empty principal"
}

// InvalidPrincipalError is returned when a principal is present but fails validation rules.
type InvalidPrincipalError struct {
	// Reason describes why the principal is invalid
	Reason string
	// Value is the invalid principal value (truncated for logging safety if necessary)
	Value string
}

// Error implements the error interface for InvalidPrincipalError.
func (e *InvalidPrincipalError) Error() string {
	return e.Reason
}

// NewMissingPrincipalError creates a new MissingPrincipalError.
func NewMissingPrincipalError(headerName string) *MissingPrincipalError {
	return &MissingPrincipalError{HeaderName: headerName}
}

// NewInvalidPrincipalError creates a new InvalidPrincipalError.
func NewInvalidPrincipalError(reason string, value string) *InvalidPrincipalError {
	return &InvalidPrincipalError{Reason: reason, Value: value}
}

// PrincipalTooLongError returns an InvalidPrincipalError for principals exceeding max length.
func PrincipalTooLongError(length int, maxLength int) *InvalidPrincipalError {
	return NewInvalidPrincipalError(
		fmt.Sprintf("principal exceeds maximum length of %d characters (got %d)", maxLength, length),
		"",
	)
}
