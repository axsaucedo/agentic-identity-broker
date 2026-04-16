package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Compile-time interface check
var _ ports.SigningKeyRepository = (*SigningKeyStore)(nil)

// SigningKeyStore is an in-memory implementation of SigningKeyRepository.
type SigningKeyStore struct {
	mu    sync.RWMutex
	byID  map[id.SigningKeyID]*storage.SigningKey
	byKID map[id.KeyID]*storage.SigningKey
}

// NewSigningKeyStore creates a new in-memory signing key store.
func NewSigningKeyStore() *SigningKeyStore {
	return &SigningKeyStore{
		byID:  make(map[id.SigningKeyID]*storage.SigningKey),
		byKID: make(map[id.KeyID]*storage.SigningKey),
	}
}

func (s *SigningKeyStore) Create(ctx context.Context, key *storage.SigningKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byKID[key.KID]; exists {
		return storage.NewStorageError("SigningKeyStore.Create", storage.ErrorKindConflict, nil,
			fmt.Sprintf("signing key with kid %s already exists", key.KID))
	}

	k := *key
	k.PrivateKeyEncrypted = append([]byte(nil), key.PrivateKeyEncrypted...)
	s.byID[k.ID] = &k
	s.byKID[k.KID] = &k
	return nil
}

func (s *SigningKeyStore) CreateAndSetCurrent(ctx context.Context, key *storage.SigningKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byKID[key.KID]; exists {
		return storage.NewStorageError("SigningKeyStore.CreateAndSetCurrent", storage.ErrorKindConflict, nil,
			fmt.Sprintf("signing key with kid %s already exists", key.KID))
	}

	// Demote all existing keys, then insert the new one as current.
	for _, existing := range s.byID {
		existing.IsCurrent = false
	}
	k := *key
	k.IsCurrent = true
	k.PrivateKeyEncrypted = append([]byte(nil), key.PrivateKeyEncrypted...)
	s.byID[k.ID] = &k
	s.byKID[k.KID] = &k
	return nil
}

func (s *SigningKeyStore) GetByKID(ctx context.Context, kid id.KeyID) (*storage.SigningKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, exists := s.byKID[kid]
	if !exists || key.RemovedAt != nil {
		return nil, storage.NewStorageError("SigningKeyStore.GetByKID", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("signing key with kid %s not found", kid))
	}
	result := *key
	result.PrivateKeyEncrypted = append([]byte(nil), key.PrivateKeyEncrypted...)
	return &result, nil
}

func (s *SigningKeyStore) GetCurrent(ctx context.Context) (*storage.SigningKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, key := range s.byID {
		if key.IsCurrent && key.RemovedAt == nil {
			result := *key
			result.PrivateKeyEncrypted = append([]byte(nil), key.PrivateKeyEncrypted...)
			return &result, nil
		}
	}
	return nil, storage.NewStorageError("SigningKeyStore.GetCurrent", storage.ErrorKindNotFound, nil, "no current signing key")
}

func (s *SigningKeyStore) ListActive(ctx context.Context) ([]*storage.SigningKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*storage.SigningKey
	for _, key := range s.byID {
		if key.RemovedAt == nil {
			k := *key
			k.PrivateKeyEncrypted = append([]byte(nil), key.PrivateKeyEncrypted...)
			result = append(result, &k)
		}
	}
	return result, nil
}

func (s *SigningKeyStore) SetCurrent(ctx context.Context, kid id.KeyID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	target, exists := s.byKID[kid]
	if !exists || target.RemovedAt != nil {
		return storage.NewStorageError("SigningKeyStore.SetCurrent", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("signing key with kid %s not found", kid))
	}

	// Demote all other keys
	for _, key := range s.byID {
		key.IsCurrent = false
	}
	target.IsCurrent = true
	return nil
}

func (s *SigningKeyStore) Delete(ctx context.Context, kid id.KeyID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.byKID[kid]
	if !exists || key.RemovedAt != nil {
		return storage.NewStorageError("SigningKeyStore.Delete", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("signing key with kid %s not found", kid))
	}

	now := time.Now()
	key.RemovedAt = &now
	return nil
}

func (s *SigningKeyStore) CountActive(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, key := range s.byID {
		if key.RemovedAt == nil {
			count++
		}
	}
	return count, nil
}
