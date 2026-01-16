# AWS KMS Architecture Decision: Hierarchical Keyring with Branch Key Caching

**Date**: 2026-01-16 | **Status**: Design Decision | **Scope**: Encryption Vault for OAuth Tokens

---

## Context

When using AWS KMS for OAuth token encryption, there are two architectural approaches:

1. **Hierarchical Keyring with Branch Key Caching** (Recommended)
2. **Direct KMS Calls** (Alternative, Higher Security)

This document clarifies the tradeoffs and recommends an approach for the encryption vault feature.

---

## Architectural Overview

### Three-Layer Key Hierarchy

```
┌──────────────────────────────────────────────────────────┐
│                   AWS KMS (Cloud)                         │
│                                                            │
│  KEK (Key Encryption Key)                                │
│  - Stored in AWS KMS, never in plaintext in application  │
│  - Managed by AWS (audit logging, key rotation)          │
│  - Accessed via AWS API calls                            │
└────────────────────┬─────────────────────────────────────┘
                     │
                     │ Called 1x per TTL window
                     │
┌────────────────────▼──────────────────────────────────────┐
│         Application Process Memory (Linux)                │
│                                                            │
│  Branch Key (Cached)                                     │
│  - Intermediate key derived from KEK                      │
│  - Cached locally for performance (TTL: 15 minutes)      │
│  - Protected by memguard (MLOCK, DONTDUMP)              │
│  - One per unique service_id (encryption context)        │
│  - ~200 bytes per key                                     │
└────────────────────┬──────────────────────────────────────┘
                     │
                     │ Used N times per TTL
                     │ (wraps/unwraps DEKs)
                     │
┌────────────────────▼──────────────────────────────────────┐
│     Session Data (PostgreSQL BYTEA Columns)              │
│                                                            │
│  DEK (Data Encryption Key)                               │
│  - Derived from Branch Key per session                    │
│  - Encrypts OAuth tokens with context binding            │
│  - Single-use (wrapped + discarded after use)            │
│  - Never stored in plaintext                              │
└────────────────────────────────────────────────────────────┘
```

---

## Approach 1: Hierarchical Keyring with Branch Key Caching (RECOMMENDED)

### How It Works

1. **Initialization**:
   - Adapter receives AWS KMS ARN from configuration
   - Initializes AWS Encryption SDK with Hierarchical Keyring
   - First encryption request for a service_id triggers KMS call

2. **First Encryption (per service_id)**:
   - Application calls AWS KMS: "Generate a Branch Key for service_id=oauth2"
   - AWS KMS returns Branch Key (32 bytes)
   - Branch Key cached locally with 15-minute TTL
   - Branch Key wrapped in memguard enclave (protected memory)

3. **Subsequent Encryptions (during TTL)**:
   - No KMS call needed
   - Use cached Branch Key to wrap/unwrap DEKs
   - Fast local operations (~1-5ms)

4. **After TTL Expiration**:
   - Cache invalidated
   - Next operation triggers fresh KMS call for new Branch Key
   - Old tokens remain decryptable (KMS maintains version history)

### Performance Characteristics

| Metric | Value |
|--------|-------|
| **First call per service_id** | 50-200ms (KMS roundtrip) |
| **Cached calls** | ~1-5ms (local cache hit) |
| **Memory per Branch Key** | ~200 bytes |
| **Memory overhead (10 services)** | ~2KB |
| **Memory overhead (100 services)** | ~20KB |
| **KMS API calls (1 hour TTL)** | 1 per service_id per hour |
| **AWS KMS cost** | ~$0.03 per 10,000 operations |

### Security Characteristics

**Strengths**:
- KEK never in plaintext in application
- KMS provides audit logging (CloudTrail)
- Branch Key TTL limits breach window
- Memguard protects cached Branch Keys from memory dumps
- Context binding prevents cross-service key reuse

**Attack Surface**:
- If attacker gains memory dump during TTL window, can extract Branch Key
- Branch Key compromise allows decryption of all data encrypted under that branch during TTL window
- Does NOT compromise KEK (which remains in AWS KMS)
- Does NOT compromise other Branch Keys (separate per service_id)
- Does NOT compromise historical data encrypted with previous Branch Key versions

**Mitigation Strategies**:
1. **Shorter TTL** (15 minutes vs. 1 hour default) limits breach window
2. **Memguard protection** makes memory extraction harder
3. **MLOCK** prevents swapping to disk
4. **Core dump exclusion** prevents post-mortem forensics
5. **Regular key rotation** in AWS KMS invalidates old Branch Keys
6. **Monitoring** alerts on unusual KMS activity or cache patterns

### Memory Protection with Memguard

```go
// Pseudocode: How memguard protects Branch Keys

type ProtectedBranchKeyCache struct {
    cache map[string]*memguard.Enclave
    mu    sync.RWMutex
}

func (c *ProtectedBranchKeyCache) Get(serviceID string) (*memguard.Enclave, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    enclave, ok := c.cache[serviceID]
    return enclave, ok
}

func (c *ProtectedBranchKeyCache) Put(serviceID string, branchKey []byte, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Move Branch Key into protected memory enclave
    enclave := memguard.NewEnclave(branchKey)

    // Schedule cleanup at TTL expiration
    go func() {
        <-time.After(ttl)
        enclave.Destroy() // Zero and unlock memory
    }()

    c.cache[serviceID] = enclave
}
```

### Configuration Example

