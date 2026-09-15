CREATE TABLE agent_client_uris (
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    client_uri TEXT NOT NULL,
    UNIQUE(client_uri)
);

CREATE INDEX idx_agent_client_uris_agent_id ON agent_client_uris(agent_id);
