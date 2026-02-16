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
  export IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN="arn:aws:kms:eu-central-1:ACCOUNT:key/KEY_ID"
  export IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_TABLE_NAME="IdentityBrokerBranchKeys-prod"
  export IDENTITY_BROKER_ENCRYPTION_AWS_IAM_ROLE_ARN="arn:aws:iam::ACCOUNT:role/IdentityBrokerEncryptionRole-prod"
  ```

- [ ] **Run smoke test**:
  ```bash
  # Test token encryption and decryption
  curl -X POST http://localhost:8000/api/v1/tokens \
    -H "X-Remote-User: testuser@example.com" \
    -H "Content-Type: application/json" \
    -d '{"service_id": "test-service"}'

  # Verify token can be decrypted
  curl -X GET http://localhost:8000/api/v1/tokens/TOKEN_ID \
    -H "X-Remote-User: testuser@example.com"
  ```

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
