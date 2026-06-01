package oauth2server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"

	domainencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// jwksCacheMaxAge is the max-age value (in seconds) sent in the Cache-Control header on
// GET /oauth2/jwks.json. The grace period below must be a multiple of this value.
// Keep in sync with the constant in internal/adapters/http/handlers/enduser/jwks_handler.go.
const jwksCacheMaxAge = 300 * time.Second

// jwksGracePeriod is how long a newly created current key waits before it starts signing tokens.
// During this window the key is already present in the JWKS response, so every client cache
// will have learned about it before the first token signed with it appears.
const jwksGracePeriod = 2 * jwksCacheMaxAge

// SigningKeyService manages signing key lifecycle including generation,
// encryption, storage, and JWKS building.
type SigningKeyService struct {
	repo             ports.SigningKeyRepository
	encryption       ports.EncryptionPort
	branchKeyManager ports.BranchKeyManager
	logger           *slog.Logger
}

// NewSigningKeyService creates a new SigningKeyService.
func NewSigningKeyService(
	repo ports.SigningKeyRepository,
	encryption ports.EncryptionPort,
	branchKeyManager ports.BranchKeyManager,
	logger *slog.Logger,
) *SigningKeyService {
	return &SigningKeyService{
		repo:             repo,
		encryption:       encryption,
		branchKeyManager: branchKeyManager,
		logger:           logger,
	}
}

// GenerateAndStoreKey generates a new ES256 signing key, encrypts the private material,
// and stores it. When makeCurrent is true the key is marked as current but will not begin
// signing tokens until jwksGracePeriod has elapsed, giving JWKS caches time to pick up the
// new key before any token signed with it is issued.
func (s *SigningKeyService) GenerateAndStoreKey(ctx context.Context, algorithm string, makeCurrent bool) (*storage.SigningKey, error) {
	activatesAt := time.Now().UTC()
	if makeCurrent {
		activatesAt = activatesAt.Add(jwksGracePeriod)
	}
	return s.generateAndStore(ctx, algorithm, makeCurrent, activatesAt)
}

// generateAndStore creates and persists a signing key with an explicit activatesAt timestamp.
func (s *SigningKeyService) generateAndStore(ctx context.Context, algorithm string, makeCurrent bool, activatesAt time.Time) (*storage.SigningKey, error) {
	if algorithm == "" {
		algorithm = "ES256"
	}

	kid := id.NewKeyID(uuid.New().String())

	var privKeyPEM []byte
	var err error

	switch algorithm {
	case "ES256":
		privKeyPEM, err = generateES256KeyPEM()
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	signingKeySubject, err := newSigningKeySubject(kid)
	if err != nil {
		return nil, err
	}

	branchKeyID, err := s.branchKeyManager.Create(ctx, signingKeySubject)
	if err != nil {
		return nil, fmt.Errorf("failed to provision branch key for signing key: %w", err)
	}

	encrypted, err := s.encryption.Encrypt(ctx, privKeyPEM, signingKeyEncCtx(kid))
	if err != nil {
		s.warnOrphanedBranchKey("orphaned branch key after encryption failure; manual cleanup required", kid, branchKeyID)
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	key := &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 kid,
		Algorithm:           algorithm,
		PrivateKeyEncrypted: encrypted,
		IsCurrent:           makeCurrent,
		ActivatesAt:         activatesAt,
		CreatedAt:           time.Now().UTC(),
	}

	if makeCurrent {
		if err := s.repo.CreateAndSetCurrent(ctx, key); err != nil {
			s.warnOrphanedBranchKey("orphaned branch key after storage failure; manual cleanup required", kid, branchKeyID)
			return nil, fmt.Errorf("failed to store and promote signing key: %w", err)
		}
	} else {
		if err := s.repo.Create(ctx, key); err != nil {
			s.warnOrphanedBranchKey("orphaned branch key after storage failure; manual cleanup required", kid, branchKeyID)
			return nil, fmt.Errorf("failed to store signing key: %w", err)
		}
	}

	s.logger.Info("signing key generated", "kid", kid, "algorithm", algorithm, "is_current", makeCurrent, "activates_at", activatesAt)
	return key, nil
}

// BuildJWKS constructs a JWK Set from all active signing keys (public keys only).
// All active keys are included regardless of activates_at, so new keys appear in the
// JWKS during the grace period and clients can cache them before they start signing.
func (s *SigningKeyService) BuildJWKS(ctx context.Context) (jwk.Set, error) {
	keys, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active keys: %w", err)
	}

	set := jwk.NewSet()
	for _, key := range keys {
		privPEM, err := s.encryption.Decrypt(ctx, key.PrivateKeyEncrypted, signingKeyEncCtx(key.KID))
		if err != nil {
			s.logger.Error("failed to decrypt signing key, skipping", "kid", key.KID, "error", err)
			continue
		}

		pubKey, err := publicKeyFromPEM(privPEM, key.Algorithm)
		if err != nil {
			s.logger.Error("failed to parse signing key, skipping", "kid", key.KID, "error", err)
			continue
		}

		jwkKey, err := jwk.Import(pubKey)
		if err != nil {
			s.logger.Error("failed to import signing key to JWK, skipping", "kid", key.KID, "error", err)
			continue
		}

		jwaAlg, err := algorithmToJWA(key.Algorithm)
		if err != nil {
			s.logger.Error("signing key has unrecognized algorithm, skipping", "kid", key.KID, "algorithm", key.Algorithm, "error", err)
			continue
		}
		_ = jwkKey.Set(jwk.KeyIDKey, key.KID.String())
		_ = jwkKey.Set(jwk.AlgorithmKey, jwaAlg)
		_ = jwkKey.Set(jwk.KeyUsageKey, "sig")

		if err := set.AddKey(jwkKey); err != nil {
			s.logger.Error("failed to add signing key to JWKS, skipping", "kid", key.KID, "error", err)
			continue
		}
	}

	if len(keys) > 0 && set.Len() == 0 {
		return nil, fmt.Errorf("failed to build JWKS: all %d active key(s) failed processing", len(keys))
	}

	return set, nil
}

