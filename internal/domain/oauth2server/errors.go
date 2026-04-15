package oauth2server

import (
	"errors"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// isNotFoundError reports whether err is a storage not-found error.
// Used to distinguish expected misses (legitimate 401) from infrastructure failures.
func isNotFoundError(err error) bool {
	var se *storage.StorageError
	return errors.As(err, &se) && se.Kind == storage.ErrorKindNotFound
}

// Sentinel errors for OAuth2 error classification.
// These enable callers to use errors.Is() instead of string matching.

// ErrInvalidScope is returned when requested scope is not allowed.
var ErrInvalidScope = errors.New("invalid_scope")

// ErrInvalidGrant is returned for authorization code exchange failures.
var ErrInvalidGrant = errors.New("invalid_grant")

// ErrInvalidRequest is returned when the request parameters are malformed.
var ErrInvalidRequest = errors.New("invalid_request")

// ErrUnsupportedResponseType is returned for unsupported response_type values.
var ErrUnsupportedResponseType = errors.New("unsupported_response_type")

// ErrUnknownClient is returned when the client_id doesn't match a known agent.
var ErrUnknownClient = errors.New("unknown_client")

// ErrInvalidRedirectURI is returned when redirect_uri doesn't match registered URIs.
var ErrInvalidRedirectURI = errors.New("invalid_redirect_uri")
