package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
)

func TestThirdpartyServiceRepository_Create(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify service was stored
	retrieved, err := repo.Get(ctx, "test-service-1")
	if err != nil {
		t.Fatalf("expected no error retrieving service, got: %v", err)
	}

	if retrieved.DisplayName != "Test Service" {
		t.Errorf("expected display_name=Test Service, got=%s", retrieved.DisplayName)
	}
}

func TestThirdpartyServiceRepository_Create_GeneratesID(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	service := &storage.ThirdpartyOAuth2Service{
		// No ID provided
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if service.ID == "" {
		t.Fatal("expected ID to be generated")
	}
}

func TestThirdpartyServiceRepository_Create_DuplicateID(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error on first create, got: %v", err)
	}

	// Try to create with same ID
	err = repo.Create(ctx, service)
	if err == nil {
		t.Fatal("expected error for duplicate ID")
	}

	storageErr, ok := err.(*storage.StorageError)
	if !ok {
		t.Fatal("expected StorageError")
	}

	if storageErr.Kind != storage.ErrorKindConflict {
		t.Errorf("expected ErrorKindConflict, got: %s", storageErr.Kind)
	}
}

func TestThirdpartyServiceRepository_Get(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error creating service, got: %v", err)
	}

	retrieved, err := repo.Get(ctx, "test-service-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if retrieved.ID != "test-service-1" {
		t.Errorf("expected ID=test-service-1, got=%s", retrieved.ID)
	}
	if retrieved.DisplayName != "Test Service" {
		t.Errorf("expected display_name=Test Service, got=%s", retrieved.DisplayName)
	}
	if retrieved.ClientSecret != "test-secret" {
		t.Errorf("expected client_secret=test-secret, got=%s", retrieved.ClientSecret)
	}
}

func TestThirdpartyServiceRepository_Get_NotFound(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	_, err := repo.Get(ctx, "non-existent")
	if err == nil {
		t.Fatal("expected error for non-existent service")
	}

	storageErr, ok := err.(*storage.StorageError)
	if !ok {
		t.Fatal("expected StorageError")
	}

	if storageErr.Kind != storage.ErrorKindNotFound {
		t.Errorf("expected ErrorKindNotFound, got: %s", storageErr.Kind)
	}
}

func TestThirdpartyServiceRepository_Update(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error creating service, got: %v", err)
	}

	// Update service
	service.DisplayName = "Updated Service"
	service.ClientSecret = "new-secret"
	err = repo.Update(ctx, service)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify update
	retrieved, err := repo.Get(ctx, "test-service-1")
	if err != nil {
		t.Fatalf("expected no error retrieving service, got: %v", err)
	}

	if retrieved.DisplayName != "Updated Service" {
		t.Errorf("expected display_name=Updated Service, got=%s", retrieved.DisplayName)
	}
	if retrieved.ClientSecret != "new-secret" {
		t.Errorf("expected client_secret=new-secret, got=%s", retrieved.ClientSecret)
	}
}

func TestThirdpartyServiceRepository_Update_NotFound(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "non-existent",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Update(ctx, service)
	if err == nil {
		t.Fatal("expected error for non-existent service")
	}

	storageErr, ok := err.(*storage.StorageError)
	if !ok {
		t.Fatal("expected StorageError")
	}

	if storageErr.Kind != storage.ErrorKindNotFound {
		t.Errorf("expected ErrorKindNotFound, got: %s", storageErr.Kind)
	}
}

func TestThirdpartyServiceRepository_Delete(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error creating service, got: %v", err)
	}

	err = repo.Delete(ctx, "test-service-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify service was deleted
	_, err = repo.Get(ctx, "test-service-1")
	if err == nil {
		t.Fatal("expected error for deleted service")
	}
}

func TestThirdpartyServiceRepository_Delete_Idempotent(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	// Delete non-existent service (should not error)
	err := repo.Delete(ctx, "non-existent")
	if err != nil {
		t.Fatalf("expected no error for idempotent delete, got: %v", err)
	}
}

