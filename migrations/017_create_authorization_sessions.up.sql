CREATE TABLE authorization_sessions (
    session_id            TEXT        PRIMARY KEY,
    agent_id              UUID        NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    client_id             TEXT        NOT NULL,
    original_url          TEXT        NOT NULL,
    redirect_uri          TEXT        NOT NULL,
    scope                 TEXT        NOT NULL,
    state                 TEXT        NOT NULL,
    code_challenge        TEXT        NOT NULL,
    code_challenge_method TEXT        NOT NULL,
    cimd_metadata         JSONB,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at            TIMESTAMPTZ NOT NULL,
    consumed_at           TIMESTAMPTZ
);

CREATE INDEX idx_authorization_sessions_expires_at ON authorization_sessions (expires_at);
