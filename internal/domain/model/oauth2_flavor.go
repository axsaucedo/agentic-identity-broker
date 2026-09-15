package model

import (
	"fmt"
	"net/url"
	"strings"
)

// OAuth2Flavor identifies the credential format used by a ThirdpartyOAuth2ProviderEntity.
// It controls how the client_secret field is interpreted and validated.
type OAuth2Flavor string

const (
	// OAuth2FlavorStandard is the default flavor: plain client_secret string.
	// Preserves existing behavior for all currently configured services.
	OAuth2FlavorStandard OAuth2Flavor = "standard"

	// OAuth2FlavorGoogle indicates a Google service account JSON credential.
	// client_id is extracted automatically from the service account JSON.
	OAuth2FlavorGoogle OAuth2Flavor = "google"

	// OAuth2FlavorGitHub indicates a GitHub OAuth2 app.
	// GitHub returns scopes as comma-separated values in token responses.
	OAuth2FlavorGitHub OAuth2Flavor = "github"
)

// DefaultOAuth2Flavor is the zero-value flavor applied when the field is omitted in requests.
const DefaultOAuth2Flavor = OAuth2FlavorStandard

// Validate returns an error if the flavor is not a recognized value.
// Returns nil for "standard", "google", and "github".
// Returns an error for any other value, including the empty string.
func (f OAuth2Flavor) Validate() error {
	switch f {
	case OAuth2FlavorStandard, OAuth2FlavorGoogle, OAuth2FlavorGitHub:
		return nil
	default:
		return fmt.Errorf("oauth2_flavor %q is not valid; accepted values are: standard, google, github", string(f))
	}
}

// ScopeSeparator returns the delimiter used to split scope strings in token responses.
// GitHub uses comma-separated scopes; all other flavors use space-separated (RFC 6749 default).
func (f OAuth2Flavor) ScopeSeparator() string {
	if f == OAuth2FlavorGitHub {
		return ","
	}
	return " "
}

// InferOAuth2Flavor determines the flavor from the token endpoint URL.
// Returns OAuth2FlavorGitHub if the token endpoint host is exactly "github.com",
// otherwise returns DefaultOAuth2Flavor.
func InferOAuth2Flavor(tokenEndpoint string) OAuth2Flavor {
	if tokenEndpoint == "" {
		return DefaultOAuth2Flavor
	}
	parsed, err := url.Parse(tokenEndpoint)
	if err != nil {
		return DefaultOAuth2Flavor
	}
	if strings.EqualFold(parsed.Hostname(), "github.com") {
		return OAuth2FlavorGitHub
	}
	return DefaultOAuth2Flavor
}

// String returns the string representation of the flavor.
func (f OAuth2Flavor) String() string {
	return string(f)
}
