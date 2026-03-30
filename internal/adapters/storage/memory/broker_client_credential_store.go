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
	byAgentID  map[id.AgentID]*storage.BrokerClientCredential
	byClientID map[id.BrokerClientID]*storage.BrokerClientCredential
}

// NewBrokerClientCredentialStore creates a new in-memory broker client credential store.
func NewBrokerClientCredentialStore() *BrokerClientCredentialStore {
	return &BrokerClientCredentialStore{
		byID:       make(map[id.CredentialID]*storage.BrokerClientCredential),
		byAgentID:  make(map[id.AgentID]*storage.BrokerClientCredential),
		byClientID: make(map[id.BrokerClientID]*storage.BrokerClientCredential),
	}
}

func (s *BrokerClientCredentialStore) Create(ctx context.Context, credential *storage.BrokerClientCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byAgentID[credential.AgentID]; exists {
		return storage.NewStorageError("BrokerClientCredentialStore.Create", storage.ErrorKindConflict, nil,
			fmt.Sprintf("credential already exists for agent %s", credential.AgentID))
	}
	if _, exists := s.byClientID[credential.BrokerClientID]; exists {
		return storage.NewStorageError("BrokerClientCredentialStore.Create", storage.ErrorKindConflict, nil,
			fmt.Sprintf("broker_client_id %s already exists", credential.BrokerClientID))
	}

	cred := *credential
	s.byID[cred.ID] = &cred
	s.byAgentID[cred.AgentID] = &cred
	s.byClientID[cred.BrokerClientID] = &cred
	return nil
}

func (s *BrokerClientCredentialStore) GetByAgentID(ctx context.Context, agentID id.AgentID) (*storage.BrokerClientCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cred, exists := s.byAgentID[agentID]
	if !exists {
		return nil, storage.NewStorageError("BrokerClientCredentialStore.GetByAgentID", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("no credential for agent %s", agentID))
	}
	result := *cred
	return &result, nil
}

func (s *BrokerClientCredentialStore) GetByBrokerClientID(ctx context.Context, clientID id.BrokerClientID) (*storage.BrokerClientCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cred, exists := s.byClientID[clientID]
	if !exists {
		return nil, storage.NewStorageError("BrokerClientCredentialStore.GetByBrokerClientID", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("no credential for broker_client_id %s", clientID))
	}
	result := *cred
	return &result, nil
}

func (s *BrokerClientCredentialStore) Delete(ctx context.Context, agentID id.AgentID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cred, exists := s.byAgentID[agentID]
	if !exists {
		return storage.NewStorageError("BrokerClientCredentialStore.Delete", storage.ErrorKindNotFound, nil,
			fmt.Sprintf("no credential for agent %s", agentID))
	}

	delete(s.byID, cred.ID)
	delete(s.byAgentID, cred.AgentID)
	delete(s.byClientID, cred.BrokerClientID)
	return nil
}
