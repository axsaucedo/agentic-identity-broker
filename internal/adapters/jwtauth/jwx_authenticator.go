// Package jwtauth provides the JWT authenticator adapter using lestrrat-go/jwx v3.
// This package implements the domain JWTAuthenticator port for signed JWT verification
// (JWKS mode) with JWKS caching, temporal validation, and CEL-based claim extraction.
package jwtauth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/lestrrat-go/jwx/v3/jwt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwtauth"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

const (
	// DefaultJWKSMinRefreshInterval is the default minimum JWKS cache refresh interval.
	DefaultJWKSMinRefreshInterval = 5 * time.Minute

	// DefaultJWKSMaxRefreshInterval is the default maximum JWKS cache refresh interval.
	DefaultJWKSMaxRefreshInterval = 15 * time.Minute
)

// JWXAuthenticatorConfig holds the configuration needed to create a JWXAuthenticator.
type JWXAuthenticatorConfig struct {
	// JWTConfig is the JWT configuration from the application config.
	JWTConfig *ports.JWTConfig

	// CELEvaluator is the domain CEL evaluator for claim extraction.
	CELEvaluator *jwtauth.CELEvaluator

	// HTTPClient is the HTTP client used for JWKS fetching.
	// If nil, http.DefaultClient is used.
	HTTPClient *http.Client

	// Logger is the structured logger for authentication events.
	Logger *slog.Logger
}

// JWXAuthenticator implements the domain JWTAuthenticator port using lestrrat-go/jwx v3.
// It handles:
//   - JWKS-based signature verification with auto-refresh caching
//   - Bearer prefix stripping for Authorization header
//   - Temporal validation (exp enforcement, aud/iss optional checks)
//   - CEL-based claim extraction via the domain CELEvaluator
//   - Error classification into domain sentinel errors
type JWXAuthenticator struct {
	config       *ports.JWTConfig
	celEvaluator *jwtauth.CELEvaluator
	jwksCache    *jwk.Cache
	jwksURL      string
	logger       *slog.Logger
}

