package helpers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega"

	"github.com/google/uuid"

	awsadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/aws"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
)

// NewEncryptionAdapter creates an encryption adapter for testing
// Supports both AWS KMS ARN and environment variable KEK references (${VAR})
func NewEncryptionAdapter(keyMaterial string) (ports.EncryptionPort, error) {
	adapter, _, err := awsadapter.NewAWSEncryption(keyMaterial, "", 0*time.Second)
	return adapter, err
}

// EncryptionTestHelper provides helper functions for encryption vault E2E testing
// This helper focuses on encryption port operations and will be extended when
// storage adapters are integrated with encryption in Phase 4-10
type EncryptionTestHelper struct {
	encryptionPort ports.EncryptionPort
	ctx            context.Context
}

// NewEncryptionTestHelper creates a new encryption test helper
func NewEncryptionTestHelper(encryptionPort ports.EncryptionPort, ctx context.Context) *EncryptionTestHelper {
	return &EncryptionTestHelper{
		encryptionPort: encryptionPort,
		ctx:            ctx,
	}
}

// CreateTestSession creates a test session with known data
// Returns the created session data for later verification
// NOTE: This will be extended in Phase 4-10 when storage integration is implemented
func (h *EncryptionTestHelper) CreateTestSession(principal, serviceID, accessToken, refreshToken string) (*fixtures.TestSessionData, error) {
	// Convert serviceID string to UUID
	svcUUID, err := uuid.Parse(serviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid service ID: %v", err)
	}

	// Create test session data with encryption context
	testData := &fixtures.TestSessionData{
		Principal:     principal,
		ServiceID:     svcUUID,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		EncryptionCtx: map[string]string{"service_id": serviceID},
	}

	// TODO Phase 4-10: Store via repository when storage integration is complete
	// For now, return the test data for verification in tests

	return testData, nil
}

// GetTestSession retrieves session data by ID
// NOTE: This will be implemented in Phase 4-10 when storage integration is complete
func (h *EncryptionTestHelper) GetTestSession(sessionID uuid.UUID) (*fixtures.TestSessionData, error) {
	// TODO Phase 4-10: Implement when storage repository integration is complete
	return nil, fmt.Errorf("not implemented - will be available in Phase 4-10 storage integration")
}

// CreateTestSessionFromFixture creates a session using fixture data
func (h *EncryptionTestHelper) CreateTestSessionFromFixture(serviceName string) (*fixtures.TestSessionData, error) {
	testData := fixtures.TestSessionWithService(serviceName)

	// TODO Phase 4-10: Store via repository when storage integration is complete
	// For now, return the fixture data for verification

	return testData, nil
}

// VerifyEncryptedToken verifies that ciphertext can be decrypted to expected plaintext
// using the provided encryption context
func (h *EncryptionTestHelper) VerifyEncryptedToken(ciphertext []byte, expectedPlaintext string, encryptionContext map[string]string) error {
	// Attempt to decrypt the ciphertext
	decrypted, err := h.encryptionPort.Decrypt(h.ctx, ciphertext, encryptionContext)
	if err != nil {
		return fmt.Errorf("decryption failed: %v", err)
	}

	// Verify the decrypted content matches expected plaintext
	if string(decrypted) != expectedPlaintext {
		return fmt.Errorf("decrypted content mismatch: expected %q, got %q", expectedPlaintext, string(decrypted))
	}

	return nil
}

// ExpectContextMismatchError verifies that decryption fails when using wrong context
// This is a negative test - we expect decryption to fail with context mismatch
func (h *EncryptionTestHelper) ExpectContextMismatchError(ciphertext []byte, wrongContext map[string]string) {
	// Attempt decryption with wrong context - this should fail
	_, err := h.encryptionPort.Decrypt(h.ctx, ciphertext, wrongContext)

	// Expect the decryption to fail due to context mismatch
	Expect(err).To(HaveOccurred(), "Expected decryption to fail with wrong context")
	Expect(err.Error()).To(ContainSubstring("context"), "Expected context-related error message")
}

