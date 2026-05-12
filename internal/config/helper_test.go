package config

import (
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func generateBase64EncodedString(t testing.TB, bytes int) string {
	key := make([]byte, bytes)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Error generating random bytes: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key)
}
