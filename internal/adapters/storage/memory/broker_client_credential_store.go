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
var _ ports.BrokerClientCredentialRepository = (*BrokerClientCredentialStore)(nil)

// BrokerClientCredentialStore is an in-memory implementation of BrokerClientCredentialRepository.
type BrokerClientCredentialStore struct {
	mu         sync.RWMutex
	byID       map[id.CredentialID]*storage.BrokerClientCredential
	byClientID map[id.ClientID]*storage.BrokerClientCredential
}

// NewBrokerClientCredentialStore creates a new in-memory broker client credential store.
func NewBrokerClientCredentialStore() *BrokerClientCredentialStore {
	return &BrokerClientCredentialStore{
		byID:       make(map[id.CredentialID]*storage.BrokerClientCredential),
		byClientID: make(map[id.ClientID]*storage.BrokerClientCredential),
	}
}

func (s *BrokerClientCredentialStore) Create(ctx context.Context, credential *storage.BrokerClientCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byClientID[credential.ClientID]; exists {
		return storage.NewStorageError("BrokerClientCredentialStore.Create", storage.ErrorKindConflict, nil,
			fmt.Sprintf("credential for client_id %s already exists", credential.ClientID))
	}

	cred := *credential
	s.byID[cred.ID] = &cred
	s.byClientID[cred.ClientID] = &cred
	return nil
}

// GetByAgentID retrieves the credential for the given agent.
// Since client_id == agentID.String(), this delegates to GetByClientID.
func (s *BrokerClientCredentialStore) GetByAgentID(ctx context.Context, agentID id.AgentID) (*storage.BrokerClientCredential, error) {
	return s.GetByClientID(ctx, id.ClientID(agentID.String()))
}

func (s *BrokerClientCredentialStore) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.BrokerClientCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cred, exists := s.byClientID[clientID]
	if !exists {
		return nil, storage.NewStorageError("BrokerClientCredentialStore.GetByClientID", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("no credential for client_id %s", clientID))
	}
	result := *cred
	return &result, nil
}

func (s *BrokerClientCredentialStore) Delete(ctx context.Context, agentID id.AgentID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientID := id.ClientID(agentID.String())
	cred, exists := s.byClientID[clientID]
	if !exists {
		return storage.NewStorageError("BrokerClientCredentialStore.Delete", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("no credential for agent %s", agentID))
	}

	delete(s.byID, cred.ID)
	delete(s.byClientID, clientID)
	return nil
}

func (s *BrokerClientCredentialStore) Rotate(ctx context.Context, agentID id.AgentID, newCredential *storage.BrokerClientCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientID := id.ClientID(agentID.String())
	old, exists := s.byClientID[clientID]
	if !exists {
		return storage.NewStorageError("BrokerClientCredentialStore.Rotate", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("no existing credential for agent %s", agentID))
	}

	newCred := *newCredential
	delete(s.byID, old.ID)
	s.byID[newCred.ID] = &newCred
	s.byClientID[clientID] = &newCred
	return nil
}
