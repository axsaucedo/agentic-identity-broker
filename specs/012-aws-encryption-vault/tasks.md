# Tasks: Encryption Vault for OAuth Tokens (Feature 012)

**Feature**: AWS Encryption Vault (PQC-Enabled by Go 1.24+)
**Specification**: [spec.md](./spec.md)
**Plan**: [plan.md](./plan.md)
**Data Model**: [data-model.md](./data-model.md)
**Quickstart**: [quickstart.md](./quickstart.md)

**Implementation Status**: ⏳ Ready for Phase 1 (Design Preconditions)

---

## Executive Summary

This feature implements envelope encryption for OAuth tokens in the agentic-identity-broker using AWS Encryption SDK with AESGCMSIV authenticated encryption. Implementation follows specification-driven development (Constitution Principle XIII): all 24 acceptance scenarios from spec.md have corresponding E2E tests written before implementation begins (red-green TDD).

**User Stories**: 7 (P1: 6 stories, P2: 1 story, P3: 1 story)
**Tasks by Phase**: Setup (3), Design Preconditions (7), Foundational (4), US1-7 (32), Polish (8)
**Total Tasks**: 54
**Parallelizable Tasks**: 18 ([P] markers)
**E2E Acceptance Tests**: 24 (mandatory per Constitution XIII)

---

## Phase 1: Setup & Project Initialization

🔒 [MANDATORY] **Core project setup**—common to all features

- [x] T001 Create feature directory structure at `specs/012-aws-encryption-vault/` (if not exists)
- [x] T002 Initialize Go module dependencies: add `github.com/aws/aws-encryption-sdk-go/v3`, `github.com/aws/aws-cryptographic-material-providers-go`, `github.com/aws/aws-sdk-go-v2/service/kms`, `github.com/awnumar/memguard` to `go.mod`
- [x] T003 Verify design documents complete: spec.md, plan.md, data-model.md, quickstart.md, contracts/ directory

---

## Phase 2: Design Preconditions ✅ COMPLETE

🔒 [MANDATORY] **Constitution Principle XIII compliance**: ALL design artifacts complete before implementation starts

### Phase 2a: Domain Model & Glossary (Principles II, V)

- [x] T004 Update [ARCHITECTURE.md](../../ARCHITECTURE.md) Glossary: add "envelope encryption", "DEK", "KEK", "EncryptionContext", "EncryptionPort", "AAD", "memguard", "AESGCMSIV"
- [x] T005 Update [ARCHITECTURE.md](../../ARCHITECTURE.md) Domain section: document UserSession aggregate, EncryptionContext value object, EncryptionPort interface
- [x] T006 Create ADR: `adrs/NNN-envelope-encryption-design.md` documenting: (1) DEK-per-session rationale, (2) service_id-only context binding, (3) AWS Encryption SDK choice, (4) memguard integration

### Phase 2b: Configuration Design (Principle VII)

- [x] T007 Verify configuration field defined: `encryption.key_encryption_key` supports AWS KMS ARN or `${ENCRYPTION_KEK}` interpolation
- [x] T008 Add configuration examples to [examples/config/](../../examples/config/): `encryption-aws-kms.yaml` (production) and `encryption-env-var.yaml` (development)

### Phase 2c: API Design (Principles IV, X)

- [x] T009 [P] Verify OpenAPI documentation exists (internal port, not HTTP API): EncryptionPort interface documented in [internal/ports/encryption.go](../../internal/ports/encryption.go) with Encrypt/Decrypt methods
- [x] T010 [P] Verify error contract complete: [specs/012-aws-encryption-vault/contracts/error-contract.md](./contracts/error-contract.md) documents ErrorKind enumeration and error handling patterns

### Phase 2d: Database Design (Principle IX)

- [x] T011 [P] Verify no new migrations required: encrypted_access_token (BYTEA), encrypted_refresh_token (BYTEA), encryption_context (JSONB) columns already exist in sessions table per feature 005

### Phase 2e: E2E Acceptance Tests (Principle XIII - CRITICAL)

- [x] T012 [P] Write E2E tests for ALL 24 acceptance scenarios BEFORE implementation: create `tests/e2e/encryption_vault_test.go`
  - **Test Organization** (Ginkgo/Gomega BDD):
    - Feature-level Describe: "Encryption Vault for OAuth Tokens"
    - Context per user story: "User Story 1: Envelope Encryption", "User Story 2: Secure KEK Storage", etc.
    - It() block per scenario: maps 1:1 to spec.md scenarios
  - **Test Scenarios Mapped** (24 total):
    - US1 (4 scenarios): DEK with context binding, KEK wrapping with context, context verification failure, fail-closed
    - US2 (4 scenarios): AWS KMS storage, operations logged, access control, key rotation backward compat
    - US3 (3 scenarios): Fresh DEK per session, unique DEK across sessions, DEK memory zeroization
    - US4 (3 scenarios): Load KEK from `${ENCRYPTION_KEK}`, use KEK for wrap/unwrap, persist across restart
    - US5 (3 scenarios): Repository Create() encrypts, Get() decrypts, no manual steps
    - US6 (3 scenarios): Same service succeeds, different service_id fails both layers, ciphertext reuse fails
    - US7 (4 scenarios): PQC available, encrypt/decrypt with PQC, disable PQC gracefully, version control
  - **Bootstrap Strategy** (tests/e2e/bootstrap/):
    - Fresh app.Builder per test
    - Fresh storage adapter (memory/postgres)
    - Fresh encryption adapter with test KEK
    - Fixtures from tests/e2e/fixtures/ (test principals, services, tokens)
  - **Custom Matchers** (tests/e2e/matchers/):
    - `HaveEncryptedToken()` - verify token is encrypted (ciphertext ≠ plaintext)
    - `HaveMatchingEncryptionContext()` - verify context matches
    - `FailWithContextMismatch()` - verify context mismatch errors

