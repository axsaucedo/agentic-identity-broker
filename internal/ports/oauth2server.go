package ports

import (
	"context"

	"github.com/lestrrat-go/jwx/v3/jwk"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// SigningKeyManager is the port for signing key lifecycle operations.
// Implemented by domain/oauth2server.SigningKeyService.
type SigningKeyManager interface {
	GenerateAndStoreKey(ctx context.Context, algorithm string, makeCurrent bool) (*storage.SigningKey, error)
	BuildJWKS(ctx context.Context) (jwk.Set, error)
}

// BrokerCredentialGenerator is the port for generating broker client credentials.
// Implemented by domain/oauth2server.ClientAuthService.
type BrokerCredentialGenerator interface {
	GenerateCredentials(agentID id.AgentID) (credential *storage.BrokerClientCredential, plaintextSecret string, err error)
}
