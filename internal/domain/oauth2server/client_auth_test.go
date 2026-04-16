package oauth2server

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

func bufLogger() (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelError})), buf
}

type mockCredentialRepo struct {
	getByAgentIDFunc        func(context.Context, id.AgentID) (*storage.BrokerClientCredential, error)
	getByBrokerClientIDFunc func(context.Context, id.BrokerClientID) (*storage.BrokerClientCredential, error)
}

func (m *mockCredentialRepo) Create(_ context.Context, _ *storage.BrokerClientCredential) error {
	return nil
}
func (m *mockCredentialRepo) GetByAgentID(ctx context.Context, agentID id.AgentID) (*storage.BrokerClientCredential, error) {
	if m.getByAgentIDFunc != nil {
		return m.getByAgentIDFunc(ctx, agentID)
	}
	return nil, storage.NewStorageError("mockCredentialRepo.GetByAgentID", storage.ErrorKindNotFound, nil, "not found")
}
func (m *mockCredentialRepo) GetByBrokerClientID(ctx context.Context, clientID id.BrokerClientID) (*storage.BrokerClientCredential, error) {
	if m.getByBrokerClientIDFunc != nil {
		return m.getByBrokerClientIDFunc(ctx, clientID)
	}
	return nil, nil
}
func (m *mockCredentialRepo) Delete(_ context.Context, _ id.AgentID) error { return nil }
func (m *mockCredentialRepo) Rotate(_ context.Context, _ id.AgentID, _ *storage.BrokerClientCredential) error {
	return nil
}

type mockAgentRepo struct {
	getFunc func(context.Context, id.AgentID) (*storage.Agent, error)
}

func (m *mockAgentRepo) Create(_ context.Context, _ *storage.Agent) error { return nil }
func (m *mockAgentRepo) Get(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
	return m.getFunc(ctx, agentID)
}
func (m *mockAgentRepo) Update(_ context.Context, _ *storage.Agent) error { return nil }
func (m *mockAgentRepo) Delete(_ context.Context, _ id.AgentID) error     { return nil }
func (m *mockAgentRepo) List(_ context.Context) ([]*storage.Agent, error) { return nil, nil }
func (m *mockAgentRepo) GetByClientID(_ context.Context, _ id.ClientID) (*storage.Agent, error) {
	return nil, nil
}

func TestArgon2Hasher_HashAndCompare(t *testing.T) {
	hasher := &Argon2Hasher{}

	t.Run("hash then compare succeeds", func(t *testing.T) {
		secret := "test-secret-value-1234"
		hash, err := hasher.Hash(secret)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		err = hasher.Compare(hash, secret)
		assert.NoError(t, err)
	})

	t.Run("wrong secret fails comparison", func(t *testing.T) {
		secret := "correct-secret"
		hash, err := hasher.Hash(secret)
		require.NoError(t, err)

		err = hasher.Compare(hash, "wrong-secret")
		assert.ErrorIs(t, err, ErrInvalidClient)
	})

	t.Run("PHC string format is correct", func(t *testing.T) {
		hash, err := hasher.Hash("test")
		require.NoError(t, err)

		assert.True(t, strings.HasPrefix(hash, "$argon2id$v="))
		parts := strings.Split(hash, "$")
		assert.Len(t, parts, 6) // empty, argon2id, version, params, salt, hash
		assert.Equal(t, "argon2id", parts[1])
		assert.Contains(t, parts[3], "m=65536,t=3,p=4")
	})

	t.Run("different hashes for same secret", func(t *testing.T) {
		secret := "same-secret"
		hash1, err := hasher.Hash(secret)
		require.NoError(t, err)
		hash2, err := hasher.Hash(secret)
		require.NoError(t, err)

		// Different salts should produce different hashes
		assert.NotEqual(t, hash1, hash2)

		// Both should verify against the same secret
		assert.NoError(t, hasher.Compare(hash1, secret))
		assert.NoError(t, hasher.Compare(hash2, secret))
	})

	t.Run("invalid PHC string fails", func(t *testing.T) {
		err := hasher.Compare("not-a-valid-phc-string", "secret")
		assert.Error(t, err)
	})
}

func TestClientAuthService_GenerateCredentials(t *testing.T) {
	t.Run("generates broker_client_id with correct format", func(t *testing.T) {
		svc := &ClientAuthService{hasher: &Argon2Hasher{}}
		cred, secret, err := svc.GenerateCredentials(testAgentID())
		require.NoError(t, err)
		require.NotNil(t, cred)
		require.NotEmpty(t, secret)

		// Verify broker_client_id format
		clientID := cred.BrokerClientID.String()
		assert.True(t, strings.HasPrefix(clientID, "broker_"), "client ID should start with broker_ prefix")
		assert.Greater(t, len(clientID), len("broker_"), "client ID should have random suffix")

		// Verify secret hash is not the plaintext
		assert.NotEqual(t, secret, cred.SecretHash)
		assert.True(t, strings.HasPrefix(cred.SecretHash, "$argon2id$"))
	})

	t.Run("generates unique credentials each time", func(t *testing.T) {
		svc := &ClientAuthService{hasher: &Argon2Hasher{}}
		cred1, secret1, err := svc.GenerateCredentials(testAgentID())
		require.NoError(t, err)
		cred2, secret2, err := svc.GenerateCredentials(testAgentID())
		require.NoError(t, err)

		assert.NotEqual(t, cred1.BrokerClientID, cred2.BrokerClientID)
		assert.NotEqual(t, secret1, secret2)
	})
}