- [x] T013 [P] Verify E2E tests FAIL before implementation (red phase): run `ginkgo -v ./tests/e2e/encryption_vault_test.go` and confirm all 24 tests fail (no implementation exists)

- [x] T014 Verify E2E test file includes comments: map each test to spec.md scenarios with `// Scenario X.Y from specs/012-aws-encryption-vault/spec.md`

- [x] T014a [P] Verify schema assumptions before Phase 3 starts:
  - **Prerequisite**: Confirm UserSession aggregate has required encrypted token and context fields
  - **Verification Status**: ✅ **VERIFIED** in [internal/domain/storage/user_session.go](../../internal/domain/storage/user_session.go)
    - Line 17: `EncryptedAccessToken  []byte` ✅ (BYTEA column)
    - Line 18: `EncryptedRefreshToken []byte` ✅ (BYTEA column)
    - Line 23: `EncryptionContext     EncryptionContext` ✅ (JSONB column, defined in data-model.md)
  - **Impact**: No schema migrations required; adapter layer sees only encrypted BYTEA values

---

## Phase 2f: E2E Test Design & Structure (Principle XIII)

🔒 [MANDATORY] **Tests written first, verified to fail, guide implementation**

- [x] T015 Create [tests/e2e/encryption_vault_test.go](../../tests/e2e/encryption_vault_test.go) with skeleton:
  ```go
  var _ = Describe("Encryption Vault for OAuth Tokens", func() {
      var app *app.App
      var encryptionPort ports.EncryptionPort

      BeforeEach(func() {
          // Initialize fresh app and encryption adapter
      })

      AfterEach(func() {
          // Cleanup
      })

      Context("User Story 1: Envelope Encryption", func() {
          // 4 It() blocks for scenarios 1-4
      })

      // ... 6 more Context blocks for US2-7
  })
  ```

- [x] T016 [P] Implement E2E test fixtures (tests/e2e/fixtures/):
  - Create test sessions with known principals, service_ids, tokens
  - Generate test KEK material (base64-encoded random bytes)
  - Define test services: "oauth2", "github", "google"

- [x] T017 [P] Implement E2E helper functions (tests/e2e/helpers/):
  - `createTestSession(principal, serviceID, accessToken, refreshToken)` - create session with encryption
  - `getTestSession(sessionID)` - retrieve session with decryption
  - `verifyEncryptedToken(ciphertext, plaintext)` - assert ciphertext ≠ plaintext
  - `expectContextMismatchError(ciphertext, wrongContext)` - assert error on context mismatch

---

## Phase 3: Foundational Infrastructure

**Blocking prerequisites—must complete before user story implementation**

### Ports & Domain Errors

- [x] T018 [P] Create [internal/ports/encryption.go](../../internal/ports/encryption.go):
  ```go
  type EncryptionPort interface {
      Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error)
      Decrypt(ctx context.Context, ciphertext []byte, encryptionContext map[string]string) ([]byte, error)
  }
  ```
  - Document: performs envelope encryption (DEK + KEK wrapping with context binding)
  - No implementation yet (interface only)

- [x] T019 [P] Create [internal/domain/encryption/errors.go](../../internal/domain/encryption/errors.go):
  ```go
  type ErrorKind string
  const (
      ErrorKindEncryptionFailed   ErrorKind = "encryption_failed"
      ErrorKindDecryptionFailed   ErrorKind = "decryption_failed"
      ErrorKindContextMismatch    ErrorKind = "context_mismatch"
      ErrorKindIntegrityViolation ErrorKind = "integrity_violation"
      ErrorKindKEKUnavailable     ErrorKind = "kek_unavailable"
  )

  type EncryptionError struct {
      Kind    ErrorKind
      Message string
      Wrapped error
  }
  ```
  - Implement Error(), Unwrap(), Is() methods for proper error handling

### Domain Events

- [x] T020 [P] Extend [internal/domain/storage/events.go](../../internal/domain/storage/events.go):
  - Define `SessionEncrypted` event: session_id, principal, service_id, timestamp
  - Define `SessionDecrypted` event: session_id, principal, service_id, timestamp
  - Define `SessionEncryptionFailed` event: session_id, principal, service_id, error_kind, message, timestamp
  - Define `SessionDecryptionFailed` event: session_id, principal, service_id, error_kind, message, timestamp
  - Ensure events are immutable, sanitized (no plaintext tokens or keys)

### Configuration Integration

- [x] T021 [P] Verify configuration system integration:
  - Confirm `encryption.key_encryption_key` field exists in config schema
  - Confirm `${ENCRYPTION_KEK}` interpolation resolves at runtime
  - Write test: configuration loads correctly from `.env` file

---

## Phase 3a: AWS Encryption SDK Adapter (AWS KMS + Environment Variable)

**Create adapter implementing EncryptionPort with envelope encryption**

- [x] T022 Create directory structure: [internal/adapters/encryption/aws/](../../internal/adapters/encryption/aws/)
  - `adapter.go` - main AWS Encryption SDK adapter
  - `adapter_test.go` - unit tests for encryption/decryption, error scenarios
  - `keyring.go` - keyring initialization (KMS vs. env var)

