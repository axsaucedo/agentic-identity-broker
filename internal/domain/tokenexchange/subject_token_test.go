package tokenexchange

import (
	"strings"
	"testing"
	"time"
)

// TestSubjectToken_ValidMinimal tests a minimal valid subject token.
func TestSubjectToken_ValidMinimal(t *testing.T) {
	now := time.Now()
	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		now.Add(1*time.Hour).Unix(),
		now.Unix(),
		"",
		nil,
	)

	if err := token.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

// TestSubjectToken_ValidWithDifferentExtraction tests when CEL extracts from different claims.
func TestSubjectToken_ValidWithDifferentExtraction(t *testing.T) {
	now := time.Now()
	// CEL expression might extract principal from 'preferred_username' instead of 'sub'
	// and agent from 'client_id' instead of 'azp'
	token := NewSubjectToken(
		"alice",    // principal extracted from preferred_username via CEL
		"my-agent", // agentClientID extracted from client_id via CEL
		"https://upstream.example.com",
		"alice@example.com", // raw sub claim (different)
		"my-agent-123",      // raw azp claim (different)
		now.Add(1*time.Hour).Unix(),
		now.Unix(),
		"read write",
		nil,
	)

	if err := token.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}

	if token.GetPrincipal() != "alice" {
		t.Errorf("GetPrincipal() = %q, want 'alice'", token.GetPrincipal())
	}

	if token.GetAgentClientID() != "my-agent" {
		t.Errorf("GetAgentClientID() = %q, want 'my-agent'", token.GetAgentClientID())
	}
}

// TestSubjectToken_ValidWithCustomClaims tests token with custom claims.
func TestSubjectToken_ValidWithCustomClaims(t *testing.T) {
	now := time.Now()
	customClaims := map[string]interface{}{
		"org":    "acme",
		"team":   "platform",
		"roles":  []string{"admin", "developer"},
		"active": true,
	}

	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		now.Add(1*time.Hour).Unix(),
		now.Unix(),
		"",
		customClaims,
	)

	if err := token.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}

	if token.GetCustomClaim("org") != "acme" {
		t.Errorf("custom claim org not found")
	}
}

// TestSubjectToken_InvalidEmptyPrincipal tests validation with empty principal.
func TestSubjectToken_InvalidEmptyPrincipal(t *testing.T) {
	now := time.Now()
	token := NewSubjectToken(
		"", // empty principal
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		now.Add(1*time.Hour).Unix(),
		now.Unix(),
		"",
		nil,
	)

	err := token.Validate()
	if err == nil {
		t.Errorf("Validate() error = nil, want error for empty principal")
	}
	if err.Error() != "principal is required for subject token" {
		t.Errorf("error = %q, want 'principal is required for subject token'", err.Error())
	}
}

// TestSubjectToken_InvalidEmptyAgentClientID tests validation with empty agent client ID.
func TestSubjectToken_InvalidEmptyAgentClientID(t *testing.T) {
	now := time.Now()
	token := NewSubjectToken(
		"user@example.com",
		"", // empty agentClientID
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		now.Add(1*time.Hour).Unix(),
		now.Unix(),
		"",
		nil,
	)

	err := token.Validate()
	if err == nil {
		t.Errorf("Validate() error = nil, want error for empty agentClientID")
	}
	if err.Error() != "agent_client_id is required for subject token" {
		t.Errorf("error = %q, want 'agent_client_id is required for subject token'", err.Error())
	}
}

// TestSubjectToken_InvalidZeroExpiresAt tests validation with zero ExpiresAt.
func TestSubjectToken_InvalidZeroExpiresAt(t *testing.T) {
	now := time.Now()
	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		0, // zero expiresAt
		now.Unix(),
		"",
		nil,
	)

	err := token.Validate()
	if err == nil {
		t.Errorf("Validate() error = nil, want error for zero expiresAt")
	}
	if err.Error() != "expiresAt (exp claim) must be a valid Unix timestamp" {
		t.Errorf("error = %q, want 'expiresAt (exp claim) must be a valid Unix timestamp'", err.Error())
	}
}

// TestSubjectToken_InvalidZeroIssuedAt tests validation with zero IssuedAt.
func TestSubjectToken_InvalidZeroIssuedAt(t *testing.T) {
	now := time.Now()
	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		now.Add(1*time.Hour).Unix(),
		0, // zero issuedAt
		"",
		nil,
	)

	err := token.Validate()
	if err == nil {
		t.Errorf("Validate() error = nil, want error for zero issuedAt")
	}
	if err.Error() != "issuedAt (iat claim) must be a valid Unix timestamp" {
		t.Errorf("error = %q, want 'issuedAt (iat claim) must be a valid Unix timestamp'", err.Error())
	}
}

