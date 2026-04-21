ALTER TABLE broker_client_credentials
    DROP CONSTRAINT IF EXISTS fk_broker_credentials_agent;

ALTER TABLE broker_client_credentials
    ADD COLUMN agent_id UUID;

UPDATE broker_client_credentials SET agent_id = client_id;

ALTER TABLE broker_client_credentials
    ALTER COLUMN agent_id SET NOT NULL;

ALTER TABLE broker_client_credentials
    ADD CONSTRAINT broker_client_credentials_agent_id_key UNIQUE (agent_id),
    ADD CONSTRAINT broker_client_credentials_agent_id_fkey
        FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE;

ALTER TABLE broker_client_credentials
    ALTER COLUMN client_id TYPE VARCHAR(255) USING client_id::text;

CREATE INDEX idx_broker_client_credentials_client_id ON broker_client_credentials(client_id);
