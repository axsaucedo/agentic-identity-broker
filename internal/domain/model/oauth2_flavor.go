package model

import "fmt"

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
)

// DefaultOAuth2Flavor is the zero-value flavor applied when the field is omitted in requests.
const DefaultOAuth2Flavor = OAuth2FlavorStandard

// Validate returns an error if the flavor is not a recognized value.
// Returns nil for "standard" and "google".
// Returns an error for any other value, including the empty string.
func (f OAuth2Flavor) Validate() error {
	switch f {
	case OAuth2FlavorStandard, OAuth2FlavorGoogle:
		return nil
	default:
		return fmt.Errorf("oauth2_flavor %q is not valid; accepted values are: standard, google", string(f))
	}
}

// String returns the string representation of the flavor.
func (f OAuth2Flavor) String() string {
	return string(f)
}
