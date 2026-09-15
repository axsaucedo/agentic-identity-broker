package impersonation

import (
	"context"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jws"
	"github.com/lestrrat-go/jwx/v4/jwt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwtclaims"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
)

// defaultClockSkew is the acceptable clock skew for impersonation credential validation.
const defaultClockSkew = 60 * time.Second

// signedValidator validates a signed impersonation credential (client assertion, actor token,
// or signed subject token) against one trusted issuer, enforcing an explicit per-issuer
// asymmetric algorithm allow-list (CR-007) in addition to signature, issuer, audience, expiry,
// and not-before checks (FR-004).
type signedValidator struct {
	jwksProvider      tokenexchange.JWKSProvider
	issuerURI         string
	allowedAlgorithms map[jwa.SignatureAlgorithm]struct{}
	clockSkew         time.Duration
}

// validate verifies the token against the issuer's keys and the expected audience, returning
// the validated claims. It rejects any algorithm not in the broker-approved asymmetric set or
// not explicitly allowed for the issuer BEFORE attempting signature verification (CR-007), so a
// symmetric or none algorithm can never be accepted for a signed role (FR-005).
func (v *signedValidator) validate(ctx context.Context, tokenString, expectedAudience string) (map[string]interface{}, error) {
	return v.validateCredential(ctx, tokenString, expectedAudience, false)
}

// validateWithoutAudience verifies the token and rejects any credential that carries an aud claim (CR-009).
func (v *signedValidator) validateWithoutAudience(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	return v.validateCredential(ctx, tokenString, "", true)
}

func (v *signedValidator) validateCredential(ctx context.Context, tokenString, expectedAudience string, requireAbsentAudience bool) (map[string]interface{}, error) {
	alg, err := protectedHeaderAlgorithm(tokenString)
	if err != nil {
		return nil, fmt.Errorf("cannot read token algorithm: %w", err)
	}
	if !isApprovedAlgorithm(alg) {
		return nil, fmt.Errorf("algorithm %q is not an approved asymmetric algorithm", alg)
	}
	if _, ok := v.allowedAlgorithms[alg]; !ok {
		return nil, fmt.Errorf("algorithm %q is not permitted for issuer", alg)
	}

	keyset, err := v.jwksProvider.GetKeySet(ctx)
	if err != nil {
		return nil, fmt.Errorf("jwks unavailable for issuer: %w", err)
	}

	var token jwt.Token
	if requireAbsentAudience {
		token, err = jwt.ParseString(
			tokenString,
			jwt.WithVerify(true),
			jwt.WithKeySet(keyset),
			jwt.WithRequiredClaim(jwt.ExpirationKey),
			jwt.WithValidate(true),
			jwt.WithIssuer(v.issuerURI),
			jwt.WithAcceptableSkew(v.clockSkew),
		)
	} else {
		token, err = jwt.ParseString(
			tokenString,
			jwt.WithVerify(true),
			jwt.WithKeySet(keyset),
			jwt.WithRequiredClaim(jwt.ExpirationKey),
			jwt.WithValidate(true),
			jwt.WithIssuer(v.issuerURI),
			jwt.WithAudience(expectedAudience),
			jwt.WithAcceptableSkew(v.clockSkew),
		)
	}
	if err != nil {
		return nil, err
	}
	if requireAbsentAudience {
		if _, ok := token.Audience(); ok {
			return nil, fmt.Errorf("token must not contain an audience claim")
		}
	}
	return jwtclaims.FromToken(token), nil
}

// soleSignature parses a compact JWS and returns its single signature without verifying it, so
// the protected-header algorithm (and, for the unverified subject, the empty signature segment)
// can be inspected before any trust decision.
func soleSignature(tokenString string) (*jws.Signature, error) {
	msg, err := jws.Parse([]byte(tokenString))
	if err != nil {
		return nil, fmt.Errorf("not a parseable compact JWS: %w", err)
	}
	sigs := msg.Signatures()
	if len(sigs) != 1 {
		return nil, fmt.Errorf("compact JWT must carry exactly one signature segment")
	}
	return sigs[0], nil
}

// protectedHeaderAlgorithm returns the typed alg parameter from a compact JWT's protected header,
// without verifying the signature, so the per-issuer allow-list can be enforced against the actual
// header value (CR-007).
func protectedHeaderAlgorithm(tokenString string) (jwa.SignatureAlgorithm, error) {
	sig, err := soleSignature(tokenString)
	if err != nil {
		return jwa.EmptySignatureAlgorithm(), err
	}
	alg, ok := sig.ProtectedHeaders().Algorithm()
	if !ok {
		return jwa.EmptySignatureAlgorithm(), fmt.Errorf("protected header has no alg")
	}
	return alg, nil
}
