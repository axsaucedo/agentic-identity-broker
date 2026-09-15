CREATE TABLE authorization_codes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code_hash       VARCHAR(64) NOT NULL UNIQUE,
    agent_id        UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    principal       VARCHAR(255) NOT NULL,
    redirect_uri    TEXT NOT NULL,
    code_challenge  VARCHAR(128) NOT NULL,
    scope           TEXT NOT NULL DEFAULT '',
    expires_at      TIMESTAMPTZ NOT NULL,
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_authorization_codes_expires ON authorization_codes(expires_at) WHERE used_at IS NULL;
