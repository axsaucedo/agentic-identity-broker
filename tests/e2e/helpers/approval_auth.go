package helpers

import (
	"fmt"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	tokenexchange "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
)

const ApprovalClientAssertionHeader = "X-Client-Assertion"

// ApprovalRequestAuthFixture holds reusable auth material for machine-facing approval API tests.
type ApprovalRequestAuthFixture struct {
	upstream               *MockUpstreamOAuth2Server
	AgentID                id.AgentID
	ClientAssertion        string
	InvalidClientAssertion string
}

// NewApprovalRequestAuthFixture creates machine-facing approval auth fixtures using the mock upstream signer.
func NewApprovalRequestAuthFixture(upstream *MockUpstreamOAuth2Server, agentID id.AgentID) (*ApprovalRequestAuthFixture, error) {
	auth := &ApprovalRequestAuthFixture{
		upstream:               upstream,
		AgentID:                agentID,
		InvalidClientAssertion: "not.a.valid.jwt.structure",
	}

	now := time.Now()
	assertionClaims := map[string]interface{}{
		"sub": "approval-gateway-client",
		"iss": upstream.URL(),
		"aud": []string{tokenexchange.DefaultBrokerAudience},
		"exp": now.Add(1 * time.Hour).Unix(),
		"iat": now.Unix(),
	}

	clientAssertion, err := SignTestJWT(assertionClaims, upstream.GetPrivateKeyPEM())
	if err != nil {
		return nil, fmt.Errorf("sign client assertion: %w", err)
	}

	auth.ClientAssertion = clientAssertion
	return auth, nil
}

// SubjectToken returns a valid subject token for the given principal and agent.
func (a *ApprovalRequestAuthFixture) SubjectToken(principal string) (string, error) {
	claims := map[string]interface{}{
		"sub": principal,
		"azp": a.AgentID.String(),
		"iss": a.upstream.URL(),
		"aud": []string{tokenexchange.DefaultBrokerAudience},
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token, err := SignTestJWT(claims, a.upstream.GetPrivateKeyPEM())
	if err != nil {
		return "", fmt.Errorf("sign subject token: %w", err)
	}

	return token, nil
}

// SubjectTokenWithoutPrincipal returns a syntactically valid subject token missing the principal claim.
func (a *ApprovalRequestAuthFixture) SubjectTokenWithoutPrincipal() (string, error) {
	claims := map[string]interface{}{
		"azp": a.AgentID.String(),
		"iss": a.upstream.URL(),
		"aud": []string{tokenexchange.DefaultBrokerAudience},
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token, err := SignTestJWT(claims, a.upstream.GetPrivateKeyPEM())
	if err != nil {
		return "", fmt.Errorf("sign subject token without principal: %w", err)
	}

	return token, nil
}

// ApprovalCreateHeaders builds headers for POST /api/approvals dual-auth requests.
func ApprovalCreateHeaders(subjectToken, clientAssertion string, _ id.AgentID) map[string]string {
	return map[string]string{
		"Authorization":               "Bearer " + subjectToken,
		ApprovalClientAssertionHeader: clientAssertion,
	}
}

// ApprovalSyncHeaders builds headers for GET /api/approvals client-assertion-only requests.
func ApprovalSyncHeaders(clientAssertion string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + clientAssertion,
	}
}

// ApprovalSubjectTokenHeaders builds headers for subject-token-only approval requests.
func ApprovalSubjectTokenHeaders(subjectToken string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + subjectToken,
	}
}
