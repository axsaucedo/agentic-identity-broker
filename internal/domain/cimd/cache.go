package cimd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CIMDCacheEntry holds a cached CIMD document with its expiry metadata.
type CIMDCacheEntry struct {
	URL       string
	Document  *ClientIDMetadataDocument
	FetchedAt time.Time
	ExpiresAt time.Time
	ETag      string
}

// CIMDCache is a thread-safe in-process cache for CIMD documents.
// TTL is derived from HTTP cache headers clamped to operator-configured bounds.
type CIMDCache struct {
	mu      sync.RWMutex
	entries map[string]*CIMDCacheEntry
	minTTL  time.Duration
	maxTTL  time.Duration
}

// NewCIMDCache creates a new CIMD cache with the given TTL bounds.
// Returns an error if minTTL exceeds maxTTL.
func NewCIMDCache(minTTL, maxTTL time.Duration) (*CIMDCache, error) {
	if minTTL > maxTTL {
		return nil, fmt.Errorf("cimd cache: minTTL (%v) must not exceed maxTTL (%v)", minTTL, maxTTL)
	}
	return &CIMDCache{
		entries: make(map[string]*CIMDCacheEntry),
		minTTL:  minTTL,
		maxTTL:  maxTTL,
	}, nil
}

// Get returns the cached entry for url if it exists and has not expired.
func (c *CIMDCache) Get(url string) *CIMDCacheEntry {
	c.mu.RLock()
	entry, ok := c.entries[url]
	c.mu.RUnlock()

	if !ok {
		return nil
	}
	if time.Now().After(entry.ExpiresAt) {
		c.mu.Lock()
		current, ok := c.entries[url]
		if ok && !time.Now().After(current.ExpiresAt) {
			cp := deepCopyEntry(current)
			c.mu.Unlock()
			return cp
		}
		if ok {
			delete(c.entries, url)
		}
		c.mu.Unlock()
		return nil
	}
	return deepCopyEntry(entry)
}

// deepCopyDocument returns a deep copy of doc with independent slice fields.
func deepCopyDocument(doc *ClientIDMetadataDocument) *ClientIDMetadataDocument {
	if doc == nil {
		return nil
	}
	cp := *doc
	cp.RedirectURIs = append([]string(nil), doc.RedirectURIs...)
	cp.GrantTypes = append([]string(nil), doc.GrantTypes...)
	cp.ResponseTypes = append([]string(nil), doc.ResponseTypes...)
	return &cp
}

// deepCopyEntry returns a copy of entry with the Document field deep-copied so
// callers cannot mutate cached slice fields (RedirectURIs, GrantTypes, ResponseTypes).
func deepCopyEntry(entry *CIMDCacheEntry) *CIMDCacheEntry {
	cp := *entry
	cp.Document = deepCopyDocument(entry.Document)
	return &cp
}

// Set stores a document in the cache for the given URL, deriving TTL from headers
// clamped to [minTTL, maxTTL].
func (c *CIMDCache) Set(url string, doc *ClientIDMetadataDocument, headers http.Header, fetchedAt time.Time) {
	ttl := c.deriveTTL(headers)
	entry := &CIMDCacheEntry{
		URL:       url,
		Document:  deepCopyDocument(doc),
		FetchedAt: fetchedAt,
		ExpiresAt: fetchedAt.Add(ttl),
		ETag:      headers.Get("ETag"),
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[url] = entry
}

// deriveTTL extracts TTL from Cache-Control max-age or Expires headers,
// clamped to [minTTL, maxTTL]. max-age takes precedence over Expires.
func (c *CIMDCache) deriveTTL(headers http.Header) time.Duration {
	ttl := c.minTTL

	if cc := headers.Get("Cache-Control"); cc != "" {
		if maxAge := extractMaxAge(cc); maxAge > 0 {
			ttl = time.Duration(maxAge) * time.Second
		}
	} else if expires := headers.Get("Expires"); expires != "" {
		if t, err := http.ParseTime(expires); err == nil && t.After(time.Now()) {
			ttl = time.Until(t)
		}
	}

	if ttl < c.minTTL {
		ttl = c.minTTL
	}
	if ttl > c.maxTTL {
		ttl = c.maxTTL
	}
	return ttl
}

// extractMaxAge parses the max-age directive from a Cache-Control header value.
func extractMaxAge(cc string) int64 {
	for _, part := range strings.Split(cc, ",") {
		part = strings.TrimSpace(part)
		if after, ok := strings.CutPrefix(part, "max-age="); ok {
			if v, err := strconv.ParseInt(strings.TrimSpace(after), 10, 64); err == nil && v >= 0 {
				return v
			}
		}
	}
	return 0
}
