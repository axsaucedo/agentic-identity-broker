package tokenexchange

import (
	"context"
	"fmt"
	"testing"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockJWKSProvider is a test double for the JWKSProvider interface
type MockJWKSProvider struct {
	keySet jwk.Set
	err    error
}

func (m *MockJWKSProvider) GetKeySet(ctx context.Context) (jwk.Set, error) {
	return m.keySet, m.err
}

func (m *MockJWKSProvider) GetKey(ctx context.Context, kid string) (jwk.Key, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.keySet == nil {
		return nil, NewInvalidClientError("no keyset")
	}
	key, found := m.keySet.LookupKeyID(kid)
	if !found {
		return nil, NewInvalidClientError("key not found")
	}
	return key, nil
}

// TestNewJWTValidator tests validator creation with various parameter combinations
func TestNewJWTValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		jwksProvider     JWKSProvider
		expectedIssuer   string
		brokerAudience   string
		clockSkewSeconds int64
		expectError      bool
		errorContains    string
	}{
		{
			name: "valid parameters",
			jwksProvider: &MockJWKSProvider{
				keySet: jwk.NewSet(),
			},
			expectedIssuer:   "https://auth.example.com",
			brokerAudience:   "broker-id",
			clockSkewSeconds: 60,
			expectError:      false,
		},
		{
			name: "zero clock skew",
			jwksProvider: &MockJWKSProvider{
				keySet: jwk.NewSet(),
			},
			expectedIssuer:   "https://auth.example.com",
			brokerAudience:   "broker-id",
			clockSkewSeconds: 0,
			expectError:      false,
		},
		{
			name:             "nil jwks provider",
			jwksProvider:     nil,
			expectedIssuer:   "https://auth.example.com",
			brokerAudience:   "broker-id",
			clockSkewSeconds: 60,
			expectError:      true,
			errorContains:    "cannot be nil",
		},
		{
			name: "empty issuer",
			jwksProvider: &MockJWKSProvider{
				keySet: jwk.NewSet(),
			},
			expectedIssuer:   "",
			brokerAudience:   "broker-id",
			clockSkewSeconds: 60,
			expectError:      true,
			errorContains:    "cannot be empty",
		},
		{
			name: "empty audience",
			jwksProvider: &MockJWKSProvider{
				keySet: jwk.NewSet(),
			},
			expectedIssuer:   "https://auth.example.com",
			brokerAudience:   "",
			clockSkewSeconds: 60,
			expectError:      true,
			errorContains:    "cannot be empty",
		},
		{
			name: "negative clock skew",
			jwksProvider: &MockJWKSProvider{
				keySet: jwk.NewSet(),
			},
			expectedIssuer:   "https://auth.example.com",
			brokerAudience:   "broker-id",
			clockSkewSeconds: -1,
			expectError:      true,
			errorContains:    "cannot be negative",
		},
		{
			name: "clock skew exceeds max",
			jwksProvider: &MockJWKSProvider{
				keySet: jwk.NewSet(),
			},
			expectedIssuer:   "https://auth.example.com",
			brokerAudience:   "broker-id",
			clockSkewSeconds: MaxClockSkewTolerance + 1,
			expectError:      true,
			errorContains:    "cannot exceed max tolerance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validator, err := NewJWTValidator(
				tt.jwksProvider,
				tt.expectedIssuer,
				tt.brokerAudience,
				tt.clockSkewSeconds,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, validator)
				if tt.errorContains != "" {
					assert.ErrorContains(t, err, tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, validator)
			}
		})
	}
}

// TestValidateSubjectToken_EmptyToken tests handling of empty token
func TestValidateSubjectToken_EmptyToken(t *testing.T) {
	t.Parallel()
	mockProvider := &MockJWKSProvider{keySet: jwk.NewSet()}
	validator, err := NewJWTValidator(mockProvider, "https://auth.example.com", "broker-id", 60)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = validator.ValidateSubjectToken(ctx, "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "empty or missing")
}

// TestValidateSubjectToken_MalformedToken tests handling of malformed token
func TestValidateSubjectToken_MalformedToken(t *testing.T) {
	t.Parallel()
	mockProvider := &MockJWKSProvider{keySet: jwk.NewSet()}
	validator, err := NewJWTValidator(mockProvider, "https://auth.example.com", "broker-id", 60)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = validator.ValidateSubjectToken(ctx, "not.a.jwt")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "malformed")
}

