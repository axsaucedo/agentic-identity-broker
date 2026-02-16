# KMS Key Management Guide

## Automatic Key Rotation (AWS Managed)

The encryption KMS key has **automatic annual rotation enabled**:

```bash
# Verify rotation is enabled
aws kms describe-key --key-id alias/identity-broker/prod/token-vault-kek \
  --query 'KeyMetadata.KeyRotationEnabled'
# Output: true
```

### What Happens During Rotation

1. AWS generates new key material for the CMK
2. Old key versions remain available for decryption
3. New encryptions use new key material automatically
4. **No downtime, no manual action required**
5. Application requires no code changes

### Key Version Timeline

```
Jan 1, 2024: Key version 1 created
Jan 1, 2025: Automatic rotation → Key version 2 created
Jan 1, 2026: Automatic rotation → Key version 3 created
```

Tokens encrypted with version 1 in 2024 remain decryptable with version 2+ key material via the AWS Encryption SDK.

## Manual Key Rotation (Emergency)

If you need to rotate the key immediately (compromised key material, etc.):

```bash
# Manually rotate KMS key
aws kms rotate-key-forward --key-id arn:aws:kms:eu-central-1:ACCOUNT:key/KEY_ID

# Verify rotation in progress
aws kms describe-key --key-id arn:aws:kms:eu-central-1:ACCOUNT:key/KEY_ID \
  --query 'KeyMetadata.MultiRegionKeyType'
```

## Key Deletion Prevention (30-Day Safety Window)

**Production keys have 30-day pending deletion window**:

```bash
# Attempting to delete production key
aws kms schedule-key-deletion \
  --key-id arn:aws:kms:eu-central-1:ACCOUNT:key/KEY_ID \
  --pending-window-in-days 30

# The key enters "PendingDeletion" state for 30 days
# During this period, the key still works for decryption
# Key material is destroyed after 30 days
```

### Cancel Deletion (if accidental)

```bash
aws kms cancel-key-deletion --key-id arn:aws:kms:eu-central-1:ACCOUNT:key/KEY_ID
```

## Monitoring Key Health

```bash
# Check if key is enabled
aws kms describe-key --key-id alias/identity-broker/prod/token-vault-kek \
  --query 'KeyMetadata.Enabled'
# Output: true

# Check key state
aws kms describe-key --key-id alias/identity-broker/prod/token-vault-kek \
  --query 'KeyMetadata.KeyState'
# Output: Enabled

# View key rotation status
aws kms get-key-rotation-status --key-id alias/identity-broker/prod/token-vault-kek
# Output:
# {
#     "KeyRotationEnabled": true
# }
```

## CloudWatch Alarms

Monitor key health via CloudWatch alarms:

```bash
# Check alarm status
aws cloudwatch describe-alarms \
  --alarm-names IdentityBroker-Encryption-prod-KMS-Throttle \
  --query 'MetricAlarms[0].StateValue'

# View recent alarm history
aws cloudwatch describe-alarm-history \
  --alarm-name IdentityBroker-Encryption-prod-KMS-Throttle \
  --max-records 10
```

## Troubleshooting

### Tokens Can't Be Decrypted

```bash
# Check if KMS key is in valid state
aws kms describe-key --key-id alias/identity-broker/prod/token-vault-kek \
  --query 'KeyMetadata.[KeyState, Enabled, PendingDeletion]'

# Check if IAM role has permission
aws iam get-role-policy --role-name IdentityBrokerEncryption-prod \
  --policy-name EncryptionPolicy

# Check CloudTrail for KMS API errors
aws cloudtrail lookup-events \
  --lookup-attributes AttributeKey=ResourceName,AttributeValue=IdentityBrokerEncryption-prod \
  --max-results 10
```

### KMS Throttling Errors

If KMS reports throttling:

```bash
# Check KMS request rate
aws cloudwatch get-metric-statistics \
  --namespace AWS/KMS \
  --metric-name UserErrorCount \
  --start-time 2026-02-16T00:00:00Z \
  --end-time 2026-02-16T23:59:59Z \
  --period 3600 \
  --statistics Sum

# Request quota increase
aws service-quotas request-service-quota-increase \
  --service-code kms \
  --quota-code L-8132254D \  # GenerateDataKey rate
  --desired-value 200
```

## Key Rotation Backward Compatibility

The AWS Encryption SDK hierarchical keyring handles key rotation transparently:
- Old key versions remain available for decryption
- New encryptions use new key material automatically
- No application code changes needed
- DEK envelope includes key version info for proper unwrapping

**Verification**: See `tests/integration/encryption_vault_keyring_test.go`
- `TestKMSKeyRotationBackwardCompatibility` - Validates tokens remain readable post-rotation
- `TestKeyVersionTransparency` - Validates key version info stored in envelope

Run: `just test -run "TestKMSKeyRotationBackwardCompatibility"`

## References

- [AWS KMS Key Rotation](https://docs.aws.amazon.com/kms/latest/developerguide/rotate-keys.html)
- [AWS Encryption SDK Keyring](https://docs.aws.amazon.com/encryption-sdk/latest/developer-guide/use-hierarchical-keyring.html)
- [CloudWatch Alarms for KMS](https://docs.aws.amazon.com/kms/latest/developerguide/monitoring-cloudwatch.html)