// NewJWXAuthenticator creates a new JWX-based JWT authenticator.
// It initializes the JWKS cache (for "jwks" verification mode) and validates configuration.
func NewJWXAuthenticator(cfg JWXAuthenticatorConfig) (*JWXAuthenticator, error) {
	if cfg.JWTConfig == nil {
		return nil, fmt.Errorf("JWTConfig is required")
	}
	if cfg.CELEvaluator == nil {
		return nil, fmt.Errorf("CELEvaluator is required")
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	auth := &JWXAuthenticator{
		config:       cfg.JWTConfig,
		celEvaluator: cfg.CELEvaluator,
		jwksURL:      cfg.JWTConfig.JWKSURI,
		logger:       logger,
	}

	// Initialize JWKS cache for signed verification mode
	if cfg.JWTConfig.Verification == "jwks" || cfg.JWTConfig.Verification == "" {
		httpClient := cfg.HTTPClient
		if httpClient == nil {
			httpClient = http.DefaultClient
		}

		// Create httprc client for JWKS fetching (follows existing pattern from internal/adapters/jwks/adapter.go)
		httprcClient := httprc.NewClient()

		cache, err := jwk.NewCache(context.Background(), httprcClient)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWKS cache: %w", err)
		}

		err = cache.Register(
			context.Background(),
			cfg.JWTConfig.JWKSURI,
			jwk.WithMinInterval(DefaultJWKSMinRefreshInterval),
			jwk.WithMaxInterval(DefaultJWKSMaxRefreshInterval),
			jwk.WithWaitReady(false),
			jwk.WithHttprcResourceOption(httprc.WithHTTPClient(httpClient)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to register JWKS URL: %w", err)
		}

		// Force an initial refresh with a 30-second timeout so startup fails fast
		// if the JWKS endpoint is unreachable (fail-closed per spec FR-004).
		initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err = cache.Refresh(initCtx, cfg.JWTConfig.JWKSURI); err != nil {
			return nil, fmt.Errorf("failed to fetch JWKS on startup from %s (verify URL is reachable and returns valid JWKS): %w", cfg.JWTConfig.JWKSURI, err)
		}

		auth.jwksCache = cache
	}

	return auth, nil
}

// Authenticate implements jwtauth.JWTAuthenticator.
// It parses the raw JWT string, verifies signature (JWKS mode), validates temporal claims,
// and extracts principal + profile via CEL expressions.
// All validation failures are logged as structured audit events (FR-020/SR-005).
func (a *JWXAuthenticator) Authenticate(ctx context.Context, rawJWT string) (*jwtauth.AuthResult, error) {
	// Strip Bearer prefix for Authorization header
	tokenStr := a.stripBearerPrefix(rawJWT)

	// Reject empty tokens early
	if strings.TrimSpace(tokenStr) == "" {
		a.logger.Warn("JWT authentication failed: malformed token",
			"reason", "empty_token",
			"jwt_header", a.config.HeaderName,
		)
		return nil, jwtauth.ErrMalformedToken
	}

	// Parse and verify the JWT
	token, err := a.parseAndVerify(ctx, tokenStr)
	if err != nil {
		a.logValidationFailure(err)
		return nil, err
	}

	// Validate temporal claims
	if err := a.validateClaims(token); err != nil {
		a.logValidationFailure(err)
		return nil, err
	}

	// Extract claims map for CEL evaluation
	claimsMap, err := a.extractClaimsMap(token)
	if err != nil {
		a.logger.Warn("JWT authentication failed: claim extraction error",
			"error", err.Error(),
		)
		return nil, fmt.Errorf("%w: %v", jwtauth.ErrClaimExtraction, err)
	}

	// Use CEL evaluator to extract principal and profile attributes
	result, err := a.celEvaluator.ExtractClaims(claimsMap)
	if err != nil {
		a.logger.Warn("JWT authentication failed: CEL claim extraction error",
			"error", err.Error(),
		)
		return nil, err
	}

	a.logger.Info("JWT authentication successful",
		"principal", result.Principal,
		"has_display_name", result.DisplayName != nil,
		"has_email", result.Email != nil,
		"has_picture_url", result.PictureURL != nil,
	)

	return result, nil
}

// stripBearerPrefix removes the "Bearer " prefix from the token string
// when the configured header is "Authorization" (case-insensitive).
func (a *JWXAuthenticator) stripBearerPrefix(rawJWT string) string {
	if strings.EqualFold(a.config.HeaderName, "Authorization") {
		if strings.HasPrefix(rawJWT, "Bearer ") {
			return strings.TrimPrefix(rawJWT, "Bearer ")
		}
		if strings.HasPrefix(rawJWT, "bearer ") {
			return strings.TrimPrefix(rawJWT, "bearer ")
		}
	}
	return rawJWT
}

// parseAndVerify parses the JWT and verifies its signature using JWKS.
func (a *JWXAuthenticator) parseAndVerify(ctx context.Context, tokenStr string) (jwt.Token, error) {
	if a.jwksCache != nil {
		// JWKS verification mode: ensure cache is ready, then fetch keyset
		if !a.jwksCache.Ready(ctx, a.jwksURL) {
			if _, err := a.jwksCache.Refresh(ctx, a.jwksURL); err != nil {
				a.logger.Error("failed to fetch JWKS", "url", a.jwksURL, "error", err)
				return nil, fmt.Errorf("%w: JWKS fetch failed: %v", jwtauth.ErrJWKSUnavailable, err)
			}
		}

		keySet, err := a.jwksCache.Lookup(ctx, a.jwksURL)
		if err != nil {
			a.logger.Error("failed to fetch JWKS", "url", a.jwksURL, "error", err)
			return nil, fmt.Errorf("%w: JWKS fetch failed: %v", jwtauth.ErrJWKSUnavailable, err)
		}

		token, err := jwt.Parse([]byte(tokenStr),
			jwt.WithKeySet(keySet, jws.WithInferAlgorithmFromKey(true)),
			jwt.WithValidate(false), // We do our own validation below
		)
		if err != nil {
			return nil, a.classifyParseError(err)
		}
		return token, nil
	}

	// Unsigned mode: parse without verification
	token, err := jwt.Parse([]byte(tokenStr),
		jwt.WithVerify(false),
		jwt.WithValidate(false),
	)
	if err != nil {
		return nil, a.classifyParseError(err)
	}
	return token, nil
}

// validateClaims validates the expiration claim (exp) and optional aud/iss claims.
// When verification is "none", exp is validated only if present in the token.
// When verification is "jwks" (or the default), exp must be present.
func (a *JWXAuthenticator) validateClaims(token jwt.Token) error {
	exp, hasExp := token.Expiration()

	// In "none" mode, exp is optional — only validate if the claim is present.
	// In signed ("jwks") mode, exp is mandatory.
	if !hasExp || exp.IsZero() {
		if a.config.Verification != "none" {
			return jwtauth.ErrMissingExpiry
		}
		// none mode + no exp: skip expiry check
	} else if time.Now().After(exp) {
		// exp is present and the token has expired — reject regardless of mode
		return jwtauth.ErrTokenExpired
	}

	// Validate audience if configured
	if a.config.ExpectedAudience != "" {
		audiences, _ := token.Audience()
		if !containsAudience(audiences, a.config.ExpectedAudience) {
			return jwtauth.ErrAudienceMismatch
		}
	}

	// Validate issuer if configured
	if a.config.ExpectedIssuer != "" {
		iss, _ := token.Issuer()
		if iss != a.config.ExpectedIssuer {
			return jwtauth.ErrIssuerMismatch
		}
	}

	return nil
}

// containsAudience checks if the expected audience is present in the audience list.
func containsAudience(audiences []string, expected string) bool {
	for _, aud := range audiences {
		if aud == expected {
			return true
		}
	}
	return false
}

// extractClaimsMap converts a jwt.Token into a map[string]interface{} for CEL evaluation.
// Follows the pattern from internal/domain/tokenexchange/service.go jwtToClaims().
func (a *JWXAuthenticator) extractClaimsMap(token jwt.Token) (map[string]interface{}, error) {
	claims := make(map[string]interface{})

	// Standard claims (jwx v3: all accessors return (value, bool))
	if sub, ok := token.Subject(); ok && sub != "" {
		claims["sub"] = sub
	}
	if iss, ok := token.Issuer(); ok && iss != "" {
		claims["iss"] = iss
	}
	if aud, ok := token.Audience(); ok && len(aud) > 0 {
		claims["aud"] = aud
	}
	if exp, ok := token.Expiration(); ok && !exp.IsZero() {
		claims["exp"] = exp.Unix()
	}
	if iat, ok := token.IssuedAt(); ok && !iat.IsZero() {
		claims["iat"] = iat.Unix()
	}
	if nbf, ok := token.NotBefore(); ok && !nbf.IsZero() {
		claims["nbf"] = nbf.Unix()
	}
	if jti, ok := token.JwtID(); ok && jti != "" {
		claims["jti"] = jti
	}

	// Private/custom claims via Keys() and Get() (follows existing pattern)
	for _, key := range token.Keys() {
		if _, exists := claims[key]; !exists {
			var value any
			if err := token.Get(key, &value); err == nil {
				claims[key] = value
			}
		}
	}

	return claims, nil
}

// classifyParseError maps jwx parsing errors to domain sentinel errors.
func (a *JWXAuthenticator) classifyParseError(err error) error {
	errStr := strings.ToLower(err.Error())

	// Signature verification failures
	if strings.Contains(errStr, "signature") ||
		strings.Contains(errStr, "verification") ||
		strings.Contains(errStr, "could not verify") ||
		strings.Contains(errStr, "key not found") {
		return fmt.Errorf("%w: %v", jwtauth.ErrInvalidSignature, err)
	}

	// Malformed token
	return fmt.Errorf("%w: %v", jwtauth.ErrMalformedToken, err)
}

// logValidationFailure logs a structured audit event for JWT validation failures.
// Per FR-020/SR-005: all authentication failures are logged with structured context
// for security monitoring and incident response.
func (a *JWXAuthenticator) logValidationFailure(err error) {
	reason := "unknown"
	switch {
	case errors.Is(err, jwtauth.ErrJWKSUnavailable):
		reason = "jwks_unavailable"
	case errors.Is(err, jwtauth.ErrInvalidSignature):
		reason = "invalid_signature"
	case errors.Is(err, jwtauth.ErrTokenExpired):
		reason = "token_expired"
	case errors.Is(err, jwtauth.ErrAudienceMismatch):
		reason = "audience_mismatch"
	case errors.Is(err, jwtauth.ErrIssuerMismatch):
		reason = "issuer_mismatch"
	case errors.Is(err, jwtauth.ErrMissingExpiry):
		reason = "missing_expiry"
	case errors.Is(err, jwtauth.ErrMalformedToken):
		reason = "malformed_token"
	case errors.Is(err, jwtauth.ErrClaimExtraction):
		reason = "claim_extraction"
	}

	a.logger.Warn("JWT authentication failed",
		"reason", reason,
		"error", err.Error(),
		"jwt_header", a.config.HeaderName,
		"verification_mode", a.config.Verification,
	)
}
