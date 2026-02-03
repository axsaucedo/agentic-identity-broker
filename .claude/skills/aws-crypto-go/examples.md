# AWS Encryption SDK Go - Code Examples

Complete, production-ready code examples for the AWS Encryption SDK for Go.

## Basic Encryption/Decryption

```go
package encryption

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/kms"

    client "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygenerated"
    esdktypes "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygeneratedtypes"
    mpl "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygenerated"
    mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
)

type Encryptor struct {
    client  *client.Client
    keyring mpltypes.IKeyring
}

func NewEncryptor(ctx context.Context, kmsKeyArn string) (*Encryptor, error) {
    // 1. Create AWS clients
    awsCfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    kmsClient := kms.NewFromConfig(awsCfg)

    // 2. Create Encryption SDK client with commitment policy
    encryptionClient, err := client.NewClient(esdktypes.AwsEncryptionSdkConfig{
        CommitmentPolicy: mpltypes.ESDKCommitmentPolicyRequireEncryptRequireDecrypt.Enum(),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create encryption client: %w", err)
    }

    // 3. Create Material Providers client
    mplClient, err := mpl.NewClient(mpltypes.MaterialProvidersConfig{})
    if err != nil {
        return nil, fmt.Errorf("failed to create MPL client: %w", err)
    }

    // 4. Create KMS keyring with explicit key specification
    keyring, err := mplClient.CreateAwsKmsKeyring(mpltypes.CreateAwsKmsKeyringInput{
        KmsClient: kmsClient,
        KmsKeyId:  kmsKeyArn,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create keyring: %w", err)
    }

    return &Encryptor{
        client:  encryptionClient,
        keyring: keyring,
    }, nil
}

func (e *Encryptor) Encrypt(ctx context.Context, plaintext []byte, resourceID string) ([]byte, error) {
    // Always use meaningful encryption context
    encryptionContext := map[string]string{
        "resource_id": resourceID,
        "purpose":     "data_protection",
    }

    output, err := e.client.Encrypt(ctx, esdktypes.EncryptInput{
        Plaintext:         plaintext,
        Keyring:           e.keyring,
        EncryptionContext: encryptionContext,
    })
    if err != nil {
        return nil, fmt.Errorf("encryption failed: %w", err)
    }

    return output.Ciphertext, nil
}

func (e *Encryptor) Decrypt(ctx context.Context, ciphertext []byte, resourceID string) ([]byte, error) {
    // Provide expected encryption context for verification
    expectedContext := map[string]string{
        "resource_id": resourceID,
        "purpose":     "data_protection",
    }

    output, err := e.client.Decrypt(ctx, esdktypes.DecryptInput{
        Ciphertext:        ciphertext,
        Keyring:           e.keyring,
        EncryptionContext: expectedContext,
    })
    if err != nil {
        return nil, fmt.Errorf("decryption failed: %w", err)
    }

    return output.Plaintext, nil
}
```

## Hierarchical Keyring Setup

