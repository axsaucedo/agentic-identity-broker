-- Migration 024: Create tool_approvals table
-- Stores human-in-the-loop tool call approval records

CREATE TABLE IF NOT EXISTS tool_approvals (
    id UUID PRIMARY KEY,
    principal VARCHAR(200) NOT NULL,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    gateway_client_id VARCHAR(255) NOT NULL,
    tool_name VARCHAR(255) NOT NULL,
    arguments JSONB NOT NULL DEFAULT '{}',
    arguments_hash VARCHAR(64) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    risk_level VARCHAR(50) NOT NULL DEFAULT '',
    mcp_session_id VARCHAR(255),
    agent_session_id VARCHAR(255),
    tool_invocation_id VARCHAR(255),
    opentelemetry_traceparent VARCHAR(55),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'denied')),
    persistence VARCHAR(20)
        CHECK (persistence IS NULL OR persistence IN ('once', 'session', 'permanent')),
    consumed BOOLEAN NOT NULL DEFAULT FALSE,
    approval_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    denied_at TIMESTAMPTZ,
    consumed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL
);

-- Partial unique index for idempotent pending approval creation.
-- Only one pending unconsumed approval per (principal, agent_id, tool_name, arguments_hash).
CREATE UNIQUE INDEX idx_tool_approvals_pending_dedup
    ON tool_approvals (principal, agent_id, tool_name, arguments_hash)
    WHERE status = 'pending' AND consumed = FALSE;

-- GIN index for JSONB arguments queries
CREATE INDEX idx_tool_approvals_arguments
    ON tool_approvals USING GIN (arguments);

-- Principal index for listing approvals by user
CREATE INDEX idx_tool_approvals_principal
    ON tool_approvals (principal);

-- Principal + agent index for pair-based queries
CREATE INDEX idx_tool_approvals_principal_agent
    ON tool_approvals (principal, agent_id);

-- Expiry index for TTL-based cleanup
CREATE INDEX idx_tool_approvals_expires_at
    ON tool_approvals (expires_at)
    WHERE status = 'pending';
