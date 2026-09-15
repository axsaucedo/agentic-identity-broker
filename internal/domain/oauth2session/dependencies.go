// Package oauth2session provides OAuth2 session management domain logic.
// This file ensures required dependencies are tracked in go.mod during Phase 1 setup.
package oauth2session

import (
	_ "github.com/lestrrat-go/jwx/v4/jwe" // JWE encryption for state tokens
	_ "golang.org/x/oauth2"               // OAuth2 client library
)
