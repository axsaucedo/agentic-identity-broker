package urivalidation

import (
	"net/url"
	"strings"
)

// IsValidRedirectURI validates a redirect URI per RFC 6749 §3.1.2:
// - Must be absolute (https for non-local, http only for loopback)
// - Must have a non-empty host
// - Must not contain a fragment
//
// HTTP is only allowed for localhost and loopback addresses (127.0.0.1, [::1]).
// All other callbacks require HTTPS.
func IsValidRedirectURI(uriStr string) bool {
	// Reject raw fragment or whitespace before parsing — url.ParseRequestURI
	// percent-encodes these instead of erroring.
	if strings.ContainsAny(uriStr, "# \t\n\r") {
		return false
	}
	u, err := url.ParseRequestURI(uriStr)
	if err != nil {
		return false
	}
	if u.Host == "" {
		return false
	}
	switch u.Scheme {
	case "https":
		return true
	case "http":
		// Allow HTTP only for loopback (development-mode callbacks)
		host := u.Hostname()
		return host == "localhost" || host == "127.0.0.1" || host == "::1"
	default:
		return false
	}
}

// MatchesRedirectURI reports whether incoming matches registered for redirect URI
// validation. For loopback hosts (localhost, 127.0.0.1, [::1]) the port is ignored
// per RFC 8252 §7.3; all other components must match exactly. For non-loopback
// hosts all four URI components must match exactly.
func MatchesRedirectURI(registered, incoming string) bool {
	if strings.ContainsAny(registered, "# \t\n\r") || strings.ContainsAny(incoming, "# \t\n\r") {
		return false
	}

	r, err := url.Parse(registered)
	if err != nil || r.Host == "" {
		return false
	}
	in, err := url.Parse(incoming)
	if err != nil || in.Host == "" {
		return false
	}
	if h := r.Hostname(); h == "localhost" || h == "127.0.0.1" || h == "::1" {
		r.Host = h
		in.Host = in.Hostname()
	}
	return r.String() == in.String()
}
