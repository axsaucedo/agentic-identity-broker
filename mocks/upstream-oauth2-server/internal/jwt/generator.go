package jwt

import (
	"github.com/go-oauth2/oauth2/v4/generates"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// NewJWTAccessGenerator creates a standard JWT access generator
func NewJWTAccessGenerator(secretKey []byte) *generates.JWTAccessGenerate {
	return generates.NewJWTAccessGenerate("", secretKey, jwtv5.SigningMethodHS512)
}
