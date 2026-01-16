package aws

import (
	"context"
	"os"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"
)

// TestNewAWSEncryptionAdapterWithEnvVar tests adapter creation with environment variable KEK
func TestNewAWSEncryptionAdapterWithEnvVar(t *testing.T) {
	// Setup: Create test KEK material (base64-encoded 32 bytes)
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("TEST_KEK", testKEK)
	defer os.Unsetenv("TEST_KEK")

	// Test: Create adapter with environment variable reference
	adapter, err := NewAWSEncryptionAdapter("${TEST_KEK}")

	// Verify: Adapter created successfully
	if err != nil {
		t.Fatalf("failed to create adapter with env var KEK: %v", err)
	}
	if adapter == nil {
		t.Fatal("adapter should not be nil")
	}
}

// TestNewAWSEncryptionAdapterWithEnvVarMissing tests adapter creation fails when env var not set
func TestNewAWSEncryptionAdapterWithEnvVarMissing(t *testing.T) {
	// Ensure env var is not set
	os.Unsetenv("MISSING_KEK_ENV_VAR")

	// Test: Try to create adapter with missing environment variable
	adapter, err := NewAWSEncryptionAdapter("${MISSING_KEK_ENV_VAR}")

	// Verify: Error returned
	if err == nil {
		t.Fatal("expected error when environment variable not set")
	}
	if adapter != nil {
		t.Fatal("adapter should be nil when environment variable not set")
	}

	// Verify: Error is KEKUnavailable type
	var kekErr *encryption.EncryptionError
	if !isKEKUnavailableError(err) {
		t.Errorf("expected KEKUnavailable error, got: %v", err)
	}
	_ = kekErr // silence unused
}

// TestNewAWSEncryptionAdapterWithInvalidEnvVarFormat tests adapter rejects invalid base64 KEK
func TestNewAWSEncryptionAdapterWithInvalidEnvVarFormat(t *testing.T) {
	// Setup: Create invalid base64 KEK material
	os.Setenv("INVALID_KEK", "not-valid-base64!!!")
	defer os.Unsetenv("INVALID_KEK")

	// Test: Try to create adapter with invalid base64
	adapter, err := NewAWSEncryptionAdapter("${INVALID_KEK}")

	// Verify: Error returned
	if err == nil {
		t.Fatal("expected error for invalid base64 KEK")
	}
	if adapter != nil {
		t.Fatal("adapter should be nil for invalid base64 KEK")
	}

	// Verify: Error is KEKUnavailable type
	if !isKEKUnavailableError(err) {
		t.Errorf("expected KEKUnavailable error, got: %v", err)
	}
}

// TestNewAWSEncryptionAdapterWithInvalidKEKLength tests adapter rejects KEK with wrong length
func TestNewAWSEncryptionAdapterWithInvalidKEKLength(t *testing.T) {
	// Setup: Create 16-byte KEK (too short, need 32)
	// "SGVsbG8gV29ybGQgSGVsbG8gV29ybGQ=" is 16 bytes base64-encoded
	os.Setenv("SHORT_KEK", "SGVsbG8gV29ybGQgSGVsbG8gV29ybGQ=")
	defer os.Unsetenv("SHORT_KEK")

	// Test: Try to create adapter with 16-byte KEK
	adapter, err := NewAWSEncryptionAdapter("${SHORT_KEK}")

	// Verify: Error returned
	if err == nil {
		t.Fatal("expected error for 16-byte KEK (need 32)")
	}
	if adapter != nil {
		t.Fatal("adapter should be nil for invalid KEK length")
	}

	// Verify: Error is KEKUnavailable type
	if !isKEKUnavailableError(err) {
		t.Errorf("expected KEKUnavailable error, got: %v", err)
	}
}

// TestNewAWSEncryptionAdapterWithInvalidFormat tests adapter rejects invalid key material format
func TestNewAWSEncryptionAdapterWithInvalidFormat(t *testing.T) {
	tests := []struct {
		name        string
		keyMaterial string
	}{
		{"empty string", ""},
		{"invalid format", "some-invalid-key"},
		{"partial KMS ARN", "arn:aws:kms:"},
		{"partial env var", "${INCOMPLETE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewAWSEncryptionAdapter(tt.keyMaterial)

			if err == nil {
				t.Fatal("expected error for invalid format")
			}
			if adapter != nil {
				t.Fatal("adapter should be nil for invalid format")
			}

			if !isKEKUnavailableError(err) {
				t.Errorf("expected KEKUnavailable error, got: %v", err)
			}
		})
	}
}

