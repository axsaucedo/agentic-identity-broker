# ADR 009: Envelope Encryption Design for OAuth2 Token Vault

**Status**: Accepted
**Date**: 2026-01-16
**Deciders**: System Architecture Team

---

## Context

The Agentic Identity Broker requires secure storage of OAuth2 access and refresh tokens for third-party services (GitHub, Google, Microsoft, etc.). These tokens are sensitive credentials that must be protected both at rest and in transit. The system needs to support:

1. **High Security**: Protection against database compromise, memory dumps, and insider threats
2. **Performance**: Sub-100ms token encryption/decryption for session operations
3. **Scalability**: Efficient encryption for thousands of concurrent sessions
4. **Compliance**: Enterprise-grade key management and audit trails
5. **Context Isolation**: Cryptographic guarantees that tokens encrypted for one service cannot be decrypted for another

Previous analysis in [ADR 008: Encryption Context Optimization](008-encryption-context-optimization.md) established the performance and security benefits of service_id-only context binding.

---

## Decision

We will implement **envelope encryption with DEK-per-session** using **AWS KMS** as the primary key management solution, with **context binding** for service isolation.

### Core Architecture

1. **Data Encryption Key (DEK) per Session**
   - Each UserSession generates a unique AES-256 DEK
   - DEK encrypts all tokens (access_token, refresh_token) for that session
   - DEK is encrypted with the Key Encryption Key (KEK) and stored alongside ciphertext

2. **AWS KMS as KEK Provider**
   - Customer-managed KMS key (CMK) serves as the Key Encryption Key
   - DEK wrapped/unwrapped via AWS KMS GenerateDataKey/Decrypt APIs
   - Single KMS call per session lifecycle (not per token operation)

3. **Service-ID Context Binding**
   - EncryptionContext contains `{"service_id": "oauth2"}` as Additional Authenticated Data (AAD)
   - AESGCMSIV authenticated encryption prevents cross-service token usage
   - Cryptographic isolation between different third-party services

4. **Memguard Memory Protection**
   - DEKs and plaintext tokens stored in memguard-protected memory regions
   - Memory locked in RAM (no swap), wiped on deallocation
   - Protection against memory dumps and process inspection

---

## Rationale

### 1. DEK-per-Session Efficiency

**Chosen**: One DEK per UserSession, single KMS call per session lifecycle

**Alternative Considered**: Per-token KMS calls
**Rejected Because**: 2x KMS calls per token operation (encrypt access_token, encrypt refresh_token) creates 200ms+ latency and scales poorly with concurrent sessions

**Benefits**:
- **Performance**: Single KMS call amortized across session lifetime (~24 hours)
- **Cost Efficiency**: Fewer KMS API calls reduce operational costs
- **Scalability**: Constant-time encryption regardless of token count per session
- **Session Isolation**: Each session gets cryptographically independent encryption

### 2. Service-ID-Only Context Binding

**Chosen**: `{"service_id": "oauth2"}` as encryption context

**Alternative Considered**: Multi-field context with principal, session_id, timestamp
**Analysis in**: [ADR 008: Encryption Context Optimization](008-encryption-context-optimization.md)
**Rejected Because**: Benchmarking showed 40% performance overhead with minimal security benefit

**Benefits**:
- **Performance**: Minimal AAD processing overhead (~8ms vs 12ms for multi-field)
- **Service Isolation**: Tokens encrypted for GitHub cannot be decrypted with Google's context
- **Simplicity**: Single field reduces complexity while maintaining core security property
- **Storage Efficiency**: Smaller encryption context reduces metadata overhead

### 3. AWS Encryption SDK Selection

**Chosen**: AWS Encryption SDK with customer-managed KMS keys

**Alternative Considered**: Custom envelope encryption implementation
**Rejected Because**:
- Higher development and audit overhead
- Potential for cryptographic implementation bugs
- No integration with AWS CloudTrail for key usage auditing

**Alternative Considered**: Database-level encryption (TDE)
**Rejected Because**:
- No application-level context binding
- Cannot prevent cross-service token access within same database
- Limited key rotation and access control

