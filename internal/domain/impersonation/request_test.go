package impersonation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
)

func baseImpersonationForm() map[string][]string {
	return map[string][]string{
		"grant_type":            {GrantType},
		"audience":              {"https://broker/impersonation"},
		"client_assertion":      {"ca.jwt"},
		"client_assertion_type": {JWTBearerType},
		"actor_token":           {"at.jwt"},
		"actor_token_type":      {JWTTokenType},
		"subject_token":         {"st.jwt"},
		"subject_token_type":    {JWTTokenType},
	}
}

func codeOf(t *testing.T, err error) string {
	t.Helper()
	var te *tokenexchange.TokenExchangeError
	require.ErrorAs(t, err, &te)
	return te.Code()
}

func TestParseRequest_Valid(t *testing.T) {
	req, err := ParseRequest(baseImpersonationForm())
	require.NoError(t, err)
	assert.Equal(t, "ca.jwt", req.ClientAssertion)
	assert.Equal(t, "at.jwt", req.ActorToken)
	assert.Equal(t, "st.jwt", req.SubjectToken)
	assert.Equal(t, JWTTokenType, req.SubjectTokenType)
}
func TestParseRequest_ParsesScope(t *testing.T) {
	t.Run("empty scope", func(t *testing.T) {
		form := baseImpersonationForm()
		form["scope"] = []string{""}

		req, err := ParseRequest(form)

		require.NoError(t, err)
		assert.Empty(t, req.Scope)
		assert.Empty(t, req.Scopes)
	})

	t.Run("literal space separated scope", func(t *testing.T) {
		form := baseImpersonationForm()
		form["scope"] = []string{"read  write"}

		req, err := ParseRequest(form)

		require.NoError(t, err)
		assert.Equal(t, "read  write", req.Scope)
		assert.Equal(t, []string{"read", "", "write"}, req.Scopes)
	})
}

func TestParseRequest_Rejections(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(f map[string][]string)
	}{
		{"resource present", func(f map[string][]string) { f["resource"] = []string{"https://api"} }},
		{"missing client_assertion", func(f map[string][]string) { delete(f, "client_assertion") }},
		{"empty actor_token", func(f map[string][]string) { f["actor_token"] = []string{""} }},
		{"repeated subject_token", func(f map[string][]string) { f["subject_token"] = []string{"a", "b"} }},
		{"wrong client_assertion_type", func(f map[string][]string) { f["client_assertion_type"] = []string{"other"} }},
		{"wrong actor_token_type", func(f map[string][]string) { f["actor_token_type"] = []string{AccessTokenType} }},
		{"requested_token_type mismatch", func(f map[string][]string) {
			f["requested_token_type"] = []string{JWTTokenType}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			form := baseImpersonationForm()
			tc.mutate(form)
			_, err := ParseRequest(form)
			require.Error(t, err)
			assert.Equal(t, tokenexchange.InvalidRequestError, codeOf(t, err))
		})
	}
}

func TestParseRequest_RequestedTokenTypeAccessOK(t *testing.T) {
	form := baseImpersonationForm()
	form["requested_token_type"] = []string{AccessTokenType}
	_, err := ParseRequest(form)
	require.NoError(t, err)
}

// FR-002a/003c: a bare or empty forbidden/optional parameter still counts as present.
func TestParseRequest_RejectsEmptyValuedParameters(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(f map[string][]string)
	}{
		{"empty resource", func(f map[string][]string) { f["resource"] = []string{""} }},
		{"empty requested_token_type", func(f map[string][]string) { f["requested_token_type"] = []string{""} }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			form := baseImpersonationForm()
			tc.mutate(form)
			_, err := ParseRequest(form)
			require.Error(t, err)
			assert.Equal(t, tokenexchange.InvalidRequestError, codeOf(t, err))
		})
	}
}
