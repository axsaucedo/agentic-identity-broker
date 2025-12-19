---
title: Use Cases
description: Real-world scenarios where Agentic Identity Broker solves critical identity challenges for multi-agent systems, autonomous services, and distributed agentic architectures.
---

# Use Cases

The Agentic Identity Broker addresses identity challenges in systems where autonomous agents need to securely identify and authenticate each other. This page explores concrete scenarios where agent-native identity brokering is essential for building reliable, secure agentic architectures.

## Use Case 1: Multi-Agent Orchestration Platform

### Scenario

You're building an enterprise AI orchestration platform where multiple specialized agents collaborate on complex business processes. For example, in a document processing pipeline:

- **Agent A (Document Parser)** extracts text from uploaded PDFs
- **Agent B (Data Validator)** checks extracted data for completeness and accuracy
- **Agent C (Enrichment Service)** augments data with external information
- **Agent D (Storage Service)** commits validated data to databases

Each agent is a separate service, potentially developed by different teams, running in different containers, and operating autonomously.

### The Identity Challenge

How does Agent B verify that a document was actually parsed by Agent A (not a malicious actor)? How does Agent D ensure it only stores data validated by Agent B? Traditional approaches fail:

- **API Keys**: Static, shared secrets that can't prove which specific agent made a request
- **Mutual TLS**: Requires certificate management infrastructure and doesn't carry semantic identity
- **OAuth2 Client Credentials**: Creates "fake user accounts" for services, not designed for agent-to-agent trust

### How Agentic Identity Broker Solves This

Each agent receives a cryptographic identity from the broker:

```text
Agent A Identity:
  agent_id: agent://aib/orchestrator/document-parser-v2
  public_key: ed25519:AAAC3NzaC1lZDI1NTE5AAAAIDfJK...
  capabilities: [parse_documents, emit_events]
  trust_level: verified
```

When Agent A sends parsed data to Agent B, it **signs the payload** with its private key. Agent B verifies the signature using Agent A's public key (retrieved from the broker), cryptographically proving the data came from the legitimate parser agent.

**Benefits**:
- **Non-repudiation**: Agent A can't deny sending the data (signed with private key)
- **Auditability**: Every agent interaction is cryptographically traceable
- **Zero-trust architecture**: No implicit trust—every agent proves identity on every interaction
- **Dynamic capabilities**: Broker can grant/revoke agent capabilities without redeploying services

### Example Flow

```python
# Agent A (Document Parser)
def parse_and_send(document):
    parsed_data = extract_text(document)

    # Sign the payload with Agent A's private key
    signature = sign_with_private_key(parsed_data, agent_a_private_key)

    # Send to Agent B with identity proof
    send_to_agent_b({
        "data": parsed_data,
        "agent_id": "agent://aib/orchestrator/document-parser-v2",
        "signature": signature
    })

# Agent B (Data Validator)
def receive_and_validate(payload):
    # Verify the sender's identity with Agentic Identity Broker
    sender_identity = broker.verify_signature(
        payload["agent_id"],
        payload["data"],
        payload["signature"]
    )

    if sender_identity.verified and "parse_documents" in sender_identity.capabilities:
        # Cryptographically proven to be from authorized parser
        validate_data(payload["data"])
    else:
        # Reject: unauthorized or invalid identity
        raise UnauthorizedError("Invalid agent identity")
```

---

## Use Case 2: Agent-to-Service Authentication in Microservices

### Scenario

You're building a microservices architecture where backend services need to authenticate to APIs and databases. For example:

- **Recommendation Service** queries the User Preferences API
- **Billing Service** writes to the Payments Database
- **Notification Service** sends events to a message queue

In traditional setups, these services use static API keys, environment variables, or service accounts. But in a dynamic environment (Kubernetes with auto-scaling), you need ephemeral, rotatable credentials that can't be compromised.

### The Identity Challenge

