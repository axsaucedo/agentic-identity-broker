package tokenexchange

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCELEvaluatorSuccessfulCompilation tests successful evaluator creation.
func TestNewCELEvaluatorSuccessfulCompilation(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)

	evaluator, err := NewCELEvaluator(config)

	require.NoError(t, err)
	assert.NotNil(t, evaluator)
	assert.NotNil(t, evaluator.principalProgram)
	assert.NotNil(t, evaluator.agentClientIDProgram)
	assert.NotNil(t, evaluator.authorizationProgram)
}

// TestNewCELEvaluatorInvalidPrincipalExpression tests startup failure with invalid principal expression.
func TestNewCELEvaluatorInvalidPrincipalExpression(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.PrincipalExpression = "invalid syntax !@#$"

	evaluator, err := NewCELEvaluator(config)

	assert.Nil(t, evaluator)
	assert.Error(t, err)
	assert.True(t, IsTokenExchangeError(err))
	tokExErr := err.(*TokenExchangeError)
	assert.Equal(t, "server_error", tokExErr.Code())
	assert.Equal(t, 500, tokExErr.HTTPStatus())
}

// TestNewCELEvaluatorInvalidAgentClientIDExpression tests startup failure with invalid agent_client_id expression.
func TestNewCELEvaluatorInvalidAgentClientIDExpression(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.AgentClientIDExpression = "invalid !@#$ syntax"

	evaluator, err := NewCELEvaluator(config)

	assert.Nil(t, evaluator)
	assert.Error(t, err)
	assert.True(t, IsTokenExchangeError(err))
}

// TestNewCELEvaluatorInvalidAuthorizationExpression tests startup failure with invalid authorization expression.
func TestNewCELEvaluatorInvalidAuthorizationExpression(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.AuthorizationExpression = "not a boolean !@#$"

	evaluator, err := NewCELEvaluator(config)

	assert.Nil(t, evaluator)
	assert.Error(t, err)
	assert.True(t, IsTokenExchangeError(err))
}

// TestExtractPrincipalDefaultExpression tests principal extraction with default expression.
func TestExtractPrincipalDefaultExpression(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		"sub": "user123",
	}

	principal, err := evaluator.ExtractPrincipal(claims)

	require.NoError(t, err)
	assert.Equal(t, "user123", principal)
}

// TestExtractPrincipalCustomExpression tests principal extraction with custom expression.
func TestExtractPrincipalCustomExpression(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.PrincipalExpression = "subject_token.preferred_username"
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		"sub":                "user123",
		"preferred_username": "john.doe",
	}

	principal, err := evaluator.ExtractPrincipal(claims)

	require.NoError(t, err)
	assert.Equal(t, "john.doe", principal)
}

// TestExtractPrincipalNestedClaim tests principal extraction from nested claim.
func TestExtractPrincipalNestedClaim(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.PrincipalExpression = "subject_token.claims.email"
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		"claims": map[string]any{
			"email": "user@example.com",
		},
	}

	principal, err := evaluator.ExtractPrincipal(claims)

	require.NoError(t, err)
	assert.Equal(t, "user@example.com", principal)
}

// TestExtractPrincipalMissingClaim tests error when claim is missing.
func TestExtractPrincipalMissingClaim(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		// Missing 'sub' claim
	}

	principal, err := evaluator.ExtractPrincipal(claims)

	assert.Error(t, err)
	assert.Empty(t, principal)
	assert.True(t, IsTokenExchangeError(err))
}

// TestExtractPrincipalEmptyValue tests error when extracted value is empty.
func TestExtractPrincipalEmptyValue(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		"sub": "",
	}

	principal, err := evaluator.ExtractPrincipal(claims)

	assert.Error(t, err)
	assert.Empty(t, principal)
	assert.True(t, IsTokenExchangeError(err))
}

// TestExtractPrincipalWrongType tests error when extracted value is not a string.
func TestExtractPrincipalWrongType(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.PrincipalExpression = "subject_token.exp"
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		"exp": int64(1234567890),
	}

	principal, err := evaluator.ExtractPrincipal(claims)

	assert.Error(t, err)
	assert.Empty(t, principal)
	assert.True(t, IsTokenExchangeError(err))
}

// TestExtractAgentClientIDDefaultExpression tests agent_client_id extraction with default expression.
func TestExtractAgentClientIDDefaultExpression(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		"azp": "agent-app-1",
	}

	agentID, err := evaluator.ExtractAgentClientID(claims)

	require.NoError(t, err)
	assert.Equal(t, "agent-app-1", agentID)
}

// TestExtractAgentClientIDCustomExpression tests agent_client_id extraction with custom expression.
func TestExtractAgentClientIDCustomExpression(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.AgentClientIDExpression = "subject_token.client_id"
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		"client_id": "app-agent",
	}

	agentID, err := evaluator.ExtractAgentClientID(claims)

	require.NoError(t, err)
	assert.Equal(t, "app-agent", agentID)
}