```go
package encryption

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/kms"

    client "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygenerated"
    esdktypes "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygeneratedtypes"
    keystore "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygenerated"
    keystoretypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygeneratedtypes"
    mpl "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygenerated"
    mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
)

type HierarchicalEncryptor struct {
    client       *client.Client
    keyring      mpltypes.IKeyring
    keyStore     *keystore.Client
}

type HierarchicalConfig struct {
    KMSKeyArn        string
    DynamoDBTable    string
    BranchKeyID      string
    CacheTTLSeconds  int64
}

func NewHierarchicalEncryptor(ctx context.Context, cfg HierarchicalConfig) (*HierarchicalEncryptor, error) {
    // 1. Create AWS clients
    awsCfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    kmsClient := kms.NewFromConfig(awsCfg)
    ddbClient := dynamodb.NewFromConfig(awsCfg)

    // 2. Create Encryption SDK client with commitment
    encryptionClient, err := client.NewClient(esdktypes.AwsEncryptionSdkConfig{
        CommitmentPolicy: mpltypes.ESDKCommitmentPolicyRequireEncryptRequireDecrypt.Enum(),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create encryption client: %w", err)
    }

    // 3. Create KeyStore backed by DynamoDB
    // Prerequisite: Table must exist with Partition Key "partition_key" (S) and Sort Key "sort_key" (S)
    keyStoreClient, err := keystore.NewClient(keystoretypes.KeyStoreConfig{
        DdbTableName:        cfg.DynamoDBTable,
        KmsConfiguration:    &keystoretypes.KMSConfigurationMemberkmsKeyArn{Value: cfg.KMSKeyArn},
        // WARNING: Changing LogicalKeyStoreName later will render existing branch keys unusable
        LogicalKeyStoreName: cfg.DynamoDBTable,
        DdbClient:           ddbClient,
        KmsClient:           kmsClient,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create keystore client: %w", err)
    }

    // 4. Create Material Providers client
    mplClient, err := mpl.NewClient(mpltypes.MaterialProvidersConfig{})
    if err != nil {
        return nil, fmt.Errorf("failed to create MPL client: %w", err)
    }

    // 5. Create Hierarchical Keyring with caching
    keyring, err := mplClient.CreateAwsKmsHierarchicalKeyring(mpltypes.CreateAwsKmsHierarchicalKeyringInput{
        KeyStore:    keyStoreClient,
        BranchKeyId: &cfg.BranchKeyID,
        TtlSeconds:  cfg.CacheTTLSeconds,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create hierarchical keyring: %w", err)
    }

    return &HierarchicalEncryptor{
        client:   encryptionClient,
        keyring:  keyring,
        keyStore: keyStoreClient,
    }, nil
}

// CreateBranchKey provisions a new branch key (run once per service/tenant)
func (e *HierarchicalEncryptor) CreateBranchKey(ctx context.Context, identifier string) (string, error) {
    output, err := e.keyStore.CreateKey(ctx, keystoretypes.CreateKeyInput{
        BranchKeyIdentifier: &identifier,
    })
    if err != nil {
        return "", fmt.Errorf("failed to create branch key: %w", err)
    }
    return output.BranchKeyIdentifier, nil
}

// RotateBranchKey creates a new version of an existing branch key
func (e *HierarchicalEncryptor) RotateBranchKey(ctx context.Context, identifier string) error {
    _, err := e.keyStore.VersionKey(ctx, keystoretypes.VersionKeyInput{
        BranchKeyIdentifier: identifier,
    })
    if err != nil {
        return fmt.Errorf("failed to rotate branch key: %w", err)
    }
    return nil
}
```

## Multi-Tenant Branch Key Supplier

```go
package encryption

import (
    "fmt"
    "sync"

    mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
)

// TenantBranchKeySupplier routes encryption to tenant-specific branch keys
type TenantBranchKeySupplier struct {
    mu               sync.RWMutex
    branchKeyMapping map[string]string // tenantID -> branchKeyID
}

func NewTenantBranchKeySupplier() *TenantBranchKeySupplier {
    return &TenantBranchKeySupplier{
        branchKeyMapping: make(map[string]string),
    }
}

// RegisterTenant maps a tenant to their branch key
func (s *TenantBranchKeySupplier) RegisterTenant(tenantID, branchKeyID string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.branchKeyMapping[tenantID] = branchKeyID
}

// GetBranchKeyId implements mpltypes.IBranchKeyIdSupplier
func (s *TenantBranchKeySupplier) GetBranchKeyId(
    input mpltypes.GetBranchKeyIdInput,
) (*mpltypes.GetBranchKeyIdOutput, error) {
    // Extract tenant from encryption context
    tenantID, exists := input.EncryptionContext["tenant_id"]
    if !exists || tenantID == "" {
        return nil, fmt.Errorf("encryption context must contain 'tenant_id'")
    }

    s.mu.RLock()
    branchKeyID, found := s.branchKeyMapping[tenantID]
    s.mu.RUnlock()

    if !found {
        return nil, fmt.Errorf("no branch key configured for tenant: %s", tenantID)
    }

    return &mpltypes.GetBranchKeyIdOutput{
        BranchKeyId: branchKeyID,
    }, nil
}

// Usage with hierarchical keyring
func createMultiTenantKeyring(
    mplClient *mpl.Client,
    keyStoreClient *keystore.Client,
    supplier *TenantBranchKeySupplier,
    cacheTTL int64,
) (mpltypes.IKeyring, error) {
    return mplClient.CreateAwsKmsHierarchicalKeyring(mpltypes.CreateAwsKmsHierarchicalKeyringInput{
        KeyStore:            keyStoreClient,
        BranchKeyIdSupplier: supplier,
        TtlSeconds:          cacheTTL,
    })
}
```

## Service-Based Branch Key Supplier

