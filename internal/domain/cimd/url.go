package cimd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ClientIDMetadataDocumentURL is a validated HTTPS URL used as a client_id.
// Invalid URLs cannot be constructed — validation occurs at parse time.
type ClientIDMetadataDocumentURL struct {
	raw string
}

// ParseClientIDMetadataDocumentURL parses and validates a URL for use as a CIMD client_id.
// Validation rules per FR-014 through FR-018:
//   - Scheme must be https
//   - Must have a non-empty path component
//   - Path must not contain dot segments (. or ..)
//   - Must not contain a fragment (#)
//   - Must not contain userinfo (credentials)
//   - Port must be 443 or absent
func ParseClientIDMetadataDocumentURL(raw string) (ClientIDMetadataDocumentURL, error) {
	if strings.Contains(raw, "#") {
		return ClientIDMetadataDocumentURL{}, fmt.Errorf("client_id URL must not contain a fragment")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return ClientIDMetadataDocumentURL{}, fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "https" {
		return ClientIDMetadataDocumentURL{}, fmt.Errorf("client_id URL must use https scheme, got %q", u.Scheme)
	}

	if u.User != nil {
		return ClientIDMetadataDocumentURL{}, fmt.Errorf("client_id URL must not contain credentials")
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
			return ClientIDMetadataDocumentURL{}, fmt.Errorf("client_id URL has malformed authority: %q", u.Host)
		}
		n, atoiErr := strconv.Atoi(p)
		if atoiErr != nil || n < 1 || n > 65535 {
			return ClientIDMetadataDocumentURL{}, fmt.Errorf("client_id URL port is not a valid port number: %q", p)
		}
		if n != 443 {
			return ClientIDMetadataDocumentURL{}, fmt.Errorf("client_id URL port must be 443 or absent, got %q", p)
		}
	} else if u.Hostname() == "" {
		return ClientIDMetadataDocumentURL{}, fmt.Errorf("client_id URL must have a host")
	}

	if u.Path == "" || u.Path == "/" {
		return ClientIDMetadataDocumentURL{}, fmt.Errorf("client_id URL must have a non-empty path")
	}

	if containsDotSegments(u.Path) {
		return ClientIDMetadataDocumentURL{}, fmt.Errorf("client_id URL path must not contain dot segments")
	}

	return ClientIDMetadataDocumentURL{raw: raw}, nil
}

// String returns the raw URL string.
func (u ClientIDMetadataDocumentURL) String() string {
	return u.raw
}

// containsDotSegments returns true if the path contains . or .. segments.
func containsDotSegments(path string) bool {
	for _, segment := range strings.Split(path, "/") {
		if segment == "." || segment == ".." {
			return true
		}
	}
	return false
}