func TestThirdpartyServiceRepository_List(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	// List empty repository
	services, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(services) != 0 {
		t.Errorf("expected 0 services, got %d", len(services))
	}

	// Add services
	for i := 1; i <= 3; i++ {
		service := &storage.ThirdpartyOAuth2Service{
			ID:           "test-service-" + string(rune(i)),
			DisplayName:  "Test Service",
			ClientID:     "test-client-id",
			ClientSecret: "test-secret",
			IssuerURI:    "https://oauth.example.com",
			Discovery: storage.DiscoveryConfig{
				EnableDiscovery: false,
			},
			Endpoints: storage.OAuth2Endpoints{
				TokenEndpoint:     "https://oauth.example.com/token",
				AuthorizeEndpoint: "https://oauth.example.com/authorize",
			},
			Scopes: []storage.OAuthScope{
				{ScopeValue: "read", Description: "Read access"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(ctx, service)
		if err != nil {
			t.Fatalf("expected no error creating service, got: %v", err)
		}
	}

	services, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(services) != 3 {
		t.Errorf("expected 3 services, got %d", len(services))
	}
}

func TestThirdpartyServiceRepository_CountGrantsReferencingService(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	// In-memory implementation always returns 0
	count, err := repo.CountGrantsReferencingService(ctx, "test-service-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count=0, got=%d", count)
	}
}

func TestThirdpartyServiceRepository_DeepCopyProtection(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error creating service, got: %v", err)
	}

	// Get service and modify it
	retrieved, err := repo.Get(ctx, "test-service-1")
	if err != nil {
		t.Fatalf("expected no error getting service, got: %v", err)
	}
	retrieved.DisplayName = "Modified"
	retrieved.Scopes[0].ScopeValue = "write"

	// Verify original is unchanged
	original, _ := repo.Get(ctx, "test-service-1")
	if original.DisplayName != "Test Service" {
		t.Errorf("external modification affected stored service: expected Test Service, got %s", original.DisplayName)
	}
	if original.Scopes[0].ScopeValue != "read" {
		t.Errorf("external modification affected stored scopes: expected read, got %s", original.Scopes[0].ScopeValue)
	}
}

func TestThirdpartyServiceRepository_ConcurrentAccess(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			service := &storage.ThirdpartyOAuth2Service{
				ID:           "test-service-" + string(rune(id)),
				DisplayName:  "Test Service",
				ClientID:     "test-client-id-" + string(rune(id)),
				ClientSecret: "test-secret",
				IssuerURI:    "https://oauth.example.com",
				Discovery: storage.DiscoveryConfig{
					EnableDiscovery: false,
				},
				Endpoints: storage.OAuth2Endpoints{
					TokenEndpoint:     "https://oauth.example.com/token",
					AuthorizeEndpoint: "https://oauth.example.com/authorize",
				},
				Scopes: []storage.OAuthScope{
					{ScopeValue: "read", Description: "Read access"},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_ = repo.Create(ctx, service)
		}(i)
	}

	wg.Wait()

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repo.List(ctx)
		}()
	}

	wg.Wait()
}

func TestThirdpartyServiceRepository_FindByProtectedResource_Single(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	// Create service with protected resources
	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-1",
		DisplayName:  "GitHub",
		ClientID:     "github-client",
		ClientSecret: "github-secret",
		IssuerURI:    "https://github.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/login/oauth/access_token",
			AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "repo", Description: "Full repository access"},
		},
		ProtectedResources: []string{
			"https://github.com/org/repo",
			"https://github.com/user/project",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error creating service, got: %v", err)
	}

	// Find by first resource
	found, err := repo.FindByProtectedResource(ctx, "https://github.com/org/repo")
	if err != nil {
		t.Fatalf("expected no error finding service, got: %v", err)
	}

	if found.ID != "service-1" {
		t.Errorf("expected service-1, got %s", found.ID)
	}

	// Find by second resource
	found, err = repo.FindByProtectedResource(ctx, "https://github.com/user/project")
	if err != nil {
		t.Fatalf("expected no error finding service, got: %v", err)
	}

	if found.ID != "service-1" {
		t.Errorf("expected service-1, got %s", found.ID)
	}
}

