package cimd

import (
	"fmt"
	"net"
)

// SSRFBlocklist is an immutable set of CIDR ranges blocked for CIMD fetches.
// Built at startup from RFC 6890 defaults plus any operator-configured extras.
type SSRFBlocklist struct {
	blocked []*net.IPNet
}

// defaultBlockedCIDRs contains RFC 6890 ranges that must never be fetched.
var defaultBlockedCIDRs = []string{
	"127.0.0.0/8",        // Loopback (IPv4)
	"::1/128",            // Loopback (IPv6)
	"10.0.0.0/8",         // Private (RFC 1918)
	"172.16.0.0/12",      // Private (RFC 1918)
	"192.168.0.0/16",     // Private (RFC 1918)
	"169.254.0.0/16",     // Link-local (IPv4) — includes cloud metadata 169.254.169.254
	"fe80::/10",          // Link-local (IPv6)
	"0.0.0.0/8",          // "This" network
	"100.64.0.0/10",      // Shared address (CGN)
	"192.0.0.0/24",       // IETF protocol assignments
	"192.0.2.0/24",       // Documentation (TEST-NET-1)
	"198.18.0.0/15",      // Benchmarking
	"198.51.100.0/24",    // Documentation (TEST-NET-2)
	"203.0.113.0/24",     // Documentation (TEST-NET-3)
	"240.0.0.0/4",        // Reserved
	"255.255.255.255/32", // Limited broadcast
	"fc00::/7",           // Unique local (IPv6)
	"2001:db8::/32",      // Documentation (IPv6)
}

// NewSSRFBlocklist creates an SSRFBlocklist from the RFC 6890 default ranges
// plus any operator-configured extra CIDRs. Returns an error if any CIDR is malformed.
func NewSSRFBlocklist(extraCIDRs []string) (SSRFBlocklist, error) {
	all := make([]string, 0, len(defaultBlockedCIDRs)+len(extraCIDRs))
	all = append(all, defaultBlockedCIDRs...)
	all = append(all, extraCIDRs...)

	nets := make([]*net.IPNet, 0, len(all))
	for _, cidr := range all {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return SSRFBlocklist{}, fmt.Errorf("invalid blocked CIDR %q: %w", cidr, err)
		}
		nets = append(nets, ipNet)
	}

	return SSRFBlocklist{blocked: nets}, nil
}

// Contains returns true if the given IP address falls within any blocked range.
func (b SSRFBlocklist) Contains(ip net.IP) bool {
	for _, network := range b.blocked {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
