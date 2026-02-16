# DynamoDB Branch Key Cache Recovery

## Point-in-Time Recovery (PITR)

Branch key cache has PITR enabled for production (35-day recovery window).

### Restore Branch Keys from PITR

```bash
# Restore table to specific point in time
aws dynamodb restore-table-to-point-in-time \
  --source-table-name IdentityBrokerBranchKeys-prod \
  --target-table-name IdentityBrokerBranchKeys-prod-restored \
  --restore-date-time 2026-02-15T14:30:00Z \
  --region eu-central-1
```

## When to Use PITR

Use Point-in-Time Recovery in these scenarios:

- **Accidental Data Deletion**: Restore table if branch keys were accidentally deleted
- **Data Corruption**: Recover to a known good state before corruption occurred
- **Configuration Rollback**: Restore to previous state after failed configuration changes
- **Disaster Recovery**: Recover from catastrophic failures within 35-day window

## Recovery Verification

After restoring the table, verify the branch keys are intact:

```bash
# Scan the restored table to verify branch keys
aws dynamodb scan \
  --table-name IdentityBrokerBranchKeys-prod-restored \
  --select COUNT

# Compare item count with original table
aws dynamodb scan \
  --table-name IdentityBrokerBranchKeys-prod \
  --select COUNT

# Sample a few branch keys to verify integrity
aws dynamodb get-item \
  --table-name IdentityBrokerBranchKeys-prod-restored \
  --key '{"BranchKeyId": {"S": "SAMPLE_KEY_ID"}}'
```

## Cutover Procedure

Once the restored table is verified:

```bash
# 1. Stop application instances (prevent writes)
# 2. Rename original table
aws dynamodb update-table \
  --table-name IdentityBrokerBranchKeys-prod \
  --table-name IdentityBrokerBranchKeys-prod-backup

# 3. Rename restored table to production name
aws dynamodb update-table \
  --table-name IdentityBrokerBranchKeys-prod-restored \
  --table-name IdentityBrokerBranchKeys-prod

# 4. Restart application instances
# 5. Monitor encryption/decryption operations
```

## 35-Day Recovery Window

- **Production tables**: PITR enabled with a 35-day recovery window
- **Non-production tables**: PITR disabled (no point-in-time recovery available)
- For tables with PITR enabled, the recovery point can be any second within the window
- PITR does not affect table performance or availability

## Cost Considerations

PITR incurs additional storage costs:
- Continuous backups stored by AWS
- Charged per GB-month of table size
- Restore operations are free (no data transfer charges)

## Monitoring PITR Status

```bash
# Verify PITR is enabled
aws dynamodb describe-continuous-backups \
  --table-name IdentityBrokerBranchKeys-prod

# Output should show:
# {
#     "ContinuousBackupsDescription": {
#         "ContinuousBackupsStatus": "ENABLED",
#         "PointInTimeRecoveryDescription": {
#             "PointInTimeRecoveryStatus": "ENABLED",
#             "EarliestRestorableDateTime": "2026-01-12T14:30:00Z",
#             "LatestRestorableDateTime": "2026-02-16T10:45:00Z"
#         }
#     }
# }
```

## References

- [AWS DynamoDB Point-in-Time Recovery](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/PointInTimeRecovery.html)
- [DynamoDB Backup and Restore](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/BackupRestore.html)
