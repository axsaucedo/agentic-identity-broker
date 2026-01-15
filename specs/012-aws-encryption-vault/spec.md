# Feature Specification: Encryption Vault for OAuth Tokens (Envelope Encryption)

**Feature Branch**: `012-aws-encryption-vault`
**Created**: 2026-01-14
**Status**: Draft
**Input**: User description: "Standard AWS Encryption Vault (PQC-Enabled by Go 1.24+)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Encrypt OAuth Tokens Using Envelope Encryption (Priority: P1)

A platform operator deploying the system to production needs to encrypt sensitive OAuth tokens (access tokens, refresh tokens) stored in the sessions table using envelope encryption. Tokens should be encrypted at the application layer using a Data Encryption Key (DEK) before storage, with the DEK itself encrypted by a Key Encryption Key (KEK), providing two layers of protection. The encryption context (principal, service_id, session_id, purpose) must be bound at the DEK level, ensuring tokens are tied to their specific usage context.

**Why this priority**: This is the core use case—protecting sensitive tokens at rest in the database using envelope encryption. Envelope encryption with DEK-level context binding provides defense in depth with the DEK and KEK managed separately.

**Independent Test**: Can be fully tested by encrypting an OAuth token with envelope encryption (DEK with context + wrapped KEK), storing it in the sessions table, retrieving it, decrypting it with the correct KEK and context, and verifying the token is correctly recovered.

**Acceptance Scenarios**:

1. **Given** a session with OAuth tokens and associated context (principal, service_id, session_id, purpose) is stored, **When** tokens are saved to the sessions table, **Then** each token is encrypted using a DEK bound to its context, the DEK is encrypted using a KEK with the same context, and both the token ciphertext and wrapped DEK are stored
2. **Given** encrypted tokens and wrapped DEKs in the sessions table, **When** a session is retrieved and tokens are requested, **Then** the KEK is used to unwrap the DEK (verifying context), the DEK is used to decrypt the token (verifying context), and the plaintext token is returned
3. **Given** an attacker gains direct database access and reads encrypted tokens with wrapped DEKs, **When** they attempt to use the encrypted tokens with different context, **Then** the tokens are invalid because context verification fails at both DEK and KEK layers
4. **Given** a session is loaded from the database, **When** context verification fails at any layer, **Then** the application fails securely and does not fall back to plaintext tokens

---

### User Story 2 - Secure KEK Storage Using Industry Best Practices (Priority: P1)

A security team needs the Key Encryption Key (KEK) to be securely stored using industry-standard key management practices. For cloud deployments, KEKs should be managed by a centralized, enterprise-grade key management service. For development, KEKs may. be stored locally with appropriate protections. Under no circumstances should KEKs exist in plaintext in application memory or on disk.

**Why this priority**: KEK security is fundamental to envelope encryption. Following industry best practices for KEK storage ensures the security of the entire token encryption system.

**Independent Test**: Can be fully tested by configuring KEK storage, encrypting/decrypting tokens, and verifying that plaintext KEKs never appear in logs, memory, or on disk, and that all KEK operations are auditeable.

**Acceptance Scenarios**:

1. **Given** a production deployment with centralized key management configured, **When** a token is encrypted, **Then** the KEK is obtained from the key management service, used to wrap the DEK with context verification, and never stored in application memory or on disk
2. **Given** KEK access is configured with centralized key management, **When** tokens are decrypted, **Then** the key management service is called to unwrap the DEK (verifying context), and all KEK operations are logged by the key management service
3. **Given** access controls on the centralized key management service, **When** an unauthorized user's session attempts to decrypt tokens, **Then** the key management service denies KEK access and the application fails securely
4. **Given** keys are rotated in the centralized key management service, **When** tokens encrypted with a previous KEK version need decryption, **Then** the key management service can still decrypt them with the previous key version

---

### User Story 3 - DEK Generation and Context Binding (Priority: P1)

