package urivalidation

import (
	"net/url"
	"strconv"
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
	u, ok := parseRedirectURI(uriStr)
	if !ok {
		return false
	}

	switch u.Scheme {
	case "https":
		return true
	case "http":
		// Allow HTTP only for loopback (development-mode callbacks)
		return isLoopbackRedirectHost(u.Hostname())
	default:
		return false
	}
}

// MatchesRedirectURI reports whether incoming matches registered for redirect URI
// validation. For loopback hosts (localhost, 127.0.0.1, [::1]) the port is ignored
// per RFC 8252 §7.3; all other components must match exactly. For non-loopback
// hosts all four URI components must match exactly.
func MatchesRedirectURI(registered, incoming string) bool {
	r, ok := parseRedirectURI(registered)
	if !ok {
		return false
	}
	in, ok := parseRedirectURI(incoming)
	if !ok {
		return false
	}
	if h := r.Hostname(); isLoopbackRedirectHost(h) {
		r.Host = normalizeRedirectHost(h)
		in.Host = normalizeRedirectHost(in.Hostname())
	}
	return r.String() == in.String()
}

func parseRedirectURI(uriStr string) (*url.URL, bool) {
	// Reject raw fragment or whitespace before parsing — url.ParseRequestURI
	// percent-encodes these instead of erroring.
	if strings.ContainsAny(uriStr, "# \t\n\r") {
		return nil, false
	}

	u, err := url.ParseRequestURI(uriStr)
	if err != nil {
		return nil, false
	}
	if !hasValidRedirectAuthority(u) {
		return nil, false
	}
	return u, true
}

func hasValidRedirectAuthority(u *url.URL) bool {
	if u.Host == "" || u.Hostname() == "" {
		return false
	}
	if strings.HasSuffix(u.Host, ":") {
		return false
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return false
		}
	}
	return true
}

func isLoopbackRedirectHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func normalizeRedirectHost(host string) string {
	if strings.Contains(host, ":") {
		return "[" + host + "]"
	}
	return host
}
