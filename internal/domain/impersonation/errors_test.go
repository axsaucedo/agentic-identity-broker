package impersonation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
)

func TestSelectNoMatchError_Precedence(t *testing.T) {
	ad := accessDenied("denied", "authorization_denied")
	ic := invalidClient("bad client", "client_assertion_invalid")
	ir := invalidRequest("bad actor", "actor_token_invalid")

	// access_denied outranks invalid_client and invalid_request, order-independently.
	assert.Equal(t, tokenexchange.AccessDeniedError, selectNoMatchError([]*tokenexchange.TokenExchangeError{ir, ic, ad}).Code())
	assert.Equal(t, tokenexchange.AccessDeniedError, selectNoMatchError([]*tokenexchange.TokenExchangeError{ad, ir, ic}).Code())

	// invalid_request outranks invalid_client.
	assert.Equal(t, tokenexchange.InvalidRequestError, selectNoMatchError([]*tokenexchange.TokenExchangeError{ir, ic}).Code())

	// invalid_request alone.
	assert.Equal(t, tokenexchange.InvalidRequestError, selectNoMatchError([]*tokenexchange.TokenExchangeError{ir}).Code())

	// empty → server_error fallback (fail closed).
	assert.Equal(t, tokenexchange.ServerErrorCode, selectNoMatchError(nil).Code())
}
