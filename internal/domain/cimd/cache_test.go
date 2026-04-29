package cimd

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustNewCIMDCache(t *testing.T, minTTL, maxTTL time.Duration) *CIMDCache {
	t.Helper()
	c, err := NewCIMDCache(minTTL, maxTTL)
	require.NoError(t, err)
	return c
}

func TestCIMDCache(t *testing.T) {
	const testURL = "https://agent.example.com/client"
	doc := &ClientIDMetadataDocument{
		ClientID:     testURL,
		ClientName:   "Test Agent",
		RedirectURIs: []string{"https://agent.example.com/callback"},
	}

	minTTL := 60 * time.Second
	maxTTL := 1 * time.Hour

	t.Run("miss returns nil", func(t *testing.T) {
		c := mustNewCIMDCache(t, minTTL, maxTTL)
		assert.Nil(t, c.Get(testURL))
	})

	t.Run("hit returns entry within TTL", func(t *testing.T) {
		c := mustNewCIMDCache(t, minTTL, maxTTL)
		h := make(http.Header)
		h.Set("Cache-Control", "max-age=300")
		c.Set(testURL, doc, h, time.Now())

		entry := c.Get(testURL)
		require.NotNil(t, entry)
		assert.Equal(t, testURL, entry.URL)
		assert.Equal(t, doc, entry.Document)
	})

	t.Run("expired entry returns nil", func(t *testing.T) {
		c := mustNewCIMDCache(t, minTTL, maxTTL)
		h := make(http.Header)
		h.Set("Cache-Control", "max-age=300")
		fetchedAt := time.Now().Add(-10 * time.Minute) // 10 min ago
		c.Set(testURL, doc, h, fetchedAt)

		assert.Nil(t, c.Get(testURL))
	})

	t.Run("max-age from Cache-Control is used", func(t *testing.T) {
		c := mustNewCIMDCache(t, minTTL, maxTTL)
		h := make(http.Header)
		h.Set("Cache-Control", "max-age=7200") // 2 hours — clamped to maxTTL (1h)

		now := time.Now()
		c.Set(testURL, doc, h, now)

		entry := c.Get(testURL)
		require.NotNil(t, entry)
		expectedExpiry := now.Add(maxTTL)
		assert.WithinDuration(t, expectedExpiry, entry.ExpiresAt, time.Second)
	})

	t.Run("operator maxTTL clamps document max-age", func(t *testing.T) {
		c := mustNewCIMDCache(t, minTTL, 30*time.Minute)
		h := make(http.Header)
		h.Set("Cache-Control", "max-age=7200")

		now := time.Now()
		c.Set(testURL, doc, h, now)

		entry := c.Get(testURL)
		require.NotNil(t, entry)
		assert.WithinDuration(t, now.Add(30*time.Minute), entry.ExpiresAt, time.Second)
	})

	t.Run("minTTL floor applied when max-age is too short", func(t *testing.T) {
		c := mustNewCIMDCache(t, 60*time.Second, maxTTL)
		h := make(http.Header)
		h.Set("Cache-Control", "max-age=10") // 10s — below minTTL (60s)

		now := time.Now()
		c.Set(testURL, doc, h, now)

		entry := c.Get(testURL)
		require.NotNil(t, entry)
		assert.WithinDuration(t, now.Add(60*time.Second), entry.ExpiresAt, time.Second)
	})

	t.Run("minTTL applied when no Cache-Control header", func(t *testing.T) {
		c := mustNewCIMDCache(t, minTTL, maxTTL)
		now := time.Now()
		c.Set(testURL, doc, make(http.Header), now)

		entry := c.Get(testURL)
		require.NotNil(t, entry)
		assert.WithinDuration(t, now.Add(minTTL), entry.ExpiresAt, time.Second)
	})

	t.Run("expired entry is evicted from map on Get", func(t *testing.T) {
		c := mustNewCIMDCache(t, minTTL, maxTTL)
		h := make(http.Header)
		h.Set("Cache-Control", "max-age=300")
		c.Set(testURL, doc, h, time.Now().Add(-10*time.Minute))

		assert.Nil(t, c.Get(testURL))

		c.mu.RLock()
		_, exists := c.entries[testURL]
		c.mu.RUnlock()
		assert.False(t, exists, "expired entry should be removed from map")
	})

	t.Run("Get does not evict entry refreshed by concurrent Set", func(t *testing.T) {
		c := mustNewCIMDCache(t, minTTL, maxTTL)
		h := make(http.Header)
		h.Set("Cache-Control", "max-age=300")

		freshDoc := &ClientIDMetadataDocument{ClientID: testURL, ClientName: "Fresh Agent"}

		// Seed expired, then immediately refresh (simulates concurrent Set winning the window).
		c.Set(testURL, doc, h, time.Now().Add(-10*time.Minute))
		c.Set(testURL, freshDoc, h, time.Now())

		entry := c.Get(testURL)
		require.NotNil(t, entry, "fresh entry must survive after concurrent Set")
		assert.Equal(t, "Fresh Agent", entry.Document.ClientName)
	})

	t.Run("Get concurrent with Set does not race", func(t *testing.T) {
		c := mustNewCIMDCache(t, minTTL, maxTTL)
		h := make(http.Header)
		h.Set("Cache-Control", "max-age=300")

		freshDoc := &ClientIDMetadataDocument{ClientID: testURL, ClientName: "Fresh"}
		c.Set(testURL, doc, h, time.Now().Add(-10*time.Minute))

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for range 200 {
				c.Set(testURL, freshDoc, h, time.Now())
			}
		}()
		go func() {
			defer wg.Done()
			for range 200 {
				_ = c.Get(testURL)
			}
		}()
		wg.Wait()
	})

	t.Run("ETag stored from response header", func(t *testing.T) {
		c := mustNewCIMDCache(t, minTTL, maxTTL)
		h := make(http.Header)
		h.Set("Cache-Control", "max-age=300")
		h.Set("ETag", `"abc123"`)
		c.Set(testURL, doc, h, time.Now())

		entry := c.Get(testURL)
		require.NotNil(t, entry)
		assert.Equal(t, `"abc123"`, entry.ETag)
	})

	t.Run("rejects minTTL greater than maxTTL", func(t *testing.T) {
		_, err := NewCIMDCache(2*time.Hour, 30*time.Minute)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "minTTL")
	})

	t.Run("accepts equal minTTL and maxTTL", func(t *testing.T) {
		c, err := NewCIMDCache(5*time.Minute, 5*time.Minute)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})
}
