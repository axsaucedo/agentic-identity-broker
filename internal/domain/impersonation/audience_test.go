package impersonation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

func TestResolveTarget(t *testing.T) {
	targetID := id.NewAgentID()
	targetAgent := &storage.Agent{ID: targetID, DisplayName: "Target", Description: "Target agent"}
	lookupErr := errors.New("storage unavailable")

	tests := []struct {
		name      string
		audiences []string
		lookup    func(context.Context, id.AgentID) (*storage.Agent, error)
		activated bool
		code      string
	}{
		{"absent audience", nil, nil, false, ""},
		{"different audience", []string{"https://elsewhere.example.com"}, nil, false, ""},
		{"multiple audiences", []string{"https://broker/impersonation/" + targetID.String(), "https://elsewhere.example.com"}, nil, false, ""},
		{"bare prefix", []string{"https://broker/impersonation"}, nil, true, tokenexchange.InvalidRequestError},
		{"empty suffix", []string{"https://broker/impersonation/"}, nil, true, tokenexchange.InvalidRequestError},
		{"malformed suffix", []string{"https://broker/impersonation/not-a-uuid"}, nil, true, tokenexchange.InvalidRequestError},
		{"noncanonical suffix", []string{"https://broker/impersonation/" + targetID.String()[0:35] + "A"}, nil, true, tokenexchange.InvalidRequestError},
		{"multi-segment suffix", []string{"https://broker/impersonation/" + targetID.String() + "/extra"}, nil, true, tokenexchange.InvalidRequestError},
		{"missing target", []string{"https://broker/impersonation/" + targetID.String()}, func(context.Context, id.AgentID) (*storage.Agent, error) { return nil, ports.ErrNotFound }, true, tokenexchange.InvalidTargetError},
		{"lookup failure", []string{"https://broker/impersonation/" + targetID.String()}, func(context.Context, id.AgentID) (*storage.Agent, error) { return nil, lookupErr }, true, tokenexchange.ServerErrorCode},
		{"registered target", []string{"https://broker/impersonation/" + targetID.String()}, func(_ context.Context, got id.AgentID) (*storage.Agent, error) {
			assert.Equal(t, targetID, got)
			return targetAgent, nil
		}, true, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, err := NewService(testImpersonationConfig(), func(ports.TrustedTokenIssuerConfig) (tokenexchange.JWKSProvider, error) {
				return stubJWKSProvider{}, nil
			}, stubAgentRepository{get: tc.lookup}, &stubIssuer{}, 0, nil)
			require.NoError(t, err)

			target, activated, err := svc.ResolveTarget(context.Background(), tc.audiences)
			assert.Equal(t, tc.activated, activated)
			if tc.code == "" {
				require.NoError(t, err)
				if tc.activated {
					require.NotNil(t, target)
					assert.Same(t, targetAgent, target.Agent)
				} else {
					assert.Nil(t, target)
				}
				return
			}
			require.Error(t, err)
			assert.Equal(t, tc.code, codeOf(t, err))
			assert.Nil(t, target)
		})
	}
}

func TestValidateAudiencePrefix(t *testing.T) {
	for _, prefix := range []string{"", "relative", "ftp://broker.example.com", "https:///missing-host", "https://user@broker.example.com", "https://broker.example.com?", "https://broker.example.com?query=x", "https://broker.example.com#fragment", "https://broker.example.com/"} {
		assert.Error(t, ValidateAudiencePrefix(prefix), prefix)
	}
	assert.NoError(t, ValidateAudiencePrefix("https://broker.example.com/impersonation"))
}
