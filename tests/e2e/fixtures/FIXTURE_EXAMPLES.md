# Example Fixture Outputs

This document shows actual outputs from fixture functions to demonstrate the test data they generate.

## Principal Fixtures

### DefaultPrincipal()

```go
principal := fixtures.DefaultPrincipal()
// Output:
// Principal {
//   Email: "user@example.com"
// }
// principal.String() = "user@example.com"
```

**Use Case**: Default test user for authenticated requests
```go
resp, err := server.AuthenticatedGET("/oauth2/authorize?...", principal, nil)
// X-Remote-User header = "user@example.com"
```

### AnotherPrincipal()

```go
principal := fixtures.AnotherPrincipal()
// Output:
// Principal {
//   Email: "another-user@example.com"
// }
// principal.String() = "another-user@example.com"
```

**Use Case**: Testing with different principals
```go
// Different users can have different grants
grant1 := fixtures.ActiveGrant(fixtures.DefaultPrincipal().String(), agentID)
grant2 := fixtures.ActiveGrant(fixtures.AnotherPrincipal().String(), agentID)
// user@example.com and another-user@example.com have different grants
```

### AdminPrincipal()

```go
principal := fixtures.AdminPrincipal()
// Output:
// Principal {
//   Email: "admin@example.com"
// }
// principal.String() = "admin@example.com"
```

**Use Case**: Testing with administrative user
```go
resp, err := server.AuthenticatedGET("/oauth2/authorize?...", principal, nil)
// X-Remote-User header = "admin@example.com"
```

## Agent Fixtures

### ValidAgent()

```go
agent := fixtures.ValidAgent()
// Output:
// Agent {
//   ID:          "550e8400-e29b-41d4-a716-446655440000"  // Generated UUID (fresh each call)
//   ClientID:    "test-client-valid"
//   DisplayName: "Test Agent Valid"
//   Description: "A valid test agent for E2E testing with all required fields"
//   CreatedAt:   2025-01-05T18:40:00Z
//   UpdatedAt:   2025-01-05T18:40:00Z
// }
```

**Domain Validation**: ✓ Passes (DisplayName and Description are required)

**Use Case**: Basic authorization flow test
```go
agent := fixtures.ValidAgent()
storage.Agents().Create(ctx, agent)
resp, _ := server.AuthenticatedGET(
    "/oauth2/authorize?client_id="+agent.ClientID+"&...",
    fixtures.DefaultPrincipal(),
    nil,
)
```

### AnotherAgent()

```go
agent := fixtures.AnotherAgent()
// Output:
// Agent {
//   ID:          "6610e501-f30c-42e5-b817-557766551111"  // Different UUID (fresh each call)
//   ClientID:    "test-client-another"
//   DisplayName: "Test Agent Another"
//   Description: "Another valid test agent for E2E testing with different client ID"
//   CreatedAt:   2025-01-05T18:40:00Z
//   UpdatedAt:   2025-01-05T18:40:00Z
// }
```

**Use Case**: Testing multiple agents
```go
agent1 := fixtures.ValidAgent()
agent2 := fixtures.AnotherAgent()
// Different ClientIDs for testing client routing
```

### AgentWithClientID("custom-client")

```go
agent := fixtures.AgentWithClientID("custom-client")
// Output:
// Agent {
//   ID:          "7721f612-g41d-53f6-c928-668877662222"  // Fresh UUID
//   ClientID:    "custom-client"
//   DisplayName: "Test Agent custom-client"
//   Description: "Test agent with custom client ID: custom-client"
//   CreatedAt:   2025-01-05T18:40:00Z
//   UpdatedAt:   2025-01-05T18:40:00Z
// }
```

**Use Case**: Testing with known client IDs
```go
agent := fixtures.AgentWithClientID("my-app-client-id")
// ClientID predictable for test assertions
```

### AgentWithURLs()

```go
agent := fixtures.AgentWithURLs()
// Output:
// Agent {
//   ID:                   "8832g723-h52e-64g7-d039-779988773333"
//   ClientID:             "test-client-with-urls"
//   DisplayName:          "Agent With URLs"
//   Description:          "Test agent with all URL fields populated for documentation and governance"
//   GovernanceURL:        &"https://governance.example.com/agent"
//   UserDocumentationURL: &"https://docs.example.com/user-guide"
//   AgentInterfaceURL:    &"https://agent.example.com"
//   CreatedAt:            2025-01-05T18:40:00Z
//   UpdatedAt:            2025-01-05T18:40:00Z
// }
```

