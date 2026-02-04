package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequirementType_Valid(t *testing.T) {
	tests := []struct {
		name string
		rt   RequirementType
		want bool
	}{
		{
			name: "mandatory is valid",
			rt:   RequirementTypeMandatory,
			want: true,
		},
		{
			name: "optional is valid",
			rt:   RequirementTypeOptional,
			want: true,
		},
		{
			name: "empty string is invalid",
			rt:   RequirementType(""),
			want: false,
		},
		{
			name: "random string is invalid",
			rt:   RequirementType("required"),
			want: false,
		},
		{
			name: "uppercase is invalid (case-sensitive)",
			rt:   RequirementType("MANDATORY"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rt.Valid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequirementType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rt      RequirementType
		wantErr bool
	}{
		{
			name:    "mandatory is valid",
			rt:      RequirementTypeMandatory,
			wantErr: false,
		},
		{
			name:    "optional is valid",
			rt:      RequirementTypeOptional,
			wantErr: false,
		},
		{
			name:    "empty string returns error",
			rt:      RequirementType(""),
			wantErr: true,
		},
		{
			name:    "invalid value returns error",
			rt:      RequirementType("required"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rt.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequirementType_String(t *testing.T) {
	assert.Equal(t, "mandatory", RequirementTypeMandatory.String())
	assert.Equal(t, "optional", RequirementTypeOptional.String())
}

func TestRequirementType_IsMandatory(t *testing.T) {
	assert.True(t, RequirementTypeMandatory.IsMandatory())
	assert.False(t, RequirementTypeOptional.IsMandatory())
	assert.False(t, RequirementType("invalid").IsMandatory())
}

func TestRequirementType_IsOptional(t *testing.T) {
	assert.True(t, RequirementTypeOptional.IsOptional())
	assert.False(t, RequirementTypeMandatory.IsOptional())
	assert.False(t, RequirementType("invalid").IsOptional())
}
