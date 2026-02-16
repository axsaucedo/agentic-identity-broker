# CDK Encryption Infrastructure Deployment Checklist

## Pre-Deployment

- [ ] Environment set correctly (-c env=prod for production)
- [ ] Trust principal specified (-c trustPrincipal=arn:...)
- [ ] AWS credentials configured (aws sts get-caller-identity)
- [ ] VPC/network access to KMS and DynamoDB verified
- [ ] IAM permissions to create KMS, DynamoDB, IAM resources
- [ ] CloudFormation template reviewed (cdk diff)
- [ ] Backup/rollback plan documented

## Deployment Steps

```bash
# 1. Synthesize CloudFormation template (review for errors)
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
  aws cloudformation describe-stacks \
    --stack-name IdentityBrokerEncryption-prod \
    --query 'Stacks[0].Outputs'
  ```

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

- [AWS CDK Deployment Guide](https://docs.aws.amazon.com/cdk/latest/guide/home.html)
- [CloudFormation Stack Operations](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-console-view-stack-data-resources.html)
- [KMS Key Management](./kms-key-management.md)
- [DynamoDB Recovery Procedures](./dynamodb-recovery.md)