**Domain Validation**: ✓ Passes (URLs validated as HTTP/HTTPS)

**Use Case**: Testing agent details display
```go
agent := fixtures.AgentWithURLs()
// URLs available for rendering in consent UI
```

## Grant Fixtures

### ActiveGrant("user@example.com", agentID)

```go
grant := fixtures.ActiveGrant("user@example.com", "550e8400-e29b-41d4-a716-446655440000")
// Output:
// UserGrant {
//   ID:         "9943h834-i63f-75h8-e140-880099884444"  // Fresh UUID
//   Principal:  "user@example.com"
//   AgentID:    "550e8400-e29b-41d4-a716-446655440000"
//   ValidUntil: &2025-01-05T19:40:00Z  // 1 hour from now
//   DelegatedOAuth2Tokens: [
//     {
//       ThirdpartyOAuth2ServiceID: "github-service"
//       Scopes: ["repo", "user"]
//     }
//   ]
//   CreatedAt:  2025-01-05T18:40:00Z
//   UpdatedAt:  2025-01-05T18:40:00Z
// }
// grant.IsActive() = true
```

**Domain Validation**: ✓ Passes (ValidUntil is in the future)

**Use Case**: Testing successful authorization with grant
```go
agent := fixtures.ValidAgent()
grant := fixtures.ActiveGrant(fixtures.DefaultPrincipal().String(), agent.ID)
storage.UserGrants().Create(ctx, grant)

resp, _ := server.AuthenticatedGET(
    "/oauth2/authorize?client_id="+agent.ClientID+"&...",
    fixtures.DefaultPrincipal(),
    nil,
)
// Should proxy to upstream (grant exists and is active)
```

### ExpiredGrant("user@example.com", agentID)

```go
grant := fixtures.ExpiredGrant("user@example.com", "550e8400-e29b-41d4-a716-446655440000")
// Output:
// UserGrant {
//   ID:         "aa54i945-j74g-86i9-f251-9911000aa555"  // Fresh UUID
//   Principal:  "user@example.com"
//   AgentID:    "550e8400-e29b-41d4-a716-446655440000"
//   ValidUntil: &2025-01-05T17:40:00Z  // 1 hour ago
//   DelegatedOAuth2Tokens: [
//     {
//       ThirdpartyOAuth2ServiceID: "github-service"
//       Scopes: ["repo", "user"]
//     }
//   ]
//   CreatedAt:  2025-01-05T18:40:00Z
//   UpdatedAt:  2025-01-05T18:40:00Z
// }
// grant.IsActive() = false
```

**Use Case**: Testing expired grant rejection
```go
grant := fixtures.ExpiredGrant(fixtures.DefaultPrincipal().String(), agent.ID)
storage.UserGrants().Create(ctx, grant)

resp, _ := server.AuthenticatedGET(
    "/oauth2/authorize?client_id="+agent.ClientID+"&...",
    fixtures.DefaultPrincipal(),
    nil,
)
// Should redirect to consent (grant expired, treat as non-existent)
```

### GrantExpiringIn("user@example.com", agentID, 24*time.Hour)

```go
grant := fixtures.GrantExpiringIn("user@example.com", agentID, 24*time.Hour)
// Output:
// UserGrant {
//   ID:         "bb65j056-k85h-97j0-g362-aa22111bb666"  // Fresh UUID
//   Principal:  "user@example.com"
//   AgentID:    "550e8400-e29b-41d4-a716-446655440000"
//   ValidUntil: &2025-01-06T18:40:00Z  // 24 hours from now
//   DelegatedOAuth2Tokens: [
//     {
//       ThirdpartyOAuth2ServiceID: "github-service"
//       Scopes: ["repo", "user"]
//     }
//   ]
//   CreatedAt:  2025-01-05T18:40:00Z
//   UpdatedAt:  2025-01-05T18:40:00Z
// }
// grant.IsActive() = true
```

**Use Case**: Testing custom expiration durations
```go
// Grant expiring in 30 seconds (test edge case)
grant := fixtures.GrantExpiringIn(principal, agentID, 30*time.Second)
```

### IndefiniteGrant("user@example.com", agentID)

