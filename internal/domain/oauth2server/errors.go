package oauth2server

import "errors"

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

// ErrServerError is returned for internal infrastructure failures (timeout, connection error, etc.)
// that are not attributable to client input. Maps to server_error / HTTP 500.
var ErrServerError = errors.New("server_error")
