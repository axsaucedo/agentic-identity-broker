// Package config contains domain-specific configuration types and errors.
package config

import "fmt"

// ConfigError represents a configuration validation or loading error.
// Compatible with Go 1.13+ error wrapping via Unwrap() method.
type ConfigError struct {
	Field    string      // Configuration field with error (e.g., "log.level")
	Value    interface{} // Invalid value provided (type-flexible, redacted if sensitive)
	Source   string      // Where the value came from (e.g., "cli flag --log-level")
	Expected string      // Expected format or valid values
	Err      error       // Underlying error (for wrapping)
}

// Error constructs the error message dynamically.
// Implements the error interface.
func (e *ConfigError) Error() string {
	var msg string
	if e.Field != "" {
		msg = fmt.Sprintf("configuration error: invalid value %q for field %q", e.Value, e.Field)
	} else {
		msg = "configuration error"
	}

	if e.Expected != "" {
		msg += fmt.Sprintf(" (expected: %s)", e.Expected)
	}

	if e.Source != "" {
		msg += fmt.Sprintf(" [source: %s]", e.Source)
	}

	if e.Err != nil {
		msg += fmt.Sprintf(": %v", e.Err)
	}

	return msg
}

// Unwrap returns the underlying error for error wrapping compatibility.
// Enables use with errors.Is() and errors.As().
func (e *ConfigError) Unwrap() error {
	return e.Err
}
