package oauth2server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/ory/fosite"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"
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
	issuerURI         string
	tokenTTL          time.Duration
	customClaimsEval  *TokenClaimsEvaluator
	logger            *slog.Logger
}

// NewJWXAccessTokenStrategy creates a new JWX-based access token strategy.
// issuerURI must be an absolute http/https URI with no fragment (RFC 8414 §2).
// tokenTTL must be positive.
func NewJWXAccessTokenStrategy(
	signingKeyService *SigningKeyService,
	issuerURI string,
	tokenTTL time.Duration,
	customClaimsEval *TokenClaimsEvaluator,
	logger *slog.Logger,
) (*JWXAccessTokenStrategy, error) {
	normalized, err := validateIssuerURI(issuerURI)
	if err != nil {
		return nil, err
	}
	if tokenTTL <= 0 {
		return nil, fmt.Errorf("tokenTTL must be positive, got %v", tokenTTL)
	}
	return &JWXAccessTokenStrategy{
		signingKeyService: signingKeyService,
		issuerURI:         normalized,
		tokenTTL:          tokenTTL,
		customClaimsEval:  customClaimsEval,
		logger:            logger,
	}, nil
}

// GenerateAccessToken creates a signed JWT access token.
func (s *JWXAccessTokenStrategy) GenerateAccessToken(ctx context.Context, requester fosite.Requester) (string, string, error) {
	// 1. Get current signing key
	key, err := s.signingKeyService.GetCurrent(ctx)
	if err != nil {
		if isStorageNotFound(err) {
			return "", "", fmt.Errorf("no signing key provisioned: create one via the admin API (POST /api/oauth2-server/signing-keys)")
		}
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
			evalErr := fmt.Errorf("token claims expression evaluation failed: %w", err)
			s.logger.ErrorContext(ctx, "token claims expression evaluation failed",
				"client_id", requester.GetClient().GetID(),
				"subject", subject,
				"granted_scopes", requester.GetGrantedScopes(),
				"error", evalErr,
			)
			return "", "", evalErr
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
		buildErr := fmt.Errorf("failed to build JWT: %w", err)
		s.logger.ErrorContext(ctx, "failed to build JWT",
			"client_id", requester.GetClient().GetID(),
			"subject", subject,
			"granted_scopes", requester.GetGrantedScopes(),
			"error", buildErr,
		)
		return "", "", buildErr
	}

	// 6. Sign with kid
	_ = privKey.Set(jwk.KeyIDKey, string(key.KID))
	alg, err := algorithmToJWA(key.Algorithm)
	if err != nil {
		return "", "", fmt.Errorf("signing key %s has unrecognized algorithm %q: %w", key.KID, key.Algorithm, err)
	}
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

// ValidateAccessToken validates an access token by verifying its JWT signature,
// expiry, and issuer against all active signing keys.
func (s *JWXAccessTokenStrategy) ValidateAccessToken(ctx context.Context, _ fosite.Requester, token string) error {
	jwks, err := s.signingKeyService.BuildJWKS(ctx)
	if err != nil {
		return fmt.Errorf("failed to build JWKS for token validation: %w", err)
	}
	if _, err := jwt.ParseString(token,
		jwt.WithVerify(true),
		jwt.WithKeySet(jwks),
		jwt.WithValidate(true),
		jwt.WithIssuer(s.issuerURI),
	); err != nil {
		return fosite.ErrTokenSignatureMismatch.WithWrap(err)
	}
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

// validateIssuerURI checks that s is a valid OAuth2 issuer identifier per RFC 8414 §2:
// absolute http/https URI, non-empty host, no query, no fragment.
// Returns the trimmed, normalized value on success.
func validateIssuerURI(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("issuerURI must not be empty or whitespace-only")
	}
	parsed, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("issuerURI is not a valid URI: %w", err)
	}
	if !parsed.IsAbs() {
		return "", fmt.Errorf("issuerURI must be an absolute URI (got %q)", s)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("issuerURI scheme must be http or https (got %q)", parsed.Scheme)
	}
	if parsed.Hostname() == "" {
		return "", fmt.Errorf("issuerURI must have a non-empty host (got %q)", s)
	}
	if parsed.RawQuery != "" || parsed.ForceQuery {
		return "", fmt.Errorf("issuerURI must not contain a query component (got %q)", s)
	}
	if parsed.Fragment != "" {
		return "", fmt.Errorf("issuerURI must not contain a fragment (got %q)", s)
	}
	return s, nil
}

// sha256Hex returns the hex-encoded SHA-256 hash of the input.
func sha256Hex(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
