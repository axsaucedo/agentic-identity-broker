---
title: "Configure encryption at rest"
description: "Configure the broker's mandatory encryption backend — a raw AES-256 key for development or an AWS KMS hierarchical keyring for production — and the separate JWE signing key for state and session tokens."
---

# Configure encryption at rest

The broker stores third-party OAuth2 access and refresh tokens, provider client
secrets, and signing-key private material **encrypted at rest**. Encryption is
mandatory: the broker refuses to start unless exactly one encryption backend is
configured, and it never falls back to plaintext. This page shows how to
configure each backend and how production pods reach the keys.

For the model behind it — the KEK → branch key → DEK envelope and per-service
context binding — see [encryption at rest](/docs/concepts/encryption).

## Choose a backend

Configure exactly one backend under the `encryption` key. Configuring both, or
neither, is a configuration error and the broker exits at startup.

| Backend | Key | Use for |
|---|---|---|
| `encryption.memory` | A raw base64 AES-256 key you supply | Development, testing, CI |
| `encryption.aws_kms` | An AWS KMS customer-managed key + a DynamoDB branch-key table | Production |

The memory backend keeps key material in the process environment, has no key
rotation, and no hardware-backed key storage or centralized audit. Use it only
outside production.

## Development: a raw AES-256 key

The memory backend encrypts with a single 32-byte AES-256 key you provide,
base64-encoded.

### Generate a key

```bash
openssl rand -base64 32
```

### Configure it

Reference the key through an environment variable rather than writing it into
the file, so the key never lands in version control:

```yaml
encryption:
  memory:
    # Base64-encoded 32-byte AES-256 key, injected from the environment.
    raw_key: "${IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY}"
```

The broker performs `${VAR}` substitution in YAML, so the value comes from the
environment at load time:

```bash
export IDENTITY_BROKER_ENCRYPTION_MEMORY_RAW_KEY="$(openssl rand -base64 32)"
```

:::warning
Never commit a key or hardcode it in a config file. The memory-backend key is
visible to anything that can read the process environment, and it does not
rotate. Treat it as a development convenience, not a production control.
:::

## Production: AWS KMS hierarchical keyring

In production the broker uses an AWS KMS customer-managed key as the root key
encryption key (KEK) and a DynamoDB table to cache intermediate branch keys.
This is the hierarchical keyring: the KMS key is called once per service and TTL
window rather than on every token operation, which keeps latency and KMS costs
low while every data key stays wrapped by KMS.

### AWS resources you need

- **A customer-managed KMS key (CMK)** — the root key that wraps branch keys.
  Key rotation is a property of the KMS key itself, managed in AWS.
- **A DynamoDB table** — caches the branch keys the keyring derives, so the
  broker avoids a KMS round trip on every encryption.

You can provision both consistently with the project's AWS CDK stack in
`infra/cdk`, which creates the KMS key, the DynamoDB table with the schema the
AWS Encryption SDK expects, and a least-privilege IAM role in one deployment.
The [deployment checklist](/docs/operations/deployment-checklist) walks through
running it and reading back the resource ARNs.

### Configure it

```yaml
encryption:
  aws_kms:
    # Customer-managed KMS key ARN.
    # Format: arn:aws:kms:<region>:<account-id>:key/<key-id>
    key_arn: "${IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN}"

    # DynamoDB table that caches branch keys.
    dynamodb_table_name: "IdentityBrokerEncryptionBranchKeys"

    # How long a cached branch key lives before regeneration (1m–24h).
    # Shorter = faster key rotation but more KMS calls.
    branch_key_ttl: "1h"

    # Optional: DynamoDB region, if it differs from the default SDK region.
    # dynamodb_region: "eu-central-1"
    # Optional: per-call timeout for AWS service calls.
    # dynamodb_timeout: "5s"
```

Set the region and structured JSON logging alongside it for a production
deployment; see the full schema in [configuration](/docs/configuration).

### IAM permissions the broker needs

The identity the broker runs as must be allowed to use the KMS key and to
read and write the branch-key table:

- **KMS**: encrypt, decrypt, and generate data keys with the customer-managed
  key (the keyring also uses describe and grant operations against that key).
- **DynamoDB**: read and write items in the branch-key table.

The CDK stack provisions an IAM role scoped to exactly these actions on exactly
these resources. Scope any hand-written policy to the specific key ARN and table
— do not grant account-wide KMS or DynamoDB access.

### How pods reach the keys

On EKS, grant pods access with **IAM Roles for Service Accounts (IRSA)** rather
than static AWS credentials: annotate the broker's service account with the IAM
role that can use the KMS key and the DynamoDB table, and the AWS SDK obtains
and rotates temporary credentials automatically. The
[Kubernetes deployment guide](/docs/guides/deploy-on-kubernetes#grant-aws-access-with-irsa)
covers the operator-level steps.

## Configure the JWE signing key

Separate from encryption at rest, the broker seals short-lived state and
consent-session tokens (the ephemeral tokens that bind an OAuth2 callback to its
initiating request and prevent consent-screen spoofing) in encrypted JWE tokens.
These use their own key: `third_party_oauth2.jwe_signing_key`, a base64-encoded
32-byte key generated the same way:

```yaml
third_party_oauth2:
  # Base64-encoded 32-byte key. Generate with: openssl rand -base64 32
  jwe_signing_key: "${IDENTITY_BROKER_THIRD_PARTY_OAUTH2_JWE_SIGNING_KEY}"
```

Provide this key in every environment. Inject it from a secret exactly as you do
the encryption key — never commit it.

## Verify

At startup the broker logs which encryption backend it initialized. Sensitive
key material is redacted in logs, so you see the backend and its parameters but
never the key itself.

- **Success**: the broker starts and logs the active backend (memory, or AWS KMS
  with the configured key ARN and table).
- **Failure closed**: a missing, duplicated, or malformed encryption backend
  makes the broker exit at startup rather than run without protection. On
  Kubernetes this surfaces as a pod in `CrashLoopBackOff`; check the logs for
  the encryption error.

To confirm end to end that KMS encryption works, create a third-party OAuth2
session and read it back — storing and retrieving the encrypted tokens exercises
the full keyring. The
[deployment checklist](/docs/operations/deployment-checklist) includes this
smoke test.

## Related

- [Encryption at rest](/docs/concepts/encryption) — the envelope model and
  per-service context binding.
- [Deploy on Kubernetes](/docs/guides/deploy-on-kubernetes) — wiring the KMS key,
  DynamoDB table, and IRSA role into the Helm release.
- [Configuration](/docs/configuration) — the full `encryption` and
  `third_party_oauth2` schema.
- [Deployment checklist](/docs/operations/deployment-checklist) — provisioning
  and verifying the encryption infrastructure.
