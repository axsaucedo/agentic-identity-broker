# Encryption Integration Guide

This guide provides developers with comprehensive instructions for integrating the Encryption Vault feature into applications using the Agentic Identity Broker.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Service Integration](#service-integration)
- [Error Handling](#error-handling)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)

## Overview

The Encryption Vault feature provides transparent envelope encryption of OAuth2 tokens at rest. Tokens are encrypted using:

1. **DEK (Data Encryption Key)**: A fresh 256-bit AES key generated for each token
2. **KEK (Key Encryption Key)**: A master key managed by AWS KMS or environment variable
3. **Context Binding**: Service-level isolation preventing token reuse across services

Encryption/decryption happens transparently in the OAuth2SessionService layer - tokens are encrypted when stored and decrypted when retrieved. Storage adapters see only encrypted ciphertext.

## Quick Start

### 1. Configure Encryption

**For Production (AWS KMS):**

```yaml
# config.yaml
encryption:
  aws_kms:
    key_arn: ${IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN}
```

**For Development (Memory backend):**

`.env.local` files do not evaluate shell command substitution. Paste a generated base64 key when editing the file directly.

```dotenv
# .env.local
IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY=base64-encoded-32-byte-key
```

```bash
# shell
export IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY="$(openssl rand -base64 32)"
```

```yaml
# config.yaml
encryption:
  memory:
    raw_key: ${IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY}
```

### 2. Initialize Application

The application builder automatically initializes encryption from the configured
backend:

```go
// From internal/app/builder.go - automatic initialization
encryptor, branchKeyManager, err := awsencryption.NewEncryptionAdapter(&config.Encryption)
if err != nil {
    return nil, fmt.Errorf("failed to initialize encryption adapter: %w", err)
}
```

### 3. Use OAuth2SessionService

The encryption is transparent in the service layer:

```go
// Creating a session (tokens automatically encrypted by service)
session, err := oauth2Service.CreateSessionFromCallback(
    ctx,
    principal,
    &HandleCallbackRequest{
        ServiceID: "github",
        Code:      authCode,
        State:     stateToken,
    },
)
// ✅ Service encrypts tokens before storing in database
// ✅ Database receives encrypted ciphertext only

// Retrieving a session (tokens automatically decrypted by service)
session, err := oauth2Service.GetSessionWithAgents(ctx, principal, "github")
// ✅ Service decrypts tokens after retrieving from database
// ✅ Application receives plaintext tokens
```

## Configuration

### AWS KMS Setup (Production)

#### Step 1: Create KMS Key

```bash
# Create customer-managed key
aws kms create-key \
  --description "Identity Broker OAuth Token Encryption Key" \
  --region eu-central-1
```

Output example:

```
KeyId: arn:aws:kms:eu-central-1:123456789012:key/12345678-1234-1234-1234-123456789012
```

#### Step 2: Create Alias

```bash
aws kms create-alias \
  --alias-name "alias/identity-broker-encryption" \
  --target-key-id "arn:aws:kms:eu-central-1:123456789012:key/12345678-1234-1234-1234-123456789012"
```

#### Step 3: Grant Service Permissions

```bash
aws kms create-grant \
  --key-id "arn:aws:kms:eu-central-1:123456789012:key/12345678-1234-1234-1234-123456789012" \
  --grantee-principal "arn:aws:iam::123456789012:role/IdentityBrokerRole" \
  --operations "Encrypt" "Decrypt" "GenerateDataKey" "DescribeKey"
```

#### Step 4: Configure Identity Broker

```yaml
# config.yaml
encryption:
  aws_kms:
    key_arn: "arn:aws:kms:eu-central-1:123456789012:alias/identity-broker-encryption"
    dynamodb_table_name: "IdentityBrokerEncryptionBranchKeys"
    branch_key_ttl: "1h"
```

### Environment Variable Setup (Development)

#### Step 1: Generate Key

```bash
# Using OpenSSL
IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY=$(openssl rand -base64 32)

# Using Python
IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY=$(python3 -c "import os, base64; print(base64.b64encode(os.urandom(32)).decode())")

# Using Go
IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY=$(go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"; "os"); func main() { key := make([]byte, 32); rand.Read(key); fmt.Println(base64.StdEncoding.EncodeToString(key)) }')
```

#### Step 2: Configure Identity Broker

```dotenv
# .env.local
IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY=base64-encoded-32-byte-key
```

```bash
# shell
export IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY="$(openssl rand -base64 32)"
```

```yaml
# config.yaml
encryption:
  memory:
    raw_key: ${IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY}
```

## Service Integration

### OAuth2SessionService Integration

The OAuth2SessionService automatically encrypts/decrypts tokens:

```go
// Service method that automatically encrypts tokens
func (s *OAuth2SessionService) CreateSessionFromCallback(
    ctx context.Context,
    principal string,
    req *HandleCallbackRequest,
) (*storage.UserSession, error) {
    // ... exchange code for tokens ...

    // Encryption context for this service
    encryptionContext := map[string]string{"service_id": serviceID}

    // Encrypt access token (transparent to caller)
    encryptedAccess, err := s.encryption.Encrypt(ctx, []byte(token.AccessToken), encryptionContext)

    // Encrypt refresh token
    encryptedRefresh, err := s.encryption.Encrypt(ctx, []byte(token.RefreshToken), encryptionContext)

    // Create session with encrypted tokens
    session := &storage.UserSession{
        EncryptedAccessToken: encryptedAccess,
        EncryptedRefreshToken: encryptedRefresh,
        // ... other fields ...
    }

    // Store to database - receives encrypted ciphertext
    return s.sessionRepo.Create(ctx, session)
}

// Service method that automatically decrypts tokens
func (s *OAuth2SessionService) GetSessionWithAgents(
    ctx context.Context,
    principal string,
    serviceID string,
) (*SessionWithAgents, error) {
    // Retrieve session from database
    session, err := s.sessionRepo.FindByPrincipalAndService(ctx, principal, serviceID)

    // Decryption context must match encryption context
    encryptionContext := session.EncryptionContext

    // Decrypt access token (transparent to caller)
    plainAccessToken, err := s.encryption.Decrypt(ctx, session.EncryptedAccessToken, encryptionContext)

    // Decrypt refresh token
    plainRefreshToken, err := s.encryption.Decrypt(ctx, session.EncryptedRefreshToken, encryptionContext)

    // Return session with plaintext tokens
    session.AccessToken = string(plainAccessToken)
    session.RefreshToken = string(plainRefreshToken)

    return &SessionWithAgents{Session: session, DependentAgents: agents}, nil
}
```

### Using EncryptionPort Interface Directly

For custom encryption needs:

```go
// EncryptionPort interface
type EncryptionPort interface {
    Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error)
    Decrypt(ctx context.Context, ciphertext []byte, encryptionContext map[string]string) ([]byte, error)
}

// Access from app builder
encryptionPort := app.EncryptionPort

// Encrypt a token
encryptionContext := map[string]string{"service_id": "github"}
ciphertext, err := encryptionPort.Encrypt(ctx, []byte(token), encryptionContext)
if err != nil {
    // Handle encryption error
}

// Decrypt a token
plaintext, err := encryptionPort.Decrypt(ctx, ciphertext, encryptionContext)
if err != nil {
    // Handle decryption error
}
```

## Error Handling

### Encryption Errors

The encryption port returns domain errors with error kinds:

```go
import "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/encryption"

_, err := encryptionPort.Encrypt(ctx, plaintext, encryptionContext)
if err != nil {
    // Check error kind
    switch err := err.(type) {
    case *encryption.EncryptionError:
        switch err.Kind {
        case encryption.ErrorKindContextMismatch:
            // Context verification failed - token not for this service
            log.Error("context mismatch", err.Message)
        case encryption.ErrorKindIntegrityViolation:
            // Token has been tampered with
            log.Error("integrity violation", err.Message)
        case encryption.ErrorKindKEKUnavailable:
            // AWS KMS or environment variable not accessible
            log.Error("kek unavailable", err.Message)
        case encryption.ErrorKindEncryptionFailed:
            // General encryption failure
            log.Error("encryption failed", err.Message)
        }
    }
}
```

### Common Scenarios

#### AWS KMS Key Not Found

**Error**: `kek_unavailable: KMS key not accessible`

**Solution**:

1. Verify KMS key ARN is correct in configuration
2. Verify IAM role has `kms:DescribeKey` permission
3. Check KMS key exists in the specified region

#### Environment Variable Not Set

**Error**: `kek_unavailable: encryption key material is empty`

**Solution**:

```bash
# Generate and export the key
export IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY=$(openssl rand -base64 32)

# Verify it's set
echo $IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY
```

#### Context Mismatch on Decryption

**Error**: `context_mismatch: context verification failed during decryption`

**Solution**:

- Ensure decryption context matches encryption context
- Tokens encrypted for service A cannot be decrypted with service B context
- Check service_id in encryption context

## Testing

### Unit Tests

```go
func TestEncryption(t *testing.T) {
    // Create mock encryption port
    mockEncryption := &MockEncryptionPort{
        encryptFn: func(ctx context.Context, plaintext []byte, ctx map[string]string) ([]byte, error) {
            // Return deterministic ciphertext for testing
            return append([]byte("encrypted_"), plaintext...), nil
        },
        decryptFn: func(ctx context.Context, ciphertext []byte, ctx map[string]string) ([]byte, error) {
            // Extract plaintext from test ciphertext
            return ciphertext[10:], nil
        },
    }

    // Use in service for testing
    service := oauth2session.NewOAuth2SessionService(
        serviceRepo, sessionRepo, grantRepo, agentRepo,
        mockEncryption,
        jweKey, config, logger,
    )

    // Test token encryption
    session, err := service.CreateSessionFromCallback(ctx, principal, cbReq)
    require.NoError(t, err)
    require.NotNil(t, session.EncryptedAccessToken)
}
```

### Integration Tests

```go
func TestEncryptionWithRealAdapter(t *testing.T) {
    // Use a real AWS adapter with a LocalStack-compatible AWS emulator for testing
    adapter, manager, err := aws.NewAWSEncryption(
        "arn:aws:kms:eu-central-1:123456789012:key/12345678",
        "IdentityBrokerEncryptionBranchKeys",
        1*time.Hour,
    )
    require.NoError(t, err)

    // Test encryption/decryption roundtrip
    plaintext := []byte("test-token")
    context := map[string]string{"service_id": "github"}

    ciphertext, err := adapter.Encrypt(ctx, plaintext, context)
    require.NoError(t, err)
    require.NotEqual(t, plaintext, ciphertext)

    decrypted, err := adapter.Decrypt(ctx, ciphertext, context)
    require.NoError(t, err)
    require.Equal(t, plaintext, decrypted)
}
```

### E2E Tests

See `tests/e2e/encryption_vault_raw_test.go` for comprehensive E2E test scenarios covering:

- Envelope encryption with context binding
- AWS KMS storage
- Environment variable KEK injection
- DEK generation and isolation
- Context verification failure scenarios
- Cross-service token reuse prevention

## Troubleshooting

### Issue: "Adapter not properly initialized"

**Cause**: Encryption adapter failed to initialize at startup

**Solution**:

1. Check KEK configuration is valid (AWS KMS ARN or environment variable)
2. Verify AWS credentials are available
3. Verify environment variable is set if using env var mode
4. Check application logs for detailed error

### Issue: "KMS key not accessible"

**Cause**: AWS KMS key cannot be accessed

**Solution**:

1. Verify KMS key exists in the specified region
2. Verify IAM role has required permissions
3. Check AWS region configuration
4. Verify key is not scheduled for deletion

### Issue: "Context verification failed"

**Cause**: Token was encrypted with different context

**Solution**:

1. Ensure encryption and decryption use the same service_id
2. Check if token is being used for a different service
3. Verify encryption context is built correctly

### Issue: Performance degradation

**Cause**: AWS KMS latency or DynamoDB caching issues

**Solution**:

1. Check AWS KMS CloudTrail logs for throttling
2. Increase branch key TTL if cache eviction is frequent
3. Consider DynamoDB provisioned capacity
4. Monitor network latency to AWS services

### Issue: Tokens cannot be decrypted after restart

**Cause**: Different KEK being used after restart

**Solution**:

1. Verify KEK configuration is identical
2. Verify environment variable hasn't changed
3. Check AWS KMS key hasn't been rotated to incompatible version
4. Ensure database contains encrypted tokens from previous run

## Additional Resources

- Encryption feature specification: see `specs/012-aws-encryption-vault/spec.md` in the repository
- Architecture documentation: see `ARCHITECTURE.md` in the repository root
- [Configuration Guide](./configuration.md)
- [AWS KMS Documentation](https://docs.aws.amazon.com/kms/)
- [AWS Encryption SDK Documentation](https://docs.aws.amazon.com/encryption-sdk/)
