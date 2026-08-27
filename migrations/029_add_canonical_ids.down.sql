DROP INDEX IF EXISTS uq_permission_sets_canonical_id;
DROP INDEX IF EXISTS uq_thirdparty_oauth2_services_canonical_id;
DROP INDEX IF EXISTS uq_agents_canonical_id;

ALTER TABLE permission_sets DROP COLUMN canonical_id;
ALTER TABLE thirdparty_oauth2_services DROP COLUMN canonical_id;
ALTER TABLE agents DROP COLUMN canonical_id;