Static credentials have serious problems:
- **Leaked secrets**: API keys in environment variables get committed to repos, logged, or exposed
- **No rotation**: Changing an API key requires redeploying all services that use it
- **Coarse-grained**: One API key = access to everything (no least-privilege principle)
- **No auditability**: Can't trace which specific service instance made a request

### How Agentic Identity Broker Solves This

Each service instance receives a **short-lived identity token** on startup:

```yaml
Service: recommendation-service
Pod: recommendation-service-7d4f5c8b-xk2jm
Identity Token:
  issued_at: 2025-12-14T10:30:00Z
  expires_at: 2025-12-14T11:30:00Z (1 hour TTL)
  agent_id: agent://aib/services/recommendation-v1/7d4f5c8b-xk2jm
  capabilities: [read_user_preferences, read_product_catalog]
```

The service uses this token to authenticate to downstream APIs. The token is:
- **Short-lived**: Expires in 1 hour, must be refreshed
- **Pod-specific**: Identifies the exact Kubernetes pod (not just "recommendation service")
- **Capability-bound**: Only grants specific permissions, not broad access
- **Auditable**: Every API call is traceable to a specific pod at a specific time

### Example Flow

```go
// Service startup: Request identity from broker
func initializeService() {
    // Authenticate to broker using pod service account
    identityToken, err := agenticBroker.RequestIdentity(context.Background(), &IdentityRequest{
        ServiceName: "recommendation-service",
        PodName:     os.Getenv("POD_NAME"),
        Namespace:   os.Getenv("POD_NAMESPACE"),
    })

    // Store token for API authentication
    cache.Set("identity_token", identityToken, 1*time.Hour)
}

// Making API requests with identity
func getRecommendations(userID string) ([]Product, error) {
    token := cache.Get("identity_token")

    // Include identity token in API request
    resp, err := http.Get(fmt.Sprintf("https://api.example.com/users/%s/preferences", userID),
        http.Header{"Authorization": fmt.Sprintf("Bearer %s", token)})

    // API server verifies token with Agentic Identity Broker
    // and checks capabilities before serving response
}
```

**Benefits**:
- **No static secrets**: Tokens are ephemeral, issued on-demand
- **Automatic rotation**: Expired tokens must be refreshed, forcing regular rotation
- **Fine-grained permissions**: Each service only gets capabilities it needs
- **Full audit trail**: Every API call tied to specific pod, service, and timestamp

---

## Use Case 3: Federated Agent Networks Across Organizations

### Scenario

You're building a supply chain platform where agents from multiple companies need to collaborate:

- **Company A (Manufacturer)**: Inventory management agent
- **Company B (Logistics)**: Shipping coordination agent
- **Company C (Retailer)**: Order fulfillment agent

These agents must securely exchange data (shipment status, inventory levels, order confirmations), but each company operates its own infrastructure with its own security policies. No single company wants to hand over credentials to another company's identity system.

### The Identity Challenge

Cross-organizational trust is hard:
- **No shared identity provider**: Each company has its own IdP (Okta, Azure AD, etc.)
- **No mutual trust**: Company A doesn't want Company B's IdP to have authority over its agents
- **Federation complexity**: Traditional SAML/OIDC federation is designed for human SSO, not agent-to-agent

### How Agentic Identity Broker Solves This

Each company runs its own **Agentic Identity Broker instance** with **federation protocols** for cross-org trust:

```text
Company A Broker ←──────────┐
  │ Issues identities for     │
  │ Company A agents          │
                              ├─── Federation Trust
Company B Broker ←──────────┤    (Broker-to-Broker)
  │ Issues identities for     │
  │ Company B agents          │
                              │
Company C Broker ←──────────┘
  │ Issues identities for
  │ Company C agents
```

When Company A's agent sends data to Company B's agent:

