package tokenexchange

import (
	"testing"
)

func TestNewResourceURI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		input       string
		expectError bool
		expectValue string
	}{
		{
			name:        "valid HTTPS URI",
			input:       "https://api.github.com",
			expectError: false,
			expectValue: "https://api.github.com",
		},
		{
			name:        "valid HTTPS URI with path",
			input:       "https://api.github.com/v3",
			expectError: false,
			expectValue: "https://api.github.com/v3",
		},
		{
			name:        "valid HTTP URI localhost",
			input:       "http://localhost:8080",
			expectError: false,
			expectValue: "http://localhost:8080",
		},
		{
			name:        "remove trailing slash",
			input:       "https://api.github.com/",
			expectError: false,
			expectValue: "https://api.github.com",
		},
		{
			name:        "remove multiple trailing slashes (only one removed)",
			input:       "https://api.github.com//",
			expectError: false,
			expectValue: "https://api.github.com/",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "no scheme",
			input:       "api.github.com",
			expectError: true,
		},
		{
			name:        "invalid URL",
			input:       "not a valid uri",
			expectError: true,
		},
		{
			name:        "scheme only",
			input:       "https://",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uri, err := NewResourceURI(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if uri.Value() != tt.expectValue {
				t.Errorf("expected %q, got %q", tt.expectValue, uri.Value())
			}
		})
	}
}

func TestResourceURIEqual(t *testing.T) {
	t.Parallel()
	uri1 := &ResourceURI{value: "https://api.github.com"}
	uri2 := &ResourceURI{value: "https://api.github.com"}
	uri3 := &ResourceURI{value: "https://api.example.com"}

	tests := []struct {
		name     string
		uri1     *ResourceURI
		uri2     *ResourceURI
		expected bool
	}{
		{
			name:     "same value",
			uri1:     uri1,
			uri2:     uri2,
			expected: true,
		},
		{
			name:     "different value",
			uri1:     uri1,
			uri2:     uri3,
			expected: false,
		},
		{
			name:     "both nil",
			uri1:     nil,
			uri2:     nil,
			expected: true,
		},
		{
			name:     "one nil",
			uri1:     uri1,
			uri2:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.uri1.Equal(tt.uri2)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestResourceURIString(t *testing.T) {
	t.Parallel()
	uri, _ := NewResourceURI("https://api.github.com")
	if uri.String() != "https://api.github.com" {
		t.Errorf("String() failed: expected %q, got %q", "https://api.github.com", uri.String())
	}

	var nilURI *ResourceURI
	if nilURI.String() != "" {
		t.Errorf("String() on nil: expected empty string, got %q", nilURI.String())
	}
}

func TestNormalize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URI without trailing slash",
			input:    "https://api.github.com",
			expected: "https://api.github.com",
		},
		{
			name:     "URI with trailing slash",
			input:    "https://api.github.com/",
			expected: "https://api.github.com",
		},
		{
			name:     "URI with path and trailing slash",
			input:    "https://api.github.com/v3/",
			expected: "https://api.github.com/v3",
		},
		{
			name:     "URI with path but no trailing slash",
			input:    "https://api.github.com/v3",
			expected: "https://api.github.com/v3",
		},
		{
			name:     "URI with port and trailing slash",
			input:    "http://localhost:8080/",
			expected: "http://localhost:8080",
		},
		{
			name:     "URI with port but no trailing slash",
			input:    "http://localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "idempotent - normalizing twice yields same result",
			input:    "https://api.example.com/",
			expected: "https://api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Normalize(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}

			// Verify idempotency - normalizing twice should yield same result
			result2 := Normalize(result)
			if result != result2 {
				t.Errorf("idempotency failed: first call %q, second call %q", result, result2)
			}
		})
	}
}
