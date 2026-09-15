package admin

import "testing"

func TestRequestsCanonicalReferences(t *testing.T) {
	tests := []struct {
		name   string
		prefer string
		want   bool
	}{
		{name: "absent", want: false},
		{name: "uuid", prefer: "reference-id=uuid", want: false},
		{name: "canonical", prefer: "reference-id=canonical", want: true},
		{name: "unrecognized", prefer: "return=minimal", want: false},
		{name: "mixed", prefer: "return=minimal, reference-id=canonical", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requestsCanonicalReferences(tt.prefer); got != tt.want {
				t.Fatalf("requestsCanonicalReferences(%q) = %t, want %t", tt.prefer, got, tt.want)
			}
		})
	}
}
