package oauth2server

import (
	"errors"

	"github.com/ory/fosite"
)

// Sentinel errors for OAuth2 conditions surfaced to callers outside oauth2server.
// These keep fosite types contained within this package.
var (
	ErrInvalidRedirectURI      = errors.New("redirect_uri is not registered or does not meet security requirements")
	ErrInvalidClient           = errors.New("invalid_client")
	ErrInvalidGrant            = errors.New("invalid_grant")
	ErrInvalidRequest          = errors.New("invalid_request")
	ErrInvalidScope            = errors.New("invalid_scope")
	ErrServerError             = errors.New("server_error")
	ErrUnsupportedResponseType = errors.New("unsupported_response_type")
)

// RFC6749Error carries a structured OAuth2 error across the domain/adapter
// boundary without exposing fosite types to callers outside oauth2server.
type RFC6749Error struct {
	ErrorCode   string
	Description string
	HTTPStatus  int
	sentinel    error
}

func (e *RFC6749Error) Error() string { return e.ErrorCode + ": " + e.Description }
func (e *RFC6749Error) Unwrap() error { return e.sentinel }

// NewRFC6749Error constructs an RFC6749Error. Use this in tests and adapters
// that produce errors outside the Provider translation path.
func NewRFC6749Error(errorCode, description string, httpStatus int, sentinel error) *RFC6749Error {
	return &RFC6749Error{
		ErrorCode:   errorCode,
		Description: description,
		HTTPStatus:  httpStatus,
		sentinel:    sentinel,
	}
}

// translateFositeError converts a *fosite.RFC6749Error to a domain-native
// *RFC6749Error. Non-fosite errors are returned unchanged.
// Errors that already carry a domain sentinel (e.g. ErrInvalidRedirectURI via %w)
// are returned unchanged so callers can still detect the domain sentinel.
func translateFositeError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrInvalidRedirectURI) {
		return err
	}
	var fe *fosite.RFC6749Error
	if !errors.As(err, &fe) {
		return err
	}
	return &RFC6749Error{
		ErrorCode:   fe.ErrorField,
		Description: fe.DescriptionField,
		HTTPStatus:  fe.CodeField,
		sentinel:    domainSentinel(err),
	}
}

func domainSentinel(err error) error {
	switch {
	case errors.Is(err, fosite.ErrInvalidClient):
		return ErrInvalidClient
	case errors.Is(err, fosite.ErrInvalidGrant):
		return ErrInvalidGrant
	case errors.Is(err, fosite.ErrInvalidRequest):
		return ErrInvalidRequest
	case errors.Is(err, fosite.ErrInvalidScope):
		return ErrInvalidScope
	case errors.Is(err, fosite.ErrServerError):
		return ErrServerError
	case errors.Is(err, fosite.ErrUnsupportedResponseType):
		return ErrUnsupportedResponseType
	default:
		return nil
	}
}