func TestThirdpartyServiceRepository_FindByProtectedResource_NotFound(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	// Create service with protected resources
	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-1",
		DisplayName:  "GitHub",
		ClientID:     "github-client",
		ClientSecret: "github-secret",
		IssuerURI:    "https://github.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/login/oauth/access_token",
			AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "repo", Description: "Full repository access"},
		},
		ProtectedResources: []string{
			"https://github.com/org/repo",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error creating service, got: %v", err)
	}

	// Try to find non-existent resource
	_, err = repo.FindByProtectedResource(ctx, "https://example.com/resource")
	if err == nil {
		t.Fatal("expected error for non-existent resource")
	}

	// Verify it's an InvalidTargetError
	txErr, ok := err.(*tokenexchange.TokenExchangeError)
	if !ok {
		t.Fatalf("expected TokenExchangeError, got %T: %v", err, err)
	}

	if txErr.Code() != "invalid_target" {
		t.Errorf("expected invalid_target, got %s", txErr.Code())
	}

	if txErr.HTTPStatus() != 400 {
		t.Errorf("expected status 400, got %d", txErr.HTTPStatus())
	}
}

func TestThirdpartyServiceRepository_FindByProtectedResource_Ambiguous(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	// Create two services with overlapping resources
	service1 := &storage.ThirdpartyOAuth2Service{
		ID:           "service-1",
		DisplayName:  "GitHub 1",
		ClientID:     "github-client-1",
		ClientSecret: "github-secret-1",
		IssuerURI:    "https://github.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/login/oauth/access_token",
			AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "repo", Description: "Full repository access"},
		},
		ProtectedResources: []string{
			"https://github.com/org/repo",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	service2 := &storage.ThirdpartyOAuth2Service{
		ID:           "service-2",
		DisplayName:  "GitHub 2",
		ClientID:     "github-client-2",
		ClientSecret: "github-secret-2",
		IssuerURI:    "https://github.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/login/oauth/access_token",
			AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "repo", Description: "Full repository access"},
		},
		ProtectedResources: []string{
			"https://github.com/org/repo", // Same resource - conflict!
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service1)
	if err != nil {
		t.Fatalf("expected no error creating service 1, got: %v", err)
	}

	err = repo.Create(ctx, service2)
	if err != nil {
		t.Fatalf("expected no error creating service 2, got: %v", err)
	}

	// Try to find ambiguous resource
	_, err = repo.FindByProtectedResource(ctx, "https://github.com/org/repo")
	if err == nil {
		t.Fatal("expected error for ambiguous resource")
	}

	// Verify it's an InvalidTargetError
	txErr, ok := err.(*tokenexchange.TokenExchangeError)
	if !ok {
		t.Fatalf("expected TokenExchangeError, got %T: %v", err, err)
	}

	if txErr.Code() != "invalid_target" {
		t.Errorf("expected invalid_target, got %s", txErr.Code())
	}

	if txErr.HTTPStatus() != 400 {
		t.Errorf("expected status 400, got %d", txErr.HTTPStatus())
	}
}

func TestThirdpartyServiceRepository_FindByProtectedResource_CaseSensitive(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	// Create service
	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-1",
		DisplayName:  "GitHub",
		ClientID:     "github-client",
		ClientSecret: "github-secret",
		IssuerURI:    "https://github.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/login/oauth/access_token",
			AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "repo", Description: "Full repository access"},
		},
		ProtectedResources: []string{
			"https://github.com/org/Repo",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error creating service, got: %v", err)
	}

	// Find with exact case
	found, err := repo.FindByProtectedResource(ctx, "https://github.com/org/Repo")
	if err != nil {
		t.Fatalf("expected no error with exact case, got: %v", err)
	}
	if found.ID != "service-1" {
		t.Errorf("expected to find service with exact case")
	}

	// Try to find with different case (should fail - case-sensitive)
	_, err = repo.FindByProtectedResource(ctx, "https://github.com/org/repo")
	if err == nil {
		t.Fatal("expected error for different case")
	}

	txErr, ok := err.(*tokenexchange.TokenExchangeError)
	if !ok {
		t.Fatalf("expected TokenExchangeError, got %T", err)
	}

	if txErr.Code() != "invalid_target" {
		t.Errorf("expected invalid_target, got %s", txErr.Code())
	}
}

func TestThirdpartyServiceRepository_FindByProtectedResource_Empty(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	// Try to find with empty resource URI
	_, err := repo.FindByProtectedResource(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty resource URI")
	}

	txErr, ok := err.(*tokenexchange.TokenExchangeError)
	if !ok {
		t.Fatalf("expected TokenExchangeError, got %T", err)
	}

	if txErr.Code() != "invalid_target" {
		t.Errorf("expected invalid_target, got %s", txErr.Code())
	}
}