```go
grant := fixtures.IndefiniteGrant("user@example.com", "550e8400-e29b-41d4-a716-446655440000")
// Output:
// UserGrant {
//   ID:         "cc76k167-l96i-08k1-h473-bb33222cc777"  // Fresh UUID
//   Principal:  "user@example.com"
//   AgentID:    "550e8400-e29b-41d4-a716-446655440000"
//   ValidUntil: nil  // No expiration
//   DelegatedOAuth2Tokens: [
//     {
//       ThirdpartyOAuth2ServiceID: "github-service"
//       Scopes: ["repo", "user"]
//     }
//   ]
//   CreatedAt:  2025-01-05T18:40:00Z
//   UpdatedAt:  2025-01-05T18:40:00Z
// }
// grant.IsActive() = true (always)
```

**Use Case**: Testing permanent grants
```go
grant := fixtures.IndefiniteGrant(principal, agentID)
// Grant never expires unless explicitly revoked
```

### GrantWithMultipleServices("user@example.com", agentID)

```go
grant := fixtures.GrantWithMultipleServices("user@example.com", agentID)
// Output:
// UserGrant {
//   ID:         "dd87l278-m07j-19l2-i584-cc44333dd888"  // Fresh UUID
//   Principal:  "user@example.com"
//   AgentID:    "550e8400-e29b-41d4-a716-446655440000"
//   ValidUntil: &2025-01-06T18:40:00Z
//   DelegatedOAuth2Tokens: [
//     {
//       ThirdpartyOAuth2ServiceID: "github-service"
//       Scopes: ["repo", "user"]
//     },
//     {
//       ThirdpartyOAuth2ServiceID: "google-service"
//       Scopes: ["calendar", "drive"]
//     },
//     {
//       ThirdpartyOAuth2ServiceID: "microsoft-service"
//       Scopes: ["mail.read", "calendar.read"]
//     }
//   ]
//   CreatedAt:  2025-01-05T18:40:00Z
//   UpdatedAt:  2025-01-05T18:40:00Z
// }
```

**Use Case**: Testing multiple delegated services
```go
grant := fixtures.GrantWithMultipleServices(principal, agentID)
// Agent can access GitHub, Google, and Microsoft services
```

## Configuration Fixtures

### DefaultOAuth2Config()

```go
config := fixtures.DefaultOAuth2Config()
// Output:
// Config {
//   Log: {
//     Level:  "info"
//     Format: "text"
//   }
//   Server: {
//     EndUser: {
//       Port:      8000
//       Bind:      "127.0.0.1"
//       PublicURL: "http://localhost:8000"
//       Authentication: {
//         Preauth: {
//           PrincipalHeaderName: "X-Remote-User"
//         }
//       }
//     }
//     Admin: {
//       Port:      14000
//       Bind:      "127.0.0.1"
//       PublicURL: "http://localhost:14000"
//     }
//     Shutdown: {
//       Timeout: 5 seconds
//     }
//   }
//   Storage: {
//     Backend: "memory"
//     Timeouts: {
//       Read:  5 seconds
//       Write: 5 seconds
//     }
//   }
//   OAuth2AuthServer: {
//     UpstreamIssuerURI:         "http://localhost:19000"
//     UpstreamAuthorizeEndpoint: "http://localhost:19000/authorize"
//     UpstreamTokenEndpoint:     "http://localhost:19000/token"
//     SupportedResponseTypes:    ["code"]
//     SupportedGrantTypes:       ["authorization_code", "refresh_token"]
//     UpstreamTimeout:           30 * time.Second
//     Mode:                      "proxy"
//   }
// }
```

**Use Case**: Basic E2E test setup
```go
config := fixtures.DefaultOAuth2Config()
server, _ := bootstrap.NewServerFactory(config).Build()
```

### OAuth2ConfigWithUpstream("http://mock-upstream:8080")

```go
config := fixtures.OAuth2ConfigWithUpstream("http://mock-upstream:8080")
// Output: DefaultOAuth2Config() with overrides:
// OAuth2AuthServer: {
//   UpstreamIssuerURI:         "http://mock-upstream:8080"
//   UpstreamAuthorizeEndpoint: "http://mock-upstream:8080/authorize"
//   UpstreamTokenEndpoint:     "http://mock-upstream:8080/token"
//   // ... other fields same as DefaultOAuth2Config()
// }
```

