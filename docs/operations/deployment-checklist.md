# CDK Encryption Infrastructure Deployment Checklist

## Pre-Deployment

### For Kubernetes IRSA Deployments

- [ ] EKS cluster has OIDC provider configured
- [ ] OIDC provider ARN extracted (`aws eks describe-cluster`)
- [ ] Kubernetes namespace and service account name defined
- [ ] Environment set correctly (-c env=prod for production)
- [ ] AWS credentials configured (aws sts get-caller-identity)
- [ ] VPC/network access to KMS and DynamoDB verified
- [ ] IAM permissions to create KMS, DynamoDB, IAM resources
- [ ] CloudFormation template reviewed (cdk diff)
- [ ] Backup/rollback plan documented

### For Legacy Trust Principal Deployments (Deprecated)

> **Note**: Trust principal support is being migrated to IRSA-only. Use IRSA for all Kubernetes deployments.

- [ ] Trust principal ARN prepared (ECS task role, Lambda role, etc.)
- [ ] Environment set correctly (-c env=prod for production)
- [ ] AWS credentials configured
- [ ] CloudFormation template reviewed

## Deployment Steps

### Kubernetes IRSA Deployment (Recommended)

The IRSA pattern uses federated identity via EKS OIDC provider for secure, credential-free AWS access from Kubernetes pods.

```bash
# 1. Extract OIDC provider information
OIDC_ISSUER=$(aws eks describe-cluster \
  --name my-cluster \
  --query 'cluster.identity.oidc.issuer' \
  --output text)

OIDC_ID=$(echo $OIDC_ISSUER | awk -F'/' '{print $NF}')
ACCOUNT_ID=$(aws sts get-caller-identity --query 'Account' --output text)
OIDC_PROVIDER_ARN="arn:aws:iam::${ACCOUNT_ID}:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/${OIDC_ID}"

# 2. Define service account details
export K8S_NAMESPACE="identity-broker"
export K8S_SERVICE_ACCOUNT="broker-sa"

# 3. Synthesize CloudFormation template (review for errors)
cd infra/cdk
npx cdk synth \
  -c env=prod \
  -c oidcProviderArn="${OIDC_PROVIDER_ARN}" \
  -c k8sNamespace="${K8S_NAMESPACE}" \
  -c k8sServiceAccountName="${K8S_SERVICE_ACCOUNT}"

# 4. Preview infrastructure changes
npx cdk diff \
  -c env=prod \
  -c oidcProviderArn="${OIDC_PROVIDER_ARN}" \
  -c k8sNamespace="${K8S_NAMESPACE}" \
  -c k8sServiceAccountName="${K8S_SERVICE_ACCOUNT}"

# 5. Deploy with confirmation prompt
npx cdk deploy \
  -c env=prod \
  -c oidcProviderArn="${OIDC_PROVIDER_ARN}" \
  -c k8sNamespace="${K8S_NAMESPACE}" \
  -c k8sServiceAccountName="${K8S_SERVICE_ACCOUNT}"
```

### Legacy Trust Principal Deployment (Deprecated)

> **Warning**: This approach is deprecated. Migrate to IRSA for Kubernetes deployments.

```bash
# 1. Synthesize CloudFormation template
cd infra/cdk
npx cdk synth -c env=prod -c trustPrincipal=arn:aws:iam::ACCOUNT:role/ROLE_NAME

# 2. Preview infrastructure changes
npx cdk diff -c env=prod -c trustPrincipal=arn:aws:iam::ACCOUNT:role/ROLE_NAME

# 3. Deploy with confirmation prompt
npx cdk deploy -c env=prod -c trustPrincipal=arn:aws:iam::ACCOUNT:role/ROLE_NAME
```

## Post-Deployment Verification

- [ ] **Verify KMS key created**:
  ```bash
  aws kms describe-key --key-id alias/identity-broker/prod/token-vault-kek
  # Expected: KeyState=Enabled, KeyRotationEnabled=true
  ```

