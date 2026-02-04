---
title: Why Not Traditional IdP?
description: Understanding when to use Agentic Identity Broker versus traditional identity providers like Keycloak, Auth0, Okta, and why agent-native identity systems differ fundamentally from human user authentication.
---

# Why Not Traditional IdP?

If you're evaluating identity solutions for your agentic system, you've likely encountered established identity providers like Keycloak, Auth0, Okta, or Azure AD. These are proven, production-ready solutions used by thousands of organizations. So why would you need a specialized identity broker for agentic systems?

This page explains the fundamental differences between traditional identity providers and the Agentic Identity Broker, helping you make an informed decision about which approach fits your architecture.

## The Core Difference: Human-Centric vs Agent-Centric

Traditional identity providers (IdPs) were architected for a specific use case: **humans logging into applications**. Every design decision—from authentication flows to session management to user interfaces—assumes a human user sitting at a browser, clicking buttons, and interacting with forms.

The Agentic Identity Broker inverts this assumption: **no human is involved in the authentication flow**. Agents authenticate autonomously using cryptographic proofs, operating at machine timescales without user interaction.

This isn't just a feature difference—it's a fundamental architectural distinction that affects every layer of the system.

## Comparison Table

| Feature | Keycloak / Auth0 / Okta | Agentic Identity Broker |
|---------|-------------------------|-------------------------|
| **Primary Use Case** | Human users logging into web/mobile apps | Autonomous agents authenticating to each other |
| **Authentication Flow** | Interactive (OAuth2, SAML, OIDC redirects) | Direct cryptographic proof exchange |
| **Identity Type** | User accounts (email, username, social login) | Machine identities (public keys, service accounts) |
| **Session Management** | Cookies, session tokens, refresh flows | Stateless verification, short-lived tokens |
| **User Interface** | Login pages, admin consoles, user portals | API-only, no UI required |
| **Directory Services** | User databases, LDAP, Active Directory integration | No user directory—identity is cryptographic proof |
| **Authentication Speed** | Seconds (human interaction time) | Milliseconds (thousands of auths/second) |
| **Failure Mode** | Degrade gracefully, retry with user | Fail closed—block operation entirely |
| **Trust Model** | Centralized authority with user credentials | Distributed trust with cryptographic verification |
| **Federation** | SAML, OIDC federation for SSO | Agent-to-agent trust chains |

## When to Use Traditional IdPs

Traditional identity providers excel when you're building systems for **human users**:

### Use Keycloak/Auth0/Okta If...

1. **You have human end users** logging into web applications, mobile apps, or dashboards
2. **You need social login** (Google, Facebook, GitHub, LinkedIn)
3. **You require user management interfaces** for administrators to manage accounts, reset passwords, and assign roles
4. **You're implementing Single Sign-On (SSO)** across multiple applications for human users
5. **You need compliance features** like MFA, passwordless login, or user consent flows
6. **You have existing identity infrastructure** (Active Directory, LDAP) that you need to integrate with
7. **You want hosted/managed services** with guaranteed uptime and support

### Example: E-commerce Platform

```text
Scenario: You're building an online store where customers create accounts,
log in to view orders, and manage their profiles.

Solution: Auth0 or Okta
- Users log in with email/password or social login
- Session cookies maintain logged-in state
- Password reset flows via email
- MFA for high-value accounts
- Admin console for customer support
```

This is the perfect use case for traditional IdPs. Don't reinvent this wheel—use proven solutions.

## When to Use Agentic Identity Broker

The Agentic Identity Broker is designed for **agent-to-agent authentication** where humans are not involved:

### Use Agentic Identity Broker If...

1. **Your system consists of autonomous agents** or services that communicate without human interaction
2. **Authentication happens machine-to-machine** using cryptographic keys, not passwords
3. **You need high-frequency identity verification** (thousands of authentications per second)
4. **Agents operate in distributed environments** across clouds, edge devices, or air-gapped networks
5. **Identity must be verifiable without central databases** or network connectivity
6. **You're building federated agent networks** where agents from different organizations interact
7. **Security posture requires fail-closed behavior** where identity failures block all operations

### Example: Multi-Agent Trading System

```text
Scenario: You're building a trading platform where multiple AI agents analyze
markets, execute trades, and settle transactions autonomously.

Problem with traditional IdP:
- No human to redirect for OAuth flow
- Agents need to authenticate thousands of times per second
- Session cookies don't make sense for agent-to-agent communication
- Password-based auth is inadequate for cryptographic non-repudiation

Solution: Agentic Identity Broker
- Each agent has a cryptographic identity (public key)
- Authentication via signature verification (no user interaction)
- Stateless token verification at microsecond latency
- Cryptographic proof for audit trails and non-repudiation
```

This scenario requires agent-native identity architecture that traditional IdPs weren't designed for.

## Real-World Architecture Patterns

### Pattern 1: Hybrid Architecture (Most Common)

Many systems need **both** human users and agent-to-agent authentication:

```text
┌─────────────────────────────────────────────────────┐
│                 Your Application                     │
├─────────────────────────────────────────────────────┤
│  Human Users               Agent Layer               │
│  (Web, Mobile)             (Backend Services)        │
│       │                         │                    │
│       ├─Auth0/Okta             ├─Agentic Identity   │
│       │ (User Login)           │  Broker             │
│       │                        │  (Service Auth)     │
│       ▼                        ▼                     │
│   Dashboard UI           AI Agents/Services          │
└─────────────────────────────────────────────────────┘
```

