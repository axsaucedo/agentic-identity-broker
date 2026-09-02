package impersonation

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Target is the registered broker agent selected by the impersonation audience suffix.
type Target struct {
	Agent *storage.Agent
}

// ValidateAudiencePrefix validates the routing-only audience prefix.
func ValidateAudiencePrefix(prefix string) error {
	u, err := url.Parse(prefix)
	if err != nil || !u.IsAbs() || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.ForceQuery || u.RawQuery != "" || u.Fragment != "" || strings.HasSuffix(u.Path, "/") {
		return fmt.Errorf("audience_prefix must be an absolute HTTP(S) URI with host and no userinfo, query, fragment, or trailing slash")
	}
	return nil
}

// ResolveTarget determines whether an audience activates impersonation and resolves its target.
func (s *Service) ResolveTarget(ctx context.Context, audiences []string) (*Target, bool, error) {
	if len(audiences) != 1 {
		return nil, false, nil
	}
	if audiences[0] == s.audiencePrefix {
		return nil, true, invalidRequest("audience target must be a canonical agent ID", "audience_target_invalid")
	}
	if !strings.HasPrefix(audiences[0], s.audiencePrefix+"/") {
		return nil, false, nil
	}

	suffix := audiences[0][len(s.audiencePrefix)+1:]
	targetID, err := id.ParseAgentID(suffix)
	if err != nil || targetID.String() != suffix {
		return nil, true, invalidRequest("audience target must be a canonical agent ID", "audience_target_invalid")
	}

	agent, err := s.agents.Get(ctx, targetID)
	if err != nil {
		if ports.IsNotFoundErr(err) {
			return nil, true, tokenexchange.NewInvalidTargetErrorWithDetails("audience target agent was not found", "target_agent_not_found")
		}
		return nil, true, serverError("audience target agent lookup failed", "target_agent_lookup_failed")
	}
	if agent == nil || agent.ID.IsZero() {
		return nil, true, serverError("audience target agent lookup returned an invalid agent", "target_agent_lookup_failed")
	}
	return &Target{Agent: agent}, true, nil
}
