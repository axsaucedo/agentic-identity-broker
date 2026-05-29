package oauth2server

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// testEncryptor is a minimal encryption implementation for testing.
// It stores ciphertext as plaintext (no actual encryption) to enable testing
// without AWS KMS infrastructure.
type testEncryptor struct{}

func (e *testEncryptor) Encrypt(_ context.Context, plaintext []byte, _ map[string]string) ([]byte, error) {
	// For testing: return plaintext with a marker prefix to distinguish from raw data
	result := make([]byte, 0, len(plaintext)+4)
	result = append(result, []byte("ENC:")...)
	result = append(result, plaintext...)
	return result, nil
}

func (e *testEncryptor) Decrypt(_ context.Context, ciphertext []byte, _ map[string]string) ([]byte, error) {
	// For testing: strip the marker prefix
	if len(ciphertext) < 4 || string(ciphertext[:4]) != "ENC:" {
		return nil, assert.AnError
	}
	return ciphertext[4:], nil
}

// mockBranchKeyManager is a hand-rolled mock for ports.BranchKeyManager.
type mockBranchKeyManager struct {
	createFn    func(ctx context.Context, serviceID id.ServiceID) (string, error)
	createCalls int
}

func (m *mockBranchKeyManager) Create(ctx context.Context, serviceID id.ServiceID) (string, error) {
	m.createCalls++
	return m.createFn(ctx, serviceID)
}

// failingEncryptor always returns an error on Encrypt, used to verify ordering.
type failingEncryptor struct{}

func (e *failingEncryptor) Encrypt(_ context.Context, _ []byte, _ map[string]string) ([]byte, error) {
	return nil, errors.New("encrypt: simulated failure")
}

func (e *failingEncryptor) Decrypt(_ context.Context, ciphertext []byte, _ map[string]string) ([]byte, error) {
	if len(ciphertext) < 4 || string(ciphertext[:4]) != "ENC:" {
		return nil, assert.AnError
	}
	return ciphertext[4:], nil
}

func newTestSigningKeyService() (*SigningKeyService, *memory.SigningKeyStore) {
	repo := memory.NewSigningKeyStore()
	enc := &testEncryptor{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewSigningKeyService(repo, enc, nil, logger), repo
}

func newTestSigningKeyServiceWithBranchKeyManager(bkm ports.BranchKeyManager) (*SigningKeyService, *memory.SigningKeyStore) {
	repo := memory.NewSigningKeyStore()
	enc := &testEncryptor{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewSigningKeyService(repo, enc, bkm, logger), repo
}

func newTestSigningKeyServiceWithBranchKeyManagerAndEncryptor(bkm ports.BranchKeyManager, enc ports.EncryptionPort, logBuf *bytes.Buffer) (*SigningKeyService, *memory.SigningKeyStore) {
	repo := memory.NewSigningKeyStore()
	var logger *slog.Logger
	if logBuf != nil {
		logger = slog.New(slog.NewJSONHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	return NewSigningKeyService(repo, enc, bkm, logger), repo
}

func TestSigningKeyService_GenerateAndStoreKey(t *testing.T) {
	t.Run("generates ES256 key pair", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		key, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.NoError(t, err)
		require.NotNil(t, key)

		assert.Equal(t, "ES256", key.Algorithm)
		assert.False(t, key.KID.IsZero())
		assert.True(t, len(key.PrivateKeyEncrypted) > 0)
	})

	t.Run("default algorithm is ES256", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		key, err := svc.GenerateAndStoreKey(context.Background(), "", true)
		require.NoError(t, err)
		assert.Equal(t, "ES256", key.Algorithm)
	})

	t.Run("unsupported algorithm rejected", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		_, err := svc.GenerateAndStoreKey(context.Background(), "RS384", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported algorithm")
	})

	t.Run("makeCurrent marks key as is_current", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		key, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.NoError(t, err)

		// Key is stored as is_current=true even though it is in the grace period.
		stored, err := repo.GetByKID(context.Background(), key.KID)
		require.NoError(t, err)
		assert.True(t, stored.IsCurrent)
	})

	t.Run("new current key has future activates_at (grace period)", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		before := time.Now()
		key, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.NoError(t, err)

		stored, err := repo.GetByKID(context.Background(), key.KID)
		require.NoError(t, err)
		assert.True(t, stored.ActivatesAt.After(before.Add(jwksGracePeriod-time.Second)),
			"activates_at should be approximately now+jwksGracePeriod")

		// GetCurrent must not return this key while it is in its grace period.
		_, err = repo.GetCurrent(context.Background())
		assert.Error(t, err, "key should not be available for signing during grace period")
	})

	t.Run("private key is encrypted (has ENC: prefix)", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		key, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.NoError(t, err)

		// Verify the stored key has the test encryption prefix
		assert.True(t, len(key.PrivateKeyEncrypted) > 4)
		assert.Equal(t, "ENC:", string(key.PrivateKeyEncrypted[:4]))
	})
}

