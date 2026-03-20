package matchers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/onsi/gomega/types"
)

// ContainAgentIDClaim returns a Gomega matcher that verifies an HTTP token response
// contains a JWT access_token with the specified agent ID claim.
//
// The matcher:
//   - Reads the response body as JSON
//   - Extracts the "access_token" field
//   - Parses the token as a JWT (without signature verification)
//   - Asserts that the named claim is present and equals agentID
//
// Usage:
//
//	Expect(resp).To(matchers.ContainAgentIDClaim("x_agent_id", agent.ID.String()))
func ContainAgentIDClaim(claimName, agentID string) types.GomegaMatcher {
	return &agentIDClaimMatcher{
		claimName: claimName,
		agentID:   agentID,
	}
}

type agentIDClaimMatcher struct {
	claimName   string
	agentID     string
	actualValue string
	error       string
}

func (m *agentIDClaimMatcher) Match(actual interface{}) (success bool, err error) {
	resp, ok := actual.(*http.Response)
	if !ok {
		return false, fmt.Errorf("ContainAgentIDClaim matcher expects an *http.Response, got %T", actual)
	}

	if resp == nil {
		m.error = "response is nil"
		return false, nil
	}

	// Read the response body
	if resp.Body == nil {
		m.error = "response body is nil"
		return false, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		m.error = fmt.Sprintf("failed to read response body: %v", err)
		return false, nil
	}

	// Restore body for potential re-reading
	resp.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// Parse JSON body
	var body map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		m.error = fmt.Sprintf("failed to parse JSON response body: %v", err)
		return false, nil
	}

	// Extract access_token field
	accessTokenRaw, exists := body["access_token"]
	if !exists {
		m.error = "response body does not contain 'access_token' field"
		return false, nil
	}

	accessToken, ok := accessTokenRaw.(string)
	if !ok {
		m.error = fmt.Sprintf("'access_token' field is not a string, got %T", accessTokenRaw)
		return false, nil
	}

	if accessToken == "" {
		m.error = "'access_token' field is empty"
		return false, nil
	}

	// Parse JWT without signature verification
	tok, err := jwt.ParseInsecure([]byte(accessToken))
	if err != nil {
		m.error = fmt.Sprintf("failed to parse access_token as JWT: %v", err)
		return false, nil
	}

	// Serialize token to JSON to access all claims uniformly
	tokenJSON, err := json.Marshal(tok)
	if err != nil {
		m.error = fmt.Sprintf("failed to marshal JWT to JSON: %v", err)
		return false, nil
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(tokenJSON, &claims); err != nil {
		m.error = fmt.Sprintf("failed to unmarshal JWT claims: %v", err)
		return false, nil
	}

	// Check for the named claim
	claimValue, exists := claims[m.claimName]
	if !exists {
		m.error = fmt.Sprintf("JWT does not contain claim %q", m.claimName)
		return false, nil
	}

	claimStr, ok := claimValue.(string)
	if !ok {
		m.error = fmt.Sprintf("claim %q has type %T, expected string", m.claimName, claimValue)
		return false, nil
	}

	m.actualValue = claimStr

	if claimStr != m.agentID {
		m.error = fmt.Sprintf("claim %q value mismatch: expected %q, got %q", m.claimName, m.agentID, claimStr)
		return false, nil
	}

	return true, nil
}

func (m *agentIDClaimMatcher) FailureMessage(actual interface{}) string {
	msg := fmt.Sprintf("Expected JWT access_token to contain claim %q = %q", m.claimName, m.agentID)
	if m.actualValue != "" {
		msg += fmt.Sprintf("\nActual claim value: %q", m.actualValue)
	}
	if m.error != "" {
		msg += fmt.Sprintf("\nError: %s", m.error)
	}
	return msg
}

func (m *agentIDClaimMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected JWT access_token NOT to contain claim %q = %q", m.claimName, m.agentID)
}
