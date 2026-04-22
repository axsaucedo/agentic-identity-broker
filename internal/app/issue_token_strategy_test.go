package app

import (
	"context"
	"errors"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueTokenCodeIssuer_InvalidClientID(t *testing.T) {
	issuer := newIssueTokenCodeIssuer(nil) // provider is never reached on parse failure

	_, err := issuer.IssueAuthorizationCode(context.Background(), &ports.AuthorizationRequest{
		ClientID: id.ClientID("not-a-uuid"),
	}, id.NewPrincipal("user@example.com"))

	require.Error(t, err)
	assert.True(t, errors.Is(err, oauth2server.ErrInvalidClient))
}