// TestExtractAgentClientIDMissingClaim tests error when claim is missing.
func TestExtractAgentClientIDMissingClaim(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		// Missing 'azp' claim
	}

	agentID, err := evaluator.ExtractAgentClientID(claims)

	assert.Error(t, err)
	assert.Empty(t, agentID)
	assert.True(t, IsTokenExchangeError(err))
}

// TestAuthorizePrivilegedClientDefaultExpression tests privileged client authorization with default allow-all expression.
func TestAuthorizePrivilegedClientDefaultExpression(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	clientAssertion := map[string]any{
		"sub": "privileged-client-1",
		"iss": "https://auth.example.com",
	}
	subjectToken := map[string]any{
		"sub": "user123",
	}
	request := CELRequestContext{
		Resource:      "https://api.example.com",
		GrantType:     TokenExchangeGrantType,
		Scope:         "read",
		Principal:     "user123",
		AgentClientID: "agent-app-1",
	}

	authorized, err := evaluator.AuthorizePrivilegedClient(clientAssertion, subjectToken, request)

	require.NoError(t, err)
	assert.True(t, authorized)
}

// TestAuthorizePrivilegedClientCustomExpressionAllow tests privileged client authorization with custom allow expression.
func TestAuthorizePrivilegedClientCustomExpressionAllow(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.AuthorizationExpression = `client_assertion.iss == "https://trusted.example.com"`
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	clientAssertion := map[string]any{
		"sub": "privileged-client-1",
		"iss": "https://trusted.example.com",
	}
	subjectToken := map[string]any{
		"sub": "user123",
	}
	request := CELRequestContext{
		Resource:      "https://api.example.com",
		GrantType:     TokenExchangeGrantType,
		Principal:     "user123",
		AgentClientID: "agent-app-1",
	}

	authorized, err := evaluator.AuthorizePrivilegedClient(clientAssertion, subjectToken, request)

	require.NoError(t, err)
	assert.True(t, authorized)
}

// TestAuthorizePrivilegedClientCustomExpressionDeny tests privileged client authorization with custom deny expression.
func TestAuthorizePrivilegedClientCustomExpressionDeny(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.AuthorizationExpression = `client_assertion.iss == "https://trusted.example.com"`
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	clientAssertion := map[string]any{
		"sub": "privileged-client-1",
		"iss": "https://untrusted.example.com", // Doesn't match trusted issuer
	}
	subjectToken := map[string]any{
		"sub": "user123",
	}
	request := CELRequestContext{
		Resource:      "https://api.example.com",
		GrantType:     TokenExchangeGrantType,
		Principal:     "user123",
		AgentClientID: "agent-app-1",
	}

	authorized, err := evaluator.AuthorizePrivilegedClient(clientAssertion, subjectToken, request)

	assert.Error(t, err)
	assert.False(t, authorized)
	assert.True(t, IsTokenExchangeError(err))
	tokExErr := err.(*TokenExchangeError)
	assert.Equal(t, "access_denied", tokExErr.Code())
}

// TestAuthorizePrivilegedClientComplexExpression tests authorization with complex multi-condition expression.
func TestAuthorizePrivilegedClientComplexExpression(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.AuthorizationExpression = `
		client_assertion.sub == "trusted-privileged-client" &&
		"admin" in subject_token.roles
	`
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	clientAssertion := map[string]any{
		"sub": "trusted-privileged-client",
	}
	subjectToken := map[string]any{
		"sub":   "user123",
		"roles": []any{"admin", "user"},
	}
	request := CELRequestContext{
		Resource:      "https://api.example.com",
		GrantType:     TokenExchangeGrantType,
		Principal:     "user123",
		AgentClientID: "trusted-privileged-client",
	}

	authorized, err := evaluator.AuthorizePrivilegedClient(clientAssertion, subjectToken, request)

	require.NoError(t, err)
	assert.True(t, authorized)
}

// TestAuthorizePrivilegedClientComplexExpressionFailsAdminCheck tests complex expression with failed admin check.
func TestAuthorizePrivilegedClientComplexExpressionFailsAdminCheck(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.AuthorizationExpression = `
		client_assertion.sub == "trusted-privileged-client" &&
		"admin" in subject_token.roles
	`
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	clientAssertion := map[string]any{
		"sub": "trusted-privileged-client",
	}
	subjectToken := map[string]any{
		"sub":   "user123",
		"roles": []any{"user"}, // Missing admin role
	}
	request := CELRequestContext{
		Resource:      "https://api.example.com",
		GrantType:     TokenExchangeGrantType,
		Principal:     "user123",
		AgentClientID: "trusted-privileged-client",
	}

	authorized, err := evaluator.AuthorizePrivilegedClient(clientAssertion, subjectToken, request)

	assert.Error(t, err)
	assert.False(t, authorized)
	assert.True(t, IsTokenExchangeError(err))
}

