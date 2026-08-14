ALTER TABLE thirdparty_oauth2_services
    ADD COLUMN protected_resources TEXT[] DEFAULT '{}';

UPDATE thirdparty_oauth2_services AS service
SET protected_resources = COALESCE(
    (
        SELECT array_agg(resource_uri ORDER BY resource_uri)
        FROM service_protected_resources
        WHERE service_id = service.id
    ),
    ARRAY[]::TEXT[]
);

CREATE INDEX idx_thirdparty_oauth2_services_protected_resources
    ON thirdparty_oauth2_services USING GIN (protected_resources);

DROP TABLE service_protected_resources;

ALTER TABLE thirdparty_oauth2_services
    DROP COLUMN version;
