// +build integration

package integration_test

import (
	"context"
	"testing"

	awsencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/aws"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
)

// TestLocalStackKMSEncryptDecryptRoundtrip tests encryption/decryption using LocalStack KMS
// This test creates a real KMS key in LocalStack and validates the entire envelope encryption flow.
// Run with: DOCKER_HOST=unix:///run/podman/podman.sock go test -v -tags=integration ./tests/integration -run "LocalStack"
func TestLocalStackKMSEncryptDecryptRoundtrip(t *testing.T) {
	ctx := context.Background()

	// Start LocalStack container with KMS and DynamoDB
	ls := bootstrap.StartLocalStack(ctx, t)
	defer ls.Terminate(ctx)
	defer ls.CleanupLocalStackEnvironment()

	// Configure AWS SDK to use LocalStack endpoint
	ls.SetupLocalStackEnvironment()

	// Create adapter with LocalStack KMS ARN
	kmsARN := "arn:aws:kms:eu-central-1:000000000000:key/" + ls.KMSKeyID
	adapter, err := awsencryption.NewAWSEncryptionAdapterWithConfig(kmsARN, "IdentityBrokerEncryptionBranchKeys", 0)
	if err != nil {
		t.Fatalf("failed to create adapter with LocalStack KMS: %v", err)
	}

	if adapter == nil {
		t.Fatal("adapter should not be nil")
	}

	// Test encryption/decryption roundtrip
	plaintext := []byte("test-oauth2-token-from-localstack")
	encryptionContext := map[string]string{
		"service_id": "oauth2",
		"principal":  "user@example.com",
	}

	// Encrypt
	ciphertext, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatal("ciphertext should not be empty")
	}

	// Verify ciphertext is different from plaintext
	if string(ciphertext) == string(plaintext) {
		t.Fatal("ciphertext should be different from plaintext")
	}

	// Decrypt
	decrypted, err := adapter.Decrypt(ctx, ciphertext, encryptionContext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	// Verify decrypted matches original
	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted plaintext mismatch: expected %q, got %q", string(plaintext), string(decrypted))
	}
}

// TestLocalStackContextMismatchDetection tests context verification with LocalStack KMS
// Verifies that context binding is enforced at both DEK and KEK layers.
func TestLocalStackContextMismatchDetection(t *testing.T) {
	ctx := context.Background()

	// Start LocalStack
	ls := bootstrap.StartLocalStack(ctx, t)
	defer ls.Terminate(ctx)
	defer ls.CleanupLocalStackEnvironment()

	// Configure AWS SDK to use LocalStack endpoint
	ls.SetupLocalStackEnvironment()

	// Create adapter with LocalStack KMS
	kmsARN := "arn:aws:kms:eu-central-1:000000000000:key/" + ls.KMSKeyID
	adapter, err := awsencryption.NewAWSEncryptionAdapterWithConfig(kmsARN, "IdentityBrokerEncryptionBranchKeys", 0)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	plaintext := []byte("test-oauth2-token")

	// Encrypt with service_id "oauth2"
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}

	ciphertext, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Try to decrypt with different service_id "github"
	wrongContext := map[string]string{
		"service_id": "github",
	}

	decrypted, err := adapter.Decrypt(ctx, ciphertext, wrongContext)

	// Decryption should fail due to context mismatch
	if err == nil {
		t.Fatal("expected decryption to fail with wrong context")
	}

	// Verify decrypted is nil/empty
	if decrypted != nil && len(decrypted) > 0 {
		t.Fatal("decrypted should be empty on context mismatch")
	}

	// Verify error is context mismatch type
	if !isContextMismatchError(err) {
		t.Errorf("expected ContextMismatch error, got: %v (type: %T)", err, err)
	}
}

// TestLocalStackUniqueEncryptionPerCall tests that LocalStack KMS generates fresh DEKs
// Verifies that two encryptions of the same plaintext produce different ciphertexts.
func TestLocalStackUniqueEncryptionPerCall(t *testing.T) {
	ctx := context.Background()

	// Start LocalStack
	ls := bootstrap.StartLocalStack(ctx, t)
	defer ls.Terminate(ctx)
	defer ls.CleanupLocalStackEnvironment()

	// Configure AWS SDK to use LocalStack endpoint
	ls.SetupLocalStackEnvironment()

	// Create adapter with LocalStack KMS
	kmsARN := "arn:aws:kms:eu-central-1:000000000000:key/" + ls.KMSKeyID
	adapter, err := awsencryption.NewAWSEncryptionAdapterWithConfig(kmsARN, "IdentityBrokerEncryptionBranchKeys", 0)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	plaintext := []byte("same-token-value")
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}

	// Encrypt same plaintext twice
	ciphertext1, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("first encryption failed: %v", err)
	}

	ciphertext2, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("second encryption failed: %v", err)
	}

	// Ciphertexts should be different (fresh DEK per encryption)
	if string(ciphertext1) == string(ciphertext2) {
		t.Fatal("ciphertexts should be different for same plaintext (unique DEK per call)")
	}

	// Both should decrypt to same plaintext
	decrypted1, err := adapter.Decrypt(ctx, ciphertext1, encryptionContext)
	if err != nil {
		t.Fatalf("first decryption failed: %v", err)
	}

	decrypted2, err := adapter.Decrypt(ctx, ciphertext2, encryptionContext)
	if err != nil {
		t.Fatalf("second decryption failed: %v", err)
	}

	if string(decrypted1) != string(plaintext) {
		t.Errorf("first decrypted mismatch: expected %q, got %q", string(plaintext), string(decrypted1))
	}

	if string(decrypted2) != string(plaintext) {
		t.Errorf("second decrypted mismatch: expected %q, got %q", string(plaintext), string(decrypted2))
	}
}