**Benefits**:
- **Battle-Tested Cryptography**: Production-hardened implementation used by AWS services
- **Automatic Key Rotation**: Support for automatic CMK rotation without application changes
- **Audit Integration**: CloudTrail logging of all key usage for compliance
- **Multi-Region Support**: Cross-region key replication for disaster recovery

### 4. Memguard Integration Rationale

**Chosen**: memguard.LockedBuffer for DEKs and plaintext tokens

**Alternative Considered**: Standard Go byte slices
**Rejected Because**:
- Sensitive data could be swapped to disk
- Memory dumps expose plaintext tokens
- No automatic zeroing on garbage collection

**Benefits**:
- **Memory Isolation**: Sensitive data locked in non-swappable memory pages
- **Automatic Cleanup**: Memory regions zeroed when no longer needed
- **Process Protection**: Protection against memory inspection and core dumps
- **Minimal Overhead**: ~1% performance impact for significant security improvement

---

## Consequences

### Positive

1. **High Security Posture**
   - Envelope encryption provides defense-in-depth
   - Context binding prevents cross-service token misuse
   - Memory protection reduces local attack surface
   - AWS KMS integration provides enterprise-grade key management

2. **Performance Efficiency**
   - Sub-50ms encryption/decryption operations
   - Single KMS call per session amortizes key management overhead
   - AESGCMSIV provides fast authenticated encryption

3. **Operational Benefits**
   - CloudTrail audit logs for all key operations
   - Automatic key rotation without application downtime
   - Multi-region key replication for high availability
   - Integration with AWS IAM for fine-grained access control

4. **Development Velocity**
   - Well-documented AWS SDK reduces implementation complexity
   - Standardized error handling patterns from AWS SDK
   - Community support and extensive documentation

### Negative

1. **AWS Dependency**
   - Tight coupling to AWS ecosystem for key management
   - Requires AWS credentials and network connectivity
   - Regional outages could impact token encryption/decryption

2. **Additional Complexity**
   - Envelope encryption adds implementation complexity vs simple encryption
   - Error handling for KMS failures, network timeouts
   - Key rotation procedures require operational processes

3. **Cost Implications**
   - AWS KMS charges per API call (though mitigated by DEK-per-session)
   - CloudTrail logging has storage costs
   - Additional AWS services for audit and monitoring

### Mitigations

1. **AWS Dependency**: Environment variable KEK fallback for development/testing
2. **Complexity**: Comprehensive error handling in EncryptionPort interface
3. **Cost**: DEK-per-session pattern minimizes KMS API calls

---

## Implementation Notes

### Configuration Structure

```yaml
encryption:
  key_encryption_key: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
  # Alternative for development:
  # key_encryption_key: "${ENCRYPTION_KEK}"
```

### Error Handling

All encryption failures are categorized per [specs/012-aws-encryption-vault/contracts/error-contract.md](../specs/012-aws-encryption-vault/contracts/error-contract.md):

- `ErrorKindEncryptionFailed`: DEK generation or encryption failure
- `ErrorKindDecryptionFailed`: KMS unwrap or decryption failure
- `ErrorKindContextMismatch`: Service isolation violation
- `ErrorKindIntegrityViolation`: AESGCMSIV authentication failure
- `ErrorKindKEKUnavailable`: AWS KMS service unreachable

### Security Considerations

1. **Key Material Protection**: DEKs and tokens stored only in memguard-protected memory
2. **Context Validation**: All decryption operations validate service_id context match
3. **Fail-Closed**: Any encryption error results in operation failure, no fallback to plaintext
4. **Audit Logging**: All key operations logged to CloudTrail for compliance

---

## Related Documents

- [ADR 008: Encryption Context Optimization](008-encryption-context-optimization.md) - Context binding performance analysis
- [AWS Encryption Vault Specification](../specs/012-aws-encryption-vault/spec.md) - Functional requirements
- [Error Contract](../specs/012-aws-encryption-vault/contracts/error-contract.md) - Error handling patterns
- [AWS KMS Developer Guide](https://docs.aws.amazon.com/kms/latest/developerguide/) - Implementation reference