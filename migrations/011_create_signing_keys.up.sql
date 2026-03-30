CREATE TABLE signing_keys (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kid                   VARCHAR(255) NOT NULL UNIQUE,
    algorithm             VARCHAR(10) NOT NULL DEFAULT 'ES256',
    private_key_encrypted BYTEA NOT NULL,
    is_current            BOOLEAN NOT NULL DEFAULT false,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    removed_at            TIMESTAMPTZ
);

CREATE INDEX idx_signing_keys_active ON signing_keys(removed_at) WHERE removed_at IS NULL;
