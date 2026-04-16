package oauth2server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
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

func newTestSigningKeyService() (*SigningKeyService, *memory.SigningKeyStore) {
	repo := memory.NewSigningKeyStore()
	enc := &testEncryptor{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewSigningKeyService(repo, enc, logger), repo
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

	t.Run("makeCurrent sets key as current", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.NoError(t, err)

		current, err := repo.GetCurrent(context.Background())
		require.NoError(t, err)
		assert.True(t, current.IsCurrent)
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

	t.Run("unrecognized algorithm returns error", func(t *testing.T) {
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

		svc := NewSigningKeyService(repo, &testEncryptor{}, testSlogger())
		_, err = svc.BuildJWKS(ctx)
		assert.Error(t, err)
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

func TestSigningKeyService_EnsureKeyExists(t *testing.T) {
	t.Run("generates key when none exist", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()

		err := svc.EnsureKeyExists(context.Background())
		require.NoError(t, err)

		count, err := repo.CountActive(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("no-op when keys already exist", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()

		// Generate a key first
		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.NoError(t, err)

		// EnsureKeyExists should be a no-op
		err = svc.EnsureKeyExists(context.Background())
		require.NoError(t, err)

		count, err := repo.CountActive(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, count, "should still have exactly 1 key")
	})
}

// partialFailureKeyRepo wraps an in-memory store and injects a SetCurrent failure.
type partialFailureKeyRepo struct {
	*memory.SigningKeyStore
	setCurrrentErr error
	deletedKIDs    []id.KeyID
}

func (r *partialFailureKeyRepo) SetCurrent(_ context.Context, kid id.KeyID) error {
	return r.setCurrrentErr
}

func (r *partialFailureKeyRepo) Delete(ctx context.Context, kid id.KeyID) error {
	r.deletedKIDs = append(r.deletedKIDs, kid)
	return r.SigningKeyStore.Delete(ctx, kid)
}

func TestSigningKeyService_GenerateAndStoreKey_CompensatingDelete(t *testing.T) {
	t.Run("deletes orphaned key when SetCurrent fails", func(t *testing.T) {
		repo := &partialFailureKeyRepo{
			SigningKeyStore: memory.NewSigningKeyStore(),
			setCurrrentErr: fmt.Errorf("simulated SetCurrent failure"),
		}
		svc := NewSigningKeyService(repo, &testEncryptor{}, testSlogger())

		_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to set key as current")

		// The compensating Delete must have been called for the orphaned key.
		require.Len(t, repo.deletedKIDs, 1)

		// The key must not appear in ListActive after the compensating delete.
		active, listErr := repo.ListActive(context.Background())
		require.NoError(t, listErr)
		assert.Empty(t, active, "orphaned key should be removed from active keys")
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