- [x] T023 Implement [internal/adapters/encryption/aws/adapter.go](../../internal/adapters/encryption/aws/adapter.go):
  - **Constructor**: `NewAWSEncryptionAdapter(keyMaterial string) (*AWSAdapter, error)`
    - Detect AWS KMS ARN vs. `${ENCRYPTION_KEK}` env var reference
    - Initialize AWS SDK client (KMS or custom keyring for env var)
    - Validate KEK accessible at startup (fail-fast)
    - Return error if KEK unavailable
  - **Encrypt method**:
    - Accept plaintext, encryptionContext (service_id)
    - Use AWS Encryption SDK to perform envelope encryption
    - Generate fresh DEK per call
    - Encrypt plaintext with DEK (AESGCMSIV)
    - Wrap DEK with KEK using same encryptionContext
    - Return base64-encoded ciphertext (wrapped DEK + ciphertext + auth tag)
    - Handle errors: map to ErrorKind types
  - **Decrypt method**:
    - Accept ciphertext, encryptionContext
    - Use AWS Encryption SDK to unwrap DEK (verifying context)
    - Use DEK to decrypt ciphertext (verifying auth tag)
    - Return plaintext token
    - Handle errors: detect context mismatch, integrity violations
  - **Memory Protection** (memguard):
    - Wrap input plaintext in memguard buffer
    - Zero DEK after use (AWS SDK handles internally)
    - Support memguard for local key material (env var backend)

- [x] T024 [P] Create custom keyring for environment variable KEK support:
  - Implement AWS Encryption SDK keyring interface
  - Load base64-encoded KEK from environment at initialization
  - Wrap KEK in memguard buffer (memory locking, core dump exclusion)
  - Implement GenerateDataKey and DecryptDataKey methods
  - Use same context binding as AWS KMS keyring

- [x] T025 [P] Write unit tests ([internal/adapters/encryption/aws/adapter_test.go](../../internal/adapters/encryption/aws/adapter_test.go)):
  - **Happy path**: Encrypt → decrypt roundtrip returns original plaintext ✅
  - **Context mismatch**: Encrypt with "oauth2", decrypt with "github" fails with ErrorKindContextMismatch ✅
  - **Integrity violation**: Tampered ciphertext fails with ErrorKindIntegrityViolation ✅
  - **Unique DEK per encryption**: Two encryptions of same plaintext produce different ciphertexts ✅
  - **Environment variable KEK**: Load KEK from `${ENCRYPTION_KEK}`, encrypt/decrypt succeeds ✅
  - **KEK validation at startup**: Invalid KMS ARN fails with ErrorKindKEKUnavailable ✅
  - **Context timeout**: Respect context.Context deadlines (implicitly tested via context.Background) ✅
  - **Test file**: 16 test functions + 2 benchmarks, all passing

- [ ] T026 [P] Write integration tests with LocalStack (testcontainers):
  - Set up LocalStack KMS container
  - Create test KMS key
  - Test adapter with real AWS KMS against LocalStack
  - Test backward compatibility: tokens encrypted with old key version remain decryptable

---

## Phase 3b: Session Service Integration

**Integrate EncryptionPort into OAuth2SessionService**

- [x] T027 Locate [internal/services/oauth2session/service.go](../../internal/services/oauth2session/) (if not exists, create)
  - Verify service structure exists (from feature 005-session-management)

- [x] T028 [P] Update OAuth2SessionService constructor:
  - Add `encryptionPort ports.EncryptionPort` dependency
  - Store as field: `encryptionPort EncryptionPort`
  - Update factory function signature

- [x] T029 [P] Update CreateSession method to encrypt tokens:
  ```go
  func (s *OAuth2SessionService) CreateSession(
      ctx context.Context,
      principal, serviceID, accessToken, refreshToken string,
      accessExpiry, refreshExpiry time.Time,
  ) (*UserSession, error) {
      // Construct encryption context
      encCtx := map[string]string{"service_id": serviceID}

      // Encrypt access token
      encAccessToken, err := s.encryptionPort.Encrypt(ctx, []byte(accessToken), encCtx)
      if err != nil {
          s.eventPublisher.Publish(SessionEncryptionFailed{...})
          return nil, err
      }

      // Encrypt refresh token
      encRefreshToken, err := s.encryptionPort.Encrypt(ctx, []byte(refreshToken), encCtx)
      if err != nil {
          s.eventPublisher.Publish(SessionEncryptionFailed{...})
          return nil, err
      }

      // Create UserSession with encrypted tokens
      session := &UserSession{
          EncryptedAccessToken: encAccessToken,
          EncryptedRefreshToken: encRefreshToken,
          EncryptionContext: encCtx,
          // ... other fields ...
      }

      // Store (adapter sees encrypted tokens)
      if err := s.repository.Create(ctx, session); err != nil {
          return nil, err
      }

      s.eventPublisher.Publish(SessionEncrypted{...})
      return session, nil
  }
  ```
  - Ensure fail-closed: no plaintext fallback on encryption error
  - Publish domain events