A security architect needs Data Encryption Keys (DEK) to be generated securely and bound to the token's usage context. Each session (containing access and refresh tokens) should use a fresh DEK generated with cryptographically secure randomness, with the context (principal, service_id, session_id, purpose) cryptographically bound during DEK encryption. DEKs should be unique per session and tightly coupled with their encrypted tokens and context.

**Why this priority**: DEK security and context binding are crucial for envelope encryption. Generating fresh, random DEKs per session with context binding ensures that tokens cannot be reused in different contexts, and compromise of one session's DEK doesn't affect other sessions.

**Independent Test**: Can be fully tested by encrypting multiple sessions with different contexts, verifying that each session has a unique DEK, and confirming that tokens encrypted in one context cannot be decrypted with different context values.

**Acceptance Scenarios**:

1. **Given** multiple sessions are encrypted with different contexts, **When** each session is created, **Then** a fresh DEK is generated for the session with cryptographically secure randomness, and the context is bound to the DEK encryption
2. **Given** two sessions with different contexts (principal/service/session_id), **When** both are stored, **Then** each session has a unique DEK with its context bound, and attempting to decrypt one session's DEK with another's context fails
3. **Given** DEKs are generated for session encryption, **When** all operations complete, **Then** DEKs used for encryption are securely erased from memory and never appear in logs

---

### User Story 4 - Local File-Based Key Storage for Development (Priority: P1)

A developer in a local development or test environment needs to use locally-stored keys for envelope encryption without requiring cloud key management services. Both DEK and KEK operations should use local key material securely stored in files with appropriate protections.

**Why this priority**: Supporting local key storage is essential for development flexibility. Developers and test environments cannot always access cloud services, so local key storage enables full-stack development.

**Independent Test**: Can be fully tested by storing KEKs in local files with restricted permissions, encrypting/decrypting tokens with context, and verifying tokens persist correctly across application restarts.

**Acceptance Scenarios**:

1. **Given** the encryption backend is configured to use local file-based keys, **When** the application starts, **Then** KEK material is loaded from the configured file path
2. **Given** local key files exist with restricted file permissions, **When** envelope encryption is used, **Then** only the application process can read the key material
3. **Given** a token is encrypted and decrypted with local file-based keys and context, **When** the application restarts, **Then** the same local KEK can still decrypt previously encrypted tokens with the same context

---

### User Story 4.5 - KEK from Environment Variables (Priority: P1)

A developer or operator in a containerized environment needs to inject KEK material via environment variables instead of file paths. This enables easier deployment in container orchestration platforms where secrets are managed through environment variables or secret management systems that inject via environment.

**Why this priority**: Environment variable support is essential for containerized deployments. Container platforms (Docker, Kubernetes) commonly use environment variables for secret injection, so supporting this pattern enables seamless integration with standard deployment practices.

**Independent Test**: Can be fully tested by providing KEK via environment variable, encrypting/decrypting tokens, and verifying the KEK from environment is used correctly.

**Acceptance Scenarios**:

1. **Given** the encryption backend is configured to use environment variables, **When** the KEK environment variable is set before application startup, **Then** KEK material is loaded from the environment variable
2. **Given** KEK material is provided via environment variable, **When** tokens are encrypted and decrypted, **Then** the environment variable KEK is used correctly for both DEK wrapping and unwrapping
3. **Given** the application is deployed with KEK in an environment variable, **When** the application restarts, **Then** tokens encrypted with previous environment KEK can still be decrypted if the environment variable is set to the same value

---

### User Story 5 - Transparent Token Encryption/Decryption in Repository (Priority: P1)

A developer using the session repository needs token envelope encryption to be transparent and automatic. When storing sessions with tokens and their context, DEK generation and encryption should happen automatically without manual intervention. When retrieving sessions, DEK unwrapping and token decryption should happen automatically with context verification.

**Why this priority**: Transparency is critical for developer experience. The encryption vault should be integrated into the session repository so developers don't need to think about encryption/decryption at the application level.

**Independent Test**: Can be fully tested by calling the repository's `Create` and `Get` methods with context data, verifying that tokens are encrypted with envelope encryption, and that the returned session has plaintext tokens.

**Acceptance Scenarios**:

