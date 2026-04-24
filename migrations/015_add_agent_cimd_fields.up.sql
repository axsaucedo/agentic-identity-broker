ALTER TABLE agents ADD COLUMN auth_method TEXT;
ALTER TABLE agents ADD COLUMN jwks_uri TEXT;
ALTER TABLE agents ADD COLUMN cimd_client_name TEXT;
ALTER TABLE agents ADD COLUMN cimd_logo_uri TEXT;

CREATE TABLE agent_client_uris (
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    client_uri TEXT NOT NULL,
    UNIQUE(client_uri)
);