1. **Company A agent** gets identity token from **Company A broker**
2. **Company A agent** sends data to **Company B agent** with token
3. **Company B agent** validates token with **Company B broker**
4. **Company B broker** verifies token signature with **Company A broker** (federation)

**Key insight**: Company B's broker doesn't issue identities for Company A's agents—it simply verifies that Company A's broker issued a valid token. This preserves organizational independence while enabling cross-org trust.

### Example Federation Configuration

```yaml
# Company A Broker Configuration
federation:
  trusted_brokers:
    - org_id: company-b
      broker_url: https://broker.company-b.com
      public_key: ed25519:BBBC3NzaC1lZDI1NTE5AAAAIDfJK...
      trust_level: verified
      allowed_capabilities: [receive_shipments, update_inventory]

    - org_id: company-c
      broker_url: https://broker.company-c.com
      public_key: ed25519:CCCC3NzaC1lZDI1NTE5AAAAIDfJK...
      trust_level: verified
      allowed_capabilities: [place_orders]
```

**Benefits**:
- **Organizational autonomy**: Each company controls its own broker and policies
- **No shared secrets**: Federation uses public key cryptography, no passwords
- **Revocable trust**: Any company can remove federation relationships instantly
- **Auditable**: All cross-org interactions are logged with cryptographic proof

---

## Use Case 4: Edge Computing with Intermittent Connectivity

### Scenario

You're deploying AI agents to edge devices (IoT sensors, autonomous vehicles, field equipment) that have intermittent or no connectivity to the cloud. For example:

- **Autonomous delivery drones** operating in remote areas
- **Industrial robots** in factories with air-gapped networks
- **Medical devices** that must operate during network outages

These edge agents need to authenticate to each other and to local services even when disconnected from central infrastructure.

### The Identity Challenge

Traditional identity systems assume always-on connectivity:
- **OAuth2/OIDC**: Requires network calls to token endpoints
- **Certificate revocation**: Checking CRLs or OCSP requires connectivity
- **Dynamic credentials**: Can't refresh tokens when offline

### How Agentic Identity Broker Solves This

Edge agents receive **offline-verifiable identities** that can be validated without network access:

1. **Pre-provisioning**: Before deployment, each edge agent receives:
   - Its own identity (public/private key pair)
   - Public keys of other agents it will interact with
   - Broker's public key for signature verification

2. **Local verification**: When Agent A (drone) meets Agent B (ground station), they verify each other's identities using pre-provisioned public keys—no network call required.

3. **Periodic sync**: When connectivity is restored, agents sync with the broker to:
   - Refresh revocation lists
   - Update trust policies
   - Report audit logs

### Example Architecture

```text
┌─────────────────────────────────────────┐
│  Edge Environment (Air-Gapped)          │
│                                          │
│  Drone Agent A ←─┐                      │
│  Ground Station  ├─ Local Verification  │
│  Sensor Agent C ←─┘  (No network)       │
│                                          │
│  ┌──────────────────────────┐           │
│  │  Local Broker Cache      │           │
│  │  - Public keys           │           │
│  │  - Revocation lists      │           │
│  │  - Trust policies        │           │
│  └──────────────────────────┘           │
└─────────────────────────────────────────┘
         │ Periodic sync when online
         ▼
  Agentic Identity Broker (Cloud)
```

**Benefits**:
- **Offline operation**: Agents verify identities without network connectivity
- **Eventually consistent**: Revocations sync when connectivity is restored
- **Resilient**: System continues operating during outages
- **Security**: Cryptographic verification doesn't require online checks

---

## Use Case 5: Multi-Tenant Agent Marketplace

### Scenario

You're building an agent marketplace where third-party developers can deploy agents that users can rent and orchestrate. For example:

- **Data analysis agents** from Vendor A
- **Report generation agents** from Vendor B
- **Forecasting agents** from Vendor C

Users (tenants) want to compose workflows using agents from multiple vendors, but need guarantees that:
- Agents can only access data they're authorized for
- Vendors can't access other vendors' agents or user data
- Users can revoke agent access at any time