// TestValidateSubjectToken_JWKSFetchError tests handling of JWKS fetch error
func TestValidateSubjectToken_JWKSFetchError(t *testing.T) {
	t.Parallel()
	mockProvider := &MockJWKSProvider{err: NewServerError("connection failed")}
	validator, err := NewJWTValidator(mockProvider, "https://auth.example.com", "broker-id", 60)
	require.NoError(t, err)

	ctx := context.Background()
	// Use a valid JWT structure (but invalid for other reasons)
	// The test will fail at JWKS fetch step
	_, err = validator.ValidateSubjectToken(ctx, "eyJhbGciOiJIUzI1NiIsImtpZCI6InRlc3Qta2lkIn0.eyJzdWIiOiJ0ZXN0In0.test")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to fetch JWKS")
}

// TestValidateClientAssertion_EmptyToken tests handling of empty client_assertion
func TestValidateClientAssertion_EmptyToken(t *testing.T) {
	t.Parallel()
	mockProvider := &MockJWKSProvider{keySet: jwk.NewSet()}
	validator, err := NewJWTValidator(mockProvider, "https://auth.example.com", "broker-id", 60)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = validator.ValidateClientAssertion(ctx, "")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "empty or missing")
}

// TestValidateClientAssertion_MalformedToken tests handling of malformed client_assertion
func TestValidateClientAssertion_MalformedToken(t *testing.T) {
	t.Parallel()
	mockProvider := &MockJWKSProvider{keySet: jwk.NewSet()}
	validator, err := NewJWTValidator(mockProvider, "https://auth.example.com", "broker-id", 60)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = validator.ValidateClientAssertion(ctx, "not.a.jwt")

	assert.Error(t, err)
	assert.ErrorContains(t, err, "malformed")
}

// TestMapParseError_WrapsUnderlyingCause verifies that JWT validation errors include the
// underlying library error as a cause so operators can log full details internally.
func TestMapParseError_WrapsUnderlyingCause(t *testing.T) {
	t.Parallel()
	mockProvider := &MockJWKSProvider{keySet: jwk.NewSet()}
	validator, err := NewJWTValidator(mockProvider, "https://auth.example.com", "broker-id", 60)
	require.NoError(t, err)

	tests := []struct {
		name            string
		rawErr          error
		wantErrContains string
	}{
		{
			name:            "signature error wrapped",
			rawErr:          fmt.Errorf("signature verification failed: key mismatch"),
			wantErrContains: "malformed",
		},
		{
			name:            "issuer error wrapped",
			rawErr:          fmt.Errorf("issuer mismatch: got example.com"),
			wantErrContains: "issuer validation",
		},
		{
			name:            "audience error wrapped",
			rawErr:          fmt.Errorf("audience not satisfied"),
			wantErrContains: "audience validation",
		},
		{
			name:            "expiry error wrapped",
			rawErr:          fmt.Errorf("token is expired"),
			wantErrContains: "expired",
		},
		{
			name:            "unknown error wrapped in default case",
			rawErr:          fmt.Errorf("some totally unknown jwt failure"),
			wantErrContains: "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			domErr := validator.mapParseError(tt.rawErr, "subject_token")
			require.Error(t, domErr)

			// The domain error should contain the expected user-facing description
			assert.ErrorContains(t, domErr, tt.wantErrContains)

			// The underlying cause must be wrapped for internal logging
			tokenErr, ok := domErr.(*TokenExchangeError)
			require.True(t, ok, "expected *TokenExchangeError")
			require.NotNil(t, tokenErr.Unwrap(), "expected underlying cause to be set for logging")
			assert.Equal(t, tt.rawErr, tokenErr.Unwrap())
		})
	}
}

// TestMapClientAssertionParseError_WrapsUnderlyingCause verifies that client assertion validation
// errors include the underlying library error as a cause for internal logging.
func TestMapClientAssertionParseError_WrapsUnderlyingCause(t *testing.T) {
	t.Parallel()
	mockProvider := &MockJWKSProvider{keySet: jwk.NewSet()}
	validator, err := NewJWTValidator(mockProvider, "https://auth.example.com", "broker-id", 60)
	require.NoError(t, err)

	rawErr := fmt.Errorf("some unexpected validation problem")
	domErr := validator.mapClientAssertionParseError(rawErr)
	require.Error(t, domErr)

	tokenErr, ok := domErr.(*TokenExchangeError)
	require.True(t, ok, "expected *TokenExchangeError")
	require.NotNil(t, tokenErr.Unwrap(), "expected underlying cause to be set for logging")
	assert.Equal(t, rawErr, tokenErr.Unwrap())
}
