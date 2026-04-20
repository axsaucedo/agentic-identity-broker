ALTER TABLE authorization_codes
    ADD COLUMN client_id VARCHAR(255) NOT NULL DEFAULT '';
