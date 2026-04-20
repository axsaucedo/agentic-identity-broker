CREATE TABLE broker_client_credentials (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id      UUID NOT NULL UNIQUE REFERENCES agents(id) ON DELETE CASCADE,
    client_id VARCHAR(255) NOT NULL UNIQUE,
    secret_hash   TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rotated_at    TIMESTAMPTZ
);

CREATE INDEX idx_broker_client_credentials_client_id ON broker_client_credentials(client_id);
