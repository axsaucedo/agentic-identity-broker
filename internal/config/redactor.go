// Package config implements the configuration loading adapter.
package config

import "strings"

const (
	// SensitivePrefix is the prefix used to identify sensitive configuration keys.
	SensitivePrefix = "IDENTITY_BROKER_"

	// RedactedValue is the placeholder shown for sensitive values.
	RedactedValue = "***REDACTED***"
)

// Redact replaces sensitive values with a redacted placeholder.
// A key is considered sensitive if it starts with IDENTITY_BROKER_ prefix.
// This prevents sensitive data from appearing in logs and startup summaries.
func Redact(key string, value interface{}) interface{} {
	// Convert key to uppercase for case-insensitive comparison
	upperKey := strings.ToUpper(key)

	// Check for sensitive prefix
	if strings.HasPrefix(upperKey, SensitivePrefix) {
		return RedactedValue
	}

	// Check for common sensitive keywords
	sensitiveKeywords := []string{
		"PASSWORD",
		"SECRET",
		"TOKEN",
		"KEY",
		"CREDENTIAL",
		"AUTH",
	}

	for _, keyword := range sensitiveKeywords {
		if strings.Contains(upperKey, keyword) {
			return RedactedValue
		}
	}

	return value
}

// IsSensitive returns true if the key should be treated as sensitive.
func IsSensitive(key string) bool {
	upperKey := strings.ToUpper(key)

	if strings.HasPrefix(upperKey, SensitivePrefix) {
		return true
	}

	sensitiveKeywords := []string{
		"PASSWORD",
		"SECRET",
		"TOKEN",
		"KEY",
		"CREDENTIAL",
		"AUTH",
	}

	for _, keyword := range sensitiveKeywords {
		if strings.Contains(upperKey, keyword) {
			return true
		}
	}

	return false
}
