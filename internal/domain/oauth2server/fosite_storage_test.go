package oauth2server

import (
	"context"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	dstorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

type mockCodeRepo struct {
	findByCodeHashFunc func(context.Context, string) (*dstorage.AuthorizationCode, error)
}

func (m *mockCodeRepo) Create(_ context.Context, _ *dstorage.AuthorizationCode) error { return nil }
func (m *mockCodeRepo) FindByCodeHash(ctx context.Context, hash string) (*dstorage.AuthorizationCode, error) {
	return m.findByCodeHashFunc(ctx, hash)
}
func (m *mockCodeRepo) MarkUsed(_ context.Context, _ id.AuthorizationCodeID) error { return nil }
func (m *mockCodeRepo) DeleteExpired(_ context.Context) (int, error)               { return 0, nil }

func newTestFositeStorage() (*FositeStorage, *memory.AuthorizationCodeStore, *memory.AgentRepository, *memory.BrokerClientCredentialStore) {
	codeRepo := memory.NewAuthorizationCodeStore()
	agentRepo := memory.NewAgentRepository()
	credRepo := memory.NewBrokerClientCredentialStore()
	return NewFositeStorage(codeRepo, agentRepo, credRepo, testSlogger()), codeRepo, agentRepo, credRepo
}

func TestFositeStorage_AuthorizeCodeSessions(t *testing.T) {
	t.Run("create and retrieve code session", func(t *testing.T) {
		store, _, agentRepo, credRepo := newTestFositeStorage()
		ctx := context.Background()

		// Setup: create agent and credential
		agent := testAgent()
		err := agentRepo.Create(ctx, agent)
		require.NoError(t, err)

		cred := &dstorage.BrokerClientCredential{
			ID:             id.NewCredentialID(),
			AgentID:        agent.ID,
			BrokerClientID: id.NewBrokerClientID("broker_test_client"),
			SecretHash:     "hash",
		}
		err = credRepo.Create(ctx, cred)
		require.NoError(t, err)

		// Create a code session
		client := &brokerClient{agent: agent, credential: cred}
		session := &fosite.DefaultSession{
			Subject: "user@example.com",
			ExpiresAt: map[fosite.TokenType]time.Time{
				fosite.AuthorizeCode: time.Now().Add(60 * time.Second),
			},
		}
		req := &fosite.Request{
			ID:             "req-1",
			Client:         client,
			Session:        session,
			RequestedScope: fosite.Arguments{"read", "write"},
			GrantedScope:   fosite.Arguments{"read", "write"},
			Form: map[string][]string{
				"redirect_uri":   {"http://localhost:8080/callback"},
				"code_challenge": {"challenge123"},
			},
			RequestedAt: time.Now(),
		}

		code := "test-code-value"
		err = store.CreateAuthorizeCodeSession(ctx, code, req)
		require.NoError(t, err)

		// Retrieve
		retrieved, err := store.GetAuthorizeCodeSession(ctx, code, session)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, "user@example.com", retrieved.GetSession().GetSubject())
	})

	t.Run("invalidate code session marks as used", func(t *testing.T) {
		store, _, agentRepo, credRepo := newTestFositeStorage()
		ctx := context.Background()

		// Setup
		agent := testAgent()
		err := agentRepo.Create(ctx, agent)
		require.NoError(t, err)

		cred := &dstorage.BrokerClientCredential{
			ID:             id.NewCredentialID(),
			AgentID:        agent.ID,
			BrokerClientID: id.NewBrokerClientID("broker_test_client2"),
			SecretHash:     "hash",
		}
		err = credRepo.Create(ctx, cred)
		require.NoError(t, err)

		client := &brokerClient{agent: agent, credential: cred}
		session := &fosite.DefaultSession{
			Subject: "user@example.com",
			ExpiresAt: map[fosite.TokenType]time.Time{
				fosite.AuthorizeCode: time.Now().Add(60 * time.Second),
			},
		}
		req := &fosite.Request{
			ID:      "req-2",
			Client:  client,
			Session: session,
			Form: map[string][]string{
				"redirect_uri": {"http://localhost:8080/callback"},
			},
			RequestedAt: time.Now(),
		}

		code := "test-code-value-2"
		err = store.CreateAuthorizeCodeSession(ctx, code, req)
		require.NoError(t, err)

		// Invalidate
		err = store.InvalidateAuthorizeCodeSession(ctx, code)
		require.NoError(t, err)

		// Should return error for invalidated code
		_, err = store.GetAuthorizeCodeSession(ctx, code, session)
		assert.Error(t, err)
	})

	t.Run("non-existent code returns not found", func(t *testing.T) {
		store, _, _, _ := newTestFositeStorage()
		_, err := store.GetAuthorizeCodeSession(context.Background(), "non-existent", &fosite.DefaultSession{})
		assert.Error(t, err)
	})
}

func TestFositeStorage_AccessTokenSessions(t *testing.T) {
	t.Run("create access token session is no-op for stateless JWT", func(t *testing.T) {
		store, _, _, _ := newTestFositeStorage()
		err := store.CreateAccessTokenSession(context.Background(), "sig", nil)
		assert.NoError(t, err)
	})

	t.Run("get access token session returns not found for stateless JWT", func(t *testing.T) {
		store, _, _, _ := newTestFositeStorage()
		_, err := store.GetAccessTokenSession(context.Background(), "sig", nil)
		assert.Error(t, err)
	})

	t.Run("delete access token session is no-op for stateless JWT", func(t *testing.T) {
		store, _, _, _ := newTestFositeStorage()
		err := store.DeleteAccessTokenSession(context.Background(), "sig")
		assert.NoError(t, err)
	})
}

