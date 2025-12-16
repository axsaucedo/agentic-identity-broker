package storage

import (
	"strings"
	"testing"
)

// TestRedactedConnection_PasswordMasking verifies passwords are redacted in connection URLs
func TestRedactedConnection_PasswordMasking(t *testing.T) {
	tests := []struct {
		name           string
		connURL        string
		shouldExist    []string
		shouldNotExist []string
	}{
		{
			name:    "postgresql URL with password",
			connURL: "postgresql://user:secretpass@localhost:5432/mydb",
			shouldExist: []string{
				"postgresql://",
				"user@localhost:5432",
				"/mydb",
			},
			shouldNotExist: []string{
				"secretpass",
			},
		},
		{
			name:    "postgres URL with password",
			connURL: "postgres://admin:MySecurePassword123@db.example.com:5432/prod_db",
			shouldExist: []string{
				"postgres://",
				"admin@db.example.com",
				"/prod_db",
			},
			shouldNotExist: []string{
				"MySecurePassword123",
			},
		},
		{
			name:    "URL with SSL certificate in query parameters",
			connURL: "postgresql://user:pass@localhost/db?sslcert=/path/to/cert.pem&sslkey=/path/to/key.pem",
			shouldExist: []string{
				"postgresql://",
				"user@localhost",
				"/db",
				"sslkey",
			},
			shouldNotExist: []string{
				"pass",
				"/path/to/cert.pem",
			},
		},
		{
			name:    "URL with sslmode parameter",
			connURL: "postgresql://user:password@host:5432/db?sslmode=require",
			shouldExist: []string{
				"sslmode=require",
			},
			shouldNotExist: []string{
				"password",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := &ConnectionParameters{ConnectionURL: tt.connURL}
			redacted := cp.Redacted()

			for _, str := range tt.shouldExist {
				if !strings.Contains(redacted, str) {
					t.Errorf("Redacted() should contain %q, got: %s", str, redacted)
				}
			}

			for _, str := range tt.shouldNotExist {
				if strings.Contains(redacted, str) {
					t.Errorf("Redacted() should NOT contain %q, got: %s", str, redacted)
				}
			}
		})
	}
}

// TestRedactedConnection_NoCredentials verifies URLs without credentials remain unchanged
func TestRedactedConnection_NoCredentials(t *testing.T) {
	tests := []struct {
		name    string
		connURL string
	}{
		{
			name:    "URL without user credentials",
			connURL: "postgresql://localhost:5432/mydb",
		},
		{
			name:    "URL with only username",
			connURL: "postgresql://user@localhost:5432/mydb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := &ConnectionParameters{ConnectionURL: tt.connURL}
			redacted := cp.Redacted()

			// Should be unchanged or very similar
			if !strings.Contains(redacted, "localhost") {
				t.Errorf("Redacted() should preserve localhost, got: %s", redacted)
			}
		})
	}
}

// TestConnectionValidation_SSLModeParameter verifies SSL mode is validated
func TestConnectionValidation_SSLModeParameter(t *testing.T) {
	tests := []struct {
		name    string
		connURL string
		valid   bool
	}{
		{
			name:    "valid with sslmode=require",
			connURL: "postgresql://user@localhost/db?sslmode=require",
			valid:   true,
		},
		{
			name:    "valid with sslmode=disable",
			connURL: "postgresql://user@localhost/db?sslmode=disable",
			valid:   true,
		},
		{
			name:    "valid with sslmode=verify-full",
			connURL: "postgresql://user@localhost/db?sslmode=verify-full",
			valid:   true,
		},
		{
			name:    "valid with connect_timeout",
			connURL: "postgresql://user@localhost/db?connect_timeout=10",
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := ValidatePostgresURL(tt.connURL)
			if valid != tt.valid {
				t.Errorf("ValidatePostgresURL(%q) = %v, want %v", tt.connURL, valid, tt.valid)
			}
		})
	}
}