// TestLocalStackTamperedCiphertextDetection tests that LocalStack detects tampering
// Verifies that modifying ciphertext fails authentication tag verification.
func TestLocalStackTamperedCiphertextDetection(t *testing.T) {
	ctx := context.Background()

	// Start LocalStack
	ls := bootstrap.StartLocalStack(ctx, t)
	defer ls.Terminate(ctx)
	defer ls.CleanupLocalStackEnvironment()

	// Configure AWS SDK to use LocalStack endpoint
	ls.SetupLocalStackEnvironment()

	// Create adapter with LocalStack KMS
	kmsARN := "arn:aws:kms:eu-central-1:000000000000:key/" + ls.KMSKeyID
	adapter, err := awsencryption.NewAWSEncryptionAdapterWithConfig(kmsARN, "IdentityBrokerEncryptionBranchKeys", 0)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	plaintext := []byte("test-token-value")
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}

	// Encrypt
	ciphertext, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Tamper with first byte of ciphertext
	if len(ciphertext) > 0 {
		tamperedCiphertext := make([]byte, len(ciphertext))
		copy(tamperedCiphertext, ciphertext)
		tamperedCiphertext[0] = tamperedCiphertext[0] ^ 0xFF // Flip all bits
		ciphertext = tamperedCiphertext
	}

	// Try to decrypt tampered ciphertext
	decrypted, err := adapter.Decrypt(ctx, ciphertext, encryptionContext)

	// Decryption should fail due to authentication tag verification
	if err == nil {
		t.Fatal("expected decryption to fail for tampered ciphertext")
	}

	// Verify decrypted is nil/empty
	if decrypted != nil && len(decrypted) > 0 {
		t.Fatal("decrypted should be empty for tampered ciphertext")
	}

	// Error should be integrity violation or decryption failed
	if !isIntegrityViolationError(err) && !isDecryptionFailedError(err) {
		t.Errorf("expected IntegrityViolation or DecryptionFailed error, got: %v", err)
	}
}

// TestLocalStackMultipleServices tests cross-service token isolation
// Encrypts tokens for different services and verifies that context binding prevents cross-service usage.
func TestLocalStackMultipleServices(t *testing.T) {
	ctx := context.Background()

	// Start LocalStack
	ls := bootstrap.StartLocalStack(ctx, t)
	defer ls.Terminate(ctx)
	defer ls.CleanupLocalStackEnvironment()

	// Configure AWS SDK to use LocalStack endpoint
	ls.SetupLocalStackEnvironment()

	// Create adapter with LocalStack KMS
	kmsARN := "arn:aws:kms:eu-central-1:000000000000:key/" + ls.KMSKeyID
	adapter, err := awsencryption.NewAWSEncryptionAdapterWithConfig(kmsARN, "IdentityBrokerEncryptionBranchKeys", 0)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	// Create tokens for two services
	oauthToken := []byte("oauth2-access-token")
	githubToken := []byte("github-access-token")

	oauthContext := map[string]string{"service_id": "oauth2"}
	githubContext := map[string]string{"service_id": "github"}

	// Encrypt both tokens
	oauthCiphertext, err := adapter.Encrypt(ctx, oauthToken, oauthContext)
	if err != nil {
		t.Fatalf("failed to encrypt oauth2 token: %v", err)
	}

	githubCiphertext, err := adapter.Encrypt(ctx, githubToken, githubContext)
	if err != nil {
		t.Fatalf("failed to encrypt github token: %v", err)
	}

	// Verify tokens can only be decrypted with matching context
	// oauth2 token with oauth2 context: should succeed
	decrypted, err := adapter.Decrypt(ctx, oauthCiphertext, oauthContext)
	if err != nil {
		t.Fatalf("failed to decrypt oauth2 token with matching context: %v", err)
	}
	if string(decrypted) != string(oauthToken) {
		t.Errorf("decrypted oauth2 token mismatch: expected %q, got %q", string(oauthToken), string(decrypted))
	}

	// github token with github context: should succeed
	decrypted, err = adapter.Decrypt(ctx, githubCiphertext, githubContext)
	if err != nil {
		t.Fatalf("failed to decrypt github token with matching context: %v", err)
	}
	if string(decrypted) != string(githubToken) {
		t.Errorf("decrypted github token mismatch: expected %q, got %q", string(githubToken), string(decrypted))
	}

	// oauth2 token with github context: should fail
	decrypted, err = adapter.Decrypt(ctx, oauthCiphertext, githubContext)
	if err == nil {
		t.Fatal("expected decryption to fail when oauth2 token used with github context")
	}
	if decrypted != nil && len(decrypted) > 0 {
		t.Fatal("decrypted should be nil when context mismatch occurs")
	}

	// github token with oauth2 context: should fail
	decrypted, err = adapter.Decrypt(ctx, githubCiphertext, oauthContext)
	if err == nil {
		t.Fatal("expected decryption to fail when github token used with oauth2 context")
	}
	if decrypted != nil && len(decrypted) > 0 {
		t.Fatal("decrypted should be nil when context mismatch occurs")
	}
}

// Helper functions for error type checking

func isContextMismatchError(err error) bool {
	encErr, ok := err.(*encryption.EncryptionError)
	return ok && encErr.Kind == encryption.ErrorKindContextMismatch
}

func isIntegrityViolationError(err error) bool {
	encErr, ok := err.(*encryption.EncryptionError)
	return ok && encErr.Kind == encryption.ErrorKindIntegrityViolation
}

func isDecryptionFailedError(err error) bool {
	encErr, ok := err.(*encryption.EncryptionError)
	return ok && encErr.Kind == encryption.ErrorKindDecryptionFailed
}