// GetCurrent returns the active signing key used for token signing.
func (s *SigningKeyService) GetCurrent(ctx context.Context) (*storage.SigningKey, error) {
	return s.repo.GetCurrent(ctx)
}

// CountActive returns the number of non-removed signing keys.
func (s *SigningKeyService) CountActive(ctx context.Context) (int, error) {
	return s.repo.CountActive(ctx)
}

// DeleteKey removes a non-current signing key.
// Returns an error if the key is the last active key or is the current key.
func (s *SigningKeyService) DeleteKey(ctx context.Context, kid id.KeyID) error {
	count, err := s.repo.CountActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to count active keys: %w", err)
	}
	if count <= 1 {
		return ports.ErrLastActiveKey
	}

	key, err := s.repo.GetByKID(ctx, kid)
	if err != nil {
		return err
	}
	if key.IsCurrent {
		return ports.ErrCurrentKey
	}

	return s.repo.Delete(ctx, kid)
}

// DecryptPrivateKey decrypts the private key material of a signing key.
func (s *SigningKeyService) DecryptPrivateKey(ctx context.Context, key *storage.SigningKey) ([]byte, error) {
	return s.encryption.Decrypt(ctx, key.PrivateKeyEncrypted, signingKeyEncCtx(key.KID))
}

func newSigningKeySubject(kid id.KeyID) (domainencryption.BranchKeySubject, error) {
	subject := domainencryption.NewSigningKeyBranchKeySubject(kid)
	if err := subject.Validate(); err != nil {
		return domainencryption.BranchKeySubject{}, fmt.Errorf("invalid signing key subject: %w", err)
	}
	return subject, nil
}

func (s *SigningKeyService) warnOrphanedBranchKey(message string, kid id.KeyID, branchKeyID string) {
	if branchKeyID == "" {
		s.logger.Debug("skipping orphan warning: branch key ID is empty (noop backend or unexpected empty return)", "kid", kid)
		return
	}
	s.logger.Warn(message, "kid", kid, "branch_key_id", branchKeyID)
}

// signingKeyEncCtx returns the encryption context AAD for a signing key.
// The signing-key subject uses the well-known JWT kid term in AAD and routes the
// hierarchical keyring to the dedicated signing-key branch key namespace.
func signingKeyEncCtx(kid id.KeyID) map[string]string {
	return domainencryption.NewSigningKeyBranchKeySubject(kid).EncryptionContext()
}

func generateES256KeyPEM() ([]byte, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}

func publicKeyFromPEM(privPEM []byte, algorithm string) (interface{}, error) {
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
	}

	switch algorithm {
	case "ES256":
		ecKey, ok := privKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("expected ECDSA private key, got %T", privKey)
		}
		return &ecKey.PublicKey, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

func algorithmToJWA(algorithm string) (jwa.SignatureAlgorithm, error) {
	switch algorithm {
	case "ES256":
		return jwa.ES256(), nil
	case "RS256":
		return jwa.RS256(), nil
	default:
		return jwa.SignatureAlgorithm{}, fmt.Errorf("unrecognized algorithm: %q", algorithm)
	}
}
