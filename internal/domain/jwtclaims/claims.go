// Package jwtclaims normalizes JWT claims for CEL evaluation.
package jwtclaims

import "github.com/lestrrat-go/jwx/v4/jwt"

// FromToken converts a JWT into a dynamic claims map. Standard claims are normalized first;
// custom claims never overwrite their normalized representation.
func FromToken(token jwt.Token) map[string]any {
	claims := make(map[string]any)

	if iss, _ := token.Issuer(); iss != "" {
		claims["iss"] = iss
	}
	if sub, _ := token.Subject(); sub != "" {
		claims["sub"] = sub
	}
	if aud, _ := token.Audience(); len(aud) > 0 {
		if len(aud) == 1 {
			claims["aud"] = aud[0]
		} else {
			claims["aud"] = aud
		}
	}
	if exp, _ := token.Expiration(); !exp.IsZero() {
		claims["exp"] = exp.Unix()
	}
	if iat, _ := token.IssuedAt(); !iat.IsZero() {
		claims["iat"] = iat.Unix()
	}
	if nbf, _ := token.NotBefore(); !nbf.IsZero() {
		claims["nbf"] = nbf.Unix()
	}
	if jti, _ := token.JwtID(); jti != "" {
		claims["jti"] = jti
	}

	for key, value := range token.Claims() {
		if _, exists := claims[key]; !exists {
			claims[key] = value
		}
	}

	return claims
}
