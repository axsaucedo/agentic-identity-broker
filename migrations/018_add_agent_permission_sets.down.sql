DROP INDEX IF EXISTS idx_agents_permission_sets;
ALTER TABLE agents DROP COLUMN IF EXISTS permission_sets;
