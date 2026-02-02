package config

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

// Helper function to create valid base64-encoded random string of given byte length
func generateBase64EncodedString(bytes int) string {
	key := make([]byte, bytes)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalf("Error generating random bytes: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key)
}
