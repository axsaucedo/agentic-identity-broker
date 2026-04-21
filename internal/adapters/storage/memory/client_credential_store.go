package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Compile-time interface check
var _ ports.ClientCredentialRepository = (*ClientCredentialStore)(nil)

// ClientCredentialStore is an in-memory implementation of ClientCredentialRepository.
type ClientCredentialStore struct {
	mu        sync.RWMutex
	byID      map[id.CredentialID]*storage.ClientCredential
	byAgentID map[id.AgentID]*storage.ClientCredential
}

// NewClientCredentialStore creates a new in-memory broker client credential store.
func NewClientCredentialStore() *ClientCredentialStore {
	return &ClientCredentialStore{
		byID:      make(map[id.CredentialID]*storage.ClientCredential),
		byAgentID: make(map[id.AgentID]*storage.ClientCredential),
	}
}

func (s *ClientCredentialStore) Create(ctx context.Context, credential *storage.ClientCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byAgentID[credential.AgentID]; exists {
		return storage.NewStorageError("ClientCredentialStore.Create", storage.ErrorKindConflict, nil,
			fmt.Sprintf("credential for agent %s already exists", credential.AgentID))
	}

	cred := *credential
	s.byID[cred.ID] = &cred
	s.byAgentID[cred.AgentID] = &cred
	return nil
}

func (s *ClientCredentialStore) GetByAgentID(ctx context.Context, agentID id.AgentID) (*storage.ClientCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cred, exists := s.byAgentID[agentID]
	if !exists {
		return nil, storage.NewStorageError("ClientCredentialStore.GetByAgentID", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("no credential for agent %s", agentID))
	}
	result := *cred
	return &result, nil
}

// GetByClientID looks up credentials by the OAuth2 client_id string.
// Since client_id equals the agent UUID, this parses the string and delegates to GetByAgentID.
func (s *ClientCredentialStore) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.ClientCredential, error) {
	agentID, err := id.ParseAgentID(clientID.String())
	if err != nil {
		return nil, storage.NewStorageError("ClientCredentialStore.GetByClientID", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("invalid client_id %s", clientID))
	}
	return s.GetByAgentID(ctx, agentID)
}

func (s *ClientCredentialStore) Delete(ctx context.Context, agentID id.AgentID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cred, exists := s.byAgentID[agentID]
	if !exists {
		return storage.NewStorageError("ClientCredentialStore.Delete", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("no credential for agent %s", agentID))
	}

	delete(s.byID, cred.ID)
	delete(s.byAgentID, agentID)
	return nil
}

func (s *ClientCredentialStore) Rotate(ctx context.Context, agentID id.AgentID, newCredential *storage.ClientCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	old, exists := s.byAgentID[agentID]
	if !exists {
		return storage.NewStorageError("ClientCredentialStore.Rotate", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("no existing credential for agent %s", agentID))
	}

	newCred := *newCredential
	delete(s.byID, old.ID)
	s.byID[newCred.ID] = &newCred
	s.byAgentID[agentID] = &newCred
	return nil
}
