package ports

import (
	"context"
	"errors"

	"github.com/lestrrat-go/jwx/v3/jwk"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

var (
	ErrLastActiveKey = errors.New("cannot remove the last active signing key")
	ErrCurrentKey    = errors.New("cannot remove the current signing key; promote another key first")
)

// SigningKeyManager is the port for signing key lifecycle operations.
// Implemented by domain/oauth2server.SigningKeyService.
type SigningKeyManager interface {
	GenerateAndStoreKey(ctx context.Context, algorithm string, makeCurrent bool) (*storage.SigningKey, error)
	BuildJWKS(ctx context.Context) (jwk.Set, error)
	DeleteKey(ctx context.Context, kid id.KeyID) error
}

// CredentialGenerator is the port for generating broker client credentials.
// Implemented by domain/oauth2server.ClientAuthService.
type CredentialGenerator interface {
	GenerateCredentials(agentID id.AgentID) (credential *storage.ClientCredential, plaintextSecret string, err error)
}
