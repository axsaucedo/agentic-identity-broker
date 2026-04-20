ALTER TABLE authorization_codes
    ADD COLUMN broker_client_id VARCHAR(255) NOT NULL;
