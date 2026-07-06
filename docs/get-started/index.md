---
title: "Get started"
description: "Run the full Agentic Identity Broker stack locally with Docker Compose and walk one delegation end to end, from the sample agent through the consent UI to a stored grant."
---

# Get started

This tutorial gets you from a fresh clone to a working delegation. You will **run the full
broker stack locally and walk one delegation end to end** — starting a sample agent's
authorization request, consenting in the browser, and confirming the grant through the admin
API.

Everything runs from a single Docker Compose stack with sample data seeded automatically, so
you can see the whole flow without registering anything by hand first.

## Before you start

You need:

- **Docker or Podman**, with a Compose plugin (`docker compose`, `docker-compose`, or
  `podman-compose`). The stack auto-detects whichever you have.
- **[just](https://github.com/casey/just)**, the command runner the project uses for its
  workflows.
- **The repository**, cloned locally:

  ```bash
  git clone https://github.com/zalando-incubator/agentic-identity-broker.git
  cd agentic-identity-broker
  ```

You do not need Go, Node.js, PostgreSQL, or any cloud credentials for this tutorial. The
stack uses an in-memory store and a local encryption key generated for you at startup.

## Step 1: Start the stack

From the repository root, bring the whole stack up in the background:

```bash
just compose-up-detached
```

This builds and starts every service the flow needs and auto-seeds sample agents, services,
and permission sets once the broker reports healthy. When it finishes, it prints a health
summary and the service URLs:

```text
Service URLs:
  - Broker (end-user): http://localhost:8000
  - Broker (admin): http://localhost:14000
  - Frontend (consent UI): http://localhost:3000
  - Sample OAuth2 client: http://localhost:9002/oauth2/authorize
```

The stack is:

| Service | URL | Role |
|---|---|---|
| Broker — end-user API | http://localhost:8000 | Consent, third-party sessions, OAuth2 server surface |
| Broker — admin API | http://localhost:14000 | Agents, services, permission sets, credentials |
| Consent UI | http://localhost:3000 | The React screen where you approve delegations |
| Mock OAuth2 services | http://localhost:9000, `:9001`, `:9002` | A mock third-party provider, a mock upstream authorization server, and a sample agent/client |

You should see each broker port respond as healthy in the summary. If a port is already in
use, stop the process holding it and re-run the command.

:::note
In local development, the consent UI's dev server injects an `X-Remote-User: dev@example.com`
header on every request it proxies to the broker. That simulates the authenticating reverse
proxy you would run in production — the broker never logs users in itself, it trusts the
principal that a proxy in front of it asserts.
:::

## Step 2: Confirm the seeded data

The stack seeds sample agents and services so there is something to delegate to. List them
through the admin API:

```bash
curl http://localhost:14000/api/agents
```

You should see a JSON array of sample agents (for example a weather assistant and a task
manager). These are ready to receive a delegation — you did not have to register anything.

## Step 3: Start the sample agent's authorization request

Open the sample OAuth2 client in your browser:

```text
http://localhost:9002/oauth2/authorize
```

The sample client acts like a real agent asking to act on your behalf. It sends an OAuth2
authorization request to the broker on port 8000. Because you have not granted this agent
anything yet, the broker redirects you to the consent UI to review the request.

You should land on the consent screen at `http://localhost:3000`, showing the agent that is
asking for access.

## Step 4: Review and grant consent

On the consent screen you see the agent's requested **permission sets**, each marked
**mandatory** or **optional**, along with the third-party services they cover. Permission sets
are business-readable bundles of scopes, so you approve capabilities rather than raw provider
scopes.

Select the permission sets to grant, optionally set an expiry, and submit.

You should see the grant confirmed. Because this flow began as an authorization request, the
broker resumes it and redirects you back to the sample client, which now holds broker-issued,
scoped access rather than any third-party credential.

## Step 5: Confirm the delegation was recorded

Back at the admin API, list the agents again and note the one you granted, or view the
broker's activity in the logs:

```bash
just compose-logs
```

You should see the consent request and grant creation in the streamed logs. Press `Ctrl+C` to
stop following them — the stack keeps running.

## Step 6: Tear down

When you are done, stop the whole stack:

```bash
just compose-down
```

All containers stop. Because this tutorial used the in-memory store, the seeded data and your
grant are discarded — the next `just compose-up-detached` starts clean.

## What you did

You ran the complete broker stack locally and walked a full delegation:

- Started the broker (end-user `8000`, admin `14000`), the consent UI (`3000`), and mock
  OAuth2 services, with sample data seeded automatically.
- Kicked off an agent's authorization request from the sample client.
- Reviewed the requested permission sets and consented in the browser — with the broker
  trusting a proxy-injected `X-Remote-User` principal, exactly as it would in production.
- Confirmed the grant was recorded, then tore the stack down.

## Where to go next

- **[Delegation and consent](/docs/concepts/delegation-and-consent)** — the model behind the
  flow you walked: principals, agents, permission sets, grants, and sessions.
- **[Manage agents and services](/docs/guides/manage-agents-and-services)** — register your
  own services, permission sets, and agents through the admin API instead of relying on seed
  data.
- **[Deploy on Kubernetes](/docs/guides/deploy-on-kubernetes)** — take the broker from this
  local stack to a real deployment.
