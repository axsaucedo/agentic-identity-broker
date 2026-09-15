package storage

import (
	"net/url"
	"strings"
)

// ConnectionParameters encapsulates PostgreSQL-specific connection details.
// Contains sensitive data (credentials) that must be redacted in logs.
type ConnectionParameters struct {
	ConnectionURL string
}

// Redacted returns a safe version of the connection URL for logging.
// Masks password and sensitive query parameters.
func (cp *ConnectionParameters) Redacted() string {
	u, err := url.Parse(cp.ConnectionURL)
	if err != nil {
		return "invalid connection URL"
	}

	// Redact password
	if u.User != nil {
		if _, ok := u.User.Password(); ok {
			u.User = url.User(u.User.Username())
		}
	}

	// Redact sensitive query parameters
	q := u.Query()
	for key := range q {
		if strings.EqualFold(key, "password") || strings.EqualFold(key, "sslcert") {
			q.Set(key, "***")
		}
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// Validate performs validation on connection parameters.
// Returns error if connection URL is invalid.
func (cp *ConnectionParameters) Validate() error {
	if cp.ConnectionURL == "" {
		return ErrConnectionURLEmpty
	}

	// Check scheme
	if !strings.HasPrefix(cp.ConnectionURL, "postgresql://") && !strings.HasPrefix(cp.ConnectionURL, "postgres://") {
		return ErrConnectionURLInvalidScheme
	}

	// Parse and validate URL
	u, err := url.Parse(cp.ConnectionURL)
	if err != nil {
		return ErrConnectionURLInvalidFormat
	}

	// Verify host is present
	if u.Host == "" {
		return ErrConnectionURLNoHost
	}

	// Verify database name is present (path component)
	if u.Path == "" || u.Path == "/" {
		return ErrConnectionURLNoDatabase
	}

	return nil
}

// Custom validation errors for connection parameters
var (
	ErrConnectionURLEmpty         = NewStorageError("ValidateConnection", ErrorKindValidation, nil, "connection URL cannot be empty")
	ErrConnectionURLInvalidScheme = NewStorageError("ValidateConnection", ErrorKindValidation, nil, "connection URL must use postgresql:// or postgres:// scheme")
	ErrConnectionURLInvalidFormat = NewStorageError("ValidateConnection", ErrorKindValidation, nil, "connection URL format is invalid")
	ErrConnectionURLNoHost        = NewStorageError("ValidateConnection", ErrorKindValidation, nil, "connection URL must include host")
	ErrConnectionURLNoDatabase    = NewStorageError("ValidateConnection", ErrorKindValidation, nil, "connection URL must include database name")
)
