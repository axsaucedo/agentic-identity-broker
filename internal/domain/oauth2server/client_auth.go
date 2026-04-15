package oauth2server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

const (
	// brokerClientIDPrefix is prepended to generated client IDs.
	brokerClientIDPrefix = "broker_"
	// brokerClientIDRandomBytes is the number of random bytes for client ID generation.
	brokerClientIDRandomBytes = 22
	// clientSecretBytes is the number of random bytes for client secret generation.
	clientSecretBytes = 32
)

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
func (s *ClientAuthService) Authenticate(ctx context.Context, clientID id.BrokerClientID, secret string) (*AuthenticatedClient, error) {
	cred, err := s.credentialRepo.GetByBrokerClientID(ctx, clientID)
	if err != nil {
		if !isNotFoundError(err) {
			s.logger.ErrorContext(ctx, "infrastructure error looking up client credential", "error", err)
		}
		return nil, ErrInvalidClient
	}

	agent, err := s.agentRepo.Get(ctx, cred.AgentID)
	if err != nil {
		if !isNotFoundError(err) {
			s.logger.ErrorContext(ctx, "infrastructure error looking up agent for client", "agent_id", cred.AgentID, "error", err)
		}
		return nil, ErrInvalidClient
	}

	if err := s.hasher.Compare(cred.SecretHash, secret); err != nil {
		return nil, ErrInvalidClient
	}

	return &AuthenticatedClient{
		Agent:      agent,
		Credential: cred,
	}, nil
}

// GenerateCredentials generates a new broker client ID and secret.
// Returns the plaintext secret (shown once to the user) and the credential for storage.
func (s *ClientAuthService) GenerateCredentials(agentID id.AgentID) (credential *storage.BrokerClientCredential, plaintextSecret string, err error) {
	// Generate broker_client_id: "broker_" + 22 chars base64url
	randomBytes := make([]byte, brokerClientIDRandomBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate client ID: %w", err)
	}
	brokerClientID := brokerClientIDPrefix + base64.RawURLEncoding.EncodeToString(randomBytes)

	// Generate secret: 32 bytes base64url
	secretBytes := make([]byte, clientSecretBytes)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate secret: %w", err)
	}
	plaintextSecret = base64.RawURLEncoding.EncodeToString(secretBytes)

	// Hash the secret
	secretHash, err := s.hasher.Hash(plaintextSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash secret: %w", err)
	}

	credential = &storage.BrokerClientCredential{
		ID:             id.NewCredentialID(),
		AgentID:        agentID,
		BrokerClientID: id.NewBrokerClientID(brokerClientID),
		SecretHash:     secretHash,
	}

	return credential, plaintextSecret, nil
}
