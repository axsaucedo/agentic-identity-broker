CREATE TABLE permission_sets (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_set_name UNIQUE (name)
);

CREATE TABLE permission_set_service_scopes (
    permission_set_id UUID   NOT NULL REFERENCES permission_sets(id) ON DELETE CASCADE,
    service_id        UUID   NOT NULL REFERENCES thirdparty_oauth2_services(id) ON DELETE RESTRICT,
    scopes            JSONB  NOT NULL,
    PRIMARY KEY (permission_set_id, service_id)
);
