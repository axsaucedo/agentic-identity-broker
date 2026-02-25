package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthScope_Validate(t *testing.T) {
	tests := []struct {
		name    string
		scope   OAuthScope
		wantErr string
	}{
		{
			name:    "valid scope",
			scope:   OAuthScope{ScopeValue: "repo", Description: "Full repository access"},
			wantErr: "",
		},
		{
			name:    "valid scope with special characters",
			scope:   OAuthScope{ScopeValue: "user:email", Description: "Access user email address"},
			wantErr: "",
		},
		{
			name:    "missing scope_value",
			scope:   OAuthScope{Description: "Full repository access"},
			wantErr: "scope_value is required",
		},
		{
			name:    "missing description",
			scope:   OAuthScope{ScopeValue: "repo"},
			wantErr: "description is required",
		},
		{
			name:    "both fields empty",
			scope:   OAuthScope{},
			wantErr: "scope_value is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scope.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOAuthScope_Copy(t *testing.T) {
	t.Run("copy is independent", func(t *testing.T) {
		original := OAuthScope{
			ScopeValue:  "repo",
			Description: "Full repository access",
		}

		copied := original.Copy()

		assert.Equal(t, original.ScopeValue, copied.ScopeValue)
		assert.Equal(t, original.Description, copied.Description)

		// Mutate copy — original must not change
		copied.Description = "Modified"
		assert.Equal(t, "Full repository access", original.Description)
		assert.Equal(t, "Modified", copied.Description)
	})

	t.Run("copy via pointer receiver", func(t *testing.T) {
		original := &OAuthScope{
			ScopeValue:  "openid",
			Description: "OpenID Connect scope",
		}
		copied := original.Copy()
		assert.Equal(t, original.ScopeValue, copied.ScopeValue)
		assert.Equal(t, original.Description, copied.Description)
	})
}