- [ ] **Verify DynamoDB table created**:
  ```bash
  aws dynamodb describe-table --table-name IdentityBrokerBranchKeys-prod
  # Expected: TableStatus=ACTIVE, PITR=ENABLED
  ```

- [ ] **Verify IAM role created**:
  ```bash
  aws iam get-role --role-name IdentityBrokerEncryptionRole-prod
  # Expected: Role exists with KMS decrypt permissions
  ```

- [ ] **Extract stack outputs**:
  ```bash
  STACK_NAME="IdentityBrokerEncryption-prod"

  aws cloudformation describe-stacks \
    --stack-name $STACK_NAME \
    --query 'Stacks[0].Outputs' \
    --output table
  ```

### For Kubernetes IRSA Deployments

- [ ] **Verify IAM role trust policy includes federated principal**:
  ```bash
  IAM_ROLE_NAME=$(aws cloudformation describe-stacks \
    --stack-name IdentityBrokerEncryption-prod \
    --query 'Stacks[0].Outputs[?OutputKey==`IamRoleName`].OutputValue' \
    --output text)

  aws iam get-role --role-name $IAM_ROLE_NAME \
    --query 'Role.AssumeRolePolicyDocument' \
    --output json | jq .

  # Expected: Principal.Federated = OIDC provider ARN
  # Expected: Condition.StringEquals includes namespace and service account
  ```

- [ ] **Extract stack outputs for Helm configuration**:
  ```bash
  # Extract outputs needed for Helm values
  export KMS_KEY_ARN=$(aws cloudformation describe-stacks \
    --stack-name IdentityBrokerEncryption-prod \
    --query 'Stacks[0].Outputs[?OutputKey==`EncryptionKeyARN`].OutputValue' \
    --output text)

  export DYNAMODB_TABLE_NAME=$(aws cloudformation describe-stacks \
    --stack-name IdentityBrokerEncryption-prod \
    --query 'Stacks[0].Outputs[?OutputKey==`BranchKeyTableName`].OutputValue' \
    --output text)

  export IAM_ROLE_ARN=$(aws cloudformation describe-stacks \
    --stack-name IdentityBrokerEncryption-prod \
    --query 'Stacks[0].Outputs[?OutputKey==`EncryptionRoleARN`].OutputValue' \
    --output text)

  echo "KMS Key ARN: $KMS_KEY_ARN"
  echo "DynamoDB Table: $DYNAMODB_TABLE_NAME"
  echo "IAM Role ARN: $IAM_ROLE_ARN"
  ```

- [ ] **Create Kubernetes namespace**:
  ```bash
  kubectl create namespace ${K8S_NAMESPACE}
  ```

- [ ] **Deploy with Helm (IRSA configured)**:
  ```bash
  helm install broker ./charts/agentic-identity-broker \
    -n ${K8S_NAMESPACE} \
    --set serviceAccount.create=true \
    --set serviceAccount.name=${K8S_SERVICE_ACCOUNT} \
    --set serviceAccount.irsa.enabled=true \
    --set serviceAccount.irsa.roleArn="${IAM_ROLE_ARN}" \
    --set broker.extraConfig.encryption.aws_kms.key_arn="${KMS_KEY_ARN}" \
    --set broker.extraConfig.encryption.aws_kms.dynamodb_table_name="${DYNAMODB_TABLE_NAME}" \
    --set storage.type=postgres \
    --set postgresql.external.enabled=true
  ```

- [ ] **Verify ServiceAccount has IRSA annotation**:
  ```bash
  kubectl get serviceaccount ${K8S_SERVICE_ACCOUNT} \
    -n ${K8S_NAMESPACE} \
    -o jsonpath='{.metadata.annotations.iam\.amazonaws\.com/role}'

  # Expected: arn:aws:iam::ACCOUNT:role/IdentityBrokerEncryptionRole-prod
  ```