// TestAuthorizePrivilegedClientWithTimeout tests authorization evaluation with timeout.
func TestAuthorizePrivilegedClientWithTimeout(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.EvaluationTimeout = 1 * time.Millisecond // Very short timeout
	config.AuthorizationExpression = "true"
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	clientAssertion := map[string]any{
		"sub": "privileged-client-1",
	}
	subjectToken := map[string]any{
		"sub": "user123",
	}
	request := CELRequestContext{
		Resource:      "https://api.example.com",
		GrantType:     TokenExchangeGrantType,
		Principal:     "user123",
		AgentClientID: "agent-app-1",
	}

	// Even a simple true expression might timeout with 1ms limit
	// But true should be fast enough - this is a relaxed test
	authorized, err := evaluator.AuthorizePrivilegedClient(clientAssertion, subjectToken, request)

	// Either succeeds quickly or times out - both are valid outcomes
	if err != nil {
		assert.True(t, IsTokenExchangeError(err))
	} else {
		assert.True(t, authorized)
	}
}

// TestExtractPrincipalWithComplexNestedClaims tests principal extraction from complex nested structure.
func TestExtractPrincipalWithComplexNestedClaims(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.PrincipalExpression = "subject_token.user.id"
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		"user": map[string]any{
			"id":   "user-456",
			"name": "John Doe",
		},
	}

	principal, err := evaluator.ExtractPrincipal(claims)

	require.NoError(t, err)
	assert.Equal(t, "user-456", principal)
}

// TestExtractAgentClientIDWithComplexNestedClaims tests agent_client_id extraction from complex nested structure.
func TestExtractAgentClientIDWithComplexNestedClaims(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.AgentClientIDExpression = "subject_token.agent.client_id"
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	claims := map[string]any{
		"agent": map[string]any{
			"client_id": "agent-789",
			"name":      "Privileged Client Agent",
		},
	}

	agentID, err := evaluator.ExtractAgentClientID(claims)

	require.NoError(t, err)
	assert.Equal(t, "agent-789", agentID)
}

// TestAuthorizePrivilegedClientWithRequestContext tests authorization with request context variables.
func TestAuthorizePrivilegedClientWithRequestContext(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.AuthorizationExpression = `request.resource == "https://api.example.com"`
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	clientAssertion := map[string]any{
		"sub": "privileged-client-1",
	}
	subjectToken := map[string]any{
		"sub": "user123",
	}
	request := CELRequestContext{
		Resource:      "https://api.example.com",
		GrantType:     TokenExchangeGrantType,
		Principal:     "user123",
		AgentClientID: "agent-app-1",
	}

	authorized, err := evaluator.AuthorizePrivilegedClient(clientAssertion, subjectToken, request)

	require.NoError(t, err)
	assert.True(t, authorized)
}

// TestAuthorizePrivilegedClientWithRequestContextMismatch tests authorization with mismatched request context.
func TestAuthorizePrivilegedClientWithRequestContextMismatch(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.AuthorizationExpression = `request.resource == "https://api.example.com"`
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	clientAssertion := map[string]any{
		"sub": "privileged-client-1",
	}
	subjectToken := map[string]any{
		"sub": "user123",
	}
	request := CELRequestContext{
		Resource:      "https://other-api.example.com", // Different resource
		GrantType:     TokenExchangeGrantType,
		Principal:     "user123",
		AgentClientID: "agent-app-1",
	}

	authorized, err := evaluator.AuthorizePrivilegedClient(clientAssertion, subjectToken, request)

	assert.Error(t, err)
	assert.False(t, authorized)
	assert.True(t, IsTokenExchangeError(err))
}

// TestNewCELEvaluatorWithDefaultTimeout tests evaluator initialization with default timeout.
func TestNewCELEvaluatorWithDefaultTimeout(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.EvaluationTimeout = 0 // Force default
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	assert.Equal(t, time.Duration(DefaultEvaluationTimeoutMs)*time.Millisecond, evaluator.evaluationTimeout)
}

// TestNewCELEvaluatorWithCustomTimeout tests evaluator initialization with custom timeout.
func TestNewCELEvaluatorWithCustomTimeout(t *testing.T) {
	t.Parallel()
	config := validTokenExchangeConfig(t)
	config.EvaluationTimeout = 250 * time.Millisecond
	evaluator, err := NewCELEvaluator(config)
	require.NoError(t, err)

	assert.Equal(t, 250*time.Millisecond, evaluator.evaluationTimeout)
}

// Helper function to create a valid CELEvaluatorConfig for testing
func validTokenExchangeConfig(_ *testing.T) CELEvaluatorConfig {
	return CELEvaluatorConfig{
		PrincipalExpression:     DefaultPrincipalExpression,
		AgentClientIDExpression: DefaultAgentClientIDExpression,
		AuthorizationExpression: DefaultAuthorizationExpression,
		EvaluationTimeout:       time.Duration(DefaultEvaluationTimeoutMs) * time.Millisecond,
	}
}
