package urivalidation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeResourceURI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "already normalized",
			input: "https://api.example.com",
			want:  "https://api.example.com",
		},
		{
			name:  "single trailing slash",
			input: "https://api.example.com/",
			want:  "https://api.example.com",
		},
		{
			name:  "multiple trailing slashes",
			input: "https://api.example.com///",
			want:  "https://api.example.com",
		},
		{
			name:  "path trailing slashes",
			input: "https://api.example.com/v1///",
			want:  "https://api.example.com/v1",
		},
		{
			name:  "query trailing slash preserved",
			input: "https://api.example.com/?target=/",
			want:  "https://api.example.com?target=/",
		},
		{
			name:  "fragment trailing slash preserved",
			input: "https://api.example.com/path/#section/",
			want:  "https://api.example.com/path#section/",
		},
		{
			name:  "escaped path slash preserved",
			input: "https://api.example.com/%2F/",
			want:  "https://api.example.com/%2F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NormalizeResourceURI(tt.input)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, got, NormalizeResourceURI(got))
		})
	}
}
