package oauth2server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ory/fosite"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

const clientSecretBytes = 32

// ClientAuthService handles client authentication and credential management.
type ClientAuthService struct {
	credentialRepo ports.BrokerClientCredentialRepository
	agentRepo      ports.AgentRepository
	hasher         *Argon2Hasher
	logger         *slog.Logger
}

// NewClientAuthService creates a new ClientAuthService.
func NewClientAuthService(
	credentialRepo ports.BrokerClientCredentialRepository,
	agentRepo ports.AgentRepository,
	logger *slog.Logger,
) *ClientAuthService {
	return &ClientAuthService{
		credentialRepo: credentialRepo,
		agentRepo:      agentRepo,
		hasher:         &Argon2Hasher{},
		logger:         logger,
	}
}

// AuthenticatedClient represents a successfully authenticated client.
type AuthenticatedClient struct {
	Agent      *storage.Agent
	Credential *storage.BrokerClientCredential
}

// Authenticate verifies client credentials and returns the associated agent.
func (s *ClientAuthService) Authenticate(ctx context.Context, agentID id.AgentID, secret string) (*AuthenticatedClient, error) {
	cred, err := s.credentialRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		s.logStorageFailure(ctx, err)
		return nil, fosite.ErrInvalidClient
	}

	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		s.logStorageFailure(ctx, err)
		return nil, fosite.ErrInvalidClient
	}

	if err := s.hasher.Compare(cred.SecretHash, secret); err != nil {
		return nil, fosite.ErrInvalidClient
	}

	return &AuthenticatedClient{
		Agent:      agent,
		Credential: cred,
	}, nil
}

// logStorageFailure logs infrastructure-level storage errors (connection, timeout, unknown)
// while silently ignoring not-found errors which are expected during authentication.
func (s *ClientAuthService) logStorageFailure(ctx context.Context, err error) {
	var se *storage.StorageError
	if errors.As(err, &se) && se.Kind != storage.ErrorKindNotFound {
		s.logger.ErrorContext(ctx, "storage failure during client authentication",
			"storage_op", se.Operation,
			"storage_kind", string(se.Kind),
		)
	}
}

// GenerateCredentials generates a new broker client secret for the given agent.
// The client_id is always the agent's UUID string. Returns the plaintext secret
// (shown once to the user) and the credential for storage.
func (s *ClientAuthService) GenerateCredentials(agentID id.AgentID) (credential *storage.BrokerClientCredential, plaintextSecret string, err error) {
	secretBytes := make([]byte, clientSecretBytes)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate secret: %w", err)
	}
	plaintextSecret = base64.RawURLEncoding.EncodeToString(secretBytes)

	secretHash, err := s.hasher.Hash(plaintextSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash secret: %w", err)
	}

	credential = &storage.BrokerClientCredential{
		ID:         id.NewCredentialID(),
		ClientID:   id.NewClientID(agentID.String()),
		SecretHash: secretHash,
	}

	return credential, plaintextSecret, nil
}
