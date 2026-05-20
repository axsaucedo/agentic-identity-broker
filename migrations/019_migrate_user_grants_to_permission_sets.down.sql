DROP INDEX IF EXISTS idx_grants_permission_sets;
ALTER TABLE user_grants ADD COLUMN delegated_oauth2_tokens JSONB NOT NULL DEFAULT '[]';
ALTER TABLE user_grants DROP COLUMN IF EXISTS granted_permission_sets;
