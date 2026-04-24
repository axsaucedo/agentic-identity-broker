package cimd

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Service orchestrates CIMD document fetch, validation, caching, and snapshot management.
type Service struct {
	fetcher       ports.CIMDFetcher
	cache         *CIMDCache
	agentRepo     ports.AgentRepository
	nameBlocklist []string
	logger        *slog.Logger
}

// NewService creates a new CIMD service.
func NewService(fetcher ports.CIMDFetcher, cache *CIMDCache, agentRepo ports.AgentRepository, nameBlocklist []string, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		fetcher:       fetcher,
		cache:         cache,
		agentRepo:     agentRepo,
		nameBlocklist: nameBlocklist,
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

	// Cache hit
	if entry := s.cache.Get(rawURL); entry != nil {
		return entry.Document, nil
	}

	// Fetch from remote
	result, err := s.fetcher.Fetch(ctx, rawURL)
	if err != nil {
		if strings.Contains(err.Error(), "SSRF protection") {
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
		s.logger.Info("cimd_brand_pin_mismatch",
			"agent_id", agent.ID,
			"agent_display_name", agent.DisplayName,
			"cimd_client_name", doc.ClientName,
		)
	}

	// Security field change detection and snapshot update
	needsUpdate := false

	if agent.AuthMethod == nil {
		// First fetch — populate baseline silently
		if doc.AuthMethod != "" {
			m := doc.AuthMethod
			agent.AuthMethod = &m
			needsUpdate = true
		}
	} else if doc.AuthMethod != *agent.AuthMethod {
		s.logger.Warn("cimd_security_field_changed",
			"agent_id", agent.ID,
			"field", "auth_method",
			"previous", *agent.AuthMethod,
			"current", doc.AuthMethod,
			"timestamp", time.Now().UTC(),
		)
		m := doc.AuthMethod
		agent.AuthMethod = &m
		needsUpdate = true
	}

	if agent.JwksURI == nil {
		if doc.JwksURI != "" {
			j := doc.JwksURI
			agent.JwksURI = &j
			needsUpdate = true
		}
	} else if doc.JwksURI != *agent.JwksURI {
		s.logger.Warn("cimd_security_field_changed",
			"agent_id", agent.ID,
			"field", "jwks_uri",
			"previous", *agent.JwksURI,
			"current", doc.JwksURI,
			"timestamp", time.Now().UTC(),
		)
		j := doc.JwksURI
		agent.JwksURI = &j
		needsUpdate = true
	}

	if agent.CIMDClientName == nil {
		if doc.ClientName != "" {
			n := doc.ClientName
			agent.CIMDClientName = &n
			needsUpdate = true
		}
	} else if doc.ClientName != *agent.CIMDClientName {
		n := doc.ClientName
		agent.CIMDClientName = &n
		needsUpdate = true
	}

	if agent.CIMDLogoURI == nil {
		if doc.LogoURI != "" {
			n := doc.LogoURI
			agent.CIMDLogoURI = &n
			needsUpdate = true
		}
	} else if doc.LogoURI != *agent.CIMDLogoURI {
		n := doc.LogoURI
		agent.CIMDLogoURI = &n
		needsUpdate = true
	}

	if agent.CIMDRedirectURIs == nil {
		// First fetch — populate baseline silently
		if len(doc.RedirectURIs) > 0 {
			agent.CIMDRedirectURIs = slices.Clone(doc.RedirectURIs)
			needsUpdate = true
		}
	} else {
		docSorted := slices.Sorted(slices.Values(doc.RedirectURIs))
		agentSorted := slices.Sorted(slices.Values(agent.CIMDRedirectURIs))
		if !slices.Equal(docSorted, agentSorted) {
			s.logger.Warn("cimd_security_field_changed",
				"agent_id", agent.ID,
				"field", "redirect_uris",
				"previous", agent.CIMDRedirectURIs,
				"current", doc.RedirectURIs,
				"timestamp", time.Now().UTC(),
			)
			agent.CIMDRedirectURIs = slices.Clone(doc.RedirectURIs)
			needsUpdate = true
		}
	}

	if needsUpdate {
		agent.UpdatedAt = time.Now().UTC()
		if err := s.agentRepo.Update(ctx, agent); err != nil {
			s.logger.Warn("cimd_snapshot_update_failed", "agent_id", agent.ID, "error", err)
			// Non-fatal: cache the document even if the snapshot update fails
		}
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
