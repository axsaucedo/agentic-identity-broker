package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
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