**Recommendation**: Use Auth0/Okta for human users AND Agentic Identity Broker for your agent layer. They solve different problems and can coexist.

### Pattern 2: Pure Agentic System

For systems with **no human users** (fully autonomous):

```text
┌─────────────────────────────────────────────────────┐
│            Autonomous Agent Network                  │
├─────────────────────────────────────────────────────┤
│  Agent A  ←──┐                                       │
│  Agent B  ←──┼──  Agentic Identity Broker           │
│  Agent C  ←──┘      (Centralized Trust)             │
│     ...                                              │
└─────────────────────────────────────────────────────┘
```

**Recommendation**: Pure Agentic Identity Broker. No need for user-facing IdP.

### Pattern 3: Federated Multi-Org Agents

For agent networks spanning organizational boundaries:

```text
Org A Agents ←→ Agentic Broker A ←──┐
                                     ├──  Federation
Org B Agents ←→ Agentic Broker B ←──┘
```

**Recommendation**: Each organization runs their own broker with federation protocols for cross-org agent trust.

## Key Technical Differences

### 1. Authentication Flow

**Traditional IdP (OAuth2 Authorization Code Flow)**:
```text
1. User clicks "Login"
2. Redirect to IdP login page
3. User enters credentials
4. IdP redirects back with authorization code
5. App exchanges code for token
6. Token used for API access
Total time: 2-5 seconds (human interaction)
```

**Agentic Identity Broker (Direct Proof Exchange)**:
```text
1. Agent creates signed request with private key
2. Broker verifies signature with public key
3. Broker issues short-lived token
4. Agent uses token for API access
Total time: <10ms (no human interaction)
```

### 2. Identity Representation

**Traditional IdP Identity**:
```json
{
  "sub": "user123",
  "email": "alice@example.com",
  "name": "Alice Smith",
  "roles": ["admin", "user"],
  "last_login": "2025-12-14T10:30:00Z"
}
```

**Agentic Identity**:
```json
{
  "agent_id": "agent://aib/trading-bot-42",
  "public_key": "ed25519:AAAC3NzaC1lZDI1NTE5AAAAIDfJK...",
  "capabilities": ["execute_trades", "read_market_data"],
  "trust_level": "verified",
  "issued_at": "2025-12-14T10:30:00Z"
}
```

Notice: No human attributes, no email, no "last login"—just cryptographic proof and capabilities.

### 3. Security Model

**Traditional IdP**:
- Secrets stored in database (password hashes)
- Session tokens track logged-in state
- Compromise of password = full account access
- MFA adds second factor (SMS, TOTP)

**Agentic Identity Broker**:
- No secrets stored—identity IS the public key
- Stateless verification (no session state)
- Compromise of private key = revoke and reissue
- Security is cryptographic, not knowledge-based

## Migration and Integration

### Migrating FROM Traditional IdP

If you're currently using Keycloak/Auth0 for service-to-service authentication (not human users), you might benefit from migrating to Agentic Identity Broker:

**Signs you've outgrown traditional IdP for services**:
- You're creating "fake user accounts" for services
- Client credentials flow feels clunky for your use case
- You need higher authentication throughput than IdP can provide
- You're implementing custom token verification logic
- Your agents operate in air-gapped or offline environments

### Integrating WITH Traditional IdP

For hybrid architectures (human users + agents):

1. **User Layer**: Keep Auth0/Okta for human authentication
2. **Agent Layer**: Use Agentic Identity Broker for service-to-service
3. **Bridge Layer**: Map human user actions to agent identity when needed

Example: User "Alice" triggers an agent to execute a trade. Alice authenticates to your app via Auth0. Your app then uses Agentic Identity Broker to authenticate the trading agent that executes on Alice's behalf.

## Cost Considerations

### Traditional IdP Costs
- **Per-user pricing**: Auth0 charges per monthly active user (MAU)
- **Enterprise features**: SSO, custom domains, advanced security often require expensive tiers
- **Scaling costs**: Costs increase linearly with user count

### Agentic Identity Broker Costs
- **Open source**: Self-hosted, no per-user licensing
- **Infrastructure costs**: Pay only for compute/storage
- **Scaling costs**: Horizontal scaling without per-agent fees
- **Trade-off**: You manage infrastructure and updates

For high-throughput agent systems with thousands of machine identities, self-hosted Agentic Identity Broker can be significantly more cost-effective than paying per-agent pricing to commercial IdPs.

## Making the Decision

### Choose Traditional IdP (Keycloak/Auth0/Okta) If:
- Human users are your primary use case
- You need user management interfaces and admin consoles
- You want managed/hosted services with SLAs
- Your authentication patterns match OAuth2/OIDC flows
- You need social login or enterprise SSO

### Choose Agentic Identity Broker If:
- Autonomous agents are your primary use case
- You need high-throughput machine-to-machine authentication
- Your agents use cryptographic identities (keys, certificates)
- You operate in distributed or offline environments
- You need fail-closed security for agent interactions

### Use BOTH If:
- You have human users AND autonomous agents
- Users interact with a frontend (use traditional IdP)
- Backend agents handle processing (use Agentic Identity Broker)

## Next Steps

- **[Use Cases](./use-cases.md)**: See concrete examples of when Agentic Identity Broker solves real problems
- **[Quick Start](/docs/quick-start)**: Try running the broker locally to see how it differs from traditional IdPs
- **[Architecture](/docs/architecture)**: Understand the design principles behind agent-native identity

Still unsure? Join the [community discussions](#) to ask about your specific use case.