1. **Given** a session with tokens and context (principal, service_id, session_id, purpose) is passed to the repository's `Create` method, **When** the session is stored, **Then** tokens are encrypted using envelope encryption with context binding and plaintext tokens are not stored
2. **Given** a session is retrieved with the repository's `Get` method, **When** the session is returned, **Then** tokens are automatically decrypted using envelope encryption with context verification and available as plaintext
3. **Given** the session repository is used normally, **When** envelope encryption/decryption happens, **Then** no manual encryption steps are required by calling code

---

### User Story 6 - Encryption Context Prevents Token Reuse Across Contexts (Priority: P2)

A security architect needs tokens to be bound to their specific usage context (principal, service_id, session_id, purpose) at the cryptographic level. Tokens encrypted in one context must not be usable in a different context, even if an attacker has access to ciphertext from both contexts.

**Why this priority**: Encryption context adds defense in depth by binding tokens to specific contexts. While not essential for MVP, it's an important security enhancement that prevents cross-context token reuse attacks.

**Independent Test**: Can be fully tested by encrypting tokens with specific context values, attempting to decrypt with different context values, and verifying decryption fails at both the DEK and KEK verification layers.

**Acceptance Scenarios**:

1. **Given** a token is encrypted with encryption context `{"principal": "user@example.com", "service_id": "oauth2", "session_id": "ABC123", "purpose": "access"}`, **When** decryption is attempted with matching context, **Then** decryption succeeds
2. **Given** a token encrypted with specific context, **When** decryption is attempted with different principal, service_id, session_id, or purpose values, **Then** decryption fails at both DEK verification and KEK unwrapping
3. **Given** tokens with different contexts stored in the database, **When** an attacker tries to use ciphertext from one context with a different context, **Then** both DEK decryption and KEK unwrapping fail, and the attack is prevented

---

### User Story 7 - Post-Quantum Cryptography Readiness (Priority: P3)

A forward-thinking security team wants the encryption vault to support post-quantum cryptographic algorithms for both DEK and KEK when available. This future-proofs token encryption against potential quantum computing threats.

**Why this priority**: Post-quantum cryptography is future-oriented. While quantum threats are not immediate, supporting PQC algorithms now ensures long-term protection of tokens. This can be deferred to follow-up work if needed.

**Independent Test**: Can be fully tested by enabling post-quantum cryptography, encrypting tokens with post-quantum algorithms, and verifying encryption/decryption work correctly.

**Acceptance Scenarios**:

1. **Given** post-quantum cryptography is available and enabled, **When** tokens are encrypted with envelope encryption, **Then** both DEK encryption and KEK wrapping use post-quantum algorithms
2. **Given** tokens encrypted with post-quantum envelope encryption, **When** decryption is performed, **Then** decryption succeeds without requiring special algorithm handling
3. **Given** systems without post-quantum cryptography support, **When** the system runs with PQC disabled, **Then** tokens are encrypted with classical algorithms without errors

---

### Edge Cases