func TestThirdpartyServiceRepository_FindByProtectedResource_ContextCancelled(t *testing.T) {
	repo := NewThirdpartyServiceRepository()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Try to find with cancelled context
	_, err := repo.FindByProtectedResource(ctx, "https://github.com/org/repo")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}

	// Verify it's a StorageError for timeout
	storageErr, ok := err.(*storage.StorageError)
	if !ok {
		t.Fatalf("expected StorageError, got %T: %v", err, err)
	}

	if storageErr.Kind != storage.ErrorKindTimeout {
		t.Errorf("expected ErrorKindTimeout, got %s", storageErr.Kind)
	}
}

func TestThirdpartyServiceRepository_FindByProtectedResource_ThreadSafe(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	// Create services with different resources
	for i := 0; i < 3; i++ {
		service := &storage.ThirdpartyOAuth2Service{
			ID:           "service-" + string(rune(i+48)),
			DisplayName:  "Service " + string(rune(i+48)),
			ClientID:     "client-" + string(rune(i+48)),
			ClientSecret: "secret",
			IssuerURI:    "https://oauth.example.com",
			Discovery: storage.DiscoveryConfig{
				EnableDiscovery: false,
			},
			Endpoints: storage.OAuth2Endpoints{
				TokenEndpoint:     "https://oauth.example.com/token",
				AuthorizeEndpoint: "https://oauth.example.com/authorize",
			},
			Scopes: []storage.OAuthScope{
				{ScopeValue: "read", Description: "Read access"},
			},
			ProtectedResources: []string{
				"https://example.com/resource-" + string(rune(i+48)),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = repo.Create(ctx, service)
	}

	// Concurrent reads
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			wg.Add(1)
			go func(serviceNum int) {
				defer wg.Done()
				resource := "https://example.com/resource-" + string(rune(serviceNum+48))
				found, err := repo.FindByProtectedResource(ctx, resource)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if found == nil {
					t.Error("expected to find service")
				}
			}(i)
		}
	}

	wg.Wait()
}

func TestThirdpartyServiceRepository_FindByProtectedResource_WithNormalize(t *testing.T) {
	repo := NewThirdpartyServiceRepository()
	ctx := context.Background()

	// Create service with protected resources (stored without trailing slashes)
	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-1",
		DisplayName:  "API Service",
		ClientID:     "api-client",
		ClientSecret: "api-secret",
		IssuerURI:    "https://api.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://api.example.com/oauth/token",
			AuthorizeEndpoint: "https://api.example.com/oauth/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "api", Description: "API access"},
		},
		ProtectedResources: []string{
			"https://api.example.com",
			"https://api.example.com/v2",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, service)
	if err != nil {
		t.Fatalf("expected no error creating service, got: %v", err)
	}

	tests := []struct {
		name        string
		requestURI  string // URI that may have trailing slash
		expectFound bool
	}{
		{
			name:        "exact match without trailing slash",
			requestURI:  "https://api.example.com",
			expectFound: true,
		},
		{
			name:        "match after normalizing trailing slash",
			requestURI:  "https://api.example.com/",
			expectFound: true,
		},
		{
			name:        "path match without trailing slash",
			requestURI:  "https://api.example.com/v2",
			expectFound: true,
		},
		{
			name:        "path match after normalizing trailing slash",
			requestURI:  "https://api.example.com/v2/",
			expectFound: true,
		},
		{
			name:        "no match",
			requestURI:  "https://api.example.com/v3",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Demonstrate the pattern: normalize before calling FindByProtectedResource
			normalizedURI := tokenexchange.Normalize(tt.requestURI)
			found, err := repo.FindByProtectedResource(ctx, normalizedURI)

			if tt.expectFound {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				if found == nil {
					t.Error("expected to find service")
				} else if found.ID != "service-1" {
					t.Errorf("expected service-1, got %s", found.ID)
				}
			} else {
				if err == nil {
					t.Error("expected error for non-matching resource")
				}
				if found != nil {
					t.Errorf("expected no service to be found")
				}
			}
		})
	}
}
