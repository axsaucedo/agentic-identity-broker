---
title: "Contributing"
description: "How to get involved with the open-source Agentic Identity Broker project: reporting issues, starting discussions, opening pull requests, and the project's governance principles."
---

# Contributing

The Agentic Identity Broker is an open-source project, released under the license in its
repository. Contributions — bug reports, discussions, and pull requests — are welcome.

The authoritative process lives in the repository. This page is an orientation; the source
documents are:

- **[CONTRIBUTING.md](https://github.com/zalando-incubator/agentic-identity-broker/blob/main/CONTRIBUTING.md)**
  — the full contribution workflow.
- **[CODE_OF_CONDUCT.md](https://github.com/zalando-incubator/agentic-identity-broker/blob/main/CODE_OF_CONDUCT.md)**
  — the community standards every participant agrees to.

## Ways to get involved

- **Report a bug.** Open a GitHub issue with clear reproduction steps.
- **Propose a feature or ask a question.** Start a GitHub discussion first, so direction can
  be agreed before code is written.
- **Open a pull request.** Fork the repository, work on a branch, and open a PR with a clear
  description, the rationale for the change, and how you tested it. Reference any related
  issues.

:::warning
Never open a public issue for a security vulnerability. Report it privately through the
process on the [Security posture](/docs/resources/security) page.
:::

## Governance principles

Contributions follow a specification-driven approach with a few standing principles:

- **API-first.** Public APIs are documented in their OpenAPI specification before they are
  implemented.
- **Security-first.** Security controls are enabled by default, never optional. Input is
  validated at system boundaries and secrets are never committed.
- **Tests required.** New functionality ships with tests, and the project's checks and
  verification gate must pass before a pull request is merged.

## Code of conduct

The project follows the Contributor Covenant. By participating you agree to uphold it —
using welcoming, inclusive language and treating other community members with respect. Report
unacceptable behavior to the maintainers as described in the
[code of conduct](https://github.com/zalando-incubator/agentic-identity-broker/blob/main/CODE_OF_CONDUCT.md).

## Before you start

If you want to understand the system before contributing, [get started](/docs/get-started) by
running the stack locally and walking a delegation end to end.
