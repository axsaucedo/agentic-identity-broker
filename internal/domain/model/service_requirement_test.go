package model

import (
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequirementType_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		rt    RequirementType
		valid bool
	}{
		{name: "mandatory is valid", rt: RequirementTypeMandatory, valid: true},
		{name: "optional is valid", rt: RequirementTypeOptional, valid: true},
		{name: "empty string is invalid", rt: RequirementType(""), valid: false},
		{name: "unknown value is invalid", rt: RequirementType("required"), valid: false},
		{name: "uppercase is invalid", rt: RequirementType("Mandatory"), valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.valid, tt.rt.Valid())
		})
	}
}

func TestRequirementType_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		rt      RequirementType
		wantErr string
	}{
		{name: "mandatory valid", rt: RequirementTypeMandatory, wantErr: ""},
		{name: "optional valid", rt: RequirementTypeOptional, wantErr: ""},
		{
			name:    "invalid type",
			rt:      RequirementType("unknown"),
			wantErr: "invalid requirement type",
		},
		{
			name:    "empty type",
			rt:      RequirementType(""),
			wantErr: "invalid requirement type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.rt.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRequirementType_Predicates(t *testing.T) {
	t.Parallel()
	t.Run("IsMandatory", func(t *testing.T) {
		t.Parallel()
		assert.True(t, RequirementTypeMandatory.IsMandatory())
		assert.False(t, RequirementTypeOptional.IsMandatory())
	})

	t.Run("IsOptional", func(t *testing.T) {
		t.Parallel()
		assert.True(t, RequirementTypeOptional.IsOptional())
		assert.False(t, RequirementTypeMandatory.IsOptional())
	})

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "mandatory", RequirementTypeMandatory.String())
		assert.Equal(t, "optional", RequirementTypeOptional.String())
	})
}

func TestServiceRequirement_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		sr      ServiceRequirement
		wantErr string
	}{
		{
			name: "valid mandatory requirement",
			sr: ServiceRequirement{
				ServiceID:       id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				RequirementType: RequirementTypeMandatory,
				RequiredScopes:  []string{"repo", "user:email"},
			},
			wantErr: "",
		},
		{
			name: "valid optional requirement",
			sr: ServiceRequirement{
				ServiceID:       id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440002"),
				RequirementType: RequirementTypeOptional,
				RequiredScopes:  []string{"read:org"},
			},
			wantErr: "",
		},
		{
			name: "missing service_id",
			sr: ServiceRequirement{
				RequirementType: RequirementTypeMandatory,
				RequiredScopes:  []string{"repo"},
			},
			wantErr: "service_id is required",
		},
		{
			name: "invalid requirement_type",
			sr: ServiceRequirement{
				ServiceID:       id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				RequirementType: RequirementType("unknown"),
				RequiredScopes:  []string{"repo"},
			},
			wantErr: "requirement_type validation failed",
		},
		{
			name: "empty required_scopes",
			sr: ServiceRequirement{
				ServiceID:       id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				RequirementType: RequirementTypeMandatory,
				RequiredScopes:  []string{},
			},
		},
		{
			name: "nil required_scopes",
			sr: ServiceRequirement{
				ServiceID:       id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				RequirementType: RequirementTypeMandatory,
				RequiredScopes:  nil,
			},
		},
		{
			name: "empty scope in required_scopes",
			sr: ServiceRequirement{
				ServiceID:       id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
				RequirementType: RequirementTypeMandatory,
				RequiredScopes:  []string{"repo", ""},
			},
			wantErr: "required_scopes[1] is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.sr.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestServiceRequirement_HasScope(t *testing.T) {
	t.Parallel()
	sr := ServiceRequirement{
		ServiceID:       id.NewServiceID(),
		RequirementType: RequirementTypeMandatory,
		RequiredScopes:  []string{"repo", "user:email", "read:org"},
	}

	assert.True(t, sr.HasScope("repo"))
	assert.True(t, sr.HasScope("user:email"))
	assert.True(t, sr.HasScope("read:org"))
	assert.False(t, sr.HasScope("write:org"))
	assert.False(t, sr.HasScope("REPO"), "scope comparison is case-sensitive")
	assert.False(t, sr.HasScope(""))
}

func TestServiceRequirement_Predicates(t *testing.T) {
	t.Parallel()
	svcID1 := id.NewServiceID()
	svcID2 := id.NewServiceID()
	mandatory := ServiceRequirement{
		ServiceID:       svcID1,
		RequirementType: RequirementTypeMandatory,
		RequiredScopes:  []string{"repo"},
	}
	optional := ServiceRequirement{
		ServiceID:       svcID2,
		RequirementType: RequirementTypeOptional,
		RequiredScopes:  []string{"repo"},
	}

	assert.True(t, mandatory.IsMandatory())
	assert.False(t, mandatory.IsOptional())
	assert.True(t, optional.IsOptional())
	assert.False(t, optional.IsMandatory())
}

func TestServiceRequirement_Copy(t *testing.T) {
	t.Parallel()
	original := &ServiceRequirement{
		ServiceID:       id.MustParseServiceID("650e8400-e29b-41d4-a716-446655440001"),
		RequirementType: RequirementTypeMandatory,
		RequiredScopes:  []string{"repo", "user:email"},
	}

	copied := original.Copy()

	assert.Equal(t, original.ServiceID, copied.ServiceID)
	assert.Equal(t, original.RequirementType, copied.RequirementType)
	assert.Equal(t, original.RequiredScopes, copied.RequiredScopes)

	// Mutate copy — original must not change
	copied.RequiredScopes[0] = "modified"
	assert.Equal(t, "repo", original.RequiredScopes[0], "mutation of copy should not affect original")

	t.Run("nil scopes copy", func(t *testing.T) {
		t.Parallel()
		sr := &ServiceRequirement{
			ServiceID:       id.NewServiceID(),
			RequirementType: RequirementTypeOptional,
			RequiredScopes:  nil,
		}
		c := sr.Copy()
		assert.Nil(t, c.RequiredScopes)
	})
}