- [ ] **Verify pod can assume IAM role**:
  ```bash
  POD_NAME=$(kubectl get pods -n ${K8S_NAMESPACE} \
    -l app.kubernetes.io/name=agentic-identity-broker \
    --output jsonpath='{.items[0].metadata.name}')

  kubectl exec -n ${K8S_NAMESPACE} $POD_NAME -- env | grep AWS_ROLE_ARN
  kubectl logs -n ${K8S_NAMESPACE} $POD_NAME | grep -i "credential\|kms\|dynamodb"

  # Expected: AWS_ROLE_ARN environment variable set
  # Expected: No credential or permission errors in logs
  ```

- [ ] **Test encryption functionality**:
  ```bash
  # Port-forward to test
  kubectl port-forward -n ${K8S_NAMESPACE} svc/broker-agentic-identity-broker 8000:8000 &

  # Test health
  curl http://localhost:8000/health

  # Test OAuth2 session (tests KMS encryption)
  curl -X GET http://localhost:8000/api/third-party/sessions \
    -H "X-Remote-User: testuser@example.com"
  ```

### For Legacy Trust Principal Deployments

- [ ] **Configure application environment variables**:
  ```bash
  # Extract stack outputs and set environment variables
  STACK_NAME="IdentityBrokerEncryption-prod"

  export IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN=$(aws cloudformation describe-stacks \
    --stack-name $STACK_NAME \
    --query 'Stacks[0].Outputs[?OutputKey==`EncryptionKeyARN`].OutputValue' --output text)

  export IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_TABLE_NAME=$(aws cloudformation describe-stacks \
    --stack-name $STACK_NAME \
    --query 'Stacks[0].Outputs[?OutputKey==`BranchKeyTableName`].OutputValue' --output text)

  # Optional: Use role assumption if required for cross-account access
  export IDENTITY_BROKER_ENCRYPTION_AWS_KMS_ASSUME_ROLE_ARN=$(aws cloudformation describe-stacks \
    --stack-name $STACK_NAME \
    --query 'Stacks[0].Outputs[?OutputKey==`EncryptionRoleARN`].OutputValue' --output text)
  ```

  This automatically populates the required variables from CDK stack outputs.

- [ ] **Run smoke test** (OAuth2 session management):
  ```bash
  # List OAuth2 sessions (verify encryption is working)
  curl -X GET http://localhost:8000/api/third-party/sessions \
    -H "X-Remote-User: testuser@example.com"

  # Verify session details can be retrieved
  curl -X GET http://localhost:8000/api/third-party/{serviceId}/session \
    -H "X-Remote-User: testuser@example.com"
  ```

  **Alternative**: Create a full OAuth2 session flow via `/api/third-party/{serviceId}/oauth2/authorize` and `/api/third-party/{serviceId}/oauth2/callback` endpoints.

- [ ] **Verify CloudWatch alarms created and subscribed**:
  ```bash
  aws cloudwatch describe-alarms \
    --alarm-name-prefix IdentityBroker-Encryption-prod

  # Subscribe to alarms
  aws sns subscribe \
    --topic-arn arn:aws:sns:eu-central-1:ACCOUNT:IdentityBroker-Encryption-prod-KMS-Throttle \
    --protocol email \
    --notification-endpoint ops-team@example.com
  ```

## Rollback Procedures

If deployment fails or issues are discovered:

### Option 1: Automatic Rollback (Default)
```bash
# CloudFormation automatically rolls back failed stacks
# No manual action required - monitors stack events
aws cloudformation describe-stack-events \
  --stack-name IdentityBrokerEncryption-prod
```

### Option 2: Manual Rollback
```bash
# Cancel in-progress deployment
aws cloudformation cancel-update-stack \
  --stack-name IdentityBrokerEncryption-prod

# Verify stack returns to previous state
aws cloudformation describe-stacks \
  --stack-name IdentityBrokerEncryption-prod \
  --query 'Stacks[0].StackStatus'
```

