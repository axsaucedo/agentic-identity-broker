package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// InMemoryUserSessionRepository is an in-memory implementation for testing/development.
type InMemoryUserSessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*storage.UserSession // Key: session ID
	index    map[string]*storage.UserSession // Key: "{principal}#{serviceID}"
}

// NewInMemoryUserSessionRepository creates a new in-memory repository.
func NewInMemoryUserSessionRepository() ports.UserSessionRepository {
	return &InMemoryUserSessionRepository{
		sessions: make(map[string]*storage.UserSession),
		index:    make(map[string]*storage.UserSession),
	}
}

// Create creates a new user session with upsert semantics.
func (r *InMemoryUserSessionRepository) Create(ctx context.Context, session *storage.UserSession) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}
	if err := session.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := principalServiceKey(session.Principal, session.ServiceID)

	// If session exists for this principal+service, update it
	if existing, ok := r.index[key]; ok {
		// Reuse ID
		session.ID = existing.ID
		delete(r.sessions, existing.ID)
	}

	r.sessions[session.ID] = session
	r.index[key] = session
	return nil
}

// Get retrieves a session by ID.
func (r *InMemoryUserSessionRepository) Get(ctx context.Context, id string) (*storage.UserSession, error) {
	if id == "" {
		return nil, errors.New("session ID cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil, storage.NewStorageError("Get", storage.ErrorKindNotFound, nil, "session not found")
	}
	return session, nil
}

// FindByPrincipalAndService retrieves the session for a principal and service.
func (r *InMemoryUserSessionRepository) FindByPrincipalAndService(ctx context.Context, principal, serviceID string) (*storage.UserSession, error) {
	if principal == "" || serviceID == "" {
		return nil, errors.New("principal and serviceID required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	key := principalServiceKey(principal, serviceID)
	session, ok := r.index[key]
	if !ok {
		return nil, nil // Not found is not an error
	}
	return session, nil
}

// ListByPrincipal retrieves all sessions for a principal.
func (r *InMemoryUserSessionRepository) ListByPrincipal(ctx context.Context, principal string) ([]*storage.UserSession, error) {
	if principal == "" {
		return nil, errors.New("principal required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*storage.UserSession
	for _, session := range r.sessions {
		if session.Principal == principal {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

// Delete deletes a session by ID.
func (r *InMemoryUserSessionRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("session ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[id]
	if ok {
		delete(r.sessions, id)
		key := principalServiceKey(session.Principal, session.ServiceID)
		delete(r.index, key)
	}
	return nil
}

// DeleteByPrincipalAndService deletes the session for a principal and service.
func (r *InMemoryUserSessionRepository) DeleteByPrincipalAndService(ctx context.Context, principal, serviceID string) error {
	if principal == "" || serviceID == "" {
		return errors.New("principal and serviceID required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := principalServiceKey(principal, serviceID)
	session, ok := r.index[key]
	if ok {
		delete(r.sessions, session.ID)
		delete(r.index, key)
	}
	return nil
}

// CountByService counts sessions referencing a service.
func (r *InMemoryUserSessionRepository) CountByService(ctx context.Context, serviceID string) (int, error) {
	if serviceID == "" {
		return 0, errors.New("serviceID required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, session := range r.sessions {
		if session.ServiceID == serviceID {
			count++
		}
	}
	return count, nil
}

// Helper function
func principalServiceKey(principal, serviceID string) string {
	return principal + "#" + serviceID
}