// TestConnectionValidation_InvalidSchemes verifies invalid schemes are rejected
func TestConnectionValidation_InvalidSchemes(t *testing.T) {
	tests := []struct {
		name    string
		connURL string
	}{
		{
			name:    "mysql scheme",
			connURL: "mysql://user@localhost/db",
		},
		{
			name:    "mongodb scheme",
			connURL: "mongodb://user@localhost/db",
		},
		{
			name:    "http scheme",
			connURL: "http://user@localhost/db",
		},
		{
			name:    "no scheme",
			connURL: "user@localhost/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := ValidatePostgresURL(tt.connURL)
			if valid {
				t.Errorf("ValidatePostgresURL(%q) should reject %q, got: %v", tt.name, tt.connURL, valid)
			}
		})
	}
}

// TestConnectionParameters_Validate verifies connection parameter validation
func TestConnectionParameters_Validate(t *testing.T) {
	tests := []struct {
		name      string
		connURL   string
		wantError bool
		errorType error
	}{
		{
			name:      "valid postgresql URL",
			connURL:   "postgresql://user:pass@localhost:5432/mydb",
			wantError: false,
		},
		{
			name:      "valid postgres URL",
			connURL:   "postgres://user:pass@localhost:5432/mydb",
			wantError: false,
		},
		{
			name:      "empty connection URL",
			connURL:   "",
			wantError: true,
			errorType: ErrConnectionURLEmpty,
		},
		{
			name:      "invalid scheme",
			connURL:   "mysql://user@localhost/db",
			wantError: true,
			errorType: ErrConnectionURLInvalidScheme,
		},
		{
			name:      "missing host",
			connURL:   "postgresql://user:pass@/mydb",
			wantError: true,
			errorType: ErrConnectionURLNoHost,
		},
		{
			name:      "missing database",
			connURL:   "postgresql://user:pass@localhost:5432",
			wantError: true,
			errorType: ErrConnectionURLNoDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := &ConnectionParameters{ConnectionURL: tt.connURL}
			err := cp.Validate()

			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.wantError && tt.errorType != nil {
				if err != tt.errorType {
					t.Errorf("Validate() error = %v, want %v", err, tt.errorType)
				}
			}
		})
	}
}

// TestConnectionURLWithSpecialCharacters verifies special characters in credentials are handled
func TestConnectionURLWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name    string
		connURL string
	}{
		{
			name:    "password with URL-encoded special characters",
			connURL: "postgresql://user:p%40ssw0rd@localhost/db",
		},
		{
			name:    "password with special characters like @ # $",
			connURL: "postgresql://user:p%40%23%24@localhost/db",
		},
		{
			name:    "username with special characters",
			connURL: "postgresql://user%40domain:pass@localhost/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := &ConnectionParameters{ConnectionURL: tt.connURL}
			redacted := cp.Redacted()

			// Should not contain the original URL unmodified
			if redacted == tt.connURL {
				t.Errorf("Redacted() should redact credentials, got unchanged: %s", redacted)
			}

			// Should be valid after redaction
			if !strings.Contains(redacted, "localhost") {
				t.Errorf("Redacted() should preserve host, got: %s", redacted)
			}
		})
	}
}

// TestConnectionParametersConsistency verifies redaction consistency
func TestConnectionParametersConsistency(t *testing.T) {
	connURL := "postgresql://admin:secretpass@db.example.com:5432/production"
	cp := &ConnectionParameters{ConnectionURL: connURL}

	// Call Redacted() multiple times - should return consistent result
	redacted1 := cp.Redacted()
	redacted2 := cp.Redacted()

	if redacted1 != redacted2 {
		t.Errorf("Redacted() should be consistent: %s vs %s", redacted1, redacted2)
	}

	// Original should be unchanged
	if cp.ConnectionURL != connURL {
		t.Errorf("ConnectionURL should not be modified: %s", cp.ConnectionURL)
	}
}