func TestSigningKeyService_BuildJWKS(t *testing.T) {
	t.Run("JWKS contains active public keys", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()

		// Generate two keys
		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.NoError(t, err)
		key2, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.NoError(t, err)

		jwks, err := svc.BuildJWKS(context.Background())
		require.NoError(t, err)
		require.NotNil(t, jwks)

		// Should have 2 keys in the set
		assert.Equal(t, 2, jwks.Len())

		// Verify the second key is in the set
		found := false
		for i := 0; i < jwks.Len(); i++ {
			k, ok := jwks.Key(i)
			if !ok {
				continue
			}
			kid, ok := k.KeyID()
			if !ok {
				continue
			}
			if kid == key2.KID.String() {
				found = true
			}
		}
		assert.True(t, found, "JWKS should contain the second key")
	})

	t.Run("empty JWKS when no keys exist", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()

		jwks, err := svc.BuildJWKS(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, jwks.Len())
	})

	t.Run("unrecognized algorithm skips key and returns empty JWKS", func(t *testing.T) {
		_, repo := newTestSigningKeyService()
		ctx := context.Background()

		// Build a key record with a bogus algorithm directly in the store,
		// bypassing GenerateAndStoreKey which rejects unknown algorithms.
		privPEM, err := generateES256KeyPEM()
		require.NoError(t, err)
		encrypted := append([]byte("ENC:"), privPEM...)

		kid := id.NewKeyID(uuid.New().String())
		err = repo.Create(ctx, &storage.SigningKey{
			ID:                  id.NewSigningKeyID(),
			KID:                 kid,
			Algorithm:           "BOGUS",
			PrivateKeyEncrypted: encrypted,
			IsCurrent:           true,
		})
		require.NoError(t, err)

		svc := NewSigningKeyService(repo, &testEncryptor{}, nil, testSlogger())
		jwks, err := svc.BuildJWKS(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, jwks.Len(), "bad key should be skipped")
	})
}

func TestAlgorithmToJWA(t *testing.T) {
	t.Run("ES256", func(t *testing.T) {
		alg, err := algorithmToJWA("ES256")
		require.NoError(t, err)
		assert.Equal(t, jwa.ES256(), alg)
	})

	t.Run("RS256", func(t *testing.T) {
		alg, err := algorithmToJWA("RS256")
		require.NoError(t, err)
		assert.Equal(t, jwa.RS256(), alg)
	})

	t.Run("unrecognized algorithm returns error and zero value", func(t *testing.T) {
		alg, err := algorithmToJWA("BOGUS")
		assert.ErrorContains(t, err, "unrecognized algorithm")
		assert.Equal(t, jwa.SignatureAlgorithm{}, alg)
	})
}

func TestSigningKeyService_GenerateAndStoreKey_Atomic(t *testing.T) {
	t.Run("CreateAndSetCurrent demotes previous current key", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		ctx := context.Background()

		key1, err := svc.GenerateAndStoreKey(ctx, "ES256", true)
		require.NoError(t, err)
		require.True(t, key1.IsCurrent)

		key2, err := svc.GenerateAndStoreKey(ctx, "ES256", true)
		require.NoError(t, err)
		require.True(t, key2.IsCurrent)

		// key1 must no longer be current.
		stored1, err := repo.GetByKID(ctx, key1.KID)
		require.NoError(t, err)
		assert.False(t, stored1.IsCurrent, "previous key must be demoted")

		// key2 is flagged is_current but still in grace period — GetCurrent falls back to key1.
		stored2, err := repo.GetByKID(ctx, key2.KID)
		require.NoError(t, err)
		assert.True(t, stored2.IsCurrent, "key2 must be flagged is_current")
	})
}

