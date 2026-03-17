package model

import (
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2/google"
)

// maxGoogleServiceAccountKeySize is the maximum allowed size for a Google service account
// JSON credential (32 KB). Enforces SR-004 to prevent denial-of-service via large payloads.
const maxGoogleServiceAccountKeySize = 32 * 1024 // 32 KB

// googleOAuthAuthURL is the Google OAuth2 authorization endpoint URL.
// Used when building OAuth2Endpoints from a Google service account credential.
const googleOAuthAuthURL = "https://accounts.google.com/o/oauth2/auth"

// GoogleServiceAccountKey is a parsed and validated Google service account JSON key.
// Required fields: type (must be "service_account"), private_key (non-empty),
// client_email (non-empty), token_uri (non-empty), client_id (non-empty).
//
// Note: private_key is stored as raw bytes from the JSON string value.
// Cryptographic validity of the key is NOT verified (per FR-014).
type GoogleServiceAccountKey struct {
	Type        string // always "service_account"
	ClientEmail string // from json:"client_email"
	PrivateKey  []byte // from json:"private_key", non-empty, not crypto-validated
	TokenURI    string // from json:"token_uri"
	ClientID    string // from json:"client_id"
}

// googleServiceAccountClientID is an auxiliary struct used only to extract the client_id
// and token_uri fields from the service account JSON, since google.JWTConfigFromJSON
// does not expose them fully (TokenURL gets a default, client_id is not exposed).
type googleServiceAccountClientID struct {
	ClientID string `json:"client_id"`
	TokenURI string `json:"token_uri"`
}

// ParseGoogleServiceAccountKey parses and validates a Google service account JSON string.
//
// Validation steps (in order):
//  1. Size check: credential must be ≤ 32 KB (FR-012, SR-004)
//  2. JSON validity and type check: google.JWTConfigFromJSON must succeed (FR-007, FR-008)
//  3. Field presence: client_email, private_key, token_uri must be non-empty (FR-005, FR-014)
//  4. client_id extraction: auxiliary unmarshal for client_id, must be non-empty (FR-005)
//
// Returns a populated GoogleServiceAccountKey or a detailed error naming the failing field(s).
// The private key is stored as raw bytes; its cryptographic format is not validated (FR-014).
// The private key value is never included in error messages (SR-002, SR-003).
func ParseGoogleServiceAccountKey(credential string) (*GoogleServiceAccountKey, error) {
	// Step 1: Size check FIRST before any JSON parsing (SR-004)
	if len(credential) > maxGoogleServiceAccountKeySize {
		return nil, fmt.Errorf("credential exceeds maximum size of 32 KB")
	}

	credBytes := []byte(credential)

	// Step 2: Parse JSON and validate type == "service_account" via google.JWTConfigFromJSON.
	// This returns an error for invalid JSON and for type != "service_account".
	jwtConfig, err := google.JWTConfigFromJSON(credBytes)
	if err != nil {
		// google.JWTConfigFromJSON returns errors mentioning "service_account" for wrong type.
		// Wrap into a user-friendly message that does not include private key material.
		return nil, fmt.Errorf("invalid Google service account JSON: %w", err)
	}

	// Step 3: Validate non-emptiness of required fields.
	// google.JWTConfigFromJSON populates Email (client_email), PrivateKey (private_key bytes),
	// and TokenURL (token_uri). Check each is non-empty.
	if jwtConfig.Email == "" {
		return nil, fmt.Errorf("client_email is required (must be a non-empty string)")
	}
	if len(jwtConfig.PrivateKey) == 0 {
		return nil, fmt.Errorf("private_key is required (must be non-empty)")
	}
	if jwtConfig.TokenURL == "" {
		return nil, fmt.Errorf("token_uri is required (must be a non-empty string)")
	}

	// Step 4: Extract client_id and token_uri via auxiliary unmarshal.
	// google.JWTConfigFromJSON defaults TokenURL when token_uri is absent/empty, so we
	// check the raw JSON value to detect missing/empty token_uri explicitly.
	var aux googleServiceAccountClientID
	if err := json.Unmarshal(credBytes, &aux); err != nil {
		// JSON is already valid (passed Step 2), so this should not happen.
		return nil, fmt.Errorf("failed to extract fields from service account JSON: %w", err)
	}
	if aux.TokenURI == "" {
		return nil, fmt.Errorf("token_uri is required (must be a non-empty string)")
	}
	if aux.ClientID == "" {
		return nil, fmt.Errorf("client_id is required in Google service account JSON")
	}

	return &GoogleServiceAccountKey{
		Type:        "service_account",
		ClientEmail: jwtConfig.Email,
		PrivateKey:  jwtConfig.PrivateKey,
		TokenURI:    aux.TokenURI,
		ClientID:    aux.ClientID,
	}, nil
}
