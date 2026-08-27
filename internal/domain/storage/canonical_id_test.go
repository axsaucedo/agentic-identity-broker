package storage

import (
	"strings"
	"testing"
)

func TestValidateCanonicalID(t *testing.T) {
	tests := []struct {
		name string
		id   *string
		want bool
	}{
		{name: "absent", id: nil, want: true},
		{name: "empty", id: stringPtr(""), want: false},
		{name: "uuid", id: stringPtr("550e8400-e29b-41d4-a716-446655440000"), want: false},
		{name: "invalid character", id: stringPtr("contains/slash"), want: false},
		{name: "maximum length", id: stringPtr(strings.Repeat("a", 128)), want: true},
		{name: "over maximum length", id: stringPtr(strings.Repeat("a", 129)), want: false},
		{name: "valid token", id: stringPtr("Research_Agent.1-Prod"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCanonicalID(tt.id)
			if tt.want && err != nil {
				t.Fatalf("ValidateCanonicalID() error = %v", err)
			}
			if !tt.want && err == nil {
				t.Fatal("ValidateCanonicalID() error = nil")
			}
		})
	}
}
