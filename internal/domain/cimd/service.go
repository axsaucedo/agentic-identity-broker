package cimd

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// SnapshotPersistenceError wraps a repository failure during CIMD snapshot update.
// Callers can check for this type to distinguish storage failures from document validation errors.
type SnapshotPersistenceError struct{ Err error }

func (e *SnapshotPersistenceError) Error() string { return e.Err.Error() }
func (e *SnapshotPersistenceError) Unwrap() error { return e.Err }

// Service orchestrates CIMD document fetch, validation, caching, and snapshot management.
type Service struct {
	fetcher       ports.CIMDFetcher
	cache         *CIMDCache
	agentRepo     ports.AgentRepository
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
func NewService(fetcher ports.CIMDFetcher, cache *CIMDCache, agentRepo ports.AgentRepository, nameBlocklist []string, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	merged := append(append([]string(nil), defaultReservedClientNames...), nameBlocklist...)
	return &Service{
		fetcher:       fetcher,
		cache:         cache,
		agentRepo:     agentRepo,
		nameBlocklist: merged,
		logger:        logger,
	}
}

// Resolve fetches and validates the CIMD document for the given URL and agent.
// It returns the validated document, updating the agent's snapshot fields
// (AuthMethod, JwksURI) and emitting structured audit events on security field changes.
//
// On the first fetch (when agent snapshot fields are nil), the snapshot is populated
// silently — no audit event is emitted. On subsequent fetches, changes trigger events.
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

	// Security field change detection and snapshot update.
	// Stage mutations on a copy so the caller's *Agent is only updated after a
	// successful persist — avoiding in-memory/DB divergence on write failure.
	staged := *agent
	needsUpdate := false

	if staged.AuthMethod == nil {
		if doc.AuthMethod != "" {
			m := doc.AuthMethod
			staged.AuthMethod = &m
			needsUpdate = true
		}
	} else if doc.AuthMethod != *staged.AuthMethod {
		s.logger.Warn("cimd_security_field_changed",
			"agent_id", staged.ID,
			"field", "auth_method",
			"previous", *staged.AuthMethod,
			"current", doc.AuthMethod,
			"timestamp", time.Now().UTC(),
		)
		m := doc.AuthMethod
		staged.AuthMethod = &m
		needsUpdate = true
	}

	if staged.JwksURI == nil {
		if doc.JwksURI != "" {
			j := doc.JwksURI
			staged.JwksURI = &j
			needsUpdate = true
		}
	} else if doc.JwksURI != *staged.JwksURI {
		s.logger.Warn("cimd_security_field_changed",
			"agent_id", staged.ID,
			"field", "jwks_uri",
			"previous", *staged.JwksURI,
			"current", doc.JwksURI,
			"timestamp", time.Now().UTC(),
		)
		j := doc.JwksURI
		staged.JwksURI = &j
		needsUpdate = true
	}

	if staged.CIMDClientName == nil {
		if doc.ClientName != "" {
			n := doc.ClientName
			staged.CIMDClientName = &n
			needsUpdate = true
		}
	} else if doc.ClientName != *staged.CIMDClientName {
		n := doc.ClientName
		staged.CIMDClientName = &n
		needsUpdate = true
	}

	if staged.CIMDLogoURI == nil {
		if doc.LogoURI != "" {
			n := doc.LogoURI
			staged.CIMDLogoURI = &n
			needsUpdate = true
		}
	} else if doc.LogoURI != *staged.CIMDLogoURI {
		n := doc.LogoURI
		staged.CIMDLogoURI = &n
		needsUpdate = true
	}

	if staged.CIMDRedirectURIs == nil {
		if len(doc.RedirectURIs) > 0 {
			staged.CIMDRedirectURIs = slices.Clone(doc.RedirectURIs)
			needsUpdate = true
		}
	} else {
		docSorted := slices.Sorted(slices.Values(doc.RedirectURIs))
		agentSorted := slices.Sorted(slices.Values(staged.CIMDRedirectURIs))
		if !slices.Equal(docSorted, agentSorted) {
			s.logger.Warn("cimd_security_field_changed",
				"agent_id", staged.ID,
				"field", "redirect_uris",
				"previous", staged.CIMDRedirectURIs,
				"current", doc.RedirectURIs,
				"timestamp", time.Now().UTC(),
			)
			staged.CIMDRedirectURIs = slices.Clone(doc.RedirectURIs)
			needsUpdate = true
		}
	}

	if needsUpdate {
		staged.UpdatedAt = time.Now().UTC()
		if err := s.agentRepo.Update(ctx, &staged); err != nil {
			s.logger.Error("cimd_snapshot_update_failed", "agent_id", staged.ID, "error", err)
			return nil, &SnapshotPersistenceError{Err: err}
		}
		*agent = staged
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
