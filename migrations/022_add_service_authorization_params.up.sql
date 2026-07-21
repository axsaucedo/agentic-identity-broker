ALTER TABLE thirdparty_oauth2_services
    ADD COLUMN authorization_params JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN thirdparty_oauth2_services.authorization_params IS
    'Static provider-defined authorization request parameters managed by administrators.';
