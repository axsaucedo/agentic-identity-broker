-- Add oauth2_flavor column to thirdparty_oauth2_services table
-- This column identifies the credential format for each configured service

ALTER TABLE thirdparty_oauth2_services
    ADD COLUMN oauth2_flavor VARCHAR(50) NOT NULL DEFAULT 'standard';

COMMENT ON COLUMN thirdparty_oauth2_services.oauth2_flavor
    IS 'OAuth2 authentication variant: standard (plain client_secret) or google (service account JSON key)';
