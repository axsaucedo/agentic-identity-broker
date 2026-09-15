ALTER TABLE thirdparty_oauth2_services
    ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

CREATE TABLE service_protected_resources (
    resource_uri TEXT PRIMARY KEY,
    service_id UUID NOT NULL REFERENCES thirdparty_oauth2_services(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_service_protected_resources_service_id
    ON service_protected_resources (service_id);

DO $$
DECLARE
    invalid_resources TEXT;
    colliding_resources TEXT;
BEGIN
    SELECT string_agg(resource_uri, ', ' ORDER BY resource_uri)
    INTO invalid_resources
    FROM (
        SELECT resource_uri
        FROM thirdparty_oauth2_services,
             unnest(COALESCE(protected_resources, ARRAY[]::TEXT[])) AS resource_uri
        WHERE resource_uri !~ '^[a-zA-Z][a-zA-Z0-9+.\-]*://[^/?#]+'
           OR regexp_replace(
               regexp_replace(resource_uri, '^[^:]+://[^/?#]*', ''),
               '[?#].*$',
               ''
           ) ~ '/$'
    ) AS invalid;

    IF invalid_resources IS NOT NULL THEN
        RAISE EXCEPTION 'non-canonical protected resource: %', invalid_resources;
    END IF;

    SELECT string_agg(resource_uri, ', ' ORDER BY resource_uri)
    INTO colliding_resources
    FROM (
        SELECT resource_uri
        FROM thirdparty_oauth2_services,
             unnest(COALESCE(protected_resources, ARRAY[]::TEXT[])) AS resource_uri
        GROUP BY resource_uri
        HAVING COUNT(*) > 1
    ) AS collisions;

    IF colliding_resources IS NOT NULL THEN
        RAISE EXCEPTION 'protected resource collision: %', colliding_resources;
    END IF;
END
$$;

INSERT INTO service_protected_resources (resource_uri, service_id)
SELECT resource_uri, id
FROM thirdparty_oauth2_services,
     unnest(COALESCE(protected_resources, ARRAY[]::TEXT[])) AS resource_uri;

DROP INDEX IF EXISTS idx_thirdparty_oauth2_services_protected_resources;

ALTER TABLE thirdparty_oauth2_services
    DROP COLUMN protected_resources;