// EncryptTestToken encrypts a test token with the given context
// Returns the ciphertext for use in other tests
func (h *EncryptionTestHelper) EncryptTestToken(plaintext string, encryptionContext map[string]string) ([]byte, error) {
	ciphertext, err := h.encryptionPort.Encrypt(h.ctx, []byte(plaintext), encryptionContext)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %v", err)
	}
	return ciphertext, nil
}

// DecryptTestToken decrypts a test token with the given context
// Returns plaintext or error if decryption fails
func (h *EncryptionTestHelper) DecryptTestToken(ciphertext []byte, encryptionContext map[string]string) ([]byte, error) {
	plaintext, err := h.encryptionPort.Decrypt(h.ctx, ciphertext, encryptionContext)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %v", err)
	}
	return plaintext, nil
}

// VerifyTokensEncryptedInDatabase verifies that tokens are stored encrypted, not as plaintext
// This is a security verification to ensure no plaintext leakage to database
func (h *EncryptionTestHelper) VerifyTokensEncryptedInDatabase(sessionID uuid.UUID, originalAccessToken string) error {
	// TODO Phase 4-10: This will require direct database access to verify ciphertext storage
	// Will be implemented when storage layer supports direct ciphertext inspection
	return fmt.Errorf("not implemented - requires direct database inspection capability (Phase 4-10)")
}

// VerifySessionTokensDecrypted verifies that session data has properly decrypted tokens
func (h *EncryptionTestHelper) VerifySessionTokensDecrypted(testData *fixtures.TestSessionData, expectedAccessToken, expectedRefreshToken string) {
	Expect(testData.AccessToken).To(Equal(expectedAccessToken), "Access token should match expected plaintext")
	Expect(testData.RefreshToken).To(Equal(expectedRefreshToken), "Refresh token should match expected plaintext")
}

// CreateMultipleTestSessions creates multiple sessions with different contexts for cross-context testing
func (h *EncryptionTestHelper) CreateMultipleTestSessions() (map[string]*fixtures.TestSessionData, error) {
	services := []string{"oauth2", "github", "google"}
	sessionData := make(map[string]*fixtures.TestSessionData)

	for _, service := range services {
		testData, err := h.CreateTestSessionFromFixture(service)
		if err != nil {
			return nil, fmt.Errorf("failed to create session for service %s: %v", service, err)
		}
		sessionData[service] = testData
	}

	return sessionData, nil
}

// VerifyUniqueEncryption verifies that the same token encrypted in different contexts
// produces different ciphertexts (ensuring DEK uniqueness and context binding)
func (h *EncryptionTestHelper) VerifyUniqueEncryption(plaintext string, contexts []map[string]string) error {
	ciphertexts := make([][]byte, len(contexts))

	// Encrypt the same plaintext with different contexts
	for i, ctx := range contexts {
		ciphertext, err := h.EncryptTestToken(plaintext, ctx)
		if err != nil {
			return fmt.Errorf("encryption failed for context %v: %v", ctx, err)
		}
		ciphertexts[i] = ciphertext
	}

	// Verify all ciphertexts are different
	for i := 0; i < len(ciphertexts); i++ {
		for j := i + 1; j < len(ciphertexts); j++ {
			if string(ciphertexts[i]) == string(ciphertexts[j]) {
				return fmt.Errorf("ciphertexts %d and %d are identical, but should be unique due to different contexts", i, j)
			}
		}
	}

	return nil
}

// GetTestEncryptionContexts returns test encryption contexts from fixtures
func (h *EncryptionTestHelper) GetTestEncryptionContexts() map[string]map[string]string {
	return fixtures.TestEncryptionContexts()
}

// GetTestKEKMaterial returns test KEK material for environment variable testing
func (h *EncryptionTestHelper) GetTestKEKMaterial() string {
	return fixtures.TestKEKMaterialDeterministic()
}

// GetTestTokenPairs returns test token pairs for encryption testing
func (h *EncryptionTestHelper) GetTestTokenPairs() map[string][2]string {
	return fixtures.TestTokenPairs()
}
