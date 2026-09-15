-- Keep approval state, sync version, and cross-instance notification atomic.
CREATE FUNCTION notify_approval_sync() RETURNS TRIGGER AS $$
BEGIN
    UPDATE approval_sync_state SET version = version + 1 WHERE id = 1;
    PERFORM pg_notify('approval_sync', '');
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_tool_approvals_sync
    AFTER INSERT OR UPDATE ON tool_approvals
    FOR EACH ROW
    EXECUTE FUNCTION notify_approval_sync();