```go
package encryption

import (
    "fmt"

    mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
)

// ServiceBranchKeySupplier derives branch key ID from service_id in encryption context
type ServiceBranchKeySupplier struct {
    branchKeyPrefix string // e.g., "service_"
    branchKeySuffix string // e.g., "_branch_key"
}

func NewServiceBranchKeySupplier(prefix, suffix string) *ServiceBranchKeySupplier {
    return &ServiceBranchKeySupplier{
        branchKeyPrefix: prefix,
        branchKeySuffix: suffix,
    }
}

// GetBranchKeyId implements mpltypes.IBranchKeyIdSupplier
func (s *ServiceBranchKeySupplier) GetBranchKeyId(
    input mpltypes.GetBranchKeyIdInput,
) (*mpltypes.GetBranchKeyIdOutput, error) {
    serviceID, exists := input.EncryptionContext["service_id"]
    if !exists || serviceID == "" {
        return nil, fmt.Errorf("encryption context must contain 'service_id'")
    }

    // Construct branch key ID: service_{service_id}_branch_key
    branchKeyID := fmt.Sprintf("%s%s%s", s.branchKeyPrefix, serviceID, s.branchKeySuffix)

    return &mpltypes.GetBranchKeyIdOutput{
        BranchKeyId: branchKeyID,
    }, nil
}
```

## Development Mode (Raw AES Keyring)

**⚠️ For development and testing only - never use in production**

```go
package encryption

import (
    "context"
    "encoding/base64"
    "fmt"

    client "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygenerated"
    esdktypes "github.com/aws/aws-encryption-sdk/releases/go/encryption-sdk/awscryptographyencryptionsdksmithygeneratedtypes"
    mpl "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygenerated"
    mpltypes "github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographymaterialproviderssmithygeneratedtypes"
)

func NewDevEncryptor(base64Key string) (*Encryptor, error) {
    // Decode base64 key (must be 32 bytes for AES-256)
    keyBytes, err := base64.StdEncoding.DecodeString(base64Key)
    if err != nil {
        return nil, fmt.Errorf("invalid base64 key: %w", err)
    }
    if len(keyBytes) != 32 {
        return nil, fmt.Errorf("key must be 32 bytes for AES-256, got %d", len(keyBytes))
    }

    // Create Encryption SDK client
    encryptionClient, err := client.NewClient(esdktypes.AwsEncryptionSdkConfig{
        CommitmentPolicy: mpltypes.ESDKCommitmentPolicyRequireEncryptRequireDecrypt.Enum(),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create encryption client: %w", err)
    }

    // Create Material Providers client
    mplClient, err := mpl.NewClient(mpltypes.MaterialProvidersConfig{})
    if err != nil {
        return nil, fmt.Errorf("failed to create MPL client: %w", err)
    }

    // Create Raw AES keyring (development only)
    keyring, err := mplClient.CreateRawAesKeyring(mpltypes.CreateRawAesKeyringInput{
        KeyNamespace: "development",
        KeyName:      "local-dev-key",
        WrappingKey:  keyBytes,
        WrappingAlg:  mpltypes.AesWrappingAlgAlgAes256GcmIv12Tag16,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create raw AES keyring: %w", err)
    }

    return &Encryptor{
        client:  encryptionClient,
        keyring: keyring,
    }, nil
}
```

## Encryption Context Verification

The SDK **automatically verifies** encryption context when you provide it in `DecryptInput`:

```go
// The SDK verifies that ALL provided key-value pairs match the ciphertext header.
// If any mismatch occurs, Decrypt() returns an error - no manual check needed.
output, err := e.client.Decrypt(ctx, esdktypes.DecryptInput{
    Ciphertext:        ciphertext,
    Keyring:           e.keyring,
    EncryptionContext: expectedContext,  // SDK verifies this automatically
})
if err != nil {
    // Includes context mismatch errors
    return nil, fmt.Errorf("decryption failed: %w", err)
}
```

## Official AWS Examples

For additional patterns and edge cases, see the official AWS examples:

- [Hierarchical Keyring](https://github.com/aws/aws-encryption-sdk/blob/mainline/releases/go/encryption-sdk/examples/keyring/awskmshierarchicalkeyring/awskmshierarchicalkeyring.go)
- [Branch Key Supplier](https://github.com/aws/aws-encryption-sdk/blob/mainline/releases/go/encryption-sdk/examples/keyring/awskmshierarchicalkeyring/branchkeysupplier.go)