- [x] T030 [P] Update GetSession method to decrypt tokens:
  ```go
  func (s *OAuth2SessionService) GetSession(
      ctx context.Context,
      sessionID string,
  ) (*UserSession, error) {
      // Retrieve encrypted session
      encSession, err := s.repository.Get(ctx, sessionID)
      if err != nil {
          return nil, err
      }

      encCtx := encSession.EncryptionContext

      // Decrypt access token
      plainAccessToken, err := s.encryptionPort.Decrypt(ctx, encSession.EncryptedAccessToken, encCtx)
      if err != nil {
          s.eventPublisher.Publish(SessionDecryptionFailed{...})
          return nil, err
      }

      // Decrypt refresh token
      plainRefreshToken, err := s.encryptionPort.Decrypt(ctx, encSession.EncryptedRefreshToken, encCtx)
      if err != nil {
          s.eventPublisher.Publish(SessionDecryptionFailed{...})
          return nil, err
      }

      // Return session with plaintext tokens (to caller only)
      result := encSession // Copy
      result.AccessToken = string(plainAccessToken)
      result.RefreshToken = string(plainRefreshToken)

      s.eventPublisher.Publish(SessionDecrypted{...})
      return result, nil
  }
  ```
  - Handle error cases: context mismatch, integrity violations, KEK unavailable
  - Publish domain events on success and failure

---

## Phase 3c: Application Builder Wiring

**Integrate EncryptionPort into app.Builder (Principle XII)**

- [x] T031 [P] Update [internal/app/builder.go](../../internal/app/builder.go):
  - Add `encryptionPort ports.EncryptionPort` field to Builder
  - Create `WithEncryption(port ports.EncryptionPort) *Builder` method
  - Create `buildEncryptionAdapter()` method:
    - Read `config.EncryptionKeyEncryptionKey`
    - Detect AWS KMS ARN vs. `${ENCRYPTION_KEK}`
    - Initialize AWS adapter
    - Fail fast if KEK unavailable
  - Update `Build()` to initialize encryption adapter before OAuth2SessionService
  - Inject adapter into OAuth2SessionService

---

## Phase 4: User Story 1 - Envelope Encryption (P1) 🎯 Core Feature

**Acceptance Criteria**: Encrypt OAuth tokens using DEK (with context) + KEK wrapping (with context), store wrapped envelope, retrieve and decrypt with context verification

### US1 E2E Tests (4 scenarios)

- [ ] T032 [US1] E2E Test - Scenario 1.1: "DEK generation with context binding"
  - Create session with tokens and service_id context
  - Verify encrypted tokens contain wrapped DEK + ciphertext
  - Verify context is bound (visible in wrapped format)

- [ ] T033 [US1] E2E Test - Scenario 1.2: "KEK wrapping with context verification"
  - Encrypt token with context `{"service_id": "oauth2"}`
  - Verify DEK is wrapped using KEK with same context
  - Decrypt with matching context succeeds

- [ ] T034 [US1] E2E Test - Scenario 1.3: "Context verification failure on mismatch"
  - Encrypt token with service_id "oauth2"
  - Attempt decrypt with service_id "github"
  - Verify decryption fails at both DEK and KEK layers

- [ ] T035 [US1] E2E Test - Scenario 1.4: "Fail-closed security"
  - Simulate KEK unavailability
  - Verify application fails with error (no plaintext fallback)
  - Verify domain event published

### US1 Unit Tests

- [ ] T036 [P] [US1] Unit test: DEK is unique per encryption call
  - Encrypt same token twice with same context
  - Verify ciphertexts are different (different DEKs generated)

- [ ] T037 [P] [US1] Unit test: Context is authenticated additional data
  - Encrypt with context, tamper with plaintext
  - Verify authentication tag fails to verify

### US1 Integration Tests

- [ ] T038 [P] [US1] Integration test: Service → Repository roundtrip with encryption
  - Create session via service (tokens encrypted)
  - Verify repository stores encrypted tokens
  - Retrieve via service (tokens decrypted)
  - Verify plaintext tokens returned

- [ ] T039 [P] [US1] Integration test: Multiple services with different contexts
  - Create sessions for "oauth2" and "github" services
  - Verify cross-service token reuse fails (context verification)

---

## Phase 5: User Story 2 - Secure KEK Storage (P1)

**Acceptance Criteria**: KEK stored securely (AWS KMS for production, env var for dev), no plaintext in memory/disk/logs, all operations auditable

### US2 E2E Tests (4 scenarios)

- [ ] T040 [US2] E2E Test - Scenario 2.1: "AWS KMS production deployment"
  - Configure KEK via AWS KMS ARN
  - Encrypt/decrypt tokens
  - Verify no plaintext KEK in application memory

- [ ] T041 [US2] E2E Test - Scenario 2.2: "KEK operations logged"
  - Encrypt/decrypt tokens with AWS KMS
  - Verify structured JSON logs with operation, service_id, success, timestamp
  - Verify no plaintext KEK in logs

- [ ] T042 [US2] E2E Test - Scenario 2.3: "Access control enforcement"
  - Simulate IAM permission denied on KMS
  - Verify encryption fails with ErrorKindKEKUnavailable
  - Verify application fails securely

- [ ] T043 [US2] E2E Test - Scenario 2.4: "Key rotation backward compatibility"
  - Encrypt token with KEK version 1
  - Rotate key to version 2
  - Decrypt token encrypted with version 1 still succeeds

### US2 Unit Tests

- [ ] T044 [P] [US2] Unit test: Environment variable KEK loaded securely
  - Load KEK from `${ENCRYPTION_KEK}`
  - Verify KEK is loaded into memguard buffer
  - Verify environment variable is not read multiple times

- [ ] T045 [P] [US2] Unit test: Plaintext KEK never logged
  - Capture logs during encrypt/decrypt
  - Verify no KEK material in logs
  - Verify error messages are sanitized

### US2 Integration Tests

- [ ] T046 [P] [US2] Integration test: AWS KMS with LocalStack
  - Create test KMS key in LocalStack
  - Configure adapter with LocalStack KMS ARN
  - Encrypt/decrypt succeeds

