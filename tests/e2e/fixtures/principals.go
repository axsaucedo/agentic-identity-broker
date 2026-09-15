// Package fixtures provides test data factories for E2E tests.
package fixtures

import "strings"

// Principal represents a test user with authentication context.
// Used for X-Remote-User header in authenticated requests.
type Principal struct {
	Email string
}

// String returns the principal identifier for use in X-Remote-User header.
// Implements fmt.Stringer interface for easy string conversion.
func (p Principal) String() string {
	return p.Email
}

// DefaultPrincipal returns the default test principal.
// Email: user@example.com
func DefaultPrincipal() Principal {
	return Principal{Email: "user@example.com"}
}

// AnotherPrincipal returns an alternative test principal.
// Email: another-user@example.com
func AnotherPrincipal() Principal {
	return Principal{Email: "another-user@example.com"}
}

// AdminPrincipal returns an admin test principal.
// Email: admin@example.com
func AdminPrincipal() Principal {
	return Principal{Email: "admin@example.com"}
}

// OversizedPrincipal returns a principal that exceeds the maximum allowed length.
// Per specification, principal must not exceed 200 characters.
// This fixture generates 250+ character principal for validation testing.
func OversizedPrincipal() Principal {
	return Principal{Email: strings.Repeat("a", 250) + "@example.com"}
}

// XSSPayload returns a test payload designed to attempt cross-site scripting.
// Used to verify the system safely handles malicious input in query parameters.
func XSSPayload() string {
	return "<script>alert('xss')</script>"
}

// SQLInjectionPayload returns a test payload designed to attempt SQL injection.
// Used to verify the system safely handles malicious input without exposing database details.
func SQLInjectionPayload() string {
	return "1' OR '1'='1"
}
