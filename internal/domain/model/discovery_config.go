package model

// DiscoveryConfig holds OAuth2 endpoint discovery configuration.
// When EnableDiscovery is true, the service can automatically discover OAuth2
// endpoints from the MetadataURL (RFC 8414 / OIDC Discovery).
type DiscoveryConfig struct {
	EnableDiscovery bool
	MetadataURL     *string
}

// OAuth2Endpoints holds explicit OAuth2 endpoint URLs for a provider.
// These are used when EnableDiscovery is false in DiscoveryConfig.
type OAuth2Endpoints struct {
	TokenEndpoint     string
	AuthorizeEndpoint string
	JWKsURI           string
}