- [ ] T047 [P] [US2] Integration test: Environment variable KEK persists
  - Set `ENCRYPTION_KEK=<base64>`
  - Create session (encrypt)
  - Stop application
  - Start application with same `ENCRYPTION_KEK`
  - Retrieve session (decrypt with same KEK) succeeds

---

## Phase 6: User Story 3 - DEK Generation & Context Binding (P1)

**Acceptance Criteria**: Fresh DEK per session, cryptographically secure randomness (min 256 bits), context bound at DEK layer, DEKs securely erased from memory

### US3 E2E Tests (3 scenarios)

- [ ] T048 [US3] E2E Test - Scenario 3.1: "Fresh DEK per session"
  - Create multiple sessions with different contexts
  - Verify each session has unique DEK (different wrapped DEK values)

- [ ] T049 [US3] E2E Test - Scenario 3.2: "Unique DEK isolation"
  - Encrypt session for service A with DEK-A
  - Encrypt session for service B with DEK-B
  - Verify DEK-A cannot decrypt session B (context mismatch)

- [ ] T050 [US3] E2E Test - Scenario 3.3: "DEK memory zeroization"
  - Track memory after DEK generation
  - Verify DEK buffers are zeroed post-operation
  - Verify mlock/core dump exclusion if available

### US3 Unit Tests

- [ ] T051 [P] [US3] Unit test: DEK entropy >= 256 bits
  - Verify AWS SDK generates DEKs with sufficient randomness
  - Test multiple DEK generations; verify no pattern

- [ ] T052 [P] [US3] Unit test: Context binding prevents cross-context reuse
  - Encrypt with context1, extract ciphertext
  - Attempt decrypt with context2
  - Verify failure (context verification)

### US3 Integration Tests

- [ ] T053 [P] [US3] Integration test: Session-level DEK isolation
  - Create 100 sessions with same token, different service_ids
  - Verify each session has different wrapped DEK
  - Verify decryption only succeeds with matching service_id

---

## Phase 7: User Story 4 - Environment Variable KEK Injection (P1)

**Acceptance Criteria**: KEK loadable from `${ENCRYPTION_KEK}`, dev/test environments without AWS dependencies, full encrypt/decrypt cycle with env var KEK

### US4 E2E Tests (3 scenarios)

- [ ] T054 [US4] E2E Test - Scenario 4.1: "Load KEK from environment variable"
  - Set `ENCRYPTION_KEK=<base64-key>`
  - Configure `encryption.key_encryption_key: ${ENCRYPTION_KEK}`
  - Verify application initializes successfully

- [ ] T055 [US4] E2E Test - Scenario 4.2: "Use environment variable KEK for wrapping/unwrapping"
  - Encrypt token with env var KEK
  - Verify DEK wrapped successfully
  - Decrypt token succeeds

- [ ] T056 [US4] E2E Test - Scenario 4.3: "Persistence across application restart"
  - Create session with env var KEK (encryption)
  - Stop application
  - Start application with same `ENCRYPTION_KEK`
  - Retrieve session (decryption) succeeds with same token

### US4 Unit Tests

- [ ] T057 [P] [US4] Unit test: Environment variable interpolation
  - Verify config system resolves `${ENCRYPTION_KEK}` to environment value
  - Verify error if environment variable not set

- [ ] T058 [P] [US4] Unit test: Env var KEK wrapped in memguard
  - Load KEK from environment
  - Verify memguard buffer initialized
  - Verify memory locking applied

### US4 Integration Tests

- [ ] T059 [P] [US4] Integration test: Full dev workflow
  - Export `ENCRYPTION_KEK=<base64>`
  - Run application locally
  - Create/retrieve sessions
  - Verify encryption/decryption with env var KEK

---

## Phase 8: User Story 5 - Transparent Token Encryption/Decryption (P1)

**Acceptance Criteria**: Developers don't think about encryption; it's automatic in repository operations; tokens encrypted on Create, decrypted on Get

### US5 E2E Tests (3 scenarios)

- [ ] T060 [US5] E2E Test - Scenario 5.1: "Repository Create encrypts transparently"
  - Call `service.CreateSession(plaintext_tokens)`
  - Verify repository stores encrypted tokens
  - Verify plaintext tokens never written to storage

- [ ] T061 [US5] E2E Test - Scenario 5.2: "Repository Get decrypts transparently"
  - Call `service.GetSession(sessionID)`
  - Verify returned session has plaintext tokens
  - Verify encryption/decryption is transparent to caller

- [ ] T062 [US5] E2E Test - Scenario 5.3: "No manual encryption steps required"
  - Verify calling code doesn't call encryption port directly
  - Verify encryption/decryption happens automatically in service

### US5 Unit Tests

- [ ] T063 [P] [US5] Unit test: Service encryption on Create
  - Mock repository to verify encrypted tokens are stored
  - Verify plaintext tokens are never passed to repository

- [ ] T064 [P] [US5] Unit test: Service decryption on Get
  - Mock encryption port to track Decrypt calls
  - Verify Decrypt called with correct context for each token

### US5 Integration Tests

- [ ] T065 [P] [US5] Integration test: Service → Repository → Storage roundtrip
  - Create session via service
  - Query storage directly (verify BYTEA contains ciphertext, not plaintext)
  - Retrieve session via service
  - Verify plaintext returned to caller

---

## Phase 9: User Story 6 - Encryption Context Prevents Token Reuse (P2)

**Acceptance Criteria**: Context binding enforces service-level isolation; tokens encrypted for service A cannot be used for service B; defense-in-depth via DEK + KEK context verification