- What happens if a token is too large to encrypt efficiently with envelope encryption?
- How does the system handle DEK generation failures during session creation?
- What happens if KEK access fails (e.g., key management service unavailable) during session retrieval?
- How does the system behave if local key file permissions are too permissive (world-readable)?
- What happens when decrypting a token that was wrapped with a different KEK version?
- How does the system recover if context verification fails for some tokens but not others?
- What happens if a token is encrypted with one context and someone tries to decrypt it with a completely different context?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST use envelope encryption to encrypt OAuth tokens: a unique Data Encryption Key (DEK) encrypts the token with context binding, and the DEK is wrapped (encrypted) by a Key Encryption Key (KEK) with the same context
- **FR-002**: System MUST generate a fresh DEK for each token encryption with cryptographically secure randomness (minimum 256 bits entropy)
- **FR-003**: System MUST bind encryption context (principal, service_id, session_id, purpose) to the DEK encryption as authenticated additional data
- **FR-004**: System MUST bind the same encryption context to the KEK wrapping operation (DEK wrapping with authenticated encryption)
- **FR-005**: System MUST wrap the DEK using the KEK with context verification and store the wrapped DEK alongside the encrypted token
- **FR-006**: System MUST decrypt tokens by first unwrapping the DEK using the KEK (verifying context), then using the DEK to decrypt the token (verifying context)
- **FR-007**: System MUST reject token decryption if context verification fails at either the DEK layer or the KEK layer
- **FR-008**: System MUST support configurable KEK storage backends: externalized key management service and local file-based storage
- **FR-009**: System MUST automatically decrypt OAuth tokens when retrieving sessions, performing DEK unwrapping and token decryption transparently with context verification
- **FR-010**: System MUST encode encrypted tokens in base64 for safe storage in database columns
- **FR-011**: System MUST validate KEK is accessible before the application starts (fail-fast on startup if KEK unavailable)
- **FR-012**: System MUST fail securely if KEK is missing, inaccessible, or unwrapping fails (no plaintext fallback)
- **FR-013**: System MUST never log plaintext OAuth tokens or encryption keys (DEK or KEK)
- **FR-014**: System MUST verify integrity of encrypted tokens during decryption; if integrity verification fails, decryption MUST fail
- **FR-015**: System MUST support multiple token types (access tokens, refresh tokens, id tokens)
- **FR-016**: System MUST support post-quantum cryptographic algorithms for both DEK encryption and KEK wrapping when available
- **FR-017**: System MUST protect against chosen-ciphertext attacks during envelope encryption and decryption

### Domain Model

**Entities** (things with unique identity):

- **EncryptedOAuthToken**: Represents an encrypted OAuth token with metadata: encrypted token data (ciphertext), wrapped DEK, encryption context, token type, algorithm identifier, timestamp
- **EncryptionConfig**: Represents envelope encryption configuration: KEK backend type, backend-specific parameters

**Aggregates** (consistency boundaries):

- **OAuthTokenVault**: Root aggregate managing envelope encryption/decryption. Manages DEK generation, DEK encryption with context, DEK wrapping with KEK and context, and transparent integration at the repository layer

**Value Objects** (things without identity):

- **EncryptionContext**: Map containing exactly four token metadata fields: principal (user identity), service_id (OAuth service), session_id (session identifier), purpose (token usage type). Immutable and bound to both DEK encryption and KEK wrapping
- **WrappedEnvelope**: Represents the complete encrypted output: token ciphertext with DEK context verification, wrapped (encrypted) DEK with KEK context verification, and authentication metadata

**Domain Events** (state changes of business significance):

- **TokenEncrypted**: OAuth token encrypted with envelope encryption and context binding, ready for storage
- **TokenDecrypted**: OAuth token decrypted from envelope with context verified, ready for use
- **TokenEncryptionFailed**: Envelope encryption failed (DEK generation, context binding, or KEK wrapping); session storage should be rejected
- **TokenDecryptionFailed**: Envelope decryption failed (context verification or DEK unwrapping or token decryption); session retrieval should fail securely
- **ContextVerificationFailed**: Encryption context verification failed during encryption or decryption, indicating context mismatch or tampering

### Configuration Requirements

**Configuration Parameters**:

The system supports two KEK storage mechanisms:

1. **Externalized key management** - for production deployments with centralized key management, audit logging, and access controls
2. **Environment-injected KEK** - for development and containerized deployments where KEK material is provided via environment variables

Configuration is managed via environment variables leveraging the existing `.env` configuration approach:
- If externalized key management is configured, the system uses it
- Otherwise, the system uses KEK material from the ENCRYPTION_KEK environment variable
- At most one mechanism should be active for any deployment

**Configuration Location**: Encryption configuration is integrated into the existing `.env` configuration approach from feature 002-flexible-configuration

### API Requirements