// TestSubjectToken_IsExpired tests expiration checking.
func TestSubjectToken_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		checkTime time.Time
		want      bool
	}{
		{
			name:      "not expired (future)",
			expiresAt: now.Add(1 * time.Hour),
			checkTime: now,
			want:      false,
		},
		{
			name:      "expired (past)",
			expiresAt: now.Add(-1 * time.Hour),
			checkTime: now,
			want:      true,
		},
		{
			name:      "expired (exactly now)",
			expiresAt: now,
			checkTime: now,
			want:      true, // <= comparison means exact time is expired
		},
		{
			name:      "not expired (near boundary)",
			expiresAt: now.Add(1 * time.Second),
			checkTime: now,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := NewSubjectToken(
				"user@example.com",
				"agent-123",
				"https://upstream.example.com",
				"user@example.com",
				"agent-123",
				tt.expiresAt.Unix(),
				tt.checkTime.Unix(),
				"",
				nil,
			)

			if token.IsExpired(tt.checkTime) != tt.want {
				t.Errorf("IsExpired() = %v, want %v", token.IsExpired(tt.checkTime), tt.want)
			}
		})
	}
}

// TestSubjectToken_GetCustomClaim tests custom claim retrieval.
func TestSubjectToken_GetCustomClaim(t *testing.T) {
	now := time.Now()
	customClaims := map[string]interface{}{
		"org":    "acme",
		"active": true,
	}

	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		now.Add(1*time.Hour).Unix(),
		now.Unix(),
		"",
		customClaims,
	)

	tests := []struct {
		name      string
		claimName string
		want      interface{}
	}{
		{
			name:      "existing string claim",
			claimName: "org",
			want:      "acme",
		},
		{
			name:      "existing bool claim",
			claimName: "active",
			want:      true,
		},
		{
			name:      "non-existent claim",
			claimName: "nonexistent",
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := token.GetCustomClaim(tt.claimName)
			if val != tt.want {
				t.Errorf("GetCustomClaim(%q) = %v, want %v", tt.claimName, val, tt.want)
			}
		})
	}
}

// TestSubjectToken_HasCustomClaim tests custom claim existence checking.
func TestSubjectToken_HasCustomClaim(t *testing.T) {
	now := time.Now()
	customClaims := map[string]interface{}{
		"org": "acme",
	}

	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		now.Add(1*time.Hour).Unix(),
		now.Unix(),
		"",
		customClaims,
	)

	if !token.HasCustomClaim("org") {
		t.Errorf("HasCustomClaim('org') = false, want true")
	}

	if token.HasCustomClaim("nonexistent") {
		t.Errorf("HasCustomClaim('nonexistent') = true, want false")
	}
}

// TestSubjectToken_GetAllClaims tests retrieving all claims as a map.
func TestSubjectToken_GetAllClaims(t *testing.T) {
	now := time.Now()
	customClaims := map[string]interface{}{
		"org": "acme",
	}

	expiresAt := now.Add(1 * time.Hour).Unix()
	issuedAt := now.Unix()

	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		expiresAt,
		issuedAt,
		"read write",
		customClaims,
	)

	claims := token.GetAllClaims()

	// Check standard claims
	if claims["sub"] != "user@example.com" {
		t.Errorf("claims['sub'] = %v, want 'user@example.com'", claims["sub"])
	}

	if claims["azp"] != "agent-123" {
		t.Errorf("claims['azp'] = %v, want 'agent-123'", claims["azp"])
	}

	if claims["iss"] != "https://upstream.example.com" {
		t.Errorf("claims['iss'] = %v, want 'https://upstream.example.com'", claims["iss"])
	}

	if claims["exp"] != expiresAt {
		t.Errorf("claims['exp'] = %v, want %d", claims["exp"], expiresAt)
	}

	if claims["iat"] != issuedAt {
		t.Errorf("claims['iat'] = %v, want %d", claims["iat"], issuedAt)
	}

	if claims["scope"] != "read write" {
		t.Errorf("claims['scope'] = %v, want 'read write'", claims["scope"])
	}

	// Check custom claims merged
	if claims["org"] != "acme" {
		t.Errorf("claims['org'] = %v, want 'acme'", claims["org"])
	}
}

