DROP TABLE IF EXISTS authorization_sessions;
DROP INDEX IF EXISTS idx_agent_client_uris_agent_id;
DROP TABLE agent_client_uris;
ALTER TABLE agents DROP COLUMN cimd_redirect_uris;
ALTER TABLE agents DROP COLUMN cimd_logo_uri;
ALTER TABLE agents DROP COLUMN cimd_client_name;
ALTER TABLE agents DROP COLUMN jwks_uri;
ALTER TABLE agents DROP COLUMN auth_method;
