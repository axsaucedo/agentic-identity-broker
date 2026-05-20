-- BREAKING: Intentional data loss — all existing user_grants are deleted before the schema change.
-- The old delegated_oauth2_tokens structure (service_id + scopes array) is fundamentally
-- incompatible with the new granted_permission_sets structure (permission_set_id +
-- included_service_ids). An automated backfill is architecturally impossible: the new model
-- requires permission set entities that did not exist before this migration.
--
-- Pre-conditions before applying to any environment with live grant data:
--   1. Ensure migration 017 (permission_sets table) and 018 (agents.permission_sets column)
--      have been applied and permission sets have been configured.
--   2. Notify users that all existing grants will be revoked and re-consent is required.
--   3. If grant history must be preserved for audit, back up the user_grants table first.
--
-- Operators must accept that all existing grants are revoked. Users will need to re-grant
-- consent after the migration completes.
--
-- Note: The down.sql restores the schema only — it does NOT restore grant data.
DELETE FROM user_grants;
ALTER TABLE user_grants ADD COLUMN granted_permission_sets JSONB NOT NULL DEFAULT '[]';
ALTER TABLE user_grants DROP COLUMN IF EXISTS delegated_oauth2_tokens;
CREATE INDEX IF NOT EXISTS idx_grants_permission_sets ON user_grants USING GIN (granted_permission_sets);