// TestSubjectToken_String tests string representation doesn't expose sensitive data.
func TestSubjectToken_String(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour).Unix()

	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		expiresAt,
		now.Unix(),
		"secret-scope",
		map[string]interface{}{
			"secret": "data",
		},
	)

	str := token.String()

	// Should include non-sensitive metadata
	if !strings.Contains(str, "SubjectToken") {
		t.Errorf("String() should contain 'SubjectToken'")
	}
	if !strings.Contains(str, "user@example.com") {
		t.Errorf("String() should contain principal")
	}
	if !strings.Contains(str, "agent-123") {
		t.Errorf("String() should contain agent client ID")
	}
	if !strings.Contains(str, "https://upstream.example.com") {
		t.Errorf("String() should contain issuer")
	}

	// Should not include sensitive data
	if strings.Contains(str, "secret") {
		t.Errorf("String() should not contain custom claims data")
	}
}

// TestSubjectToken_GetToken tests CEL context retrieval.
func TestSubjectToken_GetToken(t *testing.T) {
	now := time.Now()
	customClaims := map[string]interface{}{
		"org": "acme",
	}

	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		now.Add(1*time.Hour).Unix(),
		now.Unix(),
		"",
		customClaims,
	)

	celContext := token.GetToken()

	// Should have same content as GetAllClaims
	if celContext["sub"] != "user@example.com" {
		t.Errorf("CEL context missing subject")
	}
	if celContext["org"] != "acme" {
		t.Errorf("CEL context missing custom claim")
	}
}

// TestSubjectToken_GetterMethods tests all getter methods.
func TestSubjectToken_GetterMethods(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour).Unix()
	issuedAt := now.Unix()

	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"sub-claim-value",
		"azp-claim-value",
		expiresAt,
		issuedAt,
		"read write",
		nil,
	)

	if token.GetPrincipal() != "user@example.com" {
		t.Errorf("GetPrincipal() = %q, want 'user@example.com'", token.GetPrincipal())
	}

	if token.GetAgentClientID() != "agent-123" {
		t.Errorf("GetAgentClientID() = %q, want 'agent-123'", token.GetAgentClientID())
	}

	if token.GetIssuer() != "https://upstream.example.com" {
		t.Errorf("GetIssuer() = %q, want 'https://upstream.example.com'", token.GetIssuer())
	}

	if token.GetSubject() != "sub-claim-value" {
		t.Errorf("GetSubject() = %q, want 'sub-claim-value'", token.GetSubject())
	}

	if token.GetAuthorizedParty() != "azp-claim-value" {
		t.Errorf("GetAuthorizedParty() = %q, want 'azp-claim-value'", token.GetAuthorizedParty())
	}

	if token.GetScope() != "read write" {
		t.Errorf("GetScope() = %q, want 'read write'", token.GetScope())
	}

	if token.GetExpiresAt() != expiresAt {
		t.Errorf("GetExpiresAt() = %d, want %d", token.GetExpiresAt(), expiresAt)
	}

	if token.GetIssuedAt() != issuedAt {
		t.Errorf("GetIssuedAt() = %d, want %d", token.GetIssuedAt(), issuedAt)
	}
}

// TestSubjectToken_NilCustomClaims tests handling of nil custom claims.
func TestSubjectToken_NilCustomClaims(t *testing.T) {
	now := time.Now()
	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		now.Add(1*time.Hour).Unix(),
		now.Unix(),
		"",
		nil,
	)

	// Should not panic and should return nil
	if token.GetCustomClaim("any") != nil {
		t.Errorf("GetCustomClaim should return nil for nil custom claims")
	}

	if token.HasCustomClaim("any") {
		t.Errorf("HasCustomClaim should return false for nil custom claims")
	}

	// GetAllClaims should still work
	claims := token.GetAllClaims()
	if claims == nil {
		t.Errorf("GetAllClaims should return non-nil map")
	}
}

// TestSubjectToken_MultipleAudiences tests with multiple custom claims.
func TestSubjectToken_MultipleAudiences(t *testing.T) {
	now := time.Now()
	customClaims := map[string]interface{}{
		"audiences": []string{"api1", "api2"},
		"roles":     []string{"admin", "user"},
	}

	token := NewSubjectToken(
		"user@example.com",
		"agent-123",
		"https://upstream.example.com",
		"user@example.com",
		"agent-123",
		now.Add(1*time.Hour).Unix(),
		now.Unix(),
		"",
		customClaims,
	)

	audiences := token.GetCustomClaim("audiences")
	if audiences == nil {
		t.Errorf("audiences custom claim should exist")
		return
	}

	// Verify it's a slice
	if aud, ok := audiences.([]string); ok {
		if len(aud) != 2 || aud[0] != "api1" {
			t.Errorf("audiences claim not preserved correctly")
		}
	} else {
		t.Errorf("audiences claim is not []string")
	}
}