**Use Case**: Testing with mock upstream server
```go
mockUpstream := helpers.NewMockUpstreamOAuth2Server()
config := fixtures.OAuth2ConfigWithUpstream(mockUpstream.Server.URL)
server, _ := bootstrap.NewServerFactory(config).Build()
```

### OAuth2ConfigWithTimeout(60 * time.Second)

```go
config := fixtures.OAuth2ConfigWithTimeout(60 * time.Second)
// Output: DefaultOAuth2Config() with override:
// OAuth2AuthServer: {
//   UpstreamTimeout: 60 * time.Second
//   // ... other fields same as DefaultOAuth2Config()
// }
```

**Use Case**: Testing timeout behavior
```go
config := fixtures.OAuth2ConfigWithTimeout(2 * time.Second)
// Test upstream failures when response takes > 2 seconds
```

### OAuth2ConfigWithLogLevel("debug")

```go
config := fixtures.OAuth2ConfigWithLogLevel("debug")
// Output: DefaultOAuth2Config() with override:
// Log: {
//   Level:  "debug"
//   Format: "text"
//   // ... other fields same as DefaultOAuth2Config()
// }
```

**Use Case**: Debugging E2E test failures
```go
config := fixtures.OAuth2ConfigWithLogLevel("debug")
// Get detailed logging for debugging
```

### OAuth2ConfigWithPublicURL("https://broker.example.com")

```go
config := fixtures.OAuth2ConfigWithPublicURL("https://broker.example.com")
// Output: DefaultOAuth2Config() with override:
// Server: {
//   EndUser: {
//     PublicURL: "https://broker.example.com"
//     // ... other fields same as DefaultOAuth2Config()
//   }
// }
```

**Use Case**: Testing metadata discovery
```go
config := fixtures.OAuth2ConfigWithPublicURL("https://broker.example.com")
resp, _ := server.GET("/.well-known/oauth-authorization-server", nil)
// Metadata issuer should be "https://broker.example.com"
```

## Determinism vs Non-Determinism

### Deterministic (Same Output Every Time)

```go
// These always return the same values for same input:
p1 := fixtures.DefaultPrincipal()
p2 := fixtures.DefaultPrincipal()
p1.String() == p2.String()  // true: "user@example.com" == "user@example.com"

c1 := fixtures.DefaultOAuth2Config()
c2 := fixtures.DefaultOAuth2Config()
c1.Storage.Backend == c2.Storage.Backend  // true: "memory" == "memory"
```

### Non-Deterministic (Fresh UUIDs Each Time)

```go
// These generate different UUIDs each time:
agent1 := fixtures.ValidAgent()
agent2 := fixtures.ValidAgent()
agent1.ID == agent2.ID  // false: different UUIDs

grant1 := fixtures.ActiveGrant("user@example.com", "agent-1")
grant2 := fixtures.ActiveGrant("user@example.com", "agent-1")
grant1.ID == grant2.ID  // false: different UUIDs
```

## Summary

| Fixture | Type | Deterministic | Use Case |
|---------|------|---|---|
| DefaultPrincipal | Principal | Yes | Default test user |
| AnotherPrincipal | Principal | Yes | Alternative user |
| AdminPrincipal | Principal | Yes | Admin user |
| ValidAgent | Agent | No (UUID) | Basic agent |
| AnotherAgent | Agent | No (UUID) | Alternative agent |
| AgentWithClientID | Agent | No (UUID) | Known client ID |
| AgentWithURLs | Agent | No (UUID) | With documentation URLs |
| ActiveGrant | Grant | No (UUID) | Active permission |
| ExpiredGrant | Grant | No (UUID) | Expired permission |
| GrantExpiringIn | Grant | No (UUID) | Custom expiration |
| IndefiniteGrant | Grant | No (UUID) | Permanent permission |
| GrantWithMultipleServices | Grant | No (UUID) | Multiple OAuth2 services |
| GrantWithService | Grant | No (UUID) | Single service with custom scopes |
| DefaultOAuth2Config | Config | Yes | Default test config |
| OAuth2ConfigWithUpstream | Config | Yes | Custom upstream URL |
| OAuth2ConfigWithTimeout | Config | Yes | Custom timeout |
| OAuth2ConfigWithLogLevel | Config | Yes | Custom log level |
| OAuth2ConfigWithPublicURL | Config | Yes | Custom public URL |
| OAuth2ConfigWithStorage | Config | Yes | Custom storage backend |
