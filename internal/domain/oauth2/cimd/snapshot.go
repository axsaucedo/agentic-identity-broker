package cimd

// MetadataSnapshot stores CIMD document fields captured at authorization time.
// It is embedded in AuthorizationSessionClaims (JWE) as a trusted record, used to
// build the cimd_metadata response on the consent page without re-fetching the document.
type MetadataSnapshot struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	LogoURI      string   `json:"logo_uri,omitempty"`
	RedirectURIs []string `json:"redirect_uris"`
	AuthMethod   string   `json:"auth_method,omitempty"`
}