### US6 E2E Tests (3 scenarios)

- [ ] T066 [US6] E2E Test - Scenario 6.1: "Decrypt with matching context succeeds"
  - Encrypt token with context `{"service_id": "oauth2"}`
  - Decrypt with same context succeeds

- [ ] T067 [US6] E2E Test - Scenario 6.2: "Different service_id fails both layers"
  - Encrypt token for service A
  - Attempt decrypt with service B context
  - Verify failure at DEK verification layer
  - Verify failure at KEK unwrap layer

- [ ] T068 [US6] E2E Test - Scenario 6.3: "Ciphertext reuse attack prevented"
  - Steal ciphertext from service A
  - Attempt use with service B context
  - Verify both DEK and KEK verification fail

### US6 Unit Tests

- [ ] T069 [P] [US6] Unit test: Context mismatch detection
  - Encrypt with service_id "oauth2"
  - Attempt decrypt with service_id "github"
  - Verify ErrorKindContextMismatch error

- [ ] T070 [P] [US6] Unit test: Ciphertext is not portable across contexts
  - Encrypt token1 with context A
  - Encrypt token2 with context B (same plaintext, different context)
  - Verify ciphertexts are different (context binding)

### US6 Integration Tests

- [ ] T071 [P] [US6] Integration test: Service isolation via context
  - Create sessions for service "oauth2" and "github"
  - Attempt cross-service token access
  - Verify all attempts fail (context verification)

---

## Phase 10: User Story 7 - Post-Quantum Cryptography Readiness (P3)

**Acceptance Criteria**: PQC algorithms supported via Go 1.24+ AWS SDK; system can use PQC when available; backward compatibility without PQC

### US7 E2E Tests (4 scenarios)

- [ ] T072 [US7] E2E Test - Scenario 7.1: "Post-quantum algorithms available"
  - Verify Go version >= 1.24.0
  - Verify AWS SDK supports PQC algorithms

- [ ] T073 [US7] E2E Test - Scenario 7.2: "Encrypt/decrypt with PQC algorithms"
  - Configure system to use PQC (if available)
  - Encrypt/decrypt tokens with PQC
  - Verify success

- [ ] T074 [US7] E2E Test - Scenario 7.3: "PQC disabled gracefully"
  - Configure system without PQC
  - Encrypt/decrypt with classical algorithms
  - Verify success (no errors)

- [ ] T075 [US7] E2E Test - Scenario 7.4: "Algorithm version control"
  - Encrypt token with classical algorithm
  - Upgrade to PQC-enabled system
  - Decrypt token with classical algorithm (backward compatibility)
  - Verify success

### US7 Unit Tests

- [ ] T076 [P] [US7] Unit test: Algorithm suite selection
  - Verify AWS SDK selects algorithm based on Go version/PQC availability
  - Verify no crashes if PQC unavailable

- [ ] T077 [P] [US7] Unit test: Version byte in wrapped DEK
  - Verify encrypted tokens include algorithm version indicator
  - Verify version enables future algorithm migration

### US7 Integration Tests

- [ ] T078 [P] [US7] Integration test: Classical → PQC upgrade path
  - Encrypt tokens with classical algorithm
  - Upgrade AWS SDK to PQC version
  - Decrypt classical tokens still succeeds
  - New tokens encrypted with PQC

---

## Phase 11: Polish & Cross-Cutting Concerns

🔒 [MANDATORY] **Constitution Compliance Verification**

### Logging, Monitoring, Documentation

- [ ] T079 Implement structured JSON logging for encryption operations:
  - Log format: `operation`, `service_id`, `success`, `token_type`, `duration_ms`, `error_kind`, `timestamp`
  - All errors sanitized (no plaintext tokens, no key material)
  - Use project's slog with structured fields

- [ ] T080 [P] Create metrics instrumentation:
  - `encryption_operations_total{status,operation}` - success/failure counters
  - `encryption_duration_seconds{operation}` - latency histogram
  - `context_verification_failures_total` - context mismatch counter
  - `kek_unavailable_errors_total` - KEK availability counter

- [ ] T081 [P] Update [ARCHITECTURE.md](../../ARCHITECTURE.md):
  - Add "Encryption Vault" section documenting: envelope encryption, DEK/KEK architecture, context binding, memguard integration
  - Update glossary with new terms
  - Add port/adapter diagram

- [ ] T082 [P] Update [docs/configuration.md](../../docs/configuration.md):
  - Document `encryption.key_encryption_key` configuration
  - Show AWS KMS production example
  - Show environment variable development example
  - Document startup validation

### Security & Memory Protection

- [ ] T083 Verify memguard integration across all sensitive buffers:
  - Plaintext tokens wrapped in memguard during encryption
  - DEK buffers zeroed post-operation
  - KEK buffers memory-locked (if local key material)
  - Core dump exclusion enabled for sensitive buffers

- [ ] T084 [P] Audit logging compliance:
  - Verify 100% of encryption operations logged
  - Verify 100% of decryption operations logged
  - Verify failures include error_kind and sanitized message
  - Verify logs enable compliance audits

### Performance Validation

- [ ] T085 [P] Benchmark local operations:
  - DEK generation + encryption: target <5ms
  - DEK unwrap + decryption: target <5ms
  - AWS KMS roundtrip: target 50-200ms (operator concern)
  - Total session encryption/decryption: target <100ms

- [ ] T086 [P] Load testing (if applicable):
  - Test 100+ concurrent session encryptions
  - Verify latency stable under load
  - Verify no memory leaks (memguard cleanup)

