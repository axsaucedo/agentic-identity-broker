ALTER TABLE agents ADD COLUMN canonical_id VARCHAR(255);
ALTER TABLE thirdparty_oauth2_services ADD COLUMN canonical_id VARCHAR(255);
ALTER TABLE permission_sets ADD COLUMN canonical_id VARCHAR(255);

CREATE UNIQUE INDEX uq_agents_canonical_id ON agents (canonical_id) WHERE canonical_id IS NOT NULL;
CREATE UNIQUE INDEX uq_thirdparty_oauth2_services_canonical_id ON thirdparty_oauth2_services (canonical_id) WHERE canonical_id IS NOT NULL;
CREATE UNIQUE INDEX uq_permission_sets_canonical_id ON permission_sets (canonical_id) WHERE canonical_id IS NOT NULL;
