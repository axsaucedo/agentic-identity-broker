-- Drop UNIQUE constraint on agents.client_id to support multi-agent client sharing.
-- Uniqueness is now enforced at the application layer when multi_agent_client is disabled.
ALTER TABLE agents DROP CONSTRAINT IF EXISTS agents_client_id_key;
DROP INDEX IF EXISTS idx_agents_client_id;
-- Recreate index WITHOUT UNIQUE for performance (GetByClientID still used for CEL helper)
CREATE INDEX IF NOT EXISTS idx_agents_client_id ON agents(client_id);
