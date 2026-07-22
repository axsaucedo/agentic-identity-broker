package storage

import (
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionSet_ValidateForCreate_ScopeLessServiceScope(t *testing.T) {
	t.Parallel()

	for _, scopes := range [][]string{nil, {}} {
		ps := &PermissionSet{
			Name:        "Scope-less service",
			Description: "Requires an authenticated session without OAuth scopes",
			ServiceScopes: []ServiceScope{{
				ServiceID:       id.NewServiceID(),
				Scopes:          scopes,
				RequirementType: RequirementTypeMandatory,
			}},
		}

		require.NoError(t, ps.ValidateForCreate())
	}
}

func TestPermissionSet_ValidateForCreate_RejectsBlankSuppliedScope(t *testing.T) {
	ps := &PermissionSet{
		Name:        "Invalid scope",
		Description: "Contains a blank scope",
		ServiceScopes: []ServiceScope{{
			ServiceID:       id.NewServiceID(),
			Scopes:          []string{""},
			RequirementType: RequirementTypeOptional,
		}},
	}

	err := ps.ValidateForCreate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scope cannot be empty")
}
