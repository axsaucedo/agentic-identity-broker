package fixtures

import (
	"context"
	"time"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	storagedomain "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// PlaceholderPermissionSetID is the stable permission set ID used by grant fixtures
// that do not require specific permission set resolution. Call SeedPlaceholderGrantData
// on the test storage before exercising paths that resolve permission set scopes.
var PlaceholderPermissionSetID = id.MustParsePermissionSetID("00000000-0000-0000-0000-000000000001")

// PlaceholderServiceID is the stable service ID paired with PlaceholderPermissionSetID
// in grant fixtures. Call SeedPlaceholderGrantData to seed the backing rows.
var PlaceholderServiceID = id.MustParseServiceID("00000000-0000-0000-0000-000000000002")

// SecondaryPlaceholderPermissionSetID is a second stable permission set ID used by
// multi-service grant fixtures. Call SeedPlaceholderGrantData to seed backing rows.
var SecondaryPlaceholderPermissionSetID = id.MustParsePermissionSetID("00000000-0000-0000-0000-000000000003")

// SecondaryPlaceholderServiceID is the stable service ID paired with SecondaryPlaceholderPermissionSetID.
var SecondaryPlaceholderServiceID = id.MustParseServiceID("00000000-0000-0000-0000-000000000004")

// placeholderEntry returns a GrantedPermissionSetEntry using placeholder IDs.
func placeholderEntry() storagedomain.GrantedPermissionSetEntry {
	return storagedomain.GrantedPermissionSetEntry{
		PermissionSetID:    PlaceholderPermissionSetID,
		IncludedServiceIDs: []id.ServiceID{PlaceholderServiceID},
	}
}

// SeedPlaceholderGrantData inserts the placeholder service, permission set, and
// their FK relationship into storage. The service must exist before the permission
// set because permission_set_service_scopes.service_id references thirdparty_oauth2_services.
// Call this in BeforeEach blocks for tests that exercise paths that resolve permission
// set scopes (e.g. token exchange, consent info).
func SeedPlaceholderGrantData(ctx context.Context, store *storageadapter.Adapter) error {
	svc := ServiceWithID(PlaceholderServiceID.String())
	// Override scopes to match what the placeholder permission set declares.
	svc.Scopes = []model.OAuthScope{{ScopeValue: "read", Description: "Read access"}}
	svc.ProtectedResources = nil
	if err := store.Services().Create(ctx, svc); err != nil {
		return err
	}

	ps := &storagedomain.PermissionSet{
		ID:          PlaceholderPermissionSetID,
		Name:        "Placeholder Permission Set",
		Description: "Test fixture permission set — seeded by SeedPlaceholderGrantData",
		ServiceScopes: []storagedomain.ServiceScope{
			{ServiceID: PlaceholderServiceID, Scopes: []string{"read"}, RequirementType: storagedomain.RequirementTypeOptional},
		},
	}
	if err := store.PermissionSets().Create(ctx, ps); err != nil {
		return err
	}

	// Seed secondary service+PS for multi-service fixtures (GrantWithMultipleServices).
	svc2 := ServiceWithID(SecondaryPlaceholderServiceID.String())
	svc2.Scopes = []model.OAuthScope{{ScopeValue: "write", Description: "Write access"}}
	svc2.ProtectedResources = nil
	if err := store.Services().Create(ctx, svc2); err != nil {
		return err
	}

	ps2 := &storagedomain.PermissionSet{
		ID:          SecondaryPlaceholderPermissionSetID,
		Name:        "Secondary Placeholder Permission Set",
		Description: "Secondary test fixture permission set — seeded by SeedPlaceholderGrantData",
		ServiceScopes: []storagedomain.ServiceScope{
			{ServiceID: SecondaryPlaceholderServiceID, Scopes: []string{"write"}, RequirementType: storagedomain.RequirementTypeOptional},
		},
	}
	return store.PermissionSets().Create(ctx, ps2)
}

// ActiveGrant returns a user grant that is currently active (not expired).
// Principal and agentID must be provided by caller.
// ValidUntil is set to 1 hour in the future.
// Note: serviceID and scopes parameters are kept for backward-compatible call sites.
// The grant uses a placeholder GrantedPermissionSetEntry to satisfy domain validation.
// Call SeedPlaceholderGrantData on the test storage when the grant will be resolved
// through paths that look up permission set scopes (e.g. token exchange).
func ActiveGrant(principal, agentID, serviceID string, scopes []string) *storagedomain.UserGrant {
	now := time.Now()
	validUntil := now.Add(1 * time.Hour)

	return &storagedomain.UserGrant{
		ID:                    id.NewGrantID(),
		Principal:             id.Principal(principal),
		AgentID:               id.MustParseAgentID(agentID),
		ValidUntil:            &validUntil,
		GrantedPermissionSets: []storagedomain.GrantedPermissionSetEntry{placeholderEntry()},
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// ExpiredGrant returns a user grant that has already expired.
// ValidUntil is set to 1 hour in the past.
func ExpiredGrant(principal, agentID, serviceID string, scopes []string) *storagedomain.UserGrant {
	now := time.Now()
	validUntil := now.Add(-1 * time.Hour)

	return &storagedomain.UserGrant{
		ID:                    id.NewGrantID(),
		Principal:             id.Principal(principal),
		AgentID:               id.MustParseAgentID(agentID),
		ValidUntil:            &validUntil,
		GrantedPermissionSets: []storagedomain.GrantedPermissionSetEntry{placeholderEntry()},
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// GrantExpiringIn returns a user grant that expires in the specified duration.
// Useful for testing edge cases like grants expiring soon.
func GrantExpiringIn(principal, agentID, serviceID string, scopes []string, duration time.Duration) *storagedomain.UserGrant {
	now := time.Now()
	validUntil := now.Add(duration)

	return &storagedomain.UserGrant{
		ID:                    id.NewGrantID(),
		Principal:             id.Principal(principal),
		AgentID:               id.MustParseAgentID(agentID),
		ValidUntil:            &validUntil,
		GrantedPermissionSets: []storagedomain.GrantedPermissionSetEntry{placeholderEntry()},
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// IndefiniteGrant returns a user grant that never expires (ValidUntil is nil).
// Indefinite grants remain active until explicitly revoked.
func IndefiniteGrant(principal, agentID, serviceID string, scopes []string) *storagedomain.UserGrant {
	now := time.Now()

	return &storagedomain.UserGrant{
		ID:                    id.NewGrantID(),
		Principal:             id.Principal(principal),
		AgentID:               id.MustParseAgentID(agentID),
		ValidUntil:            nil,
		GrantedPermissionSets: []storagedomain.GrantedPermissionSetEntry{placeholderEntry()},
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// GrantWithMultipleServices returns a user grant referencing two distinct permission sets,
// each covering a different service. Requires SeedPlaceholderGrantData to be called first
// to ensure both permission sets and their backing services exist in storage.
func GrantWithMultipleServices(principal, agentID string) *storagedomain.UserGrant {
	now := time.Now()
	validUntil := now.Add(24 * time.Hour)

	return &storagedomain.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal(principal),
		AgentID:    id.MustParseAgentID(agentID),
		ValidUntil: &validUntil,
		GrantedPermissionSets: []storagedomain.GrantedPermissionSetEntry{
			placeholderEntry(),
			{
				PermissionSetID:    SecondaryPlaceholderPermissionSetID,
				IncludedServiceIDs: []id.ServiceID{SecondaryPlaceholderServiceID},
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GrantWithService returns a user grant. The serviceID and scopes are ignored
// (kept for call-site compatibility). Use GrantedPermissionSets on the returned
// struct to associate specific permission sets.
func GrantWithService(principal, agentID, serviceID string, scopes []string) *storagedomain.UserGrant {
	now := time.Now()
	validUntil := now.Add(30 * 24 * time.Hour) // 30 days

	return &storagedomain.UserGrant{
		ID:                    id.NewGrantID(),
		Principal:             id.Principal(principal),
		AgentID:               id.MustParseAgentID(agentID),
		ValidUntil:            &validUntil,
		GrantedPermissionSets: []storagedomain.GrantedPermissionSetEntry{placeholderEntry()},
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}
