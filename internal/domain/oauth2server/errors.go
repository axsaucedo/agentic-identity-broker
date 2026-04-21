package oauth2server

import (
	"errors"
	"net/http"

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
	errorCode   string
	description string
	httpStatus  int
	sentinel    error
}

func (e *RFC6749Error) Error() string       { return e.errorCode + ": " + e.description }
func (e *RFC6749Error) Unwrap() error       { return e.sentinel }
func (e *RFC6749Error) Code() string        { return e.errorCode }
func (e *RFC6749Error) Description() string { return e.description }
func (e *RFC6749Error) HTTPStatus() int     { return e.httpStatus }

// NewRFC6749Error constructs an RFC6749Error. Panics if errorCode is empty or
// httpStatus is not a valid HTTP status code (≤ 0).
func NewRFC6749Error(errorCode, description string, httpStatus int, sentinel error) *RFC6749Error {
	if errorCode == "" {
		panic("oauth2server: NewRFC6749Error: errorCode must not be empty")
	}
	if httpStatus <= 0 {
		panic("oauth2server: NewRFC6749Error: httpStatus must be a valid HTTP status code")
	}
	return &RFC6749Error{
		errorCode:   errorCode,
		description: description,
		httpStatus:  httpStatus,
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
	code := fe.ErrorField
	if code == "" {
		code = "server_error"
	}
	status := fe.CodeField
	if status <= 0 {
		status = http.StatusInternalServerError
	}
	return &RFC6749Error{
		errorCode:   code,
		description: fe.DescriptionField,
		httpStatus:  status,
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