func TestSigningKeyService_DecryptPrivateKey(t *testing.T) {
	t.Run("decrypt round-trip", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		key, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.NoError(t, err)

		decrypted, err := svc.DecryptPrivateKey(context.Background(), key)
		require.NoError(t, err)
		assert.True(t, len(decrypted) > 0)
		// Should be valid PEM
		assert.Contains(t, string(decrypted), "-----BEGIN PRIVATE KEY-----")
	})
}

func TestSigningKeyService_AdminOperations(t *testing.T) {
	t.Run("add key becomes current and previous is demoted", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		ctx := context.Background()

		key1, err := svc.GenerateAndStoreKey(ctx, "ES256", true)
		require.NoError(t, err)

		key2, err := svc.GenerateAndStoreKey(ctx, "ES256", true)
		require.NoError(t, err)

		// key1 should no longer be current
		k1, err := repo.GetByKID(ctx, key1.KID)
		require.NoError(t, err)
		assert.False(t, k1.IsCurrent, "key1 should no longer be current after key2 is added as current")

		// key2 should be current
		k2, err := repo.GetByKID(ctx, key2.KID)
		require.NoError(t, err)
		assert.True(t, k2.IsCurrent, "key2 should be current")
	})

	t.Run("list returns metadata only", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		ctx := context.Background()

		key1, err := svc.GenerateAndStoreKey(ctx, "ES256", true)
		require.NoError(t, err)
		_, err = svc.GenerateAndStoreKey(ctx, "ES256", false)
		require.NoError(t, err)

		keys, err := repo.ListActive(ctx)
		require.NoError(t, err)
		require.Len(t, keys, 2)

		// Verify metadata fields are present on all keys
		for _, k := range keys {
			assert.False(t, k.KID.IsZero(), "KID should not be zero")
			assert.Equal(t, "ES256", k.Algorithm)
			assert.True(t, len(k.PrivateKeyEncrypted) > 0, "encrypted private key should be stored")
		}

		// Verify is_current status
		kidToCurrent := make(map[string]bool)
		for _, k := range keys {
			kidToCurrent[k.KID.String()] = k.IsCurrent
		}
		assert.True(t, kidToCurrent[key1.KID.String()], "key1 should be current")
	})

	t.Run("promote key changes current", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		ctx := context.Background()

		key1, err := svc.GenerateAndStoreKey(ctx, "ES256", true)
		require.NoError(t, err)
		key2, err := svc.GenerateAndStoreKey(ctx, "ES256", true)
		require.NoError(t, err)

		// key2 is current; promote key1 back
		err = repo.SetCurrent(ctx, key1.KID)
		require.NoError(t, err)

		k1, err := repo.GetByKID(ctx, key1.KID)
		require.NoError(t, err)
		assert.True(t, k1.IsCurrent, "key1 should be current after promotion")

		k2, err := repo.GetByKID(ctx, key2.KID)
		require.NoError(t, err)
		assert.False(t, k2.IsCurrent, "key2 should no longer be current")
	})

	t.Run("remove non-current succeeds", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		ctx := context.Background()

		_, err := svc.GenerateAndStoreKey(ctx, "ES256", true) // key1 is current
		require.NoError(t, err)
		key2, err := svc.GenerateAndStoreKey(ctx, "ES256", false) // key2 is not current
		require.NoError(t, err)

		err = repo.Delete(ctx, key2.KID)
		require.NoError(t, err)

		keys, err := repo.ListActive(ctx)
		require.NoError(t, err)
		require.Len(t, keys, 1, "only one key should remain after deletion")

		// key2 should not be in active list
		for _, k := range keys {
			assert.NotEqual(t, key2.KID, k.KID, "deleted key should not appear in active list")
		}
	})

	t.Run("remove last key returns conflict", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		ctx := context.Background()

		key1, err := svc.GenerateAndStoreKey(ctx, "ES256", true)
		require.NoError(t, err)

		// Attempt to remove the only key via service — should fail
		err = svc.DeleteKey(ctx, key1.KID)
		assert.Error(t, err, "should not allow removing the last active signing key")
	})
}

