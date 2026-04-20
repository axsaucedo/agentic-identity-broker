package oauth2server

import "errors"

// ErrInvalidRedirectURI is an internal sentinel that signals the redirect_uri parameter
// failed validation before code issuance. HTTP handlers use this to comply with
// RFC 6749 §4.1.2.1: never redirect to an unverified redirect_uri.
//
// All other OAuth2 errors are represented directly by fosite.ErrXxx types.
var ErrInvalidRedirectURI = errors.New("redirect_uri is not registered or does not meet security requirements")
