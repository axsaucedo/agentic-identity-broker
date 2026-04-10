DROP INDEX IF EXISTS idx_agents_client_id;
CREATE UNIQUE INDEX idx_agents_client_id ON agents(client_id);
ALTER TABLE agents ADD CONSTRAINT agents_client_id_key UNIQUE USING INDEX idx_agents_client_id;