func TestFositeStorage_RefreshTokenSessions(t *testing.T) {
	t.Run("all refresh token operations are no-op", func(t *testing.T) {
		store, _, _, _ := newTestFositeStorage()
		ctx := context.Background()

		assert.NoError(t, store.CreateRefreshTokenSession(ctx, "sig", "access", nil))
		assert.NoError(t, store.DeleteRefreshTokenSession(ctx, "sig"))
		assert.NoError(t, store.RotateRefreshToken(ctx, "id", "sig"))
		assert.NoError(t, store.RevokeAccessToken(ctx, "sig"))

		_, err := store.GetRefreshTokenSession(ctx, "sig", nil)
		assert.Error(t, err)
	})
}

func TestFositeStorage_PKCESessions(t *testing.T) {
	t.Run("PKCE create and delete are no-op", func(t *testing.T) {
		store, _, _, _ := newTestFositeStorage()
		ctx := context.Background()

		assert.NoError(t, store.CreatePKCERequestSession(ctx, "sig", nil))
		assert.NoError(t, store.DeletePKCERequestSession(ctx, "sig"))
	})
}

func TestExtractAgentID(t *testing.T) {
	t.Run("returns agent ID for brokerClient", func(t *testing.T) {
		agent := testAgent()
		bc := &brokerClient{agent: agent}
		assert.Equal(t, agent.ID, extractAgentID(bc))
	})

	t.Run("panics for unexpected client type", func(t *testing.T) {
		assert.Panics(t, func() {
			extractAgentID(&fosite.DefaultClient{ID: "unexpected"})
		})
	})
}

func TestFositeStorage_InfrastructureErrors(t *testing.T) {
	connectionErr := dstorage.NewStorageError("FindByCodeHash", dstorage.ErrorKindConnection, nil, "connection refused")
	timeoutErr := dstorage.NewStorageError("GetByBrokerClientID", dstorage.ErrorKindTimeout, nil, "query timeout")

	t.Run("GetAuthorizeCodeSession connection error is not ErrNotFound", func(t *testing.T) {
		codeRepo := &mockCodeRepo{
			findByCodeHashFunc: func(_ context.Context, _ string) (*dstorage.AuthorizationCode, error) {
				return nil, connectionErr
			},
		}
		store := NewFositeStorage(codeRepo, memory.NewAgentRepository(), memory.NewBrokerClientCredentialStore(), testSlogger())

		_, err := store.GetAuthorizeCodeSession(context.Background(), "anycode", nil)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, fosite.ErrNotFound)
	})

	t.Run("GetAuthorizeCodeSession not-found returns ErrNotFound", func(t *testing.T) {
		notFoundErr := dstorage.NewStorageError("FindByCodeHash", dstorage.ErrorKindNotFound, nil, "not found")
		codeRepo := &mockCodeRepo{
			findByCodeHashFunc: func(_ context.Context, _ string) (*dstorage.AuthorizationCode, error) {
				return nil, notFoundErr
			},
		}
		store := NewFositeStorage(codeRepo, memory.NewAgentRepository(), memory.NewBrokerClientCredentialStore(), testSlogger())

		_, err := store.GetAuthorizeCodeSession(context.Background(), "anycode", nil)
		assert.ErrorIs(t, err, fosite.ErrNotFound)
	})

	t.Run("GetClient credential repo timeout is not ErrNotFound", func(t *testing.T) {
		credRepo := &mockCredentialRepo{
			getByBrokerClientIDFunc: func(_ context.Context, _ id.BrokerClientID) (*dstorage.BrokerClientCredential, error) {
				return nil, timeoutErr
			},
		}
		store := NewFositeStorage(memory.NewAuthorizationCodeStore(), memory.NewAgentRepository(), credRepo, testSlogger())

		_, err := store.GetClient(context.Background(), "broker_any")
		assert.Error(t, err)
		assert.NotErrorIs(t, err, fosite.ErrNotFound)
	})

	t.Run("GetClient agent repo timeout is not ErrNotFound", func(t *testing.T) {
		agentID := id.NewAgentID()
		cred := &dstorage.BrokerClientCredential{
			ID:             id.NewCredentialID(),
			AgentID:        agentID,
			BrokerClientID: id.NewBrokerClientID("broker_infra_test"),
			SecretHash:     "hash",
		}
		credRepo := memory.NewBrokerClientCredentialStore()
		require.NoError(t, credRepo.Create(context.Background(), cred))

		agentRepo := &mockAgentRepo{
			getFunc: func(_ context.Context, _ id.AgentID) (*dstorage.Agent, error) {
				return nil, dstorage.NewStorageError("Get", dstorage.ErrorKindTimeout, nil, "query timeout")
			},
		}
		store := NewFositeStorage(memory.NewAuthorizationCodeStore(), agentRepo, credRepo, testSlogger())

		_, err := store.GetClient(context.Background(), cred.BrokerClientID.String())
		assert.Error(t, err)
		assert.NotErrorIs(t, err, fosite.ErrNotFound)
	})
}
