# AWS Encryption SDK Go - API Reference

This file contains detailed API documentation for the AWS Encryption SDK for Go.

## Go Modules

```
github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk
github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl
```

**API Reference:** https://pkg.go.dev/github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk

## Import Paths

```go
import (
    // Encryption SDK - Core client and types
    client "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygenerated"
    esdktypes "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygeneratedtypes"

    // Material Providers Library - Keyrings and CMMs
    mpl "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygenerated"
    mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"

    // KeyStore - For Hierarchical Keyring with DynamoDB
    keystore "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygenerated"
    keystoretypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygeneratedtypes"
)
```

## Core Types

### Encryption SDK Client

```go
// Create client with commitment policy
encryptionClient, err := client.NewClient(esdktypes.AwsEncryptionSdkConfig{
    CommitmentPolicy: mpltypes.ESDKCommitmentPolicyRequireEncryptRequireDecrypt.Enum(),
})
```

### Encrypt/Decrypt Input Types

```go
// EncryptInput
esdktypes.EncryptInput{
    Plaintext:         []byte,
    Keyring:           mpltypes.IKeyring,
    EncryptionContext: map[string]string,
    // Optional: AlgorithmSuiteId for custom algorithm suite
}

// DecryptInput
esdktypes.DecryptInput{
    Ciphertext:        []byte,
    Keyring:           mpltypes.IKeyring,
    EncryptionContext: map[string]string,  // For verification
}
```

### Commitment Policies

| Policy | Description | Recommendation |
|--------|-------------|----------------|
| `RequireEncryptRequireDecrypt` | Both encrypt and decrypt require commitment | **Use this (default)** |
| `RequireEncryptAllowDecrypt` | Encrypt requires, decrypt allows without | Migration only |
| `ForbidEncryptAllowDecrypt` | No commitment on encrypt, allows on decrypt | **Never use** |

### Algorithm Suites

Default algorithm suite (with signing and commitment):
- `ALG_AES_256_GCM_HKDF_SHA512_COMMIT_KEY_ECDSA_P384`

Without signing (only if security requirements explicitly allow):
- `ALG_AES_256_GCM_HKDF_SHA512_COMMIT_KEY`

```go
// Specify algorithm suite (optional - default is recommended)
encryptOutput, err := encryptionClient.Encrypt(ctx, esdktypes.EncryptInput{
    Plaintext:         plaintext,
    Keyring:           keyring,
    EncryptionContext: encryptionContext,
    AlgorithmSuiteId:  mpltypes.ESDKAlgorithmSuiteIdAlgAes256GcmHkdfSha512CommitKeyEcdsaP384.Enum(),
})
```

## Keyring Types

### AWS KMS Keyring

```go
keyring, err := mplClient.CreateAwsKmsKeyring(mpltypes.CreateAwsKmsKeyringInput{
    KmsClient: kmsClient,
    KmsKeyId:  kmsKeyArn,  // Always specify explicitly
})
```

### AWS KMS Hierarchical Keyring

```go
keyring, err := mplClient.CreateAwsKmsHierarchicalKeyring(mpltypes.CreateAwsKmsHierarchicalKeyringInput{
    KeyStore:            keyStoreClient,
    BranchKeyId:         &branchKeyID,        // Static branch key
    // OR
    BranchKeyIdSupplier: supplier,            // Dynamic branch key selection
    TtlSeconds:          cacheLimitTTL,       // Cache duration
    MaxCacheSize:        &maxCacheSize,       // Optional: limit cache entries
})
```

### Raw AES Keyring (Development/Testing Only)

```go
keyring, err := mplClient.CreateRawAesKeyring(mpltypes.CreateRawAesKeyringInput{
    KeyNamespace: "development",
    KeyName:      "test-key",
    WrappingKey:  keyBytes,  // 32 bytes for AES-256
    WrappingAlg:  mpltypes.AesWrappingAlgAlgAes256GcmIv12Tag16,
})
```

## KeyStore Types

### KeyStore Configuration

```go
keyStoreClient, err := keystore.NewClient(keystoretypes.KeyStoreConfig{
    DdbTableName:        tableName,           // DynamoDB table for branch keys
    KmsConfiguration:    kmsConfig,           // KMS key for protecting branch keys
    LogicalKeyStoreName: logicalName,         // Namespace identifier
    DdbClient:           ddbClient,           // AWS DynamoDB client
    KmsClient:           kmsClient,           // AWS KMS client
})
```

### KMS Configuration Options

```go
// Single KMS key
kmsConfig := &keystoretypes.KMSConfigurationMemberkmsKeyArn{
    Value: kmsKeyArn,
}

// Discovery (not recommended - prefer explicit key specification)
kmsConfig := &keystoretypes.KMSConfigurationMemberdiscovery{
    Value: keystoretypes.Discovery{},
}
```

### Branch Key Operations

```go
// Create branch key
output, err := keyStoreClient.CreateKey(keystoretypes.CreateKeyInput{
    BranchKeyIdentifier: &branchKeyID,
})

// Get active branch key
output, err := keyStoreClient.GetActiveBranchKey(keystoretypes.GetActiveBranchKeyInput{
    BranchKeyIdentifier: branchKeyID,
})

// Version/rotate branch key
output, err := keyStoreClient.VersionKey(keystoretypes.VersionKeyInput{
    BranchKeyIdentifier: branchKeyID,
})
```

## Branch Key Supplier Interface

```go
type IBranchKeyIdSupplier interface {
    GetBranchKeyId(input mpltypes.GetBranchKeyIdInput) (*mpltypes.GetBranchKeyIdOutput, error)
}

// Input contains the encryption context
type GetBranchKeyIdInput struct {
    EncryptionContext map[string]string
}

// Output specifies which branch key to use
type GetBranchKeyIdOutput struct {
    BranchKeyId string
}
```

## Error Handling

### Domain Error Types

```go
type EncryptionErrorKind string

const (
    // Encryption operation failed (key unavailable, permissions, etc.)
    ErrorKindEncryptionFailed   EncryptionErrorKind = "encryption_failed"
    
    // Decryption operation failed (corrupted data, wrong key, etc.)
    ErrorKindDecryptionFailed   EncryptionErrorKind = "decryption_failed"
    
    // Encryption context mismatch during decryption verification
    ErrorKindContextMismatch    EncryptionErrorKind = "context_mismatch"
    
    // Ciphertext integrity check failed (tampering detected)
    ErrorKindIntegrityViolation EncryptionErrorKind = "integrity_violation"
    
    // KMS key or branch key unavailable
    ErrorKindKEKUnavailable     EncryptionErrorKind = "kek_unavailable"
)
```

### Error Wrapping Pattern

```go
func (a *Adapter) Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error) {
    output, err := a.client.Encrypt(ctx, esdktypes.EncryptInput{...})
    if err != nil {
        return nil, &EncryptionError{
            Kind:    ErrorKindEncryptionFailed,
            Message: "failed to encrypt data",
            Cause:   err,
        }
    }
    return output.Ciphertext, nil
}
```

## Message Format

The encrypted message contains:
- **Header**: Algorithm suite, encryption context (plaintext), encrypted data keys
- **Body**: Encrypted content in frames
- **Footer**: Authentication tag, optional signature

Key commitment adds ~30 bytes to ciphertext size.

## Encrypted Data Key Limits

The SDK supports up to 65,535 encrypted data keys per message. When decrypting messages from untrusted sources, be aware that malicious actors could craft messages with many encrypted data keys to cause resource exhaustion or increased KMS costs.

The Go SDK does not currently expose a configuration to limit encrypted data keys during decryption. Mitigate by validating message sources before decryption.
