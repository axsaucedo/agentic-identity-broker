-- Migration 025: Create approval_sync_state table
-- Single-row table tracking the global mutation version for ETag generation

CREATE TABLE IF NOT EXISTS approval_sync_state (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    version BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT approval_sync_state_single_row CHECK (id = 1)
);

-- Seed with initial row (idempotent)
INSERT INTO approval_sync_state (id, version) VALUES (1, 0)
ON CONFLICT (id) DO NOTHING;
