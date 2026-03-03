package jwt

import (
	"context"
	"crypto/rsa"

	"github.com/go-oauth2/oauth2/v4"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// CustomClaims extends the standard RegisteredClaims with an authorized party field.
type CustomClaims struct {
	jwtv5.RegisteredClaims
	AZP string `json:"azp,omitempty"`
}

// RSAAccessGenerator generates RS256-signed JWT access tokens with the required
// claims for the identity broker's token exchange service.
type RSAAccessGenerator struct {
	privateKey *rsa.PrivateKey
	issuerURI  string
	audience   string
	keyID      string
}

// NewRSAAccessGenerator creates a new RSAAccessGenerator.
func NewRSAAccessGenerator(privateKey *rsa.PrivateKey, issuerURI, audience, keyID string) *RSAAccessGenerator {
	return &RSAAccessGenerator{
		privateKey: privateKey,
		issuerURI:  issuerURI,
		audience:   audience,
		keyID:      keyID,
	}
}

// Token implements the oauth2.AccessGenerate interface. It returns (accessToken, refreshToken, error).
// refreshToken is a UUID string when isGenRefresh is true, otherwise empty.
func (g *RSAAccessGenerator) Token(ctx context.Context, data *oauth2.GenerateBasic, isGenRefresh bool) (string, string, error) {
	subject := data.UserID
	if subject == "" {
		subject = data.Client.GetID()
	}

	now := data.TokenInfo.GetAccessCreateAt()
	exp := now.Add(data.TokenInfo.GetAccessExpiresIn())

	claims := CustomClaims{
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    g.issuerURI,
			Subject:   subject,
			Audience:  jwtv5.ClaimStrings{g.audience},
			ExpiresAt: jwtv5.NewNumericDate(exp),
			IssuedAt:  jwtv5.NewNumericDate(now),
		},
		AZP: data.Client.GetID(),
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, claims)
	token.Header["kid"] = g.keyID

	accessToken, err := token.SignedString(g.privateKey)
	if err != nil {
		return "", "", err
	}

	refreshToken := ""
	if isGenRefresh {
		refreshToken = uuid.New().String()
	}

	return accessToken, refreshToken, nil
}