func TestClientAuthService_Authenticate(t *testing.T) {
	t.Run("valid credentials succeed", func(t *testing.T) {
		credRepo := memory.NewBrokerClientCredentialStore()
		agentRepo := memory.NewAgentRepository()

		// Create test agent
		agent := testAgent()
		err := agentRepo.Create(context.Background(), agent)
		require.NoError(t, err)

		// Create test credential
		svc := NewClientAuthService(credRepo, agentRepo, testSlogger())
		cred, plaintext, err := svc.GenerateCredentials(agent.ID)
		require.NoError(t, err)
		err = credRepo.Create(context.Background(), cred)
		require.NoError(t, err)

		// Authenticate should succeed using agent UUID
		result, err := svc.Authenticate(context.Background(), agent.ID, plaintext)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, agent.ID, result.Agent.ID)
	})

	t.Run("unknown agent_id fails", func(t *testing.T) {
		credRepo := memory.NewBrokerClientCredentialStore()
		agentRepo := memory.NewAgentRepository()
		svc := NewClientAuthService(credRepo, agentRepo, testSlogger())

		_, err := svc.Authenticate(context.Background(), id.NewAgentID(), "secret")
		assert.ErrorIs(t, err, ErrInvalidClient)
	})

	t.Run("wrong secret fails", func(t *testing.T) {
		credRepo := memory.NewBrokerClientCredentialStore()
		agentRepo := memory.NewAgentRepository()

		// Create test agent + credential
		agent := testAgent()
		err := agentRepo.Create(context.Background(), agent)
		require.NoError(t, err)

		svc := NewClientAuthService(credRepo, agentRepo, testSlogger())
		cred, _, err := svc.GenerateCredentials(agent.ID)
		require.NoError(t, err)
		err = credRepo.Create(context.Background(), cred)
		require.NoError(t, err)

		// Authenticate with wrong secret should fail
		_, err = svc.Authenticate(context.Background(), agent.ID, "wrong-secret")
		assert.ErrorIs(t, err, ErrInvalidClient)
	})

	t.Run("credential repo connection error logs at error level", func(t *testing.T) {
		agentID := id.NewAgentID()
		credRepo := &mockCredentialRepo{
			getByAgentIDFunc: func(_ context.Context, _ id.AgentID) (*storage.BrokerClientCredential, error) {
				return nil, storage.NewStorageError("GetByAgentID", storage.ErrorKindConnection, nil, "connection refused")
			},
		}
		logger, buf := bufLogger()
		svc := NewClientAuthService(credRepo, memory.NewAgentRepository(), logger)

		_, err := svc.Authenticate(context.Background(), agentID, "secret")
		assert.ErrorIs(t, err, ErrInvalidClient)
		assert.Contains(t, buf.String(), "storage failure during client authentication")
	})

	t.Run("credential repo not-found does not log an error", func(t *testing.T) {
		agentID := id.NewAgentID()
		credRepo := &mockCredentialRepo{
			getByAgentIDFunc: func(_ context.Context, _ id.AgentID) (*storage.BrokerClientCredential, error) {
				return nil, storage.NewStorageError("GetByAgentID", storage.ErrorKindNotFound, nil, "not found")
			},
		}
		logger, buf := bufLogger()
		svc := NewClientAuthService(credRepo, memory.NewAgentRepository(), logger)

		_, err := svc.Authenticate(context.Background(), agentID, "secret")
		assert.ErrorIs(t, err, ErrInvalidClient)
		assert.Empty(t, buf.String(), "not-found must not produce an error log")
	})

	t.Run("agent repo timeout error logs at error level", func(t *testing.T) {
		credRepo := memory.NewBrokerClientCredentialStore()
		agentID := id.NewAgentID()
		cred := &storage.BrokerClientCredential{
			ID:             id.NewCredentialID(),
			AgentID:        agentID,
			BrokerClientID: id.NewBrokerClientID("broker_timeout_test"),
			SecretHash:     "irrelevant",
		}
		require.NoError(t, credRepo.Create(context.Background(), cred))

		agentRepo := &mockAgentRepo{
			getFunc: func(_ context.Context, _ id.AgentID) (*storage.Agent, error) {
				return nil, storage.NewStorageError("Get", storage.ErrorKindTimeout, nil, "query timeout")
			},
		}
		logger, buf := bufLogger()
		svc := NewClientAuthService(credRepo, agentRepo, logger)

		_, err := svc.Authenticate(context.Background(), agentID, "irrelevant")
		assert.ErrorIs(t, err, ErrInvalidClient)
		assert.Contains(t, buf.String(), "storage failure during client authentication")
	})

	t.Run("agent repo not-found does not log an error", func(t *testing.T) {
		credRepo := memory.NewBrokerClientCredentialStore()
		agentID := id.NewAgentID()
		cred := &storage.BrokerClientCredential{
			ID:             id.NewCredentialID(),
			AgentID:        agentID,
			BrokerClientID: id.NewBrokerClientID("broker_nf_agent_test"),
			SecretHash:     "irrelevant",
		}
		require.NoError(t, credRepo.Create(context.Background(), cred))

		agentRepo := &mockAgentRepo{
			getFunc: func(_ context.Context, _ id.AgentID) (*storage.Agent, error) {
				return nil, storage.NewStorageError("Get", storage.ErrorKindNotFound, nil, "not found")
			},
		}
		logger, buf := bufLogger()
		svc := NewClientAuthService(credRepo, agentRepo, logger)

		_, err := svc.Authenticate(context.Background(), agentID, "irrelevant")
		assert.ErrorIs(t, err, ErrInvalidClient)
		assert.Empty(t, buf.String(), "not-found on agent lookup must not produce an error log")
	})
}
