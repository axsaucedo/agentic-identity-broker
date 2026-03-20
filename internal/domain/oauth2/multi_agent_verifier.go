package oauth2

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lestrrat-go/jwx/v3/jwt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

// AgentIDMismatchError is returned by VerifyAgentIDClaim when the claim value in the
// upstream token does not match the expected agent ID. It carries structured fields so
// callers can log or handle the mismatch without parsing the error string.
type AgentIDMismatchError struct {
	Expected  string
	Received  string
	ClaimName string
}

func (e *AgentIDMismatchError) Error() string {
	return fmt.Sprintf("agent ID claim mismatch: expected %q, got %q for claim %q", e.Expected, e.Received, e.ClaimName)
}

// MultiAgentTokenVerifier verifies that an upstream token response contains the expected
// agent ID claim, as required by feature 021 (multi-agent client sharing).
//
// When multiple agents share a single upstream OAuth2 client_id, the upstream server
// is expected to embed the requesting agent's UUID into the issued token as a custom claim.
// This verifier checks that the claim is present and matches the agent that initiated the
// authorization request (fail-closed per SR-001).
//
// The verifier parses the JSON token response body, extracts the access_token JWT,
// and verifies the configured claim without signature validation. Signature verification
// was already performed by the upstream OAuth2 server; we only need the claim value here.
type MultiAgentTokenVerifier struct {
	agentIDClaimName string
}

// NewMultiAgentTokenVerifier creates a new MultiAgentTokenVerifier with the given claim name.
// Returns an error if agentIDClaimName is empty.
func NewMultiAgentTokenVerifier(agentIDClaimName string) (*MultiAgentTokenVerifier, error) {
	if agentIDClaimName == "" {
		return nil, fmt.Errorf("agentIDClaimName cannot be empty")
	}
	return &MultiAgentTokenVerifier{agentIDClaimName: agentIDClaimName}, nil
}

// VerifyAgentIDClaim verifies that the upstream token response body contains an access_token
// JWT with the configured agent ID claim matching expectedAgentID.
//
// Steps:
//  1. Parse JSON response body to extract "access_token"
//  2. Parse access_token JWT without signature verification (upstream already validated it)
//  3. Marshal token to JSON and extract all claims as a map
//  4. Look up the configured claim name in the claims map
//  5. Compare the claim value against expectedAgentID.String()
//
// Returns nil on success.
// Returns error with "agent ID claim absent" prefix if the claim is missing or cannot be extracted.
// Returns error with "agent ID claim mismatch" prefix if the claim value does not match.
func (v *MultiAgentTokenVerifier) VerifyAgentIDClaim(_ context.Context, responseBody []byte, expectedAgentID id.AgentID) error {
	// Step 1: Extract access_token from JSON response body
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(responseBody, &tokenResponse); err != nil {
		return fmt.Errorf("agent ID claim absent from upstream token: failed to parse response body: %w", err)
	}
	if tokenResponse.AccessToken == "" {
		return fmt.Errorf("agent ID claim absent from upstream token: access_token not found in response")
	}

	// Step 2: Parse JWT without signature verification.
	// Upstream already validated the token; we only need to read the custom claim value.
	tok, err := jwt.ParseInsecure([]byte(tokenResponse.AccessToken))
	if err != nil {
		return fmt.Errorf("agent ID claim absent from upstream token: failed to parse access_token as JWT: %w", err)
	}

	// Step 3: Marshal token to JSON and unmarshal into a flat claims map.
	// This provides uniform access to both standard and custom claims.
	tokenJSON, err := json.Marshal(tok)
	if err != nil {
		return fmt.Errorf("agent ID claim absent from upstream token: failed to marshal JWT claims: %w", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(tokenJSON, &claims); err != nil {
		return fmt.Errorf("agent ID claim absent from upstream token: failed to unmarshal JWT claims: %w", err)
	}

	// Step 4: Look up the configured claim
	claimValue, exists := claims[v.agentIDClaimName]
	if !exists {
		return fmt.Errorf("agent ID claim absent from upstream token: claim %q not found in JWT", v.agentIDClaimName)
	}

	claimStr, ok := claimValue.(string)
	if !ok {
		return fmt.Errorf("agent ID claim absent from upstream token: claim %q has type %T, expected string", v.agentIDClaimName, claimValue)
	}

	// Step 5: Compare against expected agent ID
	if claimStr != expectedAgentID.String() {
		return &AgentIDMismatchError{
			Expected:  expectedAgentID.String(),
			Received:  claimStr,
			ClaimName: v.agentIDClaimName,
		}
	}

	return nil
}