// TestEncryptDecryptRoundtrip tests encryption and decryption of tokens
func TestEncryptDecryptRoundtrip(t *testing.T) {
	// Setup: Create adapter with env var KEK
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("TEST_KEK_ROUNDTRIP", testKEK)
	defer os.Unsetenv("TEST_KEK_ROUNDTRIP")

	adapter, err := NewAWSEncryptionAdapter("${TEST_KEK_ROUNDTRIP}")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()
	plaintext := []byte("test-oauth2-token-value")
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}

	// Test: Encrypt
	ciphertext, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Verify: Ciphertext is not empty
	if len(ciphertext) == 0 {
		t.Fatal("ciphertext should not be empty")
	}

	// Verify: Ciphertext is different from plaintext
	if string(ciphertext) == string(plaintext) {
		t.Fatal("ciphertext should be different from plaintext")
	}

	// Test: Decrypt
	decrypted, err := adapter.Decrypt(ctx, ciphertext, encryptionContext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	// Verify: Decrypted matches original plaintext
	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted plaintext mismatch: expected %q, got %q", string(plaintext), string(decrypted))
	}
}

// TestContextMismatchDetection tests that wrong context fails decryption
func TestContextMismatchDetection(t *testing.T) {
	// Setup: Create adapter with env var KEK
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("TEST_KEK_CONTEXT", testKEK)
	defer os.Unsetenv("TEST_KEK_CONTEXT")

	adapter, err := NewAWSEncryptionAdapter("${TEST_KEK_CONTEXT}")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()
	plaintext := []byte("test-oauth2-token")

	// Encrypt with service_id "oauth2"
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}
	ciphertext, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Test: Try to decrypt with different service_id "github"
	wrongContext := map[string]string{
		"service_id": "github",
	}
	decrypted, err := adapter.Decrypt(ctx, ciphertext, wrongContext)

	// Verify: Decryption fails
	if err == nil {
		t.Fatal("expected decryption to fail with wrong context")
	}

	// Verify: Error is ContextMismatch type
	if !isContextMismatchError(err) {
		t.Errorf("expected ContextMismatch error, got: %v", err)
	}

	// Verify: Decrypted is nil or empty
	if decrypted != nil && len(decrypted) > 0 {
		t.Fatal("decrypted should be empty/nil on context mismatch")
	}
}

// TestUniqueEncryptionPerCall tests that same plaintext produces different ciphertexts
func TestUniqueEncryptionPerCall(t *testing.T) {
	// Setup: Create adapter with env var KEK
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("TEST_KEK_UNIQUE", testKEK)
	defer os.Unsetenv("TEST_KEK_UNIQUE")

	adapter, err := NewAWSEncryptionAdapter("${TEST_KEK_UNIQUE}")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()
	plaintext := []byte("same-token-value")
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}

	// Test: Encrypt same plaintext twice
	ciphertext1, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("first encryption failed: %v", err)
	}

	ciphertext2, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("second encryption failed: %v", err)
	}

	// Verify: Ciphertexts are different (fresh DEK per encryption)
	if string(ciphertext1) == string(ciphertext2) {
		t.Fatal("ciphertexts should be different for same plaintext (unique DEK per call)")
	}

	// Verify: Both can be decrypted to same plaintext
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

// TestEncryptEmptyPlaintext tests encryption of empty plaintext
func TestEncryptEmptyPlaintext(t *testing.T) {
	// Setup: Create adapter with env var KEK
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("TEST_KEK_EMPTY", testKEK)
	defer os.Unsetenv("TEST_KEK_EMPTY")

	adapter, err := NewAWSEncryptionAdapter("${TEST_KEK_EMPTY}")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()
	emptyPlaintext := []byte{}
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}

	// Test: Encrypt empty plaintext
	ciphertext, err := adapter.Encrypt(ctx, emptyPlaintext, encryptionContext)

	// AWS SDK should handle empty plaintext (may succeed or fail depending on configuration)
	// At minimum, check that error handling doesn't panic
	if err != nil {
		t.Logf("encryption of empty plaintext returned error (expected): %v", err)
	} else {
		t.Logf("encryption of empty plaintext succeeded (ciphertext length: %d)", len(ciphertext))
	}
}

