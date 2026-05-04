package cimd

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Service orchestrates CIMD document fetch, validation, and caching.
type Service struct {
	fetcher       ports.CIMDFetcher
	cache         *CIMDCache
	nameBlocklist []string
	logger        *slog.Logger
}

// defaultReservedClientNames is always enforced regardless of operator configuration (FR-023c).
var defaultReservedClientNames = []string{
	"admin", "administrator", "system", "operator", "root", "superuser",
}

// NewService creates a new CIMD service.
// nameBlocklist is merged with defaultReservedClientNames so built-in terms are
// always enforced regardless of operator configuration (FR-023c).
func NewService(fetcher ports.CIMDFetcher, cache *CIMDCache, nameBlocklist []string, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	merged := append(append([]string(nil), defaultReservedClientNames...), nameBlocklist...)
	return &Service{
		fetcher:       fetcher,
		cache:         cache,
		nameBlocklist: merged,
		logger:        logger,
	}
}

// Resolve fetches and validates the CIMD document for the given URL and agent.
func (s *Service) Resolve(ctx context.Context, rawURL string, agent *storage.Agent) (*ClientIDMetadataDocument, error) {
	// Validate URL format
	if _, err := ParseClientIDMetadataDocumentURL(rawURL); err != nil {
		return nil, err
	}

	// RFC §3 SHOULD NOT: warn when client_id URL contains a query string
	if parsed, parseErr := url.Parse(rawURL); parseErr == nil && (parsed.RawQuery != "" || parsed.ForceQuery) {
		s.logger.WarnContext(ctx, "client_id URL contains query string (discouraged by RFC)",
			"url", rawURL)
	}

	// Cache hit
	if entry := s.cache.Get(rawURL); entry != nil {
		s.logger.DebugContext(ctx, "cimd_document_cache_hit", "url", rawURL)
		return entry.Document, nil
	}

	// Fetch from remote
	result, err := s.fetcher.Fetch(ctx, rawURL)
	if err != nil {
		var ssrfErr *ports.SSRFBlockedError
		if errors.As(err, &ssrfErr) {
			s.logger.Warn("cimd_fetch_blocked", "url", rawURL, "error", err)
		} else {
			s.logger.Warn("cimd_fetch_failed", "url", rawURL, "error", err)
		}
		return nil, err
	}

	// Parse and validate the raw response body
	doc, err := ParseDocument(result.Body, rawURL, s.nameBlocklist)
	if err != nil {
		s.logger.Warn("cimd_document_invalid", "url", rawURL, "error", err)
		return nil, err
	}

	// Brand pin check (informational — flow continues)
	if doc.ClientName != "" && agent.DisplayName != "" && doc.ClientName != agent.DisplayName {
		s.logger.Warn("cimd_brand_pin_mismatch",
			"agent_id", agent.ID,
			"agent_display_name", agent.DisplayName,
			"cimd_client_name", doc.ClientName,
		)
	}

	// Store in cache
	headers := make(http.Header)
	if result.CacheControl != "" {
		headers.Set("Cache-Control", result.CacheControl)
	}
	if result.ETag != "" {
		headers.Set("ETag", result.ETag)
	}
	if result.Expires != "" {
		headers.Set("Expires", result.Expires)
	}
	s.cache.Set(rawURL, doc, headers, time.Now())

	s.logger.Info("cimd_document_fetched",
		"url", rawURL,
		"client_name", doc.ClientName,
		"redirect_uris_count", len(doc.RedirectURIs),
	)

	return doc, nil
}
