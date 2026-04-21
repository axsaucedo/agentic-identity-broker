-- Existing credentials with broker_xxx client_ids are invalid; remove before schema change.
DELETE FROM broker_client_credentials;

-- The explicit index is redundant with the UNIQUE constraint; drop it before type change.
DROP INDEX IF EXISTS idx_broker_client_credentials_client_id;

-- Promote client_id to UUID (same type as agents.id) so a FK can be declared.
ALTER TABLE broker_client_credentials
    ALTER COLUMN client_id TYPE UUID USING client_id::uuid;

-- client_id IS the agent FK — bind it and cascade deletes.
ALTER TABLE broker_client_credentials
    ADD CONSTRAINT fk_broker_credentials_agent
        FOREIGN KEY (client_id) REFERENCES agents(id) ON DELETE CASCADE;

-- agent_id is now redundant: client_id carries the same identity.
ALTER TABLE broker_client_credentials
    DROP COLUMN agent_id;
