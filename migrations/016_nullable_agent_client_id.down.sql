UPDATE agents SET client_id = id::text WHERE client_id IS NULL;
ALTER TABLE agents ALTER COLUMN client_id SET NOT NULL;
