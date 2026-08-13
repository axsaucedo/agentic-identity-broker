package storage

import (
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRequirement_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sr      *ServiceRequirement
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mandatory requirement",
			sr: &ServiceRequirement{
				ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				RequirementType: RequirementTypeMandatory,
				RequiredScopes:  []string{"repo", "user:email"},
			},
			wantErr: false,
		},
		{
			name: "valid optional requirement",
			sr: &ServiceRequirement{
				ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				RequirementType: RequirementTypeOptional,
				RequiredScopes:  []string{"read:user"},
			},
			wantErr: false,
		},
		{
			name: "missing service_id",
			sr: &ServiceRequirement{
				RequirementType: RequirementTypeMandatory,
				RequiredScopes:  []string{"repo"},
			},
			wantErr: true,
			errMsg:  "service_id is required",
		},
		{
			name: "invalid requirement_type",
			sr: &ServiceRequirement{
				ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				RequirementType: RequirementType("invalid"),
				RequiredScopes:  []string{"repo"},
			},
			wantErr: true,
			errMsg:  "requirement_type validation failed",
		},
		{
			name: "empty required_scopes",
			sr: &ServiceRequirement{
				ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				RequirementType: RequirementTypeMandatory,
				RequiredScopes:  []string{},
			},
			wantErr: false,
		},
		{
			name: "nil required_scopes",
			sr: &ServiceRequirement{
				ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				RequirementType: RequirementTypeMandatory,
				RequiredScopes:  nil,
			},
			wantErr: false,
		},
		{
			name: "empty scope in required_scopes",
			sr: &ServiceRequirement{
				ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				RequirementType: RequirementTypeMandatory,
				RequiredScopes:  []string{"repo", "", "user:email"},
			},
			wantErr: true,
			errMsg:  "required_scopes[1] is empty",
		},
		{
			name: "require_all_scopes with empty required_scopes",
			sr: &ServiceRequirement{
				ServiceID:        id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				RequirementType:  RequirementTypeMandatory,
				RequireAllScopes: true,
			},
			wantErr: false,
		},
		{
			name: "require_all_scopes with required_scopes",
			sr: &ServiceRequirement{
				ServiceID:        id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				RequirementType:  RequirementTypeMandatory,
				RequiredScopes:   []string{"repo"},
				RequireAllScopes: true,
			},
			wantErr: true,
			errMsg:  "required_scopes must be empty when require_all_scopes is true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sr.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceRequirement_HasScope(t *testing.T) {
	sr := &ServiceRequirement{
		ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
		RequirementType: RequirementTypeMandatory,
		RequiredScopes:  []string{"repo", "user:email"},
	}

	assert.True(t, sr.HasScope("repo"))
	assert.True(t, sr.HasScope("user:email"))
	assert.False(t, sr.HasScope("read:user"))
	assert.False(t, sr.HasScope("REPO"), "scope comparison is case-sensitive")
}

func TestServiceRequirement_IsMandatory(t *testing.T) {
	mandatory := &ServiceRequirement{
		ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
		RequirementType: RequirementTypeMandatory,
		RequiredScopes:  []string{"repo"},
	}
	optional := &ServiceRequirement{
		ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
		RequirementType: RequirementTypeOptional,
		RequiredScopes:  []string{"repo"},
	}

	assert.True(t, mandatory.IsMandatory())
	assert.False(t, optional.IsMandatory())
}

func TestServiceRequirement_IsOptional(t *testing.T) {
	mandatory := &ServiceRequirement{
		ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
		RequirementType: RequirementTypeMandatory,
		RequiredScopes:  []string{"repo"},
	}
	optional := &ServiceRequirement{
		ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
		RequirementType: RequirementTypeOptional,
		RequiredScopes:  []string{"repo"},
	}

	assert.False(t, mandatory.IsOptional())
	assert.True(t, optional.IsOptional())
}
