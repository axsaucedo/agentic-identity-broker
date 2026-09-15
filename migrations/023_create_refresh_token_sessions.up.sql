CREATE TABLE refresh_token_sessions (
    signature   TEXT PRIMARY KEY,
    request_id  TEXT NOT NULL,
    agent_id    UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    client_id   VARCHAR(255) NOT NULL,
    principal   VARCHAR(255) NOT NULL,
    scope       TEXT NOT NULL DEFAULT '',
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_token_sessions_request_id ON refresh_token_sessions(request_id);
CREATE INDEX idx_refresh_token_sessions_expires ON refresh_token_sessions(expires_at) WHERE used_at IS NULL;
