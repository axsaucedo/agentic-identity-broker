package storage

import "time"

// PKCESession stores the PKCE code challenge and method for a pending authorization code.
// Keyed by the fosite code signature (opaque string, not the raw code).
// Deleted immediately after the token endpoint consumes it.
type PKCESession struct {
	Signature           string    `db:"signature"`
	CodeChallenge       string    `db:"code_challenge"`
	CodeChallengeMethod string    `db:"code_challenge_method"`
	ExpiresAt           time.Time `db:"expires_at"`
	CreatedAt           time.Time `db:"created_at"`
}

// Validate checks that all required fields are present.
func (p *PKCESession) Validate() error {
	if p.Signature == "" {
		return NewStorageError("PKCESession.Validate", ErrorKindValidation, nil, "signature is required")
	}
	if p.CodeChallenge == "" {
		return NewStorageError("PKCESession.Validate", ErrorKindValidation, nil, "code_challenge is required")
	}
	if p.CodeChallengeMethod == "" {
		return NewStorageError("PKCESession.Validate", ErrorKindValidation, nil, "code_challenge_method is required")
	}
	if p.ExpiresAt.IsZero() {
		return NewStorageError("PKCESession.Validate", ErrorKindValidation, nil, "expires_at is required")
	}
	return nil
}
