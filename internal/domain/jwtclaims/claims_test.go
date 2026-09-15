package jwtclaims

import (
	"iter"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type claimCollisionToken struct {
	jwt.Token
}

func (claimCollisionToken) Claims() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		yield(jwt.SubjectKey, "custom-subject")
	}
}

func TestFromToken_NormalizesStandardAndCustomClaims(t *testing.T) {
	expiresAt := time.Date(2025, 1, 2, 3, 4, 5, 600, time.UTC)
	issuedAt := time.Date(2025, 1, 2, 2, 4, 5, 600, time.UTC)
	notBefore := time.Date(2025, 1, 2, 1, 4, 5, 600, time.UTC)
	token, err := jwt.NewBuilder().
		Issuer("https://issuer.example.com").
		Subject("subject-1").
		Audience([]string{"audience-1"}).
		Expiration(expiresAt).
		IssuedAt(issuedAt).
		NotBefore(notBefore).
		JwtID("token-1").
		Claim("custom", "value").
		Build()
	require.NoError(t, err)

	claims := FromToken(token)

	assert.Equal(t, "https://issuer.example.com", claims["iss"])
	assert.Equal(t, "subject-1", claims["sub"])
	assert.Equal(t, "audience-1", claims["aud"])
	assert.Equal(t, expiresAt.Unix(), claims["exp"])
	assert.Equal(t, issuedAt.Unix(), claims["iat"])
	assert.Equal(t, notBefore.Unix(), claims["nbf"])
	assert.Equal(t, "token-1", claims["jti"])
	assert.Equal(t, "value", claims["custom"])
}

func TestFromToken_PreservesMultipleAudienceShape(t *testing.T) {
	token, err := jwt.NewBuilder().Audience([]string{"audience-1", "audience-2"}).Build()
	require.NoError(t, err)

	claims := FromToken(token)

	assert.Equal(t, []string{"audience-1", "audience-2"}, claims["aud"])
}

func TestFromToken_DoesNotOverwriteNormalizedStandardClaims(t *testing.T) {
	token, err := jwt.NewBuilder().Subject("normalized-subject").Build()
	require.NoError(t, err)

	claims := FromToken(claimCollisionToken{Token: token})

	assert.Equal(t, "normalized-subject", claims["sub"])
}