// TestDecryptEmptyCiphertext tests that empty ciphertext fails
func TestDecryptEmptyCiphertext(t *testing.T) {
	// Setup: Create adapter with env var KEK
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("TEST_KEK_EMPTY_CIPHER", testKEK)
	defer os.Unsetenv("TEST_KEK_EMPTY_CIPHER")

	adapter, err := NewAWSEncryptionAdapter("${TEST_KEK_EMPTY_CIPHER}")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()
	emptyCiphertext := []byte{}
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}

	// Test: Try to decrypt empty ciphertext
	decrypted, err := adapter.Decrypt(ctx, emptyCiphertext, encryptionContext)

	// Verify: Decryption fails
	if err == nil {
		t.Fatal("expected decryption to fail with empty ciphertext")
	}

	// Verify: Decrypted is nil
	if decrypted != nil {
		t.Fatal("decrypted should be nil for empty ciphertext")
	}

	// Verify: Error is DecryptionFailed type
	if !isDecryptionFailedError(err) {
		t.Errorf("expected DecryptionFailed error, got: %v", err)
	}
}

// TestTamperedCiphertextDetection tests that tampered ciphertext fails
func TestTamperedCiphertextDetection(t *testing.T) {
	// Setup: Create adapter with env var KEK
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("TEST_KEK_TAMPER", testKEK)
	defer os.Unsetenv("TEST_KEK_TAMPER")

	adapter, err := NewAWSEncryptionAdapter("${TEST_KEK_TAMPER}")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()
	plaintext := []byte("test-token-value")
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}

	// Encrypt
	ciphertext, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Test: Tamper with first byte of ciphertext
	if len(ciphertext) > 0 {
		tamperedCiphertext := make([]byte, len(ciphertext))
		copy(tamperedCiphertext, ciphertext)
		tamperedCiphertext[0] = tamperedCiphertext[0] ^ 0xFF // Flip all bits
		ciphertext = tamperedCiphertext
	}

	// Test: Try to decrypt tampered ciphertext
	decrypted, err := adapter.Decrypt(ctx, ciphertext, encryptionContext)

	// Verify: Decryption fails (authentication tag verification)
	if err == nil {
		t.Fatal("expected decryption to fail for tampered ciphertext")
	}

	// Verify: Error is IntegrityViolation or DecryptionFailed type
	if !isIntegrityViolationError(err) && !isDecryptionFailedError(err) {
		t.Errorf("expected IntegrityViolation or DecryptionFailed error, got: %v", err)
	}

	// Verify: Decrypted is nil
	if decrypted != nil && len(decrypted) > 0 {
		t.Fatal("decrypted should be empty for tampered ciphertext")
	}
}

// Helper functions for error type checking

func isKEKUnavailableError(err error) bool {
	encErr, ok := err.(*encryption.EncryptionError)
	return ok && encErr.Kind == encryption.ErrorKindKEKUnavailable
}

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

func isEncryptionFailedError(err error) bool {
	encErr, ok := err.(*encryption.EncryptionError)
	return ok && encErr.Kind == encryption.ErrorKindEncryptionFailed
}

// BenchmarkEncrypt benchmarks encryption performance
func BenchmarkEncrypt(b *testing.B) {
	// Setup: Create adapter with env var KEK
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("BENCH_KEK", testKEK)
	defer os.Unsetenv("BENCH_KEK")

	adapter, err := NewAWSEncryptionAdapter("${BENCH_KEK}")
	if err != nil {
		b.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()
	plaintext := []byte("test-oauth2-access-token-value-1234567890")
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
		if err != nil {
			b.Fatalf("encryption failed: %v", err)
		}
	}
}

