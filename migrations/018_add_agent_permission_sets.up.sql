ALTER TABLE agents ADD COLUMN permission_sets JSONB DEFAULT NULL;
CREATE INDEX idx_agents_permission_sets ON agents USING GIN (permission_sets);
