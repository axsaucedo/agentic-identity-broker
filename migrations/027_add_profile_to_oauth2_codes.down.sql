ALTER TABLE authorization_codes
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS display_name;

ALTER TABLE refresh_token_sessions
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS display_name;