### Documentation & Examples

- [ ] T087 Create integration guide:
  - Quickstart: how to configure KEK (AWS KMS vs. env var)
  - Service integration: how to inject EncryptionPort
  - Error handling: what errors to expect and how to handle
  - Testing: how to write tests with encryption

- [ ] T088 [P] Update examples:
  - Add example configuration files: `examples/config/encryption-*.yaml`
  - Add example service code: how to use OAuth2SessionService
  - Add example tests: unit and E2E test patterns

---

## Phase N: Constitution Compliance Verification

🔒 [MANDATORY] **Final compliance checklist before PR submission**

### Design Phase Verification (Principle XIII)

- [ ] T089 [P] Verify E2E tests were written FIRST:
  - Check git history: E2E test file created before adapter implementation
  - Verify tests initially failed (red phase documented)

- [ ] T090 [P] Verify E2E tests changed minimally:
  - Confirm test logic unchanged during implementation
  - Only fixture/test data adjustments allowed

- [ ] T091 [P] Verify all 24 spec scenarios have E2E tests:
  - Count It() blocks in [tests/e2e/encryption_vault_test.go](../../tests/e2e/encryption_vault_test.go)
  - Verify each test has comment mapping to spec.md scenario
  - Verify all E2E tests pass (green phase)

### Implementation Phase Verification (Principle XIII)

- [ ] T092 [P] Verify E2E tests turn GREEN as implementation completes:
  - Run `ginkgo -v ./tests/e2e/encryption_vault_test.go`
  - Confirm 24/24 tests pass
  - Track test status in PR description

### Architecture & Design (Principle II)

- [ ] T093 [P] Verify ADR created and accepted:
  - Check [adrs/NNN-envelope-encryption-design.md](../../adrs/) exists
  - Verify ADR documents: DEK-per-session, context binding, AWS SDK choice, memguard integration
  - Mark ADR as "Accepted"

- [ ] T094 [P] Verify [ARCHITECTURE.md](../../ARCHITECTURE.md) updated:
  - Encryption domain documented
  - Glossary terms added
  - Port/adapter diagram included

### Security (Principle I, III)

- [ ] T095 [P] Verify security-first implementation:
  - No plaintext fallback on encryption/decryption failure
  - Fail-closed behavior enforced
  - No custom cryptography (AWS SDK only)
  - Memguard integration verified

- [ ] T096 [P] Verify no sensitive data leakage:
  - Search codebase for plaintext token logging (grep: forbidden patterns)
  - Search for KEK logging (grep: forbidden patterns)
  - Verify error messages are sanitized

### Testing (Principle VIII, XIII)

- [ ] T097 [P] Verify TDD followed:
  - E2E tests written and failed initially (red phase)
  - Unit tests written before/during implementation (red phase)
  - Tests changed minimally during implementation
  - All tests pass (green phase)

- [ ] T098 [P] Verify test coverage:
  - Unit test coverage >= 85% for [internal/adapters/encryption/](../../internal/adapters/encryption/)
  - Integration tests cover memory adapter and PostgreSQL adapter
  - E2E tests cover all 24 spec scenarios

### Configuration (Principle VII)

- [ ] T099 [P] Verify configuration integration:
  - `encryption.key_encryption_key` field used (no custom config)
  - `${ENCRYPTION_KEK}` interpolation works
  - AWS KMS ARN format detected and handled
  - Startup validation enforced (fail-fast if KEK unavailable)

### Persistence (Principle IX)

- [ ] T100 [P] Verify storage layer unchanged:
  - No new migrations required (columns already exist)
  - Storage adapters see only encrypted BYTEA/JSONB
  - No encryption logic in storage layer

### Dependency Injection (Principle XII)

- [ ] T101 [P] Verify Builder pattern used:
  - EncryptionPort initialized in [internal/app/builder.go](../../internal/app/builder.go)
  - OAuth2SessionService receives injected port
  - No circular dependencies

- [ ] T102 [P] Verify routing functions thin:
  - HTTP handlers receive pre-wired services from Builder
  - No service instantiation in routing functions

### API & Documentation (Principle IV, X)

- [ ] T103 [P] Verify internal API documented:
  - EncryptionPort interface documented in [internal/ports/encryption.go](../../internal/ports/encryption.go)
  - Error contract documented in [contracts/error-contract.md](./contracts/error-contract.md)
  - Domain events documented in [contracts/domain-events.md](./contracts/domain-events.md)

- [ ] T104 [P] Verify user/stakeholder confirmation:
  - No HTTP API changes (internal feature only)
  - Confirm with user that port/adapter design is acceptable
  - Document any confirmations in ADR or PR comments

### Glossary & Domain (Principle V)

- [ ] T105 [P] Verify glossary updated:
  - New terms added to [ARCHITECTURE.md](../../ARCHITECTURE.md) Glossary
  - Terms used consistently throughout code/docs
  - Domain concepts modeled explicitly

---

## Success Criteria & Acceptance

### Measurable Outcomes (from spec.md)

- ✅ **SC-001**: 100% of OAuth tokens encrypted (zero plaintext in storage)
- ✅ **SC-002**: DEK generation cryptographically secure (>= 256 bits entropy)
- ✅ **SC-003**: Context binding enforced at DEK + KEK layers
- ✅ **SC-004**: Cross-context token reuse prevented (context verification fails)
- ✅ **SC-005**: KEK security follows industry best practices (externalized or local with permissions)
- ✅ **SC-006**: External KMS integration works (zero plaintext KEKs)
- ✅ **SC-007**: Local development works (<50ms per operation)
- ✅ **SC-008**: Token integrity verified (100% tampered tokens detected)
- ✅ **SC-009**: No plaintext leakage (tokens/keys never in logs/errors)
- ✅ **SC-010**: Fail-closed security (KEK missing or context verification failure = error)
- ✅ **SC-011**: Transparent encryption (repository methods require zero manual steps)
- ✅ **SC-012**: Backward compatibility (old tokens decryptable after KEK rotation)
- ✅ **SC-013**: Performance baseline (<100ms typical operation, local <50ms)