- **API-001**: Encryption port interface MUST define `Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error)` performing envelope encryption with DEK (binding context) + KEK wrapping (binding context)
- **API-002**: Encryption port interface MUST define `Decrypt(ctx context.Context, ciphertext []byte, encryptionContext map[string]string) ([]byte, error)` performing envelope decryption with DEK unwrapping (verifying context) and token decryption (verifying context)
- **API-003**: EncryptionContext parameter MUST contain exactly four keys: principal, service_id, session_id, purpose
- **API-004**: Session repository MUST accept configured encryption backend and transparently perform envelope encryption/decryption on tokens with context binding from session data
- **API-005**: Configuration system MUST support encryption settings via `config.encryption.*` namespace
- **API-006**: All encryption/decryption operations MUST be context-aware and respect context deadlines

### Security Requirements

#### Cryptographic Requirements

- **SR-001**: Encryption MUST use envelope encryption with two key layers: DEK for token encryption and KEK for DEK wrapping
- **SR-002**: DEK MUST be unique for each token, generated with cryptographically secure randomness (minimum 256 bits entropy)
- **SR-003**: Encryption MUST use authenticated encryption (AEAD ciphers) ensuring both confidentiality and integrity of tokens
- **SR-004**: Encryption MUST prevent chosen-ciphertext attacks through authenticated encryption and commitment policies
- **SR-005**: DEK encryption MUST bind encryption context (principal, service_id, session_id, purpose) as authenticated additional data (AAD)
- **SR-006**: DEK wrapping MUST bind the same encryption context as authenticated additional data to the KEK wrapping operation
- **SR-007**: Context verification MUST fail the entire decryption if context does not match at either DEK or KEK layer
- **SR-008**: Post-quantum cryptography algorithms (if enabled) MUST use NIST-approved or IETF-standardized algorithms

#### Key Management Requirements

- **SR-009**: KEK MUST be securely stored following industry best practices: externalized in centralized key management service for production or locally with restricted file permissions for development
- **SR-010**: KEK MUST NEVER exist in plaintext in application memory during normal operation (only in use for wrapping/unwrapping operations)
- **SR-011**: KEK MUST NEVER be stored in plaintext in logs, configuration files, or debug output
- **SR-012**: External key management service (when configured) MUST support audit logging of all KEK operations (wrapping and unwrapping)
- **SR-013**: Local KEK files MUST be created with restrictive permissions (0600) readable only by the application process
- **SR-014**: System MUST validate KEK accessibility at startup and fail fast if KEK cannot be accessed

#### Buffer and Memory Requirements

- **SR-015**: Plaintext OAuth tokens MUST be zeroed from memory after use via secure buffer handling (defer statements or equivalent)
- **SR-016**: DEK material MUST be zeroed from memory after use (after DEK wrapping during encryption, after DEK unwrapping during decryption)
- **SR-017**: Plaintext DEK MUST NOT persist in memory across multiple operations
- **SR-018**: KEK and DEK material MUST be protected against swapping to disk (memory locking) when possible
- **SR-019**: Core dumps MUST NOT contain plaintext KEK or DEK material (enable core dump exclusion for sensitive memory regions when supported)
- **SR-020**: All sensitive buffers (plaintext tokens, DEK, KEK) MUST be handled with memory protection best practices (no copying to temporary buffers, no intermediate allocations)

#### Operational Security Requirements

- **SR-021**: All token encryption/decryption operations MUST be auditable: successes logged with operation type and context fields, failures logged with sufficient context for debugging
- **SR-022**: OAuth tokens MUST NOT be logged as plaintext in any logs or error messages
- **SR-023**: Encryption keys (DEK and KEK) MUST NOT be logged in any logs or error messages
- **SR-024**: System MUST NOT silently fall back to plaintext token storage if envelope encryption fails
- **SR-025**: System MUST NOT silently fall back to plaintext tokens if decryption fails
- **SR-026**: Decryption failures MUST result in application-level failures, not silent acceptance of corrupted data
- **SR-027**: Context verification failures MUST cause immediate decryption failure with no fallback or retry
- **SR-028**: All DEK and KEK operations MUST include context verification as integral part of cryptographic operations

### Performance Requirements