### Option 3: Delete Stack (Emergency Only)
```bash
# WARNING: This deletes all encryption infrastructure
# Only use if stack is unrecoverable or in non-production
cd infra/cdk
npx cdk destroy -c env=prod

# Note: KMS keys have 30-day pending deletion window
# They can be recovered during this period if needed
```

## AWS KMS Configuration Options

All AWS KMS configuration can be set via environment variables:

### Key Encryption Setup
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN` - **Required**. KMS CMK ARN for envelope encryption (from CDK stack output)
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_TABLE_NAME` - DynamoDB table name for branch key caching (from CDK stack output)
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_BRANCH_KEY_TTL` - TTL for cached branch keys (default: "1h")

### DynamoDB Configuration
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_REGION` - AWS region for DynamoDB operations
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_READ_TIMEOUT` - DynamoDB read timeout (default: "5s")
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_WRITE_TIMEOUT` - DynamoDB write timeout (default: "5s")
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_ENDPOINT` - Custom DynamoDB endpoint (for LocalStack testing)

### AWS SDK Configuration
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_REGION` - AWS region for KMS operations
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_ENDPOINT` - Custom KMS endpoint URL (for LocalStack testing)
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_PROFILE` - AWS profile for credentials (~/.aws/credentials)
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_ACCESS_KEY_ID` - Static AWS access key ID (for CI/testing)
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_SECRET_ACCESS_KEY` - Static AWS secret access key (for CI/testing)
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_ASSUME_ROLE_ARN` - IAM role ARN to assume for operations (from CDK stack output)
- `IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DISABLE_SSL` - Disable SSL verification (development only, never in production)

## Success Criteria

### Before Production Deployment

- [ ] All HIGH priority fixes implemented and tested
- [ ] Production trust principal enforced (panic if missing)
- [ ] Environment validation prevents typos (dev/staging/prod only)
- [ ] KMS key rotation backward compatibility tested
- [ ] Security review completed
- [ ] Load testing performed

### Before General Availability (GA) Release

- [ ] All MEDIUM priority fixes implemented
- [ ] CloudWatch alarms configured and monitored
- [ ] Operational runbooks complete and tested
- [ ] Cost tracking and budgets configured
- [ ] Disaster recovery procedures tested
- [ ] On-call team trained on operations
- [ ] Documentation published and accessible

## Monitoring and Alerting

After deployment, ensure continuous monitoring:

```bash
# View CloudWatch dashboard
aws cloudwatch get-dashboard \
  --dashboard-name IdentityBroker-Encryption-prod

# Monitor KMS API call rate
aws cloudwatch get-metric-statistics \
  --namespace AWS/KMS \
  --metric-name ApiCallCount \
  --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Sum

# Monitor DynamoDB read/write capacity
aws cloudwatch get-metric-statistics \
  --namespace AWS/DynamoDB \
  --metric-name ConsumedReadCapacityUnits \
  --dimensions Name=TableName,Value=IdentityBrokerBranchKeys-prod \
  --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Sum
```

## Cost Estimation

Expected monthly costs for production:

- **KMS Key**: $1/month + $0.03 per 10,000 requests
- **DynamoDB Table**: Pay-per-request pricing (~$0.25 per million reads)
- **DynamoDB PITR**: ~$0.20 per GB-month of table size
- **CloudWatch Alarms**: $0.10 per alarm per month
- **CloudWatch Dashboard**: $3/month per dashboard

**Total estimated cost**: $10-50/month depending on request volume

## References

- [Kubernetes IRSA Deployment Guide](../deployment/kubernetes-irsa.md) - Complete IRSA deployment walkthrough
- [AWS CDK Deployment Guide](https://docs.aws.amazon.com/cdk/latest/guide/home.html)
- [CloudFormation Stack Operations](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-console-view-stack-data-resources.html)
- [EKS IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [KMS Key Management](./kms-key-management.md)
- [DynamoDB Recovery Procedures](./dynamodb-recovery.md)
