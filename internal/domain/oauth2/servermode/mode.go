// Package servermode defines the OAuth2 server operating mode as a zero-import leaf
// so that both internal/ports and internal/domain/oauth2 can import it without creating
// a cycle (domain/oauth2 already imports ports, so ports cannot import domain/oauth2).
package servermode

// Mode is the enumeration of valid OAuth2 server operating modes.
type Mode string

const (
	// Proxy forwards all requests to an upstream OAuth2 server.
	Proxy Mode = "proxy"
	// Local issues tokens locally (replaces deprecated "issue_token").
	Local Mode = "local"
	// Hybrid accepts proxy, CIMD, and local clients in a single deployment.
	Hybrid Mode = "hybrid"
)
