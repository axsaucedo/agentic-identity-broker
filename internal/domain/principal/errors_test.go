package principal_test

import (
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/stretchr/testify/assert"
)

func TestMissingPrincipalError(t *testing.T) {
	err := principal.NewMissingPrincipalError("X-Remote-User")

	assert.NotNil(t, err)
	assert.Equal(t, "missing or empty principal", err.Error())
	assert.Equal(t, "X-Remote-User", err.HeaderName)
}

func TestInvalidPrincipalError(t *testing.T) {
	reason := "principal exceeds maximum length"
	value := "very-long-principal"

	err := principal.NewInvalidPrincipalError(reason, value)

	assert.NotNil(t, err)
	assert.Equal(t, reason, err.Error())
	assert.Equal(t, reason, err.Reason)
	assert.Equal(t, value, err.Value)
}

func TestPrincipalTooLongError(t *testing.T) {
	err := principal.PrincipalTooLongError(201, 200)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum length")
	assert.Contains(t, err.Error(), "200")
	assert.Contains(t, err.Error(), "201")
}

func TestErrorImplementsErrorInterface(t *testing.T) {
	testCases := []struct {
		name error
	}{
		{principal.NewMissingPrincipalError("X-Remote-User")},
		{principal.NewInvalidPrincipalError("invalid", "value")},
		{principal.PrincipalTooLongError(201, 200)},
	}

	for _, tc := range testCases {
		// Should be assignable to error interface
		var _ = tc.name
		assert.NotEmpty(t, tc.name.Error())
	}
}

func TestMissingPrincipalErrorCreationVariants(t *testing.T) {
	headerNames := []string{
		"X-Remote-User",
		"X-Authenticated-User",
		"Remote-User",
		"Authorization",
	}

	for _, headerName := range headerNames {
		err := principal.NewMissingPrincipalError(headerName)
		assert.Equal(t, headerName, err.HeaderName)
		assert.Equal(t, "missing or empty principal", err.Error())
	}
}

func TestInvalidPrincipalErrorVariants(t *testing.T) {
	testCases := []struct {
		reason string
		value  string
	}{
		{"too long", "very-long-value-that-exceeds-limit"},
		{"invalid utf-8", "invalid\xFF\xFE"},
		{"contains newline", "value\nwith\nnewlines"},
	}

	for _, tc := range testCases {
		err := principal.NewInvalidPrincipalError(tc.reason, tc.value)
		assert.Equal(t, tc.reason, err.Reason)
		assert.Equal(t, tc.value, err.Value)
	}
}
