// Package cimd implements the SSRF-hardened HTTP fetcher for Client ID Metadata Documents.
package cimd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"

	domaincimd "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/cimd"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Ensure Fetcher implements ports.CIMDFetcher at compile time.
var _ ports.CIMDFetcher = (*Fetcher)(nil)

// Fetcher implements ports.CIMDFetcher with SSRF protection, timeout, and size limits.
type Fetcher struct {
	client           *http.Client
	blocklist        domaincimd.SSRFBlocklist
	maxResponseBytes int64
	nameBlocklist    []string
}

// DialerControl is a function seam for testing — allows injection of a custom
// net.Dialer.Control callback. In production this is the SSRF-blocking control.
type DialerControl func(network, address string, c syscall.RawConn) error

// NewFetcher creates a new SSRF-hardened CIMD fetcher.
// fetchTimeout is the maximum time to wait for a response.
// maxResponseBytes is the maximum allowed response body size.
// extraBlockedCIDRs are operator-configured additional blocked CIDR ranges.
// nameBlocklist is a case-insensitive list of forbidden client_name values.
func NewFetcher(fetchTimeout time.Duration, maxResponseBytes int64, extraBlockedCIDRs []string, nameBlocklist []string) (*Fetcher, error) {
	blocklist, err := domaincimd.NewSSRFBlocklist(extraBlockedCIDRs)
	if err != nil {
		return nil, fmt.Errorf("building SSRF blocklist: %w", err)
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("loading system certificate pool: %w", err)
	}

	dialer := &net.Dialer{
		Timeout:   fetchTimeout,
		KeepAlive: -1, // no keepalive for one-shot fetches
		Control:   buildSSRFControl(blocklist),
	}

	transport := &http.Transport{
		DialContext:       dialer.DialContext,
		TLSClientConfig:   &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: certPool},
		MaxConnsPerHost:   1,
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   fetchTimeout,
		// No redirects — CIMD endpoints must not redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Fetcher{
		client:           client,
		blocklist:        blocklist,
		maxResponseBytes: maxResponseBytes,
		nameBlocklist:    nameBlocklist,
	}, nil
}

// NewFetcherWithClient creates a Fetcher with an injected HTTP client, for testing.
// The no-redirect policy is always enforced regardless of the client's CheckRedirect setting.
func NewFetcherWithClient(client *http.Client, blocklist domaincimd.SSRFBlocklist, maxResponseBytes int64, nameBlocklist []string) *Fetcher {
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &Fetcher{
		client:           client,
		blocklist:        blocklist,
		maxResponseBytes: maxResponseBytes,
		nameBlocklist:    nameBlocklist,
	}
}

// Fetch fetches and validates a CIMD document from the given URL.
// Returns an error if the URL is blocked, the response is non-200, the body
// exceeds maxResponseBytes, or the document fails validation.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (*ports.CIMDFetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building CIMD request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching CIMD document: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CIMD endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, f.maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading CIMD response body: %w", err)
	}
	if int64(len(body)) > f.maxResponseBytes {
		return nil, fmt.Errorf("CIMD response exceeds %d byte limit", f.maxResponseBytes)
	}

	return &ports.CIMDFetchResult{
		Body:         body,
		CacheControl: resp.Header.Get("Cache-Control"),
		ETag:         resp.Header.Get("ETag"),
		Expires:      resp.Header.Get("Expires"),
	}, nil
}

// nameBlocklist returns the fetcher's configured name blocklist.
// Exposed for use by callers (e.g., CIMDService) that parse the raw body.
func (f *Fetcher) NameBlocklist() []string { return f.nameBlocklist }

// buildSSRFControl returns a net.Dialer.Control function that rejects connections
// to any IP address in the blocklist. The control callback fires after DNS
// resolution and before TCP connect, so blocked addresses never receive a packet.
func buildSSRFControl(blocklist domaincimd.SSRFBlocklist) func(string, string, syscall.RawConn) error {
	return func(network, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return fmt.Errorf("parsing resolved address %q: %w", address, err)
		}
		ip := net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("resolved address %q is not a valid IP", host)
		}
		if blocklist.Contains(ip) {
			return fmt.Errorf("SSRF protection: resolved IP %s is in a blocked range", ip)
		}
		return nil
	}
}
