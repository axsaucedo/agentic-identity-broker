ALTER TABLE authorization_codes
    ADD COLUMN email VARCHAR(255),
    ADD COLUMN display_name TEXT NOT NULL DEFAULT '';

ALTER TABLE refresh_token_sessions
    ADD COLUMN email VARCHAR(255),
    ADD COLUMN display_name TEXT NOT NULL DEFAULT '';
