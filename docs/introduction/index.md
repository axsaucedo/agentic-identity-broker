---
title: Introduction
description: Learn what the Agentic Identity Broker is, why it exists, and how it enables secure identity brokering for agentic systems and autonomous agents.
---

# Introduction to Agentic Identity Broker

The **Agentic Identity Broker** is an open-source identity brokering system purpose-built for agentic systems—AI agents, autonomous services, and multi-agent architectures. Unlike traditional identity providers designed for human users and web applications, this broker focuses exclusively on agent-to-agent authentication, identity verification, and secure token exchange in distributed agentic environments.

## What is the Agentic Identity Broker?

The Agentic Identity Broker serves as a trusted intermediary for identity management in systems where multiple autonomous agents need to securely identify, authenticate, and authorize each other. It acts as a central authority that issues, verifies, and manages identities specifically designed for agents—not humans.

Think of it as an identity layer for your agentic infrastructure: when Agent A needs to verify Agent B's identity, or when an autonomous service must authenticate to another system, the broker provides the cryptographic proof and verification mechanisms to enable secure, verifiable interactions.

### Core Capabilities

The broker provides three fundamental capabilities for agentic systems:

1. **Identity Issuance**: Generate and manage unique identities for agents, services, and autonomous systems with cryptographic keys and verifiable credentials.

2. **Identity Verification**: Validate agent identities through signature verification, token validation, and cryptographic proofs without requiring centralized databases or user directories.

3. **Identity Brokering**: Act as a trusted intermediary between agents that don't directly trust each other, enabling secure communication across organizational or trust boundaries.

## Why Does This Exist?

As organizations build increasingly complex multi-agent systems, they face a fundamental challenge: **how do autonomous agents securely identify and trust each other?** Traditional identity providers like Keycloak, Auth0, and Okta were designed for human users logging into web applications—they assume interactive authentication flows, session management, and human-in-the-loop workflows.

Agentic systems have different requirements:

- **No human interaction**: Agents authenticate autonomously without user input
- **Machine-to-machine trust**: Verification happens through cryptographic proofs, not passwords
- **Distributed architectures**: Agents may span multiple clouds, edge devices, or air-gapped environments
- **High-frequency verification**: Authentication happens thousands of times per second, not once per session
- **Fail-closed security**: Identity failures must block operations entirely, not degrade gracefully

The Agentic Identity Broker addresses these needs with a security-first, agent-native architecture that treats identity as a foundational infrastructure concern for agentic systems.

## Key Concepts

Understanding the broker requires familiarity with a few core concepts specific to agentic identity:

### Agentic Identity

An **agentic identity** represents a cryptographically verifiable identifier for an autonomous agent or service. Unlike human identities (tied to email addresses, usernames, or social profiles), agentic identities are typically public-private key pairs, service accounts, or machine identities with no human attributes.

Example:
```text
Agent ID: agent://aib/service-discovery/instance-42
Public Key: ed25519:AAAC3NzaC1lZDI1NTE5AAAAIDfJK...
```

### Identity Brokering vs Identity Providing

An **identity broker** acts as an intermediary between multiple identity sources and relying parties, translating and verifying identities across trust boundaries. An **identity provider** (IdP) issues and manages identities directly.

The Agentic Identity Broker can operate in both modes:
- As a **broker**: Translating identities between different agent systems
- As a **provider**: Directly issuing identities to agents it manages

This distinction matters for federated agent networks where agents from different organizations need to interact securely.

### Agent-to-Agent Authentication

Unlike user-to-application authentication (OAuth, SAML), **agent-to-agent authentication** assumes both parties are autonomous systems with cryptographic capabilities. Authentication flows use:

- Public key cryptography for mutual authentication
- Token exchange protocols for delegation
- Signature verification for non-repudiation
- Certificate chains for hierarchical trust

No cookies, sessions, or redirect flows—just direct cryptographic proof exchange.

## Who Should Use This?

The Agentic Identity Broker is designed for developers and platform engineers building:

- **Multi-agent orchestration platforms** where multiple AI agents collaborate on complex tasks
- **Autonomous microservices** requiring service-to-service authentication beyond API keys
- **Agent marketplaces** where agents from different vendors must securely interact
- **Federated agent networks** spanning multiple organizations or security domains
- **Edge computing environments** with intermittent connectivity requiring local identity verification

### When to Use This Tool

Consider the Agentic Identity Broker when you need:

- Identity verification for systems with **no human users** or where humans are not in the authentication loop
- **Cryptographic proof** of agent identity rather than password-based authentication
- **High-throughput identity verification** (thousands of authentications per second)
- **Federated trust** across organizational boundaries for agent interactions
- **Fail-closed security posture** where identity failures must completely block operations

### When NOT to Use This Tool

This tool may not be appropriate if:

- Your primary use case is **human user authentication** (use Keycloak, Auth0, or Okta instead)
- You need **social login integration** (Google, Facebook, GitHub)
- Your system requires **user directories, roles, and RBAC** for human users
- You're building a traditional web application with login pages and user sessions

For those scenarios, traditional identity providers are better suited. See [Why Not Traditional IdP?](./why-not-idp.md) for a detailed comparison.

## What's Next?

Now that you understand what the Agentic Identity Broker is and why it exists, explore the following resources:

- **[Why Not Traditional IdP?](./why-not-idp.md)**: Detailed comparison with Keycloak, Auth0, Okta, and traditional identity providers
- **[Use Cases](./use-cases.md)**: Concrete scenarios where agentic identity brokering solves real problems
- **[Quick Start](/docs/quick-start)**: Get the broker running locally in 15 minutes
- **[Architecture & Concepts](/docs/architecture)**: Deep dive into architectural decisions and design principles

## Project Status

The Agentic Identity Broker is an **open-source project** in active development. We're building the foundational identity infrastructure for the agentic systems ecosystem. Contributions, feedback, and real-world use cases are welcome.

- **License**: [To be determined - typically Apache 2.0 or MIT for OSS identity projects]
- **Repository**: [GitHub link to be added]
- **Community**: [Discussions link to be added]

Ready to get started? Head to the [Quick Start guide](/docs/quick-start) to install and run the broker locally.
