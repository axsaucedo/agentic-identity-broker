# Infrastructure — AWS CDK (`infra/`)

**Prefer retrieval-led reasoning. Read `stack.go` and ADR 010 before making infrastructure changes.**

## Overview

AWS CDK (Go SDK) infrastructure-as-code for the encryption resources required by the Token Vault's three-layer envelope encryption (KEK → Branch Key → DEK). This is a **separate Go module** (`infra/cdk/go.mod`) within the monorepo.

**ADR**: [010-cdk-encryption-infrastructure.md](../adrs/010-cdk-encryption-infrastructure.md) — binding decision for all infrastructure choices.

## Module Structure

```
infra/cdk/
  cdk.json           CDK app config (default env: "dev")
  go.mod             Separate Go module (Go 1.25.6, aws-cdk-go/awscdk/v2)
  main.go            CDK app entrypoint — env parsing, validation, stack instantiation
  stack.go           EncryptionStack definition — all resources
  stack_test.go      CDK assertions unit tests (22+ tests)
  cdk.out/           Synthesized CloudFormation (gitignored)
```

## EncryptionStack Resources

The single `NewEncryptionStack()` function provisions:

| Resource | Type | Purpose |
|---|---|---|
| **KMS CMK** | `AWS::KMS::Key` | Symmetric KEK for hierarchical keyring. Annual rotation enabled. Alias: `alias/agentic-identity-broker/{env}/token-vault-kek` |
| **DynamoDB Table** | `AWS::DynamoDB::Table` | Branch key cache. Schema: `branch-key-id` (S, HASH) + `type` (S, RANGE). PAY_PER_REQUEST billing. Name: `AgenticIdentityBrokerBranchKeys-{env}` |
| **IAM Role** | `AWS::IAM::Role` | Least-privilege KMS + DynamoDB access via IRSA. Name: `AgenticIdentityBrokerEncryptionRole-{env}` |
| **CloudWatch Alarms** | `AWS::CloudWatch::Alarm` | KMS throttle + error detection |
| **CloudWatch Dashboard** | `AWS::CloudWatch::Dashboard` | KMS API ops + DynamoDB cache metrics |

### Environment Parameterization

Two environments: `test`, `prod` (or `production`, normalized to `prod`).

| Aspect | Production | Non-Production |
|---|---|---|
| KMS RemovalPolicy | RETAIN | DESTROY |
| KMS PendingDeletion | 30 days | 7 days |
| DynamoDB PITR | Enabled | Disabled |
| DynamoDB DeletionProtection | On | Off |
| IRSA parameters | **Required** (panics if missing) | Optional (defaults to account root) |

### Stack Outputs → Environment Variables

| Stack Output | Environment Variable |
|---|---|
| `EncryptionKeyARN` | `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN` |
| `BranchKeyTableName` | `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_TABLE_NAME` |
| `EncryptionRoleARN` | `IDENTITY_BROKER_ENCRYPTION_IAM_ROLE_ARN` |

### IRSA (IAM Roles for Service Accounts)

Production deployments use federated web identity trust via EKS OIDC provider. Props:
- `OIDCProviderArn` — full OIDC provider ARN
- `K8sNamespace` — Kubernetes namespace
- `K8sServiceAccountName` — service account name

Non-production defaults to `AccountRootPrincipal` for local dev convenience.

## IAM Permissions (Least Privilege)

**KMS**: Encrypt, Decrypt, GenerateDataKey, GenerateDataKeyWithoutPlaintext, ReEncryptFrom, ReEncryptTo, DescribeKey, CreateGrant (conditioned on `kms:GrantIsForAWSResource`)

**DynamoDB**: GetItem, PutItem, Query, UpdateItem, DeleteItem, DescribeTable — scoped to branch key table ARN only.

## Development Commands

```bash
just cdk-test           # Run CDK unit tests (assertions)
just cdk-synth          # Synthesize CloudFormation template
just cdk-deploy         # Deploy stack (default: test)
```

Manual CDK commands:
```bash
cd infra/cdk
cdk deploy -c env=test
cdk deploy -c env=prod \
  -c oidcProviderArn=arn:aws:iam::ACCOUNT:oidc-provider/... \
  -c k8sNamespace=identity-broker \
  -c k8sServiceAccountName=agentic-identity-broker
cdk diff                # Preview changes
cdk destroy             # Tear down (non-prod only)
```

## Alignment Rules

- Stack resources **must** match what the encryption adapter at `internal/adapters/encryption/aws/` expects. The adapter reads `EncryptionConfig` env vars that map to stack outputs.
- DynamoDB table schema **must** match the AWS Encryption SDK KeyStore spec: `branch-key-id` (S) + `type` (S). Changing this breaks the hierarchical keyring.
- All resources tagged with: `Project=agentic-identity-broker`, `Component=encryption-vault`, `Environment={env}`, `ManagedBy=aws-cdk`. Custom tags via `EncryptionStackProps.Tags`.
- Keep documentation in sync: `docs/operations/deployment-checklist.md`, `docs/deployment/kubernetes-irsa.md`, and ADR 010.

## Testing

Tests use CDK `assertions.Template_FromStack()` to validate synthesized CloudFormation. Use `createTestStack()` helper for consistent test setup. Test categories:
- KMS key properties (symmetric, rotation, removal policy, pending window)
- DynamoDB schema, billing, PITR, deletion protection
- IAM role trust policy (IRSA conditions), permission statements
- Stack outputs and export names
- Tag propagation
- Environment parameterization (dev vs prod differences)
- IRSA validation (production panics on missing params)