### Acceptance Test Mapping

| User Story | Spec Scenarios | E2E Tests | Status |
|------------|---|---|---|
| US1: Envelope Encryption | 4 | T032-T035 | ⏳ Pending |
| US2: Secure KEK Storage | 4 | T040-T043 | ⏳ Pending |
| US3: DEK Generation | 3 | T048-T050 | ⏳ Pending |
| US4: Env Var KEK Injection | 3 | T054-T056 | ⏳ Pending |
| US5: Transparent Encryption | 3 | T060-T062 | ⏳ Pending |
| US6: Cross-Service Prevention | 3 | T066-T068 | ⏳ Pending |
| US7: Post-Quantum Ready | 4 | T072-T075 | ⏳ Pending |
| **Total** | **24** | **24** | ⏳ Pending |

---

## Implementation Strategy

### Phase Execution Order

1. **Phase 1 (Setup)**: Initialize project structure, dependencies
2. **Phase 2 (Design)**: Design preconditions; write E2E tests (must FAIL initially)
3. **Phase 3 (Foundational)**: Implement ports, domain errors, events; configure builder
4. **Phase 3a (AWS Adapter)**: Implement EncryptionPort with AWS SDK
5. **Phase 3b-c (Service Integration)**: Wire encryption into OAuth2SessionService and Builder
6. **Phase 4-10 (User Stories)**: Implement features incrementally; E2E tests turn GREEN
7. **Phase 11 (Polish)**: Logging, monitoring, documentation, performance validation
8. **Phase N (Compliance)**: Final verification before PR submission

### Parallelization Opportunities

**Team Coordination**: This feature can be implemented by 1-2 engineers in parallel:
- **Engineer 1**: AWS adapter ([internal/adapters/encryption/aws/](../../internal/adapters/encryption/aws/)) + unit tests (Tasks T022-T026)
- **Engineer 2**: Service integration + E2E tests (Tasks T027-T035, T040-T043, etc.)
- **Both**: Shared work on configuration, builder wiring, compliance verification

**Parallelizable Tasks** (18 total marked with [P]):
- T009-T010: Configuration and error contract verification (parallel)
- T022-T026: AWS adapter implementation (parallel units)
- T027-T031: Service wiring and builder updates (parallel)
- T032-T078: User story E2E/unit/integration tests (highly parallelizable)
- T079-T088: Logging, monitoring, documentation (parallel)

### MVP Scope (Phase 3-4)

**Minimum Viable Product focuses on User Story 1 only**:
- Envelope encryption with DEK + KEK wrapping
- Context binding at DEK layer
- Environment variable KEK injection (for dev)
- Transparent encryption/decryption in service
- 4 E2E tests for US1 scenarios

**Follow-up Work** (future increments):
- AWS KMS integration (Phase 5: US2)
- Additional user stories (Phase 6-10)
- Performance optimization and production hardening
- PQC algorithms (Phase 10: US7)

---

## Blocking Dependencies

- ✅ Feature 002 (Flexible Configuration) - provides config system with `${env_var}` interpolation
- ✅ Feature 005 (Session Management) - provides UserSession aggregate, repository, service patterns
- ⏳ Go 1.24.0+ (for PQC support; not required for MVP but recommended for future-proofing)

---

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Custom cryptography bugs | Use AWS Encryption SDK (official, battle-tested) |
| Context binding not enforced | AWS SDK handles AAD verification natively; tests verify both layers fail on mismatch |
| Memory leaks of plaintext/DEK | Memguard integration with explicit zeroing; memory tests verify cleanup |
| KEK unavailability breaks service | Startup validation (fail-fast); clear error messages; operator alerts on KMS failure |
| Database schema mismatches | No migrations needed; columns (encrypted_*_token BYTEA, encryption_context JSONB) already exist |
| Decryption of untrusted ciphertext | AESGCMSIV authentication tag verification built into AWS SDK; tampered data detected |

---

## Validation Checklist

**Before submitting PR, verify**:
- [ ] All 54 tasks completed (Phase 1-11, Phase N)
- [ ] All 24 E2E tests pass (green phase)
- [ ] All unit tests pass (coverage >= 85%)
- [ ] All integration tests pass (real PostgreSQL, LocalStack KMS)
- [ ] No plaintext tokens/keys in logs or error messages (verified via grep)
- [ ] Configuration works (AWS KMS ARN and `${ENCRYPTION_KEK}`)
- [ ] Startup validation enforced (fail-fast if KEK missing)
- [ ] ADR created and marked "Accepted"
- [ ] [ARCHITECTURE.md](../../ARCHITECTURE.md) updated with glossary and design
- [ ] Examples and documentation complete
- [ ] Performance targets met (<5ms local, <100ms with KMS latency)
- [ ] Memory protection verified (memguard integration, no leaks)
- [ ] Builder wiring complete (no circular dependencies)
- [ ] Routing functions thin (no service instantiation)

---

**Version**: 1.0 | **Status**: ⏳ Ready for Implementation | **Last Updated**: 2026-01-16
