# Analysis Remediation Review

**Analysis Report**: `/speckit.analyze` findings from 2026-01-19
**Status**: Ready for approval
**Total Remediations**: 5 (1 CRITICAL, 2 HIGH, 2 MEDIUM)

---

## REMEDIATION 1: Red Phase Test Implementation (CRITICAL - NOT SKIPPED)

**File**: `tests/e2e/encryption_vault_test.go`
**Location**: All 24 It() blocks
**Severity**: CRITICAL
**Issue**: E2E tests use `Skip("Not implemented - RED phase")` which violates Constitution Principle XIII. Tests must actually RUN and FAIL against non-existent implementation, not be skipped.

### Current Implementation:
```go
It("encrypts OAuth tokens using DEK bound to service context and wrapped KEK", func() {
    Skip("Not implemented - RED phase")
    // Given a session with OAuth tokens...
    // When tokens are saved...
    // Then each token is encrypted...
})
```

### Problem with Current Approach:
- ❌ Tests are SKIPPED, not RUNNING
- ❌ Tests never FAIL (they're never executed)
- ❌ Violates Constitution Principle XIII: "E2E tests MUST compile AND fail semantically initially"
- ❌ When implementation is done, tests just become un-skipped—no red→green cycle proof
- ✅ Only satisfies "tests written first" part of TDD, not "tests fail first" part

### Proposed Fix (Remove Skip, Implement Actual Test Logic):
```go
It("encrypts OAuth tokens using DEK bound to service context and wrapped KEK", func() {
    // Scenario 1.1 from specs/012-aws-encryption-vault/spec.md

    // GIVEN a session with OAuth tokens and associated service_id context
    principal := "user@example.com"
    serviceID := "oauth2"
    accessToken := "access_token_abc123"
    refreshToken := "refresh_token_xyz789"

    // WHEN tokens are saved to the sessions repository
    session := &storage.UserSession{
        ID:        "session-123",
        Principal: principal,
        ServiceID: serviceID,
        // Note: EncryptedAccessToken/EncryptedRefreshToken will be populated by service during Create()
        // This test verifies the encryption happens before storage
    }

    // THEN the encryption port encrypts tokens with DEK + KEK wrapping
    encryptionContext := map[string]string{"service_id": serviceID}

    // Encrypt access token (simulating what service does)
    encryptedAccess, err := encryptionPort.Encrypt(ctx, []byte(accessToken), encryptionContext)
    Expect(err).ToNot(HaveOccurred())
    Expect(encryptedAccess).NotTo(Equal([]byte(accessToken))) // Verify encryption happened

    // Encrypt refresh token
    encryptedRefresh, err := encryptionPort.Encrypt(ctx, []byte(refreshToken), encryptionContext)
    Expect(err).ToNot(HaveOccurred())
    Expect(encryptedRefresh).NotTo(Equal([]byte(refreshToken)))

    // THEN both ciphertexts are different (unique DEK per encryption)
    Expect(encryptedAccess).NotTo(Equal(encryptedRefresh))

    // THEN decryption with matching context succeeds
    decryptedAccess, err := encryptionPort.Decrypt(ctx, encryptedAccess, encryptionContext)
    Expect(err).ToNot(HaveOccurred())
    Expect(decryptedAccess).To(Equal([]byte(accessToken)))
})
```

### Why This Matters (Constitution Principle XIII):
Red-green TDD requires:
1. ✅ **Tests WRITTEN FIRST** (done—24 It() blocks exist)
2. ❌ **Tests FAIL SEMANTICALLY** (currently skipped—NOT failing)
3. ❌ **Minimal test changes during implementation** (can't verify if tests never run)

By removing Skip() and implementing actual test logic:
- Tests will FAIL when run (because EncryptionPort not yet implemented)
- Tests will PASS as implementation is added
- Red→green cycle is VISIBLE and VERIFIABLE
- Tests serve as executable specification of requirements

### Update T013 in tasks.md:
```
- [ ] T013 [P] Convert E2E tests from Skip() to failing red phase tests:
  - Current: All 24 It() blocks use Skip("Not implemented - RED phase")
  - Required: Remove Skip() and implement actual test logic for each scenario
  - Each test must:
    - Set up Given state (test data, context)
    - Execute When action (encrypt/decrypt, service calls)
    - Assert Then expectations (verify encryption, context binding, etc.)
    - FAIL when run against non-existent implementation
  - Command to verify red phase: `ginkgo -v ./tests/e2e/encryption_vault_test.go`
  - Expected: All 24 tests FAIL with "undefined method" or "nil pointer" errors (red phase)
  - DO NOT SKIP TESTS - they must actually run and fail to verify TDD cycle
```

### Rationale:
- Complies with Constitution Principle XIII: tests fail first, pass when implementation done
- Provides actual verification that tests specify requirements (not just placeholder comments)
- Enables red→green cycle proof in PR
- Ensures tests are independent verification, not just documentation

**Approve?** ☐ Yes ☐ No ☐ Modify (e.g., different test assertion style)

---

## REMEDIATION 2: Edge Case Handling Scenarios

**File**: `spec.md`
**Location**: After line 122 (after User Story 7 header)
**Severity**: HIGH
**Issue**: Edge cases mentioned (lines 126-132) but no acceptance scenarios defined

### Proposed Addition (new User Story 8):
```markdown
---

### User Story 8 - Edge Cases & Error Handling (Priority: P1)

A reliability engineer needs the system to handle edge cases gracefully: large tokens that exceed encryption buffer limits, DEK generation failures, KEK unavailability during decryption, and context mismatch scenarios. Errors must fail fast and clearly rather than silently corrupting data or causing data loss.

**Why this priority**: Error handling is critical for production stability. Edge cases must fail fast with clear error messages to operators.

**Independent Test**: Can be fully tested by simulating each edge case and verifying application responses are safe, predictable, and fail-closed.

**Acceptance Scenarios**:

1. **Given** a token larger than 1MB (boundary test), **When** encryption is attempted, **Then** encryption fails with ErrorKindEncryptionFailed and error message includes max buffer size (no silent truncation)
2. **Given** DEK generation fails (e.g., insufficient randomness from OS), **When** session creation is attempted, **Then** application fails with ErrorKindEncryptionFailed and no session is created (fail-closed)
3. **Given** AWS KMS becomes unavailable during token decryption (network failure, service down), **When** GetSession is called, **Then** application fails with ErrorKindKEKUnavailable after AWS SDK timeout (default 10s) with clear error message for operator
4. **Given** encryption context (service_id) mismatches between encryption and decryption, **When** token decryption is attempted, **Then** failure occurs at both DEK verification AND KEK unwrap layers with ErrorKindContextMismatch error
5. **Given** ciphertext is tampered (bytes corrupted due to storage fault), **When** decryption is attempted, **Then** AESGCMSIV authentication tag verification fails with ErrorKindIntegrityViolation (no silent data corruption)
6. **Given** plaintext token is nil or empty, **When** encryption is attempted, **Then** encryption succeeds (empty tokens are valid; edge case handled correctly)

---
```

### Update Success Criteria (add new line after SC-013):
```markdown
- **SC-014**: Edge cases handled safely—large tokens, DEK generation failures, KEK unavailability, context mismatch, integrity violations all fail fast with clear errors
```

### Rationale:
- Converts vague edge case questions to testable acceptance scenarios
- Ensures fail-closed behavior (no silent data corruption)
- Maps edge cases to E2E tests (add new US8 test scenarios)
- Improves reliability and operational visibility

**Approve?** ☐ Yes ☐ No ☐ Modify

---

## REMEDIATION 3: LocalStack Integration in docker-compose.yml (KMS + DynamoDB)

**File**: `docker-compose.yml`
**Location**: Add LocalStack service after line 141 (after frontend service)
**Severity**: MEDIUM
**Issue**: Integration tests require LocalStack for KMS and DynamoDB (branch key cache), but no service configured in main docker-compose.yml

### Proposed Addition (new LocalStack service to docker-compose.yml):
```yaml
  # LocalStack for AWS KMS + DynamoDB testing (used in Phase 5 integration tests)
  # KMS: Encrypts/decrypts data keys; DynamoDB: Caches branch keys for performance
  localstack:
    image: localstack/localstack:latest
    container_name: aib-localstack
    ports:
      - "4566:4566"  # LocalStack API endpoint (AWS SDK will connect here)
    environment:
      SERVICES: kms,dynamodb
      AWS_ACCESS_KEY_ID: test
      AWS_SECRET_ACCESS_KEY: test
      AWS_DEFAULT_REGION: eu-central-1
      DOCKER_HOST: unix:///var/run/docker.sock
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    networks:
      - aib-network
    healthcheck:
      test: ["CMD", "awslocal", "kms", "list-keys"]
      interval: 5s
      timeout: 2s
      retries: 3
      start_period: 10s
```

### Update identity-broker service (add to depends_on section):
```yaml
    depends_on:
      upstream-oauth2:
        condition: service_started
      third-party-oauth2:
        condition: service_started
      localstack:                        # Add this line
        condition: service_healthy        # Add this line
```

### Add Setup Script: `scripts/localstack-setup.sh`
```bash
#!/bin/bash
# Setup LocalStack KMS + DynamoDB for testing
set -e

ENDPOINT_URL="http://localhost:4566"
REGION="eu-central-1"
BRANCH_KEY_TABLE="IdentityBrokerEncryptionBranchKeys"

# Wait for LocalStack to be ready
echo "Waiting for LocalStack to be ready..."
for i in {1..30}; do
  if aws --endpoint-url "$ENDPOINT_URL" --region "$REGION" kms list-keys &>/dev/null; then
    echo "✅ LocalStack ready"
    break
  fi
  echo "  Attempt $i/30..."
  sleep 1
done

# Create test KMS key (alias: test-encryption-key)
echo "Creating test KMS key..."
KEY_ID=$(aws --endpoint-url "$ENDPOINT_URL" --region "$REGION" kms create-key --description "Test encryption key" --query 'KeyMetadata.KeyId' --output text)
echo "✅ Created key: $KEY_ID"

# Create alias for easy reference
aws --endpoint-url "$ENDPOINT_URL" --region "$REGION" kms create-alias --alias-name "alias/test-encryption-key" --target-key-id "$KEY_ID"
echo "✅ Created alias: alias/test-encryption-key"

# Create DynamoDB table for branch key cache
echo "Creating DynamoDB table for branch keys..."
aws --endpoint-url "$ENDPOINT_URL" --region "$REGION" dynamodb create-table \
  --table-name "$BRANCH_KEY_TABLE" \
  --attribute-definitions \
    AttributeName=branchKeyId,AttributeType=S \
    AttributeName=sort_key,AttributeType=S \
  --key-schema \
    AttributeName=branchKeyId,KeyType=HASH \
    AttributeName=sort_key,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --ttl-specification Enabled=true,AttributeName=ttl 2>/dev/null || echo "⚠️  Table may already exist"
echo "✅ DynamoDB table ready: $BRANCH_KEY_TABLE"

# Export for use in tests
export AWS_ENDPOINT_URL="$ENDPOINT_URL"
export AWS_KMS_KEY_ID="$KEY_ID"
export AWS_KMS_ENDPOINT="$ENDPOINT_URL"
export IDENTITY_BROKER_ENCRYPTION_DYNAMODB_TABLE_NAME="$BRANCH_KEY_TABLE"

echo ""
echo "LocalStack setup complete!"
echo "Export these for tests:"
echo "  export AWS_ENDPOINT_URL=$ENDPOINT_URL"
echo "  export AWS_KMS_KEY_ID=$KEY_ID"
echo "  export IDENTITY_BROKER_ENCRYPTION_DYNAMODB_TABLE_NAME=$BRANCH_KEY_TABLE"
```

### Update tasks.md (modify T026 description):
```markdown
- [x] T026 [P] Create LocalStack-compatible integration test infrastructure:
  - LocalStack service added to docker-compose.yml with KMS + DynamoDB (eu-central-1)
  - Setup script: scripts/localstack-setup.sh creates:
    - Test KMS key (alias/test-encryption-key)
    - DynamoDB table for branch key caching (IdentityBrokerEncryptionBranchKeys with TTL)
  - Usage: `docker-compose up -d` starts LocalStack, then `bash scripts/localstack-setup.sh` initializes both services
  - Integration tests use AWS_ENDPOINT_URL=http://localhost:4566
  - Exports environment variables for test configuration
  - CI/CD: GitHub Actions can use LocalStack container for KMS + DynamoDB integration tests
```

**Rationale**:
- LocalStack service integrated into existing docker-compose.yml
- Supports both KMS (key wrapping) and DynamoDB (branch key cache) for AWS Encryption SDK
- Enables full integration tests without AWS account
- Supports local development, CI/CD, and team testing
- Setup script automates key and table creation; exports environment variables
- Region: eu-central-1 (Zalando's default AWS region)
- Branch key TTL supported for automatic cache expiration
- Single LocalStack service is more efficient than separate services

**Approve?** ☐ Yes ☐ No ☐ Modify

---

## REMEDIATION 4: User Story 6 Consolidation

**File**: `spec.md`
**Location**: Lines 92-105 (User Story 6)
**Severity**: MEDIUM
**Issue**: US1 and US6 test same mechanism (context verification); US6 should be higher-level validation

### Current Text (Lines 92-105):
```markdown
### User Story 6 - Encryption Context Prevents Token Reuse Across Services (Priority: P2)

A security architect needs tokens to be bound to their specific service context (service_id) at the cryptographic level. Tokens encrypted for one service must not be usable for a different service, even if an attacker has access to ciphertext from both services.

**Why this priority**: Encryption context adds defense in depth by binding tokens to specific services. While not essential for MVP, it's an important security enhancement that prevents cross-service token reuse attacks.

**Independent Test**: Can be fully tested by encrypting tokens with specific service context values, attempting to decrypt with different service context values, and verifying decryption fails at both the DEK and KEK verification layers.

**Acceptance Scenarios**:

1. **Given** a token is encrypted with encryption context `{"service_id": "oauth2"}`, **When** decryption is attempted with matching context, **Then** decryption succeeds
2. **Given** a token encrypted for one service, **When** decryption is attempted with a different service_id value, **Then** decryption fails at both DEK verification and KEK unwrapping
3. **Given** tokens for different services stored in the database, **When** an attacker tries to use ciphertext from one service with a different service_id, **Then** both DEK decryption and KEK unwrapping fail, and the attack is prevented

---
```

### Proposed Text:
```markdown
### User Story 6 - Multi-Service Isolation via Encryption Context (Priority: P2)

A platform operator managing multiple OAuth2 services (GitHub, Google, custom OAuth providers) needs tokens from different services to be isolated at the cryptographic level: tokens encrypted for service A cannot be used for service B, preventing cross-service token reuse attacks even if an attacker compromises the database and gains direct read access to encrypted tokens and wrapped DEKs.

**Why this priority**: Multi-service isolation validates that core envelope encryption (US1) prevents high-level cross-service attacks. US1 enforces context verification at DEK/KEK layers; US6 validates this mechanism prevents practical attack scenarios. P2 because context verification is mandatory (US1), but testing it across multiple services is validation/integration work.

**Dependencies**: Depends on User Story 1 (Envelope Encryption) for the context verification mechanism.

**Independent Test**: Can be fully tested by creating sessions for multiple services, extracting encrypted ciphertexts from the database, and verifying ciphertext from service A cannot be decrypted in service B's context through any attack vector.

**Acceptance Scenarios**:

1. **Given** an application managing sessions for services "oauth2", "github", and "google", **When** sessions are created for each service with unique tokens and encryption context, **Then** each session has a unique DEK wrapped with its own service_id context (no shared DEK across services)
2. **Given** ciphertext from service "oauth2" stored in database, **When** decryption is attempted with service_id "github", **Then** decryption fails at both DEK verification layer (AAD mismatch) AND KEK unwrap layer (context mismatch), preventing cross-service reuse
3. **Given** an attacker with database access extracts encrypted tokens and wrapped DEKs from service "oauth2" and attempts to decrypt them as service "github" tokens, **When** they call the application's decryption endpoint with wrong service_id context, **Then** both cryptographic layers reject the ciphertext with ErrorKindContextMismatch, and the attack is prevented with clear audit log entry

---
```

### Rationale:
- Reframes US6 as multi-service validation layer (not core feature)
- Clarifies dependency on US1
- Adds practical attack scenario
- Improves clarity: US1 = mechanism, US6 = multi-service validation

**Approve?** ☐ Yes ☐ No ☐ Modify

---

## REMEDIATION 5: Schema Verification Task

**File**: `tasks.md`
**Location**: New task between T014 and T015 (after line 86)
**Severity**: MEDIUM
**Issue**: Schema assumptions stated but not verified before Phase 3 starts

### Proposed Addition (new task T014a):
```markdown
- [x] T014a [P] Verify schema assumptions before Phase 3 starts:
  - Prerequisite: Confirm UserSession aggregate has required BYTEA/JSONB columns
  - Verification Commands:
    - `grep -A 10 "type UserSession struct" internal/domain/storage/user_session.go | grep -E "Encrypted|EncryptionContext"`
    - Expected: EncryptedAccessToken []byte, EncryptedRefreshToken []byte, EncryptionContext struct
  - Status: ✅ VERIFIED
    - Line 17: `EncryptedAccessToken  []byte` ✅
    - Line 18: `EncryptedRefreshToken []byte` ✅
    - Line 23: `EncryptionContext     EncryptionContext` ✅
    - Line 29-36: EncryptionContext struct with JSONB serialization ✅
  - Gate: All checks passed - Phase 3 can proceed
  - Note: No new migrations required; columns already exist from feature 005-session-management
```

### Rationale:
- Documents schema verification before implementation
- Prevents false assumptions
- Gate for Phase 3 (foundational infrastructure)

**Approve?** ☐ Yes ☐ No ☐ Modify

---

## Summary Table

| ID | File | Section | Change | Priority | Status |
|----|------|---------|--------|----------|--------|
| R1 | tests/e2e/ | encryption_vault_test.go | Remove Skip(), implement failing test logic | CRITICAL | ⏳ Pending |
| R2 | spec.md | After line 122 | Add US8 (Edge Cases) with 6 scenarios | HIGH | ⏳ Pending |
| R3 | docker-compose.yml | After line 141 | Add LocalStack KMS+DynamoDB for manual testing | MEDIUM | ⏳ Pending |
| R3b | tests/e2e/bootstrap/ | New file: localstack.go | E2E testcontainers bootstrap for automated tests | MEDIUM | ⏳ Pending |
| R4 | spec.md | Lines 92-105 | Reframe US6 as multi-service validation | MEDIUM | ⏳ Pending |
| R5 | tasks.md | After line 86 | Add T014a schema verification task | MEDIUM | ⏳ Pending |

---

## REMEDIATION 3b: E2E Bootstrap for LocalStack (testcontainers)

**File**: `tests/e2e/bootstrap/localstack.go` (NEW)
**Location**: New testcontainers bootstrap module (follows existing pattern in bootstrap/ directory)
**Severity**: MEDIUM
**Issue**: E2E encryption tests need automated LocalStack setup; should follow established tests/e2e/bootstrap/ pattern (like StorageFactory, TestServerFactory)

### Proposed Addition (new file: `tests/e2e/bootstrap/localstack.go`):

Create testcontainers-based LocalStack bootstrap:

```go
package bootstrap

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// LocalStackContainer manages LocalStack KMS + DynamoDB for E2E encryption tests
type LocalStackContainer struct {
	Container testcontainers.Container
	Endpoint  string
	KMSKeyID  string
}

// StartLocalStack creates and starts a LocalStack container with KMS + DynamoDB
// Usage: In BeforeEach, `ls := bootstrap.StartLocalStack(ctx, t)`
// Usage: In AfterEach, `defer ls.Terminate(ctx)`
func StartLocalStack(ctx context.Context, t *testing.T) *LocalStackContainer {
	req := testcontainers.ContainerRequest{
		Image: "localstack/localstack:latest",
		Env: map[string]string{
			"SERVICES":              "kms,dynamodb",
			"AWS_ACCESS_KEY_ID":     "test",
			"AWS_SECRET_ACCESS_KEY": "test",
			"AWS_DEFAULT_REGION":    "eu-central-1",
		},
		ExposedPorts: []string{"4566/tcp"},
		WaitingFor:   wait.ForLog("Ready."),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start LocalStack container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "4566")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get mapped port: %v", err)
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())

	// Create KMS key
	kmsClient := kms.NewFromConfig(awsConfigForLocalStack(ctx, endpoint))
	keyOutput, err := kmsClient.CreateKey(ctx, &kms.CreateKeyInput{
		Description: wrap("Test encryption key"),
	})
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to create KMS key: %v", err)
	}

	keyID := *keyOutput.KeyMetadata.KeyId

	// Create DynamoDB table for branch key cache
	dynamoClient := dynamodb.NewFromConfig(awsConfigForLocalStack(ctx, endpoint))
	_, err = dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: wrap("IdentityBrokerEncryptionBranchKeys"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: wrap("branchKeyId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: wrap("sort_key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: wrap("branchKeyId"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: wrap("sort_key"),
				KeyType:       types.KeyTypeRange,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
		TimeToLiveSpecification: &types.TimeToLiveSpecification{
			Enabled:       wrap(true),
			AttributeName: wrap("ttl"),
		},
	})
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to create DynamoDB table: %v", err)
	}

	return &LocalStackContainer{
		Container: container,
		Endpoint:  endpoint,
		KMSKeyID:  keyID,
	}
}

// Terminate stops and removes the LocalStack container
func (ls *LocalStackContainer) Terminate(ctx context.Context) error {
	return ls.Container.Terminate(ctx)
}

// awsConfigForLocalStack creates AWS SDK config pointing to LocalStack endpoint
func awsConfigForLocalStack(ctx context.Context, endpoint string) aws.Config {
	cfg, _ := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-central-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint}, nil
			})),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	return cfg
}

func wrap(s string) *string { return &s }
func wrapBool(b bool) *bool  { return &b }
```

### Usage in E2E tests (update BeforeEach in encryption_vault_test.go):
```go
var (
	ls *bootstrap.LocalStackContainer
)

BeforeEach(func() {
	ctx = context.Background()
	ls = bootstrap.StartLocalStack(ctx, GinkgoT())
	// TODO: Initialize encryption adapter with ls.Endpoint and ls.KMSKeyID
})

AfterEach(func() {
	if ls != nil {
		ls.Terminate(ctx)
	}
})
```

### Rationale:
- Follows established tests/e2e/bootstrap/ pattern (StorageFactory, TestServerFactory)
- Automated testcontainers lifecycle: fresh LocalStack per test, automatic cleanup
- Each test isolated (no state pollution between tests)
- Combines KMS key creation + DynamoDB table setup
- Region: eu-central-1 (Zalando default)
- No manual container management needed

**Approve?** ☐ Yes ☐ No ☐ Modify

---

## Approval Instructions

**For each remediation, please indicate:**

1. **R1 (Red Phase Tests)**: ☐ Approve ☐ Reject ☐ Modify (comment)
2. **R2 (Edge Cases)**: ☐ Approve ☐ Reject ☐ Modify (comment)
3. **R3 (LocalStack docker-compose)**: ☐ Approve ☐ Reject ☐ Modify (comment)
4. **R3b (E2E LocalStack bootstrap)**: ☐ Approve ☐ Reject ☐ Modify (comment)
5. **R4 (US6 Consolidation)**: ☐ Approve ☐ Reject ☐ Modify (comment)
6. **R5 (Schema Verification)**: ☐ Approve ☐ Reject ☐ Modify (comment)

**Once approved, I will:**
- [ ] Apply all approved edits to spec.md, tasks.md, docker-compose.yml
- [ ] Create scripts/localstack-setup.sh helper
- [ ] Commit changes to feature branch with clear commit message
- [ ] Update analysis report status to ✅ COMPLETE

---

**Document Version**: 2.0 (Updated: removed R2 Performance SLA, consolidated to 5 remediations)
**Created**: 2026-01-19
**Ready for User Review**: ✅ Yes
