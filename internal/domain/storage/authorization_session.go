package storage

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

const authorizationSessionTTL = 10 * time.Minute

// CIMDMetadataSnapshot stores CIMD document fields captured at authorization time.
// It is embedded in AuthorizationSession as a trusted server-side record, used to
// build the cimd_metadata response on the consent page without re-fetching the document.
type CIMDMetadataSnapshot struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	LogoURI      string   `json:"logo_uri,omitempty"`
	RedirectURIs []string `json:"redirect_uris"`
	AuthMethod   string   `json:"auth_method,omitempty"`
	JwksURI      string   `json:"jwks_uri,omitempty"`
}

// AuthorizationSession is a short-lived server-side record of a pending CIMD-based
// authorization request. It binds the consent page to the server's trusted copy of
// the authorization context, preventing URL-parameter tampering (SR-013/SR-014).
//
// Sessions are single-use with a 10-minute TTL. Every CIMD authorization flow creates
// one session; it is consumed exactly once during consent submission (FR-028/FR-029).
type AuthorizationSession struct {
	SessionID           string       `db:"session_id"`
	AgentID             id.AgentID   `db:"agent_id"`
	Principal           id.Principal `db:"principal"`
	ClientID            string       `db:"client_id"`
	OriginalURL         string       `db:"original_url"`
	RedirectURI         string       `db:"redirect_uri"`
	Scope               string       `db:"scope"`
	State               string       `db:"state"`
	CodeChallenge       string       `db:"code_challenge"`
	CodeChallengeMethod string       `db:"code_challenge_method"`
	CIMDMetadata        *CIMDMetadataSnapshot
	CreatedAt           time.Time  `db:"created_at"`
	ExpiresAt           time.Time  `db:"expires_at"`
	ConsumedAt          *time.Time `db:"consumed_at"`
}

// NewAuthorizationSession creates a new session for a CIMD-based authorization request.
// The session ID is a 64-character lowercase hex string from 32 cryptographically random bytes
// (256 bits), per OWASP/NIST SP 800-63B recommendations for capability tokens (ADR 016).
func NewAuthorizationSession(
	agentID id.AgentID,
	principal id.Principal,
	clientID string,
	originalURL string,
	redirectURI string,
	scope string,
	state string,
	codeChallenge string,
	codeChallengeMethod string,
	meta *CIMDMetadataSnapshot,
) (*AuthorizationSession, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	now := time.Now()
	return &AuthorizationSession{
		SessionID:           hex.EncodeToString(b),
		AgentID:             agentID,
		Principal:           principal,
		ClientID:            clientID,
		OriginalURL:         originalURL,
		RedirectURI:         redirectURI,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		CIMDMetadata:        meta,
		CreatedAt:           now,
		ExpiresAt:           now.Add(authorizationSessionTTL),
	}, nil
}

// IsExpired reports whether the session TTL has elapsed.
func (s *AuthorizationSession) IsExpired() bool {
	return s.isExpiredAt(time.Now())
}

// isExpiredAt reports expiry relative to a caller-supplied now, enabling
// deterministic boundary testing without mocking the global clock.
func (s *AuthorizationSession) isExpiredAt(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

// IsConsumed reports whether the session has already been used for consent submission.
func (s *AuthorizationSession) IsConsumed() bool {
	return s.ConsumedAt != nil
}

// Consume marks the session as consumed. Idempotent: calling again is a no-op.
func (s *AuthorizationSession) Consume() {
	if s.ConsumedAt == nil {
		now := time.Now()
		s.ConsumedAt = &now
	}
}
