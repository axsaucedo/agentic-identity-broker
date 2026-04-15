package oauth2server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/ory/fosite"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Compile-time interface checks
var (
	_ fositeOAuth2.AccessTokenStrategy   = (*JWXAccessTokenStrategy)(nil)
	_ fositeOAuth2.AuthorizeCodeStrategy = (*RandomCodeStrategy)(nil)
)

// baseClaims are JWT claims that cannot be overridden by CEL expressions.
var baseClaims = map[string]bool{
	"iss": true, "sub": true, "iat": true, "exp": true,
	"jti": true, "kid": true, "agent_id": true, "scope": true,
}

// JWXAccessTokenStrategy implements fosite's AccessTokenStrategy using lestrrat-go/jwx.
type JWXAccessTokenStrategy struct {
	signingKeyService *SigningKeyService
	signingKeyRepo    ports.SigningKeyRepository
	issuerURI         string
	tokenTTL          time.Duration
	customClaimsEval  *TokenClaimsEvaluator
	logger            *slog.Logger
}

// NewJWXAccessTokenStrategy creates a new JWX-based access token strategy.
func NewJWXAccessTokenStrategy(
	signingKeyService *SigningKeyService,
	signingKeyRepo ports.SigningKeyRepository,
	issuerURI string,
	tokenTTL time.Duration,
	customClaimsEval *TokenClaimsEvaluator,
	logger *slog.Logger,
) *JWXAccessTokenStrategy {
	return &JWXAccessTokenStrategy{
		signingKeyService: signingKeyService,
		signingKeyRepo:    signingKeyRepo,
		issuerURI:         issuerURI,
		tokenTTL:          tokenTTL,
		customClaimsEval:  customClaimsEval,
		logger:            logger,
	}
}

// GenerateAccessToken creates a signed JWT access token.
func (s *JWXAccessTokenStrategy) GenerateAccessToken(ctx context.Context, requester fosite.Requester) (string, string, error) {
	// 1. Get current signing key
	key, err := s.signingKeyRepo.GetCurrent(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get current signing key: %w", err)
	}

	// 2. Decrypt private key material
	privPEM, err := s.signingKeyService.DecryptPrivateKey(ctx, key)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt signing key: %w", err)
	}

	// 3. Parse PEM → jwk.Key
	privKey, err := jwk.ParseKey(privPEM, jwk.WithPEM(true))
	if err != nil {
		return "", "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// 4. Build JWT with claims
	now := time.Now()
	jti := uuid.New().String()

	// Subject: use session subject (principal for auth_code, client ID for client_credentials)
	subject := requester.GetSession().GetSubject()
	if subject == "" {
		subject = requester.GetClient().GetID()
	}

	builder := jwt.NewBuilder().
		Issuer(s.issuerURI).
		Subject(subject).
		IssuedAt(now).
		Expiration(now.Add(s.tokenTTL)).
		JwtID(jti).
		Claim("agent_id", requester.GetClient().GetID()).
		Claim("scope", strings.Join(requester.GetGrantedScopes(), " "))

	// 5. Evaluate CEL token_claims_expression if configured
	if s.customClaimsEval != nil {
		customClaims, err := s.customClaimsEval.Evaluate(ctx, requester)
		if err != nil {
			return "", "", fmt.Errorf("token claims expression evaluation failed: %w", err)
		}
		for k, v := range customClaims {
			if baseClaims[k] {
				s.logger.Warn("CEL expression returned reserved claim, skipping", "claim", k)
				continue
			}
			builder = builder.Claim(k, v)
		}
	}

	token, err := builder.Build()
	if err != nil {
		return "", "", fmt.Errorf("failed to build JWT: %w", err)
	}

	// 6. Sign with kid
	if err := privKey.Set(jwk.KeyIDKey, string(key.KID)); err != nil {
		return "", "", fmt.Errorf("failed to set kid on signing key %s: %w", key.KID, err)
	}
	alg := algorithmToJWA(key.Algorithm)
	signed, err := jwt.Sign(token, jwt.WithKey(alg, privKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Signature = SHA-256 of the token (for storage/lookup)
	return string(signed), sha256Hex(string(signed)), nil
}

// AccessTokenSignature returns a signature for the given access token.
func (s *JWXAccessTokenStrategy) AccessTokenSignature(_ context.Context, token string) string {
	return sha256Hex(token)
}

// ValidateAccessToken validates an access token.
func (s *JWXAccessTokenStrategy) ValidateAccessToken(_ context.Context, _ fosite.Requester, _ string) error {
	// Validation happens externally via JWKS endpoint
	return nil
}

// RandomCodeStrategy implements fosite's AuthorizeCodeStrategy using crypto/rand.
type RandomCodeStrategy struct{}

// GenerateAuthorizeCode generates a random authorization code.
func (s *RandomCodeStrategy) GenerateAuthorizeCode(_ context.Context, _ fosite.Requester) (string, string, error) {
	codeBytes := make([]byte, 32) // 32 bytes → 43 char base64url
	if _, err := rand.Read(codeBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random code: %w", err)
	}
	code := base64.RawURLEncoding.EncodeToString(codeBytes)
	signature := sha256Hex(code)
	return code, signature, nil
}

// AuthorizeCodeSignature returns a signature for the given authorization code.
func (s *RandomCodeStrategy) AuthorizeCodeSignature(_ context.Context, code string) string {
	return sha256Hex(code)
}

// ValidateAuthorizeCode validates an authorization code.
// Validation happens in storage (expiry, single-use), not in the strategy.
func (s *RandomCodeStrategy) ValidateAuthorizeCode(_ context.Context, _ fosite.Requester, _ string) error {
	return nil
}

// sha256Hex returns the hex-encoded SHA-256 hash of the input.
func sha256Hex(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
