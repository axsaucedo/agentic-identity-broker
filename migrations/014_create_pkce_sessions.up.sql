CREATE TABLE pkce_sessions (
    signature            TEXT        PRIMARY KEY,
    code_challenge       TEXT        NOT NULL,
    code_challenge_method TEXT       NOT NULL,
    expires_at           TIMESTAMPTZ NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pkce_sessions_expires_at ON pkce_sessions (expires_at);