// BenchmarkDecrypt benchmarks decryption performance
func BenchmarkDecrypt(b *testing.B) {
	// Setup: Create adapter and prepare ciphertext
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("BENCH_KEK_DECRYPT", testKEK)
	defer os.Unsetenv("BENCH_KEK_DECRYPT")

	adapter, err := NewAWSEncryptionAdapter("${BENCH_KEK_DECRYPT}")
	if err != nil {
		b.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()
	plaintext := []byte("test-oauth2-access-token-value-1234567890")
	encryptionContext := map[string]string{
		"service_id": "oauth2",
	}

	// Prepare ciphertext
	ciphertext, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		b.Fatalf("encryption failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := adapter.Decrypt(ctx, ciphertext, encryptionContext)
		if err != nil {
			b.Fatalf("decryption failed: %v", err)
		}
	}
}

// TestKMSHierarchicalKeyringAdapterCreation tests adapter creation with KMS ARN
// This tests the hierarchical keyring implementation (not env var KEK)
func TestKMSHierarchicalKeyringAdapterCreation(t *testing.T) {
	// Test with invalid KMS ARN (will fail on accessibility check in test env)
	// This validates the hierarchical keyring initialization flow
	invalidKMSARN := "arn:aws:kms:us-west-2:123456789012:key/invalid-test-key"

	adapter, err := NewAWSEncryptionAdapter(invalidKMSARN)

	// Expected to fail due to invalid KMS key in test environment
	if err == nil {
		t.Fatal("expected error for invalid KMS ARN in test environment")
	}

	if adapter != nil {
		t.Fatal("adapter should be nil when KMS key is not accessible")
	}

	// Verify error is KEKUnavailable type
	if !isKEKUnavailableError(err) {
		t.Errorf("expected KEKUnavailable error, got: %v", err)
	}
}

// TestKMSARNValidation tests that KMS ARN format validation works
func TestKMSARNValidation(t *testing.T) {
	tests := []struct {
		name       string
		keyMaterial string
		shouldFail bool
	}{
		{
			name:        "valid_kms_arn_format",
			keyMaterial: "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
			shouldFail:  true, // Will fail on KMS accessibility in test, not format
		},
		{
			name:        "kms_alias_arn_format",
			keyMaterial: "arn:aws:kms:us-east-1:123456789012:alias/my-encryption-key",
			shouldFail:  true, // Will fail on KMS accessibility in test, not format
		},
		{
			name:        "invalid_kms_arn_format",
			keyMaterial: "arn:aws:s3:::my-bucket", // S3 ARN, not KMS
			shouldFail:  true, // Will fail on ARN format validation
		},
		{
			name:        "malformed_arn",
			keyMaterial: "not-an-arn",
			shouldFail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewAWSEncryptionAdapter(tt.keyMaterial)

			if !tt.shouldFail && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if tt.shouldFail && err == nil {
				t.Errorf("expected error, got nil")
			}

			if adapter != nil && !tt.shouldFail {
				// Only expect a valid adapter if we expected success
				if adapter.keyring == nil {
					t.Error("adapter keyring should not be nil")
				}
				if adapter.encryptionClient == nil {
					t.Error("adapter encryptionClient should not be nil")
				}
			}
		})
	}
}

// TestHierarchicalKeyringWithEnvVarFallback tests that env var KEK still works as fallback
func TestHierarchicalKeyringWithEnvVarFallback(t *testing.T) {
	// Test that environment variable KEK path still works (non-hierarchical)
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("FALLBACK_KEK", testKEK)
	defer os.Unsetenv("FALLBACK_KEK")

	adapter, err := NewAWSEncryptionAdapter("${FALLBACK_KEK}")

	// Should succeed with env var KEK
	if err != nil {
		t.Fatalf("failed to create adapter with env var KEK: %v", err)
	}

	if adapter == nil {
		t.Fatal("adapter should not be nil")
	}

	if adapter.keyring == nil {
		t.Error("adapter keyring should not be nil")
	}

	if adapter.encryptionClient == nil {
		t.Error("adapter encryptionClient should not be nil")
	}

	// Test that encryption/decryption still works with env var KEK
	ctx := context.Background()
	plaintext := []byte("test-oauth2-token")
	encryptionContext := map[string]string{
		"service_id": "oauth2",
		"principal":  "user@example.com",
	}

	ciphertext, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatal("ciphertext should not be empty")
	}

	decrypted, err := adapter.Decrypt(ctx, ciphertext, encryptionContext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted plaintext does not match: expected %q, got %q", string(plaintext), string(decrypted))
	}
}

// TestAdapterInterfaceImplementation tests that AWSAdapter implements EncryptionPort
func TestAdapterInterfaceImplementation(t *testing.T) {
	// Setup
	testKEK := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
	os.Setenv("INTERFACE_KEK", testKEK)
	defer os.Unsetenv("INTERFACE_KEK")

	adapter, err := NewAWSEncryptionAdapter("${INTERFACE_KEK}")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	// Verify adapter was created and has required fields
	if adapter == nil {
		t.Fatal("adapter should not be nil")
	}

	if adapter.keyring == nil {
		t.Error("adapter keyring should not be nil")
	}

	if adapter.encryptionClient == nil {
		t.Error("adapter encryptionClient should not be nil")
	}

	// Test that the methods work
	ctx := context.Background()
	plaintext := []byte("test-data")
	encryptionContext := map[string]string{"key": "value"}

	// Should be able to call Encrypt
	ciphertext, err := adapter.Encrypt(ctx, plaintext, encryptionContext)
	if err != nil {
		t.Fatalf("Encrypt method failed: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Error("Encrypt should produce non-empty ciphertext")
	}

	// Should be able to call Decrypt
	decrypted, err := adapter.Decrypt(ctx, ciphertext, encryptionContext)
	if err != nil {
		t.Fatalf("Decrypt method failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypt should return original plaintext: expected %q, got %q", string(plaintext), string(decrypted))
	}
}
