//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUserSessionTestDB creates an isolated migrated test database backed by the shared PostgreSQL container.
func setupUserSessionTestDB(t *testing.T) (*Adapter, func()) {
	t.Helper()
	return setupMigratedAdapter(t)
}

// insertTestService inserts a minimal thirdparty_oauth2_services row to satisfy the
// user_sessions FK constraint without going through the full provider adapter.
func insertTestService(t *testing.T, adapter *Adapter, serviceID id.ServiceID) {
	t.Helper()
	ctx := context.Background()
	_, err := adapter.db.ExecContext(ctx, `
		INSERT INTO thirdparty_oauth2_services
			(id, display_name, client_id, client_secret_encrypted, oauth2_flavor,
			 issuer_uri, enable_discovery, token_endpoint, authorize_endpoint, scopes)
		VALUES
			($1, 'Test Service', 'test-client', '\x00', 'standard',
			 'https://example.com', false, 'https://example.com/token',
			 'https://example.com/authorize', '[]')
	`, serviceID)
	require.NoError(t, err)
}

// TestUserSessionRepository_FindByPrincipalAndService_ScopesScan reproduces the
// bug where FindByPrincipalAndService fails with:
//
//	sql: Scan error on column index 8, name "scope": unsupported Scan,
//	storing driver.Value type string into type *[]string
//
// The root cause: pgx/v5/stdlib returns PostgreSQL TEXT[] as a string literal
// (e.g. "{genie}") and []string has no sql.Scanner implementation.
func TestUserSessionRepository_FindByPrincipalAndService_ScopesScan(t *testing.T) {
	adapter, cleanup := setupUserSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()
	serviceID := id.NewServiceID()
	insertTestService(t, adapter, serviceID)

	repo := NewUserSessionRepository(adapter)
	principal := id.Principal("user@example.com")

	session := &storage.UserSession{
		ID:                   id.NewSessionID(),
		Principal:            principal,
		ServiceID:            serviceID,
		EncryptedAccessToken: []byte("encrypted-token"),
		TokenType:            "Bearer",
		Scope:                []string{"genie"},
		EncryptionContext:    storage.EncryptionContext{ServiceID: serviceID},
		InitiatedAt:          time.Now().UTC(),
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err, "Create should succeed")

	// This is the operation that previously failed with:
	// "sql: Scan error on column index 8, name "scope": unsupported Scan,
	//  storing driver.Value type string into type *[]string"
	found, err := repo.FindByPrincipalAndService(ctx, principal, serviceID)
	require.NoError(t, err, "FindByPrincipalAndService must not return a scan error")
	require.NotNil(t, found)

	assert.Equal(t, []string{"genie"}, found.Scope, "scopes must round-trip through TEXT[]")
	assert.Equal(t, session.TokenType, found.TokenType)
	assert.Equal(t, session.Principal, found.Principal)
}

func TestUserSessionRepository_FindByPrincipalAndService_MultipleScopes(t *testing.T) {
	adapter, cleanup := setupUserSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()
	serviceID := id.NewServiceID()
	insertTestService(t, adapter, serviceID)

	repo := NewUserSessionRepository(adapter)
	principal := id.Principal("user@example.com")

	session := &storage.UserSession{
		ID:                   id.NewSessionID(),
		Principal:            principal,
		ServiceID:            serviceID,
		EncryptedAccessToken: []byte("encrypted-token"),
		TokenType:            "Bearer",
		Scope:                []string{"sql-warehouses", "genie", "clusters"},
		EncryptionContext:    storage.EncryptionContext{ServiceID: serviceID},
		InitiatedAt:          time.Now().UTC(),
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}

	require.NoError(t, repo.Create(ctx, session))

	found, err := repo.FindByPrincipalAndService(ctx, principal, serviceID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.ElementsMatch(t, []string{"sql-warehouses", "genie", "clusters"}, found.Scope)
}

func TestUserSessionRepository_Get_ScopesScan(t *testing.T) {
	adapter, cleanup := setupUserSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()
	serviceID := id.NewServiceID()
	insertTestService(t, adapter, serviceID)

	repo := NewUserSessionRepository(adapter)

	session := &storage.UserSession{
		ID:                   id.NewSessionID(),
		Principal:            id.Principal("user@example.com"),
		ServiceID:            serviceID,
		EncryptedAccessToken: []byte("encrypted-token"),
		TokenType:            "Bearer",
		Scope:                []string{"genie"},
		EncryptionContext:    storage.EncryptionContext{ServiceID: serviceID},
		InitiatedAt:          time.Now().UTC(),
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}

	require.NoError(t, repo.Create(ctx, session))

	found, err := repo.Get(ctx, session.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, []string{"genie"}, found.Scope)
}

func TestUserSessionRepository_ListByPrincipal_ScopesScan(t *testing.T) {
	adapter, cleanup := setupUserSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()
	serviceID := id.NewServiceID()
	insertTestService(t, adapter, serviceID)

	repo := NewUserSessionRepository(adapter)
	principal := id.Principal("user@example.com")

	session := &storage.UserSession{
		ID:                   id.NewSessionID(),
		Principal:            principal,
		ServiceID:            serviceID,
		EncryptedAccessToken: []byte("encrypted-token"),
		TokenType:            "Bearer",
		Scope:                []string{"genie"},
		EncryptionContext:    storage.EncryptionContext{ServiceID: serviceID},
		InitiatedAt:          time.Now().UTC(),
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}

	require.NoError(t, repo.Create(ctx, session))

	sessions, err := repo.ListByPrincipal(ctx, principal)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, []string{"genie"}, sessions[0].Scope)
}