### The Identity Challenge

Multi-tenancy with third-party agents is a security minefield:
- **Tenant isolation**: How do you prevent Vendor A's agent from accessing Tenant 1's data when running for Tenant 2?
- **Vendor trust**: Users don't trust vendor agents with full access to their systems
- **Dynamic permissions**: Access policies must change as users add/remove agents from workflows

### How Agentic Identity Broker Solves This

The broker issues **tenant-scoped identities** with **context-bound capabilities**:

```json
{
  "agent_id": "agent://marketplace/vendor-a/data-analyzer-v2",
  "tenant_id": "tenant-123",
  "capabilities": [
    "read:tenant-123:datasets",
    "write:tenant-123:analysis-results"
  ],
  "constraints": {
    "max_data_volume": "100GB",
    "allowed_endpoints": ["https://api.tenant-123.com/*"]
  },
  "expires_at": "2025-12-14T12:00:00Z"
}
```

**Key properties**:
- **Tenant-scoped**: Agent can only access Tenant 123's data, not other tenants
- **Capability-bound**: Can read datasets and write results, but can't delete or modify datasets
- **Constrained**: Can't read more than 100GB or call external endpoints
- **Time-limited**: Identity expires after job completion

### Example Marketplace Flow

```python
# User rents Data Analyzer agent for their workflow
def rent_agent(user_id, vendor_agent_id):
    # Request tenant-scoped identity from broker
    agent_identity = broker.issue_identity({
        "agent_id": vendor_agent_id,
        "tenant_id": user_id,
        "capabilities": ["read:datasets", "write:results"],
        "ttl": "1h"  # Job expected to complete in 1 hour
    })

    # Deploy agent with identity
    deploy_agent_instance(vendor_agent_id, agent_identity)

# Vendor's agent uses identity to access user data
def analyze_data(dataset_id):
    # Agent identity is tenant-scoped—can only access this tenant's data
    data = api.get_dataset(dataset_id, identity_token=my_identity)
    results = perform_analysis(data)
    api.write_results(results, identity_token=my_identity)

    # Token expires after 1 hour—agent can't access data after job completes
```

**Benefits**:
- **Tenant isolation**: Cryptographic guarantees prevent cross-tenant access
- **Least privilege**: Agents only get capabilities needed for specific jobs
- **Auditability**: All vendor agent actions logged and attributable
- **Revocable**: User can instantly revoke agent access if needed

---

## Common Patterns Across Use Cases

All these use cases share common requirements that Agentic Identity Broker addresses:

1. **Autonomous authentication**: No human in the loop—agents authenticate themselves
2. **Cryptographic proof**: Identity is verifiable through signatures, not passwords
3. **Fine-grained capabilities**: Least-privilege access control for agents
4. **High throughput**: Thousands of identity verifications per second
5. **Auditability**: Complete audit trail of agent interactions
6. **Fail-closed security**: Invalid identities block operations entirely

## When NOT to Use These Patterns

These use cases are **not appropriate** for:

- Human user authentication (use Auth0, Okta, or Keycloak)
- Web application session management (use traditional session cookies)
- Social login or SSO for end users (use OAuth2/OIDC providers)
- Systems where agents don't need to verify each other's identity

If your agents don't need to prove their identity to each other, simpler authentication mechanisms (API keys, mutual TLS) may be sufficient.

## Next Steps

Ready to implement these patterns in your system?

- **[Quick Start](/docs/quick-start)**: Install the broker and try a simple agent authentication flow
- **[Features](/docs/features)**: Explore capabilities like token issuance, signature verification, and federation
- **[API Reference](/docs/api)**: Technical details on integrating agents with the broker
- **[Architecture](/docs/architecture)**: Understand the design principles behind these patterns

Have a use case not covered here? Join the [community discussions](#) to share your scenario and get guidance.
