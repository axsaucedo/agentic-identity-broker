CREATE INDEX CONCURRENTLY idx_agent_client_uris_pattern ON agent_client_uris (client_uri) WHERE client_uri LIKE '%*%';