func TestSigningKeyService_BranchKeyProvisioning(t *testing.T) {
	t.Run("happy path: branch key created, key stored and decryptable", func(t *testing.T) {
		bkm := &mockBranchKeyManager{
			createFn: func(_ context.Context, _ id.ServiceID) (string, error) {
				return "branch-key-id", nil
			},
		}
		svc, repo := newTestSigningKeyServiceWithBranchKeyManager(bkm)

		key, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.NoError(t, err)
		require.NotNil(t, key)

		// Branch key was provisioned exactly once.
		assert.Equal(t, 1, bkm.createCalls)

		// Key was stored in the repo.
		stored, err := repo.GetByKID(context.Background(), key.KID)
		require.NoError(t, err)
		assert.Equal(t, key.KID, stored.KID)

		// Private key material is decryptable.
		decrypted, err := svc.DecryptPrivateKey(context.Background(), stored)
		require.NoError(t, err)
		assert.Contains(t, string(decrypted), "-----BEGIN PRIVATE KEY-----")
	})

	t.Run("Create failure: error propagates, nothing stored", func(t *testing.T) {
		bkm := &mockBranchKeyManager{
			createFn: func(_ context.Context, _ id.ServiceID) (string, error) {
				return "", errors.New("dynamo down")
			},
		}
		svc, repo := newTestSigningKeyServiceWithBranchKeyManager(bkm)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to provision branch key")

		// Nothing should have been stored.
		count, countErr := repo.CountActive(context.Background())
		require.NoError(t, countErr)
		assert.Equal(t, 0, count)
	})

	t.Run("ordering regression: branch key Created before Encrypt is called", func(t *testing.T) {
		// If someone moves branchKeyManager.Create after Encrypt, this test catches it:
		// the failingEncryptor errors before Create is reached, so createCalls stays 0.
		bkm := &mockBranchKeyManager{
			createFn: func(_ context.Context, _ id.ServiceID) (string, error) {
				return "branch-key-id", nil
			},
		}
		repo := memory.NewSigningKeyStore()
		enc := &failingEncryptor{}
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
		svc := NewSigningKeyService(repo, enc, bkm, logger)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.Error(t, err)

		// Create must have been called before Encrypt was attempted.
		assert.Equal(t, 1, bkm.createCalls, "branchKeyManager.Create must be called before Encrypt")
	})
}

// mockFailingSigningKeyRepo wraps memory.SigningKeyStore and fails Create/CreateAndSetCurrent.
type mockFailingSigningKeyRepo struct {
	*memory.SigningKeyStore
	createErr error
}

func (r *mockFailingSigningKeyRepo) Create(_ context.Context, _ *storage.SigningKey) error {
	return r.createErr
}

func (r *mockFailingSigningKeyRepo) CreateAndSetCurrent(_ context.Context, _ *storage.SigningKey) error {
	return r.createErr
}

func TestSigningKeyService_OrphanedBranchKeyWarning(t *testing.T) {
	t.Run("warns with kid when Encrypt fails after branch key created", func(t *testing.T) {
		var logBuf bytes.Buffer
		bkm := &mockBranchKeyManager{
			createFn: func(_ context.Context, _ id.ServiceID) (string, error) {
				return "branch-key-id", nil
			},
		}
		svc, _ := newTestSigningKeyServiceWithBranchKeyManagerAndEncryptor(bkm, &failingEncryptor{}, &logBuf)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.Error(t, err)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "orphaned branch key", "warn log must identify the orphaned entry")
		assert.Contains(t, logOutput, "kid", "warn log must include the kid field")
	})

	t.Run("warns with kid when repo.Create fails after branch key created", func(t *testing.T) {
		var logBuf bytes.Buffer
		bkm := &mockBranchKeyManager{
			createFn: func(_ context.Context, _ id.ServiceID) (string, error) {
				return "branch-key-id", nil
			},
		}
		repo := &mockFailingSigningKeyRepo{
			SigningKeyStore: memory.NewSigningKeyStore(),
			createErr:       errors.New("storage unavailable"),
		}
		logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		svc := NewSigningKeyService(repo, &testEncryptor{}, bkm, logger)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.Error(t, err)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "orphaned branch key", "warn log must identify the orphaned entry")
		assert.Contains(t, logOutput, "kid", "warn log must include the kid field")
	})

	t.Run("no orphan warning when branchKeyManager is nil", func(t *testing.T) {
		var logBuf bytes.Buffer
		// No branch key manager — warnings must not fire.
		repo := &mockFailingSigningKeyRepo{
			SigningKeyStore: memory.NewSigningKeyStore(),
			createErr:       errors.New("storage unavailable"),
		}
		logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		svc := NewSigningKeyService(repo, &testEncryptor{}, nil, logger)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.Error(t, err)

		assert.NotContains(t, logBuf.String(), "orphaned branch key")
	})
}
