package storage

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

// Agent represents an AI agent registered in the identity broker.
// An agent can request delegated permissions from users to access third-party services.
type Agent struct {
	ID                   id.AgentID           `json:"id" db:"id"`
	ClientID             id.ClientID          `json:"client_id" db:"client_id"`
	ExternalID           *id.ExternalID       `json:"external_id,omitempty" db:"external_id"`
	DisplayName          string               `json:"display_name" db:"display_name"`
	Description          string               `json:"description" db:"description"`
	GovernanceURL        *string              `json:"governance_url,omitempty" db:"governance_url"`
	UserDocumentationURL *string              `json:"user_documentation_url,omitempty" db:"user_documentation_url"`
	AgentInterfaceURL    *string              `json:"agent_interface_url,omitempty" db:"agent_interface_url"`
	ServiceRequirements  []ServiceRequirement `json:"service_requirements,omitempty" db:"service_requirements"`
	RedirectURIs         []string             `json:"redirect_uris" db:"redirect_uris"`
	AllowedScopes        []string             `json:"allowed_scopes" db:"allowed_scopes"`
	// ClientURIs holds pre-registered Client ID Metadata Document URLs.
	// Each entry must be a valid HTTPS URL, globally unique across all agents.
	ClientURIs []string  `json:"client_uris,omitempty" db:"-"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Validate performs validation on the Agent entity.
// Returns an error if any validation rules are violated.
func (a *Agent) Validate() error {
	// Required fields
	if a.ID.IsZero() {
		return errors.New("agent ID cannot be empty")
	}
	if a.ClientID.IsZero() {
		return errors.New("client_id is required")
	}
	if a.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if len(a.DisplayName) > 255 {
		return errors.New("display_name exceeds 255 characters")
	}
	if a.Description == "" {
		return errors.New("description is required")
	}
	if len(a.Description) > 1000 {
		return errors.New("description exceeds 1000 characters")
	}

	// URL validation (SR-011: prevent injection attacks)
	if a.GovernanceURL != nil && !isValidURL(*a.GovernanceURL) {
		return errors.New("governance_url is not a valid HTTP/HTTPS URL")
	}
	if a.UserDocumentationURL != nil && !isValidURL(*a.UserDocumentationURL) {
		return errors.New("user_documentation_url is not a valid HTTP/HTTPS URL")
	}
	if a.AgentInterfaceURL != nil && !isValidURL(*a.AgentInterfaceURL) {
		return errors.New("agent_interface_url is not a valid HTTP/HTTPS URL")
	}

	// Redirect URI validation (stricter than isValidURL: requires non-empty host, no fragment)
	for i, uri := range a.RedirectURIs {
		if !IsValidRedirectURI(uri) {
			return fmt.Errorf("redirect_uris[%d] is not a valid absolute HTTP/HTTPS URI without a fragment", i)
		}
	}

	if err := validateClientURIFormats(a.ClientURIs); err != nil {
		return err
	}

	// Service requirements validation
	if err := a.ValidateServiceRequirements(); err != nil {
		return fmt.Errorf("service_requirements validation failed: %w", err)
	}

	return nil
}

// ValidateClientURIsForWrite runs the full client URI validation (format and duplicates).
// Used by admin mutation paths (create and update).
func ValidateClientURIsForWrite(uris []string) error {
	return validateClientURIs(uris)
}

// validateClientURIFormats validates the format and uniqueness of each entry in the list.
func validateClientURIFormats(uris []string) error {
	seen := make(map[string]struct{}, len(uris))
	for i, uriStr := range uris {
		if _, dup := seen[uriStr]; dup {
			return fmt.Errorf("client_uris[%d] is a duplicate: %q", i, uriStr)
		}
		seen[uriStr] = struct{}{}
		if err := validateClientURI(uriStr); err != nil {
			return fmt.Errorf("client_uris[%d]: %w", i, err)
		}
	}
	return nil
}

// validateClientURIs validates CIMD client URIs for write-time paths (ValidateForCreate,
// admin UpdateAgent). Mirrors cimd.ParseClientIDMetadataDocumentURL — a direct import
// would create a circular dependency since cimd imports storage.
func validateClientURIs(uris []string) error {
	return validateClientURIFormats(uris)
}

func validateClientURI(uriStr string) error {
	if strings.Contains(uriStr, "#") {
		return errors.New("must not contain a fragment")
	}
	u, err := url.Parse(uriStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("must use https scheme, got %q", u.Scheme)
	}
	if u.User != nil {
		return errors.New("must not contain credentials")
	}
	// Validate the authority explicitly: mirrors cimd.ParseClientIDMetadataDocumentURL.
	// A circular dependency prevents a shared import, so both validators must stay in sync.
	// Check for a port separator including empty-port trailing colons ("example.com:"),
	// which url.Parse accepts but are structurally invalid per the validation rules.
	rawPort := u.Port()
	if rawPort != "" || (!strings.HasPrefix(u.Host, "[") && strings.ContainsRune(u.Host, ':')) {
		h, p, splitErr := net.SplitHostPort(u.Host)
		if splitErr != nil || h == "" || p == "" {
			return fmt.Errorf("has malformed authority: %q", u.Host)
		}
		n, atoiErr := strconv.Atoi(p)
		if atoiErr != nil || n < 1 || n > 65535 {
			return fmt.Errorf("port is not a valid port number: %q", p)
		}
		if n != 443 {
			return fmt.Errorf("port must be 443 or absent, got %q", p)
		}
	} else if u.Hostname() == "" {
		return errors.New("must have a host")
	}
	if u.Path == "" || u.Path == "/" {
		return errors.New("must have a non-empty path")
	}
	for _, seg := range strings.Split(u.Path, "/") {
		if seg == "." || seg == ".." {
			return errors.New("path must not contain dot segments")
		}
	}
	return nil
}

// isValidURL validates that a string is a valid HTTP or HTTPS URL.
func isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// IsValidRedirectURI validates a redirect URI per RFC 6749 §3.1.2:
// - Must be absolute (https for non-local, http only for loopback)
// - Must have a non-empty host
// - Must not contain a fragment
//
// HTTP is only allowed for localhost and loopback addresses (127.0.0.1, [::1])
// to support development. All other callbacks require HTTPS.
func IsValidRedirectURI(uriStr string) bool {
	// Reject raw fragment or whitespace before parsing — url.ParseRequestURI
	// percent-encodes these instead of erroring.
	if strings.ContainsAny(uriStr, "# \t\n\r") {
		return false
	}
	u, err := url.ParseRequestURI(uriStr)
	if err != nil {
		return false
	}
	if u.Host == "" {
		return false
	}
	switch u.Scheme {
	case "https":
		return true
	case "http":
		// Allow HTTP only for loopback (development-mode callbacks)
		host := u.Hostname()
		return host == "localhost" || host == "127.0.0.1" || host == "::1"
	default:
		return false
	}
}

// Copy creates a deep copy of the Agent to prevent external mutation.
func (a *Agent) Copy() *Agent {
	if a == nil {
		return nil
	}

	copy := &Agent{
		ID:          a.ID,
		ClientID:    a.ClientID,
		DisplayName: a.DisplayName,
		Description: a.Description,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}

	if a.ExternalID != nil {
		externalID := *a.ExternalID
		copy.ExternalID = &externalID
	}
	if a.GovernanceURL != nil {
		governanceURL := *a.GovernanceURL
		copy.GovernanceURL = &governanceURL
	}
	if a.UserDocumentationURL != nil {
		userDocURL := *a.UserDocumentationURL
		copy.UserDocumentationURL = &userDocURL
	}
	if a.AgentInterfaceURL != nil {
		agentURL := *a.AgentInterfaceURL
		copy.AgentInterfaceURL = &agentURL
	}

	// Deep copy service requirements
	if a.ServiceRequirements != nil {
		copy.ServiceRequirements = make([]ServiceRequirement, len(a.ServiceRequirements))
		for i, sr := range a.ServiceRequirements {
			copy.ServiceRequirements[i] = ServiceRequirement{
				ServiceID:       sr.ServiceID,
				RequirementType: sr.RequirementType,
				RequiredScopes:  append([]string(nil), sr.RequiredScopes...),
			}
		}
	}

	// Deep copy redirect URIs and allowed scopes
	if a.RedirectURIs != nil {
		copy.RedirectURIs = append([]string(nil), a.RedirectURIs...)
	}
	if a.AllowedScopes != nil {
		copy.AllowedScopes = append([]string(nil), a.AllowedScopes...)
	}

	if a.ClientURIs != nil {
		copy.ClientURIs = append([]string(nil), a.ClientURIs...)
	}

	return copy
}

// ValidateForCreate validates an agent before creation.
// ID will be generated, so it may be empty.
func (a *Agent) ValidateForCreate() error {
	if a.ClientID.IsZero() {
		return errors.New("client_id is required")
	}
	if a.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if len(a.DisplayName) > 255 {
		return fmt.Errorf("display_name exceeds 255 characters (got %d)", len(a.DisplayName))
	}
	if a.Description == "" {
		return errors.New("description is required")
	}
	if len(a.Description) > 1000 {
		return fmt.Errorf("description exceeds 1000 characters (got %d)", len(a.Description))
	}

	// URL validation
	if a.GovernanceURL != nil && !isValidURL(*a.GovernanceURL) {
		return errors.New("governance_url is not a valid HTTP/HTTPS URL")
	}
	if a.UserDocumentationURL != nil && !isValidURL(*a.UserDocumentationURL) {
		return errors.New("user_documentation_url is not a valid HTTP/HTTPS URL")
	}
	if a.AgentInterfaceURL != nil && !isValidURL(*a.AgentInterfaceURL) {
		return errors.New("agent_interface_url is not a valid HTTP/HTTPS URL")
	}

	// Redirect URI validation (stricter than isValidURL: requires non-empty host, no fragment)
	for i, uri := range a.RedirectURIs {
		if !IsValidRedirectURI(uri) {
			return fmt.Errorf("redirect_uris[%d] is not a valid absolute HTTP/HTTPS URI without a fragment", i)
		}
	}

	if err := validateClientURIs(a.ClientURIs); err != nil {
		return err
	}

	// Service requirements validation
	if err := a.ValidateServiceRequirements(); err != nil {
		return fmt.Errorf("service_requirements validation failed: %w", err)
	}

	return nil
}

// ValidateServiceRequirements validates the service requirements array.
// Returns error if:
// - Any service requirement is invalid
// - Duplicate service_id exists in the array
func (a *Agent) ValidateServiceRequirements() error {
	if len(a.ServiceRequirements) == 0 {
		return nil // Empty array is valid (no requirements)
	}

	// Validate each service requirement
	for i, sr := range a.ServiceRequirements {
		if err := sr.Validate(); err != nil {
			return fmt.Errorf("service_requirements[%d] invalid: %w", i, err)
		}
	}

	// Check for duplicate service_id (domain invariant)
	seen := make(map[id.ServiceID]int)
	for i, sr := range a.ServiceRequirements {
		if prevIdx, exists := seen[sr.ServiceID]; exists {
			return fmt.Errorf("duplicate service_id %q found at indices %d and %d", sr.ServiceID, prevIdx, i)
		}
		seen[sr.ServiceID] = i
	}

	return nil
}
