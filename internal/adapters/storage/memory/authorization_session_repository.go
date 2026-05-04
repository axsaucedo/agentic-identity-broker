package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Compile-time interface check
var _ ports.AuthorizationSessionRepository = (*AuthorizationSessionRepository)(nil)

// AuthorizationSessionRepository is an in-memory implementation of AuthorizationSessionRepository.
type AuthorizationSessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*storage.AuthorizationSession
}

// NewAuthorizationSessionRepository creates a new in-memory authorization session repository.
func NewAuthorizationSessionRepository() *AuthorizationSessionRepository {
	return &AuthorizationSessionRepository{
		sessions: make(map[string]*storage.AuthorizationSession),
	}
}

func (r *AuthorizationSessionRepository) Create(ctx context.Context, session *storage.AuthorizationSession) error {
	if session == nil {
		return storage.NewStorageError("AuthorizationSessionRepository.Create", storage.ErrorKindValidation, nil,
			"session must not be nil")
	}
	if session.SessionID == "" {
		return storage.NewStorageError("AuthorizationSessionRepository.Create", storage.ErrorKindValidation, nil,
			"session ID must not be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.SessionID]; exists {
		return storage.NewStorageError("AuthorizationSessionRepository.Create", storage.ErrorKindConflict, nil,
			fmt.Sprintf("authorization session %s already exists", session.SessionID))
	}

	cp := *session
	cp.CIMDMetadata = deepCopyCIMDMetadata(session.CIMDMetadata)
	r.sessions[session.SessionID] = &cp
	return nil
}

func (r *AuthorizationSessionRepository) GetBySessionID(ctx context.Context, sessionID string) (*storage.AuthorizationSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, storage.NewStorageError("AuthorizationSessionRepository.GetBySessionID", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("authorization session %s not found", sessionID))
	}

	cp := *session
	cp.CIMDMetadata = deepCopyCIMDMetadata(session.CIMDMetadata)
	return &cp, nil
}

func (r *AuthorizationSessionRepository) Consume(ctx context.Context, sessionID string) error {
	return r.ConsumeIf(ctx, sessionID, nil)
}

func (r *AuthorizationSessionRepository) ConsumeIf(ctx context.Context, sessionID string, fn ports.AuthorizationSessionMutation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return storage.NewStorageError("AuthorizationSessionRepository.Consume", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("authorization session %s not found", sessionID))
	}
	if session.IsExpired() {
		return storage.NewStorageError("AuthorizationSessionRepository.Consume", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("authorization session %s has expired", sessionID))
	}
	if session.IsConsumed() {
		return storage.NewStorageError("AuthorizationSessionRepository.Consume", storage.ErrorKindConflict, nil,
			fmt.Sprintf("authorization session %s has already been consumed", sessionID))
	}

	if fn != nil {
		if err := fn(ctx); err != nil {
			return err
		}
	}

	session.Consume()
	return nil
}

func (r *AuthorizationSessionRepository) DeleteExpired(ctx context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	var count int
	for id, session := range r.sessions {
		if session.ExpiresAt.Before(now) {
			delete(r.sessions, id)
			count++
		}
	}
	return count, nil
}

func deepCopyCIMDMetadata(m *storage.CIMDMetadataSnapshot) *storage.CIMDMetadataSnapshot {
	if m == nil {
		return nil
	}
	cp := *m
	cp.RedirectURIs = append([]string(nil), m.RedirectURIs...)
	return &cp
}