```yaml
# .env
encryption:
  key_encryption_key: arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012
  kms:
    branch_key_ttl: 15m           # Recommended: 15 minutes (vs. default 1 hour)
    cache_limit_entries: 100      # Support 100 concurrent service contexts
    memory_protection: true        # Enable memguard integration
```

---

## Approach 2: Direct KMS Calls (Alternative, Higher Security)

### How It Works

1. **Every encrypt/decrypt operation**:
   - Call AWS KMS directly
   - No local key caching
   - Key material only exists during operation

2. **Performance**:
   - Every operation pays 50-200ms KMS latency
   - Consistent latency profile (no variance from cache hits)

3. **Security**:
   - No keys cached in memory
   - Minimal attack surface
   - Highest memory security posture

### Performance Characteristics

| Metric | Value |
|--------|-------|
| **Every operation** | 50-200ms (KMS roundtrip) |
| **Memory overhead** | ~0 (no cached keys) |
| **KMS API calls** | 2 per operation (wrap + unwrap) |
| **AWS KMS cost** | ~$0.30 per 10,000 operations |

### When to Use

- **High-security environments** requiring minimal attack surface
- **Compliance requirements** prohibiting long-lived keys in memory
- **Low-throughput systems** where latency is acceptable
- **Air-gapped deployments** with local KMS caching unavailable

---

## Decision Matrix

| Criterion | Hierarchical Caching | Direct KMS Calls |
|-----------|-----------------|------------------|
| **Performance** | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐ Poor |
| **Cost** | ⭐⭐⭐⭐⭐ Very Low | ⭐⭐ Very High |
| **Memory Security** | ⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ Excellent |
| **Complexity** | ⭐⭐⭐ Medium | ⭐ Low |
| **Availability** | ⭐⭐⭐⭐ Good (tolerates KMS downtime) | ⭐⭐ Poor (KMS-dependent) |
| **Recommended for OAuth** | ✅ Yes | ❌ No (unless required) |

---

## Recommendation

**Use Hierarchical Keyring with Branch Key Caching for OAuth tokens**

### Rationale

1. **Performance Requirements**: OAuth tokens need <100ms response times; direct KMS calls violate SLA
2. **Throughput**: Session management requires fast, responsive encryption
3. **Cost**: KMS costs become prohibitive with direct calls (~10x more expensive)
4. **Security/Availability Tradeoff**: 15-minute TTL with memguard protection provides reasonable security while maintaining usability
5. **Operational**: Branch Keys are per-service_id, making per-service rotation granular and manageable

### Implementation Guidance

**Phase 1 (MVP)**:
- Use default Hierarchical Keyring configuration (1-hour TTL)
- Implement memguard protection for Branch Keys
- Establish baseline performance metrics

**Phase 2 (Hardening)**:
- Reduce TTL to 15 minutes based on security requirements
- Add monitoring/alerting for cache hit rates
- Implement Branch Key rotation policy

**Phase 3 (Optimization)**:
- Fine-tune TTL based on production traffic patterns
- Consider hybrid approach: Hierarchical for normal load, direct KMS for high-security ops
- Implement cost tracking and optimization

### Configuration Defaults

```yaml
# Recommended for OAuth token encryption
encryption:
  key_encryption_key: arn:aws:kms:us-east-1:123456789012:key/...
  kms:
    branch_key_ttl: 15m           # Shorter TTL for enhanced security
    cache_limit_entries: 100      # Adequate for typical deployments
    memory_protection: true        # Always enable memguard
```

---

## Key Clarifications

### Q: If attacker gets memory dump during TTL, can they decrypt all encrypted data?

**A**: No. They can only decrypt data encrypted with the compromised Branch Key during that TTL window:
- Data encrypted with previous Branch Key versions remains protected
- Data encrypted with future Branch Key versions (after TTL) remains protected
- Other service_ids' data remains protected (separate Branch Keys)
- KEK in AWS KMS remains secure

### Q: Isn't this worse than direct KMS calls?

**A**: For pure memory security, yes. But it's better overall:
- 50-200ms latency for every operation is unacceptable for OAuth
- 10x KMS cost increase is prohibitive at scale
- 15-minute TTL + memguard provides acceptable risk profile
- Availability improves (tolerates temporary KMS outages)

### Q: How does memguard help if AWS SDK doesn't use it?

**A**: Memguard protects at the application boundary:
1. **Branch Key storage**: Wrap cached keys in memguard enclaves
2. **Token processing**: Protect plaintext tokens during encrypt/decrypt
3. **Core dump exclusion**: Prevent keys from appearing in core dumps
4. **Memory locking**: Prevent sensitive pages from swapping to disk

### Q: What if I need highest-possible security?

**A**: Consider hybrid approach:
```yaml
encryption:
  kms:
    branch_key_ttl: 5m             # Very short cache window
    memory_protection: true         # Enable memguard
    high_security_mode: true       # Direct KMS for sensitive operations
```

---

## Implementation Checklist

- [ ] Configure AWS Encryption SDK Hierarchical Keyring with 15-minute TTL
- [ ] Implement memguard protection for Branch Key cache
- [ ] Add monitoring for cache hit rates and KMS latency
- [ ] Test Branch Key TTL expiration and refresh
- [ ] Verify memory protection (MLOCK, DONTDUMP)
- [ ] Establish backup/recovery for failed KMS operations
- [ ] Document configuration and operational procedures
- [ ] Plan for key rotation policy and testing
- [ ] Add alerting for unusual KMS activity
- [ ] Test graceful handling of KMS unavailability

---

**Version**: 1.0 | **Status**: Design Decision | **Last Updated**: 2026-01-16
