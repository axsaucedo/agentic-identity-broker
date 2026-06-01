package oauth2server

import (
	"bytes"
	"context"
	"encoding/pem"
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
	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
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
	createFn    func(ctx context.Context, subject domainencryption.BranchKeySubject) (string, error)
	createCalls int
	lastSubject domainencryption.BranchKeySubject
}

func (m *mockBranchKeyManager) Create(ctx context.Context, subject domainencryption.BranchKeySubject) (string, error) {
	m.createCalls++
	m.lastSubject = subject
	if m.createFn != nil {
		return m.createFn(ctx, subject)
	}
	return "", nil
}

func newNoopBranchKeyManager() *mockBranchKeyManager {
	return &mockBranchKeyManager{}
}

// failingDecryptor always errors on Decrypt, simulating KMS unavailability.
type failingDecryptor struct{}

func (e *failingDecryptor) Encrypt(_ context.Context, plaintext []byte, _ map[string]string) ([]byte, error) {
	result := make([]byte, 0, len(plaintext)+4)
	result = append(result, []byte("ENC:")...)
	result = append(result, plaintext...)
	return result, nil
}

func (e *failingDecryptor) Decrypt(_ context.Context, _ []byte, _ map[string]string) ([]byte, error) {
	return nil, errors.New("decrypt: KMS unavailable")
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
	return NewSigningKeyService(repo, enc, newNoopBranchKeyManager(), logger), repo
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

func TestNewSigningKeySubject(t *testing.T) {
	t.Run("valid kid", func(t *testing.T) {
		subject, err := newSigningKeySubject(id.NewKeyID("kid-123"))
		require.NoError(t, err)
		assert.Equal(t, domainencryption.BranchKeySubjectKindSigningKey, subject.Kind())
		assert.Equal(t, "kid-123", subject.Identifier())
	})

	t.Run("empty kid", func(t *testing.T) {
		_, err := newSigningKeySubject(id.KeyID(""))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signing key subject")
		assert.Contains(t, err.Error(), "kid is required")
	})
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

	t.Run("returns error when all active keys fail processing", func(t *testing.T) {
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

		svc := NewSigningKeyService(repo, &testEncryptor{}, newNoopBranchKeyManager(), testSlogger())
		_, err = svc.BuildJWKS(ctx)
		require.Error(t, err, "all active keys failed processing — should return error")
		assert.Contains(t, err.Error(), "failed to build JWKS")
	})

	t.Run("returns error when KMS is down and all decrypts fail", func(t *testing.T) {
		ctx := context.Background()
		repo := memory.NewSigningKeyStore()

		// Store a key that cannot be decrypted (KMS-down scenario).
		kid := id.NewKeyID(uuid.New().String())
		err := repo.Create(ctx, &storage.SigningKey{
			ID:                  id.NewSigningKeyID(),
			KID:                 kid,
			Algorithm:           "ES256",
			PrivateKeyEncrypted: []byte("ciphertext-that-will-fail-decrypt"),
			IsCurrent:           true,
		})
		require.NoError(t, err)

		// failingDecryptor simulates KMS unavailability.
		svc := NewSigningKeyService(repo, &failingDecryptor{}, newNoopBranchKeyManager(), testSlogger())
		_, buildErr := svc.BuildJWKS(ctx)
		require.Error(t, buildErr, "KMS down — all decrypts fail — should return error")
		assert.Contains(t, buildErr.Error(), "failed to build JWKS")
	})

	t.Run("partial failure: one bad key skipped, set returned with remaining key", func(t *testing.T) {
		ctx := context.Background()
		repo := memory.NewSigningKeyStore()

		// Key 1: valid, decryptable.
		privPEM, err := generateES256KeyPEM()
		require.NoError(t, err)
		goodKID := id.NewKeyID(uuid.New().String())
		err = repo.Create(ctx, &storage.SigningKey{
			ID:                  id.NewSigningKeyID(),
			KID:                 goodKID,
			Algorithm:           "ES256",
			PrivateKeyEncrypted: append([]byte("ENC:"), privPEM...),
			IsCurrent:           false,
		})
		require.NoError(t, err)

		// Key 2: bogus algorithm — will fail processing.
		badKID := id.NewKeyID(uuid.New().String())
		err = repo.Create(ctx, &storage.SigningKey{
			ID:                  id.NewSigningKeyID(),
			KID:                 badKID,
			Algorithm:           "BOGUS",
			PrivateKeyEncrypted: append([]byte("ENC:"), privPEM...),
			IsCurrent:           true,
		})
		require.NoError(t, err)

		svc := NewSigningKeyService(repo, &testEncryptor{}, newNoopBranchKeyManager(), testSlogger())
		jwks, buildErr := svc.BuildJWKS(ctx)
		require.NoError(t, buildErr, "at least one key succeeded — should not return error")
		assert.Equal(t, 1, jwks.Len(), "only the good key should be in the set")

		k, ok := jwks.Key(0)
		require.True(t, ok)
		kid, _ := k.KeyID()
		assert.Equal(t, goodKID.String(), kid)
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
		block, _ := pem.Decode(decrypted)
		require.NotNil(t, block)
		assert.Equal(t, "PRIVATE KEY", block.Type)
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
			createFn: func(_ context.Context, _ domainencryption.BranchKeySubject) (string, error) {
				return "branch-key-id", nil
			},
		}
		svc, repo := newTestSigningKeyServiceWithBranchKeyManager(bkm)

		key, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.NoError(t, err)
		require.NotNil(t, key)

		// Branch key was provisioned exactly once for the signing-key subject.
		assert.Equal(t, 1, bkm.createCalls)
		assert.Equal(t, domainencryption.BranchKeySubjectKindSigningKey, bkm.lastSubject.Kind())
		assert.Equal(t, key.KID.String(), bkm.lastSubject.Identifier())

		// Key was stored in the repo.
		stored, err := repo.GetByKID(context.Background(), key.KID)
		require.NoError(t, err)
		assert.Equal(t, key.KID, stored.KID)

		// Private key material is decryptable.
		decrypted, err := svc.DecryptPrivateKey(context.Background(), stored)
		require.NoError(t, err)
		block, _ := pem.Decode(decrypted)
		require.NotNil(t, block)
		assert.Equal(t, "PRIVATE KEY", block.Type)
	})

	t.Run("Create failure: error propagates, nothing stored", func(t *testing.T) {
		bkm := &mockBranchKeyManager{
			createFn: func(_ context.Context, _ domainencryption.BranchKeySubject) (string, error) {
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
		// Verifies branchKeyManager.Create is called before Encrypt.
		// Because failingEncryptor always errors on Encrypt, if Create were called after Encrypt
		// (or not at all), createCalls would be 0. createCalls == 1 proves Create ran first.
		bkm := &mockBranchKeyManager{
			createFn: func(_ context.Context, _ domainencryption.BranchKeySubject) (string, error) {
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
			createFn: func(_ context.Context, _ domainencryption.BranchKeySubject) (string, error) {
				return "branch-key-id", nil
			},
		}
		svc, _ := newTestSigningKeyServiceWithBranchKeyManagerAndEncryptor(bkm, &failingEncryptor{}, &logBuf)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.Error(t, err)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "orphaned branch key", "warn log must identify the orphaned entry")
		assert.Contains(t, logOutput, "kid", "warn log must include the kid field")
		assert.Contains(t, logOutput, "branch-key-id", "warn log must include the branch key ID for operator cleanup")
	})

	t.Run("warns with kid when repo.Create fails after branch key created", func(t *testing.T) {
		var logBuf bytes.Buffer
		bkm := &mockBranchKeyManager{
			createFn: func(_ context.Context, _ domainencryption.BranchKeySubject) (string, error) {
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
		assert.Contains(t, logOutput, "branch-key-id", "warn log must include the branch key ID for operator cleanup")
	})

	t.Run("warns with kid when repo.CreateAndSetCurrent fails after branch key created", func(t *testing.T) {
		var logBuf bytes.Buffer
		bkm := &mockBranchKeyManager{
			createFn: func(_ context.Context, _ domainencryption.BranchKeySubject) (string, error) {
				return "branch-key-id", nil
			},
		}
		repo := &mockFailingSigningKeyRepo{
			SigningKeyStore: memory.NewSigningKeyStore(),
			createErr:       errors.New("storage unavailable"),
		}
		logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		svc := NewSigningKeyService(repo, &testEncryptor{}, bkm, logger)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.Error(t, err)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "orphaned branch key", "warn log must identify the orphaned entry")
		assert.Contains(t, logOutput, "kid", "warn log must include the kid field")
		assert.Contains(t, logOutput, "branch-key-id", "warn log must include the branch key ID for operator cleanup")
	})

	t.Run("logs debug breadcrumb when branch key id is empty", func(t *testing.T) {
		var logBuf bytes.Buffer
		svc, _ := newTestSigningKeyServiceWithBranchKeyManagerAndEncryptor(newNoopBranchKeyManager(), &failingEncryptor{}, &logBuf)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.Error(t, err)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "branch key ID is empty", "empty branch key IDs must leave a diagnostic breadcrumb")
		assert.Contains(t, logOutput, "kid", "debug breadcrumb must retain the signing key kid")
		assert.NotContains(t, logOutput, "orphaned branch key", "empty branch key IDs must not emit manual-cleanup warnings")
		assert.NotContains(t, logOutput, "branch_key_id", "empty branch key IDs must not log a branch_key_id field")
	})

	t.Run("noop branch key manager does not warn about cleanup when Encrypt fails", func(t *testing.T) {
		var logBuf bytes.Buffer
		svc, _ := newTestSigningKeyServiceWithBranchKeyManagerAndEncryptor(newNoopBranchKeyManager(), &failingEncryptor{}, &logBuf)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.Error(t, err)

		logOutput := logBuf.String()
		assert.NotContains(t, logOutput, "orphaned branch key", "noop branch key managers must not emit manual-cleanup warnings")
		assert.NotContains(t, logOutput, "branch_key_id", "noop branch key managers must not log an empty branch_key_id")
	})

	t.Run("noop branch key manager does not warn about cleanup when repo.Create fails", func(t *testing.T) {
		var logBuf bytes.Buffer
		repo := &mockFailingSigningKeyRepo{
			SigningKeyStore: memory.NewSigningKeyStore(),
			createErr:       errors.New("storage unavailable"),
		}
		logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		svc := NewSigningKeyService(repo, &testEncryptor{}, newNoopBranchKeyManager(), logger)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.Error(t, err)

		logOutput := logBuf.String()
		assert.NotContains(t, logOutput, "orphaned branch key", "noop branch key managers must not emit manual-cleanup warnings")
		assert.NotContains(t, logOutput, "branch_key_id", "noop branch key managers must not log an empty branch_key_id")
	})

	t.Run("noop branch key manager does not warn about cleanup when repo.CreateAndSetCurrent fails", func(t *testing.T) {
		var logBuf bytes.Buffer
		repo := &mockFailingSigningKeyRepo{
			SigningKeyStore: memory.NewSigningKeyStore(),
			createErr:       errors.New("storage unavailable"),
		}
		logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		svc := NewSigningKeyService(repo, &testEncryptor{}, newNoopBranchKeyManager(), logger)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.Error(t, err)

		logOutput := logBuf.String()
		assert.NotContains(t, logOutput, "orphaned branch key", "noop branch key managers must not emit manual-cleanup warnings")
		assert.NotContains(t, logOutput, "branch_key_id", "noop branch key managers must not log an empty branch_key_id")
	})

	t.Run("default mock branch key manager never causes provisioning failure", func(t *testing.T) {
		// The default mock returns success and must never block key generation.
		bkm := newNoopBranchKeyManager()
		svc, repo := newTestSigningKeyServiceWithBranchKeyManager(bkm)

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.NoError(t, err)

		count, countErr := repo.CountActive(context.Background())
		require.NoError(t, countErr)
		assert.Equal(t, 1, count)
	})
}

func TestSigningKeyEncCtx(t *testing.T) {
	t.Run("value is bare kid UUID with no prefix", func(t *testing.T) {
		kid := id.NewKeyID("550e8400-e29b-41d4-a716-446655440000")
		ctx := signingKeyEncCtx(kid)
		require.Len(t, ctx, 1, "encryption context must contain exactly one key")
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ctx["kid"],
			"kid must be the signing key identifier used in AAD")
	})
}

func TestSigningKeyService_CountActive(t *testing.T) {
	t.Run("returns zero when no keys exist", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		count, err := svc.CountActive(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("reflects number of active keys", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.NoError(t, err)
		_, err = svc.GenerateAndStoreKey(context.Background(), "ES256", false)
		require.NoError(t, err)

		count, err := svc.CountActive(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}

func TestSigningKeyService_GetCurrent(t *testing.T) {
	t.Run("returns error when no keys exist", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		_, err := svc.GetCurrent(context.Background())
		assert.Error(t, err)
	})

	t.Run("returns the current active key", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		// Use generateAndStore with time.Now() so activates_at is in the past.
		key, err := svc.generateAndStore(context.Background(), "ES256", true, time.Now())
		require.NoError(t, err)

		got, err := svc.GetCurrent(context.Background())
		require.NoError(t, err)
		assert.Equal(t, key.KID, got.KID)
	})
}
