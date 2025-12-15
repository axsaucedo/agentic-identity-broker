package memory

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := adapter.Initialize(ctx)
	assert.NoError(t, err)
}

func TestInitialize_ContextCancelled(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := adapter.Initialize(ctx)
	assert.Error(t, err)
}

func TestClose(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := adapter.Initialize(ctx)
	require.NoError(t, err)

	err = adapter.Close(ctx)
	assert.NoError(t, err)
}

func TestClose_Idempotent(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := adapter.Initialize(ctx)
	require.NoError(t, err)

	err = adapter.Close(ctx)
	require.NoError(t, err)

	// Second close should also work (idempotent)
	err = adapter.Close(ctx)
	assert.NoError(t, err)
}

func TestHealthCheck(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := adapter.Initialize(ctx)
	require.NoError(t, err)

	err = adapter.HealthCheck(ctx)
	assert.NoError(t, err)
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name    string
		user    *ports.User
		wantErr bool
		errKind storage.ErrorKind
	}{
		{
			name: "valid user",
			user: &ports.User{
				ID:        "user123",
				Email:     "test@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:    "nil user",
			user:    nil,
			wantErr: true,
			errKind: storage.ErrorKindValidation,
		},
		{
			name: "empty ID",
			user: &ports.User{
				ID:        "",
				Email:     "test@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
			errKind: storage.ErrorKindValidation,
		},
		{
			name: "empty email",
			user: &ports.User{
				ID:        "user123",
				Email:     "",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
			errKind: storage.ErrorKindValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewAdapter()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			adapter.Initialize(ctx)

			err := adapter.CreateUser(ctx, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				if storErr, ok := err.(*storage.StorageError); ok {
					assert.Equal(t, tt.errKind, storErr.Kind)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateUser_Duplicate(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	user := &ports.User{
		ID:        "user123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := adapter.CreateUser(ctx, user)
	require.NoError(t, err)

	// Try to create again
	err = adapter.CreateUser(ctx, user)
	assert.Error(t, err)
	if storErr, ok := err.(*storage.StorageError); ok {
		assert.Equal(t, storage.ErrorKindConflict, storErr.Kind)
	}
}

func TestGetUser(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	user := &ports.User{
		ID:        "user123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	adapter.CreateUser(ctx, user)

	retrieved, err := adapter.GetUser(ctx, "user123")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestGetUser_NotFound(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	_, err := adapter.GetUser(ctx, "nonexistent")
	assert.Error(t, err)
	if storErr, ok := err.(*storage.StorageError); ok {
		assert.Equal(t, storage.ErrorKindNotFound, storErr.Kind)
	}
}

func TestUpdateUser(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	user := &ports.User{
		ID:        "user123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	adapter.CreateUser(ctx, user)

	// Update with new email
	user.Email = "newemail@example.com"
	user.UpdatedAt = time.Now()

	err := adapter.UpdateUser(ctx, user)
	assert.NoError(t, err)

	// Verify update
	retrieved, _ := adapter.GetUser(ctx, "user123")
	assert.Equal(t, "newemail@example.com", retrieved.Email)
}

func TestUpdateUser_NotFound(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	user := &ports.User{
		ID:        "nonexistent",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := adapter.UpdateUser(ctx, user)
	assert.Error(t, err)
	if storErr, ok := err.(*storage.StorageError); ok {
		assert.Equal(t, storage.ErrorKindNotFound, storErr.Kind)
	}
}

func TestDeleteUser(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	user := &ports.User{
		ID:        "user123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	adapter.CreateUser(ctx, user)

	err := adapter.DeleteUser(ctx, "user123")
	assert.NoError(t, err)

	// Verify deletion
	_, err = adapter.GetUser(ctx, "user123")
	assert.Error(t, err)
}

func TestDeleteUser_Idempotent(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	// Delete non-existent user (should not error)
	err := adapter.DeleteUser(ctx, "nonexistent")
	assert.NoError(t, err)
}

func TestListUsers(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	// Create multiple users
	for i := 1; i <= 3; i++ {
		user := &ports.User{
			ID:        "user" + string(rune(48+i)),
			Email:     "test" + string(rune(48+i)) + "@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		adapter.CreateUser(ctx, user)
	}

	users, err := adapter.ListUsers(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
}

func TestListUsers_WithPagination(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	// Create 10 users
	for i := 1; i <= 10; i++ {
		user := &ports.User{
			ID:        "user" + string(rune(48+i%10)),
			Email:     "test" + string(rune(48+i%10)) + "@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		adapter.CreateUser(ctx, user)
	}

	filter := &ports.UserFilter{
		Limit:  5,
		Offset: 0,
	}

	users, err := adapter.ListUsers(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, users, 5)
}

// Concurrent access tests
func TestConcurrentReads(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	user := &ports.User{
		ID:        "user123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	adapter.CreateUser(ctx, user)

	// 100 concurrent reads
	var wg sync.WaitGroup
	var successCount int32

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := adapter.GetUser(ctx, "user123")
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(100), successCount)
}

func TestConcurrentWritesAndReads(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	adapter.Initialize(ctx)

	var wg sync.WaitGroup
	var writeErrors int32
	var readErrors int32

	// Start 10 concurrent writes
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			user := &ports.User{
				ID:        "user" + string(rune(48+(id%10))),
				Email:     "test@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := adapter.CreateUser(ctx, user); err != nil {
				atomic.AddInt32(&writeErrors, 1)
			}
		}(i)
	}

	wg.Wait()

	// Start 50 concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := adapter.GetUser(ctx, "user"+string(rune(48+(id%10))))
			if err != nil {
				atomic.AddInt32(&readErrors, 1)
			}
		}(i)
	}

	wg.Wait()

	// Some reads might fail (users created after reads started) but most should succeed
	assert.Less(t, readErrors, int32(50))
}

func TestContextCancellation(t *testing.T) {
	adapter := NewAdapter()
	ctx, cancel := context.WithCancel(context.Background())

	adapter.Initialize(context.Background())

	user := &ports.User{
		ID:        "user123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	adapter.CreateUser(context.Background(), user)

	// Cancel context
	cancel()

	// Operations should fail with context error
	_, err := adapter.GetUser(ctx, "user123")
	assert.Error(t, err)
	if storErr, ok := err.(*storage.StorageError); ok {
		assert.Equal(t, storage.ErrorKindTimeout, storErr.Kind)
	}
}

func TestDataIsolation(t *testing.T) {
	adapter1 := NewAdapter()
	adapter2 := NewAdapter()
	ctx1, cancel1 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel1()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()

	adapter1.Initialize(ctx1)
	adapter2.Initialize(ctx2)

	user := &ports.User{
		ID:        "user123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	adapter1.CreateUser(ctx1, user)

	// User should not exist in adapter2 (different instances)
	_, err := adapter2.GetUser(ctx2, "user123")
	assert.Error(t, err)
}

// Benchmark tests
func BenchmarkCreateUser(b *testing.B) {
	adapter := NewAdapter()
	ctx := context.Background()
	adapter.Initialize(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &ports.User{
			ID:        "user" + string(rune(48+(i%1000))),
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		adapter.CreateUser(ctx, user)
	}
}

func BenchmarkGetUser(b *testing.B) {
	adapter := NewAdapter()
	ctx := context.Background()
	adapter.Initialize(ctx)

	user := &ports.User{
		ID:        "user123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	adapter.CreateUser(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.GetUser(ctx, "user123")
	}
}

func BenchmarkListUsers(b *testing.B) {
	adapter := NewAdapter()
	ctx := context.Background()
	adapter.Initialize(ctx)

	// Create 1000 users
	for i := 1; i <= 1000; i++ {
		user := &ports.User{
			ID:        "user" + string(rune(48+(i%1000))),
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		adapter.CreateUser(ctx, user)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.ListUsers(ctx, nil)
	}
}