- **PR-001**: Encryption of a typical session (both access and refresh tokens) using envelope encryption with one DEK and context binding SHOULD complete in under 100ms (excluding external key management service latency)
- **PR-002**: Decryption of a typical session (both access and refresh tokens) using envelope encryption with context verification SHOULD complete in under 100ms (excluding external key management service latency)
- **PR-003**: System MUST handle encryption and decryption operations without blocking the session repository under normal load

### Key Entities

- **EncryptedOAuthToken**: Encapsulates encrypted token envelope with wrapped DEK and context
- **EncryptionConfig**: Manages KEK storage backend configuration
- **OAuthTokenVault**: Core aggregate managing DEK generation, DEK encryption with context binding, and envelope encryption/decryption

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Envelope encryption is implemented—100% of OAuth tokens are encrypted using DEK + wrapped KEK with context binding, 0% plaintext
- **SC-002**: DEK generation is cryptographically secure—each token has a unique DEK generated with minimum 256 bits entropy
- **SC-003**: Context binding is enforced—encryption context (principal, service_id, session_id, purpose) is bound to both DEK encryption and KEK wrapping
- **SC-004**: Context verification prevents cross-context reuse—tokens encrypted in one context cannot be decrypted in a different context
- **SC-005**: KEK security follows industry best practices—KEKs are securely stored (externalized or locally with 0600 permissions), never in plaintext
- **SC-006**: External key management works—tokens encrypted/decrypted via external KMS with zero plaintext KEKs on disk
- **SC-007**: Local development works—developers can encrypt/decrypt tokens with local KEK files in under 50ms per operation
- **SC-008**: Token integrity is verified—100% of tampered encrypted tokens are detected during decryption
- **SC-009**: No plaintext leakage—OAuth tokens, DEKs, and KEKs never appear in logs, errors, or debug output
- **SC-010**: Security fail-closed—if KEK is missing or context verification fails, application fails with clear error messages
- **SC-011**: Transparent encryption—session repository envelope encryption/decryption requires zero manual steps
- **SC-012**: Backward compatibility—tokens encrypted with previous algorithm versions can be decrypted after updates
- **SC-013**: Performance baseline—envelope encryption/decryption of typical tokens with context binding completes in under 100ms

## Assumptions

- OAuth token size is typically 1-10 KB (standard OAuth2 token sizes with JWTs)
- Encryption context fields (principal, service_id, session_id, purpose) are available from session data at encryption and decryption time
- External key management service is available for production deployments and supports key versioning/rotation
- Cryptographic library supports authenticated encryption ciphers with AAD (Additional Authenticated Data) and will support post-quantum algorithms in future versions
- Local key material is stored on secure, non-networked storage
- Session repository from feature 005-session-management is available and can be extended with envelope encryption
- Operators manage KEKs according to organizational security policies and best practices
- DEK context binding is enforced as integral part of the encryption/decryption algorithm, not as separate validation step

## Notes

This specification uses **Envelope Encryption** with two key layers and context binding at the DEK level: Data Encryption Keys (DEK) that encrypt individual tokens with encryption context as authenticated additional data, and Key Encryption Keys (KEK) that wrap (encrypt) the DEKs with the same context as authenticated additional data. Each token encryption generates a fresh, random DEK, binds context during encryption, and the DEK is wrapped by the KEK with context verification. Context verification happens at both layers: DEK decryption verifies context, and DEK unwrapping verifies context. Mismatch at either layer causes decryption to fail immediately.

The KEK is securely stored following industry best practices: externalized in a centralized key management service (for production) with audit logging and access controls, or locally with restricted file permissions (for development). The system explicitly forbids plaintext fallback and requires fail-closed behavior if KEK access or context verification fails at any point.

Encryption context consists of exactly four fields from the sessions table: principal (user identity), service_id (OAuth service identifier), session_id (session identifier), and purpose (token usage type). These fields bind each token to its specific usage context, preventing token reuse across contexts.

Implementation will determine specific cryptographic algorithms, authentication modes, and key sizes during the planning phase based on the project's cryptographic library and compliance requirements.

---

**Version**: 1.0 | **Status**: Draft | **Last Updated**: 2026-01-14
