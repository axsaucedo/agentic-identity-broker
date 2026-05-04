package urivalidation

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ValidateCIMDClientURL validates a URL for use as a CIMD client_id.
//
// Rules per FR-014 through FR-018:
//   - Scheme must be https
//   - Must have a non-empty path component
//   - Path must not contain dot segments (. or ..)
//   - Must not contain a fragment (#)
//   - Must not contain userinfo (credentials)
//   - Port must be 443 or absent
func ValidateCIMDClientURL(raw string) error {
	if strings.Contains(raw, "#") {
		return fmt.Errorf("client_id URL must not contain a fragment")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "https" {
		return fmt.Errorf("client_id URL must use https scheme, got %q", u.Scheme)
	}

	if u.User != nil {
		return fmt.Errorf("client_id URL must not contain credentials")
	}

	// Validate the authority explicitly: require a non-empty hostname and, when a port
	// separator is present (including empty-port trailing colons like "example.com:"),
	// parse via net.SplitHostPort to catch malformed authorities.
	// url.Parse accepts empty ports (net.SplitHostPort allows them), so u.Port() == ""
	// does not distinguish "no port" from "trailing colon with empty port" — check for
	// a colon in the host to catch the latter.
	rawPort := u.Port()
	if rawPort != "" || (!strings.HasPrefix(u.Host, "[") && strings.ContainsRune(u.Host, ':')) {
		h, p, splitErr := net.SplitHostPort(u.Host)
		if splitErr != nil || h == "" || p == "" {
			return fmt.Errorf("client_id URL has malformed authority: %q", u.Host)
		}
		n, atoiErr := strconv.Atoi(p)
		if atoiErr != nil || n < 1 || n > 65535 {
			return fmt.Errorf("client_id URL port is not a valid port number: %q", p)
		}
		if n != 443 {
			return fmt.Errorf("client_id URL port must be 443 or absent, got %q", p)
		}
	} else if u.Hostname() == "" {
		return fmt.Errorf("client_id URL must have a host")
	}

	if u.Path == "" || u.Path == "/" {
		return fmt.Errorf("client_id URL must have a non-empty path")
	}

	for seg := range strings.SplitSeq(u.Path, "/") {
		if seg == "." || seg == ".." {
			return fmt.Errorf("client_id URL path must not contain dot segments")
		}
	}

	return nil
}
